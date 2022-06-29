package rockskip

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/amit7itz/goset"
	pg "github.com/lib/pq"
	"github.com/segmentio/fasthash/fnv1"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CommitId = int

type StatusAD int

const (
	AddedAD   StatusAD = 0
	DeletedAD StatusAD = 1
)

func GetCommitById(ctx context.Context, db dbutil.DB, givenCommit CommitId) (commitHash string, ancestor CommitId, height int, present bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT commit_id, ancestor, height
		FROM rockskip_ancestry
		WHERE id = $1
	`, givenCommit).Scan(&commitHash, &ancestor, &height)
	if err == sql.ErrNoRows {
		return "", 0, 0, false, nil
	} else if err != nil {
		return "", 0, 0, false, errors.Newf("GetCommitById: %s", err)
	}
	return commitHash, ancestor, height, true, nil
}

func GetCommitByHash(ctx context.Context, db dbutil.DB, repoId int, commitHash string) (commit CommitId, height int, present bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT id, height
		FROM rockskip_ancestry
		WHERE repo_id = $1 AND commit_id = $2
	`, repoId, commitHash).Scan(&commit, &height)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	} else if err != nil {
		return 0, 0, false, errors.Newf("GetCommitByHash: %s", err)
	}
	return commit, height, true, nil
}

func InsertCommit(ctx context.Context, db dbutil.DB, repoId int, commitHash string, height int, ancestor CommitId) (id CommitId, err error) {
	err = db.QueryRowContext(ctx, `
		INSERT INTO rockskip_ancestry (commit_id, repo_id, height, ancestor)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, commitHash, repoId, height, ancestor).Scan(&id)
	return id, errors.Wrap(err, "InsertCommit")
}

func GetSymbol(ctx context.Context, db dbutil.DB, repoId int, path string, name string, hops []CommitId) (id int, found bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT id
		FROM rockskip_symbols
		WHERE
			repo_id = $1 AND
			path = $2 AND
			name = $3 AND
		    $4 && added AND
			NOT $4 && deleted
	`, repoId, path, name, pg.Array(hops)).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	} else if err != nil {
		return 0, false, errors.Newf("GetSymbol: %s", err)
	}
	return id, true, nil
}

func GetSymbolsInFiles(ctx context.Context, db dbutil.DB, repoId int, paths []string, hops []CommitId) (map[string]*goset.Set[string], error) {
	pathToSymbols := map[string]*goset.Set[string]{}

	for _, chunk := range chunksOf(paths, 1000) {
		rows, err := db.QueryContext(ctx, `
			SELECT name, path
			FROM rockskip_symbols
			WHERE
				repo_id = $1 AND
				path = ANY($2) AND
				$3 && added AND
				NOT $3 && deleted
		`, repoId, pg.Array(chunk), pg.Array(hops))
		if err != nil {
			return nil, errors.Newf("GetSymbolsInFiles: %s", err)
		}
		for rows.Next() {
			var name string
			var path string
			if err := rows.Scan(&name, &path); err != nil {
				return nil, errors.Newf("GetSymbolsInFiles: %s", err)
			}
			if pathToSymbols[path] == nil {
				pathToSymbols[path] = goset.NewSet[string]()
			}
			pathToSymbols[path].Add(name)
		}
		err = rows.Close()
		if err != nil {
			return nil, errors.Newf("GetSymbolsInFiles: %s", err)
		}
	}

	return pathToSymbols, nil
}

func UpdateSymbolHops(ctx context.Context, db dbutil.DB, id int, status StatusAD, hop CommitId) error {
	column := statusADToColumn(status)
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE rockskip_symbols
		SET %s = array_append(%s, $1)
		WHERE id = $2
	`, column, column), hop, id)
	return errors.Wrap(err, "UpdateSymbolHops")
}

func InsertSymbol(ctx context.Context, db dbutil.DB, hop CommitId, repoId int, path string, name string) (id int, err error) {
	err = db.QueryRowContext(ctx, `
		INSERT INTO rockskip_symbols (added, deleted, repo_id, path, name)
		                      VALUES ($1   , $2     , $3     , $4  , $5  )
		RETURNING id
	`, pg.Array([]int{hop}), pg.Array([]int{}), repoId, path, name).Scan(&id)
	return id, errors.Wrap(err, "InsertSymbol")
}

func AppendHop(ctx context.Context, db dbutil.DB, repoId int, hops []CommitId, positive, negative StatusAD, newHop CommitId) error {
	pos := statusADToColumn(positive)
	neg := statusADToColumn(negative)
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE rockskip_symbols
		SET %s = array_append(%s, $1)
		WHERE $2 && singleton_integer(repo_id) AND $3 && %s AND NOT $3 && %s
	`, pos, pos, pos, neg), newHop, pg.Array([]int{repoId}), pg.Array(hops))
	return errors.Wrap(err, "AppendHop")
}

