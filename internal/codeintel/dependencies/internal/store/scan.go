package store

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func scanDependencyRepoWithVersions(s dbutil.Scanner) (shared.PackageRepoReference, error) {
	var ref shared.PackageRepoReference
	var (
		versionStrings []string
		ids            []int64
		blocked        []bool
		lastCheckedAt  []sql.NullString
	)
	err := s.Scan(
		&ref.ID,
		&ref.Scheme,
		&ref.Name,
		&ref.Blocked,
		&ref.LastCheckedAt,
		pq.Array(&ids),
		pq.Array(&versionStrings),
		pq.Array(&blocked),
		pq.Array(&lastCheckedAt),
	)
	if err != nil {
		return shared.PackageRepoReference{}, err
	}

	ref.Versions = make([]shared.PackageRepoRefVersion, 0, len(versionStrings))
	for i, version := range versionStrings {
		// because pq.Array(&[]pq.NullTime) isnt supported...
		var t *time.Time
		if lastCheckedAt[i].Valid {
			parsedT, err := pq.ParseTimestamp(nil, lastCheckedAt[i].String)
			if err != nil {
				return shared.PackageRepoReference{}, errors.Wrapf(err, "time string %q is not valid", lastCheckedAt[i].String)
			}
			t = &parsedT
		}
		ref.Versions = append(ref.Versions, shared.PackageRepoRefVersion{
			ID:            int(ids[i]),
			PackageRefID:  ref.ID,
			Version:       version,
			Blocked:       blocked[i],
			LastCheckedAt: t,
		})
	}
	return ref, err
}

func scanPackageFilter(s dbutil.Scanner) (shared.PackageRepoFilter, error) {
	var filter shared.PackageRepoFilter
	var data []byte
	err := s.Scan(
		&filter.ID,
		&filter.Behaviour,
		&filter.PackageScheme,
		&data,
		&filter.DeletedAt,
		&filter.UpdatedAt,
	)
	if err != nil {
		return shared.PackageRepoFilter{}, err
	}

	b := bytes.NewReader(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()

	if err := d.Decode(&filter.NameFilter); err != nil {
		// d.Decode will set NameFilter to != nil even if theres an error, meaning we potentially
		// have both NameFilter and VersionFilter set to not nil
		filter.NameFilter = nil
		b.Seek(0, 0)
		return filter, d.Decode(&filter.VersionFilter)
	}

	return filter, nil
}
