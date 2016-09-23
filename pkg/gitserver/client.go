package gitserver

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"hash/fnv"
	"log"
	"time"

	"github.com/neelance/chanrpc"
	"github.com/neelance/chanrpc/chanrpcutil"
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

type genericReply interface {
	repoFound() bool
}

func (c *Client) broadcastCall(ctx context.Context, newRequest func() (*request, func() (genericReply, bool))) (interface{}, error) {
	// Check that ctx is not expired before broadcasting over the network.
	select {
	case <-ctx.Done():
		deadlineExceededCounter.Inc()
		return nil, ctx.Err()
	default:
	}

	allReplies := make(chan genericReply, len(c.servers))
	for _, server := range c.servers {
		req, getReply := newRequest()
		server <- req
		go func() {
			reply, ok := getReply()
			if !ok {
				allReplies <- nil
				return
			}
			allReplies <- reply
		}()
	}

	rpcError := false
	for range c.servers {
		reply := <-allReplies
		if reply == nil {
			rpcError = true
			continue
		}
		if reply.repoFound() {
			return reply, nil
		}
	}
	if rpcError {
		return nil, errRPCFailed
	}
	return nil, vcs.RepoNotExistError{}
}

var errRPCFailed = errors.New("gitserver: rpc failed")

var deadlineExceededCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "client_deadline_exceeded",
	Help:      "Times that Client.broadcastCall() returned context.DeadlineExceeded",
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
	genReply, err := c.client.broadcastCall(ctx, func() (*request, func() (genericReply, bool)) {
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

// Search performs a remote search.
func (c *Client) Search(ctx context.Context, repo string, commit vcs.CommitID, opt vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	genReply, err := c.broadcastCall(ctx, func() (*request, func() (genericReply, bool)) {
		replyChan := make(chan *searchReply, 1)
		return &request{Search: &searchRequest{Repo: repo, Commit: commit, Opt: opt, ReplyChan: replyChan}},
			func() (genericReply, bool) { reply, ok := <-replyChan; return reply, ok }
	})
	if err != nil {
		return nil, err
	}

	reply := genReply.(*searchReply)
	if reply.CloneInProgress {
		return nil, vcs.RepoNotExistError{CloneInProgress: true}
	}
	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}
	return reply.Results, nil
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
	// We check if repo already exists by executing `git remote`. It may seem redundant since the
	// create request also checks that, but the purpose is to first do a broadcast and check if _any_
	// server already has the repo available.
	cmd := c.Command("git", "remote")
	cmd.Repo = repo
	err := cmd.Run(ctx)
	if err == nil {
		return vcs.ErrRepoExist
	}
	if !vcs.IsRepoNotExist(err) {
		// The only acceptable error is repo doesn't exist, if it's something else, there's a problem. Return the error.
		return err
	}
	if repoNotExistError := err.(vcs.RepoNotExistError); repoNotExistError.CloneInProgress {
		// If some server is already cloning this repository, report it and don't try to create another.
		return repoNotExistError
	}

	// This hash is used to avoid concurrent init on two servers, it does not need to be stable over long timespans.
	h := fnv.New32a()
	if _, err := h.Write([]byte(repo)); err != nil {
		return err
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

// Remove removes repo remotely.
func (c *Client) Remove(ctx context.Context, repo string) error {
	genReply, err := c.broadcastCall(ctx, func() (*request, func() (genericReply, bool)) {
		replyChan := make(chan *removeReply, 1)
		return &request{Remove: &removeRequest{Repo: repo, ReplyChan: replyChan}},
			func() (genericReply, bool) { reply, ok := <-replyChan; return reply, ok }
	})
	if err != nil {
		return err
	}

	reply := genReply.(*removeReply)
	if reply.CloneInProgress {
		return vcs.RepoNotExistError{CloneInProgress: true}
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}
