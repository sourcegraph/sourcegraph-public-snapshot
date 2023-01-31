package router

import (
	"net/http"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/routevar"
	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/gorilla/mux"
)

// same as spec.unresolvedRevPattern but also not allowing path
// components starting with ".".
const revSuffixNoDots = `{Rev:(?:@(?:(?:[^@=/.-]|(?:[^=/@.]{2,}))/)*(?:[^@=/.-]|(?:[^=/@.]{2,})))?}`

func addOldTreeRedirectRoute(matchRouter *mux.Router) {
	matchRouter.Path("/" + routevar.Repo + revSuffixNoDots + `/.tree{Path:.*}`).Methods("GET").Name(OldTreeRedirect).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		cleanedPath := path.Clean(v["Path"])
		if !strings.HasPrefix(cleanedPath, "/") && cleanedPath != "" {
			cleanedPath = "/" + cleanedPath
		}

		http.Redirect(w, r, URLToRepoTreeEntry(api.RepoName(v["Repo"]), v["Rev"], cleanedPath).String(), http.StatusMovedPermanently)
	})
}
