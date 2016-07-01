package gitserver

import (
	"errors"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/neelance/chanrpc/chanrpcutil"
	"github.com/prometheus/client_golang/prometheus"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type execRequest struct {
	Repo      string
	Args      []string
	Opt       *vcs.RemoteOpts
	Stdin     <-chan []byte
	ReplyChan chan<- *execReply
}

type execReply struct {
	RepoNotFound    bool // If true, exec returned with noop because repo is not found.
	CloneInProgress bool // If true, exec returned with noop because clone is in progress.
	Stdout          <-chan []byte
	Stderr          <-chan []byte
	ProcessResult   <-chan *processResult
}

func (r *execReply) repoFound() bool { return !r.RepoNotFound }

type processResult struct {
	Error      string
	ExitStatus int
}

func handleExecRequest(req *execRequest) {
	start := time.Now()
	status := ""

	defer recoverAndLog()
	defer close(req.ReplyChan)
	defer func() { observeExec(req, start, status) }()

	dir := path.Join(ReposDir, req.Repo)
	cloningMu.Lock()
	_, cloneInProgress := cloning[dir]
	cloningMu.Unlock()
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
	if err := runWithRemoteOpts(cmd, req.Opt); err != nil {
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

type Cmd struct {
	Args       []string
	Repo       string
	Opt        *vcs.RemoteOpts
	Input      []byte
	ExitStatus int
}

func Command(name string, arg ...string) *Cmd {
	if name != "git" {
		panic("gitserver: command name must be 'git'")
	}
	return &Cmd{
		Args: append([]string{"git"}, arg...),
	}
}

func (c *Cmd) DividedOutput() ([]byte, []byte, error) {
	genReply, err := broadcastCall(func() (*request, func() (genericReply, bool)) {
		replyChan := make(chan *execReply, 1)
		return &request{Exec: &execRequest{Repo: c.Repo, Args: c.Args[1:], Opt: c.Opt, Stdin: chanrpcutil.ToChunks(c.Input), ReplyChan: replyChan}},
			func() (genericReply, bool) { reply, ok := <-replyChan; return reply, ok }
	})
	if err != nil {
		return nil, nil, err
	}

	reply := genReply.(*execReply)
	if reply.CloneInProgress {
		return nil, nil, vcs.RepoNotExistError{CloneInProgress: true}
	}
	stdout := chanrpcutil.ReadAll(reply.Stdout)
	stderr := chanrpcutil.ReadAll(reply.Stderr)

	processResult, ok := <-reply.ProcessResult
	if !ok {
		return nil, nil, errors.New("connection to gitserver lost")
	}
	if processResult.Error != "" {
		err = errors.New(processResult.Error)
	}
	c.ExitStatus = processResult.ExitStatus

	return <-stdout, <-stderr, err
}

func (c *Cmd) Run() error {
	_, _, err := c.DividedOutput()
	return err
}

func (c *Cmd) Output() ([]byte, error) {
	stdout, _, err := c.DividedOutput()
	return stdout, err
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	stdout, stderr, err := c.DividedOutput()
	return append(stdout, stderr...), err
}

var execDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "exec_duration_seconds",
	Help:      "gitserver.Command latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, []string{"cmd", "repo", "status"})

func init() {
	prometheus.MustRegister(execDuration)
}

func observeExec(req *execRequest, start time.Time, status string) {
	repo := repotrackutil.GetTrackedRepo(req.Repo)
	cmd := ""
	if len(req.Args) > 0 {
		cmd = req.Args[0]
	}
	execDuration.WithLabelValues(cmd, repo, status).Observe(time.Since(start).Seconds())
}
