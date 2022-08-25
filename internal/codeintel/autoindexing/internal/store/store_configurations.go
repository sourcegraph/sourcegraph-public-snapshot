package store

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetIndexConfigurationByRepositoryID returns the index configuration for a repository.
func (s *store) GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return scanFirstIndexConfiguration(s.db.Query(ctx, sqlf.Sprintf(getIndexConfigurationByRepositoryIDQuery, repositoryID)))
}

const getIndexConfigurationByRepositoryIDQuery = `
-- source: internal/codeintel/stores/dbstore/configuration.go:GetIndexConfigurationByRepositoryID
SELECT
	c.id,
	c.repository_id,
	c.data
FROM lsif_index_configuration c WHERE c.repository_id = %s
`

// UpdateIndexConfigurationByRepositoryID updates the index configuration for a repository.
func (s *store) UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error) {
	ctx, _, endObservation := s.operations.updateIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(updateIndexConfigurationByRepositoryIDQuery, repositoryID, data, data))
}

const updateIndexConfigurationByRepositoryIDQuery = `
-- source: internal/codeintel/stores/dbstore/configuration.go:UpdateIndexConfigurationByRepositoryID
INSERT INTO lsif_index_configuration (repository_id, data) VALUES (%s, %s)
	ON CONFLICT (repository_id) DO UPDATE SET data = %s
`
