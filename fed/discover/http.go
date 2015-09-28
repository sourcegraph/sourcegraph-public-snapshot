package discover

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	urlpkg "net/url"
	"os"
	"time"

	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/wellknown"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/buildvar"
)

func init() {
	RepoFuncs = append(RepoFuncs, discoverRepoHTTP)
}

var (
	// InsecureHTTP is whether to allow discovery using plain
	// HTTP if HTTPS discovery fails.
	InsecureHTTP, _ = strconv.ParseBool(os.Getenv("HTTP_DISCOVERY_INSECURE"))

	// TestingHTTPPort is used for testing purposes only. If set, it
	// is used as the port number for HTTP/HTTPS discovery instead of
	// 80 and 443, respectively.
	TestingHTTPPort, _ = strconv.Atoi(os.Getenv("HTTP_DISCOVERY_PORT"))

	// UserAgent is the HTTP user agent string sent with all discovery
	// HTTP requests.
	UserAgent = "sourcegraph-http-discovery/" + buildvar.Version

	// HTTPClient is the HTTP client used for repo
	// discovery. Regardless of the value of this variable, httpGet is
	// used in tests.
	HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 15 * time.Second,
	}

	// httpGet can be overridden by tests. By default it adds the UserAgent string.
	httpGet = HTTPClient.Get
)

var (
	siteMu    sync.Mutex
	siteCache map[string]Info
)

func discoverSiteHTTP(ctx context.Context, scheme, host string) (Info, error) {
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("invalid scheme for site discovery: %s", scheme)
	}
	if strings.ContainsAny(host, "/@#?") {
		return nil, fmt.Errorf("invalid host for site discovery: %q", host)
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, &NotFoundError{Type: "site", Input: host, Err: fmt.Errorf("empty host for site discovery")}
	}

	// Check cache. Currently we cache successful responses forever
	// (until the process dies).
	siteMu.Lock()
	if info, ok := siteCache[scheme+host]; ok {
		siteMu.Unlock()
		return info, nil
	}
	siteMu.Unlock()

	url := &urlpkg.URL{Scheme: scheme, Host: host, Path: wellknown.ConfigPath}

	resp, err := httpGet(url.String())
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, &NotFoundError{Type: "site", Input: host, Err: err}
	}

	if ct := resp.Header.Get("content-type"); ct != "application/json" && ct != "application/json; charset=utf-8" {
		return nil, &NotFoundError{Type: "site", Input: host, Err: fmt.Errorf("invalid Content-Type %q", ct)}
	}

	var conf *sourcegraph.ServerConfig
	if err := json.NewDecoder(resp.Body).Decode(&conf); err != nil {
		return nil, err
	}

	// Validate.
	if conf.GRPCEndpoint == "" {
		return nil, &NotFoundError{
			Type:  "site",
			Input: host,
			Err:   fmt.Errorf("JSON document at %s is misssing a GRPCEndpoint value", url),
		}
	}
	if err := checkURL(conf.GRPCEndpoint); err != nil {
		return nil, &NotFoundError{Type: "site", Input: host, Err: err}
	}

	info := &remoteInfo{grpcEndpoint: conf.GRPCEndpoint}

	// Store in cache.
	siteMu.Lock()
	if siteCache == nil {
		siteCache = map[string]Info{}
	}
	siteCache[scheme+host] = info
	siteMu.Unlock()

	return info, nil
}

// checkURL returns an error iff urlStr is not a valid absolute URL.
func checkURL(urlStr string) error {
	u, err := urlpkg.Parse(urlStr)
	if err != nil {
		return err
	}
	if !u.IsAbs() {
		return fmt.Errorf("not an absolute URL: %s", u)
	}
	return nil
}

func discoverRepoHTTP(ctx context.Context, repo string) (Info, error) {
	u, err := urlpkg.Parse(repo)
	if err != nil {
		return nil, err
	}
	u.Scheme = ""
	if u.Host == "" {
		i := strings.Index(u.Path, "/")
		if i > 0 {
			u.Host = u.Path[:i]
			u.Path = u.Path[i+1:]
		}
	}
	u.Path = ""

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil && !strings.Contains(err.Error(), "missing port") {
		return nil, err
	}
	if host == "" {
		host = u.Host
	}
	if TestingHTTPPort != 0 {
		host += ":" + strconv.Itoa(TestingHTTPPort)
	}

	discoverHTTPSFallback := func(h string) (Info, error) {
		info, err := discoverSiteHTTP(ctx, "https", h)
		if err != nil && IsNotFound(err) && InsecureHTTP {
			return discoverSiteHTTP(ctx, "http", h)
		}
		return info, err
	}

	info, err := discoverHTTPSFallback(host)
	if err != nil && u.Host != host {
		info, err = discoverHTTPSFallback(u.Host)
	}
	return info, err
}
