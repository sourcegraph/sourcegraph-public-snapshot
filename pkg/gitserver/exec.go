package gitserver

import (
	"bytes"
	"errors"
	"os/exec"
	"path"
	"syscall"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type execRequest struct {
	Repo      string
	Args      []string
	Opt       *vcs.RemoteOpts
	Stdin     []byte
	ReplyChan chan<- *execReply
}

type execReply struct {
	RepoNotExist bool
	Error        string
	ExitStatus   int
	Stdout       []byte
	Stderr       []byte
}

func (r *execReply) repoNotExist() bool {
	return r.RepoNotExist
}

func handleExecRequest(req *execRequest) {
	defer recoverAndLog()
	defer close(req.ReplyChan)

	dir := path.Join(ReposDir, req.Repo)
	if !repoExists(dir) {
		req.ReplyChan <- &execReply{RepoNotExist: true}
		return
	}

	cmd := exec.Command("git", req.Args...)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(req.Stdin)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	var errStr string
	var exitStatus int
	if err := runWithRemoteOpts(cmd, req.Opt); err != nil {
		errStr = err.Error()
	}
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}

	req.ReplyChan <- &execReply{
		Error:      errStr,
		ExitStatus: exitStatus,
		Stdout:     stdoutBuf.Bytes(),
		Stderr:     stderrBuf.Bytes(),
	}
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
		return &request{Exec: &execRequest{Repo: c.Repo, Args: c.Args[1:], Opt: c.Opt, Stdin: c.Input, ReplyChan: replyChan}},
			func() (genericReply, bool) { reply, ok := <-replyChan; return reply, ok }
	})
	if err != nil {
		return nil, nil, err
	}

	reply := genReply.(*execReply)
	if reply.Error != "" {
		err = errors.New(reply.Error)
	}
	c.ExitStatus = reply.ExitStatus
	return reply.Stdout, reply.Stderr, err
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
