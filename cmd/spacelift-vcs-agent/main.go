package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bugsnag/bugsnag-go"
	"github.com/go-kit/kit/log"
	"github.com/spacelift-io/spcontext"
	"github.com/urfave/cli"

	"github.com/spacelift-io/vcs-agent/agent"
	"github.com/spacelift-io/vcs-agent/logging"
	"github.com/spacelift-io/vcs-agent/privatevcs"
	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
	"github.com/spacelift-io/vcs-agent/privatevcs/validation/allowlist"
	"github.com/spacelift-io/vcs-agent/privatevcs/validation/blocklist"
)

const (
	vendorAzureDevOps         = "azure_devops"
	vendorBitbucketDatacenter = "bitbucket_datacenter"
	vendorGitHubEnterprise    = "github_enterprise"
	vendorGitlab              = "gitlab"
)

// VERSION is the version printed by the resulting binary.
var VERSION = "development"

// BugsnagAPIKey is used to send error information to Bugsnag.
var BugsnagAPIKey string

var (
	availableVendors = []string{
		vendorAzureDevOps,
		vendorBitbucketDatacenter,
		vendorGitHubEnterprise,
		vendorGitlab,
	}

	flagAllowedProjects = cli.StringFlag{
		Name:   "allowed-projects",
		EnvVar: "SPACELIFT_VCS_AGENT_ALLOWED_PROJECTS",
		Usage:  "Regexp matching allowed projects for API calls. Projects are in the form: 'group/repository'.",
		Value:  ".*",
	}

	flagBugsnagAPIKey = cli.StringFlag{
		Name:   "bugsnag-api-key",
		EnvVar: "SPACELIFT_VCS_AGENT_BUGSNAG_API_KEY",
		Usage:  "Override the Bugsnag API key used for error reporting.",
		Value:  "",
	}

	flagBugsnagDisable = cli.BoolFlag{
		Name:   "disable-bugsnag",
		EnvVar: "SPACELIFT_VCS_AGENT_BUGSNAG_DISABLE",
		Usage:  "Disable Bugsnag reporting entirely.",
	}

	flagParallelism = cli.IntFlag{
		Name:   "parallelism",
		EnvVar: "SPACELIFT_VCS_AGENT_PARALLELISM",
		Usage:  "Number of streams to create. Each stream can handle one request simultaneously.",
		Value:  4,
	}

	flagPoolToken = cli.StringFlag{
		Name:     "token",
		EnvVar:   "SPACELIFT_VCS_AGENT_POOL_TOKEN",
		Usage:    "Token received on VCS Agent Pool creation",
		Required: true,
	}

	flagTargetBaseEndpoint = cli.StringFlag{
		Name:     "target-base-endpoint",
		EnvVar:   "SPACELIFT_VCS_AGENT_TARGET_BASE_ENDPOINT",
		Usage:    "Target endpoint this agent proxies to. Should include protocol (http/https).",
		Required: true,
	}

	flagVCSVendor = cli.StringFlag{
		Name:     "vendor",
		EnvVar:   "SPACELIFT_VCS_AGENT_VENDOR",
		Usage:    fmt.Sprintf("VCS vendor proxied by this agent. Available vendors: %s", strings.Join(availableVendors, ", ")),
		Required: true,
	}

	flagUseAllowlist = cli.BoolFlag{
		Name:   "use-allowlist",
		EnvVar: "SPACELIFT_VCS_AGENT_USE_ALLOWLIST",
		Usage:  "Whether to use the allowlist to validate API calls. Incompatible with --blocklist-path.",
	}

	flagBlocklistPath = cli.StringFlag{
		Name:   "blocklist-path",
		EnvVar: "SPACELIFT_VCS_AGENT_BLOCKLIST_PATH",
		Usage:  "Path to the YAML blocklist file. Incompatible with --use-allowlist.",
	}

	flagDebugPrintAll = cli.BoolFlag{
		Name:   "debug-print-all",
		EnvVar: "SPACELIFT_VCS_AGENT_DEBUG_PRINT_ALL",
		Usage:  "Whether to print all requests and responses to stdout.",
	}

	flagHTTPDisableResponseCompression = cli.BoolFlag{
		Name:   "http-disable-response-compression",
		EnvVar: "SPACELIFT_VCS_AGENT_HTTP_DISABLE_RESPONSE_COMPRESSION",
		Usage:  "Whether to disable HTTP response compression.",
	}
)

