package routevar

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

// DeltaRouteVars returns the route variables for generating URLs to
// the delta specified by this DeltaSpec.
func DeltaRouteVars(s sourcegraph.DeltaSpec) map[string]string {
	m := RepoRevRouteVars(s.Base)
	if rev := ResolvedRevString(s.Head); rev != "" {
		if !strings.HasPrefix(rev, "@") {
			rev = "@" + rev
		}
		m["DeltaHeadRev"] = rev
	}
	return m
}

// ToDeltaSpec marshals a map containing route variables for a
// DeltaSpec and returns the equivalent DeltaSpec struct.
func ToDeltaSpec(routeVars map[string]string) (sourcegraph.DeltaSpec, error) {
	s := sourcegraph.DeltaSpec{}

	rr, err := ToRepoRevSpec(routeVars)
	if err != nil {
		return sourcegraph.DeltaSpec{}, err
	}
	s.Base = rr

	dhr := strings.TrimPrefix(routeVars["DeltaHeadRev"], "@")
	rev, commitID := ParseResolvedRev(dhr)
	s.Head = sourcegraph.RepoRevSpec{RepoSpec: rr.RepoSpec, Rev: rev, CommitID: commitID}

	return s, nil
}
