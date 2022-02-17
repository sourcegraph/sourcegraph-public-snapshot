package rockskip

import (
	"bufio"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/regexp"
	"github.com/grafana/regexp/syntax"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	pg "github.com/lib/pq"
	"github.com/segmentio/fasthash/fnv1"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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

type ThreadStatus struct {
	Tasklog   *TaskLog
	Name      string
	HeldLocks map[string]struct{}
	Indexed   int
	Total     int
	mu        sync.Mutex
	onEnd     func()
}

func NewThreadStatus(name string, onEnd func()) *ThreadStatus {
	return &ThreadStatus{
		Tasklog:   NewTaskLog(),
		Name:      name,
		HeldLocks: map[string]struct{}{},
		Indexed:   -1,
		Total:     -1,
		mu:        sync.Mutex{},
		onEnd:     onEnd,
	}
}

func (s *ThreadStatus) WithLock(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f()
}

func (s *ThreadStatus) SetProgress(indexed, total int) {
	s.WithLock(func() { s.Indexed = indexed; s.Total = total })
}
func (s *ThreadStatus) HoldLock(name string)    { s.WithLock(func() { s.HeldLocks[name] = struct{}{} }) }
func (s *ThreadStatus) ReleaseLock(name string) { s.WithLock(func() { delete(s.HeldLocks, name) }) }

func (s *ThreadStatus) End() {
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

// RequestId is a unique int for each HTTP request.
type RequestId = int

// ServerStatus contains the status of all requests.
type ServerStatus struct {
	threadIdToThreadStatus map[RequestId]*ThreadStatus
	nextThreadId           RequestId
	mu                     sync.Mutex
}

func NewStatus() *ServerStatus {
	return &ServerStatus{
		threadIdToThreadStatus: map[int]*ThreadStatus{},
		nextThreadId:           0,
		mu:                     sync.Mutex{},
	}
}

func (s *ServerStatus) NewThreadStatus(name string) *ThreadStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	threadId := s.nextThreadId
	s.nextThreadId++

	threadStatus := NewThreadStatus(name, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.threadIdToThreadStatus, threadId)
	})

	s.threadIdToThreadStatus[threadId] = threadStatus

	return threadStatus
}

type Server struct {
	db                 *sql.DB
	git                Git
	createParser       func() ParseSymbolsFunc
	status             *ServerStatus
	repoUpdates        chan struct{}
	maxRepos           int
	logQueries         bool
	repoCommitToDone   map[string]chan struct{}
	repoCommitToDoneMu sync.Mutex
	indexRequestQueues []chan indexRequest
}

func NewServer(
	db *sql.DB,
	git Git,
	createParser func() ParseSymbolsFunc,
	maxConcurrentlyIndexing int,
	maxRepos int,
	logQueries bool,
	indexRequestsQueueSize int,
) (*Server, error) {
	indexRequestQueues := make([]chan indexRequest, maxConcurrentlyIndexing)
	for i := 0; i < maxConcurrentlyIndexing; i++ {
		indexRequestQueues[i] = make(chan indexRequest, indexRequestsQueueSize)
	}

	server := &Server{
		db:                 db,
		git:                git,
		createParser:       createParser,
		status:             NewStatus(),
		repoUpdates:        make(chan struct{}, 1),
		maxRepos:           maxRepos,
		logQueries:         logQueries,
		repoCommitToDone:   map[string]chan struct{}{},
		repoCommitToDoneMu: sync.Mutex{},
		indexRequestQueues: indexRequestQueues,
	}

	err := server.startCleanupThread()
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxConcurrentlyIndexing; i++ {
		err = server.startIndexingThread(server.indexRequestQueues[i])
		if err != nil {
			return nil, err
		}
	}

	return server, nil
}

type repoCommit struct {
	repo   string
	commit string
}

type indexRequest struct {
	repoCommit
	done chan struct{}
}

