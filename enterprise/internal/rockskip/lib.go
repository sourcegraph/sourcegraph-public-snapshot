package rockskip

import (
	"bufio"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/keegancsmith/sqlf"
	pg "github.com/lib/pq"
	"github.com/segmentio/fasthash/fnv1"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
)

type Git interface {
	LogReverseEach(repo string, commit string, n int, onLogEntry func(logEntry LogEntry) error) error
	RevListEach(repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error
	ArchiveEach(repo string, commit string, paths []string, onFile func(path string, contents []byte) error) error
}

type Symbol struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
	Kind   string `json:"kind"`
	Line   int    `json:"line"`
}

type Symbols []Symbol

func (symbols Symbols) Value() (driver.Value, error) {
	return json.Marshal(symbols)
}

func (symbols *Symbols) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("scan symbol: expected []byte")
	}

	return json.Unmarshal(bytes, &symbols)
}

type ParseSymbolsFunc func(path string, bytes []byte) (symbols []Symbol, err error)

type LogEntry struct {
	Commit       string
	PathStatuses []PathStatus
}

type PathStatus struct {
	Path   string
	Status StatusAMD
}

type CommitStatus struct {
	Commit string
	Status StatusAMD
}

type Blob struct {
	Commit  string
	Path    string
	Added   []string
	Deleted []string
	Symbols []Symbol
}

type StatusAMD int

const (
	AddedAMD    StatusAMD = 0
	ModifiedAMD StatusAMD = 1
	DeletedAMD  StatusAMD = 2
)

type StatusAD int

const (
	AddedAD   StatusAD = 0
	DeletedAD StatusAD = 1
)

const NULL = "0000000000000000000000000000000000000000"

type RequestStatus struct {
	Tasklog   *TaskLog
	Repo      string
	Commit    string
	HeldLocks map[string]struct{}
	Indexed   int
	Total     int
	mu        sync.Mutex
	onEnd     func()
}

func NewRequestStatus(repo string, commit string, onEnd func()) *RequestStatus {
	return &RequestStatus{
		Tasklog:   NewTaskLog(),
		Repo:      repo,
		Commit:    commit,
		HeldLocks: map[string]struct{}{},
		Indexed:   -1,
		Total:     -1,
		mu:        sync.Mutex{},
		onEnd:     onEnd,
	}
}

func (s *RequestStatus) WithLock(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f()
}

func (s *RequestStatus) SetProgress(indexed, total int) {
	s.WithLock(func() { s.Indexed = indexed; s.Total = total })
}
func (s *RequestStatus) HoldLock(name string)    { s.WithLock(func() { s.HeldLocks[name] = struct{}{} }) }
func (s *RequestStatus) ReleaseLock(name string) { s.WithLock(func() { delete(s.HeldLocks, name) }) }

func (s *RequestStatus) End() {
	if s.onEnd != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.onEnd()
	}
}

type TaskLog struct {
	currentName  string
	currentStart time.Time
	nameToTask   map[string]*Task
	// This mutex is only necessary to synchronize with the status page handler.
	mu sync.Mutex
}

type Task struct {
	Duration time.Duration
	Count    int
}

func NewTaskLog() *TaskLog {
	return &TaskLog{
		currentName:  "idle",
		currentStart: time.Now(),
		nameToTask:   map[string]*Task{"idle": {Duration: 0, Count: 1}},
		mu:           sync.Mutex{},
	}
}

func (t *TaskLog) Start(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	if _, ok := t.nameToTask[t.currentName]; !ok {
		t.nameToTask[t.currentName] = &Task{Duration: 0, Count: 0}
	}
	t.nameToTask[t.currentName].Duration += now.Sub(t.currentStart)

	if _, ok := t.nameToTask[name]; !ok {
		t.nameToTask[name] = &Task{Duration: 0, Count: 0}
	}
	t.nameToTask[name].Count += 1

	t.currentName = name
	t.currentStart = now
}

func (t *TaskLog) Continue(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	if _, ok := t.nameToTask[t.currentName]; !ok {
		t.nameToTask[t.currentName] = &Task{Duration: 0, Count: 0}
	}
	t.nameToTask[t.currentName].Duration += now.Sub(t.currentStart)

	if _, ok := t.nameToTask[name]; !ok {
		t.nameToTask[name] = &Task{Duration: 0, Count: 0}
	}

	t.currentName = name
	t.currentStart = now
}

func (t *TaskLog) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.currentName = "idle"
	t.currentStart = time.Now()
	t.nameToTask = map[string]*Task{"idle": {Duration: 0, Count: 1}}
}

