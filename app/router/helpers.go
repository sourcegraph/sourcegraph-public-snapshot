package router

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func (r *Router) URLToRepo(repo string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s", repo)}
}

func (r *Router) URLToRepoLanding(repo string) *url.URL {
	return &url.URL{Path: fmt.Sprintf("/%s/-/info", repo)}
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

func (r *Router) URLToDef(def routevar.DefAtRev) *url.URL {
	return r.urlToDef(def.Repo, revStr(def.Rev), def.UnitType,
		routevar.DefKeyPathToURLPath(def.Unit),
		routevar.DefKeyPathToURLPath(def.Path),
	)
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

func (r *Router) URLToDefKey(def graph.DefKey) *url.URL {
	return r.urlToDef(def.Repo, revStr(def.CommitID), def.UnitType,
		routevar.DefKeyPathToURLPath(def.Unit),
		routevar.DefKeyPathToURLPath(def.Path),
	)
}

func (r *Router) URLToDefLanding(def graph.DefKey) *url.URL {
	return &url.URL{
		Path: fmt.Sprintf("/%s%s/-/info/%s/%s/-/%s",
			def.Repo, revStr(def.CommitID), def.UnitType, routevar.DefKeyPathToURLPath(def.Unit), routevar.DefKeyPathToURLPath(def.Path)),
	}
}

// URLToLegacyDefLanding is like URLToDefLanding except it only works for Go
// symbols and it takes lsp.SymbolInformation instead of a def key. In the
// future, a def landing page route scheme supporting multiple languages will
// be created.
func (r *Router) URLToLegacyDefLanding(s lsp.SymbolInformation) (string, error) {
	uri, err := url.Parse(s.Location.URI)
	if err != nil {
		return "", err
	}

	defPath := s.Name
	if s.ContainerName != "" {
		defPath = s.ContainerName + "/" + s.Name
	}

	repo := uri.Host + uri.Path
	unit := uri.Host + path.Join(uri.Path, path.Dir(uri.Fragment))
	if repo == "github.com/golang/go" {
		// Special case golang/go to emit just "encoding/json" for the path "github.com/golang/go/src/encoding/json"
		unit = strings.TrimPrefix(path.Dir(uri.Fragment), "src/")
	}

	return r.URLToDefLanding(graph.DefKey{
		Repo:     repo,
		CommitID: "",
		UnitType: "GoPackage",
		Unit:     unit,
		Path:     defPath,
	}).String(), nil
}

func (r *Router) urlToDef(repo, rev, unitType, unit, path string) *url.URL {
	return &url.URL{
		Path: fmt.Sprintf("/%s%s/-/def/%s/%s/-/%s", repo, rev, unitType, unit, path),
	}
}

func revStr(rev string) string {
	if rev == "" || strings.HasPrefix(rev, "@") {
		return rev
	}
	return "@" + rev
}
