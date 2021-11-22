package writer

import (
	"context"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/store"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type DatabaseWriter interface {
	WriteDBFile(ctx context.Context, args types.SearchArgs, tempDBFile string) error
}

type databaseWriter struct {
	path            string
	gitserverClient gitserver.GitserverClient
	parser          parser.Parser
}

func NewDatabaseWriter(
	path string,
	gitserverClient gitserver.GitserverClient,
	parser parser.Parser,
) DatabaseWriter {
	return &databaseWriter{
		path:            path,
		gitserverClient: gitserverClient,
		parser:          parser,
	}
}

func (w *databaseWriter) WriteDBFile(ctx context.Context, args types.SearchArgs, dbFile string) error {
	if newestDBFile, oldCommit, ok, err := w.getNewestCommit(ctx, args); err != nil {
		return err
	} else if ok {
		return w.writeFileIncrementally(ctx, args, dbFile, newestDBFile, oldCommit)
	}

	return w.writeDBFile(ctx, args, dbFile)
}

func (w *databaseWriter) getNewestCommit(ctx context.Context, args types.SearchArgs) (dbFile string, commit string, ok bool, err error) {
	newest, err := findNewestFile(filepath.Join(w.path, diskcache.EncodeKeyComponent(string(args.Repo))))
	if err != nil || newest == "" {
		return "", "", false, err
	}

	err = store.WithSQLiteStore(dbFile, func(db store.Store) error {
		commit, ok, err = db.GetCommit(ctx)
		return err
	})
	return newest, commit, ok, err
}

func (w *databaseWriter) writeDBFile(ctx context.Context, args types.SearchArgs, dbFile string) error {
	return w.parseAndWriteInTransaction(ctx, args, nil, dbFile, func(tx store.Store, symbols <-chan result.Symbol) error {
		if err := tx.CreateMetaTable(ctx); err != nil {
			return err
		}
		if err := tx.CreateSymbolsTable(ctx); err != nil {
			return err
		}
		if err := tx.InsertMeta(ctx, string(args.CommitID)); err != nil {
			return err
		}
		if err := tx.WriteSymbols(ctx, symbols); err != nil {
			return err
		}
		if err := tx.CreateSymbolIndexes(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (w *databaseWriter) writeFileIncrementally(ctx context.Context, args types.SearchArgs, dbFile, newestDBFile, oldCommit string) error {
	changes, err := w.gitserverClient.GitDiff(ctx, args.Repo, api.CommitID(oldCommit), args.CommitID)
	if err != nil {
		return err
	}
	pathsToParse := append(changes.Deleted, changes.Modified...)

	if err := copyFile(newestDBFile, dbFile); err != nil {
		return err
	}

	return w.parseAndWriteInTransaction(ctx, args, pathsToParse, dbFile, func(tx store.Store, symbols <-chan result.Symbol) error {
		if err := tx.UpdateMeta(ctx, string(args.CommitID)); err != nil {
			return err
		}
		if err := tx.DeletePaths(ctx, pathsToParse); err != nil {
			return err
		}
		if err := tx.WriteSymbols(ctx, symbols); err != nil {
			return err
		}

		return nil
	})
}

func (w *databaseWriter) parseAndWriteInTransaction(ctx context.Context, args types.SearchArgs, paths []string, dbFile string, callback func(tx store.Store, symbols <-chan result.Symbol) error) error {
	symbols, err := w.parser.Parse(ctx, args, paths)
	if err != nil {
		return err
	}

	return store.WithSQLiteStoreTransaction(ctx, dbFile, func(tx store.Store) error {
		return callback(tx, symbols)
	})
}
