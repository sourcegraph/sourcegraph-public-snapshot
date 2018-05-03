package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	nettrace "golang.org/x/net/trace"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/honey"
	"github.com/sourcegraph/sourcegraph/pkg/mutablelimiter"
	"github.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
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

	// ctx is the context we use for all background jobs. It is done when the
	// server is stopped. Do not directly call this, rather call
	// Server.context()
	ctx      context.Context
	cancel   context.CancelFunc // used to shutdown background jobs
	cancelMu sync.Mutex         // protects canceled
	canceled bool
	wg       sync.WaitGroup // tracks running background jobs

	// cloningMu protects cloning
	cloningMu sync.Mutex
	// cloning tracks repositories that are in the process of being cloned
	// by the parent directory of the .git directory they are cloned to.
	// The value is the last line of output from the running clone command.
	cloning map[string]string

	// cloneLimiter and cloneableLimiter limits the number of concurrent
	// clones and ls-remotes respectively. Use s.acquireCloneLimiter() and
	// s.acquireClonableLimiter() instead of using these directly.
	cloneLimiter     *mutablelimiter.Limiter
	cloneableLimiter *mutablelimiter.Limiter

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
	s.cloning = make(map[string]string)
	s.updateRepo = s.repoUpdateLoop()
	s.repoUpdateLocks = make(map[api.RepoURI]*locks)

	// GitMaxConcurrentClones controls the maximum number of clones that
	// can happen at once. Used to prevent throttle limits from a code
	// host. Defaults to 5.
	maxConcurrentClones := conf.Get().GitMaxConcurrentClones
	if maxConcurrentClones == 0 {
		maxConcurrentClones = 5
	}
	s.cloneLimiter = mutablelimiter.New(maxConcurrentClones)
	s.cloneableLimiter = mutablelimiter.New(maxConcurrentClones)
	conf.Watch(func() {
		limit := conf.Get().GitMaxConcurrentClones
		if limit == 0 {
			limit = 5
		}
		s.cloneLimiter.SetLimit(limit)
		s.cloneableLimiter.SetLimit(limit)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", s.handleExec)
	mux.HandleFunc("/list", s.handleList)
	mux.HandleFunc("/is-repo-cloneable", s.handleIsRepoCloneable)
	mux.HandleFunc("/is-repo-cloned", s.handleIsRepoCloned)
	mux.HandleFunc("/repo", s.handleRepoInfo)
	mux.HandleFunc("/enqueue-repo-update", s.handleEnqueueRepoUpdate)
	mux.HandleFunc("/upload-pack", s.handleUploadPack)
	mux.HandleFunc("/getGitolitePhabricatorMetadata", s.handleGetGitolitePhabricatorMetadata)
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

// serverContext returns a child context tied to the lifecycle of server.
func (s *Server) serverContext() (context.Context, context.CancelFunc) {
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

func (s *Server) acquireCloneLimiter(ctx context.Context) (context.Context, context.CancelFunc, error) {
	cloneQueue.Inc()
	defer cloneQueue.Dec()
	return s.cloneLimiter.Acquire(ctx)
}

func (s *Server) acquireCloneableLimiter(ctx context.Context) (context.Context, context.CancelFunc, error) {
	lsRemoteQueue.Inc()
	defer lsRemoteQueue.Dec()
	return s.cloneableLimiter.Acquire(ctx)
}

func cloneCmd(ctx context.Context, origin, dir string, progress bool) *exec.Cmd {
	args := []string{"clone", "--mirror"}
	if progress {
		args = append(args, "--progress")
	}
	args = append(args, origin, dir)
	return exec.CommandContext(ctx, "git", args...)
}

func (s *Server) setCloneLock(dir string) {
	s.cloningMu.Lock()
	s.cloning[dir] = ""
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
	dir := path.Join(s.ReposDir, string(req.Repo))
	if !repoCloned(dir) && !skipCloneForTests {
		go func() {
			ctx, cancel1 := s.serverContext()
			defer cancel1()
			ctx, cancel2 := context.WithTimeout(ctx, longGitCommandTimeout)
			defer cancel2()
			_, err := s.cloneRepo(ctx, req.Repo, req.URL, dir)
			if err != nil {
				log15.Warn("error cloning repo", "repo", req.Repo, "err", err)
			}
		}()
	} else {
		updateQueue.Inc()
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
	var cmdStart time.Time // set once we have ensured commit
	exitStatus := -10810   // sentinel value to indicate not set
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

		tr := nettrace.New("exec."+cmd, string(req.Repo))
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

			var cmdDuration time.Duration
			var fetchDuration time.Duration
			if !cmdStart.IsZero() {
				cmdDuration = time.Since(cmdStart)
				fetchDuration = cmdStart.Sub(start)
			}

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
				if !cmdStart.IsZero() {
					ev.AddField("cmd_duration_ms", cmdDuration.Seconds()*1000)
					ev.AddField("fetch_duration_ms", fetchDuration.Seconds()*1000)
				}
				ev.Send()
			}
			if cmdDuration > 2500*time.Millisecond {
				log15.Warn("Long exec request", "repo", req.Repo, "args", req.Args, "duration", cmdDuration)
			}
			if fetchDuration > 10*time.Second {
				log15.Warn("Slow fetch/clone for exec request", "repo", req.Repo, "args", req.Args, "duration", fetchDuration)
			}
		}()
	}

	dir := path.Join(s.ReposDir, string(req.Repo))
	s.cloningMu.Lock()
	cloneProgress, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if strings.ToLower(string(req.Repo)) == "github.com/sourcegraphtest/alwayscloningtest" {
		cloneInProgress = true
		cloneProgress = "This will never finish cloning"
	}
	if cloneInProgress {
		status = "clone-in-progress"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&protocol.NotFoundPayload{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		})
		return
	}
	if !repoCloned(dir) {
		cloneProgress, err := s.cloneRepo(ctx, req.Repo, req.URL, dir)
		if err != nil {
			log15.Debug("error cloning repo", "repo", req.Repo, "err", err)
			status = "repo-not-found"
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&protocol.NotFoundPayload{CloneInProgress: false})
			return
		}
		status = "clone-in-progress"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&protocol.NotFoundPayload{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		})
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

	cmdStart = time.Now()
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
func (s *Server) cloneRepo(ctx context.Context, repo api.RepoURI, url, dir string) (string, error) {
	// PERF: Before doing the network request to check if isCloneable, lets
	// ensure we are not already cloning.
	s.cloningMu.Lock()
	progress, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if cloneInProgress {
		return progress, nil
	}

	if url == "" {
		// BACKCOMPAT: if URL is not specified in API request, look it up in the OriginMap.
		url = OriginMap(repo)
		if url == "" {
			return "", fmt.Errorf("error cloning repo: no URL provided and origin map entry found for %s", repo)
		}
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
	if err := s.isCloneable(ctx, url); err != nil {
		return "", fmt.Errorf("error cloning repo: repo %s (%s) not cloneable: %s", repo, url, err)
	}

	// Mark this repo as currently being cloned. We have to check again if someone else isn't already
	// cloning since we released the lock. We released the lock since isCloneable is a potentially
	// slow operation.
	s.cloningMu.Lock()
	progress, cloneInProgress = s.cloning[dir]
	if cloneInProgress {
		s.cloningMu.Unlock()
		return progress, nil
	}
	s.cloning[dir] = "" // Mark this repo as currently being cloned.
	s.cloningMu.Unlock()

	if skipCloneForTests {
		s.releaseCloneLock(dir)
		return "", nil
	}

	go func() {
		defer s.releaseCloneLock(dir)

		// Create a new context because this is in a background goroutine.
		ctx, cancel1 := s.serverContext()
		defer cancel1()
		ctx, cancel2, err := s.acquireCloneLimiter(ctx)
		if err != nil {
			log.Println("unexpected error while acquiring clone limiter:", err)
			return
		}
		defer cancel2()
		ctx, cancel3 := context.WithTimeout(ctx, longGitCommandTimeout)
		defer cancel3()

		path := filepath.Join(dir, ".git")
		cmd := cloneCmd(ctx, url, path, true)
		log15.Debug("cloning repo", "repo", repo, "url", url, "dir", dir)

		pr, pw := io.Pipe()
		defer pw.Close()
		go s.readCloneProgress(repo, url, dir, pr)

		if output, err := s.runWithRemoteOpts(ctx, cmd, pw); err != nil {
			log15.Error("clone failed", "error", err, "output", string(output))
			if err := s.removeAll(path); err != nil {
				log15.Error("failed to clean up after clone", "path", path, "error", err)
			}
			return
		}
		log15.Debug("repo cloned", "repo", repo)
	}()

	return "", nil
}

// readCloneProgress scans the reader and saves the most recent line of output as progress.
func (s *Server) readCloneProgress(repo api.RepoURI, url, dir string, pr io.Reader) {
	scan := bufio.NewScanner(pr)
	scan.Split(scanCRLF)
	redactor := newURLRedactor(url)
	for scan.Scan() {
		progress := scan.Text()
		log15.Debug("clone progress", "repo", repo, "url", url, "progress", progress)

		// 🚨 SECURITY: The output could include the clone url with may contain a sensitive token.
		// Redact the full url and any found HTTP credentials to be safe.
		//
		// e.g.
		// $ git clone http://token@github.com/foo/bar
		// Cloning into 'nick'...
		// fatal: repository 'http://token@github.com/foo/bar/' not found
		redactedProgress := redactor.redact(progress)

		s.cloningMu.Lock()
		if _, ok := s.cloning[dir]; ok {
			s.cloning[dir] = redactedProgress
		}
		s.cloningMu.Unlock()
	}
	if err := scan.Err(); err != nil {
		log15.Error("error reporting progress", "error", err)
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
func newURLRedactor(rawurl string) *urlRedactor {
	var sensitive []string
	parsedURL, _ := url.Parse(rawurl)
	if parsedURL != nil {
		if pw, _ := parsedURL.User.Password(); pw != "" {
			sensitive = append(sensitive, pw)
		}
		if u := parsedURL.User.Username(); u != "" {
			sensitive = append(sensitive, u)
		}
	}
	sensitive = append(sensitive, rawurl)
	return &urlRedactor{sensitive: sensitive}
}

// redact returns a redacted version of message.
// Sensitive strings are replaced with "<redacted>".
func (r *urlRedactor) redact(message string) string {
	for _, s := range r.sensitive {
		message = strings.Replace(message, s, "<redacted>", -1)
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
	out, err := s.runWithRemoteOpts(ctx, cmd, nil)
	if err != nil {
		if len(out) > 0 {
			err = fmt.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

var (
	execRunning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "exec_running",
		Help:      "number of gitserver.Command running concurrently.",
	}, []string{"cmd", "repo"})
	execDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "exec_duration_seconds",
		Help:      "gitserver.Command latencies in seconds.",
		Buckets:   trace.UserLatencyBuckets,
	}, []string{"cmd", "repo", "status"})
	cloneQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "clone_queue",
		Help:      "number of repos waiting to be cloned.",
	})
	lsRemoteQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "lsremote_queue",
		Help:      "number of repos waiting to check existence on remote code host (git ls-remote).",
	})
	updateQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "update_queue",
		Help:      "number of repos waiting to be updated (enqueue-repo-update)",
	})
)

