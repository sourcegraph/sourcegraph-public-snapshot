package router

import (
	"net/http"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/sourcegraph/mux"
)

// Note: This does not support def paths and units that are equal to
// "." or "". It is too complex for too little value to support those
// in this transition URL period.
func addOldDefRedirectRoute(genURLRouter *Router, matchRouter *mux.Router) {
	// same as spec.unresolvedRevPattern but also not allowing path
	// components starting with ".".
	const revSuffixNoDots = `{Rev:(?:@(?:(?:[^@=/.-]|(?:[^=/@.]{2,}))/)*(?:[^@=/.-]|(?:[^=/@.]{2,})))?}`

	matchRouter.Path("/" + routevar.Repo + revSuffixNoDots + `/.{UnitType:(?:GoPackage|JavaPackage|JavaArtifact|CommonJSPackage)}/{rawUnit:.*}.def{Path:(?:(?:/(?:[^/.][^/]*/)*(?:[^/.][^/]*))|)}`).Methods("GET").Name(OldDefRedirect).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		fixDefVars(v)
		defSpec, err := sourcegraph.UnmarshalDefSpec(v)
		if err != nil {
			log15.Error("Invalid def URL in oldDefRedirect", "err", err, "url", r.URL.String(), "routeVars", v)
			http.Error(w, "invalid def URL", http.StatusNotFound)
			return
		}

		u := genURLRouter.URLToDef(defSpec.DefKey())
		if u == nil {
			log15.Error("Failed to generate new URL in oldDefRedirect", "err", err, "defSpec", defSpec, "routeVars", v)
			http.Error(w, "failed to generate redirect to new def URL", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
	})
}

// pathUnescape is a limited version of url.QueryEscape that only unescapes '?'.
func pathUnescape(p string) string {
	return strings.Replace(p, "%3F", "?", -1)
}

func fixDefVars(v map[string]string) {
	v["Path"] = strings.TrimPrefix(v["Path"], "/")
	if path := v["Path"]; path == "" {
		v["Path"] = "."
	}
	v["Path"] = pathUnescape(v["Path"])

	rawUnit := v["rawUnit"]
	if rawUnit == "" {
		v["Unit"] = "."
	} else {
		v["Unit"] = strings.TrimSuffix(rawUnit, "/")
	}
	delete(v, "rawUnit")
}
