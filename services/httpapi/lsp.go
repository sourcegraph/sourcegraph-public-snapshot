package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	libhoney "github.com/honeycombio/libhoney-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"go.uber.org/atomic"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func serveLSP(w http.ResponseWriter, r *http.Request) {
	// 🚨 SECURITY: This endpoint relies on cookie based authentication. 🚨
	// websocketUpgrader.Upgrade handles checking the origin header so
	// that this endpoint is not vulnerable to CSRF like attacks.
	//
	// You can read more about this security issue here:
	// https://www.christian-schneider.net/CrossSiteWebSocketHijacking.html
	conn, err := websocketUpgrader.Upgrade(getHijacker(w), r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %s", err)
		proxyFailed.WithLabelValues("upgrade").Inc()
		// HTTP response has already been written by Upgrade
		return
	}

	ctx := r.Context()

	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP session")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	builder := honey.Builder("xlang")
	builder.AddField("client", "ws")
	builder.AddField("user_agent", r.UserAgent())
	if actor := auth.ActorFromContext(ctx); actor != nil {
		builder.AddField("uid", actor.UID)
		builder.AddField("login", actor.Login)
		builder.AddField("email", actor.Email)
	}

	proxy := &jsonrpc2Proxy{
		httpCtx: ctx,
		mode:    atomic.NewString(""),
		builder: builder,
		ready:   make(chan struct{}),
	}

	proxy.client = jsonrpc2.NewConn(ctx, websocketjsonrpc2.NewObjectStream(conn), jsonrpc2.AsyncHandler(jsonrpc2HandlerFunc(proxy.handleClientRequest)))

	serverNetConn, err := dialLSPProxy(ctx)
	if err != nil {
		log.Printf("connecting to LSP server failed: %s", err)
		proxyFailed.WithLabelValues("dial").Inc()
		// HTTP response has already been written by Upgrade
		return
	}
	proxy.server = &xclient{Client: &xlang.Client{
		Conn: jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(serverNetConn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(jsonrpc2HandlerFunc(proxy.handleServerRequest))),
	}}

	proxy.start()

	select {
	case <-proxy.client.DisconnectNotify():
		proxy.server.Close()
	case <-proxy.server.Conn.DisconnectNotify():
		proxy.client.Close()
	}
}

type jsonrpc2HandlerFunc func(context.Context, *jsonrpc2.Conn, *jsonrpc2.Request)

func (h jsonrpc2HandlerFunc) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h(ctx, conn, req)
}

// jsonrpc2Proxy is a proxy between a client WebSocket JSON-RPC 2.0
// connection (typically with a user's Web browser) and a server
// raw-TCP JSON-RPC 2.0 connection. It is necessary because we don't
// want clients to connect directly to our LSP proxy, because it does
// not perform authentication/authorization checks (and, less
// important, it does not currently accept WebSocket connections,
// although it could be easily enhanced to do so).
//
// 🚨 SECURITY: The jsonrpc2Proxy checks that the repository specified in 🚨
// the "initialize" request can be accessed by the current user. If
// the current user is forbidden, it immediately ends the connection
// and does not allow the client to send any messages to the LSP
// proxy. If the current user is permitted, it assumes the LSP proxy
// will enforce that no other repositories' files (other than the
// single repository initially specified in the "initialize" request)
// will be accessed in the current LSP session.
type jsonrpc2Proxy struct {
	httpCtx context.Context
	client  *jsonrpc2.Conn // connection to the browser
	server  *xclient       // connection to lsp proxy. We use a wrapped xlang.Client since it injects opentracing metadata
	mode    *atomic.String
	builder *libhoney.Builder
	ready   chan struct{}
}

func (p *jsonrpc2Proxy) start() {
	close(p.ready)
}

