package routevar

import "sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/spec"

var (
	Repo = `{Repo:` + NamedToNonCapturingGroups(spec.RepoPattern) + `}`
	Rev  = `{Rev:` + NamedToNonCapturingGroups(spec.RevPattern) + `}`

	RepoRevSuffix = `{Rev:` + NamedToNonCapturingGroups(`(?:@`+spec.RevPattern+`)?`) + `}`
)
