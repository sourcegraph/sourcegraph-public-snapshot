package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetUploads returns a list of uploads and the total count of records matching the given conditions.
func (s *store) GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []shared.Upload, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getUploads.With(ctx, &err, observation.Args{LogFields: buildLogFields(opts)})
	defer endObservation(1, observation.Args{})

	conds, cte := buildConditionsAndCte(opts)
	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, 0, err
	}
	conds = append(conds, authzConds)

	var orderExpression *sqlf.Query
	if opts.OldestFirst {
		orderExpression = sqlf.Sprintf("uploaded_at")
	} else {
		orderExpression = sqlf.Sprintf("uploaded_at DESC")
	}

	query := sqlf.Sprintf(
		getUploadsQuery,
		buildCTEPrefix(cte),
		sqlf.Join(conds, " AND "),
		orderExpression,
		opts.Limit,
		opts.Offset,
	)
	uploads, totalCount, err = scanUploadsWithCount(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}
	trace.Log(
		log.Int("totalCount", totalCount),
		log.Int("numUploads", len(uploads)),
	)

	return uploads, totalCount, nil
}

const getUploadsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:GetUploads
%s -- Dynamic CTE definitions for use in the WHERE clause
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
	repo.name,
	u.indexer,
	u.indexer_version,
	u.num_parts,
	u.uploaded_parts,
	u.upload_size,
	u.associated_index_id,
	s.rank,
	COUNT(*) OVER() AS count
FROM lsif_uploads u
LEFT JOIN (` + uploadRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE %s ORDER BY %s LIMIT %d OFFSET %d
`

const uploadRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_after, r.uploaded_at), r.id) as rank
FROM lsif_uploads_with_repository_name r
WHERE r.state = 'queued'
`

const visibleAtTipSubselectQuery = `SELECT 1 FROM lsif_uploads_visible_at_tip uvt WHERE uvt.repository_id = u.repository_id AND uvt.upload_id = u.id`

var rankedDependencyCandidateCTEQuery = `
SELECT
	p.dump_id as pkg_id,
	r.dump_id as ref_id,
	-- Rank each upload providing the same package from the same directory
	-- within a repository by commit date. We'll choose the oldest commit
	-- date as the canonical choice and ignore the uploads for younger
	-- commits providing the same package.
	` + packageRankingQueryFragment + ` AS rank
FROM lsif_uploads u
JOIN lsif_packages p ON p.dump_id = u.id
JOIN lsif_references r ON r.scheme = p.scheme
	AND r.name = p.name
	AND r.version = p.version
	AND r.dump_id != p.dump_id
WHERE
	-- Don't match deleted uploads
	u.state = 'completed' AND
	%s
`

// packageRankingQueryFragment uses `lsif_uploads u` JOIN `lsif_packages p` to return a rank
// for each row grouped by package and source code location and ordered by the associated Git
// commit date.
const packageRankingQueryFragment = `
rank() OVER (
	PARTITION BY
		-- Group providers of the same package together
		p.scheme, p.name, p.version,
		-- Defined by the same directory within a repository
		u.repository_id, u.indexer, u.root
	ORDER BY
		-- Rank each grouped upload by the associated commit date
		u.committed_at,
		-- Break ties via the unique identifier
		u.id
)
`

var rankedDependentCandidateCTEQuery = `
SELECT
	p.dump_id as pkg_id,
	p.scheme as scheme,
	p.name as name,
	p.version as version,
	-- Rank each upload providing the same package from the same directory
	-- within a repository by commit date. We'll choose the oldest commit
	-- date as the canonical choice and ignore the uploads for younger
	-- commits providing the same package.
	` + packageRankingQueryFragment + ` AS rank
FROM lsif_uploads u
JOIN lsif_packages p ON p.dump_id = u.id
WHERE
	-- Don't match deleted uploads
	u.state = 'completed' AND
	%s
