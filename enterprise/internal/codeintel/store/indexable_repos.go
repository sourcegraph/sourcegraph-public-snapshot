package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
)

// IndexableRepository marks a repository for eligibility to be index automatically.
type IndexableRepository struct {
	RepositoryID        int
	SearchCount         int
	PreciseCount        int
	LastIndexEnqueuedAt *time.Time
}

// UpdateableIndexableRepository is a version of IndexableRepository with pointer
// fields used to indicate which values should be updated on an upsert operation.
type UpdateableIndexableRepository struct {
	RepositoryID        int
	SearchCount         *int
	PreciseCount        *int
	LastIndexEnqueuedAt *time.Time
}

// IndexableRepositoryQueryOptions controls the result filter for IndexableRepositories.
type IndexableRepositoryQueryOptions struct {
	Limit                       int
	MinimumTimeSinceLastEnqueue time.Duration
	MinimumSearchCount          int
	MinimumPreciseCount         int
	MinimumSearchRatio          float64
	now                         time.Time
}

// scanIndexableRepositories scans a slice of indexable repositories from the return value of `*store.query`.
func scanIndexableRepositories(rows *sql.Rows, queryErr error) (_ []IndexableRepository, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	var indexableRepositories []IndexableRepository
	for rows.Next() {
		var indexableRepository IndexableRepository
		if err := rows.Scan(
			&indexableRepository.RepositoryID,
			&indexableRepository.SearchCount,
			&indexableRepository.PreciseCount,
			&indexableRepository.LastIndexEnqueuedAt,
		); err != nil {
			return nil, err
		}

		indexableRepositories = append(indexableRepositories, indexableRepository)
	}

	return indexableRepositories, nil
}

// IndexableRepositories returns the metadata of all indexable repositories.
func (s *store) IndexableRepositories(ctx context.Context, opts IndexableRepositoryQueryOptions) ([]IndexableRepository, error) {
	if opts.now.IsZero() {
		opts.now = time.Now()
	}

	if opts.Limit <= 0 {
		return nil, ErrIllegalLimit
	}

	var conds []*sqlf.Query
	if opts.MinimumTimeSinceLastEnqueue > 0 {
		conds = append(conds, sqlf.Sprintf(
			"(last_index_enqueued_at IS NULL OR %s - last_index_enqueued_at < (%s || ' second')::interval)",
			opts.now,
			opts.MinimumTimeSinceLastEnqueue/time.Second,
		))
	}
	if opts.MinimumSearchCount > 0 {
		conds = append(conds, sqlf.Sprintf("search_count >= %s", opts.MinimumSearchCount))
	}
	if opts.MinimumPreciseCount > 0 {
		conds = append(conds, sqlf.Sprintf("precise_count >= %s", opts.MinimumPreciseCount))
	}
	if opts.MinimumSearchRatio > 0 {
		conds = append(conds, sqlf.Sprintf("search_count::float / (search_count + precise_count) >= %s", opts.MinimumSearchRatio))
	}

	var whereClause *sqlf.Query
	if len(conds) > 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, " AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	return scanIndexableRepositories(s.query(ctx, sqlf.Sprintf(`
		SELECT
			repository_id,
			search_count,
			precise_count,
			last_index_enqueued_at
		FROM lsif_indexable_repositories
		%s
		LIMIT %s
	`, whereClause, opts.Limit)))
}

// UpdateIndexableRepository updates the metadata for an indexable repository. If the repository is not
// already marked as indexable, a new record will be created.
func (s *store) UpdateIndexableRepository(ctx context.Context, indexableRepository UpdateableIndexableRepository, now time.Time) error {
	// Ensure that record exists before we attempt to update it
	err := s.queryForEffect(ctx, sqlf.Sprintf(`
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
		pairs = append(pairs, sqlf.Sprintf("search_count=%s", indexableRepository.SearchCount))
	}
	if indexableRepository.PreciseCount != nil {
		pairs = append(pairs, sqlf.Sprintf("precise_count=%s", indexableRepository.PreciseCount))
	}
	if indexableRepository.LastIndexEnqueuedAt != nil {
		pairs = append(pairs, sqlf.Sprintf("last_index_enqueued_at=%s", indexableRepository.LastIndexEnqueuedAt))
	}
	if len(pairs) == 0 {
		return nil
	}

	return s.queryForEffect(ctx, sqlf.Sprintf(`
		UPDATE lsif_indexable_repositories
		SET %s, last_updated_at = %s
		WHERE repository_id = %s
	`, sqlf.Join(pairs, ","), now, indexableRepository.RepositoryID))
}

// ResetIndexableRepositories zeroes the event counts for indexable repositories that have not been updated
// since lastUpdatedBefore.
func (s *store) ResetIndexableRepositories(ctx context.Context, lastUpdatedBefore time.Time) error {
	return s.queryForEffect(ctx, sqlf.Sprintf(
		`
		UPDATE lsif_indexable_repositories
		SET search_count = 0, precise_count = 0
		WHERE last_updated_at < %s
	`,
		lastUpdatedBefore,
	))
}
