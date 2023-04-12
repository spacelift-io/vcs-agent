package agent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spacelift-io/spcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/spacelift-io/vcs-agent/privatevcs"
	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

// Agent is an agent connected to a VCS Gateway. Can handle only one concurrent request.
type Agent struct {
	poolConfig         *privatevcs.AgentPoolConfig
	targetBaseEndpoint string
	vendor             validation.Vendor
	metadata           map[string]string
	validator          validation.Strategy
}

// New creates a new Agent.
func New(poolConfig *privatevcs.AgentPoolConfig, targetBaseEndpoint, vendor string, validator validation.Strategy, metadata map[string]string) *Agent {
	return &Agent{
		metadata:           metadata,
		poolConfig:         poolConfig,
		targetBaseEndpoint: strings.TrimSuffix(targetBaseEndpoint, "/"),
		validator:          validator,
		vendor:             validation.Vendor(vendor),
	}
}

// Run runs the agent and handles any incoming requests.
func (a *Agent) Run(ctx *spcontext.Context) (outErr error) {
	insecure := os.Getenv("SPACELIFT_VCS_AGENT_DIAL_INSECURE") != ""

	var opts []grpc.DialOption
	if insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
	}

	conn, err := grpc.Dial(a.poolConfig.Host, opts...)
	if err != nil {
		return errors.Wrap(err, "couldn't dial gateway")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			if outErr == nil {
				outErr = errors.Wrap(err, "couldn't close connection")
			}
		}
	}()

	cli := privatevcs.NewGatewayClient(conn)

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
			responseMsg = a.handlePing(ctx, msg.Id, req.PingRequest)
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
	res, err := http.DefaultClient.Do(req)
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

	ctx.With(
		"elapsed", time.Since(start),
		"status_code", res.StatusCode,
	).Infof("Request served.")

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &privatevcs.Response{
			Id: id,
			Content: &privatevcs.Response_Error{
				Error: errors.Wrap(err, "couldn't read response body").Error(),
			},
		}
	}

	headers := make(map[string]string)
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

func (a *Agent) handlePing(ctx *spcontext.Context, id string, msg *privatevcs.PingRequest) *privatevcs.Response {
	return &privatevcs.Response{
		Id: id,
		Content: &privatevcs.Response_PingResponse{
			PingResponse: &privatevcs.PingResponse{},
		},
	}
}
