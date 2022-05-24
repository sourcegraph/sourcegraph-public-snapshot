package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for symbols storage.
type Store interface {
	List(ctx context.Context, opts ListOpts) (symbols []shared.Symbol, err error)
}

// store manages the symbols store.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new symbols store.
func New(db dbutil.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithDB(db, sql.TxOptions{}),
		operations: newOperations(observationContext),
	}
}

// Transact returns a new symbols store that is a transaction on the given store.
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

// ListOpts are options for listing symbols.
type ListOpts struct {
	Limit int
}

// List returns the symbols in the store.
func (s *store) List(ctx context.Context, opts ListOpts) (symbols []shared.Symbol, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numSymbols", len(symbols)),
		}})
	}()

	// This is only a stub and will be replaced or significantly modified
	// in https://github.com/sourcegraph/sourcegraph/issues/33374
	_, _ = scanSymbols(s.db.Query(ctx, sqlf.Sprintf(listQuery, opts.Limit)))
	return nil, errors.Newf("unimplemented: symbols.store.List")
}

const listQuery = `
-- source: internal/codeintel/symbols/internal/store/store.go:List
SELECT name FROM TODO
LIMIT %d
`
