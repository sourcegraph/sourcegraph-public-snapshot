package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// DeletedRepositoryGracePeriod is the minimum allowable duration between
// a repo deletion and index records for that repository being deleted.
const DeletedRepositoryGracePeriod = time.Minute * 30

// DeleteIndexesWithoutRepository deletes indexes associated with repositories that were deleted at least
// DeletedRepositoryGracePeriod ago. This returns the repository identifier mapped to the number of indexes
// that were removed for that repository.
func (s *store) DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, trace, endObservation := s.operations.deleteIndexesWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	// TODO(efritz) - this would benefit from an index on repository_id. We currently have
	// a similar one on this index, but only for uploads that are completed or visible at tip.
	repositories, err := scanCounts(tx.Query(ctx, sqlf.Sprintf(deleteIndexesWithoutRepositoryQuery, now.UTC(), DeletedRepositoryGracePeriod/time.Second)))
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

const deleteIndexesWithoutRepositoryQuery = `
-- source: internal/codeintel/autoindexing/internal/store/store_indexes.go:DeleteIndexesWithoutRepository
WITH
candidates AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_indexes u ON u.repository_id = r.id
	WHERE %s - r.deleted_at >= %s * interval '1 second'

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_indexes table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_indexes u
	WHERE id IN (SELECT id FROM candidates)
	RETURNING u.id, u.repository_id
)
SELECT d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`
