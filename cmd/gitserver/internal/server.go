// Package internal implements the gitserver service.
package internal

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/accesslog"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/perforce"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/urlredactor"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/vcssyncer"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/limiter"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// traceLogs is controlled via the env SRC_GITSERVER_TRACE. If true we trace
// logs to stderr
var traceLogs bool

var (
	lastCheckAt    = make(map[api.RepoName]time.Time)
	lastCheckMutex sync.Mutex
)

// debounce() provides some filtering to prevent spammy requests for the same
// repository. If the last fetch of the repository was within the given
// duration, returns false, otherwise returns true and updates the last
// fetch stamp.
func debounce(name api.RepoName, since time.Duration) bool {
	lastCheckMutex.Lock()
	defer lastCheckMutex.Unlock()
	if t, ok := lastCheckAt[name]; ok && time.Now().Before(t.Add(since)) {
		return false
	}
	lastCheckAt[name] = time.Now()
	return true
}

func init() {
	traceLogs, _ = strconv.ParseBool(env.Get("SRC_GITSERVER_TRACE", "false", "Toggles trace logging to stderr"))
}

// cloneJob abstracts away a repo and necessary metadata to clone it. In the future it may be
// possible to simplify this, but to do that, doClone will need to do a lot less than it does at the
// moment.
type cloneJob struct {
	repo   api.RepoName
	dir    common.GitDir
	syncer vcssyncer.VCSSyncer

	// TODO: cloneJobConsumer should acquire a new lock. We are trying to keep the changes simple
	// for the time being. When we start using the new approach of using long lived goroutines for
	// cloning we will refactor doClone to acquire a new lock.
	lock RepositoryLock

	remoteURL *vcs.URL
	options   CloneOptions
}

// cloneTask is a thin wrapper around a cloneJob to associate the doneFunc with each job.
type cloneTask struct {
	*cloneJob
	done func() time.Duration
}

// NewCloneQueue initializes a new cloneQueue.
func NewCloneQueue(obctx *observation.Context, jobs *list.List) *common.Queue[*cloneJob] {
	return common.NewQueue[*cloneJob](obctx, "clone-queue", jobs)
}

// Server is a gitserver server.
type Server struct {
	// Logger should be used for all logging and logger creation.
	Logger log.Logger

	// ObservationCtx is used to initialize an operations struct
	// with the appropriate metrics register etc.
	ObservationCtx *observation.Context

	// ReposDir is the path to the base directory for gitserver storage.
	ReposDir string

	// GetRemoteURLFunc is a function which returns the remote URL for a
	// repository. This is used when cloning or fetching a repository. In
	// production this will speak to the database to look up the clone URL. In
	// tests this is usually set to clone a local repository or intentionally
	// error.
	GetRemoteURLFunc func(context.Context, api.RepoName) (string, error)

	// GetVCSSyncer is a function which returns the VCS syncer for a repository.
	// This is used when cloning or fetching a repository. In production this will
	// speak to the database to determine the code host type. In tests this is
	// usually set to return a GitRepoSyncer.
	GetVCSSyncer func(context.Context, api.RepoName) (vcssyncer.VCSSyncer, error)

	// Hostname is how we identify this instance of gitserver. Generally it is the
	// actual hostname but can also be overridden by the HOSTNAME environment variable.
	Hostname string

	// DB provides access to datastores.
	DB database.DB

	// CloneQueue is a threadsafe queue used by DoBackgroundClones to process incoming clone
	// requests asynchronously.
	CloneQueue *common.Queue[*cloneJob]

	// Locker is used to lock repositories while fetching to prevent concurrent work.
	Locker RepositoryLocker

	// skipCloneForTests is set by tests to avoid clones.
	skipCloneForTests bool

	// ctx is the context we use for all background jobs. It is done when the
	// server is stopped. Do not directly call this, rather call
	// Server.context()
	ctx      context.Context
	cancel   context.CancelFunc // used to shutdown background jobs
	cancelMu sync.Mutex         // protects canceled
	canceled bool
	wg       sync.WaitGroup // tracks running background jobs

	// cloneLimiter and cloneableLimiter limits the number of concurrent
	// clones and ls-remotes respectively. Use s.acquireCloneLimiter() and
	// s.acquireCloneableLimiter() instead of using these directly.
	cloneLimiter     *limiter.MutableLimiter
	cloneableLimiter *limiter.MutableLimiter

	// RPSLimiter limits the remote code host git operations done per second
	// per gitserver instance
	RPSLimiter *ratelimit.InstrumentedLimiter

	repoUpdateLocksMu sync.Mutex // protects the map below and also updates to locks.once
	repoUpdateLocks   map[api.RepoName]*locks

	// GlobalBatchLogSemaphore is a semaphore shared between all requests to ensure that a
	// maximum number of Git subprocesses are active for all /batch-log requests combined.
	GlobalBatchLogSemaphore *semaphore.Weighted

	// operations provide uniform observability via internal/observation. This value is
	// set by RegisterMetrics when compiled as part of the gitserver binary. The server
	// method ensureOperations should be used in all references to avoid a nil pointer
	// dereferences.
	operations *operations

	// RecordingCommandFactory is a factory that creates recordable commands by wrapping os/exec.Commands.
	// The factory creates recordable commands with a set predicate, which is used to determine whether a
	// particular command should be recorded or not.
	RecordingCommandFactory *wrexec.RecordingCommandFactory

	// Perforce is a plugin-like service attached to Server for all things Perforce.
	Perforce *perforce.Service
}

type locks struct {
	once *sync.Once  // consolidates multiple waiting updates
	mu   *sync.Mutex // prevents updates running in parallel
}

// shortGitCommandSlow returns the threshold for regarding an git command as
// slow. Some commands such as "git archive" are inherently slower than "git
// rev-parse", so this will return an appropriate threshold given the command.
func shortGitCommandSlow(args []string) time.Duration {
	if len(args) < 1 {
		return time.Second
	}
	switch args[0] {
	case "archive":
		return 1 * time.Minute

	case "blame", "ls-tree", "log", "show":
		return 5 * time.Second

	default:
		return 2500 * time.Millisecond
	}
}

