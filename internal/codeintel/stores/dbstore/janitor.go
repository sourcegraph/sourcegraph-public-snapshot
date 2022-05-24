package dbstore

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// ScanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func ScanSourcedCommits(rows *sql.Rows, queryErr error) (_ []SourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]SourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = SourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]SourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}

// StaleSourcedCommits returns a set of commits attached to repositories that have been
// least recently checked for resolvability via gitserver. We do this periodically in
// order to determine which records in the database are unreachable by normal query
// paths and clean up that occupied (but useless) space. The output is of this method is
// ordered by repository ID then by commit.
func (s *Store) StaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []SourcedCommits, err error) {
	ctx, trace, endObservation := s.operations.staleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	now = now.UTC()
	interval := int(minimumTimeSinceLastCheck / time.Second)
	uploadSubquery := sqlf.Sprintf(staleSourcedCommitsSubquery, sqlf.Sprintf("lsif_uploads"), now, interval)
	indexesSubquery := sqlf.Sprintf(staleSourcedCommitsSubquery, sqlf.Sprintf("lsif_indexes"), now, interval)

	sourcedCommits, err := ScanSourcedCommits(s.Store.Query(ctx, sqlf.Sprintf(staleSourcedCommitsQuery, uploadSubquery, indexesSubquery, limit)))
	if err != nil {
		return nil, err
	}

	numCommits := 0
	for _, commits := range sourcedCommits {
		numCommits += len(commits.Commits)
	}
	trace.Log(
		log.Int("numRepositories", len(sourcedCommits)),
		log.Int("numCommits", numCommits),
	)

	return sourcedCommits, nil
}

const staleSourcedCommitsQuery = `
-- source: internal/codeintel/stores/dbstore/janitor.go:StaleSourcedCommits
WITH
candidates AS (%s UNION %s)
SELECT r.id, r.name, c.commit
FROM candidates c
JOIN repo r ON r.id = c.repository_id
-- Order results so that the repositories with the commits that have been updated
-- the least frequently come first. Once a number of commits are processed from a
-- given repository the ordering may change.
ORDER BY MIN(c.max_last_checked_at) OVER (PARTITION BY c.repository_id), c.commit
LIMIT %s
`

const staleSourcedCommitsSubquery = `
SELECT
	repository_id,
	commit,
	-- Keep track of the most recent update of this commit that we know about
	-- as any earlier dates for the same repository and commit pair carry no
	-- useful information.
	MAX(commit_last_checked_at) as max_last_checked_at
FROM %s
WHERE
	-- Ignore records already marked as deleted
	state NOT IN ('deleted', 'deleting') AND
	-- Ignore records that have been checked recently. Note this condition is
	-- true for a null commit_last_checked_at (which has never been checked).
	(%s - commit_last_checked_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE
GROUP BY repository_id, commit
`

// UpdateSourcedCommits updates the commit_last_checked_at field of each upload and index records belonging
// to the given repository identifier and commit. This method returns the count of upload and index records
// modified, respectively.
func (s *Store) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, indexesUpdated int, err error) {
	ctx, trace, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	uploadsUpdated, indexesUpdated, err = scanPairOfCounts(s.Query(ctx, sqlf.Sprintf(
		updateSourcedCommitsQuery,
		repositoryID, commit, // candidate_uploads
		repositoryID, commit, // candidate_indexes
		now, now, // update_uploads, update_indexes
	)))
	if err != nil {
		return 0, 0, err
	}
	trace.Log(
		log.Int("uploadsUpdated", uploadsUpdated),
		log.Int("indexesUpdated", indexesUpdated),
	)

	return uploadsUpdated, indexesUpdated, nil
}

const updateSourcedCommitsQuery = `
-- source: internal/codeintel/stores/dbstore/janitor.go:UpdateSourcedCommits
WITH
` + sourcedCommitsCandidateUploadsCTE + `,
` + sourcedCommitsCandidateIndexesCTE + `,
update_uploads AS (
	UPDATE lsif_uploads u
	SET commit_last_checked_at = %s
	WHERE id IN (SELECT id FROM candidate_uploads)
	RETURNING 1
),
update_indexes AS (
	UPDATE lsif_indexes u
	SET commit_last_checked_at = %s
	WHERE id IN (SELECT id FROM candidate_indexes)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM update_uploads) AS num_uploads,
	(SELECT COUNT(*) FROM update_indexes) AS num_indexes
`