func (t *TaskLog) Print() {
	fmt.Println(t)
}

func (t *TaskLog) String() string {
	var s strings.Builder

	t.Continue(t.currentName)

	t.mu.Lock()
	defer t.mu.Unlock()

	var total time.Duration = 0
	totalCount := 0
	for _, task := range t.nameToTask {
		total += task.Duration
		totalCount += task.Count
	}
	fmt.Fprintf(&s, "Tasks (%.0fs total, current %s): ", total.Seconds(), t.currentName)

	type kv struct {
		Key   string
		Value *Task
	}

	var kvs []kv
	for k, v := range t.nameToTask {
		kvs = append(kvs, kv{k, v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value.Duration > kvs[j].Value.Duration
	})

	for _, kv := range kvs {
		fmt.Fprintf(&s, "%s %d%% %dx, ", kv.Key, kv.Value.Duration*100/total, kv.Value.Count)
	}

	return s.String()
}

func Index(ctx context.Context, git Git, db *sql.Conn, requestStatus *RequestStatus, parse ParseSymbolsFunc, repo, givenCommit string, maxRepos int, indexingSemaphore *semaphore.Weighted) (err error) {
	// Acquire the indexing lock on the repo.
	releaseLock, err := iLock(ctx, db, requestStatus, repo)
	if err != nil {
		return err
	}
	defer func() { err = combineErrors(err, releaseLock()) }()

	tasklog := requestStatus.Tasklog

	tipCommit := NULL
	tipHeight := 0

	missingCount := 0
	tasklog.Start("RevList")
	err = git.RevListEach(repo, givenCommit, func(commit string) (shouldContinue bool, err error) {
		defer tasklog.Continue("RevList")

		tasklog.Start("GetCommit")
		_, height, present, err := GetCommit(ctx, db, repo, commit)
		if err != nil {
			return false, errors.Wrap(err, "GetCommit")
		} else if present {
			tipCommit = commit
			tipHeight = height
			return false, nil
		}
		missingCount += 1
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "RevList")
	}

	requestStatus.SetProgress(0, missingCount)

	if missingCount == 0 {
		return nil
	}

	tasklog.Start("Acquire indexing semaphore")
	err = indexingSemaphore.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	requestStatus.HoldLock("indexing semaphore")
	defer func() {
		indexingSemaphore.Release(1)
		requestStatus.ReleaseLock("indexing semaphore")
	}()

	pathToBlobIdCache := map[string]int{}

	tasklog.Start("Log")
	entriesIndexed := 0
	err = git.LogReverseEach(repo, givenCommit, missingCount, func(entry LogEntry) error {
		defer tasklog.Continue("Log")

		requestStatus.SetProgress(entriesIndexed, missingCount)
		entriesIndexed++

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "begin transaction")
		}
		defer tx.Rollback()

		hops, err := getHops(ctx, tx, repo, tipCommit, tasklog)
		if err != nil {
			return errors.Wrap(err, "getHops")
		}

		r := ruler(tipHeight + 1)
		if r >= len(hops) {
			return fmt.Errorf("ruler(%d) = %d is out of range of len(hops) = %d", tipHeight+1, r, len(hops))
		}

		tasklog.Start("InsertCommit")
		err = InsertCommit(ctx, tx, repo, entry.Commit, tipHeight+1, hops[r])
		if err != nil {
			return errors.Wrap(err, "InsertCommit")
		}

		tasklog.Start("AppendHop+")
		err = AppendHop(ctx, tx, hops[0:r], AddedAD, entry.Commit)
		if err != nil {
			return errors.Wrap(err, "AppendHop (added)")
		}
		tasklog.Start("AppendHop-")
		err = AppendHop(ctx, tx, hops[0:r], DeletedAD, entry.Commit)
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

		for _, deletedPath := range deletedPaths {
			id := 0
			ok := false
			if id, ok = pathToBlobIdCache[deletedPath]; !ok {
				found := false
				for _, hop := range hops {
					tasklog.Start("GetBlob")
					id, found, err = GetBlob(ctx, tx, hop, deletedPath)
					if err != nil {
						return errors.Wrap(err, "GetBlob")
					}
					if found {
						break
					}
				}
				if !found {
					return fmt.Errorf("could not find blob for path %s", deletedPath)
				}
			}

			tasklog.Start("UpdateBlobHops")
			err = UpdateBlobHops(ctx, tx, id, DeletedAD, entry.Commit)
			if err != nil {
				return errors.Wrap(err, "UpdateBlobHops")
			}
		}

		tasklog.Start("ArchiveEach")
		err = git.ArchiveEach(repo, entry.Commit, addedPaths, func(addedPath string, contents []byte) error {
			defer tasklog.Continue("ArchiveEach")

			tasklog.Start("parse")
			symbols, err := parse(addedPath, contents)
			if err != nil {
				return errors.Wrap(err, "parse")
			}
			blob := Blob{
				Commit:  entry.Commit,
				Path:    addedPath,
				Added:   []string{entry.Commit},
				Deleted: []string{},
				Symbols: symbols,
			}
			tasklog.Start("InsertBlob")
			id, err := InsertBlob(ctx, tx, blob, repo)
			if err != nil {
				return errors.Wrap(err, "InsertBlob")
			}
			pathToBlobIdCache[addedPath] = id
			return nil
		})
		if err != nil {
			return errors.Wrap(err, "while looping ArchiveEach")
		}

		tasklog.Start("DeleteRedundant")
		err = DeleteRedundant(ctx, tx, entry.Commit)
		if err != nil {
			return errors.Wrap(err, "DeleteRedundant")
		}

		tasklog.Start("CommitTx")
		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "commit transaction")
		}

		tipCommit = entry.Commit
		tipHeight += 1

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "LogReverseEach")
	}

	requestStatus.SetProgress(entriesIndexed, missingCount)

	return nil
}

