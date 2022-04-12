package lsifstore

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store
	serializer *Serializer
	operations *operations
	config     conftypes.SiteConfigQuerier
}

func NewStore(db dbutil.DB, siteConfig conftypes.SiteConfigQuerier, observationContext *observation.Context) *Store {
	return &Store{
		Store:      basestore.NewWithHandle(basestore.NewHandleWithDB(db, sql.TxOptions{})),
		serializer: NewSerializer(),
		operations: newOperations(observationContext),
		config:     siteConfig,
	}
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store:      tx,
		serializer: s.serializer,
		operations: s.operations,
		config:     s.config,
	}, nil
}

func (s *Store) Done(err error) error {
	return s.Store.Done(err)
}
