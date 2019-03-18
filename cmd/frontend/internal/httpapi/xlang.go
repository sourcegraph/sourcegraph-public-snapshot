package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	libhoney "github.com/honeycombio/libhoney-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/httpapi"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/honey"
	xlang_lspext "github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
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
	// Buckets are similar to trace.UserLatencyBuckets, but with more granularity for apdex measurements.
	Buckets: []float64{0.1, 0.2, 0.5, 0.8, 1, 1.5, 2, 5, 10, 15, 20, 30},
}, []string{"success", "method", "mode", "transport"})

func init() {
	prometheus.MustRegister(xlangRequestDuration)
}

func serveXLang(w http.ResponseWriter, r *http.Request) (err error) {
	start := time.Now()
	success := true
	method := "unknown"
	mode := "unknown"
	ev := honey.Event("xlang")
	emptyResponse := true
	defer func() {
		duration := time.Since(start)

		// We shouldn't make the distinction of user error, since our
		// frontend code shouldn't send "invalid" requests. An example
		// is the browser-ext sending hover requests for private repos
		// we are not authorized to view. For now we will skip
		// recording user errors in our apdex scores, but record them
		// in honeycomb for deep
		// diving. https://github.com/sourcegraph/sourcegraph/issues/2362
		isUserError := false
		if err != nil && errcode.HTTP(err) < 500 {
			isUserError = true
		}
		if !isUserError {
			labels := prometheus.Labels{
				"success":   fmt.Sprintf("%t", err == nil && success),
				"method":    method,
				"mode":      mode,
				"transport": "http",
			}
			xlangRequestDuration.With(labels).Observe(duration.Seconds())
		}

		if honey.Enabled() {
			status := strconv.FormatBool(err == nil && success)
			if isUserError {
				status = "usererror"
			}
			ev.AddField("success", status)
			ev.AddField("empty", emptyResponse)
			ev.AddField("method", method)
			ev.AddField("mode", mode)
			ev.AddField("duration_ms", duration.Seconds()*1000)
			ev.AddField("client", r.Header.Get("x-sourcegraph-client"))
			ev.AddField("referrer", r.Referer())
			ev.AddField("user_agent", r.UserAgent())
			if err != nil {
				ev.AddField("error", err.Error())
			}
			if actor := actor.FromContext(r.Context()); actor != nil {
				ev.AddField("uid", actor.UID)
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
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		return err
	}

	if len(reqs) >= 3 {
		method = reqs[len(reqs)-3].Method
	}
	span, ctx := opentracing.StartSpanFromContext(r.Context(), fmt.Sprintf("LSP HTTP gateway: %s: %s", mode, method))
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Sanity check the request body. Be strict based on what we know
	// the UI sends us.
	if l := len(reqs); l != 3 && l != 4 {
		return fmt.Errorf("got %d jsonrpc2 requests, want exactly 3 or 4", len(reqs))
	}
	if reqs[0].Method != "initialize" || reqs[len(reqs)-2].Method != "shutdown" || reqs[len(reqs)-1].Method != "exit" {
		return fmt.Errorf("invalid jsonrpc2 request methods (%s, ..., %s, %s): expected (initialize, ..., shutdown, exit)", reqs[0].Method, reqs[len(reqs)-2].Method, reqs[len(reqs)-1].Method)
	}
	if len(reqs) == 4 && reqs[1].Method == "initialize" {
		return fmt.Errorf("invalid jsonrpc2 request method for 2nd request: %q is not allowed", reqs[1].Method)
	}
	if reqs[0].Params == nil {
		return errors.New("invalid jsonrpc2 initialize request: empty params")
	}
	if reqs[len(reqs)-1].ID != (jsonrpc2.ID{}) {
		return errors.New("invalid jsonrpc2 exit request: id should NOT be present")
	}
	var initParams xlang_lspext.ClientProxyInitializeParams
	if err := json.Unmarshal(*reqs[0].Params, &initParams); err != nil {
		return fmt.Errorf("invalid jsonrpc2 initialize params: %s", err)
	}
	{
		// DEPRECATED: Be compatible with both
		// pre-Mode-field-migration LSP proxy servers and
		// post-migration LSP proxy servers.
		if initParams.InitializationOptions.Mode == "" {
			initParams.InitializationOptions.Mode = initParams.Mode
		} else {
			initParams.Mode = initParams.InitializationOptions.Mode
		}
	}
	span.SetTag("RootURI", initParams.RootURI)
	if initParams.RootURI == "" {
		return errors.New("invalid empty LSP root URI in initialize request")
	}
	rootURI, err := uri.Parse(string(initParams.RootURI))
	if err != nil {
		return fmt.Errorf("invalid LSP root URI %q: %s", initParams.RootURI, err)
	}
	addRootURIFields(ev, rootURI)
	if initParams.InitializationOptions.Mode != "" {
		mode = initParams.InitializationOptions.Mode

		// Update span operation name now that we have a mode.
		span.SetOperationName(fmt.Sprintf("LSP HTTP gateway: %s: %s", mode, method))
	}

	// Check consistency against the URL. The URL route params are for
	// ease of debugging only, but it'd be confusing if they could
	// diverge from the actual jsonrpc2 request.
	if v := mux.Vars(r)["LSPMethod"]; v != method {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: fmt.Errorf("LSP method param in URL %q != %q method in LSP message params", v, method)}
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
		if _, err := backend.Repos.GetByURI(ctx, rootURI.Repo()); err != nil {
			return err
		}
		checkedUserHasReadAccessToRepo = true
	}

	// Use a one-shot connection to the LSP proxy. This is cheap,
	// since the LSP proxy will reuse an already running server for
	// the given workspace if available.
	c, err := httpapi.XLangNewClient()
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

	// We usually modify initParams, so marshal them again
	err = reqs[0].SetParams(initParams)
	if err != nil {
		return err
	}

	// Only return the last response to the HTTP client (e.g., just
	// return the result of "textDocument/definition" if "initialize"
	// is followed by a "textDocument/definition") because that's all
	// the client needs.
	resps := make([]*jsonrpc2.Response, 0, len(reqs))
	for i, req := range reqs {
		// ?prepare indicates we are only doing the request to warm up
		// the language servers. Only the HTTP gateway understands this, so
		// we do not pass it on.
		req.Method = strings.TrimSuffix(req.Method, "?prepare")
		if req.Notif {
			if err := c.Notify(ctx, req.Method, req.Params); err != nil {
				return err
			}
		} else {
			resp := &jsonrpc2.Response{ID: reqs[i].ID}
			resps = append(resps, resp)
			err := c.Call(ctx, req.Method, req.Params, &resp.Result)
			if err == nil && resp.Result == nil {
				// c.Call sets Result to Go nil if the response has a
				// JSON "null" result (per the rules of
				// json.Unmarshal). But a JSON-RPC 2.0 response
				// requires either the "result" or "error" field, so
				// we must prevent the "result" field from being
				// omitted altogether.
				resp.Result = &jsonNull
			}
			if e, ok := err.(*jsonrpc2.Error); ok {
				// We do not mark the handler as failed, but
				// we want to record that it failed in
				// the trace.
				ext.Error.Set(span, true)
				span.LogFields(
					otlog.String("method", req.Method),
					otlog.Error(err))
				ev.AddField("lsp_error", e.Message)
				success = false
				resp.Error = e
			} else if err != nil {
				return err
			} else if err == nil && i == 1 {
				// We want to mark whether or not we've gotten a result or not
				// in the response.
				var result interface{}
				if resp.Result == nil {
					emptyResponse = true // nil result
				} else if err := json.Unmarshal(*resp.Result, &result); err != nil {
					emptyResponse = true // unmarshal error
				} else {
					emptyResponse = isEmpty(result) // empty unmarshaled result
				}
			}
		}
	}
	return writeJSON(w, resps)
}

var jsonNull = json.RawMessage("null")

// isEmpty tells if v is nil or an empty slice or map. In all other cases, it
// returns false.
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	vv := reflect.ValueOf(v)
	if vv.IsNil() {
		return true
	}
	switch vv.Kind() {
	case reflect.Slice, reflect.Map:
		return vv.Len() == 0
	default:
		return false
	}
}

func addRootURIFields(ev *libhoney.Event, u *uri.URI) {
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
