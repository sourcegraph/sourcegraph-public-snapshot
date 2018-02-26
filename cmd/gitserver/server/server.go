package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/trace"

	"github.com/neelance/parallel"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// runCommand runs the command and returns the exit status. All clients of this function should set the context
// in cmd themselves, but we have to pass the context separately here for the sake of tracing.
var runCommand = func(ctx context.Context, cmd *exec.Cmd) (err error, exitCode int) { // mocked by tests
	span, _ := opentracing.StartSpanFromContext(ctx, "runCommand")
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
	exitStatus := -10810
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return err, exitStatus
}
var skipCloneForTests = false // set by tests

// Server is a gitserver server.
type Server struct {
	// ReposDir is the path to the base directory for gitserver storage.
	ReposDir string

	// MaxConcurrentClones controls the maximum number of clones that can
	// happen at once. Used to prevent throttle limits from a code
	// host. Defaults to 100.
	MaxConcurrentClones int

	// ctx is the context we use for all background jobs. It is done when the
	// server is stopped. Do not directly call this, rather call
	// Server.context()
	ctx      context.Context
	cancel   context.CancelFunc // used to shutdown background jobs
	cancelMu sync.Mutex         // protects canceled
	canceled bool
	wg       sync.WaitGroup // tracks running background jobs

	// cloning tracks repositories that are in the process of being cloned
	// by the parent directory of the .git directory they are cloned to.
	cloningMu sync.Mutex
	cloning   map[string]struct{}

	// cloneLimiter and cloneableLimiter limits the number of concurrent
	// clones and ls-remotes respectively. Semaphore size is equal to
	// MaxConcurrentClones.
	cloneLimiter     *parallel.Run
	cloneableLimiter *parallel.Run

	updateRepo        chan<- updateRepoRequest
	repoUpdateLocksMu sync.Mutex // protects the map below and also updates to locks.once
	repoUpdateLocks   map[api.RepoURI]*locks
}

type locks struct {
	once *sync.Once  // consolidates multiple waiting updates
	mu   *sync.Mutex // prevents updates running in parallel
}

type updateRepoRequest struct {
	repo api.RepoURI
	url  string // remote URL
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
		// 10 minutes is a long time, but this never blocks a user request for
		// this long. Even repos that are not that large can take a long time,
		// for example a search over all repos in an organization may have
		// several large repos. All of those repos will be competing for IO =>
		// we need a larger timeout.
		return 10 * time.Minute

	case "ls-remote":
		return 5 * time.Second

	default:
		return time.Minute
	}
}

// This is a timeout for long git commands like clone or remote update.
// that may take a while for large repos. These types of commands should
// be run in the background.
var longGitCommandTimeout = time.Hour

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.cloning = make(map[string]struct{})
	s.updateRepo = s.repoUpdateLoop()
	s.repoUpdateLocks = make(map[api.RepoURI]*locks)
	if s.MaxConcurrentClones == 0 {
		s.MaxConcurrentClones = 100
	}
	s.cloneLimiter = parallel.NewRun(s.MaxConcurrentClones)
	s.cloneableLimiter = parallel.NewRun(s.MaxConcurrentClones)

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", s.handleExec)
	mux.HandleFunc("/list", s.handleList)
	mux.HandleFunc("/is-repo-cloneable", s.handleIsRepoCloneable)
	mux.HandleFunc("/is-repo-cloned", s.handleIsRepoCloned)
	mux.HandleFunc("/repo", s.handleRepoInfo)
	mux.HandleFunc("/enqueue-repo-update", s.handleEnqueueRepoUpdate)
	mux.HandleFunc("/upload-pack", s.handleUploadPack)
	return mux
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

// backgroundWithTimeout returns a context tied to the lifecycle of server.
func (s *Server) backgroundWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	// if we are already canceled don't increment our waitgroup. This is to
	// prevent a loop somewhere preventing us from ever finishing the
	// waitgroup, even though all calls fails instantly due to the canceled
	// context.
	s.cancelMu.Lock()
	if s.canceled {
		s.cancelMu.Unlock()
		return s.ctx, func() {}
	}
	s.wg.Add(1)
	s.cancelMu.Unlock()

	ctx, cancel := context.WithTimeout(s.ctx, timeout)

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

func cloneCmd(ctx context.Context, origin, dir string) *exec.Cmd {
	return exec.CommandContext(ctx, "git", "clone", "--mirror", origin, dir)
}

