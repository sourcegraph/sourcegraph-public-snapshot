package router

import (
	"fmt"
	"net/url"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func URLToRepoRev(repo api.RepoURI, rev string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s%s", repo, revStr(rev))}
}

func URLToRepoTreeEntry(repo api.RepoURI, rev, path string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s%s/-/tree/%s", repo, revStr(rev), path)}
}

func revStr(rev string) string {
	if rev == "" || strings.HasPrefix(rev, "@") {
		return rev
	}
	return "@" + rev
}
