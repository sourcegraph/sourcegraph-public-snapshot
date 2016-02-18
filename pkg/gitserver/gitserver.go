package gitserver

import (
	"bytes"
	"errors"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"syscall"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

type Git struct {
}

type ExecArgs struct {
	Repo string
	Args []string
	Opt  *vcs.RemoteOpts
}

type ExecReply struct {
	RepoExists bool
	Error      string
	ExitStatus int
	Stdout     []byte
	Stderr     []byte
}

var clientSingleton *rpc.Client

func RegisterHandler() {
	rpc.Register(&Git{})
	rpc.HandleHTTP()
}

func (g *Git) Exec(args *ExecArgs, reply *ExecReply) error {
	if _, err := os.Stat(args.Repo); args.Repo != "" && os.IsNotExist(err) {
		return nil
	}
	reply.RepoExists = true

	cmd := exec.Command("git", args.Args...)
	cmd.Dir = args.Repo
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if args.Opt != nil && args.Opt.SSH != nil {
		gitSSHWrapper, gitSSHWrapperDir, keyFile, err := makeGitSSHWrapper(args.Opt.SSH.PrivateKey)
		defer func() {
			if keyFile != "" {
				if err := os.Remove(keyFile); err != nil {
					log.Fatalf("Error removing SSH key file %s: %s.", keyFile, err)
				}
			}
		}()
		if err != nil {
			return err
		}
		defer os.Remove(gitSSHWrapper)
		if gitSSHWrapperDir != "" {
			defer os.RemoveAll(gitSSHWrapperDir)
		}
		cmd.Env = []string{"GIT_SSH=" + gitSSHWrapper}
	}

	if args.Opt != nil && args.Opt.HTTPS != nil {
		env := environ(os.Environ())
		env.Unset("GIT_TERMINAL_PROMPT")

		gitPassHelper, gitPassHelperDir, err := makeGitPassHelper(args.Opt.HTTPS.Pass)
		if err != nil {
			return err
		}
		defer os.Remove(gitPassHelper)
		if gitPassHelperDir != "" {
			defer os.RemoveAll(gitPassHelperDir)
		}
		env.Unset("GIT_ASKPASS")
		env = append(env, "GIT_ASKPASS="+gitPassHelper)

		cmd.Env = env
	}

	if err := cmd.Run(); err != nil {
		reply.Error = err.Error()
	}
	if cmd.ProcessState != nil { // is nil if process failed to start
		reply.ExitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	reply.Stdout = stdoutBuf.Bytes()
	reply.Stderr = stderrBuf.Bytes()
	return nil
}

func Dial(addr string) error {
	var err error
	clientSingleton, err = rpc.DialHTTP("tcp", addr)
	return err
}

type Cmd struct {
	Args       []string
	Dir        string
	Opt        *vcs.RemoteOpts
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
	var reply ExecReply
	if err := clientSingleton.Call("Git.Exec", &ExecArgs{Repo: c.Dir, Args: c.Args[1:], Opt: c.Opt}, &reply); err != nil {
		return nil, nil, err
	}
	if !reply.RepoExists {
		return nil, nil, vcs.ErrRepoNotExist
	}
	var err error
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
