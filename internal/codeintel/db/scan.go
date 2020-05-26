package db

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// scanner is the common interface shared by *sql.Row and *sql.Rows.
type scanner interface {
	// Scan copies the values of the current row into the values pointed at by dest.
	Scan(dest ...interface{}) error
}

// scanDump populates a Dump value from the given scanner.
func scanDump(scanner scanner) (dump Dump, err error) {
	err = scanner.Scan(
		&dump.ID,
		&dump.Commit,
		&dump.Root,
		&dump.VisibleAtTip,
		&dump.UploadedAt,
		&dump.State,
		&dump.FailureSummary,
		&dump.FailureStacktrace,
		&dump.StartedAt,
		&dump.FinishedAt,
		&dump.RepositoryID,
		&dump.Indexer,
	)
	return dump, err
}

// scanDumps reads the given set of dump rows and returns a slice of resulting values.
// This method should be called directly with the return value of `*db.query`.
func scanDumps(rows *sql.Rows, err error) ([]Dump, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dumps []Dump
	for rows.Next() {
		dump, err := scanDump(rows)
		if err != nil {
			return nil, err
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}

// scanFirstDump reads the given set of dump rows and returns the first value and a
// boolean flag indicating its presence. This method should be called directly with
// the return value of `*db.query`.
func scanFirstDump(rows *sql.Rows, err error) (Dump, bool, error) {
	dumps, err := scanDumps(rows, err)
	if err != nil || len(dumps) == 0 {
		return Dump{}, false, err
	}
	return dumps[0], true, nil
}

// scanUpload populates an Upload value from the given scanner.
func scanUpload(scanner scanner) (upload Upload, err error) {
	var rawUploadedParts []sql.NullInt32
	err = scanner.Scan(
		&upload.ID,
		&upload.Commit,
		&upload.Root,
		&upload.VisibleAtTip,
		&upload.UploadedAt,
		&upload.State,
		&upload.FailureSummary,
		&upload.FailureStacktrace,
		&upload.StartedAt,
		&upload.FinishedAt,
		&upload.RepositoryID,
		&upload.Indexer,
		&upload.NumParts,
		pq.Array(&rawUploadedParts),
		&upload.Rank,
	)

	var uploadedParts = []int{}
	for _, uploadedPart := range rawUploadedParts {
		uploadedParts = append(uploadedParts, int(uploadedPart.Int32))
	}
	upload.UploadedParts = uploadedParts

	return upload, err
}

// scanUploads reads the given set of upload rows and returns a slice of resulting
// values. This method should be called directly with the return value of `*db.query`.
func scanUploads(rows *sql.Rows, err error) ([]Upload, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []Upload
	for rows.Next() {
		upload, err := scanUpload(rows)
		if err != nil {
			return nil, err
		}

		uploads = append(uploads, upload)
	}

	return uploads, nil
}

// scanFirstUpload reads the given set of upload rows and returns the first value and
// a boolean flag indicating its presence. This method should be called directly with
// the return value of `*db.query`.
func scanFirstUpload(rows *sql.Rows, err error) (Upload, bool, error) {
	uploads, err := scanUploads(rows, err)
	if err != nil || len(uploads) == 0 {
		return Upload{}, false, err
	}
	return uploads[0], true, nil
}

// scanFirstUploadDequeue is scanFirstUpload with an interface return value.
func scanFirstUploadDequeue(rows *sql.Rows, err error) (interface{}, bool, error) {
	return scanFirstUpload(rows, err)
}

// scanPackageReference populates a package reference value from the given scanner.
func scanPackageReference(scanner scanner) (reference types.PackageReference, err error) {
	err = scanner.Scan(&reference.DumpID, &reference.Scheme, &reference.Name, &reference.Version, &reference.Filter)
	return reference, err
}

// scanPackageReferences reads the given set of reference rows and returns a slice of resulting
// values. This method should be called directly with the return value of `*db.queryRows`.
func scanPackageReferences(rows *sql.Rows, err error) ([]types.PackageReference, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []types.PackageReference
	for rows.Next() {
		reference, err := scanPackageReference(rows)
		if err != nil {
			return nil, err
		}

		references = append(references, reference)
	}

	return references, nil
}

// scanString populates a string value from the given scanner.
func scanString(scanner scanner) (value string, err error) {
	err = scanner.Scan(&value)
	return value, err
}

// scanStrings reads the given set of `(string)` rows and returns a slice of resulting
// values. This method should be called directly with the return value of `*db.query`.
func scanStrings(rows *sql.Rows, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		value, err := scanString(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// scanFirstString reads the given set of `(string)` rows and returns the first value
// and a boolean flag indicating its presence. This method should be called directly
// with the return value of `*db.query`.
func scanFirstString(rows *sql.Rows, err error) (string, bool, error) {
	values, err := scanStrings(rows, err)
	if err != nil || len(values) == 0 {
		return "", false, err
	}
	return values[0], true, nil
}

// scanInt populates an integer value from the given scanner.
func scanInt(scanner scanner) (value int, err error) {
	err = scanner.Scan(&value)
	return value, err
}

// scanInts reads the given set of `(int)` rows and returns a slice of resulting values.
// This method should be called directly with the return value of `*db.query`.
func scanInts(rows *sql.Rows, err error) ([]int, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []int
	for rows.Next() {
		value, err := scanInt(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// scanFirstInt reads the given set of `(int)` rows and returns the first value and a
// boolean flag indicating its presence. This method should be called directly with
// the return value of `*db.query`.
func scanFirstInt(rows *sql.Rows, err error) (int, bool, error) {
	values, err := scanInts(rows, err)
	if err != nil || len(values) == 0 {
		return 0, false, err
	}
	return values[0], true, nil
}

// scanState populates an integer and string from the given scanner.
func scanState(scanner scanner) (id int, state string, err error) {
	err = scanner.Scan(&id, &state)
	return id, state, err
}

// scanStates reads the given set of `(id, state)` rows and returns a map from id to its
// state. This method should be called directly with the return value of `*db.query`.
func scanStates(rows *sql.Rows, err error) (map[int]string, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := map[int]string{}
	for rows.Next() {
		id, state, err := scanState(rows)
		if err != nil {
			return nil, err
		}

		states[id] = state
	}

	return states, nil
}

// scanVisibility populates an integer and boolean from the given scanner.
func scanVisibility(scanner scanner) (id int, visibleAtTip bool, err error) {
	err = scanner.Scan(&id, &visibleAtTip)
	return id, visibleAtTip, err
}

// scanVisibilities reads the given set of `(id, visible_at_tip)` rows and returns a map
// from id to its visibility. This method should be called directly with the return value
// of `*db.query`.
func scanVisibilities(rows *sql.Rows, err error) (map[int]bool, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	visibilities := map[int]bool{}
	for rows.Next() {
		id, visibleAtTip, err := scanVisibility(rows)
		if err != nil {
			return nil, err
		}

		visibilities[id] = visibleAtTip
	}

	return visibilities, nil
}

// scanCommit populates a pair of strings from the given scanner.
func scanCommit(scanner scanner) (commit string, parentCommit *string, err error) {
	err = scanner.Scan(&commit, &parentCommit)
	return commit, parentCommit, err
}

// scanCommits reads the given set of `(commit, parent_commit)` rows and returns
// a map from commits to its parents. This method should be called directly from
// the return value of `*db.query`.
func scanCommits(rows *sql.Rows, err error) (map[string][]string, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commits := map[string][]string{}
	for rows.Next() {
		commit, parentCommit, err := scanCommit(rows)
		if err != nil {
			return nil, err
		}

		if _, ok := commits[commit]; !ok {
			commits[commit] = nil
		}

		if parentCommit != nil {
			commits[commit] = append(commits[commit], *parentCommit)
		}
	}

	return commits, nil
}

// scanIndexableRepository populates an IndexableRepository value from the given scanner.
func scanIndexableRepository(scanner scanner) (indexableRepository IndexableRepository, err error) {
	err = scanner.Scan(
		&indexableRepository.RepositoryID,
		&indexableRepository.SearchCount,
		&indexableRepository.PreciseCount,
		&indexableRepository.LastIndexEnqueuedAt,
	)
	return indexableRepository, err
}

// scanIndexableRepositories reads the given set of indexable repository rows and returns
// a slice of resulting values. This method should be called directly with the return value
// of `*db.query`.
func scanIndexableRepositories(rows *sql.Rows, err error) ([]IndexableRepository, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexableRepositories []IndexableRepository
	for rows.Next() {
		indexableRepository, err := scanIndexableRepository(rows)
		if err != nil {
			return nil, err
		}

		indexableRepositories = append(indexableRepositories, indexableRepository)
	}

	return indexableRepositories, nil
}

// scanIndex populates an Index value from the given scanner.
func scanIndex(scanner scanner) (index Index, err error) {
	err = scanner.Scan(
		&index.ID,
		&index.Commit,
		&index.QueuedAt,
		&index.State,
		&index.FailureSummary,
		&index.FailureStacktrace,
		&index.StartedAt,
		&index.FinishedAt,
		&index.RepositoryID,
		&index.Rank,
	)
	return index, err
}

// scanIndexes reads the given set of index rows and returns a slice of resulting
// values. This method should be called directly with the return value of `*db.query`.
func scanIndexes(rows *sql.Rows, err error) ([]Index, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []Index
	for rows.Next() {
		index, err := scanIndex(rows)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

// scanFirstIndex reads the given set of index rows and returns the first value and
// a boolean flag indicating its presence. This method should be called directly with
// the return value of `*db.query`.
func scanFirstIndex(rows *sql.Rows, err error) (Index, bool, error) {
	indexes, err := scanIndexes(rows, err)
	if err != nil || len(indexes) == 0 {
		return Index{}, false, err
	}
	return indexes[0], true, nil
}

// scanFirstIndexDequeue is scanFirstIndex with an interface return value.
func scanFirstIndexDequeue(rows *sql.Rows, err error) (interface{}, bool, error) {
	return scanFirstIndex(rows, err)
}

// scanRepoUsageStatistics populates a RepoUsageStatistics from the given scanner.
func scanRepoUsageStatistics(scanner scanner) (stats RepoUsageStatistics, err error) {
	err = scanner.Scan(&stats.RepositoryID, &stats.SearchCount, &stats.PreciseCount)
	return stats, err
}

// scanRepoUsageStatisticsSlice reads the given set of repo usage stat rows and returns
// a slice of RepoUsageStatistics values. This method should be called directly from the
// return value of `*db.query`.
func scanRepoUsageStatisticsSlice(rows *sql.Rows, err error) ([]RepoUsageStatistics, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []RepoUsageStatistics
	for rows.Next() {
		s, err := scanRepoUsageStatistics(rows)
		if err != nil {
			return nil, err
		}

		stats = append(stats, s)
	}

	return stats, nil
}