func (s *Server) startIndexingThread(indexRequestQueue chan indexRequest) (err error) {
	go func() {
		for indexRequest := range indexRequestQueue {
			// Get a fresh connection from the DB pool to get deterministic "lock stacking" behavior.
			// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
			conn, err := s.db.Conn(context.Background())
			if err != nil {
				log15.Error("Failed to get connection for indexing thread", "error", err)
				continue
			}

			err = s.Index(context.Background(), conn, indexRequest.repo, indexRequest.commit, s.createParser())
			close(indexRequest.done)
			if err != nil {
				log15.Error("indexing error", "repo", indexRequest.repo, "commit", indexRequest.commit, "err", err)
			}

			conn.Close()
		}
	}()

	return nil
}

func (s *Server) startCleanupThread() error {
	go func() {
		for range s.repoUpdates {
			// Get a fresh connection from the DB pool to get deterministic "lock stacking" behavior.
			// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
			conn, err := s.db.Conn(context.Background())
			if err != nil {
				log15.Error("Failed to get connection for deleting old repos", "error", err)
				continue
			}

			threadStatus := s.status.NewThreadStatus("cleanup")
			err = DeleteOldRepos(context.Background(), conn, s.maxRepos, threadStatus)
			threadStatus.End()
			if err != nil {
				log15.Error("Failed to delete old repos", "error", err)
			}

			conn.Close()
		}
	}()

	return nil
}

