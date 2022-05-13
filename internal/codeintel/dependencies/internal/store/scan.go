package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanDependencyRepo(s dbutil.Scanner) (dependencyRepo shared.Repo, err error) {
	return dependencyRepo, s.Scan(
		&dependencyRepo.ID,
		&dependencyRepo.Scheme,
		&dependencyRepo.Name,
		&dependencyRepo.Version,
	)
}

var scanDependencyRepos = basestore.NewSliceScanner(scanDependencyRepo)