func getHops(ctx context.Context, tx Queryable, repo, commit string, tasklog *TaskLog) ([]string, error) {
	tasklog.Start("get hops")

	current := commit
	spine := []string{current}

	for {
		ancestor, _, present, err := GetCommit(ctx, tx, repo, current)
		if err != nil {
			return nil, errors.Wrap(err, "GetCommit")
		} else if !present {
			break
		} else {
			if current == NULL {
				break
			}
			current = ancestor
			spine = append(spine, current)
		}
	}

	return spine, nil
}

var RW_LOCKS_NAMESPACE = int32(fnv1.HashString32("symbols-rw"))
var INDEXING_LOCKS_NAMESPACE = int32(fnv1.HashString32("symbols-indexing"))

func DeleteOldRepos(ctx context.Context, db *sql.Conn, maxRepos int, requestStatus *RequestStatus) error {
	// Keep deleting repos until we're back to at most maxRepos.
	for {
		more, err := tryDeleteOldestRepo(ctx, db, maxRepos, requestStatus)
		if err != nil {
			return err
		}
		if !more {
			return nil
		}
	}
}

func tryDeleteOldestRepo(ctx context.Context, db *sql.Conn, maxRepos int, requestStatus *RequestStatus) (more bool, err error) {
	defer requestStatus.Tasklog.Continue("idle")

	// Select a candidate repo to delete.
	requestStatus.Tasklog.Start("select repo to delete")
	var repo string
	var repoRank int
	err = db.QueryRowContext(ctx, `
		SELECT repo, repo_rank
		FROM (
			SELECT *, RANK() OVER (ORDER BY last_accessed_at DESC) repo_rank
			FROM rockskip_repos
		) sub
		WHERE repo_rank > $1
		ORDER BY last_accessed_at ASC
		LIMIT 1;`, maxRepos,
	).Scan(&repo, &repoRank)
	if err == sql.ErrNoRows {
		// No more repos to delete.
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "selecting repo to delete")
	}

	// Note: a search request or deletion could have intervened here.

	// Acquire the write lock on the repo.
	releaseWLock, err := wLock(ctx, db, requestStatus, repo)
	defer func() { err = combineErrors(err, releaseWLock()) }()
	if err != nil {
		return false, errors.Wrap(err, "acquiring write lock on repo")
	}

	// Make sure the repo is still old. See note above.
	var rank int
	requestStatus.Tasklog.Start("recheck repo rank")
	err = db.QueryRowContext(ctx, `
		SELECT repo_rank
		FROM (
			SELECT repo, RANK() OVER (ORDER BY last_accessed_at DESC) repo_rank
			FROM rockskip_repos
		) sub
		WHERE repo = $1;`, repo,
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
	releaseILock, err := iLock(ctx, db, requestStatus, repo)
	defer func() { err = combineErrors(err, releaseILock()) }()
	if err != nil {
		return false, errors.Wrap(err, "acquiring indexing lock on repo")
	}

	// Delete the repo.
	requestStatus.Tasklog.Start("delete repo")
	tx, err := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_ancestry WHERE repo = $1;", repo)
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_blobs WHERE repo = $1;", repo)
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_repos WHERE repo = $1;", repo)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

func updateLastAccessedAt(ctx context.Context, conn *sql.Conn, repo string, updateChan chan<- struct{}) error {
	_, err := conn.ExecContext(ctx, `
			INSERT INTO rockskip_repos (repo, last_accessed_at)
			VALUES ($1, now())
			ON CONFLICT (repo)
			DO UPDATE SET last_accessed_at = now()
		`, repo)
	if err != nil {
		return err
	}

	// Non-blocking send on updateChan to notify the background deletion goroutine.
	select {
	case updateChan <- struct{}{}:
	default:
	}

	return nil
}

// Ruler sequence
//
// input : 0, 1, 2, 3, 4, 5, 6, 7, 8, ...
// output: 0, 0, 1, 0, 2, 0, 1, 0, 3, ...
//
// https://oeis.org/A007814
func ruler(n int) int {
	if n == 0 {
		return 0
	}
	if n%2 != 0 {
		return 0
	}
	return 1 + ruler(n/2)
}

type SubprocessGit struct {
	gitDir        string
	catFileCmd    *exec.Cmd
	catFileStdin  io.WriteCloser
	catFileStdout bufio.Reader
}

func NewSubprocessGit(gitDir string) (*SubprocessGit, error) {
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = gitDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &SubprocessGit{
		gitDir:        gitDir,
		catFileCmd:    cmd,
		catFileStdin:  stdin,
		catFileStdout: *bufio.NewReader(stdout),
	}, nil
}

func (git SubprocessGit) Close() error {
	err := git.catFileStdin.Close()
	if err != nil {
		return err
	}
	return git.catFileCmd.Wait()
}

func (git SubprocessGit) LogReverseEach(repo string, givenCommit string, n int, onLogEntry func(entry LogEntry) error) (returnError error) {
	log := exec.Command("git", LogReverseArgs(n, givenCommit)...)
	log.Dir = git.gitDir
	output, err := log.StdoutPipe()
	if err != nil {
		return err
	}

	err = log.Start()
	if err != nil {
		return err
	}
	defer func() {
		err = log.Wait()
		if err != nil {
			returnError = err
		}
	}()

	return ParseLogReverseEach(output, onLogEntry)
}

func (git SubprocessGit) RevListEach(repo string, givenCommit string, onCommit func(commit string) (shouldContinue bool, err error)) (returnError error) {
	revList := exec.Command("git", RevListArgs(givenCommit)...)
	revList.Dir = git.gitDir
	output, err := revList.StdoutPipe()
	if err != nil {
		return err
	}

	err = revList.Start()
	if err != nil {
		return err
	}
	defer func() {
		err = revList.Wait()
		if err != nil {
			returnError = err
		}
	}()

	return RevListEach(output, onCommit)
}

func (git SubprocessGit) ArchiveEach(repo string, commit string, paths []string, onFile func(path string, contents []byte) error) error {
	for _, path := range paths {
		_, err := git.catFileStdin.Write([]byte(fmt.Sprintf("%s:%s\n", commit, path)))
		if err != nil {
			return errors.Wrap(err, "writing to cat-file stdin")
		}

		line, err := git.catFileStdout.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "read newline")
		}
		line = line[:len(line)-1] // Drop the trailing newline
		parts := strings.Split(line, " ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected cat-file output: %q", line)
		}
		size, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return errors.Wrap(err, "parse size")
		}

		fileContents, err := io.ReadAll(io.LimitReader(&git.catFileStdout, size))
		if err != nil {
			return errors.Wrap(err, "read contents")
		}

		discarded, err := git.catFileStdout.Discard(1) // Discard the trailing newline
		if err != nil {
			return errors.Wrap(err, "discard newline")
		}
		if discarded != 1 {
			return fmt.Errorf("expected to discard 1 byte, but discarded %d", discarded)
		}

		err = onFile(path, fileContents)
		if err != nil {
			return errors.Wrap(err, "onFile")
		}
	}

	return nil
}

