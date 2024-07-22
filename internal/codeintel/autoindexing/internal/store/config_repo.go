package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) RepositoryExceptions(ctx context.Context, repositoryID int) (canSchedule, canInfer bool, err error) {
	ctx, _, endObservation := s.operations.repositoryExceptions.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(repositoryExceptionsQuery, repositoryID))
	if err != nil {
		return false, false, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var disableSchedule, disableInference bool
	for rows.Next() {
		if err := rows.Scan(&disableSchedule, &disableInference); err != nil {
			return false, false, err
		}
	}

	return !disableSchedule, !disableInference, rows.Err()
}

const repositoryExceptionsQuery = `
SELECT
	cae.disable_scheduling,
	cae.disable_inference
FROM codeintel_autoindexing_exceptions cae
WHERE cae.repository_id = %s
`

func (s *store) SetRepositoryExceptions(ctx context.Context, repositoryID int, canSchedule, canInfer bool) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryExceptions.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(
		setRepositoryExceptionsQuery,
		repositoryID,
		!canSchedule, !canInfer,
		!canSchedule, !canInfer,
	))
}

const setRepositoryExceptionsQuery = `
INSERT INTO codeintel_autoindexing_exceptions (repository_id, disable_scheduling, disable_inference)
VALUES (%s, %s, %s)
ON CONFLICT (repository_id) DO UPDATE SET
	disable_scheduling = %s,
	disable_inference = %s
`

func (s *store) GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return scanFirstIndexConfiguration(s.db.Query(ctx, sqlf.Sprintf(getIndexConfigurationByRepositoryIDQuery, repositoryID)))
}

const getIndexConfigurationByRepositoryIDQuery = `
SELECT
	c.id,
	c.repository_id,
	c.data
FROM lsif_index_configuration c
WHERE c.repository_id = %s
`

func (s *store) UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error) {
	ctx, _, endObservation := s.operations.updateIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.Int("dataSize", len(data)),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(updateIndexConfigurationByRepositoryIDQuery, repositoryID, data, data))
}

const updateIndexConfigurationByRepositoryIDQuery = `
INSERT INTO lsif_index_configuration (repository_id, data)
VALUES (%s, %s)
ON CONFLICT (repository_id) DO UPDATE
SET data = %s
`

//
//

func scanJobConfiguration(s dbutil.Scanner) (indexConfiguration shared.IndexConfiguration, err error) {
	return indexConfiguration, s.Scan(
		&indexConfiguration.ID,
		&indexConfiguration.RepositoryID,
		&indexConfiguration.Data,
	)
}

var scanFirstIndexConfiguration = basestore.NewFirstScanner(scanJobConfiguration)
