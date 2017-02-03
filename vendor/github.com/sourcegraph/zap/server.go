package zap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/ot"
	"github.com/sourcegraph/zap/server/refdb"
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
}

// NewServer creates a new remote server.
func NewServer(backend ServerBackend) *Server {
	s := &Server{
		backend:       backend,
		repos:         map[string]*serverRepo{},
		readyToAccept: make(chan struct{}),
		work:          make(chan func() error),
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

// isWatching returns whether c is watching the given ref (either via
// explicit ref/watch or by matching one of the refspec patterns
// provided to repo/watch).
//
// The caller must hold c.mu.
func (c *serverConn) isWatching(ref RefIdentifier) bool {
	if refspec, ok := c.watchingRepos[ref.Repo]; ok {
		if refdb.MatchPattern(refspec, ref.Ref) {
			return true
		}
	}
	return false
}

type serverConn struct {
	server *Server // the server that created this conn
	conn   *jsonrpc2.Conn

	ready chan struct{} // ready to handle requests

	mu            sync.Mutex
	init          *InitializeParams
	watchingRepos map[string]string // repo -> refspec of watched repos

	*workspaceServerConn
}

func newServerConn(ctx context.Context, server *Server, stream jsonrpc2.ObjectStream) *serverConn {
	c := &serverConn{
		server: server,
		ready:  make(chan struct{}),
	}
	if server.workspaceServer != nil {
		c.workspaceServerConn = &workspaceServerConn{parent: c}
	}
	c.conn = jsonrpc2.NewConn(ctx, stream, jsonrpc2.HandlerWithError(c.handle), server.ConnOpt...)
	close(c.ready)
	go func() {
		<-c.conn.DisconnectNotify()
		server.deleteConn(c)
	}()
	return c
}

func (c *serverConn) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	// Prevent any uncaught panics from taking the entire server down.
	if os.Getenv("NO_PANIC_HANDLER") == "" {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("unexpected panic: %v", r)
				const size = 64 << 10 // copied from net/http
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				stdlog.Printf("panic serving %v: %v\n%s", req.Method, r, buf)
				return
			}
		}()
	}

	<-c.ready

	c.mu.Lock()
	inited := c.init != nil
	log := c.server.baseLogger()
	if c.init != nil {
		log = log.With("client", c.init.ID)
	}
	log = log.With("method", req.Method)
	c.mu.Unlock()
	if !inited && req.Method != "initialize" {
		return nil, &jsonrpc2.Error{Code: int64(ErrorCodeNotInitialized), Message: "connection is not initialized (client must send initialize request)"}
	}

	switch req.Method {
	case "initialize":
		c.mu.Lock()
		inited := c.init != nil
		c.mu.Unlock()
		if inited {
			return nil, &jsonrpc2.Error{Code: int64(ErrorCodeAlreadyInitialized), Message: "connection is already initialized"}
		}

		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.init = &params
		c.mu.Unlock()
		level.Debug(log.With("client", c.init.ID)).Log("msg", "new client connected")
		return InitializeResult{
			Capabilities: ServerCapabilities{WorkspaceOperationalTransformation: true},
		}, nil

	case "initialized":
		return true, nil

	case "repo/info":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RepoInfoParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo)
		repo, err := c.server.getRepo(ctx, log, params.Repo)
		if err != nil {
			return nil, err
		}
		return repo.config, nil

	case "repo/configure":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RepoConfigureParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo)
		repo, err := c.server.getRepo(ctx, log, params.Repo)
		if err != nil {
			return nil, err
		}
		repo.mu.Lock()
		defer repo.mu.Unlock()
		if err := c.server.doUpdateBulkRepoRemoteConfiguration(ctx, log, repo, repo.config.Remotes, params.Remotes); err != nil {
			return nil, err
		}
		if repo.workspace != nil {
			if err := repo.workspace.Configure(ctx, repo.config); err != nil {
				return nil, err
			}
		}
		return nil, nil

	case "repo/watch":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RepoWatchParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo, "refspec", params.Refspec)
		repo, err := c.server.getRepo(ctx, log, params.Repo)
		if err != nil {
			return nil, err
		}
		if err := c.handleRepoWatch(ctx, log, repo, params); err != nil {
			return nil, err
		}
		return nil, nil

	case "ref/list":
		if req.Params != nil && string(*req.Params) != "null" {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		c.server.reposMu.Lock()
		defer c.server.reposMu.Unlock()
		refs := []RefInfo{} // marshal to [] JSON even if empty
		for repoName, repo := range c.server.repos {
			for _, ref := range repo.refdb.List("*") {
				info := RefInfo{RefIdentifier: RefIdentifier{Repo: repoName, Ref: ref.Name}}
				if ref.IsSymbolic() {
					info.Target = ref.Target
				} else {
					refObj := ref.Object.(serverRef)
					info.GitBase = refObj.gitBase
					info.GitBranch = refObj.gitBranch
					info.Rev = refObj.rev()
				}

				watchers := c.server.watchers(RefIdentifier{Repo: repoName, Ref: ref.Name})
				info.Watchers = make([]string, len(watchers))
				for i, wc := range watchers {
					info.Watchers[i] = wc.init.ID
				}
				sort.Strings(info.Watchers)

				refs = append(refs, info)
			}
		}
		sort.Sort(sortableRefInfos(refs))
		return refs, nil

	case "ref/info":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefIdentifier
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, log, params.Repo)
		if err != nil {
			return nil, err
		}

		// Friendly check: if the ref is a remote tracking ref and the
		// repo isn't configured to track the remote, then warn the
		// user.
		//
		// TODO(sqs): make this more robust
		if strings.HasPrefix(params.Ref, "refs/remotes/") {
			parts := strings.SplitN(strings.TrimPrefix(params.Ref, "refs/remotes/"), "/", 2)
			remote := parts[0]
			branch := parts[1]
			if repoConfig, ok := repo.config.Remotes[remote]; !ok {
				return nil, fmt.Errorf("HINT: requested RefInfo for ref %q but there is no remote configured with name %q", params.Ref, remote)
			} else if !refdb.MatchPattern(repoConfig.Refspec, branch) {
				return nil, fmt.Errorf("HINT: requested RefInfo for ref %q but the remote %q refspec %q does not match the branch name", params.Ref, remote, repoConfig.Refspec)
			}
		}

		ref := repo.refdb.Lookup(params.Ref)
		if ref == nil {
			return nil, &jsonrpc2.Error{
				Code:    int64(ErrorCodeRefNotExists),
				Message: fmt.Sprintf("ref not found: %s", params.Ref),
			}
		}
		res := RefInfoResult{Target: ref.Target}
		if ref.IsSymbolic() {
			ref, err = repo.refdb.Resolve(ref.Name)
			if err != nil {
				return nil, &jsonrpc2.Error{
					Code:    int64(ErrorCodeSymbolicRefInvalid),
					Message: fmt.Sprintf("symbolic ref resolution error: %s", err),
				}
			}
		}
		if o, ok := ref.Object.(serverRef); ok {
			res.State = &RefState{
				RefBaseInfo: RefBaseInfo{GitBase: o.gitBase, GitBranch: o.gitBranch},
				History:     o.history(),
			}
			if res.State.History == nil {
				res.State.History = []ot.WorkspaceOp{}
			}

			// Extra diagnostics
			res.Wait = o.ot.Wait
			res.Buf = o.ot.Buf
			res.UpstreamRevNumber = o.ot.UpstreamRevNumber
		}
		return res, nil

	case "ref/configure":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefConfigureParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, log, params.RefIdentifier.Repo)
		if err != nil {
			return nil, err
		}
		ref := repo.refdb.Lookup(params.RefIdentifier.Ref)
		if ref == nil {
			return nil, &jsonrpc2.Error{
				Code:    int64(ErrorCodeRefNotExists),
				Message: fmt.Sprintf("configure called on nonexistent ref: %s", params.RefIdentifier.Ref),
			}
		}
		repo.mu.Lock()
		defer repo.mu.Unlock()
		if err := c.server.doUpdateRefConfiguration(ctx, log, repo, params.RefIdentifier, ref, repo.config.Refs[ref.Name], params.RefConfiguration); err != nil {
			return nil, err
		}
		if repo.workspace != nil {
			if err := repo.workspace.Configure(ctx, repo.config); err != nil {
				return nil, err
			}
		}
		return nil, nil

	case "ref/update":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefUpdateUpstreamParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, log, params.RefIdentifier.Repo)
		if err != nil {
			return nil, err
		}
		if err := c.server.handleRefUpdateFromDownstream(ctx, log, repo, params, c, true); err != nil {
			level.Error(log).Log("params", params, "err", err)
			return nil, err
		}
		return nil, nil

	default:
		res, err := c.handleWorkspaceServerMethod(ctx, log, conn, req)
		if err != errNotHandled {
			return res, err
		}

		if extHandler, present := serverExtensions[req.Method]; present {
			return extHandler(ctx, conn, req)
		}

		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not found: %q", req.Method)}
	}
}

// Close closes the connection. The connection may not be used after
// it has been closed.
func (c *serverConn) Close() error {
	c.server.deleteConn(c)
	return c.conn.Close()
}
