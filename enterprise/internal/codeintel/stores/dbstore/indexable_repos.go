package dbstore

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// IndexableRepository marks a repository for eligibility to be indexed automatically.
type IndexableRepository struct {
	RepositoryID        int
	SearchCount         int
	PreciseCount        int
	LastIndexEnqueuedAt *time.Time
	Enabled             *bool
}

// UpdateableIndexableRepository is a version of IndexableRepository with pointer
// fields used to indicate which values should be updated on an upsert operation.
type UpdateableIndexableRepository struct {
	RepositoryID        int
	SearchCount         *int
	PreciseCount        *int
	LastIndexEnqueuedAt *time.Time
	Enabled             *bool
}

// IndexableRepositoryQueryOptions controls the result filter for IndexableRepositories.
type IndexableRepositoryQueryOptions struct {
	Limit                       int
	MinimumSearchCount          int           // number of events needed to begin indexing
	MinimumSearchRatio          float64       // ratio of search/total events needed to begin indexing
	MinimumPreciseCount         int           // number of events needed to continue indexing
	MinimumTimeSinceLastEnqueue time.Duration // time between enqueues
	now                         time.Time
}

// scanIndexableRepositories scans a slice of indexable repositories from the return value of `*Store.query`.
func scanIndexableRepositories(rows *sql.Rows, queryErr error) (_ []IndexableRepository, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var indexableRepositories []IndexableRepository
	for rows.Next() {
		var indexableRepository IndexableRepository
		if err := rows.Scan(
			&indexableRepository.RepositoryID,
			&indexableRepository.SearchCount,
			&indexableRepository.PreciseCount,
			&indexableRepository.LastIndexEnqueuedAt,
			&indexableRepository.Enabled,
		); err != nil {
			return nil, err
		}

		indexableRepositories = append(indexableRepositories, indexableRepository)
	}

	return indexableRepositories, nil
}

// IndexableRepositories returns the metadata of all indexable repositories.
func (s *Store) IndexableRepositories(ctx context.Context, opts IndexableRepositoryQueryOptions) (_ []IndexableRepository, err error) {
	ctx, traceLog, endObservation := s.operations.indexableRepositories.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("limit", opts.Limit),
		log.Int("minimumSearchCount", opts.MinimumSearchCount),
		log.Float64("minimumSearchRatio", opts.MinimumSearchRatio),
		log.Int("minimumPreciseCount", opts.MinimumPreciseCount),
		log.String("minimumTimeSinceLastEnqueue", opts.MinimumTimeSinceLastEnqueue.String()),
	}})
	defer endObservation(1, observation.Args{})

	if opts.now.IsZero() {
		opts.now = time.Now()
	}

	if opts.Limit <= 0 {
		return nil, ErrIllegalLimit
	}

	var triggers []*sqlf.Query
	if opts.MinimumSearchCount > 0 || opts.MinimumSearchRatio > 0 {
		// Select which repositories with little/no precise code intel to begin indexing
		triggers = append(triggers, sqlf.Sprintf(
			"(search_count >= %s AND search_count::float / NULLIF(search_count + precise_count, 0) >= %s)",
			opts.MinimumSearchCount,
			opts.MinimumSearchRatio,
		))
	}
	if opts.MinimumPreciseCount > 0 {
		// Select which repositories with precise intel to update
		triggers = append(triggers, sqlf.Sprintf("(precise_count >= %s)", opts.MinimumPreciseCount))
	}

	var conds []*sqlf.Query
	if len(triggers) > 0 {
		conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(triggers, " OR ")))
	}
	if opts.MinimumTimeSinceLastEnqueue > 0 {
		conds = append(conds, sqlf.Sprintf(
			"(last_index_enqueued_at IS NULL OR %s - last_index_enqueued_at >= (%s || ' second')::interval)",
			opts.now,
			strconv.Itoa(int(opts.MinimumTimeSinceLastEnqueue/time.Second)),
		))
	}

	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("true"))
	}

	repositories, err := scanIndexableRepositories(s.Store.Query(ctx, sqlf.Sprintf(indexableRepositoriesQuery, sqlf.Join(conds, " AND "), opts.Limit)))
	if err != nil {
		return nil, err
	}
	traceLog(log.Int("numRepositories", len(repositories)))

	return repositories, nil
}

const indexableRepositoriesQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/indexable_repos.go:IndexableRepositories
SELECT
	repository_id,
	search_count,
	precise_count,
	last_index_enqueued_at,
	enabled
FROM lsif_indexable_repositories
WHERE enabled is not false AND (enabled is true OR (%s))
LIMIT %s
`

// UpdateIndexableRepository updates the metadata for an indexable repository. If the repository is not
// already marked as indexable, a new record will be created.
func (s *Store) UpdateIndexableRepository(ctx context.Context, indexableRepository UpdateableIndexableRepository, now time.Time) (err error) {
	ctx, endObservation := s.operations.updateIndexableRepository.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", indexableRepository.RepositoryID),
	}})
	defer endObservation(1, observation.Args{})

	// Ensure that record exists before we attempt to update it
	err = s.Store.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO lsif_indexable_repositories (repository_id)
		VALUES (%s)
		ON CONFLICT DO NOTHING
	`,
		indexableRepository.RepositoryID,
	))
	if err != nil {
		return err
	}

	var pairs []*sqlf.Query
	if indexableRepository.SearchCount != nil {
		pairs = append(pairs, sqlf.Sprintf("search_count = %s", indexableRepository.SearchCount))
	}
	if indexableRepository.PreciseCount != nil {
		pairs = append(pairs, sqlf.Sprintf("precise_count = %s", indexableRepository.PreciseCount))
	}
	if indexableRepository.LastIndexEnqueuedAt != nil {
		pairs = append(pairs, sqlf.Sprintf("last_index_enqueued_at = %s", indexableRepository.LastIndexEnqueuedAt))
	}
	if indexableRepository.Enabled != nil {
		pairs = append(pairs, sqlf.Sprintf("enabled = %s", indexableRepository.Enabled))
	}
	if len(pairs) == 0 {
		return nil
	}

	return s.Store.Exec(ctx, sqlf.Sprintf(updateIndexableRepositoryQuery, sqlf.Join(pairs, ","), now, indexableRepository.RepositoryID))
}

const updateIndexableRepositoryQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/indexable_repos.go:UpdateIndexableRepository
UPDATE lsif_indexable_repositories SET %s, last_updated_at = %s WHERE repository_id = %s
`

// ResetIndexableRepositories zeroes the event counts for indexable repositories that have not been updated
// since lastUpdatedBefore.
func (s *Store) ResetIndexableRepositories(ctx context.Context, lastUpdatedBefore time.Time) (err error) {
	ctx, endObservation := s.operations.resetIndexableRepositories.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("lastUpdatedBefore", lastUpdatedBefore.Format(time.RFC3339)), // TODO - should be a duration
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(resetIndexableRepositoriesQuery, lastUpdatedBefore))
}

const resetIndexableRepositoriesQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/indexable_repos.go:ResetIndexableRepositories
UPDATE lsif_indexable_repositories SET search_count = 0, precise_count = 0 WHERE last_updated_at < %s
`