// 🚨 SECURITY: headerXRequestedWithMiddleware will ensure that the X-Requested-With
// header contains the correct value. See "What does X-Requested-With do, anyway?" in
// https://github.com/sourcegraph/sourcegraph/pull/27931.
func headerXRequestedWithMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := log.Scoped("gitserver")

		// Do not apply the middleware to /ping and /git endpoints.
		//
		// 1. /ping is used by health check services who most likely don't set this header
		// at all.
		//
		// 2. /git may be used to run "git fetch" from another gitserver instance over
		// HTTP and the fetchCommand does not set this header yet.
		if strings.HasPrefix(r.URL.Path, "/ping") || strings.HasPrefix(r.URL.Path, "/git") {
			next.ServeHTTP(w, r)
			return
		}

		if value := r.Header.Get("X-Requested-With"); value != "Sourcegraph" {
			l.Error("header X-Requested-With is not set or is invalid", log.String("path", r.URL.Path))
			http.Error(w, "header X-Requested-With is not set or is invalid", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.repoUpdateLocks = make(map[api.RepoName]*locks)

	// GitMaxConcurrentClones controls the maximum number of clones that
	// can happen at once on a single gitserver.
	// Used to prevent throttle limits from a code host. Defaults to 5.
	//
	// The new repo-updater scheduler enforces the rate limit across all gitserver,
	// so ideally this logic could be removed here; however, ensureRevision can also
	// cause an update to happen and it is called on every exec command.
	// Max concurrent clones also means repo updates.
	maxConcurrentClones := conf.GitMaxConcurrentClones()
	s.cloneLimiter = limiter.NewMutable(maxConcurrentClones)
	s.cloneableLimiter = limiter.NewMutable(maxConcurrentClones)

	// TODO: Remove side-effects from this Handler method.
	conf.Watch(func() {
		limit := conf.GitMaxConcurrentClones()
		s.cloneLimiter.SetLimit(limit)
		s.cloneableLimiter.SetLimit(limit)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/archive", trace.WithRouteName("archive", accesslog.HTTPMiddleware(
		s.Logger.Scoped("archive.accesslog"),
		conf.DefaultClient(),
		s.handleArchive,
	)))
	mux.HandleFunc("/exec", trace.WithRouteName("exec", accesslog.HTTPMiddleware(
		s.Logger.Scoped("exec.accesslog"),
		conf.DefaultClient(),
		s.handleExec,
	)))
	mux.HandleFunc("/search", trace.WithRouteName("search", s.handleSearch))
	mux.HandleFunc("/batch-log", trace.WithRouteName("batch-log", s.handleBatchLog))
	mux.HandleFunc("/p4-exec", trace.WithRouteName("p4-exec", accesslog.HTTPMiddleware(
		s.Logger.Scoped("p4-exec.accesslog"),
		conf.DefaultClient(),
		s.handleP4Exec,
	)))
	mux.HandleFunc("/list-gitolite", trace.WithRouteName("list-gitolite", s.handleListGitolite))
	mux.HandleFunc("/is-repo-cloneable", trace.WithRouteName("is-repo-cloneable", s.handleIsRepoCloneable))
	mux.HandleFunc("/repo-clone-progress", trace.WithRouteName("repo-clone-progress", s.handleRepoCloneProgress))
	mux.HandleFunc("/delete", trace.WithRouteName("delete", s.handleRepoDelete))
	mux.HandleFunc("/repo-update", trace.WithRouteName("repo-update", s.handleRepoUpdate))
	mux.HandleFunc("/repo-clone", trace.WithRouteName("repo-clone", s.handleRepoClone))
	mux.HandleFunc("/create-commit-from-patch-binary", trace.WithRouteName("create-commit-from-patch-binary", s.handleCreateCommitFromPatchBinary))
	mux.HandleFunc("/disk-info", trace.WithRouteName("disk-info", s.handleDiskInfo))
	mux.HandleFunc("/is-perforce-path-cloneable", trace.WithRouteName("is-perforce-path-cloneable", s.handleIsPerforcePathCloneable))
	mux.HandleFunc("/check-perforce-credentials", trace.WithRouteName("check-perforce-credentials", s.handleCheckPerforceCredentials))
	mux.HandleFunc("/commands/get-object", trace.WithRouteName("commands/get-object", s.handleGetObject))
	mux.HandleFunc("/perforce-users", trace.WithRouteName("perforce-users", s.handlePerforceUsers))
	mux.HandleFunc("/perforce-protects-for-user", trace.WithRouteName("perforce-protects-for-user", s.handlePerforceProtectsForUser))
	mux.HandleFunc("/perforce-protects-for-depot", trace.WithRouteName("perforce-protects-for-depot", s.handlePerforceProtectsForDepot))
	mux.HandleFunc("/perforce-group-members", trace.WithRouteName("perforce-group-members", s.handlePerforceGroupMembers))
	mux.HandleFunc("/is-perforce-super-user", trace.WithRouteName("is-perforce-super-user", s.handleIsPerforceSuperUser))
	mux.HandleFunc("/perforce-get-changelist", trace.WithRouteName("perforce-get-changelist", s.handlePerforceGetChangelist))
	mux.HandleFunc("/ping", trace.WithRouteName("ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// This endpoint allows us to expose gitserver itself as a "git service"
	// (ETOOMANYGITS!) that allows other services to run commands like "git fetch"
	// directly against a gitserver replica and treat it as a git remote.
	//
	// Example use case for this is a repo migration from one replica to another during
	// scaling events and the new destination gitserver replica can directly clone from
	// the gitserver replica which hosts the repository currently.
	mux.HandleFunc("/git/", trace.WithRouteName("git", accesslog.HTTPMiddleware(
		s.Logger.Scoped("git.accesslog"),
		conf.DefaultClient(),
		func(rw http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/git", s.gitServiceHandler()).ServeHTTP(rw, r)
		},
	)))

	// 🚨 SECURITY: This must be wrapped in headerXRequestedWithMiddleware.
	return headerXRequestedWithMiddleware(mux)
}

func addrForRepo(ctx context.Context, repoName api.RepoName, gitServerAddrs gitserver.GitserverAddresses) string {
	return gitServerAddrs.AddrForRepo(ctx, filepath.Base(os.Args[0]), repoName)
}

// NewClonePipeline creates a new pipeline that clones repos asynchronously. It
// creates a producer-consumer pipeline that handles clone requests asychronously.
func (s *Server) NewClonePipeline(logger log.Logger, cloneQueue *common.Queue[*cloneJob]) goroutine.BackgroundRoutine {
	return &clonePipelineRoutine{
		tasks:  make(chan *cloneTask),
		logger: logger,
		s:      s,
		queue:  cloneQueue,
	}
}

type clonePipelineRoutine struct {
	logger     log.Logger
    tasks      chan *cloneTask
    db         database.GitserverRepoStore
    rpsLimiter // RPS limiter shared with the server	
}

func (p* clonePipelineroutine) doClone() ...

func (p *clonePipelineRoutine) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	// Start a go routine for each the producer and the consumer.
	go p.cloneJobConsumer(ctx, p.tasks)
	go p.cloneJobProducer(ctx, p.tasks)
}

func (p *clonePipelineRoutine) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *clonePipelineRoutine) cloneJobProducer(ctx context.Context, tasks chan<- *cloneTask) {
	defer close(tasks)

	for {
		// Acquire the cond mutex lock and wait for a signal if the queue is empty.
		p.queue.Mutex.Lock()
		if p.queue.Empty() {
			// TODO: This should only wait if ctx is not canceled.
			p.queue.Cond.Wait()
		}

		// The queue is not empty and we have a job to process! But don't forget to unlock the cond
		// mutex here as we don't need to hold the lock beyond this point for now.
		p.queue.Mutex.Unlock()

		// Keep popping from the queue until the queue is empty again, in which case we start all
		// over again from the top.
		for {
			job, doneFunc := p.queue.Pop()
			if job == nil {
				break
			}

			select {
			case tasks <- &cloneTask{
				cloneJob: *job,
				done:     doneFunc,
			}:
			case <-ctx.Done():
				p.logger.Error("cloneJobProducer", log.Error(ctx.Err()))
				return
			}
		}
	}
}

func (p *clonePipelineRoutine) cloneJobConsumer(ctx context.Context, tasks <-chan *cloneTask) {
	logger := p.logger.Scoped("cloneJobConsumer")

	for task := range tasks {
		logger := logger.With(log.String("job.repo", string(task.repo)))

		select {
		case <-ctx.Done():
			logger.Error("context done", log.Error(ctx.Err()))
			return
		default:
		}

		ctx, cancel, err := p.s.acquireCloneLimiter(ctx)
		if err != nil {
			logger.Error("acquireCloneLimiter", log.Error(err))
			continue
		}

		go func(task *cloneTask) {
			defer cancel()

			err := p.s.doClone(ctx, task.repo, task.dir, task.syncer, task.lock, task.remoteURL, task.options)
			if err != nil {
				logger.Error("failed to clone repo", log.Error(err))
			}
			// Use a different context in case we failed because the original context failed.
			p.s.setLastErrorNonFatal(p.s.ctx, task.repo, err)
			_ = task.done()
		}(task)
	}
}

// repoCloned checks if dir or `${dir}/.git` is a valid GIT_DIR.
var repoCloned = func(dir common.GitDir) bool {
	_, err := os.Stat(dir.Path("HEAD"))
	return !os.IsNotExist(err)
}

// Stop cancels the running background jobs and returns when done.
func (s *Server) Stop() {
	// idempotent so we can just always set and cancel
	s.cancel()
	s.cancelMu.Lock()
	s.canceled = true
	s.cancelMu.Unlock()
	s.wg.Wait()
}

// serverContext returns a child context tied to the lifecycle of server.
func (s *Server) serverContext() (context.Context, context.CancelFunc) {
	// if we are already canceled don't increment our WaitGroup. This is to
	// prevent a loop somewhere preventing us from ever finishing the
	// WaitGroup, even though all calls fails instantly due to the canceled
	// context.
	s.cancelMu.Lock()
	if s.canceled {
		s.cancelMu.Unlock()
		return s.ctx, func() {}
	}
	s.wg.Add(1)
	s.cancelMu.Unlock()

	ctx, cancel := context.WithCancel(s.ctx)

	// we need to track if we have called cancel, since we are only allowed to
	// call wg.Done() once, but CancelFuncs can be called any number of times.
	var canceled int32
	return ctx, func() {
		ok := atomic.CompareAndSwapInt32(&canceled, 0, 1)
		if ok {
			cancel()
			s.wg.Done()
		}
	}
}

func (s *Server) getRemoteURL(ctx context.Context, name api.RepoName) (*vcs.URL, error) {
	remoteURL, err := s.GetRemoteURLFunc(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "GetRemoteURLFunc")
	}

	return vcs.ParseURL(remoteURL)
}

// acquireCloneLimiter() acquires a cancellable context associated with the
// clone limiter.
func (s *Server) acquireCloneLimiter(ctx context.Context) (context.Context, context.CancelFunc, error) {
	pendingClones.Inc()
	defer pendingClones.Dec()
	return s.cloneLimiter.Acquire(ctx)
}

func (s *Server) acquireCloneableLimiter(ctx context.Context) (context.Context, context.CancelFunc, error) {
	lsRemoteQueue.Inc()
	defer lsRemoteQueue.Dec()
	return s.cloneableLimiter.Acquire(ctx)
}