func DeleteRedundant(ctx context.Context, db dbutil.DB, hop CommitId) error {
	_, err := db.ExecContext(ctx, `
		UPDATE rockskip_symbols
		SET added = array_remove(added, $1), deleted = array_remove(deleted, $1)
		WHERE $2 && added AND $2 && deleted
	`, hop, pg.Array([]int{hop}))
	return errors.Wrap(err, "DeleteRedundant")
}

func tryDeleteOldestRepo(ctx context.Context, db *sql.Conn, maxRepos int, threadStatus *ThreadStatus) (more bool, err error) {
	defer threadStatus.Tasklog.Continue("idle")

	// Select a candidate repo to delete.
	threadStatus.Tasklog.Start("select repo to delete")
	var repoId int
	var repo string
	var repoRank int
	err = db.QueryRowContext(ctx, `
		SELECT id, repo, repo_rank
		FROM (
			SELECT *, RANK() OVER (ORDER BY last_accessed_at DESC) repo_rank
			FROM rockskip_repos
		) sub
		WHERE repo_rank > $1
		ORDER BY last_accessed_at ASC
		LIMIT 1;`, maxRepos,
	).Scan(&repoId, &repo, &repoRank)
	if err == sql.ErrNoRows {
		// No more repos to delete.
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "selecting repo to delete")
	}

	// Note: a search request or deletion could have intervened here.

	// Acquire the write lock on the repo.
	releaseWLock, err := wLock(ctx, db, threadStatus, repo)
	defer func() { err = errors.CombineErrors(err, releaseWLock()) }()
	if err != nil {
		return false, errors.Wrap(err, "acquiring write lock on repo")
	}

	// Make sure the repo is still old. See note above.
	var rank int
	threadStatus.Tasklog.Start("recheck repo rank")
	err = db.QueryRowContext(ctx, `
		SELECT repo_rank
		FROM (
			SELECT id, RANK() OVER (ORDER BY last_accessed_at DESC) repo_rank
			FROM rockskip_repos
		) sub
		WHERE id = $1;`, repoId,
	).Scan(&rank)
	if err == sql.ErrNoRows {
		// The repo was deleted in the meantime, so retry.
		return true, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "selecting repo rank")
	}
	if rank <= maxRepos {
		// An intervening search request must have refreshed the repo, so retry.
		return true, nil
	}

	// Acquire the indexing lock on the repo.
	releaseILock, err := iLock(ctx, db, threadStatus, repo)
	defer func() { err = errors.CombineErrors(err, releaseILock()) }()
	if err != nil {
		return false, errors.Wrap(err, "acquiring indexing lock on repo")
	}

	// Delete the repo.
	threadStatus.Tasklog.Start("delete repo")
	tx, err := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_ancestry WHERE repo_id = $1;", repoId)
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_symbols WHERE repo_id = $1;", pg.Array([]int{repoId}))
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_repos WHERE id = $1;", repoId)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

func PrintInternals(ctx context.Context, db dbutil.DB) error {
	fmt.Println("Commit ancestry:")
	fmt.Println()

	// print all rows in the rockskip_ancestry table
	rows, err := db.QueryContext(ctx, `
		SELECT a1.commit_id, a1.height, a2.commit_id
		FROM rockskip_ancestry a1
		JOIN rockskip_ancestry a2 ON a1.ancestor = a2.id
		ORDER BY height ASC
	`)
	if err != nil {
		return errors.Wrap(err, "PrintInternals")
	}
	defer rows.Close()

	for rows.Next() {
		var commit, ancestor string
		var height int
		err = rows.Scan(&commit, &height, &ancestor)
		if err != nil {
			return errors.Wrap(err, "PrintInternals: Scan")
		}
		fmt.Printf("height %3d commit %s ancestor %s\n", height, commit, ancestor)
	}

	fmt.Println()
	fmt.Println("Symbols:")
	fmt.Println()

	rows, err = db.QueryContext(ctx, `
		SELECT id, path, name, added, deleted
		FROM rockskip_symbols
		ORDER BY id ASC
	`)
	if err != nil {
		return errors.Wrap(err, "PrintInternals")
	}

	for rows.Next() {
		var id int
		var path string
		var name string
		var added, deleted []int64
		err = rows.Scan(&id, &path, &name, pg.Array(&added), pg.Array(&deleted))
		if err != nil {
			return errors.Wrap(err, "PrintInternals: Scan")
		}
		fmt.Printf("  id %d path %-10s symbol %s\n", id, path, name)
		for _, a := range added {
			hash, _, _, _, err := GetCommitById(ctx, db, int(a))
			if err != nil {
				return err
			}
			fmt.Printf("    + %-40s\n", hash)
		}
		fmt.Println()
		for _, d := range deleted {
			hash, _, _, _, err := GetCommitById(ctx, db, int(d))
			if err != nil {
				return err
			}
			fmt.Printf("    - %-40s\n", hash)
		}
		fmt.Println()

	}

	fmt.Println()
	return nil
}

