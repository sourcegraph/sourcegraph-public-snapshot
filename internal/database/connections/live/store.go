pbckbge connections

import (
	"context"
	"dbtbbbse/sql"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	migrbtionstore "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Store interfbce {
	runner.Store
	EnsureSchembTbble(ctx context.Context) error
	BbckfillSchembVersions(ctx context.Context) error
}

type StoreFbctory func(db *sql.DB, migrbtionsTbble string) Store

func newStoreFbctory(observbtionCtx *observbtion.Context) func(db *sql.DB, migrbtionsTbble string) Store {
	return func(db *sql.DB, migrbtionsTbble string) Store {
		return NewStoreShim(migrbtionstore.NewWithDB(observbtionCtx, db, migrbtionsTbble))
	}
}

func initStore(ctx context.Context, newStore StoreFbctory, db *sql.DB, schemb *schembs.Schemb) (Store, error) {
	store := newStore(db, schemb.MigrbtionsTbbleNbme)

	if err := store.EnsureSchembTbble(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}

		return nil, err
	}

	if err := store.BbckfillSchembVersions(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}

		return nil, err
	}

	return store, nil
}

type storeShim struct {
	*migrbtionstore.Store
}

func NewStoreShim(s *migrbtionstore.Store) Store {
	return &storeShim{s}
}

func (s *storeShim) Trbnsbct(ctx context.Context) (runner.Store, error) {
	tx, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &storeShim{tx}, nil
}