func (s *Server) handleIsRepoCloneable(w http.ResponseWriter, r *http.Request) {
	var req protocol.IsRepoCloneableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Repo == "" {
		http.Error(w, "no Repo given", http.StatusBadRequest)
		return
	}
	resp, err := s.isRepoCloneable(r.Context(), req.Repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) isRepoCloneable(ctx context.Context, repo api.RepoName) (protocol.IsRepoCloneableResponse, error) {
	// We use an internal actor here as the repo may be private. It is safe since all
	// we return is a bool indicating whether the repo is cloneable or not. Perhaps
	// the only things that could leak here is whether a private repo exists although
	// the endpoint is only available internally so it's low risk.
	remoteURL, err := s.getRemoteURL(actor.WithInternalActor(ctx), repo)
	if err != nil {
		return protocol.IsRepoCloneableResponse{}, errors.Wrap(err, "getRemoteURL")
	}

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return protocol.IsRepoCloneableResponse{}, errors.Wrap(err, "GetVCSSyncer")
	}

	resp := protocol.IsRepoCloneableResponse{
		Cloned: repoCloned(gitserverfs.RepoDirFromName(s.ReposDir, repo)),
	}
	err = syncer.IsCloneable(ctx, repo, remoteURL)
	if err != nil {
		resp.Reason = err.Error()
	}
	resp.Cloneable = err == nil

	return resp, nil
}

// handleRepoUpdate is a synchronous (waits for update to complete or
// time out) method so it can yield errors. Updates are not
// unconditional; we debounce them based on the provided
// interval, to avoid spam.
func (s *Server) handleRepoUpdate(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := s.repoUpdate(&req)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) repoUpdate(req *protocol.RepoUpdateRequest) protocol.RepoUpdateResponse {
	logger := s.Logger.Scoped("handleRepoUpdate")
	var resp protocol.RepoUpdateResponse
	req.Repo = protocol.NormalizeRepo(req.Repo)
	dir := gitserverfs.RepoDirFromName(s.ReposDir, req.Repo)

	// despite the existence of a context on the request, we don't want to
	// cancel the git commands partway through if the request terminates.
	ctx, cancel1 := s.serverContext()
	defer cancel1()
	ctx, cancel2 := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel2()

	if !repoCloned(dir) && !s.skipCloneForTests {
		_, err := s.CloneRepo(ctx, req.Repo, CloneOptions{Block: true})
		if err != nil {
			logger.Warn("error cloning repo", log.String("repo", string(req.Repo)), log.Error(err))
			resp.Error = err.Error()
		}
		return resp
	}

	var statusErr, updateErr error

	if debounce(req.Repo, req.Since) {
		updateErr = s.doRepoUpdate(ctx, req.Repo, "")
	}

	// attempts to acquire these values are not contingent on the success of
	// the update.
	lastFetched, err := repoLastFetched(dir)
	if err != nil {
		statusErr = err
	} else {
		resp.LastFetched = &lastFetched
	}
	lastChanged, err := repoLastChanged(dir)
	if err != nil {
		statusErr = err
	} else {
		resp.LastChanged = &lastChanged
	}
	if statusErr != nil {
		logger.Error("failed to get status of repo", log.String("repo", string(req.Repo)), log.Error(statusErr))
		// report this error in-band, but still produce a valid response with the
		// other information.
		resp.Error = statusErr.Error()
	}
	// If an error occurred during update, report it but don't actually make
	// it into an http error; we want the client to get the information cleanly.
	// An update error "wins" over a status error.
	if updateErr != nil {
		resp.Error = updateErr.Error()
	} else {
		s.Perforce.EnqueueChangelistMappingJob(perforce.NewChangelistMappingJob(req.Repo, dir))
	}

	return resp
}

// handleRepoClone is an asynchronous (does not wait for update to complete or
// time out) call to clone a repository.
// Asynchronous errors will have to be checked in the gitserver_repos table under last_error.
func (s *Server) handleRepoClone(w http.ResponseWriter, r *http.Request) {
	logger := s.Logger.Scoped("handleRepoClone")
	var req protocol.RepoCloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var resp protocol.RepoCloneResponse
	req.Repo = protocol.NormalizeRepo(req.Repo)

	_, err := s.CloneRepo(context.Background(), req.Repo, CloneOptions{Block: false})
	if err != nil {
		logger.Warn("error cloning repo", log.String("repo", string(req.Repo)), log.Error(err))
		resp.Error = err.Error()
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
	var (
		logger    = s.Logger.Scoped("handleArchive")
		q         = r.URL.Query()
		treeish   = q.Get("treeish")
		repo      = q.Get("repo")
		format    = q.Get("format")
		pathspecs = q["path"]
	)

	// Log which which actor is accessing the repo.
	accesslog.Record(r.Context(), repo,
		log.String("treeish", treeish),
		log.String("format", format),
		log.Strings("path", pathspecs),
	)

	if err := git.CheckSpecArgSafety(treeish); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Logger.Error("gitserver.archive.CheckSpecArgSafety", log.Error(err))
		return
	}

	if repo == "" || format == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("gitserver.archive", log.String("error", "empty repo or format"))
		return
	}

	req := &protocol.ExecRequest{
		Repo: api.RepoName(repo),
		Args: []string{
			"archive",

			// Suppresses fatal error when the repo contains paths matching **/.git/** and instead
			// includes those files (to allow archiving invalid such repos). This is unexpected
			// behavior; the --worktree-attributes flag should merely let us specify a gitattributes
			// file that contains `**/.git/** export-ignore`, but it actually makes everything work as
			// desired. Tested by the "repo with .git dir" test case.
			"--worktree-attributes",

			"--format=" + format,
		},
	}

	if format == string(gitserver.ArchiveFormatZip) {
		// Compression level of 0 (no compression) seems to perform the
		// best overall on fast network links, but this has not been tuned
		// thoroughly.
		req.Args = append(req.Args, "-0")
	}

	req.Args = append(req.Args, treeish, "--")
	req.Args = append(req.Args, pathspecs...)

	s.execHTTP(w, r, req)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	logger := s.Logger.Scoped("handleSearch")
	tr, ctx := trace.New(r.Context(), "handleSearch")
	defer tr.End()

	// Decode the request
	protocol.RegisterGob()
	var args protocol.SearchRequest
	if err := gob.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		tr.SetError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var matchesBufMux sync.Mutex
	matchesBuf := streamhttp.NewJSONArrayBuf(8*1024, func(data []byte) error {
		tr.AddEvent("flushing data", attribute.Int("data.len", len(data)))
		return eventWriter.EventBytes("matches", data)
	})

	// Start a goroutine that periodically flushes the buffer
	var flusherWg conc.WaitGroup
	flusherCtx, flusherCancel := context.WithCancel(context.Background())
	defer flusherCancel()
	flusherWg.Go(func() {
		flushTicker := time.NewTicker(50 * time.Millisecond)
		defer flushTicker.Stop()

		for {
			select {
			case <-flushTicker.C:
				matchesBufMux.Lock()
				matchesBuf.Flush()
				matchesBufMux.Unlock()
			case <-flusherCtx.Done():
				return
			}
		}
	})

	// Create a callback that appends the match to the buffer
	var haveFlushed atomic.Bool
	onMatch := func(match *protocol.CommitMatch) error {
		matchesBufMux.Lock()
		defer matchesBufMux.Unlock()

		err := matchesBuf.Append(match)
		if err != nil {
			return err
		}

		// If we haven't sent any results yet, flush immediately
		if !haveFlushed.Load() {
			haveFlushed.Store(true)
			return matchesBuf.Flush()
		}

		return nil
	}

	// Run the search
	limitHit, searchErr := s.searchWithObservability(ctx, tr, &args, onMatch)
	if writeErr := eventWriter.Event("done", protocol.NewSearchEventDone(limitHit, searchErr)); writeErr != nil {
		if !errors.Is(writeErr, syscall.EPIPE) {
			logger.Error("failed to send done event", log.Error(writeErr))
		}
	}

	// Clean up the flusher goroutine, then do one final flush
	flusherCancel()
	flusherWg.Wait()
	matchesBuf.Flush()
}

func (s *Server) performGitLogCommand(ctx context.Context, repoCommit api.RepoCommit, format string) (output string, isRepoCloned bool, err error) {
	ctx, _, endObservation := s.operations.batchLogSingle.With(ctx, &err, observation.Args{
		Attrs: append(
			[]attribute.KeyValue{
				attribute.String("format", format),
			},
			repoCommit.Attrs()...,
		),
	})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Bool("isRepoCloned", isRepoCloned),
		}})
	}()

	dir := gitserverfs.RepoDirFromName(s.ReposDir, repoCommit.Repo)
	if !repoCloned(dir) {
		return "", false, nil
	}

	var buf bytes.Buffer

	commitId := string(repoCommit.CommitID)
	// make sure CommitID is not an arg
	if commitId[0] == '-' {
		return "", true, errors.New("commit ID starting with - is not allowed")
	}

	cmd := s.RecordingCommandFactory.Command(ctx, s.Logger, string(repoCommit.Repo), "git", "log", "-n", "1", "--name-only", format, commitId)
	dir.Set(cmd.Unwrap())
	cmd.Unwrap().Stdout = &buf

	if _, err := executil.RunCommand(ctx, cmd); err != nil {
		return "", true, err
	}

	return buf.String(), true, nil
}

