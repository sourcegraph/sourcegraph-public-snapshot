package database

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	newest, err := findNewestFile(filepath.Join(w.path, diskcache.EncodeKeyComponent(string(args.Repo))))
	if err != nil {
		return err
	}

	if newest == "" {
		symbols, err := w.parser.Parse(ctx, args, nil)
		if err != nil {
			return err
		}

		return w.writeSymbols(ctx, args, dbFile, symbols)
	}

	oldCommit, ok, err := w.getCommit(ctx, newest)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("no old commit")
	}

	// git diff
	changes, err := w.gitserverClient.GitDiff(ctx, args.Repo, api.CommitID(oldCommit), args.CommitID)
	if err != nil {
		return err
	}
	paths := append(changes.Deleted, changes.Modified...)

	symbols, err := w.parser.Parse(ctx, args, paths)
	if err != nil {
		return err
	}

	// Copy the existing DB to a new DB and update the new DB
	if err := copyFile(newest, dbFile); err != nil {
		return err
	}

	return w.updateSymbols(ctx, args, dbFile, symbols, paths)
}

func (w *databaseWriter) getCommit(ctx context.Context, dbFile string) (commit string, ok bool, err error) {
	err = store.WithSQLiteStore(dbFile, func(db store.Store) error {
		commit, ok, err = db.GetCommit(ctx)
		return err
	})

	return
}

func (w *databaseWriter) writeSymbols(ctx context.Context, args types.SearchArgs, dbFile string, symbols <-chan result.Symbol) (err error) {
	return store.WithSQLiteStoreTransaction(ctx, dbFile, func(tx store.Store) error {
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

func (w *databaseWriter) updateSymbols(ctx context.Context, args types.SearchArgs, dbFile string, symbols <-chan result.Symbol, paths []string) error {
	return store.WithSQLiteStoreTransaction(ctx, dbFile, func(tx store.Store) error {
		if err := tx.UpdateMeta(ctx, string(args.CommitID)); err != nil {
			return err
		}
		if err := tx.DeletePaths(ctx, paths); err != nil {
			return err
		}
		if err := tx.WriteSymbols(ctx, symbols); err != nil {
			return err
		}

		return nil
	})
}

// findNewestFile lists the directory and returns the newest file's path, prepended with dir.
func findNewestFile(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", nil
	}

	var mostRecentTime time.Time
	newest := ""
	for _, fi := range files {
		if fi.Type().IsRegular() {
			if !strings.HasSuffix(fi.Name(), ".zip") {
				continue
			}

			info, err := fi.Info()
			if err != nil {
				return "", err
			}

			if newest == "" || info.ModTime().After(mostRecentTime) {
				mostRecentTime = info.ModTime()
				newest = filepath.Join(dir, fi.Name())
			}
		}
	}

	return newest, nil
}

// copyFile is like the cp command.
func copyFile(from string, to string) error {
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	return err
}
