package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for autoindexing storage.
type Store interface {
	List(ctx context.Context, opts ListOpts) (indexJobs []shared.IndexJob, err error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)
	StaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration) (indexesDeleted int, err error)
}

// store manages the autoindexing store.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new autoindexing store.
func New(db database.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationContext),
	}
}

// ListOpts specifies options for listing index jobs.
type ListOpts struct {
	Limit int
}

// List returns the list of index jobs.
func (s *store) List(ctx context.Context, opts ListOpts) (indexJobs []shared.IndexJob, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numIndexJobs", len(indexJobs)),
		}})
	}()

	// This is only a stub and will be replaced or significantly modified
	// in https://github.com/sourcegraph/sourcegraph/issues/33377
	_, _ = scanIndexJobs(s.db.Query(ctx, sqlf.Sprintf(listQuery, opts.Limit)))
	return nil, errors.Newf("unimplemented: autoindexing.store.List")
}

const listQuery = `
-- source: internal/codeintel/autoindexing/store/store.go:List
SELECT id FROM TODO
LIMIT %s
`