func (s *Server) batchGitLogInstrumentedHandler(ctx context.Context, req protocol.BatchLogRequest) (resp protocol.BatchLogResponse, err error) {
	ctx, _, endObservation := s.operations.batchLog.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.String("results", fmt.Sprintf("%+v", resp.Results)),
		}})
	}()

	// Perform requests in each repository in the input batch. We perform these commands
	// concurrently, but only allow for so many commands to be in-flight at a time so that
	// we don't overwhelm a shard with either a large request or too many concurrent batch
	// requests.

	g, ctx := errgroup.WithContext(ctx)
	results := make([]protocol.BatchLogResult, len(req.RepoCommits))

	if s.GlobalBatchLogSemaphore == nil {
		return protocol.BatchLogResponse{}, errors.New("s.GlobalBatchLogSemaphore not initialized")
	}

	for i, repoCommit := range req.RepoCommits {
		// Avoid capture of loop variables
		i, repoCommit := i, repoCommit

		start := time.Now()
		if err := s.GlobalBatchLogSemaphore.Acquire(ctx, 1); err != nil {
			return resp, err
		}
		s.operations.batchLogSemaphoreWait.Observe(time.Since(start).Seconds())

		g.Go(func() error {
			defer s.GlobalBatchLogSemaphore.Release(1)

			output, isRepoCloned, gitLogErr := s.performGitLogCommand(ctx, repoCommit, req.Format)
			if gitLogErr == nil && !isRepoCloned {
				gitLogErr = errors.Newf("repo not found")
			}
			var errMessage string
			if gitLogErr != nil {
				errMessage = gitLogErr.Error()
			}

			// Concurrently write results to shared slice. This slice is already properly
			// sized, and each goroutine writes to a unique index exactly once. There should
			// be no data race conditions possible here.

			results[i] = protocol.BatchLogResult{
				RepoCommit:    repoCommit,
				CommandOutput: output,
				CommandError:  errMessage,
			}
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return
	}
	return protocol.BatchLogResponse{Results: results}, nil
}

func (s *Server) handleBatchLog(w http.ResponseWriter, r *http.Request) {
	// 🚨 SECURITY: Only allow POST requests.
	if strings.ToUpper(r.Method) != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	s.operations = s.ensureOperations()

	// Read request body
	var req protocol.BatchLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request parameters
	if len(req.RepoCommits) == 0 {
		// Early exit: implicitly writes 200 OK
		_ = json.NewEncoder(w).Encode(protocol.BatchLogResponse{Results: []protocol.BatchLogResult{}})
		return
	}
	if !strings.HasPrefix(req.Format, "--format=") {
		http.Error(w, "format parameter expected to be of the form `--format=<git log format>`", http.StatusUnprocessableEntity)
		return
	}

	// Handle unexpected error conditions
	resp, err := s.batchGitLogInstrumentedHandler(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write payload to client: implicitly writes 200 OK
	_ = json.NewEncoder(w).Encode(resp)
}

// ensureOperations returns the non-nil operations value supplied to this server
// via RegisterMetrics (when constructed as part of the gitserver binary), or
// constructs and memoizes a no-op operations value (for use in tests).
func (s *Server) ensureOperations() *operations {
	if s.operations == nil {
		s.operations = newOperations(s.ObservationCtx)
	}

	return s.operations
}

func (s *Server) handleExec(w http.ResponseWriter, r *http.Request) {
	// 🚨 SECURITY: Only allow POST requests.
	// See https://github.com/sourcegraph/security-issues/issues/213.
	if strings.ToUpper(r.Method) != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log which actor is accessing the repo.
	args := req.Args
	cmd := ""
	if len(req.Args) > 0 {
		cmd = req.Args[0]
		args = args[1:]
	}
	accesslog.Record(r.Context(), string(req.Repo),
		log.String("cmd", cmd),
		log.Strings("args", args),
	)

	s.execHTTP(w, r, &req)
}

var blockedCommandExecutedCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_gitserver_exec_blocked_command_received",
	Help: "Incremented each time a command not in the allowlist for gitserver is executed",
})

var ErrInvalidCommand = errors.New("invalid command")

type NotFoundError struct {
	Payload *protocol.NotFoundPayload
}

func (e *NotFoundError) Error() string { return "not found" }

type execStatus struct {
	ExitStatus int
	Stderr     string
	Err        error
}

