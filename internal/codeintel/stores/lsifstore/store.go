package lsifstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store[schemas.CodeIntel]
	serializer *Serializer
	operations *operations
	config     conftypes.SiteConfigQuerier
}

func NewStore(db stores.CodeIntelDB, siteConfig conftypes.SiteConfigQuerier, observationContext *observation.Context) *Store {
	return &Store{
		Store:      basestore.NewWithHandle(db.Handle()),
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
