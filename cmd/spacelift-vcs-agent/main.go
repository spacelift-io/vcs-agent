package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bugsnag/bugsnag-go"
	"github.com/go-kit/kit/log"
	"github.com/urfave/cli"

	"github.com/spacelift-io/spcontext"

	"github.com/spacelift-io/vcs-agent/agent"
	"github.com/spacelift-io/vcs-agent/privatevcs"
)

const (
	vendorBitbucketDatacenter = "bitbucket_datacenter"
	vendorGitHubEnterprise    = "github_enterprise"
	vendorGitlab              = "gitlab"
)

var VERSION = "development"

var (
	availableVendors = []string{
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
)

var app = &cli.App{
	Flags: []cli.Flag{
		flagAllowedProjects,
		flagBugsnagAPIKey,
		flagParallelism,
		flagPoolToken,
		flagTargetBaseEndpoint,
		flagVCSVendor,
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

		apiKey := "1b0ac0ea378c85618ebb2fa112fd11e0"
		if apiKeyOverride := cmdCtx.String(flagBugsnagAPIKey.Name); len(apiKeyOverride) > 0 {
			apiKey = apiKeyOverride
		}

		ctx.Notifier = bugsnag.New(bugsnag.Configuration{
			APIKey: apiKey,
			Logger: &spcontext.BugsnagLogger{
				Ctx: *ctx,
			},
			Synchronous: true,
		})

		defer ctx.Notifier.AutoNotify(ctx)

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

		allowedProjectsRegexp, err := regexp.Compile(cmdCtx.String(flagAllowedProjects.Name))
		if err != nil {
			stdlog.Fatal("couldn't compile allowed projects regexp: ", err)
		}

		agentMetadata := loadMetadata()

		a := agent.New(&poolConfig, cmdCtx.String(flagTargetBaseEndpoint.Name), vendor, allowedProjectsRegexp, agentMetadata)

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
							ctx.RawError(err, "error running agent")
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