// exec runs a git command. After the first write to w, it must not return an error.
// TODO(@camdencheek): once gRPC is the only consumer of this, do everything with errors
// because gRPC can handle trailing errors on a stream.
func (s *Server) exec(ctx context.Context, logger log.Logger, req *protocol.ExecRequest, userAgent string, w io.Writer) (execStatus, error) {
	// 🚨 SECURITY: Ensure that only commands in the allowed list are executed.
	// See https://github.com/sourcegraph/security-issues/issues/213.

	repoPath := string(protocol.NormalizeRepo(req.Repo))
	repoDir := filepath.Join(s.ReposDir, filepath.FromSlash(repoPath))

	if !gitdomain.IsAllowedGitCmd(logger, req.Args, repoDir) {
		blockedCommandExecutedCounter.Inc()
		return execStatus{}, ErrInvalidCommand
	}

	if !req.NoTimeout {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, executil.ShortGitCommandTimeout(req.Args))
		defer cancel()
	}

	start := time.Now()
	var cmdStart time.Time // set once we have ensured commit
	exitStatus := executil.UnsetExitStatus
	var stdoutN, stderrN int64
	var status string
	var execErr error
	ensureRevisionStatus := "noop"

	req.Repo = protocol.NormalizeRepo(req.Repo)
	repoName := req.Repo

	// Instrumentation
	{
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		args := strings.Join(req.Args, " ")

		var tr trace.Trace
		tr, ctx = trace.New(ctx, "exec."+cmd, repoName.Attr())
		tr.SetAttributes(
			attribute.String("args", args),
			attribute.String("ensure_revision", req.EnsureRevision),
		)
		logger = logger.WithTrace(trace.Context(ctx))

		execRunning.WithLabelValues(cmd).Inc()
		defer func() {
			tr.AddEvent(
				"done",
				attribute.String("status", status),
				attribute.Int64("stdout", stdoutN),
				attribute.Int64("stderr", stderrN),
				attribute.String("ensure_revision_status", ensureRevisionStatus),
			)
			tr.SetError(execErr)
			tr.End()

			duration := time.Since(start)
			execRunning.WithLabelValues(cmd).Dec()
			execDuration.WithLabelValues(cmd, status).Observe(duration.Seconds())

			var cmdDuration time.Duration
			var fetchDuration time.Duration
			if !cmdStart.IsZero() {
				cmdDuration = time.Since(cmdStart)
				fetchDuration = cmdStart.Sub(start)
			}

			isSlow := cmdDuration > shortGitCommandSlow(req.Args)
			isSlowFetch := fetchDuration > 10*time.Second
			if honey.Enabled() || traceLogs || isSlow || isSlowFetch {
				act := actor.FromContext(ctx)
				ev := honey.NewEvent("gitserver-exec")
				ev.SetSampleRate(honeySampleRate(cmd, act))
				ev.AddField("repo", repoName)
				ev.AddField("cmd", cmd)
				ev.AddField("args", args)
				ev.AddField("actor", act.UIDString())
				ev.AddField("ensure_revision", req.EnsureRevision)
				ev.AddField("ensure_revision_status", ensureRevisionStatus)
				ev.AddField("client", userAgent)
				ev.AddField("duration_ms", duration.Milliseconds())
				ev.AddField("stdin_size", len(req.Stdin))
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_status", exitStatus)
				ev.AddField("status", status)
				if execErr != nil {
					ev.AddField("error", execErr.Error())
				}
				if !cmdStart.IsZero() {
					ev.AddField("cmd_duration_ms", cmdDuration.Milliseconds())
					ev.AddField("fetch_duration_ms", fetchDuration.Milliseconds())
				}

				if traceID := trace.ID(ctx); traceID != "" {
					ev.AddField("traceID", traceID)
					ev.AddField("trace", trace.URL(traceID, conf.DefaultClient()))
				}

				if honey.Enabled() {
					_ = ev.Send()
				}

				if traceLogs {
					logger.Debug("TRACE gitserver exec", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
				if isSlow {
					logger.Warn("Long exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
				if isSlowFetch {
					logger.Warn("Slow fetch/clone for exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
			}
		}()
	}

	if notFoundPayload, cloned := s.maybeStartClone(ctx, logger, repoName); !cloned {
		if notFoundPayload.CloneInProgress {
			status = "clone-in-progress"
		} else {
			status = "repo-not-found"
		}

		return execStatus{}, &NotFoundError{notFoundPayload}
	}

	dir := gitserverfs.RepoDirFromName(s.ReposDir, repoName)
	if s.ensureRevision(ctx, repoName, req.EnsureRevision, dir) {
		ensureRevisionStatus = "fetched"
	}

	// Special-case `git rev-parse HEAD` requests. These are invoked by search queries for every repo in scope.
	// For searches over large repo sets (> 1k), this leads to too many child process execs, which can lead
	// to a persistent failure mode where every exec takes > 10s, which is disastrous for gitserver performance.
	if len(req.Args) == 2 && req.Args[0] == "rev-parse" && req.Args[1] == "HEAD" {
		if resolved, err := git.QuickRevParseHead(dir); err == nil && gitdomain.IsAbsoluteRevision(resolved) {
			_, _ = w.Write([]byte(resolved))
			return execStatus{}, nil
		}
	}

	// Special-case `git symbolic-ref HEAD` requests. These are invoked by resolvers determining the default branch of a repo.
	// For searches over large repo sets (> 1k), this leads to too many child process execs, which can lead
	// to a persistent failure mode where every exec takes > 10s, which is disastrous for gitserver performance.
	if len(req.Args) == 2 && req.Args[0] == "symbolic-ref" && req.Args[1] == "HEAD" {
		if resolved, err := git.QuickSymbolicRefHead(dir); err == nil {
			_, _ = w.Write([]byte(resolved))
			return execStatus{}, nil
		}
	}

	var stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &limitWriter{W: &stderrBuf, N: 1024}}

	cmdStart = time.Now()
	cmd := s.RecordingCommandFactory.Command(ctx, s.Logger, string(repoName), "git", req.Args...)
	dir.Set(cmd.Unwrap())
	cmd.Unwrap().Stdout = stdoutW
	cmd.Unwrap().Stderr = stderrW
	cmd.Unwrap().Stdin = bytes.NewReader(req.Stdin)

	exitStatus, execErr = executil.RunCommand(ctx, cmd)

	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()
	s.logIfCorrupt(ctx, repoName, dir, stderr)

	return execStatus{
		Err:        execErr,
		Stderr:     stderr,
		ExitStatus: exitStatus,
	}, nil
}

// execHTTP translates the results of an exec into the expected HTTP statuses and payloads
func (s *Server) execHTTP(w http.ResponseWriter, r *http.Request, req *protocol.ExecRequest) {
	logger := s.Logger.Scoped("exec").With(log.Strings("req.Args", req.Args))

	// Flush writes more aggressively than standard net/http so that clients
	// with a context deadline see as much partial response body as possible.
	if fw := newFlushingResponseWriter(logger, w); fw != nil {
		w = fw
		defer fw.Close()
	}

	ctx := r.Context()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache")

	w.Header().Set("Trailer", "X-Exec-Error")
	w.Header().Add("Trailer", "X-Exec-Exit-Status")
	w.Header().Add("Trailer", "X-Exec-Stderr")

	execStatus, err := s.exec(ctx, logger, req, r.UserAgent(), w)
	w.Header().Set("X-Exec-Error", errorString(execStatus.Err))
	w.Header().Set("X-Exec-Exit-Status", strconv.Itoa(execStatus.ExitStatus))
	w.Header().Set("X-Exec-Stderr", execStatus.Stderr)
	if err != nil {
		if v := (&NotFoundError{}); errors.As(err, &v) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(v.Payload)

		} else if errors.Is(err, ErrInvalidCommand) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid command"))

		} else {
			// If it's not a well-known error, send the error text
			// and a generic error code.
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}

func setLastFetched(ctx context.Context, db database.DB, shardID string, dir common.GitDir, name api.RepoName) error {
	lastFetched, err := repoLastFetched(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to get last fetched for %s", name)
	}

	lastChanged, err := repoLastChanged(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to get last changed for %s", name)
	}

	return db.GitserverRepos().SetLastFetched(ctx, name, database.GitserverFetchData{
		LastFetched: lastFetched,
		LastChanged: lastChanged,
		ShardID:     shardID,
	})
}

// setLastErrorNonFatal will set the last_error column for the repo in the gitserver table.
func (s *Server) setLastErrorNonFatal(ctx context.Context, name api.RepoName, err error) {
	var errString string
	if err != nil {
		errString = err.Error()
	}

	if err := s.DB.GitserverRepos().SetLastError(ctx, name, errString, s.Hostname); err != nil {
		s.Logger.Warn("Setting last error in DB", log.Error(err))
	}
}

func (s *Server) logIfCorrupt(ctx context.Context, repo api.RepoName, dir common.GitDir, stderr string) {
	if checkMaybeCorruptRepo(s.Logger, s.RecordingCommandFactory, repo, s.ReposDir, dir, stderr) {
		reason := stderr
		if err := s.DB.GitserverRepos().LogCorruption(ctx, repo, reason, s.Hostname); err != nil {
			s.Logger.Warn("failed to log repo corruption", log.String("repo", string(repo)), log.Error(err))
		}
	}
}

var (
	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate
	// that a repository's packfiles or commit objects might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/6676 for more
	// context.
	objectOrPackFileCorruptionRegex = lazyregexp.NewPOSIX(`^error: (Could not read|packfile) `)

	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate that
	// git's supplemental commit-graph might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/37872 for more
	// context.
	commitGraphCorruptionRegex = lazyregexp.NewPOSIX(`^fatal: commit-graph requires overflow generation data but has none`)
)

func checkMaybeCorruptRepo(logger log.Logger, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, reposDir string, dir common.GitDir, stderr string) bool {
	if !stdErrIndicatesCorruption(stderr) {
		return false
	}

	logger = logger.With(log.String("repo", string(repo)), log.String("dir", string(dir)))
	logger.Warn("marking repo for re-cloning due to stderr output indicating repo corruption",
		log.String("stderr", stderr))

	// We set a flag in the config for the cleanup janitor job to fix. The janitor
	// runs every minute.
	err := git.ConfigSet(rcf, reposDir, dir, gitConfigMaybeCorrupt, strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		logger.Error("failed to set maybeCorruptRepo config", log.Error(err))
	}

	return true
}

// stdErrIndicatesCorruption returns true if the provided stderr output from a git command indicates
// that there might be repository corruption.
func stdErrIndicatesCorruption(stderr string) bool {
	return objectOrPackFileCorruptionRegex.MatchString(stderr) || commitGraphCorruptionRegex.MatchString(stderr)
}

// testRepoCorrupter is used by tests to disrupt a cloned repository (e.g. deleting
// HEAD, zeroing it out, etc.)
var testRepoCorrupter func(ctx context.Context, tmpDir common.GitDir)

// cloneOptions specify optional behaviour for the cloneRepo function.
type CloneOptions struct {
	// Block will wait for the clone to finish before returning. If the clone
	// fails, the error will be returned. The passed in context is
	// respected. When not blocking the clone is done with a server background
	// context.
	Block bool

	// Overwrite will overwrite the existing clone.
	Overwrite bool
}

// CloneRepo performs a clone operation for the given repository. It is
// non-blocking by default.
func (s *Server) CloneRepo(ctx context.Context, repo api.RepoName, opts CloneOptions) (cloneProgress string, err error) {
	if isAlwaysCloningTest(repo) {
		return "This will never finish cloning", nil
	}

	dir := gitserverfs.RepoDirFromName(s.ReposDir, repo)

	// PERF: Before doing the network request to check if isCloneable, lets
	// ensure we are not already cloning.
	if progress, cloneInProgress := s.Locker.Status(dir); cloneInProgress {
		return progress, nil
	}

	// We always want to store whether there was an error cloning the repo, but only
	// after we checked if a clone is already in progress, otherwise we would race with
	// the actual running clone for the DB state of last_error.
	defer func() {
		// Use a different context in case we failed because the original context failed.
		s.setLastErrorNonFatal(s.ctx, repo, err)
	}()

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return "", errors.Wrap(err, "get VCS syncer")
	}

	// We may be attempting to clone a private repo so we need an internal actor.
	remoteURL, err := s.getRemoteURL(actor.WithInternalActor(ctx), repo)
	if err != nil {
		return "", err
	}

	// isCloneable causes a network request, so we limit the number that can
	// run at one time. We use a separate semaphore to cloning since these
	// checks being blocked by a few slow clones will lead to poor feedback to
	// users. We can defer since the rest of the function does not block this
	// goroutine.
	ctx, cancel, err := s.acquireCloneableLimiter(ctx)
	if err != nil {
		return "", err // err will be a context error
	}
	defer cancel()

	if err = s.RPSLimiter.Wait(ctx); err != nil {
		return "", err
	}

	if err := syncer.IsCloneable(ctx, repo, remoteURL); err != nil {
		redactedErr := urlredactor.New(remoteURL).Redact(err.Error())
		return "", errors.Errorf("error cloning repo: repo %s not cloneable: %s", repo, redactedErr)
	}

	// Mark this repo as currently being cloned. We have to check again if someone else isn't already
	// cloning since we released the lock. We released the lock since isCloneable is a potentially
	// slow operation.
	lock, ok := s.Locker.TryAcquire(dir, "starting clone")
	if !ok {
		// Someone else beat us to it
		status, _ := s.Locker.Status(dir)
		return status, nil
	}

	if s.skipCloneForTests {
		lock.Release()
		return "", nil
	}

	if opts.Block {
		ctx, cancel, err := s.acquireCloneLimiter(ctx)
		if err != nil {
			return "", err
		}
		defer cancel()

		// We are blocking, so use the passed in context.
		err = s.doClone(ctx, repo, dir, syncer, lock, remoteURL, opts)
		err = errors.Wrapf(err, "failed to clone %s", repo)
		return "", err
	}

	// We push the cloneJob to a queue and let the producer-consumer pipeline take over from this
	// point. See definitions of cloneJobProducer and cloneJobConsumer to understand how these jobs
	// are processed.
	s.CloneQueue.Push(&cloneJob{
		repo:      repo,
		dir:       dir,
		syncer:    syncer,
		lock:      lock,
		remoteURL: remoteURL,
		options:   opts,
	})

	return "", nil
}

func (s *Server) doClone(
	ctx context.Context,
	repo api.RepoName,
	dir common.GitDir,
	syncer vcssyncer.VCSSyncer,
	lock RepositoryLock,
	remoteURL *vcs.URL,
	opts CloneOptions,
) (err error) {
	logger := s.Logger.Scoped("doClone").With(log.String("repo", string(repo)))

	defer lock.Release()
	defer func() {
		if err != nil {
			repoCloneFailedCounter.Inc()
		}
	}()
	if err := s.RPSLimiter.Wait(ctx); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel()

	dstPath := string(dir)
	if !opts.Overwrite {
		// We clone to a temporary directory first, so avoid wasting resources
		// if the directory already exists.
		if _, err := os.Stat(dstPath); err == nil {
			return &os.PathError{
				Op:   "cloneRepo",
				Path: dstPath,
				Err:  os.ErrExist,
			}
		}
	}

	// We clone to a temporary location first to avoid having incomplete
	// clones in the repo tree. This also avoids leaving behind corrupt clones
	// if the clone is interrupted.
	tmpDir, err := gitserverfs.TempDir(s.ReposDir, "clone-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, ".git")

	// It may already be cloned
	if !repoCloned(dir) {
		if err := s.DB.GitserverRepos().SetCloneStatus(ctx, repo, types.CloneStatusCloning, s.Hostname); err != nil {
			s.Logger.Error("Setting clone status in DB", log.Error(err))
		}
	}
	defer func() {
		// Use a background context to ensure we still update the DB even if we time out
		if err := s.DB.GitserverRepos().SetCloneStatus(context.Background(), repo, cloneStatus(repoCloned(dir), false), s.Hostname); err != nil {
			s.Logger.Error("Setting clone status in DB", log.Error(err))
		}
	}()

	logger.Info("cloning repo", log.String("tmp", tmpDir), log.String("dst", dstPath))

	progressReader, progressWriter := io.Pipe()
	// We also capture the entire output in memory for the call to SetLastOutput
	// further down.
	// TODO: This might require a lot of memory depending on the amount of logs
	// produced, the ideal solution would be that readCloneProgress stores it in
	// chunks.
	output := &linebasedBufferedWriter{}
	eg := readCloneProgress(s.DB, logger, lock, io.TeeReader(progressReader, output), repo)

	cloneErr := syncer.Clone(ctx, repo, remoteURL, dir, tmpPath, progressWriter)
	progressWriter.Close()

	if err := eg.Wait(); err != nil {
		s.Logger.Error("reading clone progress", log.Error(err))
	}

	// best-effort update the output of the clone
	if err := s.DB.GitserverRepos().SetLastOutput(context.Background(), repo, output.String()); err != nil {
		s.Logger.Error("Setting last output in DB", log.Error(err))
	}

	if cloneErr != nil {
		// TODO: Should we really return the entire output here in an error?
		// It could be a super big error string.
		return errors.Wrapf(cloneErr, "clone failed. Output: %s", output.String())
	}

	if testRepoCorrupter != nil {
		testRepoCorrupter(ctx, common.GitDir(tmpPath))
	}

	if err := postRepoFetchActions(ctx, logger, s.DB, s.Hostname, s.RecordingCommandFactory, s.ReposDir, repo, common.GitDir(tmpPath), remoteURL, syncer); err != nil {
		return err
	}

	if opts.Overwrite {
		// remove the current repo by putting it into our temporary directory
		err := fileutil.RenameAndSync(dstPath, filepath.Join(filepath.Dir(tmpDir), "old"))
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to remove old clone")
		}
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return err
	}
	if err := fileutil.RenameAndSync(tmpPath, dstPath); err != nil {
		return err
	}

	logger.Info("repo cloned")
	repoClonedCounter.Inc()

	s.Perforce.EnqueueChangelistMappingJob(perforce.NewChangelistMappingJob(repo, dir))

	return nil
}

// linebasedBufferedWriter is an io.Writer that writes to a buffer.
// '\r' resets the write offset to the index after last '\n' in the buffer,
// or the beginning of the buffer if a '\n' has not been written yet.
//
// This exists to remove intermediate progress reports from "git clone
// --progress".
type linebasedBufferedWriter struct {
	// writeOffset is the offset in buf where the next write should begin.
	writeOffset int

	// afterLastNewline is the index after the last '\n' in buf
	// or 0 if there is no '\n' in buf.
	afterLastNewline int

	buf []byte
}

func (w *linebasedBufferedWriter) Write(p []byte) (n int, err error) {
	l := len(p)
	for {
		if len(p) == 0 {
			// If p ends in a '\r' we still want to include that in the buffer until it is overwritten.
			break
		}
		idx := bytes.IndexAny(p, "\r\n")
		if idx == -1 {
			w.buf = append(w.buf[:w.writeOffset], p...)
			w.writeOffset = len(w.buf)
			break
		}
		w.buf = append(w.buf[:w.writeOffset], p[:idx+1]...)
		switch p[idx] {
		case '\n':
			w.writeOffset = len(w.buf)
			w.afterLastNewline = len(w.buf)
			p = p[idx+1:]
		case '\r':
			// Record that our next write should overwrite the data after the most recent newline.
			// Don't slice it off immediately here, because we want to be able to return that output
			// until it is overwritten.
			w.writeOffset = w.afterLastNewline
			p = p[idx+1:]
		default:
			panic(fmt.Sprintf("unexpected char %q", p[idx]))
		}
	}
	return l, nil
}

// String returns the contents of the buffer as a string.
func (w *linebasedBufferedWriter) String() string {
	return string(w.buf)
}

// Bytes returns the contents of the buffer.
func (w *linebasedBufferedWriter) Bytes() []byte {
	return w.buf
}

func postRepoFetchActions(
	ctx context.Context,
	logger log.Logger,
	db database.DB,
	shardID string,
	rcf *wrexec.RecordingCommandFactory,
	reposDir string,
	repo api.RepoName,
	dir common.GitDir,
	remoteURL *vcs.URL,
	syncer vcssyncer.VCSSyncer,
) (errs error) {
	// Note: We use a multi error in this function to try to make as many of the
	// post repo fetch actions succeed.

	// We run setHEAD first, because other commands further down can fail when no
	// head exists.
	if err := setHEAD(ctx, logger, rcf, repo, dir, syncer, remoteURL); err != nil {
		errs = errors.Append(errs, errors.Wrapf(err, "failed to ensure HEAD exists for repo %q", repo))
	}

	if err := git.RemoveBadRefs(ctx, dir); err != nil {
		errs = errors.Append(errs, errors.Wrapf(err, "failed to remove bad refs for repo %q", repo))
	}

	if err := git.SetRepositoryType(rcf, reposDir, dir, syncer.Type()); err != nil {
		errs = errors.Append(errs, errors.Wrapf(err, "failed to set repository type for repo %q", repo))
	}

	if err := git.SetGitAttributes(dir); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "setting git attributes"))
	}

	if err := gitSetAutoGC(rcf, reposDir, dir); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "setting git gc mode"))
	}

	// Update the last-changed stamp on disk.
	if err := setLastChanged(logger, dir); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed to update last changed time"))
	}

	// Successfully updated, best-effort updating of db fetch state based on
	// disk state.
	if err := setLastFetched(ctx, db, shardID, dir, repo); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed setting last fetch in DB"))
	}

	// Successfully updated, best-effort calculation of the repo size.
	repoSizeBytes := gitserverfs.DirSize(dir.Path("."))
	if err := db.GitserverRepos().SetRepoSize(ctx, repo, repoSizeBytes, shardID); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed to set repo size"))
	}

	return errs
}

