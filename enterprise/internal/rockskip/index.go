package rockskip

import (
	"context"
	"database/sql"
	"fmt"

	"k8s.io/utils/lru"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Server) Index(ctx context.Context, conn *sql.Conn, repo, givenCommit string, parse ParseSymbolsFunc) (err error) {
	threadStatus := s.status.NewThreadStatus(fmt.Sprintf("indexing %s@%s", repo, givenCommit))
	defer threadStatus.End()

	tasklog := threadStatus.Tasklog

	// Acquire the indexing lock on the repo.
	releaseLock, err := iLock(ctx, conn, threadStatus, repo)
	if err != nil {
		return err
	}
	defer func() { err = combineErrors(err, releaseLock()) }()

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
	err = s.git.RevListEach(repo, givenCommit, func(commitHash string) (shouldContinue bool, err error) {
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

	symbolCache := newSymbolIdCache(s.symbolsCacheSize)
	pathSymbolsCache := newPathSymbolsCache(s.pathSymbolsCacheSize)

	tasklog.Start("Log")
	entriesIndexed := 0
	err = s.git.LogReverseEach(repo, givenCommit, missingCount, func(entry LogEntry) error {
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
		err = AppendHop(ctx, tx, repoId, hops[0:r], AddedAD, commit)
		if err != nil {
			return errors.Wrap(err, "AppendHop (added)")
		}
		tasklog.Start("AppendHop-")
		err = AppendHop(ctx, tx, repoId, hops[0:r], DeletedAD, commit)
		if err != nil {
			return errors.Wrap(err, "AppendHop (deleted)")
		}

		deletedPaths := []string{}
		addedPaths := []string{}
		for _, pathStatus := range entry.PathStatuses {
			if pathStatus.Status == DeletedAMD || pathStatus.Status == ModifiedAMD {
				deletedPaths = append(deletedPaths, pathStatus.Path)
			}
			if pathStatus.Status == AddedAMD || pathStatus.Status == ModifiedAMD {
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
			case DeletedAMD:
				deletedSymbols[pathStatus.Path] = symbolsFromDeletedFiles[pathStatus.Path]
			case AddedAMD:
				addedSymbols[pathStatus.Path] = symbolsFromAddedFiles[pathStatus.Path]
			case ModifiedAMD:
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
					found := false
					for _, hop := range hops {
						tasklog.Start("GetSymbol")
						id, found, err = GetSymbol(ctx, tx, repoId, path, symbol, hop)
						if err != nil {
							return err
						}
						if found {
							break
						}
					}
					if !found {
						return errors.Newf("could not find id for path %s symbol %s", path, symbol)
					}
				}

				tasklog.Start("UpdateSymbolHops")
				err = UpdateSymbolHops(ctx, tx, id, DeletedAD, commit)
				if err != nil {
					return errors.Wrap(err, "UpdateSymbolHops")
				}
			}
		}

		for path, symbols := range addedSymbols {
			for symbol := range symbols {
				tasklog.Start("InsertSymbol")
				id, err := InsertSymbol(ctx, tx, commit, repoId, path, symbol)
				if err != nil {
					return errors.Wrap(err, "InsertSymbol")
				}
				symbolCache.set(path, symbol, id)
			}
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
