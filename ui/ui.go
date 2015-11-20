package ui

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/sourcegraph/mux"

	"reflect"
	"strconv"
	"sync"
	"time"

	"src.sourcegraph.com/sourcegraph/app/appconf"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
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
// the provided router as a base. The second argument, if set to true, will enable
// testing endpoints and allow the mocking of any used services or API calls during
// the processing of this request.
func NewHandler(r *mux.Router, isTest bool) http.Handler {
	mw := []handlerutil.Middleware{
		handlerutil.CacheMiddleware,
		appauth.CookieMiddleware,
		handlerutil.UserMiddleware,
	}

	if r == nil {
		r = ui_router.New(nil, isTest)
	}

	p := payloadHandler{TestEnvironment: isTest}

	r.Get(ui_router.RepoTree).Handler(p.handler(serveRepoTree))

	r.Get(ui_router.RepoFileFinder).Handler(p.handler(serveRepoFileFinder))

	r.Get(ui_router.Definition).Handler(p.handler(serveDef))
	r.Get(ui_router.DefExamples).Handler(p.handler(serveDefExamples))

	r.Get(ui_router.Discussion).Handler(p.handler(serveDiscussion))
	r.Get(ui_router.DiscussionComment).Handler(p.handler(serveDiscussionCommentCreate))
	r.Get(ui_router.DiscussionCreate).Handler(p.handler(serveDiscussionCreate))
	r.Get(ui_router.DiscussionListDef).Handler(p.handler(serveDiscussionListDef))
	r.Get(ui_router.DiscussionListRepo).Handler(p.handler(serveDiscussionListRepo))

	r.Get(ui_router.SearchTokens).Handler(p.handler(serveTokenSearch))
	r.Get(ui_router.SearchText).Handler(p.handler(serveTextSearch))

	r.Get(ui_router.AppdashUploadPageLoad).Handler(p.handler(serveAppdashUploadPageLoad))

	if !appconf.Flags.DisableUserContent {
		r.Get(ui_router.UserContentUpload).Handler(p.handler(serveUserContentUpload))
	}

	return handlerutil.WithMiddleware(r, mw...)
}

// payloadHandler provides methods that return an http.Handler which is able to
// handle error returns and respond to them as JSON, as well as configure mock
// data for a test environment.
type payloadHandler struct {
	// TestEnvironment will cause the endpoints served by the handler to return
	// mock data, if set to true.
	TestEnvironment bool
}

func (h *payloadHandler) handler(fn func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return handlerutil.Handler(handlerutil.HandlerWithErrorReturn{
		Handler: h.serveHandler(fn),
		Error:   h.serveError,
	})
}

// serveHandler additionally augments the passed in handler with correct headers
// for JSON responses and enables mocking if this is a test environment.
func (h *payloadHandler) serveHandler(fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		if h.TestEnvironment && r.Method == "POST" && r.Header.Get("X-Mock-Response") == "yes" {
			m := new(serviceMocker)
			if err := m.Mock(r); err != nil {
				return err
			}
		}
		return fn(w, r)
	}
}

// serveError responds to the client by sending any error that might have occurred
// when processing a request.
func (h *payloadHandler) serveError(w http.ResponseWriter, req *http.Request, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	msg := err.Error() + " (Code: " + strconv.Itoa(status) + ")"
	err = json.NewEncoder(w).Encode(struct{ Error string }{msg})
	if err != nil {
		log.Printf("Error during encoding error response: %s", err)
	}
	log.Println(msg)
}