// readCloneProgress scans the reader and saves the most recent line of output
// as the lock status, writes to a log file if siteConfig.cloneProgressLog is
// enabled, and optionally to the database when the feature flag `clone-progress-logging`
// is enabled.
func readCloneProgress(db database.DB, logger log.Logger, lock RepositoryLock, pr io.Reader, repo api.RepoName) *errgroup.Group {
	// Use a background context to ensure we still update the DB even if we
	// time out. IE we intentionally don't take an input ctx.
	ctx := featureflag.WithFlags(context.Background(), db.FeatureFlags())
	enableExperimentalDBCloneProgress := featureflag.FromContext(ctx).GetBoolOr("clone-progress-logging", false)

	var logFile *os.File

	if conf.Get().CloneProgressLog {
		var err error
		logFile, err = os.CreateTemp("", "")
		if err != nil {
			logger.Warn("failed to create temporary clone log file", log.Error(err), log.String("repo", string(repo)))
		} else {
			logger.Info("logging clone output", log.String("file", logFile.Name()), log.String("repo", string(repo)))
			defer logFile.Close()
		}
	}

	dbWritesLimiter := rate.NewLimiter(rate.Limit(1.0), 1)
	scan := bufio.NewScanner(pr)
	scan.Split(scanCRLF)
	store := db.GitserverRepos()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for scan.Scan() {
			progress := scan.Text()
			lock.SetStatus(progress)

			if logFile != nil {
				// Failing to write here is non-fatal and we don't want to spam our logs if there
				// are issues
				_, _ = fmt.Fprintln(logFile, progress)
			}
			// Only write to the database persisted status if line indicates progress
			// which is recognized by presence of a '%'. We filter these writes not to waste
			// rate-limit tokens on log lines that would not be relevant to the user.
			if enableExperimentalDBCloneProgress {
				if strings.Contains(progress, "%") && dbWritesLimiter.Allow() {
					if err := store.SetCloningProgress(ctx, repo, progress); err != nil {
						logger.Error("error updating cloning progress in the db", log.Error(err))
					}
				}
			}
		}
		if err := scan.Err(); err != nil {
			return err
		}

		return nil
	})

	return eg
}

