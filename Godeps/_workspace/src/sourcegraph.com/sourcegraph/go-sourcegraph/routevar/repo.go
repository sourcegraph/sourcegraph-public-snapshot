package routevar

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/spec"

	"github.com/sourcegraph/mux"
)

var (
	// Repo captures RepoSpec strings in URL routes.
	Repo = `{Repo:` + NamedToNonCapturingGroups(spec.RepoPattern) + `}`

	// RepoRev captures RepoRevSpec strings in URL routes.
	RepoRev = Repo + `{ResolvedRev:(?:@` + NamedToNonCapturingGroups(spec.ResolvedRevPattern) + `)?}`
)

// FixRepoRevVars is a mux.PostMatchFunc that cleans and normalizes
// the route vars pertaining to a RepoRev.
func FixRepoRevVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	if _, present := match.Vars["ResolvedRev"]; present {
		match.Vars["ResolvedRev"] = strings.TrimPrefix(match.Vars["ResolvedRev"], "@")
	}
	FixResolvedRevVars(req, match, r)
}

// PrepareRepoRevRouteVars is a mux.BuildVarsFunc that converts from a
// RepoRevSpec's route vars to components used to generate routes.
func PrepareRepoRevRouteVars(vars map[string]string) map[string]string {
	vars = PrepareResolvedRevRouteVars(vars)
	if vars["ResolvedRev"] != "" {
		vars["ResolvedRev"] = "@" + vars["ResolvedRev"]
	}
	return vars
}

// FixResolvedRevVars is a mux.PostMatchFunc that cleans and
// normalizes the route vars pertaining to a ResolvedRev (Rev and CommitID).
func FixResolvedRevVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	if rrev, present := match.Vars["ResolvedRev"]; present {
		rev, commitID, err := spec.ParseResolvedRev(rrev)
		if err == nil || rrev == "" {
			// Propagate ResolvedRev if it was set and if parsing
			// failed; otherwise remove it.
			delete(match.Vars, "ResolvedRev")
		}
		if err == nil {
			if rev != "" {
				match.Vars["Rev"] = rev
			}
			if commitID != "" {
				match.Vars["CommitID"] = commitID
			}
		}
	}
}

// PrepareResolvedRevRouteVars is a mux.BuildVarsFunc that converts
// from a ResolvedRev's component route vars (Rev and CommitID) to a
// single ResolvedRev var.
func PrepareResolvedRevRouteVars(vars map[string]string) map[string]string {
	vars["ResolvedRev"] = spec.ResolvedRevString(vars["Rev"], vars["CommitID"])
	return vars
}
