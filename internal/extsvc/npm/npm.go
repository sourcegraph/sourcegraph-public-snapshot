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
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Client interface {
	// GetPackageInfo gets a package's data from the registry, including versions.
	//
	// It is preferable to use this method instead of calling GetDependencyInfo for
	// multiple versions of a package in a loop.
	GetPackageInfo(ctx context.Context, pkg *reposource.NPMPackage) (*PackageInfo, error)

	// GetDependencyInfo gets a dependency's data from the registry.
	GetDependencyInfo(ctx context.Context, dep *reposource.NPMDependency) (*DependencyInfo, error)

	// FetchTarball fetches the sources in .tar.gz format for a dependency.
	//
	// The caller should close the returned reader after reading.
	FetchTarball(ctx context.Context, dep *reposource.NPMDependency) (io.ReadCloser, error)
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

func FetchSources(ctx context.Context, client Client, dependency *reposource.NPMDependency) (tarball io.ReadCloser, err error) {
	ctx, endObservation := operations.fetchSources.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	return client.FetchTarball(ctx, dependency)
}

func Exists(ctx context.Context, client Client, dependency *reposource.NPMDependency) (exists bool, err error) {
	ctx, endObservation := operations.exists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Bool("exists", exists),
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	if _, err = client.GetDependencyInfo(ctx, dependency); err != nil {
		if errors.HasType(err, npmError{}) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

type PackageInfo struct {
	Description string                     `json:"description"`
	Versions    map[string]*DependencyInfo `json:"versions"`
}

func (client *HTTPClient) GetPackageInfo(ctx context.Context, pkg *reposource.NPMPackage) (info *PackageInfo, err error) {
	url := fmt.Sprintf("%s/%s", client.registryURL, pkg.PackageSyntax())
	body, err := client.makeGetRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	var pkgInfo PackageInfo
	if err := json.NewDecoder(body).Decode(&pkgInfo); err != nil {
		return nil, err
	}
	if len(pkgInfo.Versions) == 0 {
		return nil, errors.Newf("NPM returned empty list of versions")
	}
	return &pkgInfo, nil
}

type DependencyInfo struct {
	Dist DependencyInfoDist `json:"dist"`
}

type DependencyInfoDist struct {
	TarballURL string `json:"tarball"`
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

func (client *HTTPClient) makeGetRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if client.credentials != "" {
		req.Header.Set("Authorization", "Bearer "+client.credentials)
	}

	resp, err := client.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bodyBuffer bytes.Buffer
	if _, err := io.Copy(&bodyBuffer, resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, npmError{resp.StatusCode, errors.New(bodyBuffer.String())}
	}

	return io.NopCloser(&bodyBuffer), nil
}

func (client *HTTPClient) GetDependencyInfo(ctx context.Context, dep *reposource.NPMDependency) (*DependencyInfo, error) {
	// https://github.com/npm/registry/blob/master/docs/REGISTRY-API.md#getpackageversion
	url := fmt.Sprintf("%s/%s/%s", client.registryURL, dep.PackageSyntax(), dep.Version)
	body, err := client.makeGetRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	var info DependencyInfo
	if json.NewDecoder(body).Decode(&info) != nil {
		return nil, illFormedJSONError{url: url}
	}
	return &info, nil
}

func (client *HTTPClient) FetchTarball(ctx context.Context, dep *reposource.NPMDependency) (io.ReadCloser, error) {
	info, err := client.GetDependencyInfo(ctx, dep)
	if err != nil {
		return nil, err
	}
	return client.makeGetRequest(ctx, info.Dist.TarballURL)
}

var _ Client = &HTTPClient{}
