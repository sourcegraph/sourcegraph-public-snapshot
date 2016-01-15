package changesets

import (
	"net/http"

	"github.com/gorilla/schema"
	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/platform"
)

var (
	router        = mux.NewRouter()
	schemaDecoder = schema.NewDecoder()
)

const (
	appID = "changes"

	routeView = "changesets.view"
)

func init() {
	schemaDecoder.IgnoreUnknownKeys(true)

	router.Path("/").Methods("GET").Handler(handlerWithError(serveList))
	router.Path("/create").Methods("POST").Handler(handlerWithError(serveCreate))
	router.Path(`/{ID:\d+}`).Methods("GET").Handler(handlerWithError(serveChangeset)).Name(routeView)
	router.Path(`/{ID:\d+}/files`).Methods("GET").Handler(handlerWithError(serveChangeset))
	router.Path(`/{ID:\d+}/files/{Filter:.+}`).Methods("GET").Handler(handlerWithError(serveChangeset))
	router.Path(`/{ID:\d+}/update`).Methods("POST").Handler(handlerWithError(serveUpdate))
	router.Path(`/{ID:\d+}/submit-review`).Methods("POST").Handler(handlerWithError(serveSubmitReview))
	router.Path(`/{ID:\d+}/merge`).Methods("POST").Handler(handlerWithError(serveMerge))

	platform.RegisterFrame(platform.RepoFrame{
		ID:      appID,
		Title:   "Changes",
		Icon:    "git-pull-request",
		Handler: router,
	})
}

// handleWithError takes a custom handler that may return an error and returns
// a valid `http.Handler`. If an error is returned, it is captured and handled.
func handlerWithError(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
