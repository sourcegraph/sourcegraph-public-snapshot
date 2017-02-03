package httpapi

import (
	"net"
	"net/http"
	"net/url"
	"os"

	// Import for side effect of setting SGPATH env var.
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	zapServerURL = os.ExpandEnv(env.Get("ZAP_SERVER", "ws://${SGPATH}/zap", "zap server URL (ws:///abspath or ws://host:port)"))
)

func serveZap(w http.ResponseWriter, r *http.Request) {
	span, _ := opentracing.StartSpanFromContext(r.Context(), "Zap session")
	defer span.Finish()

	u, _ := url.Parse(zapServerURL)
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
	proxy.ServeHTTP(getHijacker(w), r)
}
