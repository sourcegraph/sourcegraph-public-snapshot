package store

import (
	"context"
	"encoding/json"

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

// TODO - test
func (s *store) TopRepositoriesToConfigure(ctx context.Context, limit int) (_ []shared.RepositoryWithCount, err error) {
	ctx, _, endObservation := s.operations.topRepositoriesToConfigure.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	repositories, err := basestore.NewSliceScanner(func(s dbutil.Scanner) (rc shared.RepositoryWithCount, _ error) {
		err := s.Scan(&rc.RepositoryID, &rc.Count)
		return rc, err
	})(s.db.Query(ctx, sqlf.Sprintf(topRepositoriesToConfigureQuery, pq.Array(eventNames), 24*30, limit)))
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
WITH candidate_repositories AS (
	SELECT
		(argument->'repositoryId')::integer AS repository_id,
		COUNT(*) as count
	FROM event_logs
	WHERE
		name = ANY(%s) AND
		timestamp >= NOW() - (%s * '1 hour'::interval)
	GROUP BY repository_id
	ORDER BY count DESC
	LIMIT %s
)
SELECT id
FROM repo
WHERE
	id IN (SELECT repository_id FROM candidate_repositories) AND
	deleted_at IS NULL AND
	blocked IS NULL
`

// TODO - test
func (s *store) SetConfigurationSummary(ctx context.Context, repositoryID int, numEvents int, availableIndexers map[string]shared.AvailableIndexer) (err error) {
	ctx, _, endObservation := s.operations.setConfigurationSummary.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	payload, err := json.Marshal(availableIndexers)
	if err != nil {
		return err
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(setConfigurationSummaryQuery, repositoryID, payload)); err != nil {
		return err
	}

	return nil
}

//
// TODO - handle expiration, conflicts

const setConfigurationSummaryQuery = `
INSERT INTO cached_available_indexers (repository_id, available_indexers) VALUES (%s, %s) ON CONFLICT DO NOTHING
`
