package rockskip

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/inconshreveable/log15"
	pg "github.com/lib/pq"
	"k8s.io/utils/lru"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Service) Index(ctx context.Context, db database.DB, repo, givenCommit string) (err error) {
	threadStatus := s.status.NewThreadStatus(fmt.Sprintf("indexing %s@%s", repo, givenCommit))
	defer threadStatus.End()

	tasklog := threadStatus.Tasklog

	// Get a fresh connection from the DB pool to get deterministic "lock stacking" behavior.
	// See doc/dev/background-information/sql/locking_behavior.md for more details.
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get connection for indexing")
	}
	defer conn.Close()

	// Acquire the indexing lock on the repo.
	releaseLock, err := iLock(ctx, conn, threadStatus, repo)
	if err != nil {
		return err
	}
	defer func() { err = errors.CombineErrors(err, releaseLock()) }()

	tipCommit := NULL
	tipCommitHash := ""
	tipHeight := 0

	var repoId int
	err = conn.QueryRowContext(ctx, "SELECT id FROM rockskip_repos WHERE repo = $1", repo).Scan(&repoId)
	if err != nil {
		return errors.Wrapf(err, "failed to get repo id for %s", repo)
	}

	missingCount := 0
	tasklog.Start("RevList")
	err = s.git.RevListEach(repo, db, givenCommit, func(commitHash string) (shouldContinue bool, err error) {
		defer tasklog.Continue("RevList")

		tasklog.Start("GetCommitByHash")
		commit, height, present, err := GetCommitByHash(ctx, conn, repoId, commitHash)
		if err != nil {
			return false, err
		} else if present {
			tipCommit = commit
			tipCommitHash = commitHash
			tipHeight = height
			return false, nil
		}
		missingCount += 1
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "RevList")
	}

	threadStatus.SetProgress(0, missingCount)

	if missingCount == 0 {
		return nil
	}

	parse := s.createParser()

	symbolCache := newSymbolIdCache(s.symbolsCacheSize)
	pathSymbolsCache := newPathSymbolsCache(s.pathSymbolsCacheSize)

	tasklog.Start("Log")
	entriesIndexed := 0
	err = s.git.LogReverseEach(repo, db, givenCommit, missingCount, func(entry gitdomain.LogEntry) error {
		defer tasklog.Continue("Log")

		threadStatus.SetProgress(entriesIndexed, missingCount)
		entriesIndexed++

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "begin transaction")
		}
		defer tx.Rollback()

		hops, err := getHops(ctx, tx, tipCommit, tasklog)
		if err != nil {
			return errors.Wrap(err, "getHops")
		}

		r := ruler(tipHeight + 1)
		if r >= len(hops) {
			return errors.Newf("ruler(%d) = %d is out of range of len(hops) = %d", tipHeight+1, r, len(hops))
		}

		tasklog.Start("InsertCommit")
		commit, err := InsertCommit(ctx, tx, repoId, entry.Commit, tipHeight+1, hops[r])
		if err != nil {
			return errors.Wrap(err, "InsertCommit")
		}

		tasklog.Start("AppendHop+")
		err = AppendHop(ctx, tx, repoId, hops[0:r], AddedAD, DeletedAD, commit)
		if err != nil {
			return errors.Wrap(err, "AppendHop (added)")
		}
		tasklog.Start("AppendHop-")
		err = AppendHop(ctx, tx, repoId, hops[0:r], DeletedAD, AddedAD, commit)
		if err != nil {
			return errors.Wrap(err, "AppendHop (deleted)")
		}

		deletedPaths := []string{}
		addedPaths := []string{}
		for _, pathStatus := range entry.PathStatuses {
			if pathStatus.Status == gitdomain.DeletedAMD || pathStatus.Status == gitdomain.ModifiedAMD {
				deletedPaths = append(deletedPaths, pathStatus.Path)
			}
			if pathStatus.Status == gitdomain.AddedAMD || pathStatus.Status == gitdomain.ModifiedAMD {
				addedPaths = append(addedPaths, pathStatus.Path)
			}
		}

		getSymbols := func(commit string, paths []string) (map[string]map[string]struct{}, error) {
			pathToSymbols := map[string]map[string]struct{}{}
			pathsToFetchSet := map[string]struct{}{}
			for _, path := range paths {
				pathsToFetchSet[path] = struct{}{}
			}

			// Don't fetch files that are already in the cache.
			if commit == tipCommitHash {
				for _, path := range paths {
					if symbols, ok := pathSymbolsCache.get(path); ok {
						pathToSymbols[path] = symbols
						delete(pathsToFetchSet, path)
					}
				}
			}

			pathsToFetch := []string{}
			for path := range pathsToFetchSet {
				pathsToFetch = append(pathsToFetch, path)
			}

			tasklog.Start("ArchiveEach")
			err = s.git.ArchiveEach(repo, commit, pathsToFetch, func(path string, contents []byte) error {
				defer tasklog.Continue("ArchiveEach")

				tasklog.Start("parse")
				symbols, err := parse(path, contents)
				if err != nil {
					return errors.Wrap(err, "parse")
				}

				pathToSymbols[path] = map[string]struct{}{}
				for _, symbol := range symbols {
					pathToSymbols[path][symbol.Name] = struct{}{}
				}

				return nil
			})

			if err != nil {
				return nil, errors.Wrap(err, "while looping ArchiveEach")
			}

			// Cache the symbols we just parsed.
			if commit != tipCommitHash {
				for path, symbols := range pathToSymbols {
					pathSymbolsCache.set(path, symbols)
				}
			}

			return pathToSymbols, nil
		}

		symbolsFromDeletedFiles, err := getSymbols(tipCommitHash, deletedPaths)
		if err != nil {
			return errors.Wrap(err, "getSymbols (deleted)")
		}
		symbolsFromAddedFiles, err := getSymbols(entry.Commit, addedPaths)
		if err != nil {
			return errors.Wrap(err, "getSymbols (added)")
		}

		// Compute the symmetric difference of symbols between the added and deleted paths.
		deletedSymbols := map[string]map[string]struct{}{}
		addedSymbols := map[string]map[string]struct{}{}
		for _, pathStatus := range entry.PathStatuses {
			switch pathStatus.Status {
			case gitdomain.DeletedAMD:
				deletedSymbols[pathStatus.Path] = symbolsFromDeletedFiles[pathStatus.Path]
			case gitdomain.AddedAMD:
				addedSymbols[pathStatus.Path] = symbolsFromAddedFiles[pathStatus.Path]
			case gitdomain.ModifiedAMD:
				deletedSymbols[pathStatus.Path] = map[string]struct{}{}
				addedSymbols[pathStatus.Path] = map[string]struct{}{}
				for name := range symbolsFromDeletedFiles[pathStatus.Path] {
					if _, ok := symbolsFromAddedFiles[pathStatus.Path][name]; !ok {
						deletedSymbols[pathStatus.Path][name] = struct{}{}
					}
				}
				for name := range symbolsFromAddedFiles[pathStatus.Path] {
					if _, ok := symbolsFromDeletedFiles[pathStatus.Path][name]; !ok {
						addedSymbols[pathStatus.Path][name] = struct{}{}
					}
				}
			}
		}

		for path, symbols := range deletedSymbols {
			for symbol := range symbols {
				id := 0
				ok := false
				if id, ok = symbolCache.get(path, symbol); !ok {
					tasklog.Start("GetSymbol")
					found := false
					id, found, err = GetSymbol(ctx, tx, repoId, path, symbol, hops)
					if !found {
						// We did not find the symbol that (supposedly) has been deleted, so ignore the
						// deletion. This will probably lead to extra symbols in search results.
						//
						// The last time this happened, it was caused by impurity in ctags where the
						// result of parsing a file was affected by previously parsed files and not fully
						// determined by the file itself:
						//
						// https://github.com/universal-ctags/ctags/pull/3300
						log15.Error("Could not find symbol that was supposedly deleted", "repo", repo, "commit", commit, "path", path, "symbol", symbol)
						continue
					}
				}

				tasklog.Start("UpdateSymbolHops")
				err = UpdateSymbolHops(ctx, tx, id, DeletedAD, commit)
				if err != nil {
					return errors.Wrap(err, "UpdateSymbolHops")
				}
			}
		}

		tasklog.Start("BatchInsertSymbols")
		err = BatchInsertSymbols(ctx, tasklog, tx, repoId, commit, symbolCache, addedSymbols)
		if err != nil {
			return errors.Wrap(err, "BatchInsertSymbols")
		}

		tasklog.Start("DeleteRedundant")
		err = DeleteRedundant(ctx, tx, commit)
		if err != nil {
			return errors.Wrap(err, "DeleteRedundant")
		}

		tasklog.Start("CommitTx")
		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "commit transaction")
		}

		tipCommit = commit
		tipCommitHash = entry.Commit
		tipHeight += 1

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "LogReverseEach")
	}

	threadStatus.SetProgress(entriesIndexed, missingCount)

	return nil
}

