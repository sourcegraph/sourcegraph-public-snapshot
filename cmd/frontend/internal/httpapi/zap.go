package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	// Import for side effect of setting SGPATH env var.
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"

	"github.com/gorilla/websocket"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/sourcegraph/websocketproxy"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/pkg/sgutil"
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
			// sourcegraph.com too (when Host=ws.sourcegraph.com but Origin=sourcegraph.com).
			origin := r.Header["Origin"]
			if len(origin) == 0 {
				return true
			}
			u, err := url.Parse(origin[0])
			if err != nil {
				return false
			}
			return u.Host == r.Host || u.Host == "sourcegraph.com"
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

	// Assign a custom proxy copy function. When called, we have the client
	// websocket connection, but not yet a proxy backend connection (which is
	// important, because we don't yet have an actor in the context which must
	// be sent as an HTTP header to the proxy backend).
	proxy.Copy = func(ioClient *websocket.Conn, ioBackend func() (*websocket.Conn, error)) error {
		backendReady := make(chan bool, 1)
		var backend, client *jsonrpc2.Conn
		backendHandler := jsonrpc2HandlerFunc(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
			// Any requests from the backend are simply round tripped to the client.
			roundTrip(ctx, backend, client, req)
		})

		// Handle jsonrpc2 requests from our client.
		clientHandler := jsonrpc2HandlerFunc(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
			// First, we look for the initialize request. This has the information
			// needed to authenticate the user context in the zap.InitializeParams.InitializationOptions.
			if req.Method == "initialize" {
				if req.Params == nil {
					conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams})
					return
				}
				var params zap.InitializeParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					e := &jsonrpc2.Error{Code: jsonrpc2.CodeParseError}
					e.SetError(err)
					conn.ReplyWithError(ctx, req.ID, e)
					return
				}

				// 🚨 SECURITY: Authenticate via the session cookie if present in 🚨
				// params.InitializationOptions.
				if auth := sgutil.AuthFromInitializationOptions(params.InitializationOptions); auth != "" {
					ctx = session.AuthenticateBySession(ctx, strings.TrimPrefix(auth, "session "))
				}

				// 🚨 SECURITY: Pass through the actor by overwriting the X-Actor HTTP header. 🚨
				//
				// DO NOT remove this or allow the user to specify an X-Actor header in any
				// way past this point.
				proxy.Director = func(incoming *http.Request, out http.Header) {
					out.Set(actor.HeaderKey, incoming.Header.Get(actor.HeaderKey))
				}
				actor.SetTrustedHeader(ctx, r.Header)

				// Now that we're authenticated and have the X-Actor HTTP header present,
				// connect to the proxy backend. This is the Zap server, which just takes
				// our X-Actor at face value / does no authentication.
				backendConn, err := ioBackend()
				if err != nil {
					e := &jsonrpc2.Error{Code: jsonrpc2.CodeInternalError}
					e.SetError(err)
					conn.ReplyWithError(ctx, req.ID, e)
					return
				}
				backend = jsonrpc2.NewConn(r.Context(), websocketjsonrpc2.NewObjectStream(backendConn), backendHandler)
				backendReady <- true
			}

			// Roundtrip the request to the backend assuming we've initialized.
			if backend != nil {
				roundTrip(ctx, client, backend, req)
			}
		})
		client = jsonrpc2.NewConn(r.Context(), websocketjsonrpc2.NewObjectStream(ioClient), clientHandler)

		select {
		case <-client.DisconnectNotify():
			if backend != nil {
				backend.Close()
			}
		case <-backendReady:
			select {
			case <-client.DisconnectNotify():
				backend.Close()
			case <-backend.DisconnectNotify():
				client.Close()
			}
		}
		return nil
	}

	// Forward to zap server.
	proxy.ServeHTTP(w, r)
}

// roundTrip performs a roundtrip of a Zap request. The request is made to the
// specified 'to' connection and the response is sent to the 'from' connection.
func roundTrip(ctx context.Context, from, to *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Notif {
		to.Notify(ctx, req.Method, req.Params)
		return
	}

	callOpts := []jsonrpc2.CallOption{
		// Proxy the ID used. Otherwise we assign our own ID, breaking
		// calls that depend on controlling the ID.
		jsonrpc2.PickID(req.ID),
	}

	var result json.RawMessage
	err := to.Call(ctx, req.Method, req.Params, &result, callOpts...)
	if err != nil {
		if _, userError := err.(*jsonrpc2.Error); !userError {
			log.Println("Zap: send req error", err.Error())
		}

		var respErr *jsonrpc2.Error
		if e, ok := err.(*jsonrpc2.Error); ok {
			respErr = e
		} else {
			respErr = &jsonrpc2.Error{Message: err.Error()}
		}
		from.ReplyWithError(ctx, req.ID, respErr)
	}
	from.Reply(ctx, req.ID, &result)
}