func init() {
	prometheus.MustRegister(execRunning)
	prometheus.MustRegister(execDuration)
	prometheus.MustRegister(cloneQueue)
	prometheus.MustRegister(lsRemoteQueue)
	prometheus.MustRegister(updateQueue)
}

func (s *Server) repoUpdateLoop() chan<- updateRepoRequest {
	updateRepo := make(chan updateRepoRequest, 10)
	lastCheckAt := make(map[api.RepoURI]time.Time)

	go func() {
		for req := range updateRepo {
			updateQueue.Dec()

			if t, ok := lastCheckAt[req.repo]; ok && time.Now().Before(t.Add(10*time.Second)) {
				continue // git data still fresh
			}
			lastCheckAt[req.repo] = time.Now()
			go func(req updateRepoRequest) {
				// Create a new context with a new timeout (instead of passing one through updateRepoRequest)
				// because the ctx of the updateRepoRequest sender will get cancelled before this goroutine runs.
				ctx, cancel1 := s.serverContext()
				defer cancel1()
				ctx, cancel2 := context.WithTimeout(ctx, longGitCommandTimeout)
				defer cancel2()
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
	ctx, cancel, err := s.acquireCloneLimiter(ctx)
	if err != nil {
		log15.Error("error acquiring clone lock for update", "err", err, "repo", repo)
		return
	}
	defer cancel()

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
	if output, err := s.runWithRemoteOpts(ctx, cmd, nil); err != nil {
		log15.Error("Failed to update", "repo", repo, "error", err, "output", string(output))
		return
	}

	headBranch := "master"

	// try to fetch HEAD from origin
	cmd = exec.CommandContext(ctx, "git", "remote", "show", url)
	cmd.Dir = path.Join(s.ReposDir, string(repo))
	output, err := s.runWithRemoteOpts(ctx, cmd, nil)
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