`

// DeletedRepositoryGracePeriod is the minimum allowable duration between a repo deletion
// and the upload and index records for that repository being deleted.
const DeletedRepositoryGracePeriod = time.Minute * 30

// DeleteUploadsWithoutRepository deletes uploads associated with repositories that were deleted at least
// DeletedRepositoryGracePeriod ago. This returns the repository identifier mapped to the number of uploads
// that were removed for that repository.
func (s *store) DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, trace, endObservation := s.operations.deleteUploadsWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	unset, _ := tx.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with repository not known to this instance")
	defer unset(ctx)

	query := sqlf.Sprintf(deleteUploadsWithoutRepositoryQuery, now.UTC(), DeletedRepositoryGracePeriod/time.Second)
	repositories, err := scanCounts(tx.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	count := 0
	for _, numDeleted := range repositories {
		count += numDeleted
	}
	trace.Log(
		log.Int("count", count),
		log.Int("numRepositories", len(repositories)),
	)

	return repositories, nil
}

const deleteUploadsWithoutRepositoryQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:DeleteUploadsWithoutRepository
WITH
candidates AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_uploads u ON u.repository_id = r.id
	WHERE %s - r.deleted_at >= %s * interval '1 second'

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	-- Note: we can go straight from completed -> deleted here as we
	-- do not need to preserve the deleted repository's current commit
	-- graph (the API cannot resolve any queries for this repository).

	UPDATE lsif_uploads u
	SET state = 'deleted'
	WHERE u.id IN (SELECT id FROM candidates)
	RETURNING u.id, u.repository_id
)
SELECT d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

// DeleteUploadsStuckUploading soft deletes any upload record that has been uploading since the given time.
func (s *store) DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error) {
	ctx, trace, endObservation := s.operations.deleteUploadsStuckUploading.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("uploadedBefore", uploadedBefore.Format(time.RFC3339)), // TODO - should be a duration
	}})
	defer endObservation(1, observation.Args{})

	unset, _ := s.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "stuck in uploading state")
	defer unset(ctx)

	query := sqlf.Sprintf(deleteUploadsStuckUploadingQuery, uploadedBefore)
	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, query))
	if err != nil {
		return 0, err
	}
	trace.Log(log.Int("count", count))

	return count, nil
}

const deleteUploadsStuckUploadingQuery = `
-- source: internal/codeintel/uploads/internal/store/store_uploads.go:DeleteUploadsStuckUploading
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'uploading' AND u.uploaded_at < %s

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	UPDATE lsif_uploads u
	SET state = 'deleted'
	WHERE id IN (SELECT id FROM candidates)
	RETURNING u.repository_id
)
SELECT count(*) FROM deleted
`

func nilTimeToString(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.String()
}

func buildConditionsAndCte(opts shared.GetUploadsOptions) ([]*sqlf.Query, []cteDefinition) {
	conds := make([]*sqlf.Query, 0, 12)
	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeSearchCondition(opts.Term))
	}
	if opts.State != "" {
		conds = append(conds, makeStateCondition(opts.State))
	} else {
		conds = append(conds, sqlf.Sprintf("u.state != 'deleted'"))
	}
	if opts.VisibleAtTip {
		conds = append(conds, sqlf.Sprintf("EXISTS ("+visibleAtTipSubselectQuery+")"))
	}

	cteDefinitions := make([]cteDefinition, 0, 2)
	if opts.DependencyOf != 0 {
		cteDefinitions = append(cteDefinitions, cteDefinition{
			name:       "ranked_dependencies",
			definition: sqlf.Sprintf(rankedDependencyCandidateCTEQuery, sqlf.Sprintf("r.dump_id = %s", opts.DependencyOf)),
		})

		// Limit results to the set of uploads canonically providing packages referenced by the given upload identifier
		// (opts.DependencyOf). We do this by selecting the top ranked values in the CTE defined above, which are the
		// referenced package providers grouped by package name, version, repository, and root.
		conds = append(conds, sqlf.Sprintf(`u.id IN (SELECT rd.pkg_id FROM ranked_dependencies rd WHERE rd.rank = 1)`))
	}
	if opts.DependentOf != 0 {
		cteCondition := sqlf.Sprintf(`(p.scheme, p.name, p.version) IN (
			SELECT p.scheme, p.name, p.version
			FROM lsif_packages p
			WHERE p.dump_id = %s
		)`, opts.DependentOf)

		cteDefinitions = append(cteDefinitions, cteDefinition{
			name:       "ranked_dependents",
			definition: sqlf.Sprintf(rankedDependentCandidateCTEQuery, cteCondition),
		})

		// Limit results to the set of uploads that reference the target upload if it canonically provides the
		// matching package. If the target upload does not canonically provide a package, the results will contain
		// no dependent uploads.
		conds = append(conds, sqlf.Sprintf(`u.id IN (
			SELECT r.dump_id
			FROM ranked_dependents rd
			JOIN lsif_references r ON r.scheme = rd.scheme
				AND r.name = rd.name
				AND r.version = rd.version
				AND r.dump_id != rd.pkg_id
			WHERE rd.pkg_id = %s AND rd.rank = 1
		)`, opts.DependentOf))
	}
	if opts.UploadedBefore != nil {
		conds = append(conds, sqlf.Sprintf("u.uploaded_at < %s", *opts.UploadedBefore))
	}
	if opts.UploadedAfter != nil {
		conds = append(conds, sqlf.Sprintf("u.uploaded_at > %s", *opts.UploadedAfter))
	}
	if opts.InCommitGraph {
		conds = append(conds, sqlf.Sprintf("u.finished_at < (SELECT updated_at FROM lsif_dirty_repositories ldr WHERE ldr.repository_id = u.repository_id)"))
	}
	if opts.LastRetentionScanBefore != nil {
		conds = append(conds, sqlf.Sprintf("(u.last_retention_scan_at IS NULL OR u.last_retention_scan_at < %s)", *opts.LastRetentionScanBefore))
	}
	if !opts.AllowExpired {
		conds = append(conds, sqlf.Sprintf("NOT u.expired"))
	}
	if !opts.AllowDeletedRepo {
		conds = append(conds, sqlf.Sprintf("repo.deleted_at IS NULL"))
	}

	return conds, cteDefinitions
}

// makeSearchCondition returns a disjunction of LIKE clauses against all searchable columns of an upload.
func makeSearchCondition(term string) *sqlf.Query {
	searchableColumns := []string{
		"u.commit",
		"u.root",
		"(u.state)::text",
		"u.failure_message",
		"repo.name",
		"u.indexer",
		"u.indexer_version",
	}

	var termConds []*sqlf.Query
	for _, column := range searchableColumns {
		termConds = append(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}

// makeStateCondition returns a disjunction of clauses comparing the upload against the target state.
func makeStateCondition(state string) *sqlf.Query {
	states := make([]string, 0, 2)
	if state == "errored" || state == "failed" {
		// Treat errored and failed states as equivalent
		states = append(states, "errored", "failed")
	} else {
		states = append(states, state)
	}

	queries := make([]*sqlf.Query, 0, len(states))
	for _, state := range states {
		queries = append(queries, sqlf.Sprintf("u.state = %s", state))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(queries, " OR "))
}

func buildCTEPrefix(cteDefinitions []cteDefinition) *sqlf.Query {
	if len(cteDefinitions) == 0 {
		return sqlf.Sprintf("")
	}

	cteQueries := make([]*sqlf.Query, 0, len(cteDefinitions))
	for _, cte := range cteDefinitions {
		cteQueries = append(cteQueries, sqlf.Sprintf("%s AS (%s)", sqlf.Sprintf(cte.name), cte.definition))
	}

	return sqlf.Sprintf("WITH\n%s", sqlf.Join(cteQueries, ",\n"))
}

func buildLogFields(opts shared.GetUploadsOptions) []log.Field {
	return []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("state", opts.State),
		log.String("term", opts.Term),
		log.Bool("visibleAtTip", opts.VisibleAtTip),
		log.Int("dependencyOf", opts.DependencyOf),
		log.Int("dependentOf", opts.DependentOf),
		log.String("uploadedBefore", nilTimeToString(opts.UploadedBefore)),
		log.String("uploadedAfter", nilTimeToString(opts.UploadedAfter)),
		log.String("lastRetentionScanBefore", nilTimeToString(opts.LastRetentionScanBefore)),
		log.Bool("inCommitGraph", opts.InCommitGraph),
		log.Bool("allowExpired", opts.AllowExpired),
		log.Bool("oldestFirst", opts.OldestFirst),
		log.Int("limit", opts.Limit),
		log.Int("offset", opts.Offset),
	}
}
