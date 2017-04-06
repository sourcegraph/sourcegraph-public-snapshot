package server

import (
	"context"
	"io"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/server/config"
	"github.com/sourcegraph/zap/server/repodb"
)

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
	// ID is used to identify this server in log messages. It should not
	// be assumed to be unique.
	ID string

	// LogWriter is were logs should be written to (os.Stderr by default)
	LogWriter io.Writer

	// IsPrivate is true if the server is not shared amongst users. If it
	// is false operations such as ref/list should not be allowed due to
	// potentially leaking data. In the future we should deprecate this
	// setting and instead use efficient access checks on listing
	// operations.
	IsPrivate bool

	// ConfigFile stores the configuration for the Zap server.
	ConfigFile *config.File

	Repos *repodb.RepoDB // the repository database

	connsMu sync.Mutex
	conns   map[*Conn]struct{} // open connections to clients

	// readyToAccept is closed when the server has been started
	readyToAccept chan struct{}

	// ConnOpt are the connection options used on all connections that are
	// accepted.
	ConnOpt []jsonrpc2.ConnOpt

	remotes serverRemotes
	// workspaceServer *workspaceServer TODO(sqs8)

	// Background is the context used to start this server. It should
	// be used as the context for any background operations done by
	// the server (i.e., operations that aren't performed in the path
	// of handling a client request).
	Background context.Context

	ext []Extension

	// TestingClientConnect is used in tests to delay client
	// connections until the server receives on the returned channel.
	TestingClientConnect func() <-chan struct{}
}

// New creates a new unstarted Zap server.
//
// Call (*Server).Start to start the server, and call (*Server).Accept
// to handle an incoming connection.
func New(backend repodb.Backend) *Server {
	s := &Server{
		Repos:         repodb.New(backend),
		readyToAccept: make(chan struct{}),
	}
	s.remotes.parent = s
	return s
}

// Start starts the remote server. If startup fails (with Start
// returning a non-nil error), the server is no longer usable and a
// new server must be created with New.
func (s *Server) Start(ctx context.Context) error {
	if s.Background != nil {
		panic("server was already started")
	}
	s.Background = ctx

	// Start extensions.
	for _, ext := range s.ext {
		if err := ext.Start(ctx); err != nil {
			return err
		}
	}

	close(s.readyToAccept)
	return nil
}

// Accept accepts a new connection to the remote server from a
// client. It returns a channel that is closed when the client
// disconnects.
//
// (Server).Start must be called before any (Server).Accept call will
// return.
func (s *Server) Accept(ctx context.Context, stream jsonrpc2.ObjectStream) <-chan struct{} {
	<-s.readyToAccept

	if s.TestingClientConnect != nil {
		wait := s.TestingClientConnect()
		<-wait
	}

	c := newConn(ctx, s, stream)
	s.connsMu.Lock()
	if s.conns == nil {
		s.conns = make(map[*Conn]struct{}, 1)
	}
	s.conns[c] = struct{}{}
	s.connsMu.Unlock()
	return c.Conn.DisconnectNotify()
}

// isClosed returns true if the server is closed.
func (s *Server) isClosed() bool {
	return s.Background.Err() != nil
}

func (s *Server) deleteConn(c *Conn) {
	s.connsMu.Lock()
	delete(s.conns, c)
	s.connsMu.Unlock()
}

// TestingCloseClientConns is used in tests to close all connections
// from clients without stopping the server.
func (s *Server) TestingCloseClientConns() {
	s.connsMu.Lock()
	conns := make([]*Conn, 0, len(s.conns))
	for c := range s.conns {
		conns = append(conns, c)
	}
	s.connsMu.Unlock()
	for _, c := range conns {
		_ = c.Close()
	}
}

// TestingCloseClientConn is used in tests to close specific client
// connections given the ID that they provided to the server in the
// "initialize" request. Usually there is only 1 client connection per
// ID, but it's possible for there to be multiple. If there are no
// client connections with the given ID, it panics.
func (s *Server) TestingCloseClientConn(id string) {
	var conns []*Conn
	s.connsMu.Lock()
	for c := range s.conns {
		c.mu.Lock()
		match := c.init != nil && c.init.ID == id
		c.mu.Unlock()
		if match {
			conns = append(conns, c)
		}
	}
	s.connsMu.Unlock()
	if len(conns) == 0 {
		panic("no client connections with ID: " + id)
	}
	for _, c := range conns {
		_ = c.Close()
	}
}
