package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanDependencyRepo(s dbutil.Scanner) (shared.Repo, error) {
	var v shared.Repo
	err := s.Scan(&v.ID, &v.Scheme, &v.Name, &v.Version)
	return v, err
}

var scanDependencyRepos = basestore.NewSliceScanner(scanDependencyRepo)
