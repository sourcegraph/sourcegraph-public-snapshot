package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store provides the interface for codenav storage.
type Store interface {
	GetUnsafeDB() database.DB
	GetUploadsForRepository(ctx context.Context, repositoryID int) ([]int, error)
}

// store manages the codenav store.
type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new codenav store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("codenav.store", ""),
		operations: newOperations(observationCtx),
	}
}

// GetUnsafeDB returns the underlying database handle. This is used by the
// resolvers that have the old convention of using the database handle directly.
func (s *store) GetUnsafeDB() database.DB {
	return database.NewDBWith(s.logger, s.db)
}

func (s *store) GetUploadsForRepository(ctx context.Context, repositoryID int) ([]int, error) {
	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(getUploadsForRepositoryQuery, repositoryID)))
}

const getUploadsForRepositoryQuery = `
SELECT u.id
FROM lsif_uploads u
JOIN repo r ON r.id = u.repository_id
JOIN lsif_uploads_visible_at_tip uvt ON uvt.upload_id = u.id
WHERE
	r.id = %s AND
	r.deleted_at IS NULL AND
	r.blocked IS NULL AND
	uvt.is_default_branch
`
