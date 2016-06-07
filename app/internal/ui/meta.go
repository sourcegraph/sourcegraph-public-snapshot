package ui

import (
	"fmt"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// meta holds document metadata about the HTML page to be rendered in
// the initial HTTP response for the page.
type meta struct {
	// Title is the page's title (in the <title> tag). If empty,
	// "Sourcegraph" is used.
	Title string
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
		Title: fmt.Sprintf("%s: %s", trimRepo(repo.URI), desc),
	}
}

func defMeta(def *sourcegraph.Def, repo string) *meta {
	f := graph.PrintFormatter(&def.Def)
	return &meta{
		Title: repoPageTitle(repo, f.Name("dep")),
	}
}