func BatchInsertSymbols(ctx context.Context, tasklog *TaskLog, tx *sql.Tx, repoId, commit int, symbolCache *symbolIdCache, symbols map[string]map[string]struct{}) error {
	callback := func(inserter *batch.Inserter) error {
		for path, pathSymbols := range symbols {
			for symbol := range pathSymbols {
				if err := inserter.Insert(ctx, pg.Array([]int{commit}), pg.Array([]int{}), repoId, path, symbol); err != nil {
					return err
				}
			}
		}

		return nil
	}

	returningScanner := func(rows dbutil.Scanner) error {
		var path string
		var symbol string
		var id int
		if err := rows.Scan(&path, &symbol, &id); err != nil {
			return err
		}
		symbolCache.set(path, symbol, id)
		return nil
	}

	return batch.WithInserterWithReturn(
		ctx,
		tx,
		"rockskip_symbols",
		batch.MaxNumPostgresParameters,
		[]string{"added", "deleted", "repo_id", "path", "name"},
		"",
		[]string{"path", "name", "id"},
		returningScanner,
		callback,
	)
}

type repoCommit struct {
	repo   string
	commit string
}

type indexRequest struct {
	repoCommit
	done chan struct{}
}

type symbolIdCache struct {
	cache *lru.Cache
}

func newSymbolIdCache(size int) *symbolIdCache {
	return &symbolIdCache{cache: lru.New(size)}
}

func (s *symbolIdCache) get(path, symbol string) (int, bool) {
	v, ok := s.cache.Get(symbolIdCacheKey(path, symbol))
	if !ok {
		return 0, false
	}
	return v.(int), true
}

func (s *symbolIdCache) set(path, symbol string, id int) {
	s.cache.Add(symbolIdCacheKey(path, symbol), id)
}

func symbolIdCacheKey(path, symbol string) string {
	return path + ":" + symbol
}

type pathSymbolsCache struct {
	cache *lru.Cache
}

func newPathSymbolsCache(size int) *pathSymbolsCache {
	return &pathSymbolsCache{cache: lru.New(size)}
}

func (s *pathSymbolsCache) get(path string) (map[string]struct{}, bool) {
	v, ok := s.cache.Get(path)
	if !ok {
		return nil, false
	}
	return v.(map[string]struct{}), true
}

func (s *pathSymbolsCache) set(path string, symbols map[string]struct{}) {
	s.cache.Add(path, symbols)
}
