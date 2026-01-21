package agent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spacelift-io/spcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	insecurePkg "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/spacelift-io/vcs-agent/privatevcs"
	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

// RequestDoer is an interface for an entity that can perform http request
type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// AgentConfig contains configuration parameters for creating a new Agent
type AgentConfig struct {
	PoolConfig                     *privatevcs.AgentPoolConfig
	TargetBaseEndpoint             string
	Vendor                         string
	Validator                      validation.Strategy
	Metadata                       map[string]string
	HTTPClient                     RequestDoer
	HTTPDisableResponseCompression bool
	DialInsecure                   bool
}

// Agent is an agent connected to a VCS Gateway. Can handle only one concurrent request.
type Agent struct {
	poolConfig                     *privatevcs.AgentPoolConfig
	targetBaseEndpoint             string
	vendor                         validation.Vendor
	metadata                       map[string]string
	validator                      validation.Strategy
	httpClient                     RequestDoer
	httpDisableResponseCompression bool
	dialInsecure                   bool
}

// New creates a new Agent.
func New(config *AgentConfig) (*Agent, error) {
	if config.PoolConfig == nil {
		return nil, errors.New("PoolConfig must be supplied")
	}

	if config.TargetBaseEndpoint == "" {
		return nil, errors.New("TargetBaseEndpoint must be supplied")
	}

	return &Agent{
		metadata:                       config.Metadata,
		poolConfig:                     config.PoolConfig,
		targetBaseEndpoint:             strings.TrimSuffix(config.TargetBaseEndpoint, "/"),
		validator:                      config.Validator,
		vendor:                         validation.Vendor(config.Vendor),
		httpClient:                     config.HTTPClient,
		httpDisableResponseCompression: config.HTTPDisableResponseCompression,
		dialInsecure:                   config.DialInsecure,
	}, nil
}

// Run runs the agent and handles any incoming requests.
func (a *Agent) Run(ctx *spcontext.Context) (outErr error) {
	var opts []grpc.DialOption
	if a.dialInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecurePkg.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
	}

	var host string
	if host = a.poolConfig.Host; host == "" {
		return errors.New("pool config host is empty, is the Spacelift backend misconfigured?")
	}

	client, err := grpc.NewClient(host, opts...)
	if err != nil {
		return errors.Wrap(err, "couldn't dial gateway")
	}
	defer func() {
		if err := client.Close(); err != nil {
			if outErr == nil {
				outErr = errors.Wrap(err, "couldn't close connection")
			}
		}
	}()

	cli := privatevcs.NewGatewayClient(client)

	metadataJSON, err := json.Marshal(a.metadata)
	if err != nil {
		return errors.Wrap(err, "couldn't marshal agent metadata as JSON")
	}

	md := metadata.New(map[string]string{
		privatevcs.MetadataFieldVcsAgentPoolULID: a.poolConfig.PoolULID,
		privatevcs.MetadataFieldVCSAgentKey:      a.poolConfig.Key,
		privatevcs.MetadataFieldVCSAgentMetadata: string(metadataJSON),
	})
	stream, err := cli.Connect(metadata.NewOutgoingContext(ctx, md))
	if err != nil {
		return errors.Wrap(err, "couldn't connect to gateway")
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "couldn't receive message from gateway")
		}

		var responseMsg *privatevcs.Response
		switch req := msg.Request.(type) {
		case *privatevcs.Request_HttpRequest:
			responseMsg = a.handleRequest(ctx, msg.Id, req.HttpRequest)
		case *privatevcs.Request_PingRequest:
			responseMsg = a.handlePing(msg.Id)
		}

		if err := stream.Send(responseMsg); err != nil {
			return errors.Wrap(err, "couldn't send response to gateway")
		}
	}

	if err := stream.CloseSend(); err != nil {
		return errors.Wrap(err, "couldn't close send stream")
	}

	return nil
}

func (a *Agent) handleRequest(ctx *spcontext.Context, id string, msg *privatevcs.HTTPRequest) *privatevcs.Response {
	req, err := http.NewRequest(msg.Method, a.targetBaseEndpoint+msg.Path, bytes.NewReader(msg.Body))
	if err != nil {
		return &privatevcs.Response{
			Id: id,
			Content: &privatevcs.Response_Error{
				Error: errors.Wrap(err, "couldn't create request").Error(),
			},
		}
	}

	ctx = validation.RewriteGitHubTarballRequest(ctx, a.vendor, req).With(
		"id", id,
		"pool_id", a.poolConfig.PoolULID,
		"method", req.Method,
		"raw_path", req.URL.EscapedPath(),
		"path", req.URL.Path,
	)

	for key, value := range msg.Headers {
		req.Header.Set(key, value)
	}

	if a.httpDisableResponseCompression {
		req.Header.Set("Accept-Encoding", "identity")
	}

	ctx, err = a.validator.Validate(ctx, a.vendor, req)
	if err != nil {
		return &privatevcs.Response{
			Id: id,
			Content: &privatevcs.Response_Error{
				Error: err.Error(),
			},
		}
	}

	timeoutCtx, cancel := spcontext.WithTimeout(ctx, time.Second*25)
	defer cancel()
	req = req.WithContext(timeoutCtx)

	start := time.Now()
	res, err := a.httpClient.Do(req)
	if err != nil {
		ctx.With(
			"error", err.Error(),
		).Errorf("Error serving request.")
		return &privatevcs.Response{
			Id: id,
			Content: &privatevcs.Response_Error{
				Error: errors.Wrap(err, "couldn't do request").Error(),
			},
		}
	}
	defer res.Body.Close() //nolint:errcheck // error not actionable after response is read

	ctx.With(
		"elapsed", time.Since(start),
		"status_code", res.StatusCode,
	).Infof("Request served.")

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &privatevcs.Response{
			Id: id,
			Content: &privatevcs.Response_Error{
				Error: errors.Wrap(err, "couldn't read response body").Error(),
			},
		}
	}

	headers := make(map[string]string, len(res.Header))
	for key, values := range res.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &privatevcs.Response{
		Id: id,
		Content: &privatevcs.Response_HttpResponse{
			HttpResponse: &privatevcs.HTTPResponse{
				Status:  int64(res.StatusCode),
				Headers: headers,
				Body:    data,
			},
		},
	}
}

func (a *Agent) handlePing(id string) *privatevcs.Response {
	return &privatevcs.Response{
		Id: id,
		Content: &privatevcs.Response_PingResponse{
			PingResponse: &privatevcs.PingResponse{},
		},
	}
}
