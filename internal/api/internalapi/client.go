package internalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	proto "github.com/sourcegraph/sourcegraph/internal/api/internalapi/v1"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/syncx"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var frontendInternal = env.Get("SRC_FRONTEND_INTERNAL", defaultFrontendInternal(), "HTTP address for internal frontend HTTP API.")

// NOTE: this intentionally does not use the site configuration option because we need to make the decision
// about whether or not to use gRPC to fetch the site configuration in the first place.
var enableGRPC = env.MustGetBool("SRC_GRPC_ENABLE_CONF", false, "Enable gRPC for configuration updates")

func defaultFrontendInternal() string {
	if deploy.IsApp() {
		return "localhost:3090"
	}
	return "sourcegraph-frontend-internal"
}

type internalClient struct {
	// URL is the root to the internal API frontend server.
	URL string

	getConfClient func() (proto.ConfigServiceClient, error)
}

var Client = &internalClient{
	URL: "http://" + frontendInternal,
	getConfClient: syncx.OnceValues(func() (proto.ConfigServiceClient, error) {
		logger := log.Scoped("internalapi", "")
		conn, err := defaults.Dial(frontendInternal, logger)
		if err != nil {
			return nil, err
		}
		return proto.NewConfigServiceClient(conn), nil
	}),
}

var requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_frontend_internal_request_duration_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"category", "code"})

// MockClientConfiguration mocks (*internalClient).Configuration.
var MockClientConfiguration func() (conftypes.RawUnified, error)

func (c *internalClient) Configuration(ctx context.Context) (conftypes.RawUnified, error) {
	if MockClientConfiguration != nil {
		return MockClientConfiguration()
	}

	if enableGRPC {
		cc, err := c.getConfClient()
		if err != nil {
			return conftypes.RawUnified{}, err
		}
		resp, err := cc.GetConfig(ctx, &proto.GetConfigRequest{})
		if err != nil {
			return conftypes.RawUnified{}, err
		}
		var raw conftypes.RawUnified
		raw.FromProto(resp.RawUnified)
		return raw, nil
	}

	var cfg conftypes.RawUnified
	err := c.postInternal(ctx, "configuration", nil, &cfg)
	return cfg, err
}

// postInternal sends an HTTP post request to the internal route.
func (c *internalClient) postInternal(ctx context.Context, route string, reqBody, respBody any) error {
	return c.meteredPost(ctx, "/.internal/"+route, reqBody, respBody)
}

func (c *internalClient) meteredPost(ctx context.Context, route string, reqBody, respBody any) error {
	start := time.Now()
	statusCode, err := c.post(ctx, route, reqBody, respBody)
	d := time.Since(start)

	code := strconv.Itoa(statusCode)
	if err != nil {
		code = "error"
	}
	requestDuration.WithLabelValues(route, code).Observe(d.Seconds())
	return err
}

// post sends an HTTP post request to the provided route. If reqBody is
// non-nil it will Marshal it as JSON and set that as the Request body. If
// respBody is non-nil the response body will be JSON unmarshalled to resp.
func (c *internalClient) post(ctx context.Context, route string, reqBody, respBody any) (int, error) {
	var data []byte
	if reqBody != nil {
		var err error
		data, err = json.Marshal(reqBody)
		if err != nil {
			return -1, err
		}
	}

	req, err := http.NewRequest("POST", c.URL+route, bytes.NewBuffer(data))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Check if we have an actor, if not, ensure that we use our internal actor since
	// this is an internal request.
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() && !a.IsInternal() {
		ctx = actor.WithInternalActor(ctx)
	}

	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if err := checkAPIResponse(resp); err != nil {
		return resp.StatusCode, err
	}

	if respBody != nil {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(respBody)
	}
	return resp.StatusCode, nil
}

func checkAPIResponse(resp *http.Response) error {
	if 200 > resp.StatusCode || resp.StatusCode > 299 {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		b := buf.Bytes()
		errString := string(b)
		if errString != "" {
			return errors.Errorf(
				"internal API response error code %d: %s (%s)",
				resp.StatusCode,
				errString,
				resp.Request.URL,
			)
		}
		return errors.Errorf("internal API response error code %d (%s)", resp.StatusCode, resp.Request.URL)
	}
	return nil
}
