package gitserver

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/neelance/chanrpc"
	"github.com/neelance/chanrpc/chanrpcutil"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

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
}

// Serve serves incoming gitserver requests on listener l.
func (s *Server) Serve(l net.Listener) error {
	s.cloning = make(map[string]struct{})

	s.registerMetrics()
	requests := make(chan *request, 100)
	go s.processRequests(requests)
	srv := &chanrpc.Server{RequestChan: requests}
	return srv.Serve(l)
}

func (s *Server) processRequests(requests <-chan *request) {
	for req := range requests {
		if req.Exec != nil {
			go s.handleExecRequest(req.Exec)
		}
		if req.Search != nil {
			go s.handleSearchRequest(req.Search)
		}
		if req.Create != nil {
			go s.handleCreateRequest(req.Create)
		}
		if req.Remove != nil {
			go s.handleRemoveRequest(req.Remove)
		}
	}
}

// handleExecRequest handles a exec request.
func (s *Server) handleExecRequest(req *execRequest) {
	start := time.Now()
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)

	// Instrumentation
	{
		repo := repotrackutil.GetTrackedRepo(req.Repo)
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		execRunning.WithLabelValues(cmd, repo).Inc()
		defer func() {
			execRunning.WithLabelValues(cmd, repo).Dec()
			execDuration.WithLabelValues(cmd, repo, status).Observe(time.Since(start).Seconds())
		}()
	}

	dir := path.Join(s.ReposDir, req.Repo)
	s.cloningMu.Lock()
	_, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if cloneInProgress {
		chanrpcutil.Drain(req.Stdin)
		req.ReplyChan <- &execReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if !repoExists(dir) {
		chanrpcutil.Drain(req.Stdin)
		req.ReplyChan <- &execReply{RepoNotFound: true}
		status = "repo-not-found"
		return
	}

	stdoutC, stdoutW := chanrpcutil.NewWriter()
	stderrC, stderrW := chanrpcutil.NewWriter()

	cmd := exec.Command("git", req.Args...)
	cmd.Dir = dir
	cmd.Stdin = chanrpcutil.NewReader(req.Stdin)
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	processResultChan := make(chan *processResult, 1)
	req.ReplyChan <- &execReply{
		Stdout:        stdoutC,
		Stderr:        stderrC,
		ProcessResult: processResultChan,
	}

	var errStr string
	var exitStatus int
	if err := s.runWithRemoteOpts(cmd, req.Opt); err != nil {
		errStr = err.Error()
	}
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}

	chanrpcutil.Drain(req.Stdin)
	stdoutW.Close()
	stderrW.Close()

	processResultChan <- &processResult{
		Error:      errStr,
		ExitStatus: exitStatus,
	}
	close(processResultChan)
	status = strconv.Itoa(exitStatus)
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
	Buckets:   statsutil.UserLatencyBuckets,
}, []string{"cmd", "repo", "status"})

func init() {
	prometheus.MustRegister(execRunning)
	prometheus.MustRegister(execDuration)
}

// handleSearchRequest handles a search request.
func (s *Server) handleSearchRequest(req *searchRequest) {
	start := time.Now()
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)
	defer func() { defer observeSearch(req, start, status) }()

	dir := path.Join(s.ReposDir, req.Repo)
	s.cloningMu.Lock()
	_, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if cloneInProgress {
		req.ReplyChan <- &searchReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if !repoExists(dir) {
		req.ReplyChan <- &searchReply{RepoNotFound: true}
		status = "repo-not-found"
		return
	}

	var queryType string
	switch req.Opt.QueryType {
	case vcs.FixedQuery:
		queryType = "--fixed-strings"
	default:
		req.ReplyChan <- &searchReply{Error: fmt.Sprintf("unrecognized QueryType: %q", req.Opt.QueryType)}
		status = "error"
		return
	}

	cmd := exec.Command("git", "grep", "--null", "--line-number", "-I", "--no-color", "--context", strconv.Itoa(int(req.Opt.ContextLines)), queryType, "-e", req.Opt.Query, string(req.Commit))
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		req.ReplyChan <- &searchReply{Error: err.Error()}
		status = "error"
		return
	}
	defer out.Close()
	if err := cmd.Start(); err != nil {
		req.ReplyChan <- &searchReply{Error: err.Error()}
		status = "error"
		return
	}

	var results []*vcs.SearchResult
	errc := make(chan error)
	go func() {
		rd := bufio.NewReader(out)
		var r *vcs.SearchResult
		addResult := func(rr *vcs.SearchResult) bool {
			if rr != nil {
				if req.Opt.Offset == 0 {
					results = append(results, rr)
				} else {
					req.Opt.Offset--
				}
				r = nil
			}
			// Return true if no more need to be added.
			return len(results) == int(req.Opt.N)
		}
		for {
			line, err := rd.ReadBytes('\n')
			if err == io.EOF {
				// git-grep output ends with a newline, so if we hit EOF, there's nothing left to
				// read
				break
			} else if err != nil {
				errc <- err
				return
			}
			// line is guaranteed to be '\n' terminated according to the contract of ReadBytes
			line = line[0 : len(line)-1]

			if bytes.Equal(line, []byte("--")) {
				// Match separator.
				if addResult(r) {
					break
				}
			} else {
				// Match line looks like: "HEAD:filename\x00lineno\x00matchline\n".
				fileEnd := bytes.Index(line, []byte{'\x00'})
				file := string(line[len(req.Commit)+1 : fileEnd])
				lineNoStart, lineNoEnd := fileEnd+1, fileEnd+1+bytes.Index(line[fileEnd+1:], []byte{'\x00'})
				lineNo, err := strconv.Atoi(string(line[lineNoStart:lineNoEnd]))
				if err != nil {
					panic("bad line number on line: " + string(line) + ": " + err.Error())
				}
				if r == nil || r.File != file {
					if r != nil {
						if addResult(r) {
							break
						}
					}
					r = &vcs.SearchResult{File: file, StartLine: uint32(lineNo)}
				}
				r.EndLine = uint32(lineNo)
				if r.Match != nil {
					r.Match = append(r.Match, '\n')
				}
				r.Match = append(r.Match, line[lineNoEnd+1:]...)
			}
		}
		addResult(r)

		if err := cmd.Process.Kill(); err != nil {
			if runtime.GOOS != "windows" {
				errc <- err
				return
			}
		}
		if err := cmd.Wait(); err != nil {
			if c := exitStatus(err); c != -1 && c != 1 {
				// -1 exit code = killed (by cmd.Process.Kill() call
				// above), 1 exit code means grep had no match (but we
				// don't translate that to a Go error)
				errc <- fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
				return
			}
		}
		errc <- nil
	}()

	err = <-errc
	cmd.Process.Kill()
	if err != nil {
		req.ReplyChan <- &searchReply{Error: err.Error()}
		status = "error"
		return
	}

	req.ReplyChan <- &searchReply{
		Results: results,
	}
	status = "success"
}

