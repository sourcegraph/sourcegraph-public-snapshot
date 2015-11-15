package discover

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/context"
)

// Site performs the Sourcegraph site configuration discovery process
// for the specified hostname. If successful, the returned Info
// enables communication with the specified site (you just need to
// call its NewContext method on the context you're using).
//
// Site is used when you have a host, such as "src.example.com:1234",
// and you want to communicate with the Sourcegraph site located at
// that host.
func Site(ctx context.Context, host string) (Info, error) {
	if strings.Contains(host, ":") {
		panic(fmt.Sprintf("host cannot contain colon: %q", host))
	}

	scheme, port := schemeAndPortForSiteDiscovery()
	return discoverSiteHTTP(ctx, scheme, host, port)
}

// SiteURL is like Site, but it accepts a URL instead of just a
// hostname. If the URL is "http://" (non-HTTPS), then insecure HTTP
// site discovery is allowed even if InsecureHTTP is false.
//
// If urlStr does not begin with either "https://" or "http://", it is
// prepended with "https://". This is so that you must be explicit if
// you want HTTP (non-HTTPS) discovery.
func SiteURL(ctx context.Context, urlStr string) (Info, error) {
	defaultScheme, _ := schemeAndPortForSiteDiscovery()

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = defaultScheme + "://" + urlStr
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(url.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			err = nil
			host = url.Host

			switch url.Scheme {
			case "http":
				port = "80"
			case "https":
				port = "443"
			}
			if TestingHTTPPort != "" {
				port = TestingHTTPPort
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return discoverSiteHTTP(ctx, url.Scheme, host, port)
}

func schemeAndPortForSiteDiscovery() (scheme, port string) {
	if InsecureHTTP {
		scheme = "http"
		port = "80"
	} else {
		scheme = "https"
		port = "443"
	}
	if TestingHTTPPort != "" {
		port = TestingHTTPPort
	}

	return scheme, port
}
