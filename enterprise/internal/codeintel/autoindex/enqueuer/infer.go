package enqueuer

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func InferRepositoryAndRevision(pkg precise.Package) (repoName, gitTagOrCommit string, ok bool) {
	for _, fn := range []func(pkg precise.Package) (string, string, bool){
		inferGoRepositoryAndRevision,
		inferJVMRepositoryAndRevision,
		inferNPMRepositoryAndRevision,
	} {
		if repoName, gitTagOrCommit, ok := fn(pkg); ok {
			return repoName, gitTagOrCommit, true
		}
	}

	return "", "", false
}

const GitHubScheme = "https://"

var goVersionPattern = lazyregexp.New(`^v?[\d\.]+-([a-f0-9]+)`)

func inferGoRepositoryAndRevision(pkg precise.Package) (string, string, bool) {
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

func inferJVMRepositoryAndRevision(pkg precise.Package) (string, string, bool) {
	if pkg.Scheme != dbstore.JVMPackagesScheme {
		return "", "", false
	}
	return pkg.Name, "v" + pkg.Version, true
}

func inferNPMRepositoryAndRevision(pkg precise.Package) (string, string, bool) {
	if pkg.Scheme != dbstore.NPMPackagesScheme {
		return "", "", false
	}
	return pkg.Name, "v" + pkg.Version, true
}
