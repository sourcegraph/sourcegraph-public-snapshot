package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrUnknownRepository = errors.New("unknown repository")

func (s *store) GetRepoName(ctx context.Context, repositoryID int) (name string, err error) {
	ctx, _, endObservation := s.operations.getRepoName.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.String("repositoryName", name),
		}})
	}()

	name, exists, err := basestore.ScanFirstString(s.db.Query(ctx, sqlf.Sprintf(repoNameQuery, repositoryID)))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrUnknownRepository
	}
	return name, nil
}

const repoNameQuery = `
SELECT name FROM repo WHERE id = %s
`

func (s *store) RepositoryIDsWithConfiguration(ctx context.Context, offset, limit int) (_ []shared.RepositoryWithAvailableIndexers, totalCount int, err error) {
	ctx, _, endObservation := s.operations.repositoryIDsWithConfiguration.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	return scanRepositoryWithAvailableIndexers(s.db.Query(ctx, sqlf.Sprintf(repositoriesWithConfigurationQuery, limit, offset)))
}

var scanRepositoryWithAvailableIndexers = basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (rai shared.RepositoryWithAvailableIndexers, count int, _ error) {
	var payload string
	if err := s.Scan(&rai.RepositoryID, &payload, &count); err != nil {
		return rai, 0, err
	}
	if err := json.Unmarshal([]byte(payload), &rai.AvailableIndexers); err != nil {
		return rai, 0, err
	}

	return rai, count, nil
})

const repositoriesWithConfigurationQuery = `
SELECT
	repository_id,
	available_indexers,
	COUNT(*) OVER() AS count
FROM cached_available_indexers
WHERE
	available_indexers != '{}'::jsonb
ORDER BY num_events DESC LIMIT %s OFFSET %s
`

// about one month
const eventLogsWindow = time.Hour * 24 * 30

func (s *store) TopRepositoriesToConfigure(ctx context.Context, limit int) (_ []shared.RepositoryWithCount, err error) {
	ctx, _, endObservation := s.operations.topRepositoriesToConfigure.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	repositories, err := basestore.NewSliceScanner(func(s dbutil.Scanner) (rc shared.RepositoryWithCount, _ error) {
		err := s.Scan(&rc.RepositoryID, &rc.Count)
		return rc, err
	})(s.db.Query(ctx, sqlf.Sprintf(topRepositoriesToConfigureQuery, pq.Array(eventNames), eventLogsWindow/time.Hour, limit)))
	if err != nil {
		return nil, err
	}

	return repositories, nil
}

var eventNames = []string{
	"codeintel.searchDefinitions.xrepo",
	"codeintel.searchDefinitions",
	"codeintel.searchHover",
	"codeintel.searchReferences.xrepo",
	"codeintel.searchReferences",
}

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

func (s *store) SetConfigurationSummary(ctx context.Context, repositoryID int, numEvents int, availableIndexers map[string]shared.AvailableIndexer) (err error) {
	ctx, _, endObservation := s.operations.setConfigurationSummary.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	payload, err := json.Marshal(availableIndexers)
	if err != nil {
		return err
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(setConfigurationSummaryQuery, repositoryID, numEvents, payload)); err != nil {
		return err
	}

	return nil
}

const setConfigurationSummaryQuery = `
INSERT INTO cached_available_indexers(repository_id, num_events, available_indexers) VALUES (%s, %s, %s)
ON CONFLICT(repository_id) DO UPDATE
SET
	num_events = EXCLUDED.num_events,
	available_indexers = EXCLUDED.available_indexers
`

func (s *store) TruncateConfigurationSummary(ctx context.Context, numRecordsToRetain int) (err error) {
	ctx, _, endObservation := s.operations.truncateConfigurationSummary.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("total", numRecordsToRetain),
	}})
	defer endObservation(1, observation.Args{})

	if err := s.db.Exec(ctx, sqlf.Sprintf(truncateConfigurationSummaryQuery, numRecordsToRetain)); err != nil {
		return err
	}

	return nil
}

const truncateConfigurationSummaryQuery = `
WITH safe AS (
	SELECT id
	FROM cached_available_indexers
	ORDER BY num_events DESC
	LIMIT %s
)
DELETE FROM cached_available_indexers WHERE id NOT IN (SELECT id FROM safe)
`
