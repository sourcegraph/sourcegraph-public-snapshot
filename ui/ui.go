package ui

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/schema"
	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	ui_router "sourcegraph.com/sourcegraph/sourcegraph/ui/router"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/metricutil"
)

var (
	schemaDecoder = schema.NewDecoder()
	once          sync.Once
)

func init() {
	once.Do(func() {
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
	})
}

// NewHandler creates a new http.Handler for all UI endpoints, optionally using
// the provided router as a base.
func NewHandler(r *mux.Router) http.Handler {
	var mw []handlerutil.Middleware
	if authutil.ActiveFlags.HasUserAccounts() {
		mw = append(mw, appauth.CookieMiddleware, handlerutil.UserMiddleware)
	}
	if !metricutil.DisableMetricsCollection() {
		// TODO(rothfels): check what other special characters (besides semis) need to be sanitized
		// mw = append(mw, eventsutil.AgentMiddleware)
		mw = append(mw, eventsutil.DeviceIdMiddleware)
	}

	if r == nil {
		r = ui_router.New(nil)
	}

	r.Get(ui_router.RepoTree).Handler(handler(serveRepoTree))

	r.Get(ui_router.Definition).Handler(handler(serveDef))
	r.Get(ui_router.DefExamples).Handler(handler(serveDefExamples))

	r.Get(ui_router.RepoCommits).Handler(handler(serveRepoCommits))

	r.Get(ui_router.SearchTokens).Handler(handler(serveTokenSearch))
	r.Get(ui_router.SearchText).Handler(handler(serveTextSearch))

	r.Get(ui_router.AppdashUploadPageLoad).Handler(handler(serveAppdashUploadPageLoad))

	if !appconf.Flags.DisableUserContent {
		r.Get(ui_router.UserContentUpload).Handler(handler(serveUserContentUpload))
	}

	r.Get(ui_router.UserInviteBulk).Handler(handler(serveUserInviteBulk))

	return handlerutil.WithMiddleware(r, mw...)
}

func handler(fn func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return handlerutil.Handler(handlerutil.HandlerWithErrorReturn{
		Handler: jsonContentType(fn),
		Error:   serveError,
	})
}

func jsonContentType(fn func(w http.ResponseWriter, r *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		return fn(w, r)
	}
}

// serveError responds to the client by sending any error that might have occurred
// when processing a request.
func serveError(w http.ResponseWriter, req *http.Request, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if status < 200 || status >= 500 {
		log15.Error("UI HTTP handler error response", "method", req.Method, "request_uri", req.URL.RequestURI(), "status_code", status, "error", err)
	}

	msg := err.Error() + " (Code: " + strconv.Itoa(status) + ")"
	err = json.NewEncoder(w).Encode(struct{ Error string }{msg})
	if err != nil {
		log.Printf("Error during encoding error response: %s", err)
	}
}
