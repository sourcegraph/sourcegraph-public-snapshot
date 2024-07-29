package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) GetIndexers(ctx context.Context, opts shared.GetIndexersOptions) (_ []string, err error) {
	ctx, _, endObservation := s.operations.getIndexers.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", opts.RepositoryID),
	}})
	defer endObservation(1, observation.Args{})

	if opts.RepositoryID == 0 {
		return basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(getGlobalIndexersQuery)))
	}

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, err
	}

	return basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(getIndexersForRepositoryQuery, opts.RepositoryID, authzConds)))
}

const getGlobalIndexersQuery = `
WITH
combined_indexers AS (
	SELECT DISTINCT u.indexer FROM lsif_uploads u
	UNION ALL
	SELECT DISTINCT u.indexer FROM lsif_indexes u
)
SELECT DISTINCT u.indexer
FROM combined_indexers u
ORDER BY u.indexer
`

const getIndexersForRepositoryQuery = `
WITH
repo_candidate AS (
	SELECT repo.id
	FROM repo
	WHERE
		repo.id = %s AND
		%s AND
		repo.deleted_at IS NULL AND
		repo.blocked IS NULL
),
combined_indexers AS (
	SELECT DISTINCT u.indexer FROM lsif_uploads u JOIN repo_candidate r ON r.id = u.repository_id
	UNION ALL
	SELECT DISTINCT u.indexer FROM lsif_indexes u JOIN repo_candidate r ON r.id = u.repository_id
)
SELECT DISTINCT u.indexer
FROM combined_indexers u
ORDER BY u.indexer
`

// GetRecentUploadsSummary returns a set of "interesting" uploads for the repository with the given identifeir.
// The return value is a list of uploads grouped by root and indexer. In each group, the set of uploads should
// include the set of unprocessed records as well as the latest finished record. These values allow users to
// quickly determine if a particular root/indexer pair is up-to-date or having issues processing.
func (s *store) GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error) {
	ctx, logger, endObservation := s.operations.getRecentUploadsSummary.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	uploads, err := scanUploadComplete(s.db.Query(ctx, sqlf.Sprintf(recentUploadsSummaryQuery, repositoryID, repositoryID)))
	if err != nil {
		return nil, err
	}
	logger.AddEvent("scanUploadComplete", attribute.Int("numUploads", len(uploads)))

	groupedUploads := make([]shared.UploadsWithRepositoryNamespace, 1, len(uploads)+1)
	for _, index := range uploads {
		if last := groupedUploads[len(groupedUploads)-1]; last.Root != index.Root || last.Indexer != index.Indexer {
			groupedUploads = append(groupedUploads, shared.UploadsWithRepositoryNamespace{
				Root:    index.Root,
				Indexer: index.Indexer,
			})
		}

		n := len(groupedUploads)
		groupedUploads[n-1].Uploads = append(groupedUploads[n-1].Uploads, index)
	}

	return groupedUploads[1:], nil
}

const recentUploadsSummaryQuery = `
WITH ranked_completed AS (
	SELECT
		u.id,
		u.root,
		u.indexer,
		u.finished_at,
		RANK() OVER (PARTITION BY root, ` + sanitizedIndexerExpression + ` ORDER BY finished_at DESC) AS rank
	FROM lsif_uploads u
	WHERE
		u.repository_id = %s AND
		u.state NOT IN ('uploading', 'queued', 'processing', 'deleted')
),
latest_uploads AS (
	SELECT u.id, u.root, u.indexer, u.uploaded_at
	FROM lsif_uploads u
	WHERE
		u.id IN (
			SELECT rc.id
			FROM ranked_completed rc
			WHERE rc.rank = 1
		)
	ORDER BY u.root, u.indexer
),
new_uploads AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE
		u.repository_id = %s AND
		u.state IN ('uploading', 'queued', 'processing') AND
		u.uploaded_at >= (
			SELECT lu.uploaded_at
			FROM latest_uploads lu
			WHERE
				lu.root = u.root AND
				lu.indexer = u.indexer
			-- condition passes when latest_uploads is empty
			UNION SELECT u.queued_at LIMIT 1
		)
)
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_at_tip,
	u.uploaded_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	u.repository_name,
	u.indexer,
	u.indexer_version,
	u.num_parts,
	u.uploaded_parts,
	u.upload_size,
	u.associated_index_id,
	u.content_type,
	u.should_reindex,
	s.rank,
	u.uncompressed_size
FROM lsif_uploads_with_repository_name u
LEFT JOIN (` + uploadRankQueryFragment + `) s
ON u.id = s.id
WHERE u.id IN (
	SELECT lu.id FROM latest_uploads lu
	UNION
	SELECT nu.id FROM new_uploads nu
)
ORDER BY u.root, u.indexer
`

const sanitizedIndexerExpression = `
(
    split_part(
        split_part(
            CASE
                -- Strip sourcegraph/ prefix if it exists
                WHEN strpos(indexer, 'sourcegraph/') = 1 THEN substr(indexer, length('sourcegraph/') + 1)
                ELSE indexer
            END,
        '@', 1), -- strip off @sha256:...
    ':', 1) -- strip off tag
)
`

