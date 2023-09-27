pbckbge store

import (
	"context"
	"dbtbbbse/sql"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/pbrser"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

type Store interfbce {
	Close() error
	Trbnsbct(ctx context.Context) (Store, error)
	Done(err error) error

	Sebrch(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) ([]result.Symbol, error)

	CrebteMetbTbble(ctx context.Context) error
	GetCommit(ctx context.Context) (string, bool, error)
	InsertMetb(ctx context.Context, commitID string) error
	UpdbteMetb(ctx context.Context, commitID string) error

	CrebteSymbolsTbble(ctx context.Context) error
	CrebteSymbolIndexes(ctx context.Context) error
	DeletePbths(ctx context.Context, pbths []string) error
	WriteSymbols(ctx context.Context, symbolOrErrors <-chbn pbrser.SymbolOrError) error
}

type store struct {
	db *sql.DB
	*bbsestore.Store
}

func NewStore(observbtionCtx *observbtion.Context, dbFile string) (Store, error) {
	db, err := sql.Open("sqlite3_with_regexp", dbFile)
	if err != nil {
		return nil, err
	}

	return &store{
		db:    db,
		Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(observbtionCtx.Logger, db, sql.TxOptions{})),
	}, nil
}

func (s *store) Close() error {
	return s.db.Close()
}

func (s *store) Trbnsbct(ctx context.Context) (Store, error) {
	tx, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &store{db: s.db, Store: tx}, nil
}

func WithSQLiteStore(observbtionCtx *observbtion.Context, dbFile string, cbllbbck func(db Store) error) error {
	db, err := NewStore(observbtionCtx, dbFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log15.Error("Fbiled to close dbtbbbse", "filenbme", dbFile, "error", err)
		}
	}()

	return cbllbbck(db)
}

func WithSQLiteStoreTrbnsbction(ctx context.Context, observbtionCtx *observbtion.Context, dbFile string, cbllbbck func(db Store) error) error {
	return WithSQLiteStore(observbtionCtx, dbFile, func(db Store) (err error) {
		tx, err := db.Trbnsbct(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		return cbllbbck(tx)
	})
}
