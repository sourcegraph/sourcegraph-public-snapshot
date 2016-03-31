package sourcegraph

import (
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func defURLPathToKeyPath(s string) string {
	if s == "_._" {
		return "."
	}
	return s
}

func defKeyPathToURLPath(s string) string {
	if s == "." {
		return "_._"
	}
	return s
}

func (s DefSpec) RouteVars() map[string]string {
	rev := s.CommitID
	if !strings.HasPrefix(s.CommitID, "@") && rev != "" {
		rev = "@" + rev
	}
	return map[string]string{
		"Repo":     s.Repo,
		"Rev":      rev,
		"UnitType": s.UnitType,
		"Unit":     defKeyPathToURLPath(s.Unit),
		"Path":     defKeyPathToURLPath(pathEscape(s.Path)),
	}
}

func UnmarshalDefSpec(routeVars map[string]string) (DefSpec, error) {
	repoRev, err := UnmarshalRepoRevSpec(routeVars)
	if err != nil {
		return DefSpec{}, err
	}

	return DefSpec{
		Repo:     repoRev.URI,
		CommitID: repoRev.ResolvedRevString(),
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

// DefKey returns the def key specified by s, using the Repo, UnitType,
// Unit, and Path fields of s.
func (s *DefSpec) DefKey() graph.DefKey {
	if s.Repo == "" {
		panic("Repo is empty")
	}
	if s.UnitType == "" {
		panic("UnitType is empty")
	}
	if s.Unit == "" {
		panic("Unit is empty")
	}
	return graph.DefKey{
		Repo:     s.Repo,
		CommitID: s.CommitID,
		UnitType: s.UnitType,
		Unit:     s.Unit,
		Path:     s.Path,
	}
}

// NewDefSpecFromDefKey returns a DefSpec that specifies the same
// def as the given key.
func NewDefSpecFromDefKey(key graph.DefKey) DefSpec {
	return DefSpec{
		Repo:     key.Repo,
		CommitID: key.CommitID,
		UnitType: key.UnitType,
		Unit:     key.Unit,
		Path:     key.Path,
	}
}

// DefSpec returns the DefSpec that specifies s.
func (s *Def) DefSpec() DefSpec {
	spec := NewDefSpecFromDefKey(s.Def.DefKey)
	return spec
}