// GetRecentAutoIndexJobsSummary returns the set of "interesting" indexes for the repository with the given identifier.
// The return value is a list of indexes grouped by root and indexer. In each group, the set of indexes should
// include the set of unprocessed records as well as the latest finished record. These values allow users to
// quickly determine if a particular root/indexer pair os up-to-date or having issues processing.
func (s *store) GetRecentAutoIndexJobsSummary(ctx context.Context, repositoryID int) (summaries []uploadsshared.GroupedAutoIndexJobs, err error) {
	ctx, logger, endObservation := s.operations.getRecentAutoIndexJobsSummary.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	indexes, err := scanJobs(s.db.Query(ctx, sqlf.Sprintf(recentIndexesSummaryQuery, repositoryID, repositoryID)))
	if err != nil {
		return nil, err
	}
	logger.AddEvent("scanJobs", attribute.Int("numIndexes", len(indexes)))

	groupedIndexes := make([]uploadsshared.GroupedAutoIndexJobs, 1, len(indexes)+1)
	for _, index := range indexes {
		if last := groupedIndexes[len(groupedIndexes)-1]; last.Root != index.Root || last.Indexer != index.Indexer {
			groupedIndexes = append(groupedIndexes, uploadsshared.GroupedAutoIndexJobs{
				Root:    index.Root,
				Indexer: index.Indexer,
			})
		}

		n := len(groupedIndexes)
		groupedIndexes[n-1].Indexes = append(groupedIndexes[n-1].Indexes, index)
	}

	return groupedIndexes[1:], nil
}

const recentIndexesSummaryQuery = `
WITH ranked_completed AS (
	SELECT
		u.id,
		u.root,
		u.indexer,
		u.finished_at,
		RANK() OVER (PARTITION BY root, ` + sanitizedIndexerExpression + ` ORDER BY finished_at DESC) AS rank
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.state NOT IN ('queued', 'processing', 'deleted')
),
latest_indexes AS (
	SELECT u.id, u.root, u.indexer, u.queued_at
	FROM lsif_indexes u
	WHERE
		u.id IN (
			SELECT rc.id
			FROM ranked_completed rc
			WHERE rc.rank = 1
		)
	ORDER BY u.root, u.indexer
),
new_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.state IN ('queued', 'processing') AND
		u.queued_at >= (
			SELECT lu.queued_at
			FROM latest_indexes lu
			WHERE
				lu.root = u.root AND
				lu.indexer = u.indexer
			-- condition passes when latest_indexes is empty
			UNION SELECT u.queued_at LIMIT 1
		)
)
SELECT
	u.id,
	u.commit,
	u.queued_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	u.repository_name,
	u.docker_steps,
	u.root,
	u.indexer,
	u.indexer_args,
	u.outfile,
	u.execution_logs,
	s.rank,
	u.local_steps,
	` + indexAssociatedUploadIDQueryFragment + `,
	u.should_reindex,
	u.requested_envvars,
	u.enqueuer_user_id
FROM lsif_indexes_with_repository_name u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
WHERE u.id IN (
	SELECT lu.id FROM latest_indexes lu
	UNION
	SELECT nu.id FROM new_indexes nu
)
ORDER BY u.root, u.indexer
`

func (s *store) RepositoryIDsWithErrors(ctx context.Context, offset, limit int) (_ []uploadsshared.RepositoryWithCount, totalCount int, err error) {
	ctx, _, endObservation := s.operations.repositoryIDsWithErrors.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return scanRepositoryWithCounts(s.db.Query(ctx, sqlf.Sprintf(repositoriesWithErrorsQuery, limit, offset)))
}

var scanRepositoryWithCounts = basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (rc uploadsshared.RepositoryWithCount, count int, _ error) {
	err := s.Scan(&rc.RepositoryID, &rc.Count, &count)
	return rc, count, err
})

const repositoriesWithErrorsQuery = `
WITH

-- Return unique (repository, root, indexer) triples for each "project" (root/indexer pair)
-- within a repository that has a failing record without a newer completed record shadowing
-- it. Group these by the project triples so that we only return one row for the count we
-- perform below.
candidates_from_uploads AS (
	SELECT u.repository_id
	FROM lsif_uploads_with_repository_name u
	WHERE
		u.state = 'failed' AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_uploads u2
			WHERE
				u2.state = 'completed' AND
				u2.repository_id = u.repository_id AND
				u2.root = u.root AND
				u2.indexer = u.indexer AND
				u2.finished_at > u.finished_at
		)
	GROUP BY u.repository_id, u.root, u.indexer
),

-- Same as above for index records
candidates_from_indexes AS (
	SELECT u.repository_id
	FROM lsif_indexes u
	WHERE
		u.state = 'failed' AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_indexes u2
			WHERE
				u2.state = 'completed' AND
				u2.repository_id = u.repository_id AND
				u2.root = u.root AND
				u2.indexer = u.indexer AND
				u2.finished_at > u.finished_at
		)
	GROUP BY u.repository_id, u.root, u.indexer
),

candidates AS (
	SELECT * FROM candidates_from_uploads UNION ALL
	SELECT * FROM candidates_from_indexes
),
grouped_candidates AS (
	SELECT
		r.repository_id,
		COUNT(*) AS num_failures
	FROM candidates r
	GROUP BY r.repository_id
)
SELECT
	r.repository_id,
	r.num_failures,
	COUNT(*) OVER() AS count
FROM grouped_candidates r
ORDER BY num_failures DESC, repository_id
LIMIT %s
OFFSET %s
`

func (s *store) NumRepositoriesWithCodeIntelligence(ctx context.Context) (_ int, err error) {
	ctx, _, endObservation := s.operations.numRepositoriesWithCodeIntelligence.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	numRepositoriesWithCodeIntelligence, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(countRepositoriesQuery)))
	if err != nil {
		return 0, err
	}

	return numRepositoriesWithCodeIntelligence, err
}

const countRepositoriesQuery = `
WITH candidate_repositories AS (
	SELECT
	DISTINCT uvt.repository_id AS id
	FROM lsif_uploads_visible_at_tip uvt
	WHERE is_default_branch
)
SELECT COUNT(*)
FROM candidate_repositories s
JOIN repo r ON r.id = s.id
WHERE
	r.deleted_at IS NULL AND
	r.blocked IS NULL
`
