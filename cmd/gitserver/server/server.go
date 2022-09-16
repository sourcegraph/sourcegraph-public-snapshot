// Package server implements the gitserver service.
package server

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/internal/accesslog"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/adapters"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/search"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/mutablelimiter"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repotrackutil"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// tempDirName is the name used for the temporary directory under ReposDir.
const tempDirName = ".tmp"

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

// runCommandMock is set by tests. When non-nil it is run instead of
// runCommand
var runCommandMock func(context.Context, *exec.Cmd) (int, error)

// runCommand runs the command and returns the exit status. All clients of this function should set the context
// in cmd themselves, but we have to pass the context separately here for the sake of tracing.
func runCommand(ctx context.Context, cmd *exec.Cmd) (exitCode int, err error) {
	if runCommandMock != nil {
		return runCommandMock(ctx, cmd)
	}
	span, _ := ot.StartSpanFromContext(ctx, "runCommand")
	span.SetTag("path", cmd.Path)
	span.SetTag("args", cmd.Args)
	span.SetTag("dir", cmd.Dir)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
			span.SetTag("exitCode", exitCode)
		}
		span.Finish()
	}()

	err = cmd.Run()
	exitStatus := -10810         // sentinel value to indicate not set
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return exitStatus, err
}

// runCommandGraceful runs the command and returns the exit status. If the
// supplied context is cancelled we attempt to send SIGINT to the command to
// allow it to gracefully shutdown. All clients of this function should pass in a
// command *without* a context.
func runCommandGraceful(ctx context.Context, logger log.Logger, cmd *exec.Cmd) (exitCode int, err error) {
	span, _ := ot.StartSpanFromContext(ctx, "runCommandGraceful")
	span.SetTag("path", cmd.Path)
	span.SetTag("args", cmd.Args)
	span.SetTag("dir", cmd.Dir)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
			span.SetTag("exitCode", exitCode)
		}
		span.Finish()
	}()

	exitCode = -10810 // sentinel value to indicate not set
	err = cmd.Start()
	if err != nil {
		return exitCode, err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		err = cmd.Wait()
		if err != nil {
			logger.Error("running command", log.Error(err))
		}
	}()

	// Wait for command to exit or context to be done
	select {
	case <-ctx.Done():
		logger.Debug("context cancelled, sending SIGINT")
		// Attempt to send SIGINT
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			logger.Warn("Sending SIGINT to command", log.Error(err))
			if err := cmd.Process.Kill(); err != nil {
				logger.Warn("killing process", log.Error(err))
			}
			return exitCode, err
		}
		// Now, continue waiting for command for up to two seconds before killing it
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-done:
			logger.Debug("process exited after SIGINT sent")
			timer.Stop()
			if err == nil {
				exitCode = 0
			}
		case <-timer.C:
			logger.Debug("timed out, killing process")
			if err := cmd.Process.Kill(); err != nil {
				logger.Warn("killing process", log.Error(err))
			}
			logger.Debug("process killed, waiting for done")
			// Wait again to ensure we can access cmd.ProcessState below
			<-done
		}

		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		err = ctx.Err()
		return exitCode, err
	case <-done:
		// Happy path, command exits
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode = exitError.ExitCode()
	}
	if err == nil {
		exitCode = 0
	}
	return exitCode, err
}

// cloneJob abstracts away a repo and necessary metadata to clone it. In the future it may be
// possible to simplify this, but to do that, doClone will need to do a lot less than it does at the
// moment.
type cloneJob struct {
	repo   api.RepoName
	dir    GitDir
	syncer VCSSyncer

	// TODO: cloneJobConsumer should acquire a new lock. We are trying to keep the changes simple
	// for the time being. When we start using the new approach of using long lived goroutines for
	// cloning we will refactor doClone to acquire a new lock.
	lock *RepositoryLock

	remoteURL *vcs.URL
	options   *cloneOptions
}

// cloneQueue is a threadsafe list.List of cloneJobs that functions as a queue in practice.
type cloneQueue struct {
	mu   sync.Mutex
	jobs *list.List

	cmu  sync.Mutex
	cond *sync.Cond
}

// push will queue the cloneJob to the end of the queue.
func (c *cloneQueue) push(cj *cloneJob) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.jobs.PushBack(cj)
	c.cond.Signal()
}

// pop will return the next cloneJob. If there's no next job available, it returns nil.
func (c *cloneQueue) pop() *cloneJob {
	c.mu.Lock()
	defer c.mu.Unlock()

	next := c.jobs.Front()
	if next == nil {
		return nil
	}

	return c.jobs.Remove(next).(*cloneJob)
}

func (c *cloneQueue) empty() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.jobs.Len() == 0
}

// NewCloneQueue initializes a new cloneQueue.
func NewCloneQueue(jobs *list.List) *cloneQueue {
	cq := cloneQueue{jobs: jobs}
	cq.cond = sync.NewCond(&cq.cmu)

	return &cq
}

// Server is a gitserver server.
type Server struct {
	// Logger should be used for all logging and logger creation.
	Logger log.Logger

	// ReposDir is the path to the base directory for gitserver storage.
	ReposDir string

	// DesiredPercentFree is the desired percentage of disk space to keep free.
	DesiredPercentFree int

	// DiskSizer tells how much disk is free and how large the disk is.
	DiskSizer DiskSizer

	// GetRemoteURLFunc is a function which returns the remote URL for a
	// repository. This is used when cloning or fetching a repository. In
	// production this will speak to the database to look up the clone URL. In
	// tests this is usually set to clone a local repository or intentionally
	// error.
	//
	// Note: internal uses should call getRemoteURL which will handle
	// GetRemoteURLFunc being nil.
	GetRemoteURLFunc func(context.Context, api.RepoName) (string, error)

	// GetVCSSyncer is a function which returns the VCS syncer for a repository.
	// This is used when cloning or fetching a repository. In production this will
	// speak to the database to determine the code host type. In tests this is
	// usually set to return a GitRepoSyncer.
	GetVCSSyncer func(context.Context, api.RepoName) (VCSSyncer, error)

	// Hostname is how we identify this instance of gitserver. Generally it is the
	// actual hostname but can also be overridden by the HOSTNAME environment variable.
	Hostname string

	// shared db handle
	DB database.DB

	// CloneQueue is a threadsafe queue used by DoBackgroundClones to process incoming clone
	// requests asynchronously.
	CloneQueue *cloneQueue

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

	locker *RepositoryLocker

	// cloneLimiter and cloneableLimiter limits the number of concurrent
	// clones and ls-remotes respectively. Use s.acquireCloneLimiter() and
	// s.acquireClonableLimiter() instead of using these directly.
	cloneLimiter     *mutablelimiter.Limiter
	cloneableLimiter *mutablelimiter.Limiter

	// rpsLimiter limits the remote code host git operations done per second
	// per gitserver instance
	rpsLimiter *ratelimit.InstrumentedLimiter

	repoUpdateLocksMu sync.Mutex // protects the map below and also updates to locks.once
	repoUpdateLocks   map[api.RepoName]*locks

	// GlobalBatchLogSemaphore is a semaphore shared between all requests to ensure that a
	// maximum number of Git subprocesses are active for all /batch-log requests combined.
	GlobalBatchLogSemaphore *semaphore.Weighted

	// operations provide uniform observability via internal/observation. This value is
	// set by RegisterMetrics when compiled as part of the gitserver binary. The server
	// method ensureOperations should be used in all references to avoid a nil pointer
	// dereferencs.
	operations *operations
}

type locks struct {
	once *sync.Once  // consolidates multiple waiting updates
	mu   *sync.Mutex // prevents updates running in parallel
}

