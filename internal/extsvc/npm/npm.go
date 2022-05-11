// Code for interfacing with Javascript and Typescript package registries such
// as npmjs.com.
package npm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	otlog "github.com/opentracing/opentracing-go/log"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Client interface {
	// GetPackageInfo gets a package's data from the registry, including versions.
	//
	// It is preferable to use this method instead of calling GetDependencyInfo for
	// multiple versions of a package in a loop.
	GetPackageInfo(ctx context.Context, pkg *reposource.NpmPackage) (*PackageInfo, error)

	// GetDependencyInfo gets a dependency's data from the registry.
	GetDependencyInfo(ctx context.Context, dep *reposource.NpmDependency) (*DependencyInfo, error)

	// FetchTarball fetches the sources in .tar.gz format for a dependency.
	//
	// The caller should close the returned reader after reading.
	FetchTarball(ctx context.Context, dep *reposource.NpmDependency) (io.ReadCloser, error)
}

func init() {
	// The HTTP client will transparently handle caching,
	// so we don't need to set up any on-disk caching here.
}

func FetchSources(ctx context.Context, client Client, dependency *reposource.NpmDependency) (tarball io.ReadCloser, err error) {
	operations := getOperations()

	ctx, _, endObservation := operations.fetchSources.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	return client.FetchTarball(ctx, dependency)
}

type HTTPClient struct {
	registryURL string
	doer        httpcli.Doer
	limiter     *rate.Limiter
	credentials string
}

func NewHTTPClient(urn string, registryURL string, credentials string, doer httpcli.Doer) *HTTPClient {
	return &HTTPClient{
		registryURL: registryURL,
		doer:        doer,
		limiter:     ratelimit.DefaultRegistry.Get(urn),
		credentials: credentials,
	}
}

type PackageInfo struct {
	Description string                     `json:"description"`
	Versions    map[string]*DependencyInfo `json:"versions"`
}

func (client *HTTPClient) GetPackageInfo(ctx context.Context, pkg *reposource.NpmPackage) (info *PackageInfo, err error) {
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
		return nil, errors.Newf("npm returned empty list of versions")
	}
	return &pkgInfo, nil
}

type DependencyInfo struct {
	Description string             `json:"description"`
	Dist        DependencyInfoDist `json:"dist"`
}

type DependencyInfoDist struct {
	TarballURL string `json:"tarball"`
}

type illFormedJSONError struct {
	url string
}

func (i illFormedJSONError) Error() string {
	return fmt.Sprintf("unexpected JSON output from npm request: url=%s", i.url)
}

func (client *HTTPClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req, ht := nethttp.TraceRequest(ot.GetTracer(ctx),
		req.WithContext(ctx),
		nethttp.OperationName("npm"),
		nethttp.ClientTrace(false))
	defer ht.Finish()
	startWait := time.Now()
	if err := client.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	if d := time.Since(startWait); d > 200*time.Millisecond {
		log15.Warn("npm self-enforced API rate limit: request delayed longer than expected due to rate limit", "delay", d)
	}
	return client.doer.Do(req)
}

type npmError struct {
	statusCode int
	err        error
}

func (n npmError) Error() string {
	if 100 <= n.statusCode && n.statusCode <= 599 {
		return fmt.Sprintf("npm HTTP response %d: %s", n.statusCode, n.err.Error())
	}
	return n.err.Error()
}

func (n npmError) NotFound() bool {
	return n.statusCode == http.StatusNotFound
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

func (client *HTTPClient) GetDependencyInfo(ctx context.Context, dep *reposource.NpmDependency) (*DependencyInfo, error) {
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

func (client *HTTPClient) FetchTarball(ctx context.Context, dep *reposource.NpmDependency) (io.ReadCloser, error) {
	if dep.TarballURL == "" {
		return nil, errors.New("empty TarballURL")
	}
	return client.makeGetRequest(ctx, dep.TarballURL)
}

var _ Client = &HTTPClient{}
