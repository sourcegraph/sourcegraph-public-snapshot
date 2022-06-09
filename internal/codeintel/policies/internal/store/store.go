package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for policies storage.
type Store interface {
	List(ctx context.Context, opts ListOpts) (policies []shared.Policy, err error)
}

// store manages the policies store.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new policies store.
func New(db dbutil.DB, observationContext *observation.Context) *store {
	return &store{
		db:         basestore.NewWithDB(db, sql.TxOptions{}),
		operations: newOperations(observationContext),
	}
}

// Transaction returns a transaction for the store.
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

// ListOpts specifies options for listing policies.
type ListOpts struct {
	Limit int
}

// List returns a list of policies.
func (s *store) List(ctx context.Context, opts ListOpts) (policies []shared.Policy, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numPolicies", len(policies)),
		}})
	}()

	// This is only a stub and will be replaced or significantly modified
	// in https://github.com/sourcegraph/sourcegraph/issues/33376
	_, _ = scanPolicies(s.db.Query(ctx, sqlf.Sprintf(listQuery, opts.Limit)))
	return nil, errors.Newf("unimplemented: policies.store.List")
}

const listQuery = `
-- source: internal/codeintel/policies/store/internal/store.go:List
SELECT id FROM TODO
LIMIT %s
`
