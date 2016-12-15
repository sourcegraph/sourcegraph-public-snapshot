package gitserver

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"log"
	"time"

	"github.com/neelance/chanrpc"
	"github.com/neelance/chanrpc/chanrpcutil"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// DefaultClient is the default Client. It is not connected to any gitservers by default.
var DefaultClient = new(Client)

// Client is a gitserver client.
type Client struct {
	servers [](chan<- *request)
}

// Connect connects the client to a remote gitserver.
// It can be called multiple times to connect to a gitserver cluster.
func (c *Client) Connect(addr string) {
	requestsChan := make(chan *request, 100)
	c.servers = append(c.servers, requestsChan)

	go func() {
		for {
			err := chanrpc.DialAndDeliver(addr, requestsChan)
			log.Printf("gitserver: DialAndDeliver error: %v", err)
			time.Sleep(time.Second)
		}
	}()
}

func (c *Cmd) sendExec(ctx context.Context) (_ *execReply, errRes error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Client.sendExec")
	defer func() {
		if errRes != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", errRes.Error())
		}
		span.Finish()
	}()
	span.SetTag("request", "Exec")
	span.SetTag("repo", c.Repo)
	span.SetTag("args", c.Args[1:])
	span.SetTag("opt", c.Opt)

	// Check that ctx is not expired.
	if err := ctx.Err(); err != nil {
		deadlineExceededCounter.Inc()
		return nil, err
	}

	sum := md5.Sum([]byte(c.Repo))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(c.client.servers))
	replyChan := make(chan *execReply, 1)
	c.client.servers[serverIndex] <- &request{Exec: &execRequest{Repo: c.Repo, Args: c.Args[1:], Opt: c.Opt, Stdin: chanrpcutil.ToChunks(c.Input), ReplyChan: replyChan}}
	reply, ok := <-replyChan
	if !ok {
		return nil, errRPCFailed
	}

	if !reply.repoFound() {
		return nil, vcs.RepoNotExistError{}
	}

	return reply, nil
}

var errRPCFailed = errors.New("gitserver: rpc failed")

var deadlineExceededCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "client_deadline_exceeded",
	Help:      "Times that Client.sendExec() returned context.DeadlineExceeded",
})

func init() {
	prometheus.MustRegister(deadlineExceededCounter)
}

// Cmd represents a command to be executed remotely.
type Cmd struct {
	client *Client

	Args       []string
	Repo       string
	Opt        *vcs.RemoteOpts
	Input      []byte
	ExitStatus int
}

// Command creates a new Cmd. Command name must be 'git',
// otherwise it panics.
func (c *Client) Command(name string, arg ...string) *Cmd {
	if name != "git" {
		panic("gitserver: command name must be 'git'")
	}
	return &Cmd{
		client: c,
		Args:   append([]string{"git"}, arg...),
	}
}

// DividedOutput runs the command and returns its standard output and standard error.
func (c *Cmd) DividedOutput(ctx context.Context) ([]byte, []byte, error) {
	reply, err := c.sendExec(ctx)
	if err != nil {
		return nil, nil, err
	}

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

// Run starts the specified command and waits for it to complete.
func (c *Cmd) Run(ctx context.Context) error {
	_, _, err := c.DividedOutput(ctx)
	return err
}

// Output runs the command and returns its standard output.
func (c *Cmd) Output(ctx context.Context) ([]byte, error) {
	stdout, _, err := c.DividedOutput(ctx)
	return stdout, err
}

// CombinedOutput runs the command and returns its combined standard output and standard error.
func (c *Cmd) CombinedOutput(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := c.DividedOutput(ctx)
	return append(stdout, stderr...), err
}

// Init creates a new empty repository remotely.
func (c *Client) Init(ctx context.Context, repo string) error {
	return c.create(ctx, repo, "", nil)
}

// Clone performs a clone operation remotely.
func (c *Client) Clone(ctx context.Context, repo string, remote string, opt *vcs.RemoteOpts) error {
	if remote == "" {
		return errors.New("empty remote")
	}
	return c.create(ctx, repo, remote, opt)
}

// create creates a new repository in the gitserver cluster by initializing an empty repository
// if mirrorRemote is empty or clones the given remote otherwise, using opt for authentication.
// The gitserver is selected pseudo-randomly.
//
// A nil error is returned if the new repository was created successfully, but vcs.ErrRepoExist
// is returned if there was already an existing repository in its place, causing create to be noop.
// If the repository is in process of being cloned, vcs.RepoNotExistError{CloneInProgress: true} is returned.
func (c *Client) create(ctx context.Context, repo string, mirrorRemote string, opt *vcs.RemoteOpts) error {
	// "github.com/sourcegraphtest/AlwaysCloningTest" is a special repository for triggering
	// an infinite "Cloning this repository" response. It exists to aid development and testing
	// of various features that need to handle the case of repositories being cloned.
	if repo == "github.com/sourcegraphtest/AlwaysCloningTest" {
		return vcs.RepoNotExistError{CloneInProgress: true}
	}

	sum := md5.Sum([]byte(repo))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(c.servers))

	replyChan := make(chan *createReply, 1)
	c.servers[serverIndex] <- &request{Create: &createRequest{
		Repo:         repo,
		MirrorRemote: mirrorRemote,
		Opt:          opt,
		ReplyChan:    replyChan,
	}}

	reply, ok := <-replyChan
	if !ok {
		return errors.New("gitserver: create failed")
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	if reply.CloneInProgress {
		return vcs.RepoNotExistError{CloneInProgress: true}
	}
	if reply.RepoExist {
		return vcs.ErrRepoExist
	}
	// Repo did not exist and was successfully created.
	return nil
}
