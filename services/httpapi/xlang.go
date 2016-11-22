package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	libhoney "github.com/honeycombio/libhoney-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

// We need to multiplex an entire xlang connection pool on an HTTP
// endpoint. Clients obtain a "session key" after the "initialize"
// request. Subsequent requests to the xlang HTTP endpoint with the
// same session key will reuse the same connection.

var xlangRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "xlang",
	Name:      "request_duration_seconds",
	Help:      "The xlang request latencies in seconds.",
	// Buckets are similar to statsutil.UserLatencyBuckets, but with more granularity for apdex measurements.
	Buckets: []float64{0.1, 0.2, 0.5, 0.8, 1, 1.5, 2, 5, 10, 15, 20, 30},
}, []string{"success", "method", "mode"})

func init() {
	prometheus.MustRegister(xlangRequestDuration)
}

var xlangNewClient = func() (xlangClient, error) { return xlang.NewDefaultClient() }

type xlangClient interface {
	Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error
	Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error
	Close() error
}

func serveXLang(w http.ResponseWriter, r *http.Request) (err error) {
	return serveXLangMethod(r.Context(), w, r.Body)
}

// serveXLangMethod was split out from serveXLang to support the old
// hover-info and jump-to-def httpapi endpoints. Once those are gone we
// extract this back into serveXLang.
func serveXLangMethod(ctx context.Context, w http.ResponseWriter, body io.Reader) (err error) {
	start := time.Now()
	success := true
	method := "unknown"
	mode := "unknown"
	ev := honey.Event("xlang")
	defer func() {
		duration := time.Now().Sub(start)
		labels := prometheus.Labels{
			"success": fmt.Sprintf("%t", err == nil && success),
			"method":  method,
			"mode":    mode,
		}
		xlangRequestDuration.With(labels).Observe(duration.Seconds())

		if honey.Enabled() {
			ev.AddField("success", err == nil && success)
			ev.AddField("method", method)
			ev.AddField("mode", mode)
			ev.AddField("duration_ms", duration.Seconds()*1000)
			if err != nil {
				ev.AddField("error", err.Error())
			}
			if actor := auth.ActorFromContext(ctx); actor != nil {
				ev.AddField("uid", actor.UID)
				ev.AddField("login", actor.Login)
				ev.AddField("email", actor.Email)
			}
			ev.Send()
		}
	}()

	if os.Getenv("DISABLE_XLANG_HTTP_GATEWAY") != "" {
		// Escape valve in case it causes production issues and we
		// need to quickly disable it.
		return &errcode.HTTPErr{Status: http.StatusGatewayTimeout, Err: errors.New("xlang http gateway is disabled")}
	}

	// Decode this early so we can print more useful log messages.
	var reqs []jsonrpc2.Request
	if err := json.NewDecoder(body).Decode(&reqs); err != nil {
		return err
	}

	method = reqs[1].Method
	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP HTTP gateway: "+method)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Sanity check the request body. Be strict based on what we know
	// the UI sends us.
	if len(reqs) != 4 {
		return fmt.Errorf("got %d jsonrpc2 requests, want exactly 4", len(reqs))
	}
	if reqs[0].Method != "initialize" || reqs[1].Method == "initialize" || reqs[2].Method != "shutdown" || reqs[3].Method != "exit" {
		return fmt.Errorf("invalid jsonrpc2 request methods %q: expected initialize, anything but initialize, shutdown, exit (in that order)", []string{reqs[0].Method, reqs[1].Method, reqs[2].Method, reqs[3].Method})
	}
	if reqs[0].Params == nil {
		return errors.New("invalid jsonrpc2 initialize request: empty params")
	}
	var initParams xlang.ClientProxyInitializeParams
	if err := json.Unmarshal(*reqs[0].Params, &initParams); err != nil {
		return fmt.Errorf("invalid jsonrpc2 initialize params: %s", err)
	}
	if initParams.RootPath == "" {
		return errors.New("invalid empty LSP root path in initialize request")
	}
	rootPathURI, err := uri.Parse(initParams.RootPath)
	if err != nil {
		return fmt.Errorf("invalid LSP root path %q: %s", initParams.RootPath, err)
	}
	addRootPathFields(ev, rootPathURI)
	if initParams.Mode != "" {
		mode = initParams.Mode
	}

	// Check consistency against the URL. The URL route params are for
	// ease of debugging only, but it'd be confusing if they could
	// diverge from the actual jsonrpc2 request.
	if method != strings.TrimSuffix(reqs[1].Method, "?prepare") {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: fmt.Errorf("LSP method param in URL %q != %q method in LSP message params", method, reqs[1].Method)}
	}

	// Check that the user has permission to read this repo. Calling
	// Repos.Resolve will fail if the user does not have access to the
	// specified repo.
	//
	// SECURITY NOTE: The LSP client proxy DOES NOT check
	// permissions. It accesses the gitserver directly and relies on
	// its callers to check permissions.
	checkedUserHasReadAccessToRepo := false // safeguard to make sure we don't accidentally delete the check below
	{
		// SECURITY NOTE: Do not delete this block. If you delete this
		// block, anyone can access any private code, even if they are
		// not authorized to do so.
		repo := rootPathURI.Host + strings.TrimSuffix(rootPathURI.Path, ".git") // of the form "github.com/foo/bar"
		if _, err := backend.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repo}); err != nil {
			return err
		}
		checkedUserHasReadAccessToRepo = true
	}

	// Use a one-shot connection to the LSP proxy. This is cheap,
	// since the LSP proxy will reuse an already running server for
	// the given workspace if available.
	c, err := xlangNewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if !checkedUserHasReadAccessToRepo {
		// SECURITY NOTE: If we somehow got here without checking that
		// the user has read access to the repo, then there is a
		// serious issue (e.g., the permission check code above got
		// deleted). This if-statement is not what enforces security;
		// it just makes it harder to make a stupid mistake and remove
		// the permission check.
		return &errcode.HTTPErr{Status: http.StatusUnauthorized, Err: errors.New("authorization check failed")}
	}

	// Only return the last response to the HTTP client (e.g., just
	// return the result of "textDocument/definition" if "initialize"
	// is followed by a "textDocument/definition") because that's all
	// the client needs.
	resps := make([]*jsonrpc2.Response, len(reqs))
	for i, req := range reqs {
		// ?prepare indicates we are only doing the request to warm up
		// the LSP servers. Only the HTTP gateway understands this, so
		// we do not pass it on.
		req.Method = strings.TrimSuffix(req.Method, "?prepare")
		if req.Notif {
			if err := c.Notify(ctx, req.Method, req.Params); err != nil {
				return err
			}
		} else {
			resps[i] = &jsonrpc2.Response{}
			err := c.Call(ctx, req.Method, req.Params, &resps[i].Result)
			if e, ok := err.(*jsonrpc2.Error); ok {
				// We do not mark the handler as failed, but
				// we want to record that it failed in
				// lightstep.
				ext.Error.Set(span, true)
				span.LogEvent(fmt.Sprintf("error: %s failed with %v", req.Method, err))
				ev.AddField("lsp_error", e.Message)
				success = false
				if !handlerutil.DebugMode {
					e.Message = "(error message omitted)"
				}
				resps[i].Error = e
			} else if err != nil {
				return err
			}
		}
	}
	return writeJSON(w, resps)
}

func addRootPathFields(ev *libhoney.Event, u *uri.URI) {
	// u usually looks something like git://github.com/foo/bar?commithash
	ev.AddField("repo", u.Host+u.Path)
	ev.AddField("commit", u.RawQuery)
	if u.Host == "github.com" && len(u.Path) > 1 {
		i := strings.Index(u.Path[1:], "/")
		if i > 0 {
			ev.AddField("repo_org", u.Path[1:i+1])
		}
	}
}
