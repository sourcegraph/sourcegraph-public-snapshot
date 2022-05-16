package store

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var scanPackageDependencies = basestore.NewSliceScanner(func(s dbutil.Scanner) (shared.PackageDependency, error) {
	var v shared.PackageDependencyLiteral
	err := s.Scan(&v.RepoNameValue, &v.GitTagFromVersionValue, &v.SchemeValue, &v.PackageSyntaxValue, &v.PackageVersionValue)
	return v, err
})

var scanRepoCommits = basestore.NewSliceScanner(func(s dbutil.Scanner) (api.RepoCommit, error) {
	var v api.RepoCommit
	err := s.Scan(&v.Repo, &v.CommitID)
	return v, err
})

func scanDependencyRepo(s dbutil.Scanner) (shared.Repo, error) {
	var v shared.Repo
	err := s.Scan(&v.ID, &v.Scheme, &v.Name, &v.Version)
	return v, err
}

var scanDependencyRepos = basestore.NewSliceScanner(scanDependencyRepo)
