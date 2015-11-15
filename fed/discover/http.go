package discover

import (
	"fmt"
	"net"
	"net/http"
	urlpkg "net/url"
	"os"
	"time"

	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
)

func init() {
	RepoFuncs = append(RepoFuncs, discoverRepoHTTP)
}

var (
	// InsecureHTTP is whether to perform discovery using plain
	// HTTP (not HTTPS).
	InsecureHTTP, _ = strconv.ParseBool(os.Getenv("HTTP_DISCOVERY_INSECURE"))

	// TestingHTTPPort is used for testing purposes only. If set, it
	// is used as the port number for HTTP/HTTPS discovery instead of
	// 80 and 443, respectively.
	TestingHTTPPort = os.Getenv("HTTP_DISCOVERY_PORT")

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

func discoverSiteHTTP(ctx context.Context, scheme, host, port string) (Info, error) {
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

	return &remoteInfo{grpcEndpoint: fmt.Sprintf("%s://%s:%s", scheme, host, port)}, nil
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

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil && !strings.Contains(err.Error(), "missing port") {
		return nil, err
	}
	if host == "" {
		host = u.Host
	}

	lookupRepo := func(info Info, path string) error {
		ctx, err := info.NewContext(ctx)
		if err != nil {
			return err
		}

		cl := sourcegraph.NewClientFromContext(ctx)
		if _, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: path}); err != nil {
			if grpc.Code(err) == codes.NotFound {
				return &NotFoundError{
					Type:  "repo",
					Input: repo,
					Err:   err,
				}
			}

			// Probably unable to even contact the host.
			return &NotFoundError{
				Type:  "site",
				Input: host,
				Err:   err,
			}
		}
		return nil
	}
	discoverHTTPSFallback := func(h string) (Info, error) {
		info, err := SiteURL(ctx, h)
		if err != nil {
			return nil, err
		}

		if err := lookupRepo(info, u.Path); err != nil {
			return nil, err
		}
		return info, nil
	}

	info, err := discoverHTTPSFallback(host)
	if err != nil && u.Host != host {
		info, err = discoverHTTPSFallback(u.Host)
	}
	if err != nil {
		return nil, err
	}
	return info, nil
}
