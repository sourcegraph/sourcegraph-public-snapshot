package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

// We need to multiplex an entire xlang connection pool on an HTTP
// endpoint. Clients obtain a "session key" after the "initialize"
// request. Subsequent requests to the xlang HTTP endpoint with the
// same session key will reuse the same connection.

type xlangClient interface {
	Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error
	Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error
	Close() error
}

var xlangCreateConnection = func() (xlangClient, error) {
	addr := os.Getenv("LSP_PROXY")
	if addr == "" {
		return nil, errors.New("no LSP_PROXY env var set (need address to LSP proxy)")
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return xlang.DialProxy(dialCtx, addr, nil)
}

func serveXLang(w http.ResponseWriter, r *http.Request) error {
	if os.Getenv("DISABLE_XLANG_HTTP_GATEWAY") != "" {
		// Escape valve in case it causes production issues and we
		// need to quickly disable it.
		return &errcode.HTTPErr{Status: http.StatusGatewayTimeout, Err: errors.New("xlang http gateway is disabled")}
	}

	// Decode this early so we can print more useful log messages.
	var reqs []jsonrpc2.Request
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		return err
	}

	// We only ever need to send 4 (initialize, the actual method,
	// shutdown, exit) on the frontend, so be strict here for now.
	if len(reqs) != 4 {
		return fmt.Errorf("got %d jsonrpc2 requests, want exactly 4", len(reqs))
	}

	// Check consistency against the URL. The URL route params are for
	// ease of debugging only, but it'd be confusing if they could
	// diverge from the actual jsonrpc2 request.
	if v := mux.Vars(r); v["LSPMethod"] != reqs[1].Method {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: fmt.Errorf("LSP method param in URL %q != %q method in LSP message params", v["LSPMethod"], reqs[1].Method)}
	}

	// Use a one-shot connection to the LSP proxy. This is cheap,
	// since the LSP proxy will reuse an already running server for
	// the given workspace if available.
	c, err := xlangCreateConnection()
	if err != nil {
		return err
	}
	defer c.Close()

	// Inject tracing info.
	opName := "LSP HTTP gateway: " + reqs[1].Method
	span, ctx := opentracing.StartSpanFromContext(r.Context(), opName)
	defer span.Finish()
	carrier := opentracing.TextMapCarrier{}
	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		return err
	}
	addMeta := jsonrpc2.Meta(carrier)

	// Only return the last response to the HTTP client (e.g., just
	// return the result of "textDocument/definition" if "initialize"
	// is followed by a "textDocument/definition") because that's all
	// the client needs.
	resps := make([]*jsonrpc2.Response, len(reqs))
	for i, req := range reqs {
		if req.Notif {
			if err := c.Notify(ctx, req.Method, req.Params, addMeta); err != nil {
				return err
			}
		} else {
			resps[i] = &jsonrpc2.Response{}
			err := c.Call(ctx, req.Method, req.Params, &resps[i].Result, addMeta)
			if err != nil {
				return err
			}
		}
	}
	return writeJSON(w, resps)
}
