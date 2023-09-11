package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) TopRepositoriesToConfigure(ctx context.Context, limit int) (_ []shared.RepositoryWithCount, err error) {
	ctx, _, endObservation := s.operations.topRepositoriesToConfigure.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	return scanRepositoryWithCounts(s.db.Query(ctx, sqlf.Sprintf(
		topRepositoriesToConfigureQuery,
		pq.Array(eventLogNames),
		eventLogsWindow/time.Hour,
		limit,
	)))
}

var eventLogNames = []string{
	"codeintel.searchDefinitions.xrepo",
	"codeintel.searchDefinitions",
	"codeintel.searchHover",
	"codeintel.searchReferences.xrepo",
	"codeintel.searchReferences",
}

// about one month
const eventLogsWindow = time.Hour * 24 * 30

const topRepositoriesToConfigureQuery = `
SELECT
	r.id,
	COUNT(*) as num_events
FROM event_logs e
JOIN repo r ON r.id = (e.argument->'repositoryId')::integer
WHERE
	e.name = ANY(%s) AND
	e.timestamp >= NOW() - (%s * '1 hour'::interval) AND
	r.deleted_at IS NULL AND
	r.blocked IS NULL
GROUP BY r.id
ORDER BY num_events DESC, id
LIMIT %s
`

func (s *store) RepositoryIDsWithConfiguration(ctx context.Context, offset, limit int) (_ []shared.RepositoryWithAvailableIndexers, totalCount int, err error) {
	ctx, _, endObservation := s.operations.repositoryIDsWithConfiguration.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("offset", offset),
		attribute.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	return scanRepositoryWithAvailableIndexersSlice(s.db.Query(ctx, sqlf.Sprintf(
		repositoriesWithConfigurationQuery,
		limit,
		offset,
	)))
}

const repositoriesWithConfigurationQuery = `
SELECT
	r.id,
	cai.available_indexers,
	COUNT(*) OVER() AS count
FROM cached_available_indexers cai
JOIN repo r ON r.id = cai.repository_id
WHERE
	available_indexers != '{}'::jsonb AND
	r.deleted_at IS NULL AND
	r.blocked IS NULL
ORDER BY num_events DESC
LIMIT %s
OFFSET %s
`

func (s *store) GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := s.operations.getLastIndexScanForRepository.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	t, ok, err := basestore.ScanFirstTime(s.db.Query(ctx, sqlf.Sprintf(lastIndexScanForRepositoryQuery, repositoryID)))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &t, nil
}

const lastIndexScanForRepositoryQuery = `
SELECT last_index_scan_at FROM lsif_last_index_scan WHERE repository_id = %s
`

func (s *store) SetConfigurationSummary(ctx context.Context, repositoryID int, numEvents int, availableIndexers map[string]shared.AvailableIndexer) (err error) {
	ctx, _, endObservation := s.operations.setConfigurationSummary.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.Int("numEvents", numEvents),
		attribute.Int("numIndexers", len(availableIndexers)),
	}})
	defer endObservation(1, observation.Args{})

	payload, err := json.Marshal(availableIndexers)
	if err != nil {
		return err
	}

	return s.db.Exec(ctx, sqlf.Sprintf(setConfigurationSummaryQuery, repositoryID, numEvents, payload))
}

const setConfigurationSummaryQuery = `
INSERT INTO cached_available_indexers (repository_id, num_events, available_indexers)
VALUES (%s, %s, %s)
ON CONFLICT(repository_id) DO UPDATE
SET
	num_events = EXCLUDED.num_events,
	available_indexers = EXCLUDED.available_indexers
`

func (s *store) TruncateConfigurationSummary(ctx context.Context, numRecordsToRetain int) (err error) {
	ctx, _, endObservation := s.operations.truncateConfigurationSummary.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numRecordsToRetain", numRecordsToRetain),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(truncateConfigurationSummaryQuery, numRecordsToRetain))
}

const truncateConfigurationSummaryQuery = `
WITH safe AS (
	SELECT id
	FROM cached_available_indexers
	ORDER BY num_events DESC
	LIMIT %s
)
DELETE FROM cached_available_indexers
WHERE id NOT IN (SELECT id FROM safe)
`

//
//

func scanRepositoryWithCount(s dbutil.Scanner) (rc shared.RepositoryWithCount, _ error) {
	return rc, s.Scan(&rc.RepositoryID, &rc.Count)
}

var scanRepositoryWithCounts = basestore.NewSliceScanner(scanRepositoryWithCount)

func scanRepositoryWithAvailableIndexers(s dbutil.Scanner) (rai shared.RepositoryWithAvailableIndexers, count int, _ error) {
	var rawPayload string
	if err := s.Scan(&rai.RepositoryID, &rawPayload, &count); err != nil {
		return rai, 0, err
	}
	if err := json.Unmarshal([]byte(rawPayload), &rai.AvailableIndexers); err != nil {
		return rai, 0, err
	}

	return rai, count, nil
}

var scanRepositoryWithAvailableIndexersSlice = basestore.NewSliceWithCountScanner(scanRepositoryWithAvailableIndexers)
