package routevar

import "strings"

// Delta is like sourcegraph.DeltaSpec, but it allows non-absolute
// commit IDs.
type Delta struct {
	Base RepoRev
	Head RepoRev
}

// DeltaRouteVars returns the route variables for generating URLs to
// the delta specified by this Delta.
func DeltaRouteVars(s Delta) map[string]string {
	m := RepoRevRouteVars(s.Base)
	m["DeltaHeadRev"] = "@" + s.Head.Rev
	return m
}

// ToDelta marshals a map containing route variables for a
// DeltaSpec and returns the equivalent DeltaSpec struct.
func ToDelta(routeVars map[string]string) Delta {
	repoRev := ToRepoRev(routeVars)
	return Delta{
		Base: repoRev,
		Head: RepoRev{
			RepoSpec: repoRev.RepoSpec,
			Rev:      strings.TrimPrefix(routeVars["DeltaHeadRev"], "@"),
		},
	}
}
