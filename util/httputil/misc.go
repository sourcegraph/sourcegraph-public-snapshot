package httputil

import "net"

// StripPort removes the port specification from an address.
func StripPort(s string) string {
	if h, _, err := net.SplitHostPort(s); err == nil {
		s = h
	}
	return s
}
