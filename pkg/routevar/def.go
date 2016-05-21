package routevar

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

// Def captures def paths in URL routes.
const Def = "{UnitType}/{Unit:.+?}/-/{Path:.*?}"

func defURLPathToKeyPath(s string) string {
	if s == "_._" {
		return "."
	}
	return s
}

func DefKeyPathToURLPath(s string) string {
	if s == "." {
		return "_._"
	}
	return s
}

func DefRouteVars(s sourcegraph.DefSpec) map[string]string {
	rev := s.CommitID
	if !strings.HasPrefix(s.CommitID, "@") && rev != "" {
		rev = "@" + rev
	}
	return map[string]string{
		"Repo":     s.Repo,
		"Rev":      rev,
		"UnitType": s.UnitType,
		"Unit":     DefKeyPathToURLPath(s.Unit),
		"Path":     DefKeyPathToURLPath(pathEscape(s.Path)),
	}
}

func ToDefSpec(routeVars map[string]string) (sourcegraph.DefSpec, error) {
	repoRev, err := ToRepoRevSpec(routeVars)
	if err != nil {
		return sourcegraph.DefSpec{}, err
	}

	return sourcegraph.DefSpec{
		Repo:     repoRev.URI,
		CommitID: ResolvedRevString(repoRev),
		UnitType: routeVars["UnitType"],
		Unit:     defURLPathToKeyPath(routeVars["Unit"]),
		Path:     defURLPathToKeyPath(pathUnescape(routeVars["Path"])),
	}, nil
}

// pathEscape is a limited version of url.QueryEscape that only escapes '?'.
func pathEscape(p string) string {
	return strings.Replace(p, "?", "%3F", -1)
}

// pathUnescape is a limited version of url.QueryEscape that only unescapes '?'.
func pathUnescape(p string) string {
	return strings.Replace(p, "%3F", "?", -1)
}