func GetCommit(ctx context.Context, db Queryable, repo, givenCommit string) (ancestor string, height int, present bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT ancestor_id, height
		FROM rockskip_ancestry
		WHERE repo = $1 AND commit_id = $2
	`, repo, givenCommit).Scan(&ancestor, &height)
	if err == sql.ErrNoRows {
		return "", 0, false, nil
	} else if err != nil {
		return "", 0, false, fmt.Errorf("GetCommit: %s", err)
	}
	return ancestor, height, true, nil
}

func InsertCommit(ctx context.Context, db Queryable, repo, commit string, height int, ancestor string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO rockskip_ancestry (commit_id, repo, height, ancestor_id)
		VALUES ($1, $2, $3, $4)
	`, commit, repo, height, ancestor)
	return errors.Wrap(err, "InsertCommit")
}

func GetBlob(ctx context.Context, db Queryable, hop string, path string) (id int, found bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT id
		FROM rockskip_blobs
		WHERE $1 = ANY (path) AND $2 = ANY (added) AND NOT $2 = ANY (deleted)
	`, path, hop).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	} else if err != nil {
		return 0, false, fmt.Errorf("GetBlob: %s", err)
	}
	return id, true, nil
}

func UpdateBlobHops(ctx context.Context, db Queryable, id int, status StatusAD, hop string) error {
	column := statusADToColumn(status)
	// TODO also try `||` instead of `array_append``
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE rockskip_blobs
		SET %s = array_append(%s, $1)
		WHERE id = $2
	`, column, column), hop, id)
	return errors.Wrap(err, "UpdateBlobHops")
}

func InsertBlob(ctx context.Context, db Queryable, blob Blob, repo string) (id int, err error) {
	symbolNames := []string{}
	for _, symbol := range blob.Symbols {
		symbolNames = append(symbolNames, symbol.Name)
	}

	lastInsertId := 0
	err = db.QueryRowContext(ctx, `
		INSERT INTO rockskip_blobs (repo, commit_id, path, added, deleted, symbol_names, symbol_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, repo, blob.Commit, pg.Array([]string{blob.Path}), pg.Array(blob.Added), pg.Array(blob.Deleted), pg.Array(symbolNames), Symbols(blob.Symbols)).Scan(&lastInsertId)
	return lastInsertId, errors.Wrap(err, "InsertBlob")
}