// scanCRLF is similar to bufio.ScanLines except it splits on both '\r' and '\n'
// and it does not return tokens that contain only whitespace.
func scanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	trim := func(data []byte) []byte {
		data = bytes.TrimSpace(data)
		if len(data) == 0 {
			// Don't pass back a token that is all whitespace.
			return nil
		}
		return data
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, trim(data[:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), trim(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

var (
	execRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_gitserver_exec_running",
		Help: "number of gitserver.GitCommand running concurrently.",
	}, []string{"cmd"})
	execDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_gitserver_exec_duration_seconds",
		Help:    "gitserver.GitCommand latencies in seconds.",
		Buckets: trace.UserLatencyBuckets,
	}, []string{"cmd", "status"})

	searchRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_search_running",
		Help: "number of gitserver.Search running concurrently.",
	})
	searchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_gitserver_search_duration_seconds",
		Help:    "gitserver.Search duration in seconds.",
		Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
	}, []string{"error"})
	searchLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "src_gitserver_search_latency_seconds",
		Help:    "gitserver.Search latency (time until first result is sent) in seconds.",
		Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
	})

	pendingClones = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_clone_queue",
		Help: "number of repos waiting to be cloned.",
	})
	lsRemoteQueue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_lsremote_queue",
		Help: "number of repos waiting to check existence on remote code host (git ls-remote).",
	})
	repoClonedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repo_cloned",
		Help: "number of successful git clones run",
	})
	repoCloneFailedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repo_cloned_failed",
		Help: "number of failed git clones",
	})
)

var headBranchPattern = lazyregexp.New(`HEAD branch: (.+?)\n`)

func (s *Server) doRepoUpdate(ctx context.Context, repo api.RepoName, revspec string) (err error) {
	tr, ctx := trace.New(ctx, "doRepoUpdate", repo.Attr())
	defer tr.EndWithErr(&err)

	s.repoUpdateLocksMu.Lock()
	l, ok := s.repoUpdateLocks[repo]
	if !ok {
		l = &locks{
			once: new(sync.Once),
			mu:   new(sync.Mutex),
		}
		s.repoUpdateLocks[repo] = l
	}
	once := l.once
	mu := l.mu
	s.repoUpdateLocksMu.Unlock()

	// doBackgroundRepoUpdate can block longer than our context deadline. done will
	// close when its done. We can return when either done is closed or our
	// deadline has passed.
	done := make(chan struct{})
	err = errors.New("another operation is already in progress")
	go func() {
		defer close(done)
		once.Do(func() {
			mu.Lock() // Prevent multiple updates in parallel. It works fine, but it wastes resources.
			defer mu.Unlock()

			s.repoUpdateLocksMu.Lock()
			l.once = new(sync.Once) // Make new requests wait for next update.
			s.repoUpdateLocksMu.Unlock()

			err = s.doBackgroundRepoUpdate(repo, revspec)
			if err != nil {
				// We don't want to spam our logs when the rate limiter has been set to block all
				// updates
				if !errors.Is(err, ratelimit.ErrBlockAll) {
					s.Logger.Error("performing background repo update", log.Error(err))
				}

				// The repo update might have failed due to the repo being corrupt
				var gitErr *common.GitCommandError
				if errors.As(err, &gitErr) {
					s.logIfCorrupt(ctx, repo, gitserverfs.RepoDirFromName(s.ReposDir, repo), gitErr.Output)
				}
			}
			s.setLastErrorNonFatal(s.ctx, repo, err)
		})
	}()

	select {
	case <-done:
		return errors.Wrapf(err, "repo %s:", repo)
	case <-ctx.Done():
		return ctx.Err()
	}
}

var doBackgroundRepoUpdateMock func(api.RepoName) error

func (s *Server) doBackgroundRepoUpdate(repo api.RepoName, revspec string) error {
	logger := s.Logger.Scoped("backgroundRepoUpdate").With(log.String("repo", string(repo)))

	if doBackgroundRepoUpdateMock != nil {
		return doBackgroundRepoUpdateMock(repo)
	}
	// background context.
	ctx, cancel1 := s.serverContext()
	defer cancel1()

	// ensure the background update doesn't hang forever
	ctx, cancel2 := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel2()

	// This background process should use our internal actor
	ctx = actor.WithInternalActor(ctx)

	ctx, cancel2, err := s.acquireCloneLimiter(ctx)
	if err != nil {
		return err
	}
	defer cancel2()

	if err = s.RPSLimiter.Wait(ctx); err != nil {
		return err
	}

	repo = protocol.NormalizeRepo(repo)
	dir := gitserverfs.RepoDirFromName(s.ReposDir, repo)

	remoteURL, err := s.getRemoteURL(ctx, repo)
	if err != nil {
		return errors.Wrap(err, "failed to determine Git remote URL")
	}

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return errors.Wrap(err, "get VCS syncer")
	}

	// drop temporary pack files after a fetch. this function won't
	// return until this fetch has completed or definitely-failed,
	// either way they can't still be in use. we don't care exactly
	// when the cleanup happens, just that it does.
	// TODO: Should be done in janitor.
	defer git.CleanTmpPackFiles(s.Logger, dir)

	output, err := syncer.Fetch(ctx, remoteURL, repo, dir, revspec)
	// TODO: Move the redaction also into the VCSSyncer layer here, to be in line
	// with what clone does.
	redactedOutput := urlredactor.New(remoteURL).Redact(string(output))
	// best-effort update the output of the fetch
	if err := s.DB.GitserverRepos().SetLastOutput(context.Background(), repo, redactedOutput); err != nil {
		s.Logger.Warn("Setting last output in DB", log.Error(err))
	}

	if err != nil {
		if output != nil {
			return errors.Wrapf(err, "failed to fetch repo %q with output %q", repo, redactedOutput)
		} else {
			return errors.Wrapf(err, "failed to fetch repo %q", repo)
		}
	}

	return postRepoFetchActions(ctx, logger, s.DB, s.Hostname, s.RecordingCommandFactory, s.ReposDir, repo, dir, remoteURL, syncer)
}

