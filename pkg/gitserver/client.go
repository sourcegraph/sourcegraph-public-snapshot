package gitserver

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/neelance/chanrpc"
	"github.com/neelance/chanrpc/chanrpcutil"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var gitservers = env.Get("SRC_GIT_SERVERS", "", "addresses of the remote gitservers; a local gitserver process is used by default")

// DefaultClient is the default Client. Unless overwritten it is connected to servers specified by SRC_GIT_SERVERS.
var DefaultClient = NewClient(strings.Fields(gitservers))

func NewClient(addrs []string) *Client {
	client := &Client{}
	for _, addr := range addrs {
		client.connect(addr)
	}
	return client
}

// Client is a gitserver client.
type Client struct {
	servers [](chan<- *request)
}

func (c *Client) connect(addr string) {
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
	c.client.servers[serverIndex] <- &request{Exec: &execRequest{
		Repo:           c.Repo,
		EnsureRevision: c.EnsureRevision,
		Args:           c.Args[1:],
		Opt:            c.Opt,
		Stdin:          chanrpcutil.ToChunks(c.Input),
		ReplyChan:      replyChan,
	}}
	reply, ok := <-replyChan
	if !ok {
		return nil, errRPCFailed
	}

	if reply.RepoNotFound {
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

	Args           []string
	Repo           string
	EnsureRevision string
	Opt            *vcs.RemoteOpts
	Input          []byte
	ExitStatus     int
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