func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	repositoryCount, _, err := basestore.ScanFirstInt(s.db.QueryContext(ctx, "SELECT COUNT(*) FROM rockskip_repos"))
	if err != nil {
		log15.Error("Failed to count repos", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type repoRow struct {
		repo           string
		lastAccessedAt time.Time
	}

	repoRows := []repoRow{}
	repoSqlRows, err := s.db.QueryContext(ctx, "SELECT repo, last_accessed_at FROM rockskip_repos ORDER BY last_accessed_at DESC LIMIT 5")
	if err != nil {
		log15.Error("Failed to list repoRows", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer repoSqlRows.Close()
	for repoSqlRows.Next() {
		var repo string
		var lastAccessedAt time.Time
		if err := repoSqlRows.Scan(&repo, &lastAccessedAt); err != nil {
			log15.Error("Failed to scan repo", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		repoRows = append(repoRows, repoRow{repo: repo, lastAccessedAt: lastAccessedAt})
	}

	blobsSize, _, err := basestore.ScanFirstString(s.db.QueryContext(ctx, "SELECT pg_size_pretty(pg_total_relation_size('rockskip_blobs'))"))
	if err != nil {
		log15.Error("Failed to get size of blobs table", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "This is the symbols service status page.")
	fmt.Fprintln(w, "")

	fmt.Fprintf(w, "Number of repositories: %d\n", repositoryCount)
	fmt.Fprintf(w, "Size of blobs table: %s\n", blobsSize)
	fmt.Fprintln(w, "")

	if repositoryCount > 0 {
		fmt.Fprintf(w, "Most recently searched repositories (at most 5 shown)\n")
		for _, repo := range repoRows {
			fmt.Fprintf(w, "  %s %s\n", repo.lastAccessedAt, repo.repo)
		}
		fmt.Fprintln(w, "")
	}

	s.status.mu.Lock()
	defer s.status.mu.Unlock()

	if len(s.status.threadIdToThreadStatus) == 0 {
		fmt.Fprintln(w, "No requests in flight.")
		return
	}
	fmt.Fprintln(w, "Here are all in-flight requests:")
	fmt.Fprintln(w, "")

	ids := []int{}
	for id := range s.status.threadIdToThreadStatus {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		status := s.status.threadIdToThreadStatus[id]
		status.WithLock(func() {
			fmt.Fprintf(w, "%s\n", status.Name)
			if status.Total > 0 {
				fmt.Fprintf(w, "    progress %.2f%% (indexed %d of %d commits)\n", float64(status.Indexed)/float64(status.Total)*100, status.Indexed, status.Total)
			}
			fmt.Fprintf(w, "    %s\n", status.Tasklog)
			locks := []string{}
			for lock := range status.HeldLocks {
				locks = append(locks, lock)
			}
			sort.Strings(locks)
			for _, lock := range locks {
				fmt.Fprintf(w, "    holding %s\n", lock)
			}
			fmt.Fprintln(w)
		})
	}
}

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
	tipHeight := 0

	missingCount := 0
	tasklog.Start("RevList")
	err = s.git.RevListEach(repo, givenCommit, func(commit string) (shouldContinue bool, err error) {
		defer tasklog.Continue("RevList")

		tasklog.Start("GetCommit")
		_, height, present, err := GetCommit(ctx, conn, repo, commit)
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

	threadStatus.SetProgress(0, missingCount)

	if missingCount == 0 {
		return nil
	}

	pathToBlobIdCache := map[string]int{}

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

		hops, err := getHops(ctx, tx, repo, tipCommit, tasklog)
		if err != nil {
			return errors.Wrap(err, "getHops")
		}

		r := ruler(tipHeight + 1)
		if r >= len(hops) {
			return errors.Newf("ruler(%d) = %d is out of range of len(hops) = %d", tipHeight+1, r, len(hops))
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
					return errors.Newf("could not find blob for path %s", deletedPath)
				}
			}

			tasklog.Start("UpdateBlobHops")
			err = UpdateBlobHops(ctx, tx, id, DeletedAD, entry.Commit)
			if err != nil {
				return errors.Wrap(err, "UpdateBlobHops")
			}
		}

		tasklog.Start("ArchiveEach")
		err = s.git.ArchiveEach(repo, entry.Commit, addedPaths, func(addedPath string, contents []byte) error {
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

	threadStatus.SetProgress(entriesIndexed, missingCount)

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

func DeleteOldRepos(ctx context.Context, db *sql.Conn, maxRepos int, threadStatus *ThreadStatus) error {
	// Keep deleting repos until we're back to at most maxRepos.
	for {
		more, err := tryDeleteOldestRepo(ctx, db, maxRepos, threadStatus)
		if err != nil {
			return err
		}
		if !more {
			return nil
		}
	}
}

func tryDeleteOldestRepo(ctx context.Context, db *sql.Conn, maxRepos int, threadStatus *ThreadStatus) (more bool, err error) {
	defer threadStatus.Tasklog.Continue("idle")

	// Select a candidate repo to delete.
	threadStatus.Tasklog.Start("select repo to delete")
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
	releaseWLock, err := wLock(ctx, db, threadStatus, repo)
	defer func() { err = combineErrors(err, releaseWLock()) }()
	if err != nil {
		return false, errors.Wrap(err, "acquiring write lock on repo")
	}

	// Make sure the repo is still old. See note above.
	var rank int
	threadStatus.Tasklog.Start("recheck repo rank")
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
	releaseILock, err := iLock(ctx, db, threadStatus, repo)
	defer func() { err = combineErrors(err, releaseILock()) }()
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
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_ancestry WHERE repo = $1;", repo)
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_blobs WHERE $1 && repo;", pg.Array([]string{repo}))
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

func updateLastAccessedAt(ctx context.Context, db Queryable, repo string) error {
	_, err := db.ExecContext(ctx, `
			INSERT INTO rockskip_repos (repo, last_accessed_at)
			VALUES ($1, now())
			ON CONFLICT (repo)
			DO UPDATE SET last_accessed_at = now()
		`, repo)
	if err != nil {
		return err
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
			return errors.Newf("unexpected cat-file output: %q", line)
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
			return errors.Newf("expected to discard 1 byte, but discarded %d", discarded)
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
		return "", 0, false, errors.Newf("GetCommit: %s", err)
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
		WHERE $1 && path AND $2 && added AND NOT $2 && deleted
	`, pg.Array([]string{path}), pg.Array([]string{hop})).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	} else if err != nil {
		return 0, false, errors.Newf("GetBlob: %s", err)
	}
	return id, true, nil
}

func UpdateBlobHops(ctx context.Context, db Queryable, id int, status StatusAD, hop string) error {
	column := statusADToColumn(status)
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
	`, pg.Array([]string{repo}), blob.Commit, pg.Array([]string{blob.Path}), pg.Array(blob.Added), pg.Array(blob.Deleted), pg.Array(symbolNames), Symbols(blob.Symbols)).Scan(&lastInsertId)
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

func (s *Server) Search(ctx context.Context, args types.SearchArgs) (blobs []Blob, err error) {
	repo := string(args.Repo)
	commit := string(args.CommitID)

	threadStatus := s.status.NewThreadStatus(fmt.Sprintf("searching %+v", args))
	if s.logQueries {
		defer threadStatus.Tasklog.Print()
	}
	defer threadStatus.End()

	// Acquire a read lock on the repo.
	locked, releaseRLock, err := tryRLock(ctx, s.db, threadStatus, repo)
	if err != nil {
		return nil, err
	}
	defer func() { err = combineErrors(err, releaseRLock()) }()
	if !locked {
		return nil, errors.Newf("deletion in progress", repo)
	}

	// Insert or set the last_accessed_at column for this repo to now() in the rockskip_repos table.
	threadStatus.Tasklog.Start("update last_accessed_at")
	err = updateLastAccessedAt(ctx, s.db, repo)
	if err != nil {
		return nil, err
	}

	// Non-blocking send on repoUpdates to notify the background deletion goroutine.
	select {
	case s.repoUpdates <- struct{}{}:
	default:
	}

	// Check if the commit has already been indexed, and if not then index it.
	threadStatus.Tasklog.Start("check commit presence")
	_, _, present, err := GetCommit(ctx, s.db, repo, commit)
	if err != nil {
		return nil, err
	} else if !present {

		// Try to send an index request.
		done, err := s.emitIndexRequest(repoCommit{repo: repo, commit: commit})
		if err != nil {
			return nil, err
		}

		// Wait for indexing to complete or the request to be canceled.
		threadStatus.Tasklog.Start("awaiting indexing completion")
		select {
		case <-done:
			threadStatus.Tasklog.Start("recheck commit presence")
			_, _, present, err = GetCommit(ctx, s.db, repo, commit)
			if err != nil {
				return nil, err
			}
			if !present {
				return nil, errors.Newf("indexing failed, check server logs")
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}

	// Finally search.
	blobs, err = queryBlobs(ctx, s.db, args, threadStatus, s.logQueries)
	if err != nil {
		return nil, err
	}

	return blobs, nil
}

func (s *Server) emitIndexRequest(rc repoCommit) (chan struct{}, error) {
	key := fmt.Sprintf("%s@%s", rc.repo, rc.commit)

	s.repoCommitToDoneMu.Lock()

	if done, ok := s.repoCommitToDone[key]; ok {
		s.repoCommitToDoneMu.Unlock()
		return done, nil
	}

	done := make(chan struct{})

	s.repoCommitToDone[key] = done
	s.repoCommitToDoneMu.Unlock()
	go func() {
		<-done
		s.repoCommitToDoneMu.Lock()
		delete(s.repoCommitToDone, key)
		s.repoCommitToDoneMu.Unlock()
	}()

	request := indexRequest{
		repoCommit: repoCommit{
			repo:   rc.repo,
			commit: rc.commit,
		},
		done: done}

	// Route the index request to the indexer associated with the repo.
	ix := int(fnv1.HashString32(rc.repo)) % len(s.indexRequestQueues)

	select {
	case s.indexRequestQueues[ix] <- request:
	default:
		return nil, errors.Newf("the indexing queue is full")
	}

	return done, nil
}

func queryBlobs(ctx context.Context, db Queryable, args types.SearchArgs, threadStatus *ThreadStatus, logQueries bool) ([]Blob, error) {
	hops, err := getHops(ctx, db, string(args.Repo), string(args.CommitID), threadStatus.Tasklog)
	if err != nil {
		return nil, err
	}

	limit := 100
	if args.First > 0 {
		limit = args.First
	}

	threadStatus.Tasklog.Start("run query")
	q := sqlf.Sprintf(`
		SELECT id, commit_id, path, added, deleted, symbol_data
		FROM rockskip_blobs
		WHERE
			%s && repo
			AND     %s && added
			AND NOT %s && deleted
			AND %s
		LIMIT %s;`,
		pg.Array([]string{string(args.Repo)}),
		pg.Array(hops),
		pg.Array(hops),
		convertSearchArgsToSqlQuery(args),
		limit,
	)

	if logQueries {
		err = logQuery(ctx, db, args, q)
		if err != nil {
			return nil, errors.Wrap(err, "logQuery")
		}
	}

	var rows *sql.Rows
	rows, err = db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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
			return nil, errors.Newf("Search: expected 1 path, got %d", len(paths))
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

func logQuery(ctx context.Context, db Queryable, args types.SearchArgs, q *sqlf.Query) error {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "Search args: %+v\n", args)

	fmt.Fprintln(sb, "Query:")
	query, err := sqlfToString(q)
	if err != nil {
		return errors.Wrap(err, "sqlfToString")
	}

	fmt.Fprintln(sb, query)

	fmt.Println(bracket(sb.String()))

	return nil
}

func bracket(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i, line := range lines {
		if i == 0 {
			lines[i] = "┌ " + line
		} else if i == len(lines)-1 {
			lines[i] = "└ " + line
		} else {
			lines[i] = "│ " + line
		}
	}
	return strings.Join(lines, "\n")
}

func sqlfToString(q *sqlf.Query) (string, error) {
	s := q.Query(sqlf.PostgresBindVar)
	for i, arg := range q.Args() {
		argString, err := argToString(arg)
		if err != nil {
			return "", err
		}
		s = strings.ReplaceAll(s, fmt.Sprintf("$%d", i+1), argString)
	}
	return s, nil
}

func argToString(arg interface{}) (string, error) {
	switch arg := arg.(type) {
	case string:
		return fmt.Sprintf("'%s'", sqlEscapeQuotes(arg)), nil
	case *pg.StringArray:
		value, err := arg.Value()
		if err != nil {
			return "", err
		}
		switch value := value.(type) {
		case string:
			return fmt.Sprintf("'%s'", sqlEscapeQuotes(value)), nil
		default:
			return "", errors.Newf("unrecognized array type %T", value)
		}
	case int:
		return fmt.Sprintf("%d", arg), nil
	default:
		return "", errors.Newf("unrecognized type %T", arg)
	}
}

func sqlEscapeQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
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
		return sqlf.Sprintf(fmt.Sprintf("%%s && %s", column), pg.Array([]string{literal}))
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

func lock(ctx context.Context, db Queryable, threadStatus *ThreadStatus, namespace int32, name, repo, lockFn, unlockFn string) (func() error, error) {
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

func tryLock(ctx context.Context, db Queryable, threadStatus *ThreadStatus, namespace int32, name, repo, lockFn, unlockFn string) (bool, func() error, error) {
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
func tryRLock(ctx context.Context, db Queryable, threadStatus *ThreadStatus, repo string) (bool, func() error, error) {
	return tryLock(ctx, db, threadStatus, RW_LOCKS_NAMESPACE, "rLock", repo, "pg_try_advisory_lock_shared", "pg_advisory_unlock_shared")
}

// wLock acquires the write lock on the repo. It blocks only when another connection holds a read or the
// write lock. That means a single connection can acquire the write lock while holding a read lock.
func wLock(ctx context.Context, db Queryable, threadStatus *ThreadStatus, repo string) (func() error, error) {
	return lock(ctx, db, threadStatus, RW_LOCKS_NAMESPACE, "wLock", repo, "pg_advisory_lock", "pg_advisory_unlock")
}

// iLock acquires the indexing lock on the repo.
func iLock(ctx context.Context, db Queryable, threadStatus *ThreadStatus, repo string) (func() error, error) {
	return lock(ctx, db, threadStatus, INDEXING_LOCKS_NAMESPACE, "iLock", repo, "pg_advisory_lock", "pg_advisory_unlock")
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
			return errors.Newf("Search: expected 1 path, got %d", len(paths))
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
				return errors.Newf("discarded %d bytes, expected 1", discarded)
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
					return errors.Newf("read %d bytes, expected 99", read)
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
					return errors.Newf("unexpected status 'C' given --no-renames was specified")
				case 'R':
					// Renamed
					return errors.Newf("unexpected status 'R' given --no-renames was specified")
				case 'X':
					return errors.Newf("unexpected status 'X' indicates a bug in git")
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
func combineErrors(errOrNils ...error) error {
	errs := []error{}
	for _, err := range errOrNils {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return errors.Append(errs[0], errs[1:]...)
}
