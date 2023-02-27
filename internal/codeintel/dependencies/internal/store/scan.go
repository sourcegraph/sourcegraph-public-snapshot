package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanDependencyRepo(s dbutil.Scanner) (shared.PackageRepoReference, error) {
	var ref shared.PackageRepoReference
	var version shared.PackageRepoRefVersion
	err := s.Scan(&ref.ID, &ref.Scheme, &ref.Name, &version.ID, &version.PackageRefID, &version.Version)
	ref.Versions = []shared.PackageRepoRefVersion{version}
	return ref, err
}
