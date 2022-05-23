package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for uploads storage.
type Store interface {
	List(ctx context.Context, opts ListOpts) (uploads []shared.Upload, err error)
}

// store manages the database operations for uploads.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new uploads store.
func New(db dbutil.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithDB(db, sql.TxOptions{}),
		operations: newOperations(observationContext),
	}
}

// Transact returns a store with a transaction.
func (s *store) Transact(ctx context.Context) (*store, error) {
	txBase, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		db:         txBase,
		operations: s.operations,
	}, nil
}

// ListOpts specifies options for listing uploads.
type ListOpts struct {
	Limit int
}

// List returns a list of uploads.
func (s *store) List(ctx context.Context, opts ListOpts) (uploads []shared.Upload, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numUploads", len(uploads)),
		}})
	}()

	// This is only a stub and will be replaced or significantly modified
	// in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = scanUploads(s.db.Query(ctx, sqlf.Sprintf(listQuery, opts.Limit)))
	return nil, errors.Newf("unimplemented: uploads.store.List")
}

const listQuery = `
-- source: internal/codeintel/uploads/internal/store/store.go:List
SELECT id FROM TODO
LIMIT %s
`
