package enqueuer

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func InferRepositoryAndRevision(pkg precise.Package) (repoName api.RepoName, gitTagOrCommit string, ok bool) {
	for _, fn := range []func(pkg precise.Package) (api.RepoName, string, bool){
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

func inferGoRepositoryAndRevision(pkg precise.Package) (api.RepoName, string, bool) {
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

	return api.RepoName(strings.Join(repoParts, "/")), version, true
}

func inferJVMRepositoryAndRevision(pkg precise.Package) (api.RepoName, string, bool) {
	if pkg.Scheme != dbstore.JVMPackagesScheme {
		return "", "", false
	}
	// TODO: [Varun] Is this correct, or do we need extra steps like for NPM?
	return api.RepoName(pkg.Name), "v" + pkg.Version, true
}

func inferNPMRepositoryAndRevision(pkg precise.Package) (api.RepoName, string, bool) {
	if pkg.Scheme != dbstore.NPMPackagesScheme {
		return "", "", false
	}
	// TODO: [Varun] What should we do about the error here?
	npmPkg, _ := reposource.ParseNPMPackageFromPackageSyntax(pkg.Name)
	return npmPkg.RepoName(), "v" + pkg.Version, true
}
