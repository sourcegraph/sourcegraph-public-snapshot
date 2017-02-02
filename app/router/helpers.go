package router

import (
	"fmt"
	"net/url"
	"strings"
)

func (r *Router) URLToRepo(repo string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s", repo)}
}

func (r *Router) URLToSitemap(lang string) *url.URL {
	if lang != "" {
		return &url.URL{Path: fmt.Sprintf("/sitemap/%s", lang)}
	}
	return &url.URL{Path: "/sitemap"}
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

func (r *Router) URLToBlob(repo, rev, file string, line int) *url.URL {
	return &url.URL{
		Path:     fmt.Sprintf("/%s%s/-/blob/%s", repo, revStr(rev), file),
		Fragment: fmt.Sprintf("L%d", line),
	}
}

func (r *Router) URLToBlobRange(repo, rev, file string, startLine, endLine, startCharacter, endCharacter int) *url.URL {
	return &url.URL{
		Path:     fmt.Sprintf("/%s%s/-/blob/%s", repo, revStr(rev), file),
		Fragment: fmt.Sprintf("L%d:%d-%d:%d", startLine, startCharacter, endLine, endCharacter),
	}
}

func revStr(rev string) string {
	if rev == "" || strings.HasPrefix(rev, "@") {
		return rev
	}
	return "@" + rev
}
