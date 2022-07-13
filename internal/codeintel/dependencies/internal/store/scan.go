package store

import (
	"database/sql"

	"github.com/lib/pq"

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

// scanLockfileIndex scans `shared.LockfileIndex`
func scanLockfileIndex(s dbutil.Scanner) (shared.LockfileIndex, error) {
	var (
		i            shared.LockfileIndex
		commit       dbutil.CommitBytea
		referenceIDs pq.Int32Array
	)

	err := s.Scan(&i.ID, &i.RepositoryID, &commit, &referenceIDs, &i.Lockfile, &i.Fidelity)
	if err != nil {
		return i, err
	}

	i.Commit = string(commit)

	i.LockfileReferenceIDs = make([]int, 0, len(referenceIDs))
	for _, id := range referenceIDs {
		i.LockfileReferenceIDs = append(i.LockfileReferenceIDs, int(id))
	}

	return i, nil
}

// scanLockfileIndexes scans `[]shared.LockfileIndex`
var scanLockfileIndexes = basestore.NewSliceScanner(scanLockfileIndex)

// scanIntString scans a int, string pair.
func scanIntString(s dbutil.Scanner) (int, string, error) {
	var (
		i   int
		str string
	)

	if err := s.Scan(&i, &str); err != nil {
		return 0, "", err
	}

	return i, str, nil
}

func scanIdNames(rows *sql.Rows, queryErr error) (nameIDs map[string]int, ids []int, err error) {
	if queryErr != nil {
		return nil, nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	nameIDs = make(map[string]int)
	ids = []int{}

	for rows.Next() {
		id, name, err := scanIntString(rows)
		if err != nil {
			return nil, nil, err
		}

		nameIDs[name] = id
		ids = append(ids, id)
	}

	return nameIDs, ids, nil
}
