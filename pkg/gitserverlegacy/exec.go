package gitserverlegacy

import (
	"bytes"
	"errors"
	"os/exec"
	"path"
	"syscall"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type ExecArgs struct {
	Repo  string
	Args  []string
	Opt   *vcs.RemoteOpts
	Stdin []byte
}

type ExecReply struct {
	RepoExists bool
	Error      string
	ExitStatus int
	Stdout     []byte
	Stderr     []byte
}

func (r *ExecReply) repoExists() bool {
	return r.RepoExists
}

func (g *Git) Exec(args *ExecArgs, reply *ExecReply) error {
	dir := path.Join(ReposDir, args.Repo)
	if !repoExists(dir) {
		return nil
	}
	reply.RepoExists = true

	cmd := exec.Command("git", args.Args...)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(args.Stdin)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := runWithRemoteOpts(cmd, args.Opt); err != nil {
		reply.Error = err.Error()
	}
	if cmd.ProcessState != nil { // is nil if process failed to start
		reply.ExitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	reply.Stdout = stdoutBuf.Bytes()
	reply.Stderr = stderrBuf.Bytes()
	return nil
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
	rawReply, err := broadcastCall(
		"Git.Exec",
		&ExecArgs{Repo: c.Repo, Args: c.Args[1:], Opt: c.Opt, Stdin: c.Input},
		func() repoExistsReply { return new(ExecReply) },
	)
	if err != nil {
		return nil, nil, err
	}
	reply := rawReply.(*ExecReply)
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
