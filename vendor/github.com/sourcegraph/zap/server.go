package zap

import (
	"context"
	"io"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/ws"
)

// ServerBackend is how the Server creates server
// workspaces. It is pluggable, so a server can be configured to
// create workspaces based on a git repository, for example.
type ServerBackend interface {
	// Create creates a new workspace.
	Create(ctx context.Context, log *log.Context, repo, gitBase string) (*ws.Proxy, error)

	// CanAccess is called to determine if the client can access the
	// given repo (and all of its refs).
	CanAccess(ctx context.Context, log *log.Context, repo string) (bool, error)
}

// Server is the server that manages the source of truth for all
// workspaces. It is like a git remote endpoint.
//
// SERVER EXTENSIONS
//
// A Zap bare server is concerned with storing and retrieving
// information about Zap branches and workspaces.
//
// A server can be extended to include additional functionality.
//
// WORKSPACE SERVER
//
// The Zap workspace server sits on top of a Zap bare server,
// constantly scans for changes in repository worktrees on disk, and
// presents a simple UI based on branches and directories.
//
// Analogy to Git: The Zap bare server by itself is like a Git bare
// repo, concerned with storing and retrieving information, not
// usability. The Zap workspace server is like a Git worktree. Users
// interact with Git by using higher-level commands in a worktree (git
// commit, git checkout, git status, etc.); these commands simplify
// the UI and understand both the on-disk worktree and the Git
// repository database.
//
// A workspace server is intended to run on a user's local machine and
// provides additional capabilities and APIs for a user's local
// workspaces:
//
// - workspace/* API methods that operate on local git repository
//   directories directly. This allows users and editors to interact
//   with Zap without needing to specify the "repo" or "ref"/"branch"
//   parameters (which the workspace server automatically and
//   continuously computes based on the git repository).
// - File system watching of directories added via
//   workspace/add. Changes to a workspace's files are encoded as OT
//   ops.
//
// REMOTES & UPSTREAMS
//
// A Zap branch can be configured to track an upstream branch that lives on
// another Zap server. A Zap local tracking branch "proxies" OT
// operations between its downstream clients and the upstream
// server. To its downstream clients, a Zap local tracking branch behaves
// as an OT server. To its upstream, a Zap local tracking branch behaves
// as an OT client.
//
// This makes it easier to implement and run editors and other local
// tools because:
//
// - Editors do not need to implement "compose" and "transform" for OT
//   workspace ops. They have a local (low-latency) connection to the
//   local server, which has the Go reference implementation of
//   "compose" and "transform". Because this connection is low
//   latency, the editor can assume it will never have pending or
//   buffered ops waiting for the server.
// - If you're using multiple editors or tools locally on the same Zap
//   branch, they will all use a consistent view of the workspace
//   (i.e., the local server is the single, shared source of
//   truth). Also, only one remote connection needs to be established
//   (from the local server to the upstream), which cuts down on
//   bandwidth usage.
//
// Consider the following scenario. The local server and editor are
// intended to run on the same machine.
//
//      Alice's editor             Bob's editor
//          ^                           ^
//     	    |                           |
//     	    |                           |
//     	    v                           v
//     	Alice's local server       Bob's local server
//          ^                           ^
//     	    |                           |
//     	    |                           |
//     	    v                           |
//     	Upstream server<----------------+
//
type Server struct {
	ID string

	LogWriter io.Writer // where logs should be written to (os.Stderr by default)

	backend ServerBackend

	reposMu sync.Mutex
	repos   map[string]*serverRepo

	connsMu sync.Mutex
	conns   map[*serverConn]struct{} // open connections to clients

	recvMu sync.Mutex

	readyToAccept chan struct{}

	ConnOpt []jsonrpc2.ConnOpt

	remotes         serverRemotes
	workspaceServer *workspaceServer

	closedMu sync.Mutex
	closed   bool

	bgCtx context.Context

	work chan func() error

	updateFromDownstreamMu    sync.Mutex
	updateRemoteTrackingRefMu sync.Mutex
}

// NewServer creates a new remote server.
func NewServer(backend ServerBackend) *Server {
	s := &Server{
		backend:       backend,
		repos:         map[string]*serverRepo{},
		readyToAccept: make(chan struct{}),
		work:          make(chan func() error, 25),
	}
	s.remotes.parent = s
	return s
}

// Start starts the remote server.
func (s *Server) Start(ctx context.Context) {
	if s.bgCtx != nil {
		panic("server is already started")
	}
	s.bgCtx = ctx
	close(s.readyToAccept)
	go s.startWorker(s.bgCtx)

	go func() {
		<-ctx.Done()
		s.closedMu.Lock()
		s.closed = true
		s.closedMu.Unlock()
	}()
}

// Accept accepts a new connection to the remote server from a
// client. It returns a channel that is closed when the client
// disconnects.
//
// (Server).Start must be called before any (Server).Accept call will
// return.
func (s *Server) Accept(ctx context.Context, stream jsonrpc2.ObjectStream) <-chan struct{} {
	<-s.readyToAccept
	sc := newServerConn(ctx, s, stream)
	s.connsMu.Lock()
	if s.conns == nil {
		s.conns = make(map[*serverConn]struct{}, 1)
	}
	s.conns[sc] = struct{}{}
	s.connsMu.Unlock()
	return sc.conn.DisconnectNotify()
}

func (s *Server) deleteConn(c *serverConn) {
	s.connsMu.Lock()
	delete(s.conns, c)
	s.connsMu.Unlock()
}
