package server

import (
	"context"
	"encoding/json"
	"fmt"
	stdlog "log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/slimsag/gup"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/internal/debugutil"
	ot2 "github.com/sourcegraph/zap/op"
	"github.com/sourcegraph/zap/pkg/fpath"
	"github.com/sourcegraph/zap/server/refdb"
	"github.com/sourcegraph/zap/server/refstate"
	"github.com/sourcegraph/zap/server/repodb"
)

// A Conn is a connection to this Zap server.
type Conn struct {
	Server *Server        // the server that owns this connection
	Conn   *jsonrpc2.Conn // the underlying JSON-RPC 2.0 connection

	ready chan struct{} // ready to handle requests

	mu            sync.Mutex
	closed        bool
	init          *zap.InitializeParams
	watchingRepos map[fpath.KeyString][]string // repo -> watch refspecs, for watched repos

	// TODO(sqs8)
	//
	// *Conn
}

const (
	sendTimeout = 20 * time.Second
)

var (
	// Vars in this block are guarded by debugutil.Mu.

	// TestSimulateResetAfterErrorInSendToUpstream is used in tests
	// only and causes SendToUpstream to simulate an error condition
	// that triggers this server to reset the ref on its upstream.
	TestSimulateResetAfterErrorInSendToUpstream bool
)

func newConn(ctx context.Context, server *Server, stream jsonrpc2.ObjectStream) *Conn {
	c := &Conn{
		Server: server,
		ready:  make(chan struct{}),
	}
	// TODO(sqs8)
	//
	// if server.workspaceServer != nil {
	// 	c.Conn = &Conn{parent: c}
	// }
	c.Conn = jsonrpc2.NewConn(ctx, stream, jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(c.handle)), server.ConnOpt...)
	close(c.ready)
	go func() {
		select {
		case <-gup.UpdateAvailable:
			// Ensure the channel remains full for other client connections..
			select {
			case gup.UpdateAvailable <- true:
			default:
			}

			// After the VS Code extension client connects, handlers are
			// registered. If we were to immediately respond right now (which is
			// very possible), then we would be in a race between the client
			// registering its handlers and us sending the notification.
			time.Sleep(1 * time.Second)

			if err := c.Conn.Notify(ctx, "window/showMessage", &zap.ShowMessageParams{
				Message: "Zap binary update available, use 'zap upgrade' in terminal to update",
				Type:    zap.MessageTypeInfo,
			}); err != nil {
				log.WithPrefix(server.BaseLogger(), "updater").Log("err", err)
			}

			// Now just wait for disconnection.
			<-c.Conn.DisconnectNotify()
			server.deleteConn(c)

		case <-c.Conn.DisconnectNotify():
			server.deleteConn(c)
			return
		}
	}()
	return c
}

// send sends a ref update to the peer client. If the update fails to
// send, then it must be a connection error, and we will
// disconnect/remove the client.
func (c *Conn) send(ctx context.Context, logger log.Logger, params zap.RefUpdateDownstreamParams) {
	c.mu.Lock()
	var clientID string
	if c.init != nil {
		clientID = c.init.ID
	}
	logger = log.With(logger, "client", clientID)
	c.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	debugutil.SimulateLatency()

	// No wait for the client to reply (we use Notify not Call). We
	// are the server and we are the source of truth. If the client
	// experiences an error applying our op, then it is the client's
	// fault and there is nothing we can do to help resolve it; the
	// recovery must be done on the client.
	if err := c.Conn.Notify(ctx, "ref/update", params); err != nil {
		level.Warn(logger).Log("send-error", err, "id", c.init.ID)
		if err := c.Conn.Close(); err != nil {
			level.Warn(logger).Log("send-close-error", err, "id", c.init.ID)
		}
		return
	}
}

