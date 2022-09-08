package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetStaleSourcedCommits returns a set of commits attached to repositories that have been
// least recently checked for resolvability via gitserver. We do this periodically in
// order to determine which records in the database are unreachable by normal query
// paths and clean up that occupied (but useless) space. The output is of this method is
// ordered by repository ID then by commit.
func (s *store) GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, trace, endObservation := s.operations.getStaleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	now = now.UTC()
	interval := int(minimumTimeSinceLastCheck / time.Second)
	uploadSubquery := sqlf.Sprintf(staleSourcedCommitsSubquery, now, interval)
	query := sqlf.Sprintf(staleSourcedCommitsQuery, uploadSubquery, limit)

	sourcedCommits, err := scanSourcedCommits(tx.Query(ctx, query))
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
-- source: internal/codeintel/uploads/internal/store/store_commits.go:StaleSourcedCommits
WITH
	candidates AS (%s)
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
FROM lsif_uploads
WHERE
	-- Ignore records already marked as deleted
	state NOT IN ('deleted', 'deleting') AND
	-- Ignore records that have been checked recently. Note this condition is
	-- true for a null commit_last_checked_at (which has never been checked).
	(%s - commit_last_checked_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE
GROUP BY repository_id, commit
`

// UpdateSourcedCommits updates the commit_last_checked_at field of each upload records belonging to
// the given repository identifier and commit. This method returns the count of upload records modified
func (s *store) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error) {
	ctx, trace, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	candidateUploadsSubquery := sqlf.Sprintf(candidateUploadsCTE, repositoryID, commit)
	updateSourcedCommitQuery := sqlf.Sprintf(updateSourcedCommitsQuery, candidateUploadsSubquery, now)

	uploadsUpdated, err = scanCount(s.db.Query(ctx, updateSourcedCommitQuery))
	if err != nil {
		return 0, err
	}
	trace.Log(log.Int("uploadsUpdated", uploadsUpdated))

	return uploadsUpdated, nil
}

const updateSourcedCommitsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_commits.go:UpdateSourcedCommits
WITH
candidate_uploads AS (%s),
update_uploads AS (
	UPDATE lsif_uploads u
	SET commit_last_checked_at = %s
	WHERE id IN (SELECT id FROM candidate_uploads)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM update_uploads) AS num_uploads
`

const candidateUploadsCTE = `
SELECT u.id, u.state, u.uploaded_at
FROM lsif_uploads u
WHERE u.repository_id = %s AND u.commit = %s

-- Lock these rows in a deterministic order so that we don't
-- deadlock with other processes updating the lsif_uploads table.
ORDER BY u.id FOR UPDATE
`

// DeleteSourcedCommits deletes each upload record belonging to the given repository identifier
// and commit. Uploads are soft deleted. This method returns the count of upload modified.
//
// If a maximum commit lag is supplied, then any upload records in the uploading, queued, or processing states
// younger than the provided lag will not be deleted, but its timestamp will be modified as if the sibling method
// UpdateSourcedCommits was called instead. This configurable parameter enables support for remote code hosts
// that are not the source of truth; if we deleted all pending records without resolvable commits introduce races
// between the customer's Sourcegraph instance and their CI (and their CI will usually win).
func (s *store) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (
	uploadsUpdated, uploadsDeleted int,
	err error,
) {
	ctx, trace, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	unset, _ := s.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with unknown commit")
	defer unset(ctx)

	now = now.UTC()
	interval := int(maximumCommitLag / time.Second)

	candidateUploadsSubquery := sqlf.Sprintf(candidateUploadsCTE, repositoryID, commit)
	taggedCandidateUploadsSubquery := sqlf.Sprintf(taggedCandidateUploadsCTE, now, interval)
	deleteSourcedCommitsQuery := sqlf.Sprintf(deleteSourcedCommitsQuery, candidateUploadsSubquery, taggedCandidateUploadsSubquery, now)

	uploadsUpdated, uploadsDeleted, err = scanPairOfCounts(s.db.Query(ctx, deleteSourcedCommitsQuery))
	if err != nil {
		return 0, 0, err
	}
	trace.Log(
		log.Int("uploadsUpdated", uploadsUpdated),
		log.Int("uploadsDeleted", uploadsDeleted),
	)

	return uploadsUpdated, uploadsDeleted, nil
}

const deleteSourcedCommitsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_commits.go:DeleteSourcedCommits
WITH
candidate_uploads AS (%s),
tagged_candidate_uploads AS (%s),
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
)
SELECT
	(SELECT COUNT(*) FROM update_uploads) AS num_uploads_updated,
	(SELECT COUNT(*) FROM delete_uploads) AS num_uploads_deleted
`

const taggedCandidateUploadsCTE = `
SELECT
	u.*,
	(u.state IN ('uploading', 'queued', 'processing') AND %s - u.uploaded_at <= (%s * '1 second'::interval)) AS protected
