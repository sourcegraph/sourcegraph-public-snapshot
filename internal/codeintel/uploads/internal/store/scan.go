package store

import (
	"database/sql"
	"sort"

	"github.com/jackc/pgtype"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanUploadByID(s dbutil.Scanner) (upload shared.Upload, err error) {
	return upload, s.Scan(
		&upload.ID,
	)
}

func scanCompleteUpload(s dbutil.Scanner) (upload shared.Upload, _ error) {
	var rawUploadedParts []sql.NullInt32
	if err := s.Scan(
		&upload.ID,
		&upload.Commit,
		&upload.Root,
		&upload.VisibleAtTip,
		&upload.UploadedAt,
		&upload.State,
		&upload.FailureMessage,
		&upload.StartedAt,
		&upload.FinishedAt,
		&upload.ProcessAfter,
		&upload.NumResets,
		&upload.NumFailures,
		&upload.RepositoryID,
		&upload.RepositoryName,
		&upload.Indexer,
		&dbutil.NullString{S: &upload.IndexerVersion},
		&upload.NumParts,
		pq.Array(&rawUploadedParts),
		&upload.UploadSize,
		&upload.AssociatedIndexID,
		&upload.Rank,
		&upload.UncompressedSize,
	); err != nil {
		return upload, err
	}

	upload.UploadedParts = make([]int, 0, len(rawUploadedParts))
	for _, uploadedPart := range rawUploadedParts {
		upload.UploadedParts = append(upload.UploadedParts, int(uploadedPart.Int32))
	}

	return upload, nil
}

var (
	scanUploadComplete = basestore.NewSliceScanner(scanCompleteUpload)
	scanUploads        = basestore.NewSliceScanner(scanUploadByID)
)

// scanFirstUpload scans a slice of uploads from the return value of `*Store.query` and returns the first.
var scanFirstUpload = basestore.NewFirstScanner(scanCompleteUpload)

var scanUploadsWithCount = basestore.NewSliceWithCountScanner(scanUploadWithCount)

func scanUploadWithCount(s dbutil.Scanner) (upload shared.Upload, count int, err error) {
	var rawUploadedParts []sql.NullInt32
	if err = s.Scan(
		&upload.ID,
		&upload.Commit,
		&upload.Root,
		&upload.VisibleAtTip,
		&upload.UploadedAt,
		&upload.State,
		&upload.FailureMessage,
		&upload.StartedAt,
		&upload.FinishedAt,
		&upload.ProcessAfter,
		&upload.NumResets,
		&upload.NumFailures,
		&upload.RepositoryID,
		&upload.RepositoryName,
		&upload.Indexer,
		&dbutil.NullString{S: &upload.IndexerVersion},
		&upload.NumParts,
		pq.Array(&rawUploadedParts),
		&upload.UploadSize,
		&upload.AssociatedIndexID,
		&upload.Rank,
		&upload.UncompressedSize,
		&count,
	); err != nil {
		return upload, 0, err
	}

	upload.UploadedParts = make([]int, 0, len(rawUploadedParts))
	for _, uploadedPart := range rawUploadedParts {
		upload.UploadedParts = append(upload.UploadedParts, int(uploadedPart.Int32))
	}

	return upload, count, nil
}

// scanCounts scans pairs of id/counts from the return value of `*Store.query`.
func scanCounts(rows *sql.Rows, queryErr error) (_ map[int]int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}

		visibilities[id] = count
	}

	return visibilities, nil
}

// scanDumps scans a slice of dumps from the return value of `*Store.query`.
func scanDump(s dbutil.Scanner) (dump shared.Dump, err error) {
	return dump, s.Scan(
		&dump.ID,
		&dump.Commit,
		&dump.Root,
		&dump.VisibleAtTip,
		&dump.UploadedAt,
		&dump.State,
		&dump.FailureMessage,
		&dump.StartedAt,
		&dump.FinishedAt,
		&dump.ProcessAfter,
		&dump.NumResets,
		&dump.NumFailures,
		&dump.RepositoryID,
		&dump.RepositoryName,
		&dump.Indexer,
		&dbutil.NullString{S: &dump.IndexerVersion},
		&dump.AssociatedIndexID,
	)
}

var scanDumps = basestore.NewSliceScanner(scanDump)

// scanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func scanSourcedCommits(rows *sql.Rows, queryErr error) (_ []shared.SourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]shared.SourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = shared.SourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]shared.SourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}

func scanCount(rows *sql.Rows, queryErr error) (value int, err error) {
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return 0, err
		}
	}

	return value, nil
}

func scanPairOfCounts(rows *sql.Rows, queryErr error) (value1, value2 int, err error) {
	if queryErr != nil {
		return 0, 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value1, &value2); err != nil {
			return 0, 0, err
		}
	}

	return value1, value2, nil
}

func scanIntPairs(rows *sql.Rows, queryErr error) (_ map[int]int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	values := map[int]int{}
	for rows.Next() {
		var value1 int
		var value2 int
		if err := rows.Scan(&value1, &value2); err != nil {
			return nil, err
		}

		values[value1] = value2
	}

	return values, nil
}

// scanCommitGraphView scans a commit graph view from the return value of `*Store.query`.
func scanCommitGraphView(rows *sql.Rows, queryErr error) (_ *commitgraph.CommitGraphView, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	commitGraphView := commitgraph.NewCommitGraphView()

	for rows.Next() {
		var meta commitgraph.UploadMeta
		var commit, token string

		if err := rows.Scan(&meta.UploadID, &commit, &token, &meta.Distance); err != nil {
			return nil, err
		}

		commitGraphView.Add(meta, commit, token)
	}

	return commitGraphView, nil
}

func scanRepoNames(rows *sql.Rows, queryErr error) (_ map[int]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	names := map[int]string{}

	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		names[id] = name
	}

	return names, nil
}

func scanUploadAuditLog(s dbutil.Scanner) (log shared.UploadLog, _ error) {
	hstores := pgtype.HstoreArray{}
	err := s.Scan(
		&log.LogTimestamp,
		&log.RecordDeletedAt,
		&log.UploadID,
		&log.Commit,
		&log.Root,
		&log.RepositoryID,
		&log.UploadedAt,
		&log.Indexer,
		&log.IndexerVersion,
		&log.UploadSize,
		&log.AssociatedIndexID,
		&hstores,
		&log.Reason,
		&log.Operation,
	)

	for _, hstore := range hstores.Elements {
		m := make(map[string]*string)
		if err := hstore.AssignTo(&m); err != nil {
			return log, err
		}
		log.TransitionColumns = append(log.TransitionColumns, m)
	}

	return log, err
}

var scanUploadAuditLogs = basestore.NewSliceScanner(scanUploadAuditLog)
