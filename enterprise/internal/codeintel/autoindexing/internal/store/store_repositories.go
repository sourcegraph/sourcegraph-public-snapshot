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

func (s *store) NumRepositoriesWithCodeIntelligence(ctx context.Context) (_ int, err error) {
	ctx, _, endObservation := s.operations.numRepositoriesWithCodeIntelligence.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
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

func (s *store) RepositoryIDsWithErrors(ctx context.Context, offset, limit int) (_ []shared.RepositoryWithCount, totalCount int, err error) {
	ctx, _, endObservation := s.operations.repositoryIDsWithErrors.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	return scanRepositoryWithCounts(s.db.Query(ctx, sqlf.Sprintf(repositoriesWithErrorsQuery, limit, offset)))
}

var scanRepositoryWithCounts = basestore.NewSliceWithCountScanner(func(s dbutil.Scanner) (rc shared.RepositoryWithCount, count int, _ error) {
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
	FROM lsif_uploads u
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
