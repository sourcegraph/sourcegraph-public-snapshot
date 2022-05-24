package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for documents storage.
type Store interface {
	List(ctx context.Context, opts ListOpts) (documents []shared.Document, err error)
}

// store manages the documents store.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new documents store.
func New(db dbutil.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithDB(db, sql.TxOptions{}),
		operations: newOperations(observationContext),
	}
}

// Transact returns a new store transaction.
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

// ListOpts specifies options for listing documents.
type ListOpts struct {
	Limit int
}

// List returns the list of documents.
func (s *store) List(ctx context.Context, opts ListOpts) (documents []shared.Document, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDocuments", len(documents)),
		}})
	}()

	// This is only a stub and will be replaced or significantly modified
	// in https://github.com/sourcegraph/sourcegraph/issues/33373
	_, _ = scanDocuments(s.db.Query(ctx, sqlf.Sprintf(listQuery, opts.Limit)))
	return nil, errors.Newf("unimplemented: documents.store.List")
}

const listQuery = `
-- source: internal/codeintel/documents/internal/store/store.go:List
SELECT path FROM TODO
LIMIT %d
`
