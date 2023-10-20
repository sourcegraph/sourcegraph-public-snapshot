package internalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

var frontendInternal = func() *url.URL {
	rawURL := env.Get("SRC_FRONTEND_INTERNAL", defaultFrontendInternal(), "HTTP address for internal frontend HTTP API.")
	return mustParseSourcegraphInternalURL(rawURL)
}()

// NOTE: this intentionally does not use the site configuration option because we need to make the decision
// about whether or not to use gRPC to fetch the site configuration in the first place.
var enableGRPC = env.MustGetBool("SRC_GRPC_ENABLE_CONF", false, "Enable gRPC for configuration updates")

func defaultFrontendInternal() string {
	if deploy.IsSingleBinary() {
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
	URL: frontendInternal.String(),
	getConfClient: syncx.OnceValues(func() (proto.ConfigServiceClient, error) {
		logger := log.Scoped("internalapi")
		conn, err := defaults.Dial(frontendInternal.Host, logger)
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

// mustParseSourcegraphInternalURL parses a frontend internal URL string and panics if it is invalid.
//
// The URL will be parsed with a default scheme of "http" and a default port of "80" if no scheme or port is specified.
func mustParseSourcegraphInternalURL(rawURL string) *url.URL {
	u, err := parseAddress(rawURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse frontend internal URL %q: %s", rawURL, err))
	}

	u = addDefaultScheme(u, "http")
	u = addDefaultPort(u)

	return u
}

// parseAddress parses rawAddress into a URL object. It accommodates cases where the rawAddress is a
// simple host:port pair without a URL scheme (e.g., "example.com:8080").
//
// This function aims to provide a flexible way to parse addresses that may or may not strictly adhere to the URL format.
func parseAddress(rawAddress string) (*url.URL, error) {
	addedScheme := false

	// Temporarily prepend "http://" if no scheme is present
	if !strings.Contains(rawAddress, "://") {
		rawAddress = "http://" + rawAddress
		addedScheme = true
	}

	parsedURL, err := url.Parse(rawAddress)
	if err != nil {
		return nil, err
	}

	// If we added the "http://" scheme, remove it from the final URL
	if addedScheme {
		parsedURL.Scheme = ""
	}

	return parsedURL, nil
}

// addDefaultScheme adds a default scheme to a URL if one is not specified.
//
// The original URL is not mutated. A copy is modified and returned.
func addDefaultScheme(original *url.URL, scheme string) *url.URL {
	if original == nil {
		return nil // don't panic
	}

	if original.Scheme != "" {
		return original
	}

	u := cloneURL(original)
	u.Scheme = scheme

	return u
}

// addDefaultPort adds a default port to a URL if one is not specified.
//
// If the URL scheme is "http" and no port is specified, "80" is used.
// If the scheme is "https", "443" is used.
//
// The original URL is not mutated. A copy is modified and returned.
func addDefaultPort(original *url.URL) *url.URL {
	if original == nil {
		return nil // don't panic
	}

	if original.Scheme == "http" && original.Port() == "" {
		u := cloneURL(original)
		u.Host = net.JoinHostPort(u.Host, "80")
		return u
	}

	if original.Scheme == "https" && original.Port() == "" {
		u := cloneURL(original)
		u.Host = net.JoinHostPort(u.Host, "443")
		return u
	}

	return original
}

// cloneURL returns a copy of the URL. It is safe to mutate the returned URL.
// This is copied from net/http/clone.go
func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}
