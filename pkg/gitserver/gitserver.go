package gitserver

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os/exec"
	"sync"
	"syscall"
)

type Git struct {
}

type ExecArgs struct {
	Repo string
	Args []string
}

type ExecReply struct {
	Error      string
	ExitStatus int
	Stdout     []byte
	Stderr     []byte
}

var clientSingleton *rpc.Client
var clientSingletonMutex sync.Mutex

func ListenAndServe() error {
	rpc.Register(&Git{})
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":3081")
	if e != nil {
		return e
	}
	log.Print("Git server listening on :3081")
	return http.Serve(l, nil)
}

func (g *Git) Exec(args *ExecArgs, reply *ExecReply) error {
	cmd := exec.Command("git", args.Args...)
	cmd.Dir = args.Repo
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		reply.Error = err.Error()
	}
	reply.ExitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	reply.Stdout = stdoutBuf.Bytes()
	reply.Stderr = stderrBuf.Bytes()
	return nil
}

func client() (*rpc.Client, error) {
	clientSingletonMutex.Lock()
	defer clientSingletonMutex.Unlock()
	if clientSingleton == nil {
		var err error
		clientSingleton, err = rpc.DialHTTP("tcp", ":3081")
		if err != nil {
			return nil, err
		}
	}
	return clientSingleton, nil
}

type Cmd struct {
	Args       []string
	Dir        string
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
	cl, err := client()
	if err != nil {
		return nil, nil, err
	}
	var reply ExecReply
	if err = cl.Call("Git.Exec", &ExecArgs{Repo: c.Dir, Args: c.Args[1:]}, &reply); err != nil {
		return nil, nil, err
	}
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
