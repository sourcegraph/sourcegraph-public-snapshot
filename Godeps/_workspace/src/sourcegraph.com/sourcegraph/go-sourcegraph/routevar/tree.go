package routevar

import (
	"net/http"
	"path"
	"strings"

	"github.com/sourcegraph/mux"
)

// TreeEntryPath captures tree entry paths in URL routes.
var TreeEntryPath = `{Path:(?:/.*)*}`

// FixTreeEntryVars is a mux.PostMatchFunc that cleans and normalizes
// the path to a tree entry.
func FixTreeEntryVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	path := path.Clean(strings.TrimPrefix(match.Vars["Path"], "/"))
	if path == "" || path == "." {
		match.Vars["Path"] = "."
	} else {
		match.Vars["Path"] = path
	}
}

// PrepareTreeEntryRouteVars is a mux.BuildVarsFunc that converts from
// a cleaned and normalized Path to a Path that we use to generate
// tree entry URLs.
func PrepareTreeEntryRouteVars(vars map[string]string) map[string]string {
	if path := vars["Path"]; path == "." {
		vars["Path"] = ""
	} else {
		vars["Path"] = "/" + path
	}
	return vars
}
