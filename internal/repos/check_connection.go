package repos

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getHostname(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", errors.Wrap(err, "invalid or bad url for connection check")
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
			return "", errors.Newf("unsupported url format (%q) for connection check", rawURL)
		}
	}

	return hostname, nil
}

// checkConnection parses the rawURL and makes a best effort attempt to obtain a hostname. It then
// performs an IP lookup on that hostname and returns a an error on failure.
//
// At the moment this function is only limited to doing IP lookups. We may want/have to expand this
// to support other code hosts or to add more checks (for example making a test API call to verify
// the authorization, etc).
func checkConnection(rawURL string) error {
	if err := dnsLookup(rawURL); err != nil {
		return errors.Wrap(err, "DNS lookup failed")
	}

	if err := ping(rawURL); err != nil {
		return errors.Wrap(err, "ping failed")
	}

	return nil
}

func dnsLookup(rawURL string) error {
	hostname, err := getHostname(rawURL)
	if err != nil {
		return errors.Wrap(err, "getHostname failed")
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return errors.Newf("no IP addresses found for hostname %q", hostname)
	}

	return nil
}

// ping attempts to connect to the given rawURL. Technically it is not exactly a ping request in the
// UNIX sense since it uses TCP instead of ICMP. But we use the name to signifiy the intent here,
// which is to check if we can connect to the URL.
func ping(rawURL string) error {
	hostname, err := getHostname(rawURL)
	if err != nil {
		return errors.Wrap(err, "getHostname failed")
	}

	baseURL := rawURL
	var protocol string
	if strings.Contains(rawURL, "://") {
		parts := strings.Split(rawURL, "://")
		// Technically we can never have this condition because of the getHostname check above which
		// should detect any malformed URLs. But if I've learnt anything out of implementing the
		// methods in this file, is that URL parsing in itself is very brittle so I don't want to
		// take any chances with a panic here.
		if len(parts) < 2 {
			return errors.Newf("potentially malformed URL: %q", rawURL)
		}

		protocol, baseURL = parts[0], parts[1]
	}

	// Check if the URL includes a port.
	_, port, err := net.SplitHostPort(baseURL)
	if err != nil {
		switch protocol {
		// Assume HTTP if URL has no protocol or port.
		case "", "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return errors.Wrap(err, "failed to get port for URL which is likely a non HTTP based URL")
		}
	}

	address := net.JoinHostPort(hostname, port)
	_, err = net.DialTimeout("tcp", address, 2*time.Second)
	return err
}
