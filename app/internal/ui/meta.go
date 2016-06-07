package ui

import (
	"fmt"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/textutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// meta holds document metadata about the HTML page to be rendered in
// the initial HTTP response for the page.
type meta struct {
	// Title is the page's title. If empty, "Sourcegraph" is used.
	Title string

	// ShortTitle is used for the Open Graph and Twitter titles.
	ShortTitle string

	// Description is the page's description.
	Description string
}

// repoPageTitle produces the page title for a repo route or subroute
// page by joining the title component with the abbreviated repo name
// (e.g., "mydir/myfile.go · my/repo").
//
// NOTE: This should be (roughly) kept in sync with the page titles in
// JavaScript.
func repoPageTitle(repo, title string) string {
	repoTitle := trimRepo(repo)
	if title == "" {
		return repoTitle
	}
	return title + " · " + repoTitle
}

func trimRepo(repo string) string {
	return strings.TrimPrefix(strings.TrimPrefix(repo, "github.com/"), "sourcegraph.com/")
}

func repoMeta(repo *sourcegraph.Repo) *meta {
	desc := repo.Description
	if len(desc) > 40 {
		desc = desc[:40] + "..."
	}

	return &meta{
		Title:       fmt.Sprintf("%s: %s", trimRepo(repo.URI), desc),
		ShortTitle:  trimRepo(repo.URI),
		Description: repo.Description,
	}
}

func defMeta(def *sourcegraph.Def, repo string) *meta {
	var html string
	if def.DocHTML != nil {
		html = def.DocHTML.HTML
	}
	doc := textutil.TextFromHTML(html)
	if len(doc) > 250 {
		doc = doc[:250] + "..."
	}

	f := graph.PrintFormatter(&def.Def)
	return &meta{
		Title:       repoPageTitle(repo, f.Name("dep")),
		ShortTitle:  f.Name("dep"),
		Description: doc,
	}
}

func treeOrBlobMeta(path string, repo *sourcegraph.Repo) *meta {
	return &meta{
		Title:       repoPageTitle(repo.URI, path),
		ShortTitle:  path,
		Description: fmt.Sprintf("%s — %s", trimRepo(repo.URI), repo.Description),
	}
}
