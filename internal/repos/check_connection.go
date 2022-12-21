package repos

import (
	"net"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// checkConnection parses the rawURL and makes a best effort attempt to obtain a hostname. It then
// performs an IP lookup on that hostname and returns a non-nil error on failure.
//
// At the moment this function is only limited to doing IP lookups. We may want/have to expand this
// to support other code hosts or to add more checks (for example making a test API call to verify
// the authorization, etc).
func checkConnection(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return errors.Wrap(err, "invalid or bad url for connection check")
	}

	// Best effort at finding a hostname. For example if rawURL is sourcegraph.com, then u.Host is
	// empty but Path is sourcegraph.com. Use that as a result.
	//
	// 👉 Also, we need to use u.Hostname() here because we want to strip any port numbers if they
	// are present in u.Host.
	hostname := u.Hostname()
	if hostname == "" {
		if u.Scheme != "" {
			// rawURL is most likely something like "sourcegraph.com:80", read from u.Scheme.
			hostname = u.Scheme
		} else if u.Path != "" {
			// rawURL is most likely something like "sourcegraph.com:80", read from u.Path.
			hostname = u.Path
		} else {
			return errors.Newf("unsupported url format (%q) for connection check", rawURL)
		}
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return errors.Wrap(err, "connection check failed")
	}

	if len(ips) == 0 {
		return errors.Newf("connection check failed, no IP addresses found for hostname %q", hostname)
	}

	return nil
}
