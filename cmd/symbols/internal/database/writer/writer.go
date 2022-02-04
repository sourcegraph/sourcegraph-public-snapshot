package writer

import (
	"context"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api/observability"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/store"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
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
		if ok, err := w.writeFileIncrementally(ctx, args, dbFile, newestDBFile, oldCommit); err != nil || ok {
			return err
		}
	}

	return w.writeDBFile(ctx, args, dbFile)
}

func (w *databaseWriter) getNewestCommit(ctx context.Context, args types.SearchArgs) (dbFile string, commit string, ok bool, err error) {
	newest, err := findNewestFile(filepath.Join(w.path, diskcache.EncodeKeyComponent(string(args.Repo))))
	if err != nil || newest == "" {
		return "", "", false, err
	}

	err = store.WithSQLiteStore(newest, func(db store.Store) (err error) {
		if commit, ok, err = db.GetCommit(ctx); err != nil {
			return errors.Wrap(err, "store.GetCommit")
		}

		return nil
	})

	return newest, commit, ok, err
}

func (w *databaseWriter) writeDBFile(ctx context.Context, args types.SearchArgs, dbFile string) error {
	observability.SetParseAmount(ctx, observability.FullParse)

	return w.parseAndWriteInTransaction(ctx, args, nil, dbFile, func(tx store.Store, symbolOrErrors <-chan parser.SymbolOrError) error {
		if err := tx.CreateMetaTable(ctx); err != nil {
			return errors.Wrap(err, "store.CreateMetaTable")
		}
		if err := tx.CreateSymbolsTable(ctx); err != nil {
			return errors.Wrap(err, "store.CreateSymbolsTable")
		}
		if err := tx.InsertMeta(ctx, string(args.CommitID)); err != nil {
			return errors.Wrap(err, "store.InsertMeta")
		}
		if err := tx.WriteSymbols(ctx, symbolOrErrors); err != nil {
			return errors.Wrap(err, "store.WriteSymbols")
		}
		if err := tx.CreateSymbolIndexes(ctx); err != nil {
			return errors.Wrap(err, "store.CreateSymbolIndexes")
		}

		return nil
	})
}

// The maximum number of paths when doing incremental indexing. Diffs with more paths than this will
// not be incrementally indexed, and instead we will process all symbols.
const maxTotalPaths = 999

// The maximum sum of bytes in paths in a diff when doing incremental indexing. Diffs bigger than this
// will not be incrementally indexed, and instead we will process all symbols. Without this limit, we
// could hit the error "argument list too long" by exceeding the limit on the number of arguments to a
// command.
//
// Mac  : getconf ARG_MAX returns 1,048,576
// Linux: getconf ARG_MAX returns 2,097,152
//
// We want to remain well under that limit, so 100,000 seems safe.
const maxTotalPathsLength = 100_000

func (w *databaseWriter) writeFileIncrementally(ctx context.Context, args types.SearchArgs, dbFile, newestDBFile, oldCommit string) (bool, error) {
	observability.SetParseAmount(ctx, observability.PartialParse)

	changes, err := w.gitserverClient.GitDiff(ctx, args.Repo, api.CommitID(oldCommit), args.CommitID)
	if err != nil {
		return false, errors.Wrap(err, "gitserverClient.GitDiff")
	}

	// Paths to re-parse
	addedOrModifiedPaths := append(changes.Added, changes.Modified...)

	// Paths to modify in the database
	addedModifiedOrDeletedPaths := append(addedOrModifiedPaths, changes.Deleted...)

	// Too many entries
	if len(addedModifiedOrDeletedPaths) > maxTotalPaths {
		return false, nil
	}

	totalPathsLength := 0
	for _, path := range addedModifiedOrDeletedPaths {
		totalPathsLength += len(path)
	}
	// Argument lists too long
	if totalPathsLength > maxTotalPathsLength {
		return false, nil
	}

	if err := copyFile(newestDBFile, dbFile); err != nil {
		return false, err
	}

	return true, w.parseAndWriteInTransaction(ctx, args, addedOrModifiedPaths, dbFile, func(tx store.Store, symbolOrErrors <-chan parser.SymbolOrError) error {
		if err := tx.UpdateMeta(ctx, string(args.CommitID)); err != nil {
			return errors.Wrap(err, "store.UpdateMeta")
		}
		if err := tx.DeletePaths(ctx, addedModifiedOrDeletedPaths); err != nil {
			return errors.Wrap(err, "store.DeletePaths")
		}
		if err := tx.WriteSymbols(ctx, symbolOrErrors); err != nil {
			return errors.Wrap(err, "store.WriteSymbols")
		}

		return nil
	})
}

func (w *databaseWriter) parseAndWriteInTransaction(ctx context.Context, args types.SearchArgs, paths []string, dbFile string, callback func(tx store.Store, symbolOrErrors <-chan parser.SymbolOrError) error) (err error) {
	symbolOrErrors, err := w.parser.Parse(ctx, args, paths)
	if err != nil {
		return errors.Wrap(err, "parser.Parse")
	}
	defer func() {
		if err != nil {
			go func() {
				// Drain channel on early exit
				for range symbolOrErrors {
				}
			}()
		}
	}()

	return store.WithSQLiteStoreTransaction(ctx, dbFile, func(tx store.Store) error {
		return callback(tx, symbolOrErrors)
	})
}
