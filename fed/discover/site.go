package discover

import (
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
	info, err := discoverSiteHTTP(ctx, "https", host)
	if err == nil {
		return info, nil
	} else if IsNotFound(err) && InsecureHTTP {
		return discoverSiteHTTP(ctx, "http", host)
	}
	return nil, err
}

// SiteURL is like Site, but it accepts a URL instead of just a
// hostname. If the URL is "http://" (non-HTTPS), then insecure HTTP
// site discovery is allowed even if InsecureHTTP is false.
//
// If urlStr does not begin with either "https://" or "http://", it is
// prepended with "https://". This is so that you must be explicit if
// you want HTTP (non-HTTPS) discovery.
func SiteURL(ctx context.Context, urlStr string) (Info, error) {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return discoverSiteHTTP(ctx, url.Scheme, url.Host)
}
