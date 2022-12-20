package repos

import (
	"net"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func checkConnection(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	// Best effort at finding a hostname. For example if rawURL is sourcegraph.com, then u.Host is
	// empty but Path is sourcegraph.com. Use that as a result.
	//
	// 👉 Also, we need to use u.Hostname() here because we want to strip any port numbers if they
	// are present in u.Host.
	addr := u.Hostname()
	if addr == "" {
		addr = u.Path
	}

	ips, err := net.LookupIP(addr)
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return errors.Newf("no IPs found for: %q", addr)
	}

	return nil
}
