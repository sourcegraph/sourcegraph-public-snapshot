package ui

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// serveShortSearch redirects /short/abcdefgh= queries to the appropriate
// search query result page
func serveShortSearch(db database.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var savedSearchID int32
		err := relay.UnmarshalSpec(graphql.ID(mux.Vars(r)["id"]), &savedSearchID)
		if err != nil {
			serveError(w, r, db, err, http.StatusNotFound)
			return
		}
		savedSearch, err := db.SavedSearches().GetByID(r.Context(), savedSearchID)
		if err != nil {
			serveError(w, r, db, err, http.StatusNotFound)
			return
		}
		http.Redirect(w, r, "/search?q="+url.QueryEscape(savedSearch.Config.Query), http.StatusMovedPermanently)
	}
}
