package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

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