func (s *Server) setCloneLock(dir string) {
	s.cloningMu.Lock()
	s.cloning[dir] = struct{}{}
	s.cloningMu.Unlock()
}

func (s *Server) releaseCloneLock(dir string) {
	s.cloningMu.Lock()
	delete(s.cloning, dir)
	s.cloningMu.Unlock()
}

func (s *Server) handleIsRepoCloneable(w http.ResponseWriter, r *http.Request) {
	var req protocol.IsRepoCloneableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Repo = protocol.NormalizeRepo(req.Repo)

	if req.URL == "" {
		req.URL = OriginMap(req.Repo)
	}
	if req.URL == "" {
		// BACKCOMPAT: Determine URL from the existing repo on disk if the client didn't send it.
		dir := path.Join(s.ReposDir, string(req.Repo))
		var err error
		req.URL, err = repoRemoteURL(r.Context(), dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if req.URL == "" {
			http.Error(w, "no URL in IsRepoCloneableRequest and no Git remote URL in .git/config", http.StatusInternalServerError)
			return
		}
	}

	var resp protocol.IsRepoCloneableResponse
	if err := s.isCloneable(r.Context(), req.URL); err == nil {
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

func (s *Server) handleIsRepoCloned(w http.ResponseWriter, r *http.Request) {
	var req protocol.IsRepoClonedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dir := path.Join(s.ReposDir, string(req.Repo))
	if repoCloned(dir) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) handleEnqueueRepoUpdate(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Repo = protocol.NormalizeRepo(req.Repo)
	if req.URL == "" {
		log15.Warn("RepoUpdate request is missing Git remote URL.", "repo", req.Repo)
	}
	dir := path.Join(s.ReposDir, string(req.Repo))
	if !repoCloned(dir) && !skipCloneForTests {
		go func() {
			ctx, cancel := s.backgroundWithTimeout(longGitCommandTimeout)
			defer cancel()
			err := s.cloneRepo(ctx, req.Repo, req.URL, dir)
			if err != nil {
				log15.Warn("error cloning repo", "repo", req.Repo, "err", err)
			}
		}()
	} else {
		s.updateRepo <- updateRepoRequest{repo: req.Repo, url: req.URL}
	}
}

func (s *Server) handleExec(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "Server.handleExec")
	defer span.Finish()

	var req protocol.ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Flush writes more aggressively than standard net/http so that clients
	// with a context deadline see as much partial response body as possible.
	if fw := newFlushingResponseWriter(w); fw != nil {
		w = fw
		defer fw.Close()
	}

	ctx, cancel := context.WithTimeout(ctx, shortGitCommandTimeout(req.Args))
	defer cancel()

	start := time.Now()
	exitStatus := -10810 // sentinel value to indicate not set
	var stdoutN, stderrN int64
	var status string
	var errStr string

	req.Repo = protocol.NormalizeRepo(req.Repo)

	// Instrumentation
	{
		repo := repotrackutil.GetTrackedRepo(req.Repo)
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		args := strings.Join(req.Args, " ")

		tr := trace.New("exec."+cmd, string(req.Repo))
		tr.LazyPrintf("args: %s", args)
		execRunning.WithLabelValues(cmd, repo).Inc()
		defer func() {
			tr.LazyPrintf("status=%s stdout=%d stderr=%d", status, stdoutN, stderrN)
			if errStr != "" {
				tr.LazyPrintf("error: %s", errStr)
				tr.SetError()
			}
			tr.Finish()

			duration := time.Since(start)
			execRunning.WithLabelValues(cmd, repo).Dec()
			execDuration.WithLabelValues(cmd, repo, status).Observe(duration.Seconds())
			if honey.Enabled() {
				ev := honey.Event("gitserver-exec")
				ev.AddField("repo", req.Repo)
				ev.AddField("cmd", cmd)
				ev.AddField("args", args)
				ev.AddField("duration_ms", duration.Seconds()*1000)
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_status", exitStatus)
				ev.AddField("status", status)
				if errStr != "" {
					ev.AddField("error", errStr)
				}
				ev.Send()
			}
			if duration > 2500*time.Millisecond {
				log15.Warn("Long exec request", "repo", req.Repo, "args", req.Args, "duration", duration)
			}
		}()
	}

	dir := path.Join(s.ReposDir, string(req.Repo))
	s.cloningMu.Lock()
	_, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if cloneInProgress || strings.ToLower(string(req.Repo)) == "github.com/sourcegraphtest/alwayscloningtest" {
		status = "clone-in-progress"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: true})
		return
	}
	if !repoCloned(dir) {
		err := s.cloneRepo(ctx, req.Repo, req.URL, dir)
		if err != nil {
			log15.Debug("error cloning repo", "repo", req.Repo, "err", err)
			status = "repo-not-found"
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: false})
			return
		}
		status = "clone-in-progress"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: true})
		return
	}

	s.ensureRevision(ctx, req.Repo, req.URL, req.EnsureRevision, dir)

	w.Header().Set("Trailer", "X-Exec-Error, X-Exec-Exit-Status, X-Exec-Stderr")
	w.WriteHeader(http.StatusOK)

	// Special-case `git rev-parse HEAD` requests. These are invoked by search queries for every repo in scope.
	// For searches over large repo sets (> 1k), this leads to too many child process execs, which can lead
	// to a persistent failure mode where every exec takes > 10s, which is disastrous for gitserver performance.
	if len(req.Args) == 2 && req.Args[0] == "rev-parse" && req.Args[1] == "HEAD" {
		if resolved, err := quickRevParseHead(dir); err == nil && len(resolved) == 40 {
			w.Write([]byte(resolved))
			w.Header().Set("X-Exec-Error", "")
			w.Header().Set("X-Exec-Exit-Status", "0")
			w.Header().Set("X-Exec-Stderr", "")
			return
		}
	}

	var stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &stderrBuf}

	cmd := exec.CommandContext(ctx, "git", req.Args...)
	cmd.Dir = dir
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	var err error
	err, exitStatus = runCommand(ctx, cmd)
	if err != nil {
		errStr = err.Error()
	}

	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()
	if len(stderr) > 1024 {
		stderr = stderr[:1024]
	}

	// write trailer
	w.Header().Set("X-Exec-Error", errStr)
	w.Header().Set("X-Exec-Exit-Status", status)
	w.Header().Set("X-Exec-Stderr", string(stderr))
}

