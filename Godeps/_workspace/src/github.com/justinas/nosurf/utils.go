package nosurf

import (
	"net/url"
)

func sContains(slice []string, s string) bool {
	// checks if the given slice contains the given string
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// Checks if the given URLs have the same origin
// (that is, they share the host, the port and the scheme)
func sameOrigin(u1, u2 *url.URL) bool {
	// we take pointers, as url.Parse() returns a pointer
	// and http.Request.URL is a pointer as well

	// Host is either host or host:port
	return (u1.Scheme == u2.Scheme && u1.Host == u2.Host)
}