var app = &cli.App{
	Flags: []cli.Flag{
		flagAllowedProjects,
		flagBugsnagAPIKey,
		flagParallelism,
		flagPoolToken,
		flagTargetBaseEndpoint,
		flagVCSVendor,
		flagDebugPrintAll,
		flagHTTPDisableResponseCompression,
	},
	Action: func(cmdCtx *cli.Context) error {
		availableVendorsMap := make(map[string]bool)
		for _, vendor := range availableVendors {
			availableVendorsMap[vendor] = true
		}
		vendor := cmdCtx.String(flagVCSVendor.Name)
		if !availableVendorsMap[vendor] {
			stdlog.Fatalf("invalid vendor specified: '%s', available vendors: [%s]", vendor, strings.Join(availableVendors, ", "))
		}

		var opts []spcontext.ContextOption
		ctx := spcontext.New(log.NewJSONLogger(os.Stdout), opts...)

		apiKey := BugsnagAPIKey
		if apiKeyOverride := cmdCtx.String(flagBugsnagAPIKey.Name); len(apiKeyOverride) > 0 {
			apiKey = apiKeyOverride
		}

		if !cmdCtx.Bool(flagBugsnagDisable.Name) {
			ctx.Notifier = bugsnag.New(bugsnag.Configuration{
				APIKey: apiKey,
				Logger: &spcontext.BugsnagLogger{
					Ctx: *ctx,
				},
				Synchronous: true,
			})

			defer ctx.Notifier.AutoNotify(ctx)
		}

		var poolConfig privatevcs.AgentPoolConfig
		configBytes, err := base64.StdEncoding.DecodeString(cmdCtx.String(flagPoolToken.Name))
		if err != nil {
			stdlog.Fatal("invalid pool token: ", err.Error())
		}
		if err := json.Unmarshal(configBytes, &poolConfig); err != nil {
			stdlog.Fatal("invalid pool token: ", err.Error())
		}

		if cmdCtx.IsSet(flagAllowedProjects.Name) && vendor == vendorGitHubEnterprise {
			stdlog.Fatal("--allowed-projects is not currently supported for the GitHub Enterprise integration")
		}

		agentMetadata := loadMetadata()

		var validationStrategy validation.Strategy = new(blocklist.List)

		useAllowlist := cmdCtx.Bool(flagUseAllowlist.Name)
		if useAllowlist {
			if validationStrategy, err = allowlist.New(cmdCtx.String(flagAllowedProjects.Name)); err != nil {
				stdlog.Fatal("could not create request allowlist: ", err.Error())
			}
		}

		if cmdCtx.IsSet(flagBlocklistPath.Name) {
			if useAllowlist {
				stdlog.Fatal("--use-allowlist and --blocklist-path are mutually exclusive")
			}

			if validationStrategy, err = blocklist.Load(cmdCtx.String(flagBlocklistPath.Name)); err != nil {
				stdlog.Fatal("could not create request blocklist: ", err.Error())
			}
		}

		var httpClient agent.RequestDoer = http.DefaultClient
		if cmdCtx.Bool(flagDebugPrintAll.Name) {
			httpClient = &logging.HTTPClient{
				Wrapped: http.DefaultClient,
				Out:     &logging.ConcurrentSafeWriter{Out: os.Stdout},
			}
		}

		a := agent.New(
			&poolConfig,
			cmdCtx.String(flagTargetBaseEndpoint.Name),
			vendor,
			validationStrategy,
			agentMetadata,
			httpClient,
		)
		a.HTTPDisableResponseCompression = cmdCtx.Bool(flagHTTPDisableResponseCompression.Name)

		parallelismSemaphore := make(chan struct{}, cmdCtx.Int(flagParallelism.Name))

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM, syscall.SIGHUP)
		ctx, cancel := spcontext.WithCancel(ctx)

		go func() {
			s := <-signals
			ctx.Infof("signal received: %s", s.String())
			cancel()
		}()

		wg := sync.WaitGroup{}

	runLoop:
		for {
			select {
			case parallelismSemaphore <- struct{}{}:
			case <-ctx.Done():
				break runLoop
			}
			wg.Add(1)
			ctx.Infof("Starting new stream.")
			go func() {
				func() {
					defer wg.Done()
					defer func() {
						// Recover error which has already been sent by bugsnag below.
						recover()
					}()
					defer ctx.Notifier.AutoNotify(ctx)

					if err := a.Run(ctx); err != nil {
						if !strings.Contains(err.Error(), "context canceled") {
							_ = ctx.RawError(err, "error running agent")
						}
					}
				}()
				time.Sleep(time.Second * 5)
				<-parallelismSemaphore
			}()
		}

		wg.Wait()

		return nil
	},
	Copyright: "Spacelift, Inc.",
	Usage:     "The VCS Agent is used to proxy requests to your VCS provider if Spacelift cannot access it directly.",
	Version:   VERSION,
}

func main() {
	if err := app.Run(os.Args); err != nil {
		stdlog.Fatal(err)
	}
}

func loadMetadata() map[string]string {
	const metadataPrefix = "SPACELIFT_METADATA_"

	metadata := make(map[string]string)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], metadataPrefix) {
			name := strings.TrimPrefix(pair[0], metadataPrefix)

			if name == "" {
				continue
			}

			metadata[name] = pair[1]
		}
	}

	return metadata
}
