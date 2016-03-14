package gitserver

import (
	"bytes"
	"errors"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"path"
	"syscall"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"
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

func (g *Git) Exec(args *ExecArgs, reply *ExecReply) error {
	dir := path.Join(ReposDir, args.Repo)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	reply.RepoExists = true

	cmd := exec.Command("git", args.Args...)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(args.Stdin)
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
	done := make(chan *rpc.Call, len(servers))
	for _, server := range servers {
		server <- &rpc.Call{
			ServiceMethod: "Git.Exec",
			Args:          &ExecArgs{Repo: c.Repo, Args: c.Args[1:], Opt: c.Opt, Stdin: c.Input},
			Reply:         &ExecReply{},
			Done:          done,
		}
	}
	var rpcError error
	for range servers {
		call := <-done
		if call.Error != nil {
			rpcError = call.Error
			continue
		}
		reply := call.Reply.(*ExecReply)
		if reply.RepoExists {
			var err error
			if reply.Error != "" {
				err = errors.New(reply.Error)
			}
			c.ExitStatus = reply.ExitStatus
			return reply.Stdout, reply.Stderr, err
		}
	}
	if rpcError != nil {
		return nil, nil, rpcError
	}
	return nil, nil, vcs.ErrRepoNotExist
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
