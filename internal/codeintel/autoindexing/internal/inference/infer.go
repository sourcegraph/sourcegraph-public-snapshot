package inference

import (
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

func InferRepositoryAndRevision(pkg dependencies.MinimialVersionedPackageRepo) (repoName api.RepoName, gitTagOrCommit string, ok bool) {
	for _, fn := range []func(pkg dependencies.MinimialVersionedPackageRepo) (api.RepoName, string, bool){
		inferGoRepositoryAndRevision,
		inferJVMRepositoryAndRevision,
		inferNpmRepositoryAndRevision,
		inferRustRepositoryAndRevision,
		inferPythonRepositoryAndRevision,
		inferRubyRepositoryAndRevision,
	} {
		if repoName, gitTagOrCommit, ok := fn(pkg); ok {
			return repoName, gitTagOrCommit, true
		}
	}

	return "", "", false
}

const GitHubScheme = "https://"

var goVersionPattern = lazyregexp.New(`^v?[\d\.]+-([a-f0-9]+)`)

func inferGoRepositoryAndRevision(pkg dependencies.MinimialVersionedPackageRepo) (api.RepoName, string, bool) {
	if pkg.Scheme != "gomod" || !strings.HasPrefix(string(pkg.Name), GitHubScheme+"github.com/") {
		return "", "", false
	}

	repoParts := strings.Split(string(pkg.Name[len(GitHubScheme):]), "/")[:3]
	if len(repoParts) > 3 {
		repoParts = repoParts[:3]
	}

	version := pkg.Version
	if match := goVersionPattern.FindAllStringSubmatch(version, 1); len(match) > 0 {
		version = match[0][1]
	}

	return api.RepoName(strings.Join(repoParts, "/")), version, true
}

func inferJVMRepositoryAndRevision(pkg dependencies.MinimialVersionedPackageRepo) (api.RepoName, string, bool) {
	if pkg.Scheme != dependencies.JVMPackagesScheme {
		return "", "", false
	}
	return api.RepoName(pkg.Name), "v" + pkg.Version, true
}

func inferNpmRepositoryAndRevision(pkg dependencies.MinimialVersionedPackageRepo) (api.RepoName, string, bool) {
	if pkg.Scheme != dependencies.NpmPackagesScheme {
		return "", "", false
	}

	logger := log.Scoped("inferNpmRepositoryAndRevision")
	npmPkg, err := reposource.ParseNpmPackageFromPackageSyntax(pkg.Name)
	if err != nil {
		logger.Error("invalid npm package name in database", log.Error(err))
		return "", "", false
	}
	return npmPkg.RepoName(), "v" + pkg.Version, true
}

func inferRustRepositoryAndRevision(pkg dependencies.MinimialVersionedPackageRepo) (api.RepoName, string, bool) {
	if pkg.Scheme != dependencies.RustPackagesScheme {
		return "", "", false
	}

	rustPkg := reposource.ParseRustVersionedPackage(string(pkg.Name))
	return rustPkg.RepoName(), "v" + pkg.Version, true
}

func inferPythonRepositoryAndRevision(pkg dependencies.MinimialVersionedPackageRepo) (api.RepoName, string, bool) {
	if pkg.Scheme != dependencies.PythonPackagesScheme {
		return "", "", false
	}

	pythonPkg := reposource.ParsePythonPackageFromName(pkg.Name)

	return pythonPkg.RepoName(), pkg.Version, true
}

func inferRubyRepositoryAndRevision(pkg dependencies.MinimialVersionedPackageRepo) (api.RepoName, string, bool) {
	if pkg.Scheme != dependencies.RubyPackagesScheme {
		return "", "", false
	}

	rubyPkg := reposource.ParseRubyPackageFromName(pkg.Name)

	return rubyPkg.RepoName(), pkg.Version, true
}