const sourcedCommitsCandidateUploadsCTE = `
candidate_uploads AS (
	SELECT u.id, u.state, u.uploaded_at
	FROM lsif_uploads u
	WHERE u.repository_id = %s AND u.commit = %s

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
)
`

const sourcedCommitsCandidateIndexesCTE = `
candidate_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE u.repository_id = %s AND u.commit = %s

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_indexes table.
	ORDER BY u.id FOR UPDATE
)
`

// DeleteSourcedCommits deletes each upload and index records belonging to the given repository identifier
// and commit. Uploads are soft deleted and indexes are hard-deleted. This method returns the count of upload
// and index records modified.
//
// If a maximum commit lag is supplied, then any upload records in the uploading, queued, or processing states
// younger than the provided lag will not be deleted, but its timestamp will be modified as if the sibling method
// UpdateSourcedCommits was called instead. This configurable parameter enables support for remote code hosts
// that are not the source of truth; if we deleted all pending records without resolvable commits introduce races
// between the customer's Sourcegraph instance and their CI (and their CI will usually win).
func (s *Store) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (
	uploadsUpdated int,
	uploadsDeleted int,
	indexesDeleted int,
	err error,
) {
	ctx, trace, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	now = now.UTC()
	interval := int(maximumCommitLag / time.Second)

	unset, _ := s.Store.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with unknown commit")
	defer unset(ctx)
	uploadsUpdated, uploadsDeleted, indexesDeleted, err = scanTripleOfCounts(s.Query(ctx, sqlf.Sprintf(
		deleteSourcedCommitsQuery,
		repositoryID, commit, // candidate_uploads
		repositoryID, commit, // candidate_indexes
		now, interval, // tagged_candidate_uploads
		now, // update_uploads
	)))
	if err != nil {
		return 0, 0, 0, err
	}
	trace.Log(
		log.Int("uploadsUpdated", uploadsUpdated),
		log.Int("uploadsDeleted", uploadsDeleted),
		log.Int("indexesDeleted", indexesDeleted),
	)

	return uploadsUpdated, uploadsDeleted, indexesDeleted, nil
}

const deleteSourcedCommitsQuery = `
-- source: internal/codeintel/stores/dbstore/janitor.go:DeleteSourcedCommits
WITH
` + sourcedCommitsCandidateUploadsCTE + `,
` + sourcedCommitsCandidateIndexesCTE + `,
tagged_candidate_uploads AS (
	SELECT
		u.*,
		(u.state IN ('uploading', 'queued', 'processing') AND %s - u.uploaded_at <= (%s * '1 second'::interval)) AS protected
	FROM candidate_uploads u
),
update_uploads AS (
	UPDATE lsif_uploads u
	SET commit_last_checked_at = %s
	WHERE EXISTS (SELECT 1 FROM tagged_candidate_uploads tu WHERE tu.id = u.id AND tu.protected)
	RETURNING 1
),
delete_uploads AS (
	UPDATE lsif_uploads u
	SET state = CASE WHEN u.state = 'completed' THEN 'deleting' ELSE 'deleted' END
	WHERE EXISTS (SELECT 1 FROM tagged_candidate_uploads tu WHERE tu.id = u.id AND NOT tu.protected)
	RETURNING 1
),
delete_indexes AS (
	DELETE FROM lsif_indexes u
	WHERE id IN (SELECT id FROM candidate_indexes)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM update_uploads) AS num_uploads_updated,
	(SELECT COUNT(*) FROM delete_uploads) AS num_uploads_deleted,
	(SELECT COUNT(*) FROM delete_indexes) AS num_indexes_deleted
`

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

func scanTripleOfCounts(rows *sql.Rows, queryErr error) (value1, value2, value3 int, err error) {
	if queryErr != nil {
		return 0, 0, 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value1, &value2, &value3); err != nil {
			return 0, 0, 0, err
		}
	}

	return value1, value2, value3, nil
}

// DeleteOldAuditLogs removes lsif_upload audit log records older than the given max age.
func (s *Store) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (_ int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(deleteOldAuditLogsQuery, now, int(maxAge/time.Second))))
	return count, err
}

const deleteOldAuditLogsQuery = `
-- source: internal/codeintel/stores/dbstore/janitor.go:DeleteOldAuditLogs
WITH deleted AS (
	DELETE FROM lsif_uploads_audit_logs
	WHERE %s - log_timestamp > (%s * '1 second'::interval)
	RETURNING upload_id
)
SELECT count(*) FROM deleted
`
