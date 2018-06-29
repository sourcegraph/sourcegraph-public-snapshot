package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	libhoney "github.com/honeycombio/libhoney-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/honey"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/xlang"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

var websocketUpgrader = websocket.Upgrader{}

func init() {
	if v, _ := strconv.ParseBool(os.Getenv("WEBSOCKET_INSECURE_SKIP_ORIGIN_CHECK")); v {
		websocketUpgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
}

func serveLSP(w http.ResponseWriter, r *http.Request) {
	if conf.Platform() == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var err error
	ctx := r.Context()

	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP session")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	builder := honey.Builder("lsp")
	builder.AddField("client", "ws")
	builder.AddField("user_agent", r.UserAgent())
	if actor := actor.FromContext(ctx); actor != nil {
		builder.AddField("uid", actor.UID)
	}

	// Connect to the backend server (proxy target) before upgrading to WebSocket connection. This
	// is so we can return BadGateway if the backend server is down. Otherwise we would have to do
	// some sort of custom response on the WebSocket connection.
	dialCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	serverConn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", xlang.DefaultAddr)
	if err != nil {
		proxyFailed.WithLabelValues("dial").Inc()
		http.Error(w, "Connecting to LSP server failed.", http.StatusBadGateway)
		return
	}

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		proxyFailed.WithLabelValues("upgrade").Inc()
		// We do not need to do http.Error, since the websocketUpgrader will do it for us.
		return
	}

	// Some clients (such as web browsers) close the WebSocket for a lot of unexpected reasons. To
	// better understand these, we instrument how a connection is closed by the peer. We also
	// normalisze it to prevent noisy logs from the jsonrpc2 package.
	closeHandler := conn.CloseHandler()
	conn.SetCloseHandler(func(code int, text string) error {
		// Instrument why we closed
		span.SetTag("ws.closeerror.code", code)
		span.SetTag("ws.closeerror.text", text)
		wsCloseError.WithLabelValues(strconv.Itoa(code)).Inc()

		// Call the wrapped default close handler.
		err := closeHandler(code, text)
		if err != nil {
			return err
		}

		// jsonrpc2 "unwraps" this as an io.ErrUnexpectedEOF. We want to treat all
		// websocket.CloseErrors as unexpected.
		return &websocket.CloseError{
			Code: websocket.CloseAbnormalClosure,
			Text: io.ErrUnexpectedEOF.Error(),
		}
	})

	proxy := &jsonrpc2Proxy{
		httpCtx: ctx,
		builder: builder,
		ready:   make(chan struct{}),
	}

	proxy.client = jsonrpc2.NewConn(ctx, websocketjsonrpc2.NewObjectStream(conn), jsonrpc2HandlerFunc(proxy.handleClientRequest))
	proxy.server = &xclient{Client: &xlang.Client{
		Conn: jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(serverConn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(jsonrpc2HandlerFunc(proxy.handleServerRequest))),
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

// jsonrpc2Proxy is a proxy between a client WebSocket JSON-RPC 2.0 connection (typically with a
// user's Web browser) and a server raw-TCP JSON-RPC 2.0 connection. It is necessary because we
// don't want clients to connect directly to our LSP proxy, because it does not perform
// authentication/authorization checks (and, less important, it does not currently accept WebSocket
// connections, although it could be easily enhanced to do so).
type jsonrpc2Proxy struct {
	httpCtx    context.Context
	client     *jsonrpc2.Conn // connection to the HTTP client (e.g., the user's browser)
	server     *xclient       // connection to LSP proxy; use a wrapped xlang.Client since it injects OpenTracing metadata
	initParams *trackedInitParams
	builder    *libhoney.Builder
	ready      chan struct{}
}

func (p *jsonrpc2Proxy) start() {
	close(p.ready)
}

// jsonrpc2FromConn defines the subset of jsonrpc2.Conn we use to pass on a response.
type jsonrpc2FromConn interface {
	ReplyWithError(context.Context, jsonrpc2.ID, *jsonrpc2.Error) error
	Reply(context.Context, jsonrpc2.ID, interface{}) error
}

func (p *jsonrpc2Proxy) roundTrip(ctx context.Context, from jsonrpc2FromConn, to jsonrpc2.JSONRPC2, req *jsonrpc2.Request) error {
	if req.Notif {
		if err := to.Notify(ctx, req.Method, req.Params); err != nil {
			if _, userError := err.(*jsonrpc2.Error); !userError {
				proxyFailed.WithLabelValues("send-notif").Inc()
			}
		}
		return nil
	}

	callOpts := []jsonrpc2.CallOption{
		// Proxy the ID used. Otherwise we assign our own ID, breaking calls that depend on
		// controlling the ID such as $/cancelRequest and $/partialResult.
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
			// "from" being closed is common in this code path. The reason is "to" can be closed
			// during an inflight request, causing "from" to be closed via DisconnectNotify. So we
			// only increment the counter for the uncommon case.
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

	// This assumes that the LSP proxy only allows access to the repository specified in the
	// "initialize" request. (I.e., you can't send an "initialize" with a repository of
	// "github.com/foo/bar" then a "textDocument/hover" for a file in "github.com/qux/other".)
	if req.Method == "initialize" {
		initParams, accessErr := authorizeInitialize(p.httpCtx, req)
		if accessErr != nil {
			p.initParams = nil
			proxyFailed.WithLabelValues("auth").Inc()
			if err := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Message: accessErr.Error()}); err != nil {
				proxyFailed.WithLabelValues("reply-auth").Inc()
			}
			return // 🚨 SECURITY: Do not pass on unauthorized request to server. 🚨
		}
		p.initParams = initParams
	}
	initParams := p.initParams
	if initParams == nil {
		if !req.Notif {
			if err := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Message: "connection not initialized"}); err != nil {
				proxyFailed.WithLabelValues("reply-init").Inc()
			}
		}
		return
	}

	// We don't need to measure notifications, so just send straight away.
	if req.Notif {
		go p.roundTrip(ctx, conn, p.server, req)
		return
	}

	go func() {
		// No parent, otherwise request is associated with all spans in session. This makes it hard
		// to understand the UI.
		span := opentracing.GlobalTracer().StartSpan("LSP WebSocket Proxy")
		ext.Component.Set(span, "httpapi")
		ext.SpanKindRPCClient.Set(span)
		span.SetTag("jsonrpc2.id", req.ID.String())
		span.SetTag("jsonrpc2.method", req.Method)
		if req.Params != nil {
			span.LogFields(otlog.Object("jsonrpc2.params", string(*req.Params)))
		}
		ctx = opentracing.ContextWithSpan(ctx, span)
		meta := map[string]string{
			"X-Trace": trace.SpanURL(span),
		}

		err := p.roundTrip(ctx, replyWithMeta{conn, meta}, p.server, req)

		duration := time.Since(start)
		success := strconv.FormatBool(err == nil)

		// Update span now that we have initParams.
		if u := initParams.rootURI; u != nil {
			span.SetTag("rooturi", u.String())
			span.SetTag("repo", u.Repo())
			span.SetTag("commit", u.Rev())
		}
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()

		labels := prometheus.Labels{
			"success":   success,
			"method":    req.Method,
			"mode":      initParams.mode,
			"transport": "ws",
		}
		xlangRequestDuration.With(labels).Observe(duration.Seconds())

		if honey.Enabled() {
			ev := p.builder.NewEvent()
			ev.AddField("success", success)
			ev.AddField("method", req.Method)
			ev.AddField("mode", initParams.mode)
			ev.AddField("duration_ms", duration.Seconds()*1000)
			if initParams.rootURI != nil {
				addRootURIFields(ev, initParams.rootURI)
			}
			if err != nil {
				ev.AddField("error", err.Error())
			}
			ev.Send()
		}
	}()
}