func (c *Conn) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
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
	logger := c.Server.BaseLogger()
	if c.init != nil {
		logger = log.With(logger, "client", c.init.ID)
	}
	logger = log.With(logger, "method", req.Method)
	c.mu.Unlock()
	if !inited && req.Method != "initialize" {
		return nil, &jsonrpc2.Error{Code: int64(zap.ErrorCodeNotInitialized), Message: "connection is not initialized (client must send initialize request)"}
	}

	switch req.Method {
	case "initialize":
		// Hold lock since we check and modify c.init
		c.mu.Lock()
		defer c.mu.Unlock()
		inited := c.init != nil
		if inited {
			return nil, &jsonrpc2.Error{Code: int64(zap.ErrorCodeAlreadyInitialized), Message: "connection is already initialized"}
		}

		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		c.init = &params
		level.Debug(logger).Log("client-connected", c.init.ID)
		return zap.InitializeResult{
			Capabilities: zap.ServerCapabilities{WorkspaceOperationalTransformation: true},
		}, nil

	case "initialized":
		return true, nil

	case "repo/info":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.RepoInfoParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		logger = log.With(logger, "repo", params.Repo)
		repo, err := c.Server.Repos.Add(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}
		defer repo.Unlock()

		return repo.Repo.Config, nil

	case "repo/configure":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.RepoConfigureParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		logger = log.With(logger, "repo", params.Repo)
		repo, err := c.Server.Repos.Add(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}
		defer repo.Unlock()

		oldConfig, newConfig, err := c.Server.updateRepoConfiguration(ctx, *repo, func(config *zap.RepoConfiguration) error {
			if len(params.Remotes) > 1 {
				return fmt.Errorf("a repository may have at most 1 remote (got %+v)", params.Remotes)
			}
			config.Remotes = params.Remotes
			return nil
		})
		if err != nil {
			return nil, err
		}

		if err := c.Server.ApplyRepoConfiguration(ctx, logger, *repo, oldConfig, newConfig); err != nil {
			return nil, err
		}
		return nil, nil

	case "repo/watch":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.RepoWatchParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		logger = log.With(logger, "repo", params.Repo, "refspecs", fmt.Sprint(params.Refspecs))
		repo, err := c.Server.Repos.Add(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}
		defer repo.Unlock()

		if err := c.handleRepoWatch(ctx, logger, *repo, params); err != nil {
			return nil, err
		}
		return nil, nil

	case "ref/list":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.RefListParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		logger = log.With(logger, "repo", params.Repo)

		if params.Repo == "" {
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidParams,
				Message: "ref/list requires repo to be specified",
			}
		}

		repo, err := c.Server.Repos.Add(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}
		defer repo.Unlock()

		refs := c.refsInRepo(params.Repo, *repo)
		if refs == nil {
			refs = []zap.RefInfo{} // marshal to [] JSON even if empty
		}

		sort.Sort(sortableRefInfos(refs))
		return refs, nil

	case "repo/list":
		if req.Params != nil && string(*req.Params) != "null" {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		if !c.Server.IsPrivate {
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: "repo/list not allowed on shared server",
			}
		}
		return c.Server.Repos.List(), nil

	case "ref/info":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.RefInfoParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if !params.Fuzzy {
			zap.CheckRefName(params.Ref)
		}

		logger = log.With(logger, "repo", params.Repo, "ref", params.Ref)
		repo, err := c.Server.Repos.Add(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}
		defer repo.Unlock()

		var ref refdb.OwnedRef
		if params.Fuzzy {
			ref = lookupRefByFuzzyName(repo.RefDB, params.Ref)
		} else {
			ref = repo.RefDB.Lookup(params.Ref)
		}
		defer ref.Unlock()

		if ref.Ref == nil {
			return nil, &jsonrpc2.Error{
				Code: int64(zap.ErrorCodeRefNotExists),
				// TODO(slimsag): do not use CLI instructions here, instead
				// type assert to *jsonrpc2.Error in the CLI and then print
				// instructions.
				Message: fmt.Sprintf("ref not found: %s\n\nHINT: Maybe you need to run 'zap auth'?", params.Ref),
			}
		}

		res := zap.RefInfo{
			RefIdentifier: zap.RefIdentifier{Repo: params.Repo, Ref: ref.Ref.Name}, // use resolved (not fuzzy) name
			RefState:      ref.Ref.Data.(refstate.RefState).RefState,
		}
		if res.Data != nil && res.Data.History == nil {
			res.Data.History = []ot2.Ops{}
		}

		// Add watchers.
		if watchers := c.Server.watchers(params.RefIdentifier); len(watchers) > 0 {
			res.Watchers = make([]string, len(watchers))
			for i, wc := range watchers {
				res.Watchers[i] = wc.init.ID
			}
			sort.Strings(res.Watchers)
		}

		// Deep copy to avoid race conditions (the ref.Unlock() must
		// execute *after* we have performed a deep copy, or else our
		// caller would get a racy return value).
		res.RefState = res.RefState.DeepCopy()

		return res, nil

	case "ref/configure":
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: "ref/configure method was removed",
		}

	case "ref/update":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.RefUpdateUpstreamParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		zap.CheckRefName(params.RefIdentifier.Ref)
		if err := params.Validate(); err != nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()}
		}

		logger = log.With(logger, "repo", params.Repo, "ref", params.Ref)
		repo, err := c.Server.Repos.Add(ctx, logger, params.RefIdentifier.Repo)
		if err != nil {
			return nil, err
		}
		defer repo.Unlock()

		ref := repo.RefDB.Lookup(params.RefIdentifier.Ref)
		defer ref.Unlock()

		// Receive the update and apply to our internal ref state.
		level.Debug(logger).Log("recv-from-downstream", params)
		var u refstate.RefUpdate
		u.FromUpdateUpstream(params)
		if err := c.Server.RefUpdate(ctx, logger, c, repo, &ref, u); err != nil {
			return nil, err
		}
		return nil, nil

	case "ping":
		return "pong", nil

	case "debug/log":
		if v, _ := strconv.ParseBool(os.Getenv("ZAP_ENABLE_DEBUG_LOG_REQUEST")); !v {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "debug/log is not enabled (set env var ZAP_ENABLE_DEBUG_LOG_REQUEST=1)"}
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params zap.DebugLogParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		time.Sleep(10 * time.Millisecond) // flush
		if params.Header {
			fmt.Fprintf(os.Stderr, strings.Repeat("━", 70)+"\n███ %s\n", params.Text)
		} else {
			level.Info(logger).Log("text", params.Text)
		}
		time.Sleep(10 * time.Millisecond) // flush
		return nil, nil

	case "shutdown":
		level.Debug(logger).Log("message", "client "+req.Method)
		return nil, nil

	case "exit":
		level.Debug(logger).Log("message", "client "+req.Method)
		if err := c.Close(); err != nil {
			return nil, err
		}
		return nil, nil

	default:
		// Try handling the method by using a server extension.
		for _, ext := range c.Server.ext {
			if ext.Handle != nil {
				res, err := ext.Handle(ctx, logger, c, req)
				if e, ok := err.(*jsonrpc2.Error); ok && e.Code == jsonrpc2.CodeMethodNotFound {
					continue
				}
				return res, err
			}
		}
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not found: %q", req.Method)}
	}
}