func exitStatus(err error) int {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// There is no platform independent way to retrieve
			// the exit code, but the following will work on Unix
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		return 0
	}
	return 0
}

var searchDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "search_duration_seconds",
	Help:      "gitserver.Search latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, []string{"query_type", "repo", "status"})

func init() {
	prometheus.MustRegister(searchDuration)
}

func observeSearch(req *searchRequest, start time.Time, status string) {
	repo := repotrackutil.GetTrackedRepo(req.Repo)
	searchDuration.WithLabelValues(req.Opt.QueryType, repo, status).Observe(time.Since(start).Seconds())
}

// handleCreateRequest handles a create request.
func (s *Server) handleCreateRequest(req *createRequest) {
	start := time.Now()
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)
	defer func() { defer observeCreate(start, status) }()

	dir := path.Join(s.ReposDir, req.Repo)
	s.cloningMu.Lock()
	if _, ok := s.cloning[dir]; ok {
		s.cloningMu.Unlock()
		req.ReplyChan <- &createReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if repoExists(dir) {
		s.cloningMu.Unlock()
		req.ReplyChan <- &createReply{RepoExist: true}
		status = "repo-exists"
		return
	}

	// We'll take this repo and start cloning it.
	// Mark it as being cloned so no one else starts to.
	s.cloning[dir] = struct{}{}
	s.cloningMu.Unlock()

	defer func() {
		s.cloningMu.Lock()
		delete(s.cloning, dir)
		s.cloningMu.Unlock()
	}()

	if req.MirrorRemote != "" {
		cmd := exec.Command("git", "clone", "--mirror", req.MirrorRemote, dir)

		var outputBuf bytes.Buffer
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
		if err := s.runWithRemoteOpts(cmd, req.Opt); err != nil {
			req.ReplyChan <- &createReply{Error: fmt.Sprintf("cloning repository %s failed with output:\n%s", req.Repo, outputBuf.String())}
			status = "clone-fail"
			return
		}
		req.ReplyChan <- &createReply{}
		status = "clone-success"
		return
	}

	cmd := exec.Command("git", "init", "--bare", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		req.ReplyChan <- &createReply{Error: fmt.Sprintf("initializing repository %s failed with output:\n%s", req.Repo, string(out))}
		status = "init-fail"
		return
	}
	status = "init-success"
	req.ReplyChan <- &createReply{}
}

var createDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "create_duration_seconds",
	Help:      "gitserver.Init and gitserver.Clone latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, []string{"status"})

func init() {
	prometheus.MustRegister(createDuration)
}

func observeCreate(start time.Time, status string) {
	createDuration.WithLabelValues(status).Observe(time.Since(start).Seconds())
}

// handleRemoveRequest handles a remove request.
func (s *Server) handleRemoveRequest(req *removeRequest) {
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)
	defer func() { defer observeRemove(status) }()

	dir := path.Join(s.ReposDir, req.Repo)
	s.cloningMu.Lock()
	_, cloneInProgress := s.cloning[dir]
	s.cloningMu.Unlock()
	if cloneInProgress {
		req.ReplyChan <- &removeReply{CloneInProgress: true}
		status = "clone-in-progress"
		return
	}
	if !repoExists(dir) {
		req.ReplyChan <- &removeReply{RepoNotFound: true}
		status = "repo-not-found"
		return
	}

	cmd := exec.Command("git", "remote")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		req.ReplyChan <- &removeReply{Error: fmt.Sprintf("not a repository: %s", req.Repo)}
		status = "not-a-repository"
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		req.ReplyChan <- &removeReply{Error: err.Error()}
		status = "failed"
		return
	}
	req.ReplyChan <- &removeReply{}
	status = "success"
}

// Remove should be pretty much instant, so we just track counts.
var removeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "remove_total",
	Help:      "Total calls to gitserver.Remove",
}, []string{"status"})

func init() {
	prometheus.MustRegister(removeCounter)
}

func observeRemove(status string) {
	removeCounter.WithLabelValues(status).Inc()
}