// shortGitCommandTimeout returns the timeout for git commands that should not
// take a long time. Some commands such as "git archive" are allowed more time
// than "git rev-parse", so this will return an appropriate timeout given the
// command.
func shortGitCommandTimeout(args []string) time.Duration {
	if len(args) < 1 {
		return time.Minute
	}
	switch args[0] {
	case "archive":
		// This is a long time, but this never blocks a user request for this
		// long. Even repos that are not that large can take a long time, for
		// example a search over all repos in an organization may have several
		// large repos. All of those repos will be competing for IO => we need
		// a larger timeout.
		return conf.GitLongCommandTimeout()

	case "ls-remote":
		return 30 * time.Second

	default:
		return time.Minute
	}
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

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.locker = &RepositoryLocker{}
	s.repoUpdateLocks = make(map[api.RepoName]*locks)

	// GitMaxConcurrentClones controls the maximum number of clones that
	// can happen at once on a single gitserver.
	// Used to prevent throttle limits from a code host. Defaults to 5.
	//
	// The new repo-updater scheduler enforces the rate limit across all gitserver,
	// so ideally this logic could be removed here; however, ensureRevision can also
	// cause an update to happen and it is called on every exec command.
	maxConcurrentClones := conf.GitMaxConcurrentClones()
	s.cloneLimiter = mutablelimiter.New(maxConcurrentClones)
	s.cloneableLimiter = mutablelimiter.New(maxConcurrentClones)
	conf.Watch(func() {
		limit := conf.GitMaxConcurrentClones()
		s.cloneLimiter.SetLimit(limit)
		s.cloneableLimiter.SetLimit(limit)
	})

	s.rpsLimiter = ratelimit.NewInstrumentedLimiter("RpsLimiter", rate.NewLimiter(rate.Inf, 10))
	setRPSLimiter := func() {
		if maxRequestsPerSecond := conf.GitMaxCodehostRequestsPerSecond(); maxRequestsPerSecond == -1 {
			// As a special case, -1 means no limiting
			s.rpsLimiter.SetLimit(rate.Inf)
			s.rpsLimiter.SetBurst(10)
		} else if maxRequestsPerSecond == 0 {
			// A limiter with zero limit but a non-zero burst is not rejecting all events
			// because the bucket is initially full with N tokens and refilled N tokens
			// every second, where N is the burst size. See
			// https://github.com/golang/go/issues/18763 for details.
			s.rpsLimiter.SetLimit(0)
			s.rpsLimiter.SetBurst(0)
		} else {
			s.rpsLimiter.SetLimit(rate.Limit(maxRequestsPerSecond))
			s.rpsLimiter.SetBurst(10)
		}
	}
	conf.Watch(func() {
		setRPSLimiter()
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/archive", trace.WithRouteName("archive", accesslog.HTTPMiddleware(
		s.Logger.Scoped("archive.accesslog", "archive endpoint access log"),
		conf.DefaultClient(),
		s.handleArchive,
	)))
	mux.HandleFunc("/exec", trace.WithRouteName("exec", accesslog.HTTPMiddleware(
		s.Logger.Scoped("exec.accesslog", "exec endpoint access log"),
		conf.DefaultClient(),
		s.handleExec,
	)))
	mux.HandleFunc("/search", trace.WithRouteName("search", s.handleSearch))
	mux.HandleFunc("/batch-log", trace.WithRouteName("batch-log", s.handleBatchLog))
	mux.HandleFunc("/p4-exec", trace.WithRouteName("p4-exec", accesslog.HTTPMiddleware(
		s.Logger.Scoped("p4-exec.accesslog", "p4-exec endpoint access log"),
		conf.DefaultClient(),
		s.handleP4Exec,
	)))
	mux.HandleFunc("/list-gitolite", trace.WithRouteName("list-gitolite", s.handleListGitolite))
	mux.HandleFunc("/is-repo-cloneable", trace.WithRouteName("is-repo-cloneable", s.handleIsRepoCloneable))
	mux.HandleFunc("/repos-stats", trace.WithRouteName("repos-stats", s.handleReposStats))
	mux.HandleFunc("/repo-clone-progress", trace.WithRouteName("repo-clone-progress", s.handleRepoCloneProgress))
	mux.HandleFunc("/delete", trace.WithRouteName("delete", s.handleRepoDelete))
	mux.HandleFunc("/repo-update", trace.WithRouteName("repo-update", s.handleRepoUpdate))
	mux.HandleFunc("/create-commit-from-patch", trace.WithRouteName("create-commit-from-patch", s.handleCreateCommitFromPatch))
	mux.HandleFunc("/ping", trace.WithRouteName("ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	mux.HandleFunc("/git/", trace.WithRouteName("git", accesslog.HTTPMiddleware(
		s.Logger.Scoped("git.accesslog", "git endpoint access log"),
		conf.DefaultClient(),
		func(rw http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/git", s.gitServiceHandler()).ServeHTTP(rw, r)
		},
	)))

	// Migration to hexagonal architecture starting here:

	gitAdapter := &adapters.Git{
		ReposDir: s.ReposDir,
	}
	getObjectService := gitdomain.GetObjectService{
		RevParse:      gitAdapter.RevParse,
		GetObjectType: gitAdapter.GetObjectType,
	}
	getObjectFunc := gitdomain.GetObjectFunc(func(ctx context.Context, repo api.RepoName, objectName string) (*gitdomain.GitObject, error) {
		// Tracing is server concern, so add it here. Once generics lands we should be
		// able to create some simple wrappers
		span, ctx := ot.StartSpanFromContext(ctx, "Git: GetObject")
		span.SetTag("objectName", objectName)
		defer span.Finish()
		return getObjectService.GetObject(ctx, repo, objectName)
	})

	mux.HandleFunc("/commands/get-object", trace.WithRouteName("commands/get-object",
		accesslog.HTTPMiddleware(
			s.Logger.Scoped("commands/get-object.accesslog", "commands/get-object endpoint access log"),
			conf.DefaultClient(),
			handleGetObject(s.Logger.Scoped("commands/get-object", "handles get object"), getObjectFunc),
		)))

	return mux
}

// Janitor does clean up tasks over s.ReposDir and is expected to run in a
// background goroutine.
func (s *Server) Janitor(interval time.Duration) {
	for {
		gitserverAddrs := currentGitserverAddresses()
		s.cleanupRepos(gitserverAddrs)
		time.Sleep(interval)
	}
}

// SyncRepoState syncs state on disk to the database for all repos and is
// expected to run in a background goroutine. We perform a full sync if the known
// gitserver addresses has changed since the last run. Otherwise, we only sync
// repos that have not yet been assigned a shard.
func (s *Server) SyncRepoState(interval time.Duration, batchSize, perSecond int) {
	var previousAddrs string
	var previousPinned string
	for {
		gitServerAddrs := currentGitserverAddresses()
		addrs := gitServerAddrs.Addresses
		// We turn addrs into a string here for easy comparison and storage of previous
		// addresses since we'd need to take a copy of the slice anyway.
		currentAddrs := strings.Join(addrs, ",")
		fullSync := currentAddrs != previousAddrs
		previousAddrs = currentAddrs

		// We turn PinnedServers into a string here for easy comparison and storage
		// of previous pins.
		pinnedServerPairs := make([]string, 0, len(gitServerAddrs.PinnedServers))
		for k, v := range gitServerAddrs.PinnedServers {
			pinnedServerPairs = append(pinnedServerPairs, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(pinnedServerPairs)
		currentPinned := strings.Join(pinnedServerPairs, ",")
		fullSync = fullSync || currentPinned != previousPinned
		previousPinned = currentPinned

		if err := s.syncRepoState(gitServerAddrs, batchSize, perSecond, fullSync); err != nil {
			s.Logger.Error("Syncing repo state", log.Error(err))
		}

		time.Sleep(interval)
	}
}

func (s *Server) addrForRepo(ctx context.Context, repoName api.RepoName, gitServerAddrs gitserver.GitServerAddresses) (string, error) {
	return gitserver.AddrForRepo(ctx, filepath.Base(os.Args[0]), s.DB, repoName, gitServerAddrs)
}

func currentGitserverAddresses() gitserver.GitServerAddresses {
	cfg := conf.Get()
	gitServerAddrs := gitserver.GitServerAddresses{
		Addresses: cfg.ServiceConnectionConfig.GitServers,
	}
	if cfg.ExperimentalFeatures != nil {
		gitServerAddrs.PinnedServers = cfg.ExperimentalFeatures.GitServerPinnedRepos
	}

	return gitServerAddrs
}

// StartClonePipeline clones repos asynchronously. It creates a producer-consumer
// pipeline.
func (s *Server) StartClonePipeline(ctx context.Context) {
	jobs := make(chan *cloneJob)

	go s.cloneJobConsumer(ctx, jobs)
	go s.cloneJobProducer(ctx, jobs)
}

func (s *Server) cloneJobProducer(ctx context.Context, jobs chan<- *cloneJob) {
	defer close(jobs)

	for {
		// Acquire the cond mutex lock and wait for a signal if the queue is empty.
		s.CloneQueue.cmu.Lock()
		if s.CloneQueue.empty() {
			s.CloneQueue.cond.Wait()
		}

		// The queue is not empty and we have a job to process! But don't forget to unlock the cond
		// mutex here as we don't need to hold the lock beyond this point for now.
		s.CloneQueue.cmu.Unlock()

		// Keep popping from the queue until the queue is empty again, in which case we start all
		// over again from the top.
		for {
			job := s.CloneQueue.pop()
			if job == nil {
				break
			}

			select {
			case jobs <- job:
			case <-ctx.Done():
				s.Logger.Error("cloneJobProducer: ", log.Error(ctx.Err()))
				return
			}
		}
	}
}

func (s *Server) cloneJobConsumer(ctx context.Context, jobs <-chan *cloneJob) {
	logger := s.Logger.Scoped("cloneJobConsumer", "process clone jobs")

	for j := range jobs {
		logger := logger.With(log.String("job.repo", string(j.repo)))

		select {
		case <-ctx.Done():
			logger.Error("context done", log.Error(ctx.Err()))
			return
		default:
		}

		ctx, cancel, err := s.acquireCloneLimiter(ctx)
		if err != nil {
			logger.Error("acquireCloneLimiter", log.Error(err))
			continue
		}

		go func(job *cloneJob) {
			defer cancel()

			err := s.doClone(ctx, job.repo, job.dir, job.syncer, job.lock, job.remoteURL, job.options)
			if err != nil {
				logger.Error("failed to clone repo", log.Error(err))
			}
			// Use a different context in case we failed because the original context failed.
			s.setLastErrorNonFatal(s.ctx, job.repo, err)
		}(j)
	}
}

// hostnameMatch checks whether the hostname matches the given address.
// If we don't find an exact match, we look at the initial prefix.
func (s *Server) hostnameMatch(addr string) bool {
	if !strings.HasPrefix(addr, s.Hostname) {
		return false
	}
	if addr == s.Hostname {
		return true
	}
	// We know that s.Hostname is shorter than addr so we can safely check the next
	// char
	next := addr[len(s.Hostname)]
	return next == '.' || next == ':'
}

var (
	repoSyncStateCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repo_sync_state_counter",
		Help: "Incremented each time we check the state of repo",
	}, []string{"type"})
	repoStateUpsertCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repo_sync_state_upsert_counter",
		Help: "Incremented each time we upsert repo state in the database",
	}, []string{"success"})
	wrongShardReposTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_repo_wrong_shard",
		Help: "The number of repos that are on disk on the wrong shard",
	})
	wrongShardReposSizeTotalBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_repo_wrong_shard_bytes",
		Help: "Size (in bytes) of repos that are on disk on the wrong shard",
	})
	wrongShardReposDeletedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repo_wrong_shard_deleted",
		Help: "The number of repos on the wrong shard that we deleted",
	})
)

