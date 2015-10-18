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

func init() {
	router.StrictSlash(true)
	schemaDecoder.IgnoreUnknownKeys(true)

	router.Path("/").Methods("GET").Handler(handlerWithError(serveList))
	router.Path("/create").Methods("POST").Handler(handlerWithError(serveCreate))
	router.Path(`/{ID:\d+}`).Methods("GET").Handler(handlerWithError(serveChangeset))
	router.Path(`/{ID:\d+}/files`).Methods("GET").Handler(handlerWithError(serveChangeset))
	router.Path(`/{ID:\d+}/files/{Filter:.+}`).Methods("GET").Handler(handlerWithError(serveChangeset))
	router.Path(`/{ID:\d+}/update`).Methods("POST").Handler(handlerWithError(serveUpdate))
	router.Path(`/{ID:\d+}/submit-review`).Methods("POST").Handler(handlerWithError(serveSubmitReview))

	platform.RegisterFrame(platform.RepoFrame{
		ID:      "changes",
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
