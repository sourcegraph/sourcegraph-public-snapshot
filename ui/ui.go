package ui

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/schema"
	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	ui_router "src.sourcegraph.com/sourcegraph/ui/router"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
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
	mw := []handlerutil.Middleware{
		appauth.CookieMiddleware,
		handlerutil.UserMiddleware,
	}

	if r == nil {
		r = ui_router.New(nil)
	}

	r.Get(ui_router.RepoTree).Handler(handler(serveRepoTree))

	r.Get(ui_router.RepoFileFinder).Handler(handler(serveRepoFileFinder))

	r.Get(ui_router.Definition).Handler(handler(serveDef))
	r.Get(ui_router.DefExamples).Handler(handler(serveDefExamples))

	r.Get(ui_router.RepoCreate).Handler(handler(serveRepoCreate))
	r.Get(ui_router.RepoCommits).Handler(handler(serveRepoCommits))

	r.Get(ui_router.SearchTokens).Handler(handler(serveTokenSearch))
	r.Get(ui_router.SearchText).Handler(handler(serveTextSearch))

	r.Get(ui_router.AppdashUploadPageLoad).Handler(handler(serveAppdashUploadPageLoad))

	if !appconf.Flags.DisableUserContent {
		r.Get(ui_router.UserContentUpload).Handler(handler(serveUserContentUpload))
	}

	if authutil.ActiveFlags.PrivateMirrors {
		r.Get(ui_router.RepoMirror).Handler(handler(serveRepoMirror))
		r.Get(ui_router.UserInviteBulk).Handler(handler(serveUserInviteBulk))
	}

	r.Get(ui_router.UserInvite).Handler(handler(serveUserInvite))
	r.Get(ui_router.UserKeys).Handler(handler(serveUserKeys))

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
	msg := err.Error() + " (Code: " + strconv.Itoa(status) + ")"
	err = json.NewEncoder(w).Encode(struct{ Error string }{msg})
	if err != nil {
		log.Printf("Error during encoding error response: %s", err)
	}
	log.Println("ui.serveError serving error:", msg)
}