func (s *Server) syncRepoState(gitServerAddrs gitserver.GitServerAddresses, batchSize, perSecond int, fullSync bool) error {
	s.Logger.Info("starting syncRepoState", log.Bool("fullSync", fullSync))
	addrs := gitServerAddrs.Addresses

	// When fullSync is true we'll scan all repos in the database and ensure we set
	// their clone state and assign any that belong to this shard with the correct
	// shard_id.
	//
	// When fullSync is false, we assume that we only need to check repos that have
	// not yet had their shard_id allocated.

	// Sanity check our host exists in addrs before starting any work
	var found bool
	for _, a := range addrs {
		if s.hostnameMatch(a) {
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf("gitserver hostname, %q, not found in list", s.Hostname)
	}

	ctx := s.ctx
	store := s.DB.GitserverRepos()

	// The rate limit should be enforced across all instances
	perSecond = perSecond / len(addrs)
	if perSecond < 0 {
		perSecond = 1
	}
	limiter := ratelimit.NewInstrumentedLimiter("SyncRepoState", rate.NewLimiter(rate.Limit(perSecond), perSecond))

	// The rate limiter doesn't allow writes that are larger than the burst size
	// which we've set to perSecond.
	if batchSize > perSecond {
		batchSize = perSecond
	}

	batch := make([]*types.GitserverRepo, 0)

	writeBatch := func() {
		if len(batch) == 0 {
			return
		}
		// We always clear the batch
		defer func() {
			batch = batch[0:0]
		}()
		err := limiter.WaitN(ctx, len(batch))
		if err != nil {
			s.Logger.Error("Waiting for rate limiter", log.Error(err))
			return
		}

		if err := store.Update(ctx, batch...); err != nil {
			repoStateUpsertCounter.WithLabelValues("false").Add(float64(len(batch)))
			s.Logger.Error("Updating GitserverRepos", log.Error(err))
			return
		}
		repoStateUpsertCounter.WithLabelValues("true").Add(float64(len(batch)))
	}

	options := database.IterateRepoGitserverStatusOptions{
		// We also want to include deleted repos as they may still be cloned on disk
		IncludeDeleted: true,
	}
	if !fullSync {
		options.OnlyWithoutShard = true
	}
	err := store.IterateRepoGitserverStatus(ctx, options, func(repo types.RepoGitserverStatus) error {
		repoSyncStateCounter.WithLabelValues("check").Inc()

		// We may have a deleted repo, we need to extract the original name both to
		// ensure that the shard check is correct and also so that we can find the
		// directory.
		repo.Name = api.UndeletedRepoName(repo.Name)

		// Ensure we're only dealing with repos we are responsible for.
		addr, err := s.addrForRepo(ctx, repo.Name, gitServerAddrs)
		if err != nil {
			return err
		}
		if !s.hostnameMatch(addr) {
			repoSyncStateCounter.WithLabelValues("other_shard").Inc()
			return nil
		}
		repoSyncStateCounter.WithLabelValues("this_shard").Inc()

		dir := s.dir(repo.Name)
		cloned := repoCloned(dir)
		_, cloning := s.locker.Status(dir)

		var shouldUpdate bool
		if repo.ShardID != s.Hostname {
			repo.ShardID = s.Hostname
			shouldUpdate = true
		}
		cloneStatus := cloneStatus(cloned, cloning)
		if repo.CloneStatus != cloneStatus {
			repo.CloneStatus = cloneStatus
			shouldUpdate = true
		}

		if !shouldUpdate {
			return nil
		}

		batch = append(batch, repo.GitserverRepo)

		if len(batch) >= batchSize {
			writeBatch()
		}

		return nil
	})

	// Attempt final write
	writeBatch()

	return err
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
	if s.GetRemoteURLFunc == nil {
		return nil, errors.New("gitserver GetRemoteURLFunc is unset")
	}

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

// queryCloneLimiter reports the capacity and length of the clone limiter's queue
func (s *Server) queryCloneLimiter() (cap, len int) {
	return s.cloneLimiter.GetLimit()
}

func (s *Server) acquireCloneableLimiter(ctx context.Context) (context.Context, context.CancelFunc, error) {
	lsRemoteQueue.Inc()
	defer lsRemoteQueue.Dec()
	return s.cloneableLimiter.Acquire(ctx)
}

// tempDir is a wrapper around os.MkdirTemp, but using the server's
// temporary directory filepath.Join(s.ReposDir, tempDirName).
//
// This directory is cleaned up by gitserver and will be ignored by repository
// listing operations.
func (s *Server) tempDir(prefix string) (name string, err error) {
	dir := filepath.Join(s.ReposDir, tempDirName)

	// Create tmpdir directory if doesn't exist yet.
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	return os.MkdirTemp(dir, prefix)
}

func (s *Server) ignorePath(path string) bool {
	// We ignore any path which starts with .tmp in ReposDir
	if filepath.Dir(path) != s.ReposDir {
		return false
	}
	return strings.HasPrefix(filepath.Base(path), tempDirName)
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

	var syncer VCSSyncer
	// We use an internal actor here as the repo may be private. It is safe since all
	// we return is a bool indicating whether the repo is cloneable or not. Perhaps
	// the only things that could leak here is whether a private repo exists although
	// the endpoint is only available internally so it's low risk.
	remoteURL, err := s.getRemoteURL(actor.WithInternalActor(r.Context()), req.Repo)
	if err != nil {
		// We use this endpoint to verify if a repo exists without consuming
		// API rate limit, since many users visit private or bogus repos,
		// so we deduce the unauthenticated clone URL from the repo name.
		remoteURL, _ = vcs.ParseURL("https://" + string(req.Repo) + ".git")

		// At this point we are assuming it's a git repo
		syncer = &GitRepoSyncer{}
	} else {
		syncer, err = s.GetVCSSyncer(r.Context(), req.Repo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var resp protocol.IsRepoCloneableResponse
	if err := syncer.IsCloneable(r.Context(), remoteURL); err == nil {
		resp = protocol.IsRepoCloneableResponse{Cloneable: true}
	} else {
		resp = protocol.IsRepoCloneableResponse{
			Cloneable: false,
			Reason:    err.Error(),
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleRepoUpdate is a synchronous (waits for update to complete or
// time out) method so it can yield errors. Updates are not
// unconditional; we debounce them based on the provided
// interval, to avoid spam.
func (s *Server) handleRepoUpdate(w http.ResponseWriter, r *http.Request) {
	logger := s.Logger.Scoped("handleRepoUpdate", "synchronous http handler for repo updates")
	var req protocol.RepoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var resp protocol.RepoUpdateResponse
	req.Repo = protocol.NormalizeRepo(req.Repo)
	dir := s.dir(req.Repo)

	// despite the existence of a context on the request, we don't want to
	// cancel the git commands partway through if the request terminates.
	ctx, cancel1 := s.serverContext()
	defer cancel1()
	ctx, cancel2 := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel2()
	if !repoCloned(dir) && !s.skipCloneForTests {
		// We do not need to check if req.CloneFromShard is non-zero here since that has no effect on
		// the code path at this point. Since the repo is already not cloned at this point, either
		// this request was received for a repo migration or a regular clone - for both of which we
		// want to go ahead and clone the repo. The responsibility of figuring out where to clone
		// the repo from (upstream URL of the external service or the gitserver instance) lies with
		// the implementation details of cloneRepo.
		_, err := s.cloneRepo(ctx, req.Repo, &cloneOptions{Block: true, CloneFromShard: req.CloneFromShard})
		if err != nil {
			logger.Warn("error cloning repo", log.String("repo", string(req.Repo)), log.Error(err))
			resp.Error = err.Error()
		}
	} else {
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
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
	var (
		logger    = s.Logger.Scoped("handleArchive", "http handler for repo archive")
		q         = r.URL.Query()
		treeish   = q.Get("treeish")
		repo      = q.Get("repo")
		format    = q.Get("format")
		pathspecs = q["path"]
	)

	// Log which which actor is accessing the repo.
	accesslog.Record(r.Context(), repo, map[string]string{
		"treeish": treeish,
		"format":  format,
		"path":    strings.Join(pathspecs, ","),
	})

	if err := checkSpecArgSafety(treeish); err != nil {
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

	s.exec(w, r, req)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	logger := s.Logger.Scoped("handleSearch", "http handler for search")
	tr, ctx := trace.New(r.Context(), "search", "")
	defer tr.Finish()

	// Decode the request
	protocol.RegisterGob()
	var args protocol.SearchRequest
	if err := gob.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tr.SetAttributes(
		attribute.String("repo", string(args.Repo)),
		attribute.Bool("include_diff", args.IncludeDiff),
		attribute.String("query", args.Query.String()),
		attribute.Int("limit", args.Limit),
		attribute.Bool("include_modified_files", args.IncludeModifiedFiles),
	)

	searchStart := time.Now()
	searchRunning.Inc()
	defer searchRunning.Dec()

	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		tr.SetError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var latencyOnce sync.Once
	matchesBuf := streamhttp.NewJSONArrayBuf(8*1024, func(data []byte) error {
		tr.AddEvent("flushing data", attribute.Int("data.len", len(data)))
		latencyOnce.Do(func() {
			searchLatency.Observe(time.Since(searchStart).Seconds())
		})
		return eventWriter.EventBytes("matches", data)
	})

	// Run the search
	limitHit, searchErr := s.search(ctx, &args, matchesBuf)
	if writeErr := eventWriter.Event("done", protocol.NewSearchEventDone(limitHit, searchErr)); writeErr != nil {
		logger.Error("failed to send done event", log.Error(writeErr))
	}
	tr.AddEvent("done", attribute.Bool("limit_hit", limitHit))
	tr.SetError(searchErr)
	searchDuration.
		WithLabelValues(strconv.FormatBool(searchErr != nil)).
		Observe(time.Since(searchStart).Seconds())

	if honey.Enabled() || traceLogs {
		act := actor.FromContext(ctx)
		ev := honey.NewEvent("gitserver-search")
		ev.SetSampleRate(honeySampleRate("", act.IsInternal()))
		ev.AddField("repo", args.Repo)
		ev.AddField("revisions", args.Revisions)
		ev.AddField("include_diff", args.IncludeDiff)
		ev.AddField("include_modified_files", args.IncludeModifiedFiles)
		ev.AddField("actor", act.UIDString())
		ev.AddField("query", args.Query.String())
		ev.AddField("limit", args.Limit)
		ev.AddField("duration_ms", time.Since(searchStart).Milliseconds())
		if searchErr != nil {
			ev.AddField("error", searchErr.Error())
		}
		if traceID := trace.ID(ctx); traceID != "" {
			ev.AddField("traceID", traceID)
			ev.AddField("trace", trace.URL(traceID, conf.DefaultClient()))
		}
		if honey.Enabled() {
			_ = ev.Send()
		}
		if traceLogs {
			logger.Debug("TRACE gitserver search", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
		}
	}
}

// search handles the core logic of the search. It is passed a matchesBuf so it doesn't need to
// concern itself with event types, and all instrumentation is handled in the calling function.
func (s *Server) search(ctx context.Context, args *protocol.SearchRequest, matchesBuf *streamhttp.JSONArrayBuf) (limitHit bool, err error) {
	args.Repo = protocol.NormalizeRepo(args.Repo)
	if args.Limit == 0 {
		args.Limit = math.MaxInt32
	}

	dir := s.dir(args.Repo)
	if !repoCloned(dir) {
		if conf.Get().DisableAutoGitUpdates {
			s.Logger.Debug("not cloning on demand as DisableAutoGitUpdates is set")
			return false, &gitdomain.RepoNotExistError{
				Repo: args.Repo,
			}
		}

		cloneProgress, cloneInProgress := s.locker.Status(dir)
		if cloneInProgress {
			return false, &gitdomain.RepoNotExistError{
				Repo:            args.Repo,
				CloneInProgress: true,
				CloneProgress:   cloneProgress,
			}
		}

		cloneProgress, err := s.cloneRepo(ctx, args.Repo, nil)
		if err != nil {
			s.Logger.Debug("error starting repo clone", log.String("repo", string(args.Repo)), log.Error(err))
			return false, &gitdomain.RepoNotExistError{
				Repo:            args.Repo,
				CloneInProgress: false,
			}
		}

		return false, &gitdomain.RepoNotExistError{
			Repo:            args.Repo,
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		}
	}

	if !conf.Get().DisableAutoGitUpdates {
		for _, rev := range args.Revisions {
			// TODO add result to trace
			if rev.RevSpec != "" {
				_ = s.ensureRevision(ctx, args.Repo, rev.RevSpec, dir)
			} else if rev.RefGlob != "" {
				_ = s.ensureRevision(ctx, args.Repo, rev.RefGlob, dir)
			}
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Search all commits, sending matching commits down resultChan
	resultChan := make(chan *protocol.CommitMatch, 128)
	g.Go(func() error {
		defer close(resultChan)
		done := ctx.Done()

		mt, err := search.ToMatchTree(args.Query)
		if err != nil {
			return err
		}

		searcher := &search.CommitSearcher{
			Logger:               s.Logger,
			RepoName:             args.Repo,
			RepoDir:              dir.Path(),
			Revisions:            args.Revisions,
			Query:                mt,
			IncludeDiff:          args.IncludeDiff,
			IncludeModifiedFiles: args.IncludeModifiedFiles,
		}

		return searcher.Search(ctx, func(match *protocol.CommitMatch) {
			select {
			case <-done:
			case resultChan <- match:
			}
		})
	})

	// Write matching commits to the stream, flushing occasionally
	g.Go(func() error {
		defer cancel()
		defer matchesBuf.Flush()

		flushTicker := time.NewTicker(50 * time.Millisecond)
		defer flushTicker.Stop()

		sentCount := 0
		firstMatch := true
		for {
			select {
			case result, ok := <-resultChan:
				if !ok {
					return nil
				}

				if sentCount >= args.Limit {
					limitHit = true
					return nil
				}
				sentCount += matchCount(result)

				_ = matchesBuf.Append(result) // EOF only

				// Send immediately if this if the first result we've seen
				if firstMatch {
					_ = matchesBuf.Flush() // EOF only
					firstMatch = false
				}
			case <-flushTicker.C:
				_ = matchesBuf.Flush() // EOF only
			}
		}
	})

	return limitHit, g.Wait()
}

// matchCount returns either:
// 1) the number of diff matches if there are any
// 2) the number of messsage matches if there are any
// 3) one, to represent matching the commit, but nothing inside it
func matchCount(cm *protocol.CommitMatch) int {
	if len(cm.Diff.MatchedRanges) > 0 {
		return len(cm.Diff.MatchedRanges)
	}
	if len(cm.Message.MatchedRanges) > 0 {
		return len(cm.Message.MatchedRanges)
	}
	return 1
}

func (s *Server) handleBatchLog(w http.ResponseWriter, r *http.Request) {
	// 🚨 SECURITY: Only allow POST requests.
	if strings.ToUpper(r.Method) != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	operations := s.ensureOperations()

	// Run git log for a single repository.
	// Invoked multiple times from the handler defined below.
	performGitLogCommand := func(ctx context.Context, repoCommit api.RepoCommit, format string) (output string, isRepoCloned bool, err error) {
		ctx, _, endObservation := operations.batchLogSingle.With(ctx, &err, observation.Args{
			LogFields: append(
				[]otlog.Field{
					otlog.String("format", format),
				},
				repoCommit.LogFields()...,
			),
		})
		defer func() {
			endObservation(1, observation.Args{LogFields: []otlog.Field{
				otlog.Bool("isRepoCloned", isRepoCloned),
			}})
		}()

		dir := s.dir(repoCommit.Repo)
		if !repoCloned(dir) {
			return "", false, nil
		}

		var buf bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "log", "-n", "1", "--name-only", format, string(repoCommit.CommitID))
		dir.Set(cmd)
		cmd.Stdout = &buf

		if _, err := runCommand(ctx, cmd); err != nil {
			return "", true, err
		}

		return buf.String(), true, nil
	}

	// Handles the /batch-log route
	instrumentedHandler := func(ctx context.Context) (statusCodeOnError int, err error) {
		ctx, logger, endObservation := operations.batchLog.With(ctx, &err, observation.Args{})
		defer func() {
			endObservation(1, observation.Args{LogFields: []otlog.Field{
				otlog.Int("statusCodeOnError", statusCodeOnError),
			}})
		}()

		// Read request body
		var req protocol.BatchLogRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return http.StatusBadRequest, err
		}
		logger.Log(req.LogFields()...)

		// Validate request parameters
		if len(req.RepoCommits) == 0 {
			// Early exit: implicitly writes 200 OK
			_ = json.NewEncoder(w).Encode(protocol.BatchLogResponse{Results: []protocol.BatchLogResult{}})
			return 0, nil
		}
		if !strings.HasPrefix(req.Format, "--format=") {
			return http.StatusUnprocessableEntity, errors.New("format parameter expected to be of the form `--format=<git log format>`")
		}

		// Perform requests in each repository in the input batch. We perform these commands
		// concurrently, but only allow for so many commands to be in-flight at a time so that
		// we don't overwhelm a shard with either a large request or too many concurrent batch
		// requests.

		g, ctx := errgroup.WithContext(ctx)
		results := make([]protocol.BatchLogResult, len(req.RepoCommits))

		if s.GlobalBatchLogSemaphore == nil {
			return http.StatusInternalServerError, errors.New("s.GlobalBatchLogSemaphore not initialized")
		}

		for i, repoCommit := range req.RepoCommits {
			// Avoid capture of loop variables
			i, repoCommit := i, repoCommit

			start := time.Now()
			if err := s.GlobalBatchLogSemaphore.Acquire(ctx, 1); err != nil {
				return http.StatusInternalServerError, err
			}
			s.operations.batchLogSemaphoreWait.Observe(time.Since(start).Seconds())

			g.Go(func() error {
				defer s.GlobalBatchLogSemaphore.Release(1)

				output, isRepoCloned, err := performGitLogCommand(ctx, repoCommit, req.Format)
				if err == nil && !isRepoCloned {
					err = errors.Newf("repo not found")
				}
				var errMessage string
				if err != nil {
					errMessage = err.Error()
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

		if err := g.Wait(); err != nil {
			return http.StatusInternalServerError, err
		}

		// Write payload to client: implicitly writes 200 OK
		_ = json.NewEncoder(w).Encode(protocol.BatchLogResponse{Results: results})
		return 0, nil
	}

	// Handle unexpected error conditions. We expect the instrumented handler to not
	// have written the status code or any of the body if this error value is non-nil.
	if statusCodeOnError, err := instrumentedHandler(r.Context()); err != nil {
		http.Error(w, err.Error(), statusCodeOnError)
		return
	}
}

// ensureOperations returns the non-nil operations value supplied to this server
// via RegisterMetrics (when constructed as part of the gitserver binary), or
// constructs and memoizes a no-op operations value (for use in tests).
func (s *Server) ensureOperations() *operations {
	if s.operations == nil {
		s.operations = newOperations(&observation.TestContext)
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

	// Log which which actor is accessing the repo.
	args := req.Args
	cmd := ""
	if len(req.Args) > 0 {
		cmd = req.Args[0]
		args = args[1:]
	}
	accesslog.Record(r.Context(), string(req.Repo), map[string]string{
		"cmd":  cmd,
		"args": strings.Join(args, " "),
	})

	s.exec(w, r, &req)
}

var blockedCommandExecutedCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_gitserver_exec_blocked_command_received",
	Help: "Incremented each time a command not in the allowlist for gitserver is executed",
})

func (s *Server) exec(w http.ResponseWriter, r *http.Request, req *protocol.ExecRequest) {
	logger := s.Logger.Scoped("exec", "").With(log.Strings("req.Args", req.Args))

	// Flush writes more aggressively than standard net/http so that clients
	// with a context deadline see as much partial response body as possible.
	if fw := newFlushingResponseWriter(logger, w); fw != nil {
		w = fw
		defer fw.Close()
	}

	// 🚨 SECURITY: Ensure that only commands in the allowed list are executed.
	// See https://github.com/sourcegraph/security-issues/issues/213.
	if !gitdomain.IsAllowedGitCmd(logger, req.Args) {
		blockedCommandExecutedCounter.Inc()
		logger.Warn("exec: bad command", log.String("RemoteAddr", r.RemoteAddr))

		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid command"))
		return
	}

	ctx := r.Context()

	if !req.NoTimeout {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, shortGitCommandTimeout(req.Args))
		defer cancel()
	}

	start := time.Now()
	var cmdStart time.Time // set once we have ensured commit
	exitStatus := -10810   // sentinel value to indicate not set
	var stdoutN, stderrN int64
	var status string
	var execErr error
	ensureRevisionStatus := "noop"

	req.Repo = protocol.NormalizeRepo(req.Repo)

	// Instrumentation
	{
		repo := repotrackutil.GetTrackedRepo(req.Repo)
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		args := strings.Join(req.Args, " ")

		var tr *trace.Trace
		tr, ctx = trace.New(ctx, "exec."+cmd, string(req.Repo))
		tr.SetAttributes(
			attribute.String("args", args),
			attribute.String("ensure_revision", req.EnsureRevision),
		)
		logger = logger.WithTrace(trace.Context(ctx))

		execRunning.WithLabelValues(cmd, repo).Inc()
		defer func() {
			tr.AddEvent(
				"done",
				attribute.String("status", status),
				attribute.Int64("stdout", stdoutN),
				attribute.Int64("stderr", stderrN),
				attribute.String("ensure_revision_status", ensureRevisionStatus),
			)
			tr.SetError(execErr)
			tr.Finish()

			duration := time.Since(start)
			execRunning.WithLabelValues(cmd, repo).Dec()
			execDuration.WithLabelValues(cmd, repo, status).Observe(duration.Seconds())

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
				ev.SetSampleRate(honeySampleRate(cmd, act.IsInternal()))
				ev.AddField("repo", req.Repo)
				ev.AddField("cmd", cmd)
				ev.AddField("args", args)
				ev.AddField("actor", act.UIDString())
				ev.AddField("ensure_revision", req.EnsureRevision)
				ev.AddField("ensure_revision_status", ensureRevisionStatus)
				ev.AddField("client", r.UserAgent())
				ev.AddField("duration_ms", duration.Milliseconds())
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

	dir := s.dir(req.Repo)
	if !repoCloned(dir) {
		if conf.Get().DisableAutoGitUpdates {
			logger.Debug("not cloning on demand as DisableAutoGitUpdates is set")
			status = "repo-not-found"
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(&protocol.NotFoundPayload{})
			return
		}

		cloneProgress, cloneInProgress := s.locker.Status(dir)
		if cloneInProgress {
			status = "clone-in-progress"
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(&protocol.NotFoundPayload{
				CloneInProgress: true,
				CloneProgress:   cloneProgress,
			})
			return
		}

		cloneProgress, err := s.cloneRepo(ctx, req.Repo, nil)
		if err != nil {
			logger.Debug("error starting repo clone", log.String("repo", string(req.Repo)), log.Error(err))
			status = "repo-not-found"
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: false})
			return
		}
		status = "clone-in-progress"
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(&protocol.NotFoundPayload{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		})
		return
	}

	if !conf.Get().DisableAutoGitUpdates {
		// ensureRevision may kick off a git fetch operation which we don't want if we've
		// configured DisableAutoGitUpdates.
		if s.ensureRevision(ctx, req.Repo, req.EnsureRevision, dir) {
			ensureRevisionStatus = "fetched"
		}
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache")

	w.Header().Set("Trailer", "X-Exec-Error")
	w.Header().Add("Trailer", "X-Exec-Exit-Status")
	w.Header().Add("Trailer", "X-Exec-Stderr")
	w.WriteHeader(http.StatusOK)

	// Special-case `git rev-parse HEAD` requests. These are invoked by search queries for every repo in scope.
	// For searches over large repo sets (> 1k), this leads to too many child process execs, which can lead
	// to a persistent failure mode where every exec takes > 10s, which is disastrous for gitserver performance.
	if len(req.Args) == 2 && req.Args[0] == "rev-parse" && req.Args[1] == "HEAD" {
		if resolved, err := quickRevParseHead(dir); err == nil && isAbsoluteRevision(resolved) {
			_, _ = w.Write([]byte(resolved))
			w.Header().Set("X-Exec-Error", "")
			w.Header().Set("X-Exec-Exit-Status", "0")
			w.Header().Set("X-Exec-Stderr", "")
			return
		}
	}
	// Special-case `git symbolic-ref HEAD` requests. These are invoked by resolvers determining the default branch of a repo.
	// For searches over large repo sets (> 1k), this leads to too many child process execs, which can lead
	// to a persistent failure mode where every exec takes > 10s, which is disastrous for gitserver performance.
	if len(req.Args) == 2 && req.Args[0] == "symbolic-ref" && req.Args[1] == "HEAD" {
		if resolved, err := quickSymbolicRefHead(dir); err == nil {
			_, _ = w.Write([]byte(resolved))
			w.Header().Set("X-Exec-Error", "")
			w.Header().Set("X-Exec-Exit-Status", "0")
			w.Header().Set("X-Exec-Stderr", "")
			return
		}
	}

	var stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &limitWriter{W: &stderrBuf, N: 1024}}

	cmdStart = time.Now()
	cmd := exec.CommandContext(ctx, "git", req.Args...)
	dir.Set(cmd)
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	exitStatus, execErr = runCommand(ctx, cmd)

	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()
	checkMaybeCorruptRepo(s.Logger, req.Repo, dir, stderr)

	// write trailer
	w.Header().Set("X-Exec-Error", errorString(execErr))
	w.Header().Set("X-Exec-Exit-Status", status)
	w.Header().Set("X-Exec-Stderr", stderr)
}

func (s *Server) handleP4Exec(w http.ResponseWriter, r *http.Request) {
	var req protocol.P4ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Args) < 1 {
		http.Error(w, "args must be greater than or equal to 1", http.StatusBadRequest)
		return
	}

	// Make sure the subcommand is explicitly allowed
	allowlist := []string{"protects", "groups", "users", "group"}
	allowed := false
	for _, arg := range allowlist {
		if req.Args[0] == arg {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, fmt.Sprintf("subcommand %q is not allowed", req.Args[0]), http.StatusBadRequest)
		return
	}

	// Log which actor is accessing p4-exec.
	//
	// p4-exec is currently only used for fetching user based permissions information
	// so, we don't have a repo name.
	accesslog.Record(r.Context(), "<no-repo>", map[string]string{
		"p4user": req.P4User,
		"p4port": req.P4Port,
		"args":   strings.Join(req.Args, " "),
	})

	// Make sure credentials are valid before heavier operation
	err := p4pingWithTrust(r.Context(), req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.p4exec(w, r, &req)
}

func (s *Server) p4exec(w http.ResponseWriter, r *http.Request, req *protocol.P4ExecRequest) {
	logger := s.Logger.Scoped("p4exec", "")

	// Flush writes more aggressively than standard net/http so that clients
	// with a context deadline see as much partial response body as possible.
	if fw := newFlushingResponseWriter(logger, w); fw != nil {
		w = fw
		defer fw.Close()
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	start := time.Now()
	var cmdStart time.Time // set once we have ensured commit
	exitStatus := -10810   // sentinel value to indicate not set
	var stdoutN, stderrN int64
	var status string
	var execErr error

	// Instrumentation
	{
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		args := strings.Join(req.Args, " ")

		var tr *trace.Trace
		tr, ctx = trace.New(ctx, "p4exec."+cmd, req.P4Port)
		tr.SetAttributes(attribute.String("args", args))
		logger = logger.WithTrace(trace.Context(ctx))

		execRunning.WithLabelValues(cmd, req.P4Port).Inc()
		defer func() {
			tr.AddEvent("done",
				attribute.String("status", status),
				attribute.Int64("stdout", stdoutN),
				attribute.Int64("stderr", stderrN),
			)
			tr.SetError(execErr)
			tr.Finish()

			duration := time.Since(start)
			execRunning.WithLabelValues(cmd, req.P4Port).Dec()
			execDuration.WithLabelValues(cmd, req.P4Port, status).Observe(duration.Seconds())

			var cmdDuration time.Duration
			if !cmdStart.IsZero() {
				cmdDuration = time.Since(cmdStart)
			}

			isSlow := cmdDuration > 30*time.Second
			if honey.Enabled() || traceLogs || isSlow {
				act := actor.FromContext(ctx)
				ev := honey.NewEvent("gitserver-p4exec")
				ev.SetSampleRate(honeySampleRate(cmd, act.IsInternal()))
				ev.AddField("p4port", req.P4Port)
				ev.AddField("cmd", cmd)
				ev.AddField("args", args)
				ev.AddField("actor", act.UIDString())
				ev.AddField("client", r.UserAgent())
				ev.AddField("duration_ms", duration.Milliseconds())
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_status", exitStatus)
				ev.AddField("status", status)
				if execErr != nil {
					ev.AddField("error", execErr.Error())
				}
				if !cmdStart.IsZero() {
					ev.AddField("cmd_duration_ms", cmdDuration.Milliseconds())
				}

				if traceID := trace.ID(ctx); traceID != "" {
					ev.AddField("traceID", traceID)
					ev.AddField("trace", trace.URL(traceID, conf.DefaultClient()))
				}

				_ = ev.Send()

				if traceLogs {
					logger.Debug("TRACE gitserver p4exec", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
				if isSlow {
					logger.Warn("Long p4exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
			}
		}()
	}

	w.Header().Set("Trailer", "X-Exec-Error")
	w.Header().Add("Trailer", "X-Exec-Exit-Status")
	w.Header().Add("Trailer", "X-Exec-Stderr")
	w.WriteHeader(http.StatusOK)

	var stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &limitWriter{W: &stderrBuf, N: 1024}}

	cmdStart = time.Now()
	cmd := exec.CommandContext(ctx, "p4", req.Args...)
	cmd.Env = append(os.Environ(),
		"P4PORT="+req.P4Port,
		"P4USER="+req.P4User,
		"P4PASSWD="+req.P4Passwd,
	)
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	exitStatus, execErr = runCommand(ctx, cmd)

	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()

	// write trailer
	w.Header().Set("X-Exec-Error", errorString(execErr))
	w.Header().Set("X-Exec-Exit-Status", status)
	w.Header().Set("X-Exec-Stderr", stderr)
}

func (s *Server) setLastFetched(ctx context.Context, name api.RepoName) error {
	dir := s.dir(name)

	lastFetched, err := repoLastFetched(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to get last fetched for %s", name)
	}

	lastChanged, err := repoLastChanged(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to get last changed for %s", name)
	}

	return s.DB.GitserverRepos().SetLastFetched(ctx, name, database.GitserverFetchData{
		LastFetched: lastFetched,
		LastChanged: lastChanged,
		ShardID:     s.Hostname,
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

func (s *Server) setCloneStatus(ctx context.Context, name api.RepoName, status types.CloneStatus) (err error) {
	return s.DB.GitserverRepos().SetCloneStatus(ctx, name, status, s.Hostname)
}

// setCloneStatusNonFatal is the same as setCloneStatus but only logs errors
func (s *Server) setCloneStatusNonFatal(ctx context.Context, name api.RepoName, status types.CloneStatus) {
	if err := s.setCloneStatus(ctx, name, status); err != nil {
		s.Logger.Warn("Setting clone status in DB", log.Error(err))
	}
}

// setRepoSize calculates the size of the repo and stores it in the database.
func (s *Server) setRepoSize(ctx context.Context, name api.RepoName) error {
	return s.DB.GitserverRepos().SetRepoSize(ctx, name, dirSize(s.dir(name).Path(".")), s.Hostname)
}

// setGitAttributes writes our global gitattributes to
// gitDir/info/attributes. This will override .gitattributes inside of
// repositories. It is used to unset attributes such as export-ignore.
func setGitAttributes(dir GitDir) error {
	infoDir := dir.Path("info")
	if err := os.Mkdir(infoDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "failed to set git attributes")
	}

	_, err := fileutil.UpdateFileIfDifferent(
		filepath.Join(infoDir, "attributes"),
		[]byte(`# Managed by Sourcegraph gitserver.

# We want every file to be present in git archive.
* -export-ignore
`))
	if err != nil {
		return errors.Wrap(err, "failed to set git attributes")
	}
	return nil
}

// testRepoCorrupter is used by tests to disrupt a cloned repository (e.g. deleting
// HEAD, zeroing it out, etc.)
var testRepoCorrupter func(ctx context.Context, tmpDir GitDir)

// cloneOptions specify optional behaviour for the cloneRepo function.
type cloneOptions struct {
	// Block will wait for the clone to finish before returning. If the clone
	// fails, the error will be returned. The passed in context is
	// respected. When not blocking the clone is done with a server background
	// context.
	Block bool

	// Overwrite will overwrite the existing clone.
	Overwrite bool

	// CloneFromShard is the hostname of the gitserver instance which is the current owner of the
	// repository. If this is a non-zero string, then gitserver will attempt to clone the repo from
	// that gitserver instance instead of the upstream repo URL of the external service.
	CloneFromShard string
}

// cloneRepo performs a clone operation for the given repository. It is
// non-blocking by default.
func (s *Server) cloneRepo(ctx context.Context, repo api.RepoName, opts *cloneOptions) (cloneProgress string, err error) {
	if isAlwaysCloningTest(repo) {
		return "This will never finish cloning", nil
	}

	// We always want to store whether there was an error cloning the repo
	defer func() {
		// Use a different context in case we failed because the original context failed.
		s.setLastErrorNonFatal(s.ctx, repo, err)
	}()

	dir := s.dir(repo)

	// PERF: Before doing the network request to check if isCloneable, lets
	// ensure we are not already cloning.
	if progress, cloneInProgress := s.locker.Status(dir); cloneInProgress {
		return progress, nil
	}

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return "", errors.Wrap(err, "get VCS syncer")
	}

	var remoteURL *vcs.URL
	if opts != nil && opts.CloneFromShard != "" {
		// are we cloning from the same gitserver instance?
		if s.hostnameMatch(strings.TrimPrefix(opts.CloneFromShard, "http://")) {
			return "", errors.Errorf("cannot clone from the same gitserver instance")
		}

		remoteURL, err = vcs.ParseURL(filepath.Join(opts.CloneFromShard, "git", string(repo)))
	} else {
		// We may be attempting to clone a private repo so we need an internal actor.
		remoteURL, err = s.getRemoteURL(actor.WithInternalActor(ctx), repo)
	}
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

	if err = s.rpsLimiter.Wait(ctx); err != nil {
		return "", err
	}

	if err := syncer.IsCloneable(ctx, remoteURL); err != nil {
		redactedErr := newURLRedactor(remoteURL).redact(err.Error())
		return "", errors.Errorf("error cloning repo: repo %s not cloneable: %s", repo, redactedErr)
	}

	// Mark this repo as currently being cloned. We have to check again if someone else isn't already
	// cloning since we released the lock. We released the lock since isCloneable is a potentially
	// slow operation.
	lock, ok := s.locker.TryAcquire(dir, "starting clone")
	if !ok {
		// Someone else beat us to it
		status, _ := s.locker.Status(dir)
		return status, nil
	}

	if s.skipCloneForTests {
		lock.Release()
		return "", nil
	}

	// We clone to a temporary location first to avoid having incomplete
	// clones in the repo tree. This also avoids leaving behind corrupt clones
	// if the clone is interrupted.
	if opts != nil && opts.Block {
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
	s.CloneQueue.push(&cloneJob{
		repo:      repo,
		dir:       dir,
		syncer:    syncer,
		lock:      lock,
		remoteURL: remoteURL,
		options:   opts,
	})

	return "", nil
}

func (s *Server) doClone(ctx context.Context, repo api.RepoName, dir GitDir, syncer VCSSyncer, lock *RepositoryLock, remoteURL *vcs.URL, opts *cloneOptions) (err error) {
	logger := s.Logger.Scoped("doClone", "").With(log.String("repo", string(repo)))

	defer lock.Release()
	defer func() {
		if err != nil {
			repoCloneFailedCounter.Inc()
		}
	}()
	if err := s.rpsLimiter.Wait(ctx); err != nil {
		return err
	}
	ctx, cancel2 := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel2()

	dstPath := string(dir)
	overwrite := opts != nil && opts.Overwrite
	if !overwrite {
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

	tmpPath, err := s.tempDir("clone-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpPath)
	tmpPath = filepath.Join(tmpPath, ".git")
	tmp := GitDir(tmpPath)

	// It may already be cloned
	if !repoCloned(dir) {
		s.setCloneStatusNonFatal(ctx, repo, types.CloneStatusCloning)
	}
	defer func() {
		// Use a background context to ensure we still update the DB even if we time out
		s.setCloneStatusNonFatal(context.Background(), repo, cloneStatus(repoCloned(dir), false))
	}()

	cmd, err := syncer.CloneCommand(ctx, remoteURL, tmpPath)
	if err != nil {
		return errors.Wrap(err, "get clone command")
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	// see issue #7322: skip LFS content in repositories with Git LFS configured
	cmd.Env = append(cmd.Env, "GIT_LFS_SKIP_SMUDGE=1")
	logger.Info("cloning repo", log.String("tmp", tmpPath), log.String("dst", dstPath))

	pr, pw := io.Pipe()
	defer pw.Close()

	go readCloneProgress(logger, newURLRedactor(remoteURL), lock, pr, repo)

	if output, err := runWith(ctx, cmd, true, pw); err != nil {
		return errors.Wrapf(err, "clone failed. Output: %s", string(output))
	}

	if testRepoCorrupter != nil {
		testRepoCorrupter(ctx, tmp)
	}

	removeBadRefs(ctx, tmp)

	if err := setHEAD(ctx, logger, tmp, syncer, repo, remoteURL); err != nil {
		s.Logger.Error("Failed to ensure HEAD exists", log.String("repo", string(repo)), log.Error(err))
		return errors.Wrap(err, "failed to ensure HEAD exists")
	}

	if err := setRepositoryType(tmp, syncer.Type()); err != nil {
		return errors.Wrap(err, `git config set "sourcegraph.type"`)
	}

	// Update the last-changed stamp.
	if err := setLastChanged(logger, tmp); err != nil {
		return errors.Wrapf(err, "failed to update last changed time")
	}

	// Set gitattributes
	if err := setGitAttributes(tmp); err != nil {
		return err
	}

	// Set gc.auto depending on gitGCMode.
	if err := gitSetAutoGC(tmp); err != nil {
		return err
	}

	if overwrite {
		// remove the current repo by putting it into our temporary directory
		err := fileutil.RenameAndSync(dstPath, filepath.Join(filepath.Dir(tmpPath), "old"))
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

	// Successfully updated, best-effort updating of db fetch state based on
	// disk state.
	if err := s.setLastFetched(ctx, repo); err != nil {
		logger.Warn("failed setting last fetch in DB", log.Error(err))
	}

	// Successfully updated, best-effort calculation of the repo size.
	if err := s.setRepoSize(ctx, repo); err != nil {
		logger.Warn("failed setting repo size", log.Error(err))
	}

	logger.Info("repo cloned")
	repoClonedCounter.Inc()

	return nil
}

// readCloneProgress scans the reader and saves the most recent line of output
// as the lock status.
func readCloneProgress(logger log.Logger, redactor *urlRedactor, lock *RepositoryLock, pr io.Reader, repo api.RepoName) {
	var logFile *os.File
	var err error

	if conf.Get().CloneProgressLog {
		logFile, err = os.CreateTemp("", "")
		if err != nil {
			logger.Warn("failed to create temporary clone log file", log.Error(err), log.String("repo", string(repo)))
		} else {
			logger.Info("logging clone output", log.String("file", logFile.Name()), log.String("repo", string(repo)))
			defer logFile.Close()
		}
	}

	scan := bufio.NewScanner(pr)
	scan.Split(scanCRLF)
	for scan.Scan() {
		progress := scan.Text()

		// 🚨 SECURITY: The output could include the clone url with may contain a sensitive token.
		// Redact the full url and any found HTTP credentials to be safe.
		//
		// e.g.
		// $ git clone http://token@github.com/foo/bar
		// Cloning into 'nick'...
		// fatal: repository 'http://token@github.com/foo/bar/' not found
		redactedProgress := redactor.redact(progress)

		lock.SetStatus(redactedProgress)

		if logFile != nil {
			// Failing to write here is non-fatal and we don't want to spam our logs if there
			// are issues
			_, _ = fmt.Fprintln(logFile, progress)
		}
	}
	if err := scan.Err(); err != nil {
		logger.Error("error reporting progress", log.Error(err))
	}
}

// urlRedactor redacts all sensitive strings from a message.
type urlRedactor struct {
	// sensitive are sensitive strings to be redacted.
	// The strings should not be empty.
	sensitive []string
}

// newURLRedactor returns a new urlRedactor that redacts
// credentials found in rawurl, and the rawurl itself.
func newURLRedactor(parsedURL *vcs.URL) *urlRedactor {
	var sensitive []string
	pw, _ := parsedURL.User.Password()
	u := parsedURL.User.Username()
	if pw != "" && u != "" {
		// Only block password if we have both as we can
		// assume that the username isn't sensitive in this case
		sensitive = append(sensitive, pw)
	} else {
		if pw != "" {
			sensitive = append(sensitive, pw)
		}
		if u != "" {
			sensitive = append(sensitive, u)
		}
	}
	sensitive = append(sensitive, parsedURL.String())
	return &urlRedactor{sensitive: sensitive}
}

// redact returns a redacted version of message.
// Sensitive strings are replaced with "<redacted>".
func (r *urlRedactor) redact(message string) string {
	for _, s := range r.sensitive {
		message = strings.ReplaceAll(message, s, "<redacted>")
	}
	return message
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

// testGitRepoExists is a test fixture that overrides the return value for
// GitRepoSyncer.IsCloneable when it is set.
var testGitRepoExists func(ctx context.Context, remoteURL *vcs.URL) error

var (
	execRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_gitserver_exec_running",
		Help: "number of gitserver.GitCommand running concurrently.",
	}, []string{"cmd", "repo"})
	execDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_gitserver_exec_duration_seconds",
		Help:    "gitserver.GitCommand latencies in seconds.",
		Buckets: trace.UserLatencyBuckets,
	}, []string{"cmd", "repo", "status"})

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

// Send 1 in 16 events to honeycomb. This is hardcoded since we only use this
// for Sourcegraph.com.
//
// 2020-05-29 1 in 4. We are currently at the top tier for honeycomb (before
// enterprise) and using double our quota. This gives us room to grow. If you
// find we keep bumping this / missing data we care about we can look into
// more dynamic ways to sample in our application code.
//
// 2020-07-20 1 in 16. Again hitting very high usage. Likely due to recent
// scaling up of the indexed search cluster. Will require more investigation,
// but we should probably segment user request path traffic vs internal batch
// traffic.
//
// 2020-11-02 Dynamically sample. Again hitting very high usage. Same root
// cause as before, scaling out indexed search cluster. We update our sampling
// to instead be dynamic, since "rev-parse" is 12 times more likely than the
// next most common command.
//
// 2021-08-20 over two hours we did 128 * 128 * 1e6 rev-parse requests
// internally. So we update our sampling to heavily downsample internal
// rev-parse, while upping our sampling for non-internal.
// https://ui.honeycomb.io/sourcegraph/datasets/gitserver-exec/result/67e4bLvUddg
func honeySampleRate(cmd string, internal bool) uint {
	switch {
	case cmd == "rev-parse" && internal:
		return 1 << 14 // 16384

	case internal:
		// we care more about user requests, so downsample internal more.
		return 16

	default:
		return 8
	}
}

var headBranchPattern = lazyregexp.New(`HEAD branch: (.+?)\n`)

func (s *Server) doRepoUpdate(ctx context.Context, repo api.RepoName, revspec string) error {
	span, ctx := ot.StartSpanFromContext(ctx, "Server.doRepoUpdate")
	span.SetTag("repo", repo)
	defer span.Finish()

	if msg, ok := isPaused(filepath.Join(s.ReposDir, string(protocol.NormalizeRepo(repo)))); ok {
		s.Logger.Warn("doRepoUpdate paused", log.String("repo", string(repo)), log.String("reason", msg))
		return nil
	}

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
	err := errors.New("another operation is already in progress")
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
				s.Logger.Error("performing background repo update", log.Error(err))
			}
			s.setLastErrorNonFatal(s.ctx, repo, err)
		})
	}()

	select {
	case <-done:
		return errors.Wrapf(err, "repo %s:", repo)
	case <-ctx.Done():
		span.LogFields(otlog.String("event", "context canceled"))
		return ctx.Err()
	}
}

var doBackgroundRepoUpdateMock func(api.RepoName) error

func (s *Server) doBackgroundRepoUpdate(repo api.RepoName, revspec string) error {
	logger := s.Logger.Scoped("backgroundRepoUpdate", "")

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

	if err = s.rpsLimiter.Wait(ctx); err != nil {
		return err
	}

	repo = protocol.NormalizeRepo(repo)
	dir := s.dir(repo)

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
	defer s.cleanTmpFiles(dir)

	err = syncer.Fetch(ctx, remoteURL, dir, revspec)
	if err != nil {
		return errors.Wrap(err, "failed to fetch")
	}

	removeBadRefs(ctx, dir)

	if err := setHEAD(ctx, logger, dir, syncer, repo, remoteURL); err != nil {
		logger.Error("Failed to ensure HEAD exists", log.String("repo", string(repo)), log.Error(err))
		return errors.Wrap(err, "failed to ensure HEAD exists")
	}

	if err := setRepositoryType(dir, syncer.Type()); err != nil {
		return errors.Wrap(err, `git config set "sourcegraph.type"`)
	}

	// Update the last-changed stamp on disk.
	if err := setLastChanged(logger, dir); err != nil {
		logger.Warn("Failed to update last changed time", log.String("repo", string(repo)), log.Error(err))
	}

	// Successfully updated, best-effort updating of db fetch state based on
	// disk state.
	if err := s.setLastFetched(ctx, repo); err != nil {
		logger.Warn("failed setting last fetch in DB", log.String("repo", string(repo)), log.Error(err))
	}

	// Successfully updated, best-effort calculation of the repo size.
	if err := s.setRepoSize(ctx, repo); err != nil {
		logger.Warn("failed setting repo size", log.String("repo", string(repo)), log.Error(err))
	}

	return nil
}

var (
	badRefsOnce sync.Once
	badRefs     []string
)

// removeBadRefs removes bad refs and tags from the git repo at dir. This
// should be run after a clone or fetch. If your repository contains a ref or
// tag called HEAD (case insensitive), most commands will output a warning
// from git:
//
//	warning: refname 'HEAD' is ambiguous.
//
// Instead we just remove this ref.
func removeBadRefs(ctx context.Context, dir GitDir) {
	// older versions of git do not remove tags case insensitively, so we
	// generate every possible case of HEAD (2^4 = 16)
	badRefsOnce.Do(func() {
		for bits := uint8(0); bits < (1 << 4); bits++ {
			s := []byte("HEAD")
			for i, c := range s {
				// lowercase if the i'th bit of bits is 1
				if bits&(1<<i) != 0 {
					s[i] = c - 'A' + 'a'
				}
			}
			badRefs = append(badRefs, string(s))
		}
	})

	args := append([]string{"branch", "-D"}, badRefs...)
	cmd := exec.CommandContext(ctx, "git", args...)
	dir.Set(cmd)
	_ = cmd.Run()

	args = append([]string{"tag", "-d"}, badRefs...)
	cmd = exec.CommandContext(ctx, "git", args...)
	dir.Set(cmd)
	_ = cmd.Run()
}

// ensureHEAD verifies that there is a HEAD file within the repo, and that it
// is of non-zero length. If either condition is met, we configure a
// best-effort default.
func ensureHEAD(dir GitDir) {
	head, err := os.Stat(dir.Path("HEAD"))
	if os.IsNotExist(err) || head.Size() == 0 {
		os.WriteFile(dir.Path("HEAD"), []byte("ref: refs/heads/master"), 0600)
	}
}

// setHEAD configures git repo defaults (such as what HEAD is) which are
// needed for git commands to work.
func setHEAD(ctx context.Context, logger log.Logger, dir GitDir, syncer VCSSyncer, repo api.RepoName, remoteURL *vcs.URL) error {
	// Verify that there is a HEAD file within the repo, and that it is of
	// non-zero length.
	ensureHEAD(dir)

	// Fallback to git's default branch name if git remote show fails.
	headBranch := "master"

	// try to fetch HEAD from origin
	cmd, err := syncer.RemoteShowCommand(ctx, remoteURL)
	if err != nil {
		return errors.Wrap(err, "get remote show command")
	}
	dir.Set(cmd)
	output, err := runWith(ctx, cmd, true, nil)
	if err != nil {
		logger.Error("Failed to fetch remote info", log.String("repo", string(repo)), log.Error(err), log.String("output", string(output)))
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
		list, err := cmd.Output()
		if err != nil {
			logger.Error("Failed to list branches", log.String("repo", string(repo)), log.Error(err), log.String("output", string(output)))
			return errors.Wrap(err, "failed to list branches")
		}
		lines := strings.Split(string(list), "\n")
		branch := strings.TrimPrefix(strings.TrimPrefix(lines[0], "* "), "  ")
		if branch != "" {
			headBranch = branch
		}
	}

	// set HEAD
	cmd = exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD", "refs/heads/"+headBranch)
	dir.Set(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Error("Failed to set HEAD", log.String("repo", string(repo)), log.Error(err), log.String("output", string(output)))
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
func setLastChanged(logger log.Logger, dir GitDir) error {
	hashFile := dir.Path("sg_refhash")

	hash, err := computeRefHash(dir)
	if err != nil {
		return errors.Wrapf(err, "computeRefHash failed for %s", dir)
	}

	var stamp time.Time
	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		// This is the first time we are calculating the hash. Give a more
		// approriate timestamp for sg_refhash than the current time.
		stamp = computeLatestCommitTimestamp(logger, dir)
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

// computeLatestCommitTimestamp returns the timestamp of the most recent
// commit if any. If there are no commits or the latest commit is in the
// future, or there is any error, time.Now is returned.
func computeLatestCommitTimestamp(logger log.Logger, dir GitDir) time.Time {
	logger = logger.Scoped("computeLatestCommitTimestamp", "compute the timestamp of the most recent commit").
		With(log.String("repo", string(dir)))

	now := time.Now() // return current time if we don't find a more accurate time
	cmd := exec.Command("git", "rev-list", "--all", "--timestamp", "-n", "1")
	dir.Set(cmd)
	output, err := cmd.Output()
	// If we don't have a more specific stamp, we'll return the current time,
	// and possibly an error.
	if err != nil {
		logger.Warn("failed to execute, defaulting to time.Now", log.Error(err))
		return now
	}

	words := bytes.Split(output, []byte(" "))
	// An empty rev-list output, without an error, is okay.
	if len(words) < 2 {
		return now
	}

	// We should have a timestamp and a commit hash; format is
	// 1521316105 ff03fac223b7f16627b301e03bf604e7808989be
	epoch, err := strconv.ParseInt(string(words[0]), 10, 64)
	if err != nil {
		logger.Warn("ignoring corrupted timestamp, defaulting to time.Now", log.String("timestamp", string(words[0])))
		return now
	}
	stamp := time.Unix(epoch, 0)
	if stamp.After(now) {
		return now
	}
	return stamp
}

// computeRefHash returns a hash of the refs for dir. The hash should only
// change if the set of refs and the commits they point to change.
func computeRefHash(dir GitDir) ([]byte, error) {
	// Do not use CommandContext since this is a fast operation we do not want
	// to interrupt.
	cmd := exec.Command("git", "show-ref")
	dir.Set(cmd)
	output, err := cmd.Output()
	if err != nil {
		// Ignore the failure for an empty repository: show-ref fails with
		// empty output and an exit code of 1
		var e *exec.ExitError
		if !errors.As(err, &e) || len(output) != 0 || len(e.Stderr) != 0 || e.Sys().(syscall.WaitStatus).ExitStatus() != 1 {
			return nil, err
		}
	}

	lines := bytes.Split(output, []byte("\n"))
	sort.Slice(lines, func(i, j int) bool {
		return bytes.Compare(lines[i], lines[j]) < 0
	})
	hasher := sha256.New()
	for _, b := range lines {
		_, _ = hasher.Write(b)
		_, _ = hasher.Write([]byte("\n"))
	}
	hash := make([]byte, hex.EncodedLen(hasher.Size()))
	hex.Encode(hash, hasher.Sum(nil))
	return hash, nil
}

func (s *Server) ensureRevision(ctx context.Context, repo api.RepoName, rev string, repoDir GitDir) (didUpdate bool) {
	if rev == "" || rev == "HEAD" {
		return false
	}
	// rev-parse on an OID does not check if the commit actually exists, so it always
	// works. So we append ^0 to force the check
	if isAbsoluteRevision(rev) {
		rev = rev + "^0"
	}
	cmd := exec.Command("git", "rev-parse", rev, "--")
	repoDir.Set(cmd)
	if err := cmd.Run(); err == nil {
		return false
	}
	// Revision not found, update before returning.
	err := s.doRepoUpdate(ctx, repo, rev)
	if err != nil {
		s.Logger.Warn("failed to perform background repo update", log.Error(err), log.String("repo", string(repo)), log.String("rev", rev))
	}
	return true
}

const headFileRefPrefix = "ref: "

// quickSymbolicRefHead best-effort mimics the execution of `git symbolic-ref HEAD`, but doesn't exec a child process.
// It just reads the .git/HEAD file from the bare git repository directory.
func quickSymbolicRefHead(dir GitDir) (string, error) {
	// See if HEAD contains a commit hash and fail if so.
	head, err := os.ReadFile(dir.Path("HEAD"))
	if err != nil {
		return "", err
	}
	head = bytes.TrimSpace(head)
	if isAbsoluteRevision(string(head)) {
		return "", errors.New("ref HEAD is not a symbolic ref")
	}

	// HEAD doesn't contain a commit hash. It contains something like "ref: refs/heads/master".
	if !bytes.HasPrefix(head, []byte(headFileRefPrefix)) {
		return "", errors.New("unrecognized HEAD file format")
	}
	headRef := bytes.TrimPrefix(head, []byte(headFileRefPrefix))
	return string(headRef), nil
}

// quickRevParseHead best-effort mimics the execution of `git rev-parse HEAD`, but doesn't exec a child process.
// It just reads the relevant files from the bare git repository directory.
func quickRevParseHead(dir GitDir) (string, error) {
	// See if HEAD contains a commit hash and return it if so.
	head, err := os.ReadFile(dir.Path("HEAD"))
	if err != nil {
		return "", err
	}
	head = bytes.TrimSpace(head)
	if h := string(head); isAbsoluteRevision(h) {
		return h, nil
	}

	// HEAD doesn't contain a commit hash. It contains something like "ref: refs/heads/master".
	if !bytes.HasPrefix(head, []byte(headFileRefPrefix)) {
		return "", errors.New("unrecognized HEAD file format")
	}
	// Look for the file in refs/heads. If it exists, it contains the commit hash.
	headRef := bytes.TrimPrefix(head, []byte(headFileRefPrefix))
	if bytes.HasPrefix(headRef, []byte("../")) || bytes.Contains(headRef, []byte("/../")) || bytes.HasSuffix(headRef, []byte("/..")) {
		// 🚨 SECURITY: prevent leakage of file contents outside repo dir
		return "", errors.Errorf("invalid ref format: %s", headRef)
	}
	headRefFile := dir.Path(filepath.FromSlash(string(headRef)))
	if refs, err := os.ReadFile(headRefFile); err == nil {
		return string(bytes.TrimSpace(refs)), nil
	}

	// File didn't exist in refs/heads. Look for it in packed-refs.
	f, err := os.Open(dir.Path("packed-refs"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := bytes.Fields(scanner.Bytes())
		if len(fields) != 2 {
			continue
		}
		commit, ref := fields[0], fields[1]
		if bytes.Equal(ref, headRef) {
			return string(commit), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Didn't find the refs/heads/$HEAD_BRANCH in packed_refs
	return "", errors.New("could not compute `git rev-parse HEAD` in-process, try running `git` process")
}

// errorString returns the error string. If err is nil it returns the empty
// string.
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// IsAbsoluteRevision checks if the revision is a git OID SHA string.
//
// Note: This doesn't mean the SHA exists in a repository, nor does it mean it
// isn't a ref. Git allows 40-char hexadecimal strings to be references.
//
// copied from internal/vcs/git to avoid cyclic import
func isAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(('0' <= r && r <= '9') ||
			('a' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return false
		}
	}
	return true
}