// refsInRepo returns all references in the given repo.
func (c *Conn) refsInRepo(repoName string, repo repodb.OwnedRepo) []zap.RefInfo {
	var refs []zap.RefInfo
	for _, ref := range repo.RefDB.List("*") {
		info := zap.RefInfo{RefIdentifier: zap.RefIdentifier{Repo: repoName, Ref: ref.Name}}
		if ref.IsSymbolic() {
			info.Target = ref.Target()
		} else {
			// Acquire exclusive ref lock so we can safely access its
			// Object field.
			ownedRef := repo.RefDB.Lookup(ref.Name)
			info.RefState = ref.Data.(refstate.RefState).RefState
			ownedRef.Unlock()
		}

		watchers := c.Server.watchers(zap.RefIdentifier{Repo: repoName, Ref: ref.Name})
		info.Watchers = make([]string, len(watchers))
		for i, wc := range watchers {
			info.Watchers[i] = wc.init.ID
		}
		sort.Strings(info.Watchers)

		zap.CheckRefName(info.RefIdentifier.Ref)
		if info.Target != "" {
			zap.CheckRefName(info.Target)
		}
		refs = append(refs, info)
	}
	return refs
}

// Close closes the connection. The connection may not be used after
// it has been closed.
func (c *Conn) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()
	c.Server.deleteConn(c)
	return c.Conn.Close()
}