func updateLastAccessedAt(ctx context.Context, db dbutil.DB, repo string) (id int, err error) {
	err = db.QueryRowContext(ctx, `
			INSERT INTO rockskip_repos (repo, last_accessed_at)
			VALUES ($1, now())
			ON CONFLICT (repo)
			DO UPDATE SET last_accessed_at = now()
			RETURNING id
		`, repo).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func statusADToColumn(status StatusAD) string {
	switch status {
	case AddedAD:
		return "added"
	case DeletedAD:
		return "deleted"
	default:
		fmt.Println("unexpected status StatusAD: ", status)
		return "unknown_status"
	}
}

var RW_LOCKS_NAMESPACE = int32(fnv1.HashString32("symbols-rw"))
var INDEXING_LOCKS_NAMESPACE = int32(fnv1.HashString32("symbols-indexing"))

func lock(ctx context.Context, db dbutil.DB, threadStatus *ThreadStatus, namespace int32, name, repo, lockFn, unlockFn string) (func() error, error) {
	key := int32(fnv1.HashString32(repo))

	threadStatus.Tasklog.Start(name)
	_, err := db.ExecContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, lockFn), namespace, key)
	if err != nil {
		return nil, errors.Newf("acquire %s: %s", name, err)
	}
	threadStatus.HoldLock(name)

	release := func() error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, unlockFn), namespace, key)
		if err != nil {
			return errors.Newf("release %s: %s", name, err)
		}
		threadStatus.ReleaseLock(name)
		return nil
	}

	return release, nil
}

func tryLock(ctx context.Context, db dbutil.DB, threadStatus *ThreadStatus, namespace int32, name, repo, lockFn, unlockFn string) (bool, func() error, error) {
	key := int32(fnv1.HashString32(repo))

	threadStatus.Tasklog.Start(name)
	locked, _, err := basestore.ScanFirstBool(db.QueryContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, lockFn), namespace, key))
	if err != nil {
		return false, nil, errors.Newf("try acquire %s: %s", name, err)
	}

	if !locked {
		return false, nil, nil
	}

	threadStatus.HoldLock(name)

	release := func() error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, unlockFn), namespace, key)
		if err != nil {
			return errors.Newf("release %s: %s", name, err)
		}
		threadStatus.ReleaseLock(name)
		return nil
	}

	return true, release, nil
}

// tryRLock attempts to acquire a read lock on the repo.
func tryRLock(ctx context.Context, db dbutil.DB, threadStatus *ThreadStatus, repo string) (bool, func() error, error) {
	return tryLock(ctx, db, threadStatus, RW_LOCKS_NAMESPACE, "rLock", repo, "pg_try_advisory_lock_shared", "pg_advisory_unlock_shared")
}

// wLock acquires the write lock on the repo. It blocks only when another connection holds a read or the
// write lock. That means a single connection can acquire the write lock while holding a read lock.
func wLock(ctx context.Context, db dbutil.DB, threadStatus *ThreadStatus, repo string) (func() error, error) {
	return lock(ctx, db, threadStatus, RW_LOCKS_NAMESPACE, "wLock", repo, "pg_advisory_lock", "pg_advisory_unlock")
}

// iLock acquires the indexing lock on the repo.
func iLock(ctx context.Context, db dbutil.DB, threadStatus *ThreadStatus, repo string) (func() error, error) {
	return lock(ctx, db, threadStatus, INDEXING_LOCKS_NAMESPACE, "iLock", repo, "pg_advisory_lock", "pg_advisory_unlock")
}

func chunksOf[T any](xs []T, size int) [][]T {
	if xs == nil {
		return nil
	}

	chunks := [][]T{}

	for i := 0; i < len(xs); i += size {
		end := i + size

		if end > len(xs) {
			end = len(xs)
		}

		chunks = append(chunks, xs[i:end])
	}

	return chunks
}
