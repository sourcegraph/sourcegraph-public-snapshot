package router

import (
	"fmt"
	"net/url"
	"strings"
)

func (r *Router) URLToRepoRev(repo, rev string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s%s", repo, revStr(rev))}
}

func (r *Router) URLToRepoTreeEntry(repo, rev, path string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s%s/-/tree/%s", repo, revStr(rev), path)}
}

func revStr(rev string) string {
	if rev == "" || strings.HasPrefix(rev, "@") {
		return rev
	}
	return "@" + rev
}
