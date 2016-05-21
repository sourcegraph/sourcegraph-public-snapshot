package router

import (
	"fmt"
	"net/url"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func (r *Router) URLToRepo(repo string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s", repo)}
}

func (r *Router) URLToRepoSitemap(repo string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s/sitemap.xml", repo)}
}

func (r *Router) URLToRepoRev(repo, rev string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s%s", repo, revStr(rev))}
}

func (r *Router) URLToRepoTreeEntry(repo, rev, path string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s%s/-/tree/%s", repo, revStr(rev), path)}
}

func (r *Router) URLToDef(key graph.DefKey) *url.URL {
	return &url.URL{
		Path: fmt.Sprintf("/%s%s/-/def/%s/%s/-/%s",
			key.Repo, revStr(key.CommitID), key.UnitType,
			routevar.DefKeyPathToURLPath(key.Unit),
			routevar.DefKeyPathToURLPath(key.Path),
		),
	}
}

func revStr(rev string) string {
	if rev == "" || strings.HasPrefix(rev, "@") {
		return rev
	}
	return "@" + rev
}
