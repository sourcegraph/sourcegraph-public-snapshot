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
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
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

	builder := honey.Builder("xlang")
	builder.AddField("client", "ws")
	builder.AddField("user_agent", r.UserAgent())
	if actor := actor.FromContext(ctx); actor != nil {
		builder.AddField("uid", actor.UID)
		builder.AddField("login", actor.Login)
		builder.AddField("email", actor.Email)
	}

	// Connect to server before upgrading to websocket connection. This is
	// so we can return BadGateway if our gateway is down. Otherwise we
	// would have to do some sort of custom response on the websocket
	// connection.
	serverNetConn, err := dialLSPProxy(ctx)
	if err != nil {
		proxyFailed.WithLabelValues("dial").Inc()
		http.Error(w, "connecting to LSP server failed", http.StatusBadGateway)
		return
	}

	// 🚨 SECURITY: This endpoint relies on cookie based authentication. 🚨
	// websocketUpgrader.Upgrade handles checking the origin header so
	// that this endpoint is not vulnerable to CSRF like attacks.
	//
	// You can read more about this security issue here:
	// https://www.christian-schneider.net/CrossSiteWebSocketHijacking.html
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		proxyFailed.WithLabelValues("upgrade").Inc()
		// We do not need to do http.Error, since the
		// websocketUpgrader will do it for us.
		return
	}

	// Our peer is the user in the browser. As such we can get lots of
	// interesting close responses from them. We instrument how a
	// connection is closed by the peer. We also normalise it to prevent
	// noise in the jsonrpc2 pkg.
	closeHandler := conn.CloseHandler()
	conn.SetCloseHandler(func(code int, text string) error {
		// Instrument why we closed
		span.SetTag("ws.closeerror.code", code)
		span.SetTag("ws.closeerror.text", text)
		wsCloseError.WithLabelValues(strconv.Itoa(code)).Inc()

		// Call the wrapped default close handler
		err := closeHandler(code, text)
		if err != nil {
			return err
		}

		// jsonrpc2 "unwraps" this as an io.ErrUnexpectedEOF. We want
		// to treat all websocket.CloseErrors as unexpected.
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
	httpCtx    context.Context
	client     *jsonrpc2.Conn // connection to the browser
	server     *xclient       // connection to lsp proxy. We use a wrapped xlang.Client since it injects opentracing metadata
	initParams *trackedInitParams
	builder    *libhoney.Builder
	ready      chan struct{}
}

func (p *jsonrpc2Proxy) start() {
	close(p.ready)
}

// jsonrpc2FromConn defines the subset of jsonrpc2.Conn we use to pass on a
// response.
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
		// No parent, otherwise request is associated with all spans in
		// session. This makes it hard to understand the UI.
		span := opentracing.GlobalTracer().StartSpan("LSP WebSocket Proxy")
		ext.Component.Set(span, "httpapi")
		ext.SpanKindRPCClient.Set(span)
		span.SetTag("jsonrpc2.id", req.ID.String())
		span.SetTag("jsonrpc2.method", req.Method)
		ctx = opentracing.ContextWithSpan(ctx, span)
		meta := map[string]string{
			"X-Trace": traceutil.SpanURL(span),
		}

		err := p.roundTrip(ctx, replyWithMeta{conn, meta}, p.server, req)

		duration := time.Since(start)
		success := strconv.FormatBool(err == nil)

		// Update span now that we have initParams.
		if u := initParams.rootPathURI; u != nil {
			span.SetTag("rootpath", u.String())
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
			if initParams.rootPathURI != nil {
				addRootPathFields(ev, initParams.rootPathURI)
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

	var params struct {
		lspext.ClientProxyInitializeParams
		RootURI string `json:"rootUri"`
	}
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: LSP recently introduced a rootUri field on 🚨
	// InitializeParams and deprecated rootPath. Until we
	// support rootUri in the LSP proxy, unset rootUri so we
	// guarantee only rootPath can specify the workspace.
	params.RootURI = ""
	if err := req.SetParams(params); err != nil {
		return nil, err
	}

	rootPathURI, err := uri.Parse(params.RootPath)
	if err != nil {
		return nil, err
	}

	t := &trackedInitParams{
		mode:        params.InitializationOptions.Mode,
		rootPathURI: rootPathURI,
	}

	// 🚨 SECURITY: Check that the the user can access the repo. 🚨
	if _, err := backend.Repos.GetByURI(ctx, rootPathURI.Repo()); err != nil {
		return t, err
	}
	return t, nil
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

// trackedInitParams stores fields we want to track in our instrumentation for
// each request. The fields are extracted from the initialize request.
type trackedInitParams struct {
	mode        string
	rootPathURI *uri.URI
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
		Name:      "websocket_proxy_failed",
		Help:      "Total number of failures in the xlang websocket gateway.",
	}, []string{"why"})
	wsCloseError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "websocket_close_error",
		Help:      "Total number of websocket close errors received.",
	}, []string{"code"})
)

func init() {
	prometheus.MustRegister(proxyFailed)
	prometheus.MustRegister(wsCloseError)
}
