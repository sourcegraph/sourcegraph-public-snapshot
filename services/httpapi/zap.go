package httpapi

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	// Import for side effect of setting SGPATH env var.
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"

	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	zapServerURL = os.ExpandEnv(env.Get("ZAP_SERVER", "ws://${SGPATH}/zap", "zap server URL (ws:///abspath or ws://host:port)"))
)

func zapServerDial() (net.Conn, error) {
	u, err := url.Parse(zapServerURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "ws" {
		return nil, fmt.Errorf("bad dial URL %s (must be ws:///abspath or ws://host:port)", zapServerURL)
	}
	if u.Host == "" {
		return net.Dial("unix", u.Path)
	}
	return net.Dial("tcp", u.Host)
}

func serveZap(w http.ResponseWriter, r *http.Request) {
	span, _ := opentracing.StartSpanFromContext(r.Context(), "Zap session")
	defer span.Finish()

	// Forward to zap server.
	webSocketProxy(zapServerDial).ServeHTTP(w, r)
}
