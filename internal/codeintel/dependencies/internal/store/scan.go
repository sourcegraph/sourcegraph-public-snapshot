package store

import (
	"bytes"
	"encoding/json"

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

func scanPackageFilter(s dbutil.Scanner) (shared.PackageFilter, error) {
	var filter shared.PackageFilter
	var data []byte
	err := s.Scan(&filter.ID, &filter.Behaviour, &filter.ExternalService, &data)
	if err != nil {
		return shared.PackageFilter{}, err
	}

	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()

	if err := d.Decode(&filter.NameMatcher); err != nil {
		return filter, d.Decode(&filter.VersionMatcher)
	}

	return filter, nil
}
