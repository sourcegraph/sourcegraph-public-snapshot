package rockskip

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"
	"unicode/utf8"

	"github.com/amit7itz/goset"
	pg "github.com/lib/pq"
	"k8s.io/utils/lru"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Service) Index(ctx context.Context, repo, givenCommit string) (err error) {
	logger := s.logger.With(log.String("repo", repo), log.String("commit", givenCommit))
	logger.Info("indexing repository with rockskip")

	s.metrics.indexRunning.Inc()
	threadStatus := s.status.NewThreadStatus(fmt.Sprintf("indexing %s@%s", repo, givenCommit))
	defer func(start time.Time) {
		threadStatus.End()
		s.metrics.indexRunning.Dec()
		if err != nil {
			s.metrics.indexFailed.Inc()
		}
		s.metrics.indexDuration.Observe(time.Since(start).Seconds())
	}(time.Now())

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
	tipHeight := 0

	var repoId int
	err = conn.QueryRowContext(ctx, "SELECT id FROM rockskip_repos WHERE repo = $1", repo).Scan(&repoId)
	if err != nil {
		return errors.Wrapf(err, "failed to get repo id for %s", repo)
	}

	missingCommits := []api.CommitID{}
	tasklog.Start("RevList")
	err = s.git.RevList(ctx, repo, givenCommit, func(commitHash string) (shouldContinue bool, err error) {
		defer tasklog.Continue("RevList")

		tasklog.Start("GetCommitByHash")
		commit, height, present, err := GetCommitByHash(ctx, conn, repoId, commitHash)
		if err != nil {
			return false, err
		} else if present {
			tipCommit = commit
			tipHeight = height
			return false, nil
		}
		missingCommits = append(missingCommits, api.CommitID(commitHash))
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "RevList")
	}

	threadStatus.SetProgress(0, len(missingCommits))

	if len(missingCommits) == 0 {
		logger.Info("no commits missing, indexing complete")
		return nil
	}

	logger.Info("rockskip indexing missing commits", log.Int("missingCount", len(missingCommits)))

	symbolCache := lru.New(s.symbolsCacheSize)
	pathSymbolsCache := lru.New(s.pathSymbolsCacheSize)

	tasklog.Start("Log")
	entriesIndexed := 0
	for i := len(missingCommits) - 1; i >= 0; i-- {
		missingCommit := missingCommits[i]
		err := func() error {
			defer tasklog.Continue("Log")

			it, err := s.git.ChangedFiles(ctx, api.RepoName(repo), "", string(missingCommit))
			if err != nil {
				return errors.Wrap(err, "git.ChangedFiles")
			}
			pathStatuses := []gitdomain.PathStatus{}
			for {
				ps, err := it.Next()
				if err != nil {
					it.Close()
					if errors.Is(err, io.EOF) {
						break
					}
					return errors.Wrap(err, "git.ChangedFilesIterator")
				}
				pathStatuses = append(pathStatuses, ps)
			}

			threadStatus.SetProgress(entriesIndexed, len(missingCommits))
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
			commit, err := InsertCommit(ctx, tx, repoId, string(missingCommit), tipHeight+1, hops[r])
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
			for _, pathStatus := range pathStatuses {
				if !utf8.ValidString(pathStatus.Path) {
					s.logger.Warn(
						"Rockskip skipping file due to path not being utf-8 encoded",
						log.String("repo", repo),
						log.String("path", pathStatus.Path),
					)
					continue
				}

				if pathStatus.Status == gitdomain.StatusDeleted || pathStatus.Status == gitdomain.StatusModified {
					deletedPaths = append(deletedPaths, pathStatus.Path)
				}
				if pathStatus.Status == gitdomain.StatusAdded || pathStatus.Status == gitdomain.StatusModified {
					addedPaths = append(addedPaths, pathStatus.Path)
				}
				// Ignoring StatusTypeChanged because there are no changes to the contents of a file.
			}

			symbolsFromDeletedFiles := map[string]*goset.Set[string]{}
			{
				// Fill from the cache.
				for _, path := range deletedPaths {
					if symbols, ok := pathSymbolsCache.Get(path); ok {
						symbolsFromDeletedFiles[path] = symbols.(*goset.Set[string])
					}
				}

				// Fetch the rest from the DB.
				pathsToFetch := goset.NewSet[string]()
				for _, path := range deletedPaths {
					if _, ok := pathSymbolsCache.Get(path); !ok {
						pathsToFetch.Add(path)
					}
				}

				pathToSymbols, err := GetSymbolsInFiles(ctx, tx, repoId, pathsToFetch.Items(), hops)
				if err != nil {
					return err
				}

				for path, symbols := range pathToSymbols {
					symbolsFromDeletedFiles[path] = symbols
				}
			}

			symbolsFromAddedFiles := map[string]*goset.Set[string]{}
			{
				tasklog.Start("ArchiveEach")
				err = archiveEach(ctx, s.fetcher, repo, string(missingCommit), addedPaths, func(path string, contents []byte) error {
					defer tasklog.Continue("ArchiveEach")

					tasklog.Start("parse")
					symbols, err := s.parseSymbols(ctx, path, contents)
					if err != nil {
						return errors.Wrap(err, "parse")
					}

					symbolsFromAddedFiles[path] = goset.NewSet[string]()
					for _, symbol := range symbols {
						symbolsFromAddedFiles[path].Add(symbol.Name)
					}

					// Cache the symbols we just parsed.
					pathSymbolsCache.Add(path, symbolsFromAddedFiles[path])

					return nil
				})

				if err != nil {
					return errors.Wrap(err, "while looping ArchiveEach")
				}

			}

			// Compute the symmetric difference of symbols between the added and deleted paths.
			deletedSymbols := map[string]*goset.Set[string]{}
			addedSymbols := map[string]*goset.Set[string]{}
			for _, pathStatus := range pathStatuses {
				deleted := symbolsFromDeletedFiles[pathStatus.Path]
				if deleted == nil {
					deleted = goset.NewSet[string]()
				}
				added := symbolsFromAddedFiles[pathStatus.Path]
				if added == nil {
					added = goset.NewSet[string]()
				}
				switch pathStatus.Status {
				case gitdomain.StatusDeleted:
					deletedSymbols[pathStatus.Path] = deleted
				case gitdomain.StatusAdded:
					addedSymbols[pathStatus.Path] = added
				case gitdomain.StatusModified:
					deletedSymbols[pathStatus.Path] = deleted.Difference(added)
					addedSymbols[pathStatus.Path] = added.Difference(deleted)
				case gitdomain.StatusTypeChanged:
					// a type change does not change the contents of a file,
					// so this is safe to ignore.
				}
			}

			for path, symbols := range deletedSymbols {
				for _, symbol := range symbols.Items() {
					id := 0
					id_, ok := symbolCache.Get(pathSymbol{path: path, symbol: symbol})
					if ok {
						id = id_.(int)
					} else {
						tasklog.Start("GetSymbol")
						found := false
						id, found, err = GetSymbol(ctx, tx, repoId, path, symbol, hops)
						if err != nil {
							return errors.Wrap(err, "GetSymbol")
						}
						if !found {
							// We did not find the symbol that (supposedly) has been deleted, so ignore the
							// deletion. This will probably lead to extra symbols in search results.
							//
							// The last time this happened, it was caused by impurity in ctags where the
							// result of parsing a file was affected by previously parsed files and not fully
							// determined by the file itself:
							//
							// https://github.com/universal-ctags/ctags/pull/3300
							s.logger.Error(
								"could not find symbol that was supposedly deleted",
								log.String("repo", repo),
								log.Int("commit", commit),
								log.String("path", path),
								log.String("symbol", symbol),
							)
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
			tipHeight += 1

			return nil
		}()
		if err != nil {
			return err
		}
	}

	threadStatus.SetProgress(entriesIndexed, len(missingCommits))

	return nil
}

func BatchInsertSymbols(ctx context.Context, tasklog *TaskLog, tx *sql.Tx, repoId, commit int, symbolCache *lru.Cache, symbols map[string]*goset.Set[string]) error {
	callback := func(inserter *batch.Inserter) error {
		for path, pathSymbols := range symbols {
			for _, symbol := range pathSymbols.Items() {
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
		symbolCache.Add(pathSymbol{path: path, symbol: symbol}, id)
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
	dateAddedToQueue time.Time
	done             chan struct{}
}

type pathSymbol struct {
	path   string
	symbol string
}
