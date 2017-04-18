package httpapi

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	// Import for side effect of setting SGPATH env var.
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	opentracing "github.com/opentracing/opentracing-go"
)

func serveZap(w http.ResponseWriter, r *http.Request) {
	// Reject WebSocket connections that are to sourcegraph.com instead of
	// ws.sourcegraph.com (the canonical URL). Because connections made to
	// sourcegraph.com directly are subject to the CloudFlare proxy, and are
	// inherently rate limited AND subject to an uncontrollable timeout every
	// 100s.
	//
	// Due to the WebSocket spec not requiring clients to follow redirects,
	// i.e. because RFC6455 (https://tools.ietf.org/html/rfc6455) states:
	//
	// 	1.  If the status code received from the server is not 101, the
	// 	    client handles the response per HTTP [RFC2616] procedures.  In
	// 	    particular, the client might perform authentication if it
	// 	    receives a 401 status code; the server might redirect the client
	// 	    using a 3xx status code (but clients are not required to follow
	// 	    them), etc.  Otherwise, proceed as follows.
	//
	// And because of the fact that the clients we care about do not follow
	// redirects (the Go websocket client we use in the zap binary AND
	// WebSocket connections made by Google Chrome itself), we cannot use a
	// redirect to achieve this.
	//
	// Note: We use == below instead of != to consider the case of e.g. local
	// deployments or other non-sourcegraph.com deployments.
	if r.Host == "sourcegraph.com" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: WebSocket connections must be made to ws.sourcegraph.com (not %q)\n", r.Host)
		return
	}

	span, _ := opentracing.StartSpanFromContext(r.Context(), "Zap session")
	defer span.Finish()

	u, _ := url.Parse(backend.ZapServerURL)
	proxy := websocketproxy.NewProxy(u)
	proxy.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024, // default used by websocketproxy
		WriteBufferSize: 1024, // default used by websocketproxy
		CheckOrigin: func(r *http.Request) bool {
			// Same as the default used by websocketproxy, except it allows
			// ws.sourcegraph.com too.
			origin := r.Header["Origin"]
			if len(origin) == 0 {
				return true
			}
			u, err := url.Parse(origin[0])
			if err != nil {
				return false
			}
			return u.Host == r.Host || u.Host == "ws.sourcegraph.com"
		},
	}
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
		out.Set(actor.HeaderKey, incoming.Header.Get(actor.HeaderKey))
	}
	actor.SetTrustedHeader(r.Context(), r.Header)

	// Forward to zap server.
	proxy.ServeHTTP(w, r)
}
