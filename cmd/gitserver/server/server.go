package server

import (
	"log"
	"net"
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
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type originMapEntry struct {
	Prefix string
	Origin string
}

var originMap = []*originMapEntry{
	{Prefix: "github.com/", Origin: "https://github.com/%.git"},
}

func init() {
	var providedOriginMap []*originMapEntry
	for _, e := range strings.Fields(env.Get("ORIGIN_MAP", "", `space separated list of mappings from repo name prefix to origin url, for example "github.com/!https://github.com/%.git"`)) {
		p := strings.Split(e, "!")
		if len(p) != 2 {
			log.Fatalf("invalid ORIGIN_MAP entry: %s", e)
		}
		providedOriginMap = append(providedOriginMap, &originMapEntry{Prefix: p[0], Origin: p[1]})
	}
	originMap = append(providedOriginMap, originMap...)
}

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

// Serve serves incoming gitserver requests on listener l.
func (s *Server) Serve(l net.Listener) error {
	if err := initializeSSH(); err != nil {
		log.Printf("SSH initialization error: %s", err)
	}

	s.cloning = make(map[string]struct{})
	s.updateRepo = s.repoUpdateLoop()
	s.repoUpdateLocks = make(map[string]*locks)

	s.registerMetrics()
	requests := make(chan *protocol.Request, 100)
	go s.processRequests(requests)
	srv := &chanrpc.Server{RequestChan: requests}
	return srv.Serve(l)
}

func (s *Server) processRequests(requests <-chan *protocol.Request) {
	for req := range requests {
		if req.Exec != nil {
			go s.handleExecRequest(req.Exec)
		}
	}
}

// handleExecRequest handles a exec request.
func (s *Server) handleExecRequest(req *protocol.ExecRequest) {
	start := time.Now()
	exitStatus := -10810 // sentinel value to indicate not set
	var stdoutN, stderrN int64
	var status string
	var errStr string

	defer recoverAndLog()
	defer close(req.ReplyChan)

	req.Repo = protocol.NormalizeRepo(req.Repo)

	// This is a repo that we use for testing the cloning state of the UI
	if req.Repo == "github.com/sourcegraphtest/alwayscloningtest" {
		chanrpcutil.Drain(req.Stdin)
		req.ReplyChan <- &protocol.ExecReply{CloneInProgress: true}
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
		chanrpcutil.Drain(req.Stdin)
		req.ReplyChan <- &protocol.ExecReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if !repoExists(dir) {
		var origin string
		for _, entry := range originMap {
			if strings.HasPrefix(req.Repo, entry.Prefix) {
				origin = strings.Replace(entry.Origin, "%", strings.TrimPrefix(req.Repo, entry.Prefix), 1)
				break
			}
		}
		if origin != "" && !req.NoAutoUpdate {
			s.cloning[dir] = struct{}{} // Mark this repo as currently being cloned.
			s.cloningMu.Unlock()

			go func() {
				defer func() {
					s.cloningMu.Lock()
					delete(s.cloning, dir)
					s.cloningMu.Unlock()
				}()

				cmd := exec.Command("git", "clone", "--mirror", origin, dir)
				if output, err := s.runWithRemoteOpts(cmd, req.Opt); err != nil {
					log15.Error("clone failed", "error", err, "output", string(output))
					return
				}
			}()

			chanrpcutil.Drain(req.Stdin)
			req.ReplyChan <- &protocol.ExecReply{CloneInProgress: true}
			status = "clone-in-progress"
			return
		}

		s.cloningMu.Unlock()
		chanrpcutil.Drain(req.Stdin)
		req.ReplyChan <- &protocol.ExecReply{RepoNotFound: true}
		status = "repo-not-found"
		return
	}
	s.cloningMu.Unlock()

	if req.EnsureRevision != "" {
		cmd := exec.Command("git", "rev-parse", req.EnsureRevision)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			// Revision not found, update before running actual command.
			s.doRepoUpdate(req.Repo, req.Opt)
		}
	}

	stdoutC, stdoutWRaw := chanrpcutil.NewWriter()
	stderrC, stderrWRaw := chanrpcutil.NewWriter()
	stdoutW := &writeCounter{w: stdoutWRaw}
	stderrW := &writeCounter{w: stderrWRaw}

	cmd := exec.Command("git", req.Args...)
	cmd.Dir = dir
	cmd.Stdin = chanrpcutil.NewReader(req.Stdin)
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	processResultChan := make(chan *protocol.ProcessResult, 1)
	req.ReplyChan <- &protocol.ExecReply{
		Stdout:        stdoutC,
		Stderr:        stderrC,
		ProcessResult: processResultChan,
	}

	if err := cmd.Run(); err != nil {
		errStr = err.Error()
	}
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}

	chanrpcutil.Drain(req.Stdin)
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

	if !req.NoAutoUpdate {
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

			go s.doRepoUpdate(req.repo, req.opt)
		}
	}()

	return updateRepo
}

var headBranchPattern = regexp.MustCompile("HEAD branch: (.+?)\\n")

func (s *Server) doRepoUpdate(repo string, opt *vcs.RemoteOpts) {
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

		s.doRepoUpdate2(repo, opt)
	})
}

func (s *Server) doRepoUpdate2(repo string, opt *vcs.RemoteOpts) {
	cmd := exec.Command("git", "remote", "update", "--prune")
	cmd.Dir = path.Join(s.ReposDir, repo)
	if output, err := s.runWithRemoteOpts(cmd, opt); err != nil {
		log15.Error("Failed to update", "repo", repo, "error", err, "output", string(output))
		return
	}

	headBranch := "master"

	// try to fetch HEAD from origin
	cmd = exec.Command("git", "remote", "show", "origin")
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
	cmd = exec.Command("git", "rev-parse", headBranch)
	cmd.Dir = path.Join(s.ReposDir, repo)
	if err := cmd.Run(); err != nil {
		// branch does not exist, pick first branch
		cmd := exec.Command("git", "branch")
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
	cmd = exec.Command("git", "symbolic-ref", "HEAD", "refs/heads/"+headBranch)
	cmd.Dir = path.Join(s.ReposDir, repo)
	if output, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to set HEAD", "repo", repo, "error", err, "output", string(output))
		return
	}
}
