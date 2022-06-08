package store

import (
	"context"
	"database/sql"

	"github.com/inconshreveable/log15"
	"github.com/jmoiron/sqlx"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/parser"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Store interface {
	Close() error
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	Search(ctx context.Context, args search.SymbolsParameters) ([]result.Symbol, error)

	CreateMetaTable(ctx context.Context) error
	GetCommit(ctx context.Context) (string, bool, error)
	InsertMeta(ctx context.Context, commitID string) error
	UpdateMeta(ctx context.Context, commitID string) error

	CreateSymbolsTable(ctx context.Context) error
	CreateSymbolIndexes(ctx context.Context) error
	DeletePaths(ctx context.Context, paths []string) error
	WriteSymbols(ctx context.Context, symbolOrErrors <-chan parser.SymbolOrError) error
}

type store struct {
	db *sqlx.DB
	*basestore.Store
}

func NewStore(dbFile string) (Store, error) {
	db, err := sqlx.Open("sqlite3_with_regexp", dbFile)
	if err != nil {
		return nil, err
	}

	return &store{
		db:    db,
		Store: basestore.NewWithDB(db, sql.TxOptions{}),
	}, nil
}

func (s *store) Close() error {
	return s.db.Close()
}

func (s *store) Transact(ctx context.Context) (Store, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{db: s.db, Store: tx}, nil
}

func WithSQLiteStore(dbFile string, callback func(db Store) error) error {
	db, err := NewStore(dbFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log15.Error("Failed to close database", "filename", dbFile, "error", err)
		}
	}()

	return callback(db)
}

func WithSQLiteStoreTransaction(ctx context.Context, dbFile string, callback func(db Store) error) error {
	return WithSQLiteStore(dbFile, func(db Store) (err error) {
		tx, err := db.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		return callback(tx)
	})
}
