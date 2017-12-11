package httpapi

import (
	"log"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	opentracing "github.com/opentracing/opentracing-go"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
	apirouter "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// NewHandler returns a new API handler that uses the provided API
// router, which must have been created by httpapi/router.New, or
// creates a new one if nil.
func NewHandler(m *mux.Router) http.Handler {
	if m == nil {
		m = apirouter.New(nil)
	}
	m.StrictSlash(true)

	// Set handlers for the installed routes.
	m.Get(apirouter.RepoShield).Handler(traceutil.TraceRoute(handler(serveRepoShield)))

	m.Get(apirouter.SubmitForm).Handler(traceutil.TraceRoute(handler(serveSubmitForm)))

	m.Get(apirouter.Telemetry).Handler(traceutil.TraceRoute(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "https"
			req.URL.Host = "sourcegraph-logging.telligentdata.com"
			req.Host = "sourcegraph-logging.telligentdata.com"
			req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]
		},
	}))

	m.Get(apirouter.XLang).Handler(traceutil.TraceRoute(handler(serveXLang)))

	// 🚨 SECURITY: The LSP endpoints specifically allows cookie authorization 🚨
	// because the JavaScript WebSocket API does not allow us to set custom
	// headers. It is possible to send a basic authorization header, but hacking
	// it to send our auth cookie doesn't seem worth the complexity.
	//
	// This does not introduce a CSRF vulnerability (mentioned in the security comment below), because
	// gorilla/websocket verifies the origin of the HTTP request before upgrading it to a web socket:
	// https://sourcegraph.com/github.com/gorilla/websocket/-/blob/server.go#L126:1-130:1
	//
	// You can read more about this security issue here:
	// https://www.christian-schneider.net/CrossSiteWebSocketHijacking.html
	m.Get(apirouter.LSP).Handler(traceutil.TraceRoute(session.CookieMiddleware(httpapiauth.AuthorizationMiddleware(http.HandlerFunc(serveLSP)))))

	m.Get(apirouter.GraphQL).Handler(traceutil.TraceRoute(handler(serveGraphQL)))

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API no route: %s %s from %s", r.Method, r.URL, r.Referer())
		http.Error(w, "no route", http.StatusNotFound)
	})

	// SECURITY NOTE: The HTTP API should not accept cookies as
	// authentication (except with CookieMiddlewareIfHeader). Doing so
	// would open it up to CSRF attacks.
	var h http.Handler = m
	h = session.CookieMiddlewareIfHeader(h, "X-Requested-By")
	h = httpapiauth.AuthorizationMiddleware(h)

	return h
}

// NewInternalHandler returns a new API handler for internal endpoints that uses
// the provided API router, which must have been created by httpapi/router.NewInternal.
//
// 🚨 SECURITY: This handler should not be served on a publicly exposed port. 🚨
// This handler is not guarenteed to provide the same authorization checks as
// public API handlers.
func NewInternalHandler(m *mux.Router) http.Handler {
	if m == nil {
		m = apirouter.New(nil)
	}
	m.StrictSlash(true)

	m.Get(apirouter.PhabricatorRepoCreate).Handler(traceutil.TraceRoute(handler(servePhabricatorRepoCreate)))
	m.Get(apirouter.ReposCreateIfNotExists).Handler(traceutil.TraceRoute(handler(serveReposCreateIfNotExists)))
	m.Get(apirouter.ReposUpdateIndex).Handler(traceutil.TraceRoute(handler(serveReposUpdateIndex)))
	m.Get(apirouter.ReposUnindexedDependencies).Handler(traceutil.TraceRoute(handler(serveReposUnindexedDependencies)))
	m.Get(apirouter.ReposInventoryUncached).Handler(traceutil.TraceRoute(handler(serveReposInventoryUncached)))
	m.Get(apirouter.ReposGetByURI).Handler(traceutil.TraceRoute(handler(serveReposGetByURI)))
	m.Get(apirouter.DefsRefreshIndex).Handler(traceutil.TraceRoute(handler(serveDefsRefreshIndex)))
	m.Get(apirouter.GitoliteUpdateRepos).Handler(traceutil.TraceRoute(handler(serveGitoliteUpdateRepos)))

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API no route: %s %s from %s", r.Method, r.URL, r.Referer())
		http.Error(w, "no route", http.StatusNotFound)
	})

	return m
}

// handler is a wrapper func for API handlers.
func handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return handlerutil.HandlerWithErrorReturn{
		Handler: func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "application/json")
			return h(w, r)
		},
		Error: handleError,
	}
}

var schemaDecoder = schema.NewDecoder()

func init() {
	schemaDecoder.IgnoreUnknownKeys(true)

	// Register a converter for unix timestamp strings -> time.Time values
	// (needed for Appdash PageLoadEvent type).
	schemaDecoder.RegisterConverter(time.Time{}, func(s string) reflect.Value {
		ms, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.ValueOf(err)
		}
		return reflect.ValueOf(time.Unix(0, ms*int64(time.Millisecond)))
	})
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	// Handle custom errors
	if ee, ok := err.(*handlerutil.URLMovedError); ok {
		err := handlerutil.RedirectToNewRepoURI(w, r, ee.NewURL)
		if err != nil {
			log15.Error("error redirecting to new URI", "err", err, "new_url", ee.NewURL)
		}
		return
	}

	// Never cache error responses.
	w.Header().Set("cache-control", "no-cache, max-age=0")

	errBody := err.Error()

	var displayErrBody string
	if handlerutil.DebugMode {
		// Only display error message to admins when in debug mode, since it may
		// contain sensitive info (like API keys in net/http error messages).
		displayErrBody = string(errBody)
	}
	http.Error(w, displayErrBody, status)
	traceSpan := opentracing.SpanFromContext(r.Context())
	var spanURL string
	if traceSpan != nil {
		spanURL = traceutil.SpanURL(traceSpan)
	}
	if status < 200 || status >= 500 {
		log15.Error("API HTTP handler error response", "method", r.Method, "request_uri", r.URL.RequestURI(), "status_code", status, "error", err, "trace", spanURL)
	}
}