// setHEAD configures git repo defaults (such as what HEAD is) which are
// needed for git commands to work.
func setHEAD(ctx context.Context, logger log.Logger, rcf *wrexec.RecordingCommandFactory, repoName api.RepoName, dir common.GitDir, syncer vcssyncer.VCSSyncer, remoteURL *vcs.URL) error {
	// Verify that there is a HEAD file within the repo, and that it is of
	// non-zero length.
	if err := git.EnsureHEAD(dir); err != nil {
		logger.Error("failed to ensure HEAD exists", log.Error(err), log.String("repo", string(repoName)))
	}

	// Fallback to git's default branch name if git remote show fails.
	headBranch := "master"

	// try to fetch HEAD from origin
	cmd, err := syncer.RemoteShowCommand(ctx, remoteURL)
	if err != nil {
		return errors.Wrap(err, "get remote show command")
	}
	dir.Set(cmd)
	r := urlredactor.New(remoteURL)
	output, err := executil.RunRemoteGitCommand(ctx, rcf.WrapWithRepoName(ctx, logger, repoName, cmd).WithRedactorFunc(r.Redact), true)
	if err != nil {
		logger.Error("Failed to fetch remote info", log.Error(err), log.String("output", string(output)))
		return errors.Wrap(err, "failed to fetch remote info")
	}

	submatches := headBranchPattern.FindSubmatch(output)
	if len(submatches) == 2 {
		submatch := string(submatches[1])
		if submatch != "(unknown)" {
			headBranch = submatch
		}
	}

	// check if branch pointed to by HEAD exists
	cmd = exec.CommandContext(ctx, "git", "rev-parse", headBranch, "--")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		// branch does not exist, pick first branch
		cmd := exec.CommandContext(ctx, "git", "branch")
		dir.Set(cmd)
		output, err := cmd.Output()
		if err != nil {
			logger.Error("Failed to list branches", log.Error(err), log.String("output", string(output)))
			return errors.Wrap(err, "failed to list branches")
		}
		lines := strings.Split(string(output), "\n")
		branch := strings.TrimPrefix(strings.TrimPrefix(lines[0], "* "), "  ")
		if branch != "" {
			headBranch = branch
		}
	}

	// set HEAD
	cmd = exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD", "refs/heads/"+headBranch)
	dir.Set(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Error("Failed to set HEAD", log.Error(err), log.String("output", string(output)))
		return errors.Wrap(err, "Failed to set HEAD")
	}

	return nil
}

// setLastChanged discerns an approximate last-changed timestamp for a
// repository. This can be approximate; it's used to determine how often we
// should run `git fetch`, but is not relied on strongly. The basic plan
// is as follows: If a repository has never had a timestamp before, we
// guess that the right stamp is *probably* the timestamp of the most
// chronologically-recent commit. If there are no commits, we just use the
// current time because that's probably usually a temporary state.
//
// If a timestamp already exists, we want to update it if and only if
// the set of references (as determined by `git show-ref`) has changed.
//
// To accomplish this, we assert that the file `sg_refhash` in the git
// directory should, if it exists, contain a hash of the output of
// `git show-ref`, and have a timestamp of "the last time this changed",
// except that if we're creating that file for the first time, we set
// it to the timestamp of the top commit. We then compute the hash of
// the show-ref output, and store it in the file if and only if it's
// different from the current contents.
//
// If show-ref fails, we use rev-list to determine whether that's just
// an empty repository (not an error) or some kind of actual error
// that is possibly causing our data to be incorrect, which should
// be reported.
func setLastChanged(logger log.Logger, dir common.GitDir) error {
	hashFile := dir.Path("sg_refhash")

	hash, err := git.ComputeRefHash(dir)
	if err != nil {
		return errors.Wrapf(err, "computeRefHash failed for %s", dir)
	}

	var stamp time.Time
	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		// This is the first time we are calculating the hash. Give a more
		// approriate timestamp for sg_refhash than the current time.
		stamp = git.LatestCommitTimestamp(logger, dir)
	}

	_, err = fileutil.UpdateFileIfDifferent(hashFile, hash)
	if err != nil {
		return errors.Wrapf(err, "failed to update %s", hashFile)
	}

	// If stamp is non-zero we have a more approriate mtime.
	if !stamp.IsZero() {
		err = os.Chtimes(hashFile, stamp, stamp)
		if err != nil {
			return errors.Wrapf(err, "failed to set mtime to the lastest commit timestamp for %s", dir)
		}
	}

	return nil
}

// errorString returns the error string. If err is nil it returns the empty
// string.
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *Server) handleIsPerforcePathCloneable(w http.ResponseWriter, r *http.Request) {
	var req protocol.IsPerforcePathCloneableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.DepotPath == "" {
		http.Error(w, "no DepotPath given", http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.IsDepotPathCloneable(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd, req.DepotPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(protocol.IsPerforcePathCloneableResponse{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleCheckPerforceCredentials(w http.ResponseWriter, r *http.Request) {
	var req protocol.CheckPerforceCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(protocol.CheckPerforceCredentialsResponse{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleGetObject(w http.ResponseWriter, r *http.Request) {
	var req protocol.GetObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, errors.Wrap(err, "decoding body").Error(), http.StatusBadRequest)
		return
	}

	// Log which actor is accessing the repo.
	accesslog.Record(r.Context(), string(req.Repo), log.String("objectname", req.ObjectName))

	obj, err := git.GetObject(r.Context(), s.RecordingCommandFactory, s.ReposDir, req.Repo, req.ObjectName)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting object").Error(), http.StatusInternalServerError)
		return
	}

	resp := protocol.GetObjectResponse{
		Object: *obj,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handlePerforceUsers(w http.ResponseWriter, r *http.Request) {
	var req protocol.PerforceUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accesslog.Record(
		r.Context(),
		"<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
	)

	users, err := perforce.P4Users(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &protocol.PerforceUsersResponse{
		Users: make([]protocol.PerforceUser, 0, len(users)),
	}

	for _, user := range users {
		resp.Users = append(resp.Users, protocol.PerforceUser{
			Username: user.Username,
			Email:    user.Email,
		})
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handlePerforceProtectsForUser(w http.ResponseWriter, r *http.Request) {
	var req protocol.PerforceProtectsForUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accesslog.Record(
		r.Context(),
		"<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
	)

	protects, err := perforce.P4ProtectsForUser(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd, req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonProtects := make([]protocol.PerforceProtect, len(protects))
	for i, p := range protects {
		jsonProtects[i] = protocol.PerforceProtect{
			Level:       p.Level,
			EntityType:  p.EntityType,
			EntityName:  p.EntityName,
			Match:       p.Match,
			IsExclusion: p.IsExclusion,
			Host:        p.Host,
		}
	}

	resp := &protocol.PerforceProtectsForUserResponse{
		Protects: jsonProtects,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handlePerforceProtectsForDepot(w http.ResponseWriter, r *http.Request) {
	var req protocol.PerforceProtectsForDepotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accesslog.Record(
		r.Context(),
		"<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
	)

	protects, err := perforce.P4ProtectsForDepot(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd, req.Depot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonProtects := make([]protocol.PerforceProtect, len(protects))
	for i, p := range protects {
		jsonProtects[i] = protocol.PerforceProtect{
			Level:       p.Level,
			EntityType:  p.EntityType,
			EntityName:  p.EntityName,
			Match:       p.Match,
			IsExclusion: p.IsExclusion,
			Host:        p.Host,
		}
	}

	resp := &protocol.PerforceProtectsForDepotResponse{
		Protects: jsonProtects,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (s *Server) handlePerforceGroupMembers(w http.ResponseWriter, r *http.Request) {
	var req protocol.PerforceGroupMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accesslog.Record(
		r.Context(),
		"<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
	)

	members, err := perforce.P4GroupMembers(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd, req.Group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &protocol.PerforceGroupMembersResponse{
		Usernames: members,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleIsPerforceSuperUser(w http.ResponseWriter, r *http.Request) {
	var req protocol.IsPerforceSuperUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accesslog.Record(
		r.Context(),
		"<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
	)

	err = perforce.P4UserIsSuperUser(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		if err == perforce.ErrIsNotSuperUser {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &protocol.IsPerforceSuperUserResponse{}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handlePerforceGetChangelist(w http.ResponseWriter, r *http.Request) {
	var req protocol.PerforceGetChangelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accesslog.Record(
		r.Context(),
		"<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
	)

	changelist, err := perforce.GetChangelistByID(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd, req.ChangelistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &protocol.PerforceGetChangelistResponse{
		Changelist: protocol.PerforceChangelist{
			ID:           changelist.ID,
			CreationDate: changelist.CreationDate,
			State:        string(changelist.State),
			Author:       changelist.Author,
			Title:        changelist.Title,
			Message:      changelist.Message,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
