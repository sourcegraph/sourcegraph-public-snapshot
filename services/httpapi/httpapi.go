package httpapi

import (
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/auth"
	apirouter "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/router"
)

// NewHandler returns a new API handler that uses the provided API
// router, which must have been created by httpapi/router.New, or
// creates a new one if nil.
func NewHandler(m *mux.Router) http.Handler {
	if m == nil {
		m = apirouter.New(nil)
	}
	m.StrictSlash(true)

	// SECURITY NOTE: The HTTP API should not accept cookies as
	// authentication. Doing so would open it up to CSRF
	// attacks.
	var mw []handlerutil.Middleware
	mw = append(mw, httpapiauth.AuthorizationMiddleware)
	mw = append(mw, eventsutil.AgentMiddleware)

	// Set handlers for the installed routes.
	m.Get(apirouter.RepoCreate).Handler(httptrace.TraceRoute(handler(serveRepoCreate)))
	m.Get(apirouter.RepoRefresh).Handler(httptrace.TraceRoute(handler(func(w http.ResponseWriter, r *http.Request) error { return nil }))) // legacy
	m.Get(apirouter.RepoResolveRev).Handler(httptrace.TraceRoute(handler(serveRepoResolveRev)))
	m.Get(apirouter.RepoDefLanding).Handler(httptrace.TraceRoute(handler(serveRepoDefLanding)))
	m.Get(apirouter.Repos).Handler(httptrace.TraceRoute(handler(serveRepos)))

	m.Get(apirouter.Orgs).Handler(httptrace.TraceRoute(handler(serveOrgs)))
	m.Get(apirouter.OrgMembers).Handler(httptrace.TraceRoute(handler(serveOrgMembers)))
	m.Get(apirouter.OrgInvites).Handler(httptrace.TraceRoute(handler(serveOrgInvites)))

	m.Get(apirouter.BetaSubscription).Handler(httptrace.TraceRoute(handler(serveBetaSubscription)))

	m.Get(apirouter.XLang).Handler(httptrace.TraceRoute(handler(serveXLang)))

	m.Get(apirouter.GraphQL).Handler(httptrace.TraceRoute(handler(serveGraphQL)))

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API no route: %s %s from %s", r.Method, r.URL, r.Referer())
		http.Error(w, "no route", http.StatusNotFound)
	})

	return handlerutil.WithMiddleware(m, mw...)
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
	if status < 200 || status >= 500 {
		log15.Error("API HTTP handler error response", "method", r.Method, "request_uri", r.URL.RequestURI(), "status_code", status, "error", err, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
	}
}