func AppendHop(ctx context.Context, db Queryable, hops []string, givenStatus StatusAD, newHop string) error {
	column := statusADToColumn(givenStatus)
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE rockskip_blobs
		SET %s = array_append(%s, $1)
		WHERE $2 && %s
	`, column, column, column), newHop, pg.Array(hops))
	return errors.Wrap(err, "AppendHop")
}

func Search(ctx context.Context, args types.SearchArgs, git Git, db *sql.DB, parse ParseSymbolsFunc, maxRepos int, searchSemaphore *semaphore.Weighted, indexingSemaphore *semaphore.Weighted, requestStatus *RequestStatus, updateChan chan<- struct{}) (blobs []Blob, err error) {
	repo := string(args.Repo)
	commit := string(args.CommitID)

	requestStatus.Tasklog.Start("Acquire search semaphore")
	err = searchSemaphore.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	requestStatus.HoldLock("search semaphore")
	defer func() {
		searchSemaphore.Release(1)
		requestStatus.ReleaseLock("search semaphore")
	}()

	// Get a fresh connection from the DB pool to get deterministic "lock stacking" behavior.
	// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = combineErrors(err, conn.Close()) }()

	// Acquire a read lock on the repo.
	releaseRLock, err := rLock(ctx, conn, requestStatus, repo)
	if err != nil {
		return nil, err
	}
	defer func() { err = combineErrors(err, releaseRLock()) }()

	// Insert or set the last_accessed_at column for this repo to now() in the rockskip_repos table.
	requestStatus.Tasklog.Start("update last_accessed_at")
	err = updateLastAccessedAt(ctx, conn, repo, updateChan)
	if err != nil {
		return nil, err
	}

	// Check if the commit has already been indexed, and if not then index it.
	requestStatus.Tasklog.Start("check commit presence")
	_, _, present, err := GetCommit(ctx, conn, repo, commit)
	if err != nil {
		return nil, err
	} else if !present {
		err = Index(ctx, git, conn, requestStatus, parse, repo, commit, maxRepos, indexingSemaphore)
		if err != nil {
			return nil, err
		}
	}

	// Finally search.
	blobs, err = queryBlobs(ctx, conn, args, requestStatus)
	if err != nil {
		return nil, err
	}

	return blobs, nil
}

func queryBlobs(ctx context.Context, conn *sql.Conn, args types.SearchArgs, requestStatus *RequestStatus) ([]Blob, error) {
	hops, err := getHops(ctx, conn, string(args.Repo), string(args.CommitID), requestStatus.Tasklog)
	if err != nil {
		return nil, err
	}

	limit := 100
	if args.First > 0 {
		limit = args.First
	}

	requestStatus.Tasklog.Start("run query")
	q := sqlf.Sprintf(`
		SELECT id, commit_id, path, added, deleted, symbol_data
		FROM rockskip_blobs
		WHERE
			repo = %s
			AND     %s && added
			AND NOT %s && deleted
			AND %s
		LIMIT %s;`,
		string(args.Repo),
		pg.Array(hops),
		pg.Array(hops),
		convertSearchArgsToSqlQuery(args),
		limit,
	)

	var rows *sql.Rows
	rows, err = conn.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, errors.Wrap(err, "Search")
	}
	defer rows.Close()

	symbolCount := 0
	blobs := []Blob{}
	for rows.Next() {
		var id int
		var commit string
		var paths []string
		var added, deleted []string
		var allSymbols Symbols
		err = rows.Scan(&id, &commit, pg.Array(&paths), pg.Array(&added), pg.Array(&deleted), &allSymbols)
		if err != nil {
			return nil, errors.Wrap(err, "Search: Scan")
		}
		if len(paths) != 1 {
			return nil, fmt.Errorf("Search: expected 1 path, got %d", len(paths))
		}
		path := paths[0]

		symbols := []Symbol{}
		for _, symbol := range allSymbols {
			if symbolCount >= limit {
				break
			}
			if args.Query == "" {
				symbols = append(symbols, symbol)
				symbolCount++
			} else {
				regex, err := regexp.Compile(args.Query)
				if err != nil {
					return nil, errors.Wrap(err, "Search compile regex")
				}
				if regex.MatchString(symbol.Name) {
					symbols = append(symbols, symbol)
					symbolCount++
				}
			}
		}

		if len(symbols) > 0 {
			blobs = append(blobs, Blob{Commit: commit, Path: path, Added: added, Deleted: deleted, Symbols: symbols})
		}

		if symbolCount >= limit {
			break
		}
	}

	return blobs, nil
}

func convertSearchArgsToSqlQuery(args types.SearchArgs) *sqlf.Query {
	// TODO support non regexp queries once the frontend supports it.

	conjunctOrNils := []*sqlf.Query{}

	// Query
	conjunctOrNils = append(conjunctOrNils, arrayMatchesRegex("symbol_names", args.Query, args.IsCaseSensitive))

	// IncludePatterns
	for _, includePattern := range args.IncludePatterns {
		conjunctOrNils = append(conjunctOrNils, arrayMatchesRegex("path", includePattern, args.IsCaseSensitive))
	}

	// ExcludePattern
	conjunctOrNils = append(conjunctOrNils, negate(arrayMatchesRegex("path", args.ExcludePattern, args.IsCaseSensitive)))

	// Drop nils
	conjuncts := []*sqlf.Query{}
	for _, condition := range conjunctOrNils {
		if condition != nil {
			conjuncts = append(conjuncts, condition)
		}
	}

	if len(conjuncts) == 0 {
		return sqlf.Sprintf("TRUE")
	}

	return sqlf.Join(conjuncts, "AND")
}

func arrayMatchesRegex(column string, regex string, isCaseSensitive bool) *sqlf.Query {
	if regex == "" {
		return nil
	}

	if literal, isExact, err := isLiteralEquality(regex); err == nil && isExact && isCaseSensitive {
		return sqlf.Sprintf("%s && %s", pg.Array([]string{literal}), column)
	}

	operator := "~"
	if !isCaseSensitive {
		operator = "~*"
	}

	return sqlf.Sprintf(
		fmt.Sprintf(`
			EXISTS (
				SELECT
				FROM   unnest(%s) col
				WHERE  col %s %%s
			)`, column, operator),
		regex,
	)
}

// isLiteralEquality returns true if the given regex matches literal strings exactly.
// If so, this function returns true along with the literal search query. If not, this
// function returns false.
func isLiteralEquality(expr string) (string, bool, error) {
	regexp, err := syntax.Parse(expr, syntax.Perl)
	if err != nil {
		return "", false, errors.Wrap(err, "regexp/syntax.Parse")
	}

	// want a concat of size 3 which is [begin, literal, end]
	if regexp.Op == syntax.OpConcat && len(regexp.Sub) == 3 {
		// starts with ^
		if regexp.Sub[0].Op == syntax.OpBeginLine || regexp.Sub[0].Op == syntax.OpBeginText {
			// is a literal
			if regexp.Sub[1].Op == syntax.OpLiteral {
				// ends with $
				if regexp.Sub[2].Op == syntax.OpEndLine || regexp.Sub[2].Op == syntax.OpEndText {
					return string(regexp.Sub[1].Rune), true, nil
				}
			}
		}
	}

	return "", false, nil
}

func negate(query *sqlf.Query) *sqlf.Query {
	if query == nil {
		return nil
	}

	return sqlf.Sprintf("NOT %s", query)
}

// rLock acquires a read lock on the repo. It blocks only when another connection holds the write lock.
func rLock(ctx context.Context, conn *sql.Conn, requestStatus *RequestStatus, repo string) (func() error, error) {
	requestStatus.Tasklog.Start("rLock")
	_, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock_shared($1, $2)`, RW_LOCKS_NAMESPACE, int32(fnv1.HashString32(repo)))
	if err != nil {
		return nil, errors.Wrap(err, "acquire rLock")
	}
	requestStatus.HoldLock("rLock")

	return func() error {
		_, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock_shared($1, $2)`, RW_LOCKS_NAMESPACE, int32(fnv1.HashString32(repo)))
		if err == nil {
			requestStatus.ReleaseLock("rLock")
		}
		return errors.Wrap(err, "release rLock")
	}, nil
}

// wLock acquires the write lock on the repo. It blocks only when another connection holds a read or the
// write lock. That means a single connection can acquire the write lock while holding a read lock.
func wLock(ctx context.Context, conn *sql.Conn, requestStatus *RequestStatus, repo string) (func() error, error) {
	requestStatus.Tasklog.Start("wLock")
	_, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1, $2)`, RW_LOCKS_NAMESPACE, int32(fnv1.HashString32(repo)))
	if err != nil {
		return nil, errors.Wrap(err, "acquire wLock")
	}
	requestStatus.HoldLock("wLock")

	return func() error {
		_, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock($1, $2)`, RW_LOCKS_NAMESPACE, int32(fnv1.HashString32(repo)))
		if err == nil {
			requestStatus.ReleaseLock("wLock")
		}
		return errors.Wrap(err, "release wLock")
	}, nil
}

// iLock acquires the indexing lock on the repo.
func iLock(ctx context.Context, conn *sql.Conn, requestStatus *RequestStatus, repo string) (func() error, error) {
	requestStatus.Tasklog.Start("iLock")
	_, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1, $2)`, INDEXING_LOCKS_NAMESPACE, int32(fnv1.HashString32(repo)))
	if err != nil {
		return nil, errors.Wrap(err, "acquire iLock")
	}
	requestStatus.HoldLock("iLock")

	return func() error {
		_, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock($1, $2)`, INDEXING_LOCKS_NAMESPACE, int32(fnv1.HashString32(repo)))
		if err == nil {
			requestStatus.ReleaseLock("iLock")
		}
		return errors.Wrap(err, "release iLock")
	}, nil
}

func DeleteRedundant(ctx context.Context, db Queryable, hop string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE rockskip_blobs
		SET added = array_remove(added, $1), deleted = array_remove(deleted, $1)
		WHERE $2 && added AND $2 && deleted
	`, hop, pg.Array([]string{hop}))
	return errors.Wrap(err, "DeleteRedundant")
}

