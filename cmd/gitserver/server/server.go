package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/neelance/chanrpc"
	"github.com/neelance/chanrpc/chanrpcutil"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/originmap"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var runCommand = func(cmd *exec.Cmd) (error, int) { // mocked by tests
	err := cmd.Run()
	exitStatus := -10810
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return err, exitStatus
}
var noUpdates = false // set by tests

// Server is a gitserver server.
type Server struct {
	// ReposDir is the path to the base directory for gitserver storage.
	ReposDir string

	// InsecureSkipCheckVerifySSH controls whether the client verifies the
	// SSH server's certificate or host key. If InsecureSkipCheckVerifySSH
	// is true, the program is susceptible to a man-in-the-middle
	// attack. This should only be used for testing.
	InsecureSkipCheckVerifySSH bool

	// cloning tracks repositories (key is '/'-separated path) that are
	// in the process of being cloned.
	cloningMu sync.Mutex
	cloning   map[string]struct{}

	updateRepo        chan<- updateRepoRequest
	repoUpdateLocksMu sync.Mutex // protects the map below and also updates to locks.once
	repoUpdateLocks   map[string]*locks
}

type locks struct {
	once *sync.Once  // consolidates multiple waiting updates
	mu   *sync.Mutex // prevents updates running in parallel
}

type updateRepoRequest struct {
	repo string
	opt  *vcs.RemoteOpts
}

// Serve serves incoming http requests on listener l.
func (s *Server) Handler() http.Handler {
	s.cloning = make(map[string]struct{})
	s.updateRepo = s.repoUpdateLoop()
	s.repoUpdateLocks = make(map[string]*locks)

	if err := initializeSSH(); err != nil {
		log.Printf("SSH initialization error: %s", err)
	}

	s.registerMetrics()

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {
		var req protocol.ExecRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		stdout, err := s.handleExecRequest(&req)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(err); err != nil {
				log.Print(err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		io.Copy(w, stdout)
	})
	return mux
}

// ServeLegacy serves incoming legacy gitserver requests on listener l.
func (s *Server) ServeLegacy(l net.Listener) error {
	s.cloning = make(map[string]struct{})
	s.updateRepo = s.repoUpdateLoop()
	s.repoUpdateLocks = make(map[string]*locks)

	requests := make(chan *protocol.LegacyRequest, 100)
	go s.processRequests(requests)
	srv := &chanrpc.Server{RequestChan: requests}
	return srv.Serve(l)
}

func (s *Server) processRequests(requests <-chan *protocol.LegacyRequest) {
	for req := range requests {
		if req.Exec != nil {
			go s.handleLegacyExecRequest(req.Exec)
		}
	}
}

func (s *Server) handleExecRequest(req *protocol.ExecRequest) (io.Reader, *protocol.ExecError) {
	replyChan := make(chan *protocol.LegacyExecReply, 1)
	s.handleLegacyExecRequest(&protocol.LegacyExecRequest{
		Repo:           req.Repo,
		EnsureRevision: req.EnsureRevision,
		Args:           req.Args,
		Opt:            req.Opt,
		NoAutoUpdate:   req.NoAutoUpdate,
		ReplyChan:      replyChan,
	})
	reply := <-replyChan
	if reply.RepoNotFound || reply.CloneInProgress {
		return nil, &protocol.ExecError{
			RepoNotFound:    reply.RepoNotFound,
			CloneInProgress: reply.CloneInProgress,
		}
	}

	for b := range reply.Stdout {
		if len(b) == 0 {
			continue
		}

		// got first output, assume no error and stream stdout to reader
		go func() { <-reply.ProcessResult }() // discard process result
		go chanrpcutil.Drain(reply.Stderr)    // discard stderr

		stdout := make(chan []byte, 10)
		stdout <- b
		go func() {
			for b := range reply.Stdout {
				stdout <- b
			}
			close(stdout)
		}()
		return chanrpcutil.NewReader(stdout), nil
	}

	// stdout closed without any output, check for errors
	result := <-reply.ProcessResult
	if result.Error != "" || result.ExitStatus != 0 {
		return nil, &protocol.ExecError{
			Error:      result.Error,
			ExitStatus: result.ExitStatus,
			Stderr:     string(<-chanrpcutil.ReadAll(reply.Stderr)),
		}
	}

	// no output and no error
	chanrpcutil.Drain(reply.Stderr)
	return bytes.NewReader(nil), nil
}

// This is a timeout for git commands that should not take a long time.
var shortGitCommandTimeout = time.Minute

// This is a timeout for long git commands like clone or remote update.
// that may take a while for large repos. These types of commands should
// be run in the background.
var longGitCommandTimeout = time.Hour

// handleExecRequest handles a exec request.
func (s *Server) handleLegacyExecRequest(req *protocol.LegacyExecRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), shortGitCommandTimeout)
	defer cancel()

	start := time.Now()
	exitStatus := -10810 // sentinel value to indicate not set
	var stdoutN, stderrN int64
	var status string
	var errStr string

	defer recoverAndLog()
	defer close(req.ReplyChan)

	if req.Stdin != nil {
		go chanrpcutil.Drain(req.Stdin) // deprecated
	}

	req.Repo = protocol.NormalizeRepo(req.Repo)

	// This is a repo that we use for testing the cloning state of the UI
	if req.Repo == "github.com/sourcegraphtest/alwayscloningtest" {
		req.ReplyChan <- &protocol.LegacyExecReply{CloneInProgress: true}
		return
	}

	// Instrumentation
	{
		repo := repotrackutil.GetTrackedRepo(req.Repo)
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		execRunning.WithLabelValues(cmd, repo).Inc()
		defer func() {
			duration := time.Since(start)
			execRunning.WithLabelValues(cmd, repo).Dec()
			execDuration.WithLabelValues(cmd, repo, status).Observe(duration.Seconds())
			// Only log to honeycomb if we have the repo to reduce noise
			if ranGit := exitStatus != -10810; ranGit && honey.Enabled() {
				ev := honey.Event("gitserver-exec")
				ev.AddField("repo", req.Repo)
				ev.AddField("cmd", cmd)
				ev.AddField("args", strings.Join(req.Args, " "))
				ev.AddField("duration_ms", duration.Seconds()*1000)
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_status", exitStatus)
				if errStr != "" {
					ev.AddField("error", errStr)
				}
				ev.Send()
			}
		}()
	}

	dir := path.Join(s.ReposDir, req.Repo)
	s.cloningMu.Lock()
	_, cloneInProgress := s.cloning[dir]
	if cloneInProgress {
		s.cloningMu.Unlock()
		req.ReplyChan <- &protocol.LegacyExecReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if !repoExists(dir) {
		if origin := originmap.Map(req.Repo); origin != "" && !req.NoAutoUpdate && !noUpdates {
			s.cloning[dir] = struct{}{} // Mark this repo as currently being cloned.
			s.cloningMu.Unlock()

			go func() {
				// Create a new context because this is in a background goroutine.
				ctx, cancel := context.WithTimeout(context.Background(), longGitCommandTimeout)
				defer func() {
					cancel()
					s.cloningMu.Lock()
					delete(s.cloning, dir)
					s.cloningMu.Unlock()
				}()

				cmd := exec.CommandContext(ctx, "git", "clone", "--mirror", origin, dir)
				if output, err := s.runWithRemoteOpts(cmd, req.Opt); err != nil {
					log15.Error("clone failed", "error", err, "output", string(output))
					return
				}
			}()

			req.ReplyChan <- &protocol.LegacyExecReply{CloneInProgress: true}
			status = "clone-in-progress"
			return
		}

		s.cloningMu.Unlock()
		req.ReplyChan <- &protocol.LegacyExecReply{RepoNotFound: true}
		status = "repo-not-found"
		return
	}
	s.cloningMu.Unlock()

	if req.EnsureRevision != "" {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", req.EnsureRevision)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			// Revision not found, update before running actual command.
			s.doRepoUpdate(ctx, req.Repo, req.Opt)
		}
	}

	stdoutC, stdoutWRaw := chanrpcutil.NewWriter()
	stderrC, stderrWRaw := chanrpcutil.NewWriter()
	stdoutW := &writeCounter{w: stdoutWRaw}
	stderrW := &writeCounter{w: stderrWRaw}

	cmd := exec.CommandContext(ctx, "git", req.Args...)
	cmd.Dir = dir
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	processResultChan := make(chan *protocol.ProcessResult, 1)
	req.ReplyChan <- &protocol.LegacyExecReply{
		Stdout:        stdoutC,
		Stderr:        stderrC,
		ProcessResult: processResultChan,
	}

	var err error
	err, exitStatus = runCommand(cmd)
	if err != nil {
		errStr = err.Error()
	}

	stdoutW.Close()
	stderrW.Close()

	processResultChan <- &protocol.ProcessResult{
		Error:      errStr,
		ExitStatus: exitStatus,
	}
	close(processResultChan)
	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	if !req.NoAutoUpdate && !noUpdates {
		s.updateRepo <- updateRepoRequest{req.Repo, req.Opt}
	}
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
	lastCheckAt := make(map[string]time.Time)

	go func() {
		for req := range updateRepo {
			if t, ok := lastCheckAt[req.repo]; ok && time.Now().Before(t.Add(10*time.Second)) {
				continue // git data still fresh
			}
			lastCheckAt[req.repo] = time.Now()
			go func(req updateRepoRequest) {
				// Create a new context with a new timeout (instead of passing one through updateRepoRequest)
				// because the ctx of the updateRepoRequest sender will get cancelled before this goroutine runs.
				ctx, cancel := context.WithTimeout(context.Background(), longGitCommandTimeout)
				defer cancel()
				s.doRepoUpdate(ctx, req.repo, req.opt)
			}(req)
		}
	}()

	return updateRepo
}

