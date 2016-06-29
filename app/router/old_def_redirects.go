package router

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"

	"github.com/gorilla/mux"
)

// same as spec.unresolvedRevPattern but also not allowing path
// components starting with ".".
const revSuffixNoDots = `{Rev:(?:@(?:(?:[^@=/.-]|(?:[^=/@.]{2,}))/)*(?:[^@=/.-]|(?:[^=/@.]{2,})))?}`

// Note: This does not support def paths and units that are equal to
// "." or "". It is too complex for too little value to support those
// in this transition URL period.
func addOldDefRedirectRoute(genURLRouter *Router, matchRouter *mux.Router) {
	matchRouter.Path("/" + routevar.Repo + revSuffixNoDots + `/.{UnitType:(?:GoPackage|JavaPackage|JavaArtifact|CommonJSPackage)}/{rawUnit:.*}.def{Path:(?:(?:/(?:[^/.][^/]*/)*(?:[^/.][^/]*))|)}`).Methods("GET").Name(OldDefRedirect).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		fixDefVars(v)
		def := routevar.ToDefAtRev(v)
		http.Redirect(w, r, genURLRouter.URLToDef(def).String(), http.StatusMovedPermanently)
	})

	// Match old "DEF/-/refs" URLs. These are also handled in
	// JavaScript, but the prefetching on the server will try to
	// handle them before JS can handle them unless we deal with them
	// here.
	matchRouter.Path("/" + routevar.Repo + routevar.RepoRevSuffix + "/-/def/" + routevar.Def + "/-/refs").Methods("GET").Name(OldDefRedirect).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		def := routevar.ToDefAtRev(mux.Vars(r))
		u := genURLRouter.URLToDef(def)
		u.Path = strings.Replace(u.Path, "/def/", "/info/", 1)
		u.RawQuery = ""
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
