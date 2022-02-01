// Code for interfacing with Javascript and Typescript package registries such
// as NPMJS.com.
package npm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Client interface {
	// AvailablePackageVersions lists the available versions for an NPM package.
	//
	// It is preferable to use this method instead of calling DoesDependencyExist
	// in a loop, if different dependencies may share the same underlying package.
	//
	// If err is nil, versions should be non-empty.
	AvailablePackageVersions(ctx context.Context, pkg reposource.NPMPackage) (versions map[string]struct{}, err error)

	// DoesDependencyExist checks if a particular dependency exists on a particular registry.
	//
	// exists should be checked even if err is nil.
	DoesDependencyExist(ctx context.Context, dep reposource.NPMDependency) (exists bool, err error)

	// FetchTarball fetches the sources in .tar.gz format for a dependency.
	//
	// The caller should close the returned reader after reading.
	//
	// The return value is an io.ReadSeekCloser instead of an io.ReadCloser
	// to allow callers to iterate over the reader multiple times if needed.
	FetchTarball(ctx context.Context, dep reposource.NPMDependency) (io.ReadSeekCloser, error)
}

var (
	observationContext *observation.Context
	operations         *Operations
)

func init() {
	observationContext = &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	operations = NewOperations(observationContext)

	// The HTTP client will transparently handle caching,
	// so we don't need to set up any on-disk caching here.
}

func FetchSources(ctx context.Context, client Client, dependency reposource.NPMDependency) (tarball io.ReadSeekCloser, err error) {
	ctx, endObservation := operations.fetchSources.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})
	return client.FetchTarball(ctx, dependency)
}

func Exists(ctx context.Context, client Client, dependency reposource.NPMDependency) (err error) {
	ctx, endObservation := operations.exists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	exists, err := client.DoesDependencyExist(ctx, dependency)
	if err != nil {
		return errors.Wrapf(err, "tried to check if npm package %s exists but failed", dependency.PackageManagerSyntax())
	}
	if !exists {
		return errors.Newf("npm package %s does not exist", dependency.PackageManagerSyntax())
	}
	return nil
}

type HTTPClient struct {
	registryURL string
	doer        httpcli.Doer
	limiter     *rate.Limiter
	credentials string
}

func NewHTTPClient(registryURL string, rateLimit *schema.NPMRateLimit, credentials string) *HTTPClient {
	var requestsPerHour float64
	if rateLimit == nil || !rateLimit.Enabled {
		requestsPerHour = math.Inf(1)
	} else {
		requestsPerHour = rateLimit.RequestsPerHour
	}
	defaultLimiter := rate.NewLimiter(rate.Limit(requestsPerHour/3600.0), 100)
	cachedLimiter := ratelimit.DefaultRegistry.GetOrSet(registryURL, defaultLimiter)
	return &HTTPClient{
		registryURL,
		httpcli.ExternalDoer,
		cachedLimiter,
		credentials,
	}
}

type packageInfo struct {
	Versions map[string]interface{} `json:"versions"`
}

func (client *HTTPClient) AvailablePackageVersions(ctx context.Context, pkg reposource.NPMPackage) (versions map[string]struct{}, err error) {
	url := fmt.Sprintf("%s/%s", client.registryURL, pkg.PackageSyntax())
	jsonBytes, err := client.makeGetRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	var pkgInfo packageInfo
	if err := json.Unmarshal(jsonBytes, &pkgInfo); err != nil {
		return nil, err
	}
	if len(pkgInfo.Versions) == 0 {
		return nil, fmt.Errorf("NPM returned empty list of versions")
	}
	versions = map[string]struct{}{}
	for k := range pkgInfo.Versions {
		versions[k] = struct{}{}
	}
	return versions, nil
}

type npmDependencyDist struct {
	TarballURL string `json:"tarball"`
}

type npmDependencyInfo struct {
	Dist npmDependencyDist `json:"dist"`
}

type illFormedJSONError struct {
	url string
}

func (i illFormedJSONError) Error() string {
	return fmt.Sprintf("unexpected JSON output from NPM request: url=%s", i.url)
}

func (client *HTTPClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req, ht := nethttp.TraceRequest(ot.GetTracer(ctx),
		req.WithContext(ctx),
		nethttp.OperationName("NPM"),
		nethttp.ClientTrace(false))
	defer ht.Finish()
	startWait := time.Now()
	if err := client.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	if d := time.Since(startWait); d > 200*time.Millisecond {
		log15.Warn("NPM self-enforced API rate limit: request delayed longer than expected due to rate limit", "delay", d)
	}
	return client.doer.Do(req)
}

type npmError struct {
	statusCode int
	err        error
}

func (n npmError) Error() string {
	if 100 <= n.statusCode && n.statusCode <= 599 {
		return fmt.Sprintf("NPM HTTP response %d: %s", n.statusCode, n.err.Error())
	}
	return n.err.Error()
}

func (client *HTTPClient) makeGetRequest(ctx context.Context, url string) (responseBody []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if client.credentials != "" {
		req.Header.Set("Authorization", "Bearer "+client.credentials)
	}
	resp, err := client.do(ctx, req)
	if err != nil {
		if resp == nil { // possible if you pass in an incorrect registry URL
			return nil, err
		}
		return nil, npmError{resp.StatusCode, err}
	}
	var bodyBuffer bytes.Buffer
	if _, err := io.Copy(&bodyBuffer, resp.Body); err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, npmError{resp.StatusCode, fmt.Errorf("%s", bodyBuffer.String())}
	}
	return bodyBuffer.Bytes(), nil
}

func (client *HTTPClient) getDependencyInfo(ctx context.Context, dep reposource.NPMDependency) (info npmDependencyInfo, err error) {
	// https://github.com/npm/registry/blob/master/docs/REGISTRY-API.md#getpackageversion
	url := fmt.Sprintf("%s/%s/%s", client.registryURL, dep.Package.PackageSyntax(), dep.Version)
	respBytes, err := client.makeGetRequest(ctx, url)
	if err != nil {
		return info, err
	}
	if json.Unmarshal(respBytes, &info) != nil {
		return info, illFormedJSONError{url: url}
	}
	return info, nil
}

func (client *HTTPClient) DoesDependencyExist(ctx context.Context, dep reposource.NPMDependency) (exists bool, err error) {
	_, err = client.getDependencyInfo(ctx, dep)
	var npmErr npmError
	if err != nil && errors.As(err, &npmErr) && npmErr.statusCode == http.StatusNotFound {
		log15.Info("npm dependency does not exist", "dependency", dep.PackageManagerSyntax())
		return false, err
	}
	var e illFormedJSONError
	if errors.As(err, &e) {
		log15.Warn("received ill-formed JSON payload from NPM Registry API", "error", e)
		return false, nil
	}
	return err == nil, err
}

func (client *HTTPClient) FetchTarball(ctx context.Context, dep reposource.NPMDependency) (io.ReadSeekCloser, error) {
	info, err := client.getDependencyInfo(ctx, dep)
	if err != nil {
		return nil, err
	}
	respBytes, err := client.makeGetRequest(ctx, info.Dist.TarballURL)
	if err != nil {
		return nil, err
	}
	return &NopSeekCloser{bytes.NewReader(respBytes)}, nil
}

var _ Client = &HTTPClient{}

type NopSeekCloser struct {
	io.ReadSeeker
}

func (n *NopSeekCloser) Close() error {
	return nil
}

var _ io.ReadSeekCloser = &NopSeekCloser{}