// cloneRepo issues a non-blocking git clone command for the given repo to the given directory.
// The repository will be cloned to ${dir}/.git.
func (s *Server) cloneRepo(ctx context.Context, repo api.RepoURI, url, dir string) error {
	// PERF: Before doing the network request to check if isCloneable, lets
	// ensure we are not already cloning.
	s.cloningMu.Lock()
	_, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if cloneInProgress {
		return nil
	}

	if url == "" {
		// BACKCOMPAT: if URL is not specified in API request, look it up in the OriginMap.
		url = OriginMap(repo)
		if url == "" {
			return fmt.Errorf("error cloning repo: no URL provided and origin map entry found for %s", repo)
		}
	}

	// isCloneable causes a network request, so we limit the number that can
	// run at one time. We use a seperate semaphore to cloning since these
	// checks being blocked by a few slow clones will lead to poor feedback to
	// users. We can defer since the rest of the function does not block this
	// goroutine.
	s.cloneableLimiter.Acquire()
	defer s.cloneableLimiter.Release()
	if err := s.isCloneable(ctx, url); err != nil {
		return fmt.Errorf("error cloning repo: repo %s (%s) not cloneable: %s", repo, url, err)
	}

	// Mark this repo as currently being cloned. We have to check again if someone else isn't already
	// cloning since we released the lock. We released the lock since isCloneable is a potentially
	// slow operation.
	s.cloningMu.Lock()
	_, cloneInProgress = s.cloning[dir]
	if cloneInProgress {
		s.cloningMu.Unlock()
		return nil
	}
	s.cloning[dir] = struct{}{} // Mark this repo as currently being cloned.
	s.cloningMu.Unlock()

	if skipCloneForTests {
		s.releaseCloneLock(dir)
		return nil
	}

	go func() {
		s.cloneLimiter.Acquire()
		defer s.cloneLimiter.Release()

		// Create a new context because this is in a background goroutine.
		ctx, cancel := s.backgroundWithTimeout(longGitCommandTimeout)
		defer func() {
			cancel()
			s.releaseCloneLock(dir)
		}()

		path := filepath.Join(dir, ".git")
		cmd := cloneCmd(ctx, url, path)
		if err := s.removeAll(path); err != nil {
			log15.Error("failed to clean up before clone", "path", path, "error", err)
		}

		log15.Debug("cloning repo", "repo", repo, "url", url, "dir", dir)
		if output, err := s.runWithRemoteOpts(ctx, cmd); err != nil {
			log15.Error("clone failed", "error", err, "output", string(output))
			if err := s.removeAll(path); err != nil {
				log15.Error("failed to clean up after clone", "path", path, "error", err)
			}
			return
		}
		log15.Debug("repo cloned", "repo", repo)
	}()

	return nil
}

