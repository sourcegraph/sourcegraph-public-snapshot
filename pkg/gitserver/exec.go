package gitserver

import (
	"errors"
	"os/exec"
	"path"
	"syscall"

	"github.com/neelance/chanrpc/chanrpcutil"

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
	RepoNotExist  bool
	Stdout        <-chan []byte
	Stderr        <-chan []byte
	ProcessResult <-chan *processResult
}

type processResult struct {
	Error      string
	ExitStatus int
}

func (r *execReply) repoNotExist() bool {
	return r.RepoNotExist
}

func handleExecRequest(req *execRequest) {
	defer recoverAndLog()
	defer close(req.ReplyChan)

	dir := path.Join(ReposDir, req.Repo)
	if !repoExists(dir) {
		chanrpcutil.Drain(req.Stdin)
		req.ReplyChan <- &execReply{RepoNotExist: true}
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
	stdout := chanrpcutil.ReadAll(reply.Stdout)
	stderr := chanrpcutil.ReadAll(reply.Stderr)

	processResult := <-reply.ProcessResult
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
