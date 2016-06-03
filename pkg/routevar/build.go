package routevar

import (
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func BuildRouteVars(s sourcegraph.BuildSpec) map[string]string {
	m := RepoRouteVars(s.Repo)
	m["Build"] = fmt.Sprintf("%d", s.ID)
	return m
}

func TaskRouteVars(s sourcegraph.TaskSpec) map[string]string {
	v := BuildRouteVars(s.Build)
	v["Task"] = fmt.Sprintf("%d", s.ID)
	return v
}