// testRepoExists is a test fixture that overrides the return value
// for isCloneable when it is set.
var testRepoExists func(ctx context.Context, url string) error

// isCloneable checks to see if the Git remote URL is cloneable.
func (s *Server) isCloneable(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, shortGitCommandTimeout([]string{"ls-remote"}))
	defer cancel()

	if strings.ToLower(string(protocol.NormalizeRepo(api.RepoURI(url)))) == "github.com/sourcegraphtest/alwayscloningtest" {
		return nil
	}
	if testRepoExists != nil {
		return testRepoExists(ctx, url)
	}

	cmd := exec.CommandContext(ctx, "git", "ls-remote", url, "HEAD")
	out, err := s.runWithRemoteOpts(ctx, cmd)
	if err != nil {
		if len(out) > 0 {
			err = fmt.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

var execRunning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "exec_running",
	Help:      "number of gitserver.Command running concurrently.",
}, []string{"cmd", "repo"})
var execDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "exec_duration_seconds",
	Help:      "gitserver.Command latencies in seconds.",
	Buckets:   traceutil.UserLatencyBuckets,
}, []string{"cmd", "repo", "status"})

func init() {
	prometheus.MustRegister(execRunning)
	prometheus.MustRegister(execDuration)
}

func (s *Server) repoUpdateLoop() chan<- updateRepoRequest {
	updateRepo := make(chan updateRepoRequest, 10)
	lastCheckAt := make(map[api.RepoURI]time.Time)

	go func() {
		for req := range updateRepo {
			if t, ok := lastCheckAt[req.repo]; ok && time.Now().Before(t.Add(10*time.Second)) {
				continue // git data still fresh
			}
			lastCheckAt[req.repo] = time.Now()
			go func(req updateRepoRequest) {
				// Create a new context with a new timeout (instead of passing one through updateRepoRequest)
				// because the ctx of the updateRepoRequest sender will get cancelled before this goroutine runs.
				ctx, cancel := s.backgroundWithTimeout(longGitCommandTimeout)
				defer cancel()
				s.doRepoUpdate(ctx, req.repo, req.url)
			}(req)
		}
	}()

	return updateRepo
}

var headBranchPattern = regexp.MustCompile("HEAD branch: (.+?)\\n")

func (s *Server) doRepoUpdate(ctx context.Context, repo api.RepoURI, url string) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Server.doRepoUpdate")
	span.SetTag("repo", repo)
	span.SetTag("url", url)
	defer span.Finish()

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

	once.Do(func() {
		mu.Lock() // Prevent multiple updates in parallel. It works fine, but it wastes resources.
		defer mu.Unlock()

		s.repoUpdateLocksMu.Lock()
		l.once = new(sync.Once) // Make new requests wait for next update.
		s.repoUpdateLocksMu.Unlock()

		s.doRepoUpdate2(ctx, repo, url)
	})
}

