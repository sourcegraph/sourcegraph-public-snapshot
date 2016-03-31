package router

import (
	"net/http"
	"path"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/routevar"

	"github.com/gorilla/mux"
)

func addOldTreeRedirectRoute(genURLRouter *Router, matchRouter *mux.Router) {
	matchRouter.Path("/" + routevar.Repo + revSuffixNoDots + `/.tree{Path:.*}`).Methods("GET").Name(OldTreeRedirect).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		path := path.Clean(v["Path"])
		if !strings.HasPrefix(path, "/") && path != "" {
			path = "/" + path
		}

		u := genURLRouter.URLToRepoTreeEntry(v["Repo"], v["Rev"], path)
		if u == nil {
			log15.Error("Failed to generate new URL in oldTreeRedirect", "routeVars", v)
			http.Error(w, "failed to generate redirect to new tree URL", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
	})
}