var headBranchPattern = regexp.MustCompile("HEAD branch: (.+?)\\n")

func (s *Server) doRepoUpdate(ctx context.Context, repo string, opt *vcs.RemoteOpts) {
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

		s.doRepoUpdate2(ctx, repo, opt)
	})
}

func (s *Server) doRepoUpdate2(ctx context.Context, repo string, opt *vcs.RemoteOpts) {
	cmd := exec.CommandContext(ctx, "git", "remote", "update", "--prune")
	cmd.Dir = path.Join(s.ReposDir, repo)
	if output, err := s.runWithRemoteOpts(cmd, opt); err != nil {
		log15.Error("Failed to update", "repo", repo, "error", err, "output", string(output))
		return
	}

	headBranch := "master"

	// try to fetch HEAD from origin
	cmd = exec.CommandContext(ctx, "git", "remote", "show", "origin")
	cmd.Dir = path.Join(s.ReposDir, repo)
	output, err := s.runWithRemoteOpts(cmd, opt)
	if err != nil {
		log15.Error("Failed to fetch remote info", "repo", repo, "error", err, "output", string(output))
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
	cmd = exec.CommandContext(ctx, "git", "rev-parse", headBranch)
	cmd.Dir = path.Join(s.ReposDir, repo)
	if err := cmd.Run(); err != nil {
		// branch does not exist, pick first branch
		cmd := exec.CommandContext(ctx, "git", "branch")
		cmd.Dir = path.Join(s.ReposDir, repo)
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
	cmd.Dir = path.Join(s.ReposDir, repo)
	if output, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to set HEAD", "repo", repo, "error", err, "output", string(output))
		return
	}
}