func (s *Server) doRepoUpdate2(ctx context.Context, repo api.RepoURI, url string) {
	dir := path.Join(s.ReposDir, string(repo))

	var urlIsGitRemote bool
	if url == "" {
		// BACKCOMPAT: if URL is not specified in API request, look it up in the OriginMap.
		url = OriginMap(repo)
	}
	if url == "" {
		// log15.Warn("Deprecated: use of saved Git remote for repo updating (API client should set URL)", "repo", repo)
		var err error
		url, err = repoRemoteURL(ctx, dir)
		if err != nil || url == "" {
			log15.Error("Failed to determine Git remote URL", "repo", repo, "error", err, "url", url)
			return
		}
		urlIsGitRemote = true
	}

	{
		// Update Git remote URL if it differs from the configured remote URL (so that callers that
		// don't specify the URL in the request will use the latest remote URL).
		var gitRemoteURL string
		if urlIsGitRemote {
			gitRemoteURL = url
		} else {
			var err error
			gitRemoteURL, err = repoRemoteURL(ctx, dir)
			if err != nil || url == "" {
				log15.Error("Failed to determine Git remote URL", "repo", repo, "error", err, "url", url)
				return
			}
		}

		cmd := exec.CommandContext(ctx, "git", "remote", "set-url", "origin", "--", gitRemoteURL)
		cmd.Dir = dir
		if err, _ := runCommand(ctx, cmd); err != nil {
			log15.Error("Failed to update repository's Git remote URL.", "repo", repo, "url", url)
		}
	}

	cmd := exec.CommandContext(ctx, "git", "fetch", "--prune", url, "+refs/*:refs/*")
	cmd.Dir = dir
	if output, err := s.runWithRemoteOpts(ctx, cmd); err != nil {
		log15.Error("Failed to update", "repo", repo, "error", err, "output", string(output))
		return
	}

	headBranch := "master"

	// try to fetch HEAD from origin
	cmd = exec.CommandContext(ctx, "git", "remote", "show", url)
	cmd.Dir = path.Join(s.ReposDir, string(repo))
	output, err := s.runWithRemoteOpts(ctx, cmd)
	if err != nil {
		log15.Error("Failed to fetch remote info", "repo", repo, "url", url, "error", err, "output", string(output))
		return
	}
	submatches := headBranchPattern.FindSubmatch(output)
	if len(submatches) == 2 {
		submatch := string(submatches[1])
		if submatch != "(unknown)" {
			headBranch = string(submatch)
		}
	}

	// check if branch pointed to by HEAD exists
	cmd = exec.CommandContext(ctx, "git", "rev-parse", headBranch, "--")
	cmd.Dir = path.Join(s.ReposDir, string(repo))
	if err := cmd.Run(); err != nil {
		// branch does not exist, pick first branch
		cmd := exec.CommandContext(ctx, "git", "branch")
		cmd.Dir = path.Join(s.ReposDir, string(repo))
		list, err := cmd.Output()
		if err != nil {
			log15.Error("Failed to list branches", "repo", repo, "error", err, "output", string(output))
			return
		}
		lines := strings.Split(string(list), "\n")
		branch := strings.TrimPrefix(strings.TrimPrefix(lines[0], "* "), "  ")
		if branch != "" {
			headBranch = branch
		}
	}

	// set HEAD
	cmd = exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD", "refs/heads/"+headBranch)
	cmd.Dir = path.Join(s.ReposDir, string(repo))
	if output, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to set HEAD", "repo", repo, "error", err, "output", string(output))
		return
	}
}

func (s *Server) ensureRevision(ctx context.Context, repo api.RepoURI, url, rev, repoDir string) {
	if rev == "" {
		return
	}
	if rev == "HEAD" {
		if _, err := quickRevParseHead(repoDir); err == nil {
			return
		}
	}
	cmd := exec.CommandContext(ctx, "git", "rev-parse", rev, "--")
	cmd.Dir = repoDir
	if err := cmd.Run(); err == nil {
		return
	}
	if rev == "HEAD" {
		return
	}
	// Revision not found, update before returning.
	s.doRepoUpdate(ctx, repo, url)
}

// quickRevParseHead best-effort mimics the execution of `git rev-parse HEAD`, but doesn't exec a child process.
// It just reads the relevant files from the bare git repository directory.
func quickRevParseHead(dir string) (string, error) {
	// See if HEAD contains a commit hash and return it if so.
	head, err := ioutil.ReadFile(filepath.Join(dir, "HEAD"))
	if os.IsNotExist(err) {
		dir = filepath.Join(dir, ".git")
		head, err = ioutil.ReadFile(filepath.Join(dir, "HEAD"))
	}
	if err != nil {
		return "", err
	}
	head = bytes.TrimSpace(head)
	if len(head) == 40 {
		return string(head), nil
	}

	// HEAD doesn't contain a commit hash. It contains something like "ref: refs/heads/master".
	if !bytes.HasPrefix(head, []byte("ref: ")) {
		return "", errors.New("unrecognized HEAD file format")
	}
	// Look for the file in refs/heads. If it exists, it contains the commit hash.
	headRef := bytes.TrimPrefix(head, []byte("ref: "))
	if bytes.HasPrefix(headRef, []byte("../")) || bytes.Contains(headRef, []byte("/../")) || bytes.HasSuffix(headRef, []byte("/..")) {
		// 🚨 SECURITY: prevent leakage of file contents outside repo dir
		return "", fmt.Errorf("invalid ref format: %s", headRef)
	}
	headRefFile := filepath.Join(dir, filepath.FromSlash(string(headRef)))
	if refs, err := ioutil.ReadFile(headRefFile); err == nil {
		return string(bytes.TrimSpace(refs)), nil
	}

	// File didn't exist in refs/heads. Look for it in packed-refs.
	f, err := os.Open(filepath.Join(dir, "packed-refs"))
	if err != nil {
		return "", err
	}
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