func (p *jsonrpc2Proxy) handleServerRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	<-p.ready
	p.roundTrip(ctx, conn, p.client, req)
}

func authorizeInitialize(ctx context.Context, req *jsonrpc2.Request) (*trackedInitParams, error) {
	if req.Method != "initialize" {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lspext.ClientProxyInitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Ensure that the client has set a session value. LSP sessions with different
	// session keys are isolated from each other, even if they are for the same
	// repository/commit/etc. This WebSocket LSP API, as opposed to the simple HTTP POST interface,
	// allows clients to send arbitrary LSP requests, including ones that cause mutation (such as
	// textDocument/didChange). For example, a malicious client could send a textDocument/didChange
	// saying that a popular library's documentation comment for a common function recommends that
	// users run `rm -rf`, and the language server would report that in the textDocument/hover
	// response to all users.
	//
	// We must prevent those mutations from affecting readonly clients using the simple HTTP POST
	// interface. These clients all share a session because the session value is empty. This is OK
	// because the simple HTTP POST interface does not allow clients to mutate state.
	//
	// Note that any other clients who know the session value can join or hijack this session. It is
	// the client's responsibility to generate a random session value and keep it secret.
	if params.InitializationOptions.Session == "" {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "initializationOptions.session must be set",
		}
	}

	rootURI, err := uri.Parse(string(params.RootURI))
	if err != nil {
		return nil, err
	}

	t := &trackedInitParams{
		mode:    params.InitializationOptions.Mode,
		rootURI: rootURI,
	}

	return t, nil
}

// trackedInitParams stores fields we want to track in our instrumentation for
// each request. The fields are extracted from the initialize request.
type trackedInitParams struct {
	mode    string
	rootURI *uri.URI
}

type replyWithMeta struct {
	conn *jsonrpc2.Conn
	meta interface{}
}

func (r replyWithMeta) ReplyWithError(ctx context.Context, id jsonrpc2.ID, respErr *jsonrpc2.Error) error {
	meta, err := r.marshalMeta()
	if err != nil {
		return err
	}
	return r.conn.SendResponse(ctx, &jsonrpc2.Response{ID: id, Meta: meta, Error: respErr})
}

func (r replyWithMeta) Reply(ctx context.Context, id jsonrpc2.ID, result interface{}) error {
	meta, err := r.marshalMeta()
	if err != nil {
		return err
	}
	resp := &jsonrpc2.Response{ID: id, Meta: meta}
	if err := resp.SetResult(result); err != nil {
		return err
	}
	return r.conn.SendResponse(ctx, resp)
}

func (r replyWithMeta) marshalMeta() (*json.RawMessage, error) {
	b, err := json.Marshal(r.meta)
	if err != nil {
		return nil, err
	}
	return (*json.RawMessage)(&b), nil
}

var (
	proxyFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "lsp_websocket_proxy_failed",
		Help:      "Total number of failures in the LSP WebSocket.",
	}, []string{"why"})
	wsCloseError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "lsp_websocket_close_error",
		Help:      "Total number of LSP WebSocket close errors received.",
	}, []string{"code"})
)

func init() {
	prometheus.MustRegister(proxyFailed)
	prometheus.MustRegister(wsCloseError)
}
