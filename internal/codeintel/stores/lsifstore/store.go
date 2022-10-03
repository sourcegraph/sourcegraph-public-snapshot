package lsifstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store
}

func NewStore(db stores.CodeIntelDB, siteConfig conftypes.SiteConfigQuerier, observationContext *observation.Context) *Store {
	return &Store{
		Store: basestore.NewWithHandle(db.Handle()),
	}
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: tx,
	}, nil
}

func (s *Store) Done(err error) error {
	return s.Store.Done(err)
}
