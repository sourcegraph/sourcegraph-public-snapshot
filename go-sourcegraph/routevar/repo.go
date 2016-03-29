package routevar

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/spec"

	"github.com/sourcegraph/mux"
)

var (
	// Repo captures RepoSpec strings in URL routes.
	Repo = `{Repo:` + NamedToNonCapturingGroups(spec.RepoPattern) + `}`

	// RepoRev captures RepoRevSpec strings in URL routes.
	RepoRev = Repo + `{Rev:(?:@` + NamedToNonCapturingGroups(spec.RevPattern) + `)?}`
)

// FixRepoRevVars is a mux.PostMatchFunc that cleans and normalizes
// the route vars pertaining to a RepoRev.
func FixRepoRevVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	if match.Vars["Rev"] == "" {
		delete(match.Vars, "Rev")
	}
	if _, present := match.Vars["Rev"]; present {
		match.Vars["Rev"] = strings.TrimPrefix(match.Vars["Rev"], "@")
	}
}

// PrepareRepoRevRouteVars is a mux.BuildVarsFunc that converts from a
// RepoRevSpec's route vars to components used to generate routes.
func PrepareRepoRevRouteVars(vars map[string]string) map[string]string {
	if vars["Rev"] == "" {
		vars["Rev"] = ""
	} else {
		vars["Rev"] = "@" + vars["Rev"]
	}
	return vars
}
