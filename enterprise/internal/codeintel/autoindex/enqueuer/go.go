package enqueuer

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

const GitHubScheme = "https://"

var goVersionPattern = lazyregexp.New(`^v?[\d\.]+-([a-f0-9]+)`)

func InferGoRepositoryAndRevision(pkg precise.Package) (repoName, gitTagOrCommit string, ok bool) {
	if pkg.Scheme != "gomod" || !strings.HasPrefix(pkg.Name, GitHubScheme+"github.com/") {
		return "", "", false
	}

	repoParts := strings.Split(pkg.Name[len(GitHubScheme):], "/")[:3]
	if len(repoParts) > 3 {
		repoParts = repoParts[:3]
	}

	version := pkg.Version
	if match := goVersionPattern.FindAllStringSubmatch(version, 1); len(match) > 0 {
		version = match[0][1]
	}

	return strings.Join(repoParts, "/"), version, true
}