FROM candidate_uploads u
`

// GetCommitsVisibleToUpload returns the set of commits for which the given upload can answer code intelligence queries.
// To paginate, supply the token returned from this method to the invocation for the next page.
func (s *store) GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error) {
	ctx, _, endObservation := s.operations.getCommitsVisibleToUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
		log.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	after := ""
	if token != nil {
		after = *token
	}

	commits, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(commitsVisibleToUploadQuery, strconv.Itoa(uploadID), after, limit)))
	if err != nil {
		return nil, nil, err
	}

	if len(commits) > 0 {
		last := commits[len(commits)-1]
		nextToken = &last
	}

	return commits, nextToken, nil
}

const commitsVisibleToUploadQuery = `
-- source: internal/codeintel/uploads/internal/store/store_commits.go:GetCommitsVisibleToUpload
WITH
direct_commits AS (
	SELECT nu.repository_id, nu.commit_bytea
	FROM lsif_nearest_uploads nu
	WHERE nu.uploads ? %s
),
linked_commits AS (
	SELECT ul.commit_bytea
	FROM direct_commits dc
	JOIN lsif_nearest_uploads_links ul
	ON
		ul.repository_id = dc.repository_id AND
		ul.ancestor_commit_bytea = dc.commit_bytea
),
combined_commits AS (
	SELECT dc.commit_bytea FROM direct_commits dc
	UNION ALL
	SELECT lc.commit_bytea FROM linked_commits lc
)
SELECT encode(c.commit_bytea, 'hex') as commit
FROM combined_commits c
WHERE decode(%s, 'hex') < c.commit_bytea
ORDER BY c.commit_bytea
LIMIT %s
`

type backfillIncompleteError struct {
	repositoryID int
}

func (e backfillIncompleteError) Error() string {
	return fmt.Sprintf("repository %d has not yet completed its backfill of column committed_at", e.repositoryID)
}

// GetOldestCommitDate returns the oldest commit date for all uploads for the given repository. If there are no
// non-nil values, a false-valued flag is returned. If there are any null values, the committed_at backfill job
// has not yet completed and an error is returned to prevent downstream expiration errors being made due to
// outdated commit graph data.
func (s *store) GetOldestCommitDate(ctx context.Context, repositoryID int) (_ time.Time, _ bool, err error) {
	ctx, _, endObservation := s.operations.getOldestCommitDate.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	t, ok, err := basestore.ScanFirstNullTime(s.db.Query(ctx, sqlf.Sprintf(getOldestCommitDateQuery, repositoryID)))
	if err != nil || !ok {
		return time.Time{}, false, err
	}
	if t == nil {
		return time.Time{}, false, &backfillIncompleteError{repositoryID}
	}

	return *t, true, nil
}

// Note: we check against '-infinity' here, as the backfill operation will use this sentinel value in the case
// that the commit is no longer know by gitserver. This allows the backfill migration to make progress without
// having pristine database.
const getOldestCommitDateQuery = `
-- source: internal/codeintel/uploads/internal/store/store_commits.go:GetOldestCommitDate
SELECT
	committed_at
FROM lsif_uploads
WHERE
	repository_id = %s AND
	state = 'completed' AND
	(committed_at != '-infinity' OR committed_at IS NULL)
ORDER BY committed_at NULLS FIRST
LIMIT 1
`

// HasCommit determines if the given commit is known for the given repository.
func (s *store) HasCommit(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservation := s.operations.hasCommit.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(
		ctx,
		sqlf.Sprintf(
			hasCommitQuery,
			repositoryID, dbutil.CommitBytea(commit),
			repositoryID, dbutil.CommitBytea(commit),
		),
	))

	return count > 0, err
}

const hasCommitQuery = `
-- source: internal/codeintel/stores/dbstore/commits.go:HasCommit
SELECT
	(SELECT COUNT(*) FROM lsif_nearest_uploads WHERE repository_id = %s AND commit_bytea = %s) +
	(SELECT COUNT(*) FROM lsif_nearest_uploads_links WHERE repository_id = %s AND commit_bytea = %s)
`

// CommitGraphMetadata returns whether or not the commit graph for the given repository is stale, along with the date of
// the most recent commit graph refresh for the given repository.
func (s *store) GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error) {
	ctx, _, endObservation := s.operations.getCommitGraphMetadata.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	updateToken, dirtyToken, updatedAt, exists, err := scanCommitGraphMetadata(s.db.Query(ctx, sqlf.Sprintf(commitGraphQuery, repositoryID)))
	if err != nil {
		return false, nil, err
	}
	if !exists {
		return false, nil, nil
	}

	return updateToken != dirtyToken, updatedAt, err
}

const commitGraphQuery = `
-- source: internal/codeintel/stores/dbstore/commits.go:CommitGraphMetadata
SELECT update_token, dirty_token, updated_at FROM lsif_dirty_repositories WHERE repository_id = %s LIMIT 1
`

// scanCommitGraphMetadata scans a a commit graph metadata row from the return value of `*Store.query`.
func scanCommitGraphMetadata(rows *sql.Rows, queryErr error) (updateToken, dirtyToken int, updatedAt *time.Time, _ bool, err error) {
	if queryErr != nil {
		return 0, 0, nil, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&updateToken, &dirtyToken, &updatedAt); err != nil {
			return 0, 0, nil, false, err
		}

		return updateToken, dirtyToken, updatedAt, true, nil
	}

	return 0, 0, nil, false, nil
}
