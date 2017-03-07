package httpapi

import (
	"net"
	"net/http"
	"net/url"

	// Import for side effect of setting SGPATH env var.
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	opentracing "github.com/opentracing/opentracing-go"
)

func serveZap(w http.ResponseWriter, r *http.Request) {
	span, _ := opentracing.StartSpanFromContext(r.Context(), "Zap session")
	defer span.Finish()

	u, _ := url.Parse(backend.ZapServerURL)
	proxy := websocketproxy.NewProxy(u)
	d := *websocket.DefaultDialer
	d.NetDial = func(network, addr string) (net.Conn, error) {
		if u.Host == "" {
			network = "unix"
			addr = u.Path
		}
		return net.Dial(network, addr)
	}
	proxy.Dialer = &d

	// Forward to zap server.
	proxy.ServeHTTP(w, r)
}
