package httpserver

import (
	"net"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

// NewListener returns a TCP listener accepting connections
// on the given address.
func NewListener(addr string) (_ net.Listener, err error) {
	addr, err = SanitizeAddr(addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return listener, err
}

// SanitizeAddr replaces the host in the given address with
// 127.0.0.1 if no host is supplied or if running in insecure
// dev mode.
func SanitizeAddr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}

	if host == "" && env.InsecureDev {
		host = "127.0.0.1"
	}

	return net.JoinHostPort(host, port), nil
}
