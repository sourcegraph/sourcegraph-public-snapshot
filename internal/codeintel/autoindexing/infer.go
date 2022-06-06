package autoindexing

import (
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func InferRepositoryAndRevision(pkg precise.Package) (repoName api.RepoName, gitTagOrCommit string, ok bool) {
	for _, fn := range []func(pkg precise.Package) (api.RepoName, string, bool){
		inferGoRepositoryAndRevision,
		inferJVMRepositoryAndRevision,
		inferNpmRepositoryAndRevision,
		inferRustRepositoryAndRevision,
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
	if pkg.Scheme != dependencies.JVMPackagesScheme {
		return "", "", false
	}
	return api.RepoName(pkg.Name), "v" + pkg.Version, true
}

func inferNpmRepositoryAndRevision(pkg precise.Package) (api.RepoName, string, bool) {
	if pkg.Scheme != dependencies.NpmPackagesScheme {
		return "", "", false
	}
	npmPkg, err := reposource.ParseNpmPackageFromPackageSyntax(pkg.Name)
	if err != nil {
		log15.Error("invalid npm package name in database", "error", err)
		return "", "", false
	}
	return npmPkg.RepoName(), "v" + pkg.Version, true
}

func inferRustRepositoryAndRevision(pkg precise.Package) (api.RepoName, string, bool) {
	if pkg.Scheme != dependencies.RustPackagesScheme {
		return "", "", false
	}

	rustPkg, err := reposource.ParseRustDependency(pkg.Name)
	if err != nil {
		log15.Error("invalid rust package name in database", "error", err, "pkg", pkg.Name)
		return "", "", false
	}

	return rustPkg.RepoName(), "v" + pkg.Version, true
}