func PrintInternals(ctx context.Context, db Queryable) error {
	fmt.Println("Commit ancestry:")
	fmt.Println()

	// print all rows in the rockskip_ancestry table
	rows, err := db.QueryContext(ctx, `
		SELECT commit_id, height, ancestor_id
		FROM rockskip_ancestry
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
	fmt.Println("Blobs:")
	fmt.Println()

	rows, err = db.QueryContext(ctx, `
		SELECT id, path, added, deleted
		FROM rockskip_blobs
		ORDER BY id ASC
	`)
	if err != nil {
		return errors.Wrap(err, "PrintInternals")
	}

	for rows.Next() {
		var id int
		var paths []string
		var added, deleted []string
		err = rows.Scan(&id, pg.Array(&paths), pg.Array(&added), pg.Array(&deleted))
		if err != nil {
			return errors.Wrap(err, "PrintInternals: Scan")
		}
		if len(paths) != 1 {
			return fmt.Errorf("Search: expected 1 path, got %d", len(paths))
		}
		path := paths[0]
		fmt.Printf("  id %d path %-10s\n", id, path)
		for _, a := range added {
			fmt.Printf("    + %-40s\n", a)
		}
		fmt.Println()
		for _, d := range deleted {
			fmt.Printf("    - %-40s\n", d)
		}
		fmt.Println()

	}

	fmt.Println()
	return nil
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

func LogReverseArgs(n int, givenCommit string) []string {
	return []string{
		"log",
		"--pretty=%H %P",
		"--raw",
		"-z",
		"-m",
		// --no-abbrev speeds up git log a lot
		"--no-abbrev",
		"--no-renames",
		"--first-parent",
		"--reverse",
		"--ignore-submodules",
		fmt.Sprintf("-%d", n),
		givenCommit,
	}
}

func ParseLogReverseEach(stdout io.Reader, onLogEntry func(entry LogEntry) error) error {
	reader := bufio.NewReader(stdout)

	var buf []byte

	for {
		// abc... ... NULL '\n'?

		// Read the commit
		commitBytes, err := reader.Peek(40)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		commit := string(commitBytes)

		// Skip past the NULL byte
		_, err = reader.ReadBytes(0)
		if err != nil {
			return err
		}

		// A '\n' indicates a list of paths and their statuses is next
		buf, err = reader.Peek(1)
		if err == io.EOF {
			err = onLogEntry(LogEntry{Commit: commit, PathStatuses: []PathStatus{}})
			if err != nil {
				return err
			}
			break
		} else if err != nil {
			return err
		}
		if buf[0] == '\n' {
			// A list of paths and their statuses is next

			// Skip the '\n'
			discarded, err := reader.Discard(1)
			if discarded != 1 {
				return fmt.Errorf("discarded %d bytes, expected 1", discarded)
			} else if err != nil {
				return err
			}

			pathStatuses := []PathStatus{}
			for {
				// :100644 100644 abc... def... M NULL file.txt NULL
				// ^ 0                          ^ 97   ^ 99

				// A ':' indicates a path and its status is next
				buf, err = reader.Peek(1)
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}
				if buf[0] != ':' {
					break
				}

				// Read the status from index 97 and skip to the path at index 99
				buf = make([]byte, 99)
				read, err := io.ReadFull(reader, buf)
				if read != 99 {
					return fmt.Errorf("read %d bytes, expected 99", read)
				} else if err != nil {
					return err
				}

				// Read the path
				path, err := reader.ReadBytes(0)
				if err != nil {
					return err
				}
				path = path[:len(path)-1] // Drop the trailing NULL byte

				// Inspect the status
				var status StatusAMD
				statusByte := buf[97]
				switch statusByte {
				case 'A':
					status = AddedAMD
				case 'M':
					status = ModifiedAMD
				case 'D':
					status = DeletedAMD
				case 'T':
					// Type changed. Check if it changed from a file to a submodule or vice versa,
					// treating submodules as empty.

					isSubmodule := func(mode string) bool {
						// Submodules are mode "160000". https://stackoverflow.com/questions/737673/how-to-read-the-mode-field-of-git-ls-trees-output#comment3519596_737877
						return mode == "160000"
					}

					oldMode := string(buf[1:7])
					newMode := string(buf[8:14])

					if isSubmodule(oldMode) && !isSubmodule(newMode) {
						// It changed from a submodule to a file, so consider it added.
						status = AddedAMD
						break
					}

					if !isSubmodule(oldMode) && isSubmodule(newMode) {
						// It changed from a file to a submodule, so consider it deleted.
						status = DeletedAMD
						break
					}

					// Otherwise, it remained the same, so ignore the type change.
					continue
				case 'C':
					// Copied
					return fmt.Errorf("unexpected status 'C' given --no-renames was specified")
				case 'R':
					// Renamed
					return fmt.Errorf("unexpected status 'R' given --no-renames was specified")
				case 'X':
					return fmt.Errorf("unexpected status 'X' indicates a bug in git")
				default:
					fmt.Printf("LogReverse commit %q path %q: unrecognized diff status %q, skipping\n", commit, path, string(statusByte))
					continue
				}

				pathStatuses = append(pathStatuses, PathStatus{Path: string(path), Status: status})
			}

			err = onLogEntry(LogEntry{Commit: commit, PathStatuses: pathStatuses})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RevListArgs(givenCommit string) []string {
	return []string{"rev-list", "--first-parent", givenCommit}
}

func RevListEach(stdout io.Reader, onCommit func(commit string) (shouldContinue bool, err error)) error {
	reader := bufio.NewReader(stdout)

	for {
		commit, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		commit = commit[:len(commit)-1] // Drop the trailing newline
		shouldContinue, err := onCommit(commit)
		if !shouldContinue {
			return err
		}
	}

	return nil
}

type Queryable interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// combineErrors is like errors.Append, but it returns nil when there are no errors or all errors are
// nil so that success is propagated.
func combineErrors(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	for _, err := range errs {
		if err != nil {
			return errors.Append(errs[0], errs[1:]...)
		}
	}
	return nil
}
