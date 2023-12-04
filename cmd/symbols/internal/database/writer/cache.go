package writer

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api/observability"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CachedDatabaseWriter interface {
	GetOrCreateDatabaseFile(ctx context.Context, args search.SymbolsParameters) (string, error)
}

type cachedDatabaseWriter struct {
	databaseWriter DatabaseWriter
	cache          diskcache.Store
}

func NewCachedDatabaseWriter(databaseWriter DatabaseWriter, cache diskcache.Store) CachedDatabaseWriter {
	return &cachedDatabaseWriter{
		databaseWriter: databaseWriter,
		cache:          cache,
	}
}

// The version of the symbols database schema. This is included in the database filenames to prevent a
// newer version of the symbols service from attempting to read from a database created by an older and
// likely incompatible symbols service. Increment this when you change the database schema.
const symbolsDBVersion = 5

func (w *cachedDatabaseWriter) GetOrCreateDatabaseFile(ctx context.Context, args search.SymbolsParameters) (string, error) {
	// set to noop parse originally, this will be overridden if the fetcher func below is called
	observability.SetParseAmount(ctx, observability.CachedParse)
	cacheFile, err := w.cache.OpenWithPath(ctx, repoCommitKey(args.Repo, args.CommitID), func(fetcherCtx context.Context, tempDBFile string) error {
		if err := w.databaseWriter.WriteDBFile(fetcherCtx, args, tempDBFile); err != nil {
			return errors.Wrap(err, "databaseWriter.WriteDBFile")
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	defer cacheFile.File.Close()

	return cacheFile.File.Name(), err
}

// repoCommitKey returns the diskcache key for a repo and commit (points to a SQLite DB file).
func repoCommitKey(repo api.RepoName, commitID api.CommitID) []string {
	return []string{
		fmt.Sprint(symbolsDBVersion),
		string(repo),
		string(commitID),
	}
}

// repoKey returns the diskcache key for a repo (points to a directory).
func repoKey(repo api.RepoName) []string {
	return []string{
		fmt.Sprint(symbolsDBVersion),
		string(repo),
	}
}
