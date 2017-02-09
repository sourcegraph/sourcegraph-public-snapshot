package httpapi

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	libhoney "github.com/honeycombio/libhoney-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"go.uber.org/atomic"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
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
		log15.Error("serveLSP: Upgrade to WebSocket failed.", "err", err)
		http.Error(w, "upgrade to WebSocket failed", http.StatusInternalServerError)
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
		log15.Info("LSP: user connected", "login", actor.Login, "uid", actor.UID)
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

	proxy.client = jsonrpc2.NewConn(ctx, websocketjsonrpc2.NewObjectStream(conn), jsonrpc2HandlerFunc(proxy.handleClientRequest))

	addr := os.Getenv("LSP_PROXY")
	if addr == "" {
		log15.Error("serveLSP: no LSP_PROXY env var set (need address to LSP proxy)")
		http.Error(w, "locating LSP server failed", http.StatusBadGateway)
		return
	}
	serverNetConn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		log15.Error("serveLSP: dialing LSP server failed", "addr", addr, "err", err)
		http.Error(w, "connecting to LSP server failed", http.StatusBadGateway)
		return
	}
	proxy.server = jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(serverNetConn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2HandlerFunc(proxy.handleServerRequest))

	proxy.start()

	select {
	case <-proxy.client.DisconnectNotify():
		proxy.server.Close()
	case <-proxy.server.DisconnectNotify():
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
	httpCtx        context.Context
	client, server *jsonrpc2.Conn //  connection to the client (typically a browser) and server (LSP proxy)
	mode           *atomic.String
	builder        *libhoney.Builder
	ready          chan struct{}
}

func (p *jsonrpc2Proxy) start() {
	close(p.ready)
}

func (p *jsonrpc2Proxy) roundTrip(ctx context.Context, from, to *jsonrpc2.Conn, req *jsonrpc2.Request) error {
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
			if _, err := backend.Repos.Resolve(p.httpCtx, &sourcegraph.RepoResolveOp{Path: rootPathURI.Repo()}); err != nil {
				log15.Error("jsonrpc2Proxy: access check failed", "workspace", params.RootPath, "err", err)
				return err
			}
			return nil
		}
		if accessErr := checkAccess(); accessErr != nil {
			if err := from.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Message: accessErr.Error()}); err != nil {
				log15.Error("jsonrpc2Proxy: error sending access-check-failed reply", "method", req.Method, "accessErr", accessErr, "err", err)
			}
			return accessErr // 🚨 SECURITY: Do not pass on unauthorized request to server. 🚨
		}
	}

	if req.Notif {
		if err := to.Notify(ctx, req.Method, req.Params); err != nil {
			log15.Error("jsonrpc2Proxy: error sending notification", "method", req.Method)
		}
		return nil
	}

	var result json.RawMessage
	if err := to.Call(ctx, req.Method, req.Params, &result); err != nil {
		log15.Error("jsonrpc2Proxy: error sending request", "method", req.Method, "err", err)
		var respErr *jsonrpc2.Error
		if e, ok := err.(*jsonrpc2.Error); ok {
			respErr = e
		} else {
			respErr = &jsonrpc2.Error{Message: err.Error()}
		}
		if err := from.ReplyWithError(ctx, req.ID, respErr); err != nil {
			log15.Error("jsonrpc2Proxy: error sending error reply", "method", req.Method, "payload", respErr, "err", err)
		}
		return respErr
	}
	if err := from.Reply(ctx, req.ID, &result); err != nil {
		log15.Error("jsonrpc2Proxy: error sending reply", "method", req.Method, "err", err)
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
