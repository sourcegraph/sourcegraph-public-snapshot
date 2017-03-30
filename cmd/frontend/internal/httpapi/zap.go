package httpapi

import (
	"net"
	"net/http"
	"net/url"

	// Import for side effect of setting SGPATH env var.
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"

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

	// SECURITY: Pass through the actor by overwriting the X-Actor HTTP header.
	//
	// DO NOT remove this or allow the user to specify an X-Actor header in any
	// way past this point.
	proxy.Director = func(incoming *http.Request, out http.Header) {
		out.Set(auth.ActorHeaderKey, incoming.Header.Get(auth.ActorHeaderKey))
	}
	auth.SetActorTrustedHeader(r.Context(), r.Header)

	// Forward to zap server.
	proxy.ServeHTTP(w, r)
}
