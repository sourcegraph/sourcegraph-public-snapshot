package store

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

//
// Scans `[]shared.PackageDepencency`

var scanPackageDependencies = basestore.NewSliceScanner(func(s dbutil.Scanner) (shared.PackageDependency, error) {
	var v shared.PackageDependencyLiteral
	err := s.Scan(&v.RepoNameValue, &v.GitTagFromVersionValue, &v.SchemeValue, &v.PackageSyntaxValue, &v.PackageVersionValue)
	return v, err
})

//
// Scans `[]api.RepoCommit`

var scanRepoCommits = basestore.NewSliceScanner(func(s dbutil.Scanner) (api.RepoCommit, error) {
	var v api.RepoCommit
	err := s.Scan(&v.Repo, &v.CommitID)
	return v, err
})

//
// Scans `map[api.RepoName]types.RevSpecSet`

var scanRepoRevSpecSets = basestore.NewKeyedCollectionScanner[api.RepoName, api.RevSpec, types.RevSpecSet](scanRepoNameRevSpecPair, revSpecSetReducer{})

var scanRepoNameRevSpecPair = func(s dbutil.Scanner) (repoName api.RepoName, revSpec api.RevSpec, _ error) {
	err := s.Scan(&repoName, &revSpec)
	return repoName, revSpec, err
}

type revSpecSetReducer struct{}

func (r revSpecSetReducer) Create() types.RevSpecSet {
	return types.RevSpecSet{}
}

func (r revSpecSetReducer) Reduce(collection types.RevSpecSet, value api.RevSpec) types.RevSpecSet {
	collection[value] = struct{}{}
	return collection
}

//
// Scans `shared.Repo`

func scanDependencyRepo(s dbutil.Scanner) (shared.Repo, error) {
	var v shared.Repo
	err := s.Scan(&v.ID, &v.Scheme, &v.Name, &v.Version)
	return v, err
}

//
// Scans `[]shared.Repo`

var scanDependencyRepos = basestore.NewSliceScanner(scanDependencyRepo)