func (p *jsonrpc2Proxy) roundTrip(ctx context.Context, from *jsonrpc2.Conn, to jsonrpc2.JSONRPC2, req *jsonrpc2.Request) error {
	// 🚨 SECURITY: If this is the "initialize" request, we MUST check 🚨
	// that the current user can access the workspace root's
	// repository. This is the ONLY PLACE that access is checked; the
	// LSP proxy does not perform any access checking.
	//
	// This assumes that the LSP proxy only allows access to the
	// repository specified in the "initialize" request. (I.e., you
	// can't send an "initialize" with a repository of
	// "github.com/foo/bar" then a "textDocument/hover" for a file in
	// "github.com/qux/other".)
	if req.Method == "initialize" {
		checkAccess := func() error {
			if req.Params == nil {
				return &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
			}

			var params struct {
				xlang.ClientProxyInitializeParams
				RootURI string `json:"rootUri"`
			}
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				return err
			}

			// 🚨 SECURITY: LSP recently introduced a rootUri field on 🚨
			// InitializeParams and deprecated rootPath. Until we
			// support rootUri in the LSP proxy, unset rootUri so we
			// guarantee only rootPath can specify the workspace.
			params.RootURI = ""
			if err := req.SetParams(params); err != nil {
				return err
			}

			p.mode.Store(params.InitializationOptions.Mode)

			rootPathURI, err := uri.Parse(params.RootPath)
			if err != nil {
				return err
			}

			// 🚨 SECURITY: Check that the the user can access the repo. 🚨
			if _, err := backend.Repos.GetByURI(p.httpCtx, rootPathURI.Repo()); err != nil {
				proxyFailed.WithLabelValues("auth").Inc()
				return err
			}
			return nil
		}
		if accessErr := checkAccess(); accessErr != nil {
			if err := from.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Message: accessErr.Error()}); err != nil {
				proxyFailed.WithLabelValues("reply-auth").Inc()
			}
			return accessErr // 🚨 SECURITY: Do not pass on unauthorized request to server. 🚨
		}
	}

	if req.Notif {
		if err := to.Notify(ctx, req.Method, req.Params); err != nil {
			if _, userError := err.(*jsonrpc2.Error); !userError {
				proxyFailed.WithLabelValues("send-notif").Inc()
			}
		}
		return nil
	}

	callOpts := []jsonrpc2.CallOption{
		// Proxy the ID used. Otherwise we assign our own ID, breaking
		// calls that depend on controlling the ID such as
		// $/cancelRequest and $/partialResult.
		jsonrpc2.PickID(req.ID),
	}

	var result json.RawMessage
	if err := to.Call(ctx, req.Method, req.Params, &result, callOpts...); err != nil {
		if _, userError := err.(*jsonrpc2.Error); !userError {
			log.Println("LSP: send req error", err.Error())
			proxyFailed.WithLabelValues("send-req").Inc()
		}

		var respErr *jsonrpc2.Error
		if e, ok := err.(*jsonrpc2.Error); ok {
			respErr = e
		} else {
			respErr = &jsonrpc2.Error{Message: err.Error()}
		}
		if err := from.ReplyWithError(ctx, req.ID, respErr); err != nil {
			// "from" being closed is common in this codepath. The
			// reason is "to" can be closed during an inflight
			// request, causing "from" to be closed via
			// DisconnectNotify. So we only increment the counter
			// for the uncommon case.
			if err != jsonrpc2.ErrClosed {
				proxyFailed.WithLabelValues("reply-req-err").Inc()
			}
		}
		return respErr
	}
	if err := from.Reply(ctx, req.ID, &result); err != nil {
		proxyFailed.WithLabelValues("reply-req").Inc()
	}
	return nil
}

func (p *jsonrpc2Proxy) handleClientRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	start := time.Now()
	<-p.ready

	err := p.roundTrip(ctx, conn, p.server, req)

	if req.Notif {
		// We don't need to measure notifications
		return
	}

	duration := time.Since(start)
	success := strconv.FormatBool(err == nil)
	mode := p.mode.Load()
	if mode == "" {
		mode = "unknown"
	}

	labels := prometheus.Labels{
		"success":   success,
		"method":    req.Method,
		"mode":      mode,
		"transport": "ws",
	}
	xlangRequestDuration.With(labels).Observe(duration.Seconds())

	if honey.Enabled() {
		ev := p.builder.NewEvent()
		ev.AddField("success", success)
		ev.AddField("method", req.Method)
		ev.AddField("mode", mode)
		ev.AddField("duration_ms", duration.Seconds()*1000)
		if err != nil {
			ev.AddField("error", err.Error())
		}
		ev.Send()
	}
}

func (p *jsonrpc2Proxy) handleServerRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	<-p.ready
	p.roundTrip(ctx, conn, p.client, req)
}

func dialLSPProxy(ctx context.Context) (net.Conn, error) {
	addr := os.Getenv("LSP_PROXY")
	if addr == "" {
		return nil, errors.New("lsp proxy not found")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
}

var proxyFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "xlang",
	Name:      "websocket_proxy_failed",
	Help:      "Total number of failures in the xlang websocket gateway.",
}, []string{"why"})

func init() {
	prometheus.MustRegister(proxyFailed)
}
