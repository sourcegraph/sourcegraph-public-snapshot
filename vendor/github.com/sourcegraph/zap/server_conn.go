package zap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/ot"
)

// isWatching returns whether c is watching the given ref (either via
// explicit ref/watch or by matching one of the refspec patterns
// provided to repo/watch).
//
// The caller must hold c.mu.
func (c *serverConn) isWatching(ref RefIdentifier) bool {
	if refspecs, ok := c.watchingRepos[ref.Repo]; ok {
		if matchAnyRefspec(refspecs, ref.Ref) {
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
	closed        bool
	init          *InitializeParams
	watchingRepos map[string][]string // repo -> watch refspecs, for watched repos
	toSend        chan refUpdateItem  // holds ref updates that are waiting to be sent over this connection

	*workspaceServerConn
}

const (
	sendBufferSize = 50
	sendTimeout    = 20 * time.Second
)

var (
	DebugMu                           sync.Mutex
	DebugSimulatedLatency, _          = time.ParseDuration(os.Getenv("SIMULATED_LATENCY"))
	DebugRandomizeSimulatedLatency, _ = strconv.ParseBool(os.Getenv("RANDOMIZE_SIMULATED_LATENCY"))
)

func debugSimulateLatency() {
	DebugMu.Lock()
	if DebugSimulatedLatency == 0 {
		DebugMu.Unlock()
		return
	}
	d := DebugSimulatedLatency
	if DebugRandomizeSimulatedLatency {
		x := math.Abs(rand.NormFloat64()*0.6 + 1)
		if x < 0.1 || x > 3 {
			x = 1
		}
		d = time.Duration(float64(d) * x)
	}
	DebugMu.Unlock()
	time.Sleep(d)
}

func newServerConn(ctx context.Context, server *Server, stream jsonrpc2.ObjectStream) *serverConn {
	c := &serverConn{
		server: server,
		ready:  make(chan struct{}),
		toSend: make(chan refUpdateItem, sendBufferSize),
	}
	if server.workspaceServer != nil {
		c.workspaceServerConn = &workspaceServerConn{parent: c}
	}
	c.conn = jsonrpc2.NewConn(ctx, stream, jsonrpc2.HandlerWithError(c.handle), server.ConnOpt...)
	close(c.ready)
	go c.sendRefUpdatesLoop(ctx, server.baseLogger())
	go func() {
		<-c.conn.DisconnectNotify()
		server.deleteConn(c)
	}()
	return c
}

// refUpdateItem is a unit of work for
// (*serverConn).sendRefUpdatesLoop. Exactly 1 field is set.
type refUpdateItem struct {
	nonSymbolic *RefUpdateDownstreamParams
	symbolic    *RefUpdateSymbolicParams
}

func (c *serverConn) sendRefUpdatesLoop(ctx context.Context, log *log.Context) {
	// If an error occurs here (which is the only way we return), then
	// it means the connection is no longer viable, so close it.
	defer c.Close()

	for {
		c.mu.Lock()
		var clientID string
		if c.init != nil {
			clientID = c.init.ID
		}
		log := log.With("client", clientID)
		c.mu.Unlock()

		select {
		case item, ok := <-c.toSend:
			if !ok {
				return
			}

			// TODO(sqs): add timeout
			var params interface{}
			var method string
			if item.nonSymbolic != nil {
				params = item.nonSymbolic
				method = "ref/update"
			} else {
				params = item.symbolic
				method = "ref/updateSymbolic"
			}

			ctx, cancel := context.WithTimeout(ctx, sendTimeout)
			debugSimulateLatency()
			err := c.conn.Call(ctx, method, params, nil)
			cancel()
			if err == io.ErrUnexpectedEOF {
				// This means the client connection no longer
				// exists. Continue sending to the other watchers.
			} else if err != nil {
				level.Warn(log).Log("send-error", err, "id", c.init.ID)
				if err := c.conn.Close(); err != nil {
					level.Warn(log).Log("send-close-error", err, "id", c.init.ID)
				}
				return
			}

		case <-ctx.Done():
			return
		}
	}
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
		return repo.getConfig(), nil

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

		var res RepoConfigureResult

		oldConfig, newConfig, err := c.server.updateRepoConfiguration(ctx, repo, func(config *RepoConfiguration) error {
			var removedRemotes []string
			for remoteName := range config.Remotes {
				if _, ok := params.Remotes[remoteName]; !ok {
					removedRemotes = append(removedRemotes, remoteName)
				}
			}
			sort.Strings(removedRemotes)

			// Handle refs whose upstream remotes will be removed.
			for _, removedRemote := range removedRemotes {
				for refName, ref := range config.Refs {
					if ref.Upstream == removedRemote {
						// Unset this ref's upstream in the config.
						ref.Upstream = ""
						config.Refs[refName] = ref

						// Notify the client so it can tell the user.
						res.UpstreamConfigurationRemovedFromRefs = append(res.UpstreamConfigurationRemovedFromRefs, refName)
					}
				}
			}

			// TODO(sqs): Handle refs that are no longer matched by their
			// upstream repo's refspec.

			if config.Remotes == nil {
				config.Remotes = map[string]RepoRemoteConfiguration{}
			}
			config.Remotes = params.Remotes
			return nil
		})
		if err != nil {
			return nil, err
		}

		if err := c.server.applyRepoConfiguration(ctx, log, params.Repo, repo, oldConfig, newConfig, false, false /* TODO(sqs): these 2 bools are not meaningful */); err != nil {
			return nil, err
		}
		return res, nil

	case "repo/watch":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RepoWatchParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo, "refspecs", fmt.Sprint(params.Refspecs))
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
			repoConfig := repo.getConfig()
			if remoteConfig, ok := repoConfig.Remotes[remote]; !ok {
				return nil, fmt.Errorf("HINT: requested RefInfo for ref %q but there is no remote configured with name %q (remotes: %+v)", params.Ref, remote, repoConfig.Remotes)
			} else if !matchAnyRefspec(remoteConfig.Refspecs, branch) {
				return nil, fmt.Errorf("HINT: requested RefInfo for ref %q but no remote %q refspecs %q match the branch name", params.Ref, remote, remoteConfig.Refspecs)
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

		oldConfig, newConfig, err := c.server.updateRepoConfiguration(ctx, repo, func(config *RepoConfiguration) error {
			// Validate new config.
			if remoteName := params.RefConfiguration.Upstream; remoteName != "" {
				remote, ok := config.Remotes[remoteName]
				if !ok {
					return &jsonrpc2.Error{
						Code:    int64(ErrorCodeRemoteNotExists),
						Message: fmt.Sprintf("remote does not exist: %s", remoteName),
					}
				}
				if !matchAnyRefspec(remote.Refspecs, params.RefIdentifier.Ref) {
					return &jsonrpc2.Error{
						Code:    int64(ErrorCodeInvalidConfig),
						Message: fmt.Sprintf("ref is not matched by refspec %q configured for remote %s", remote.Refspecs, remoteName),
					}
				}
			}

			if config.Refs == nil {
				config.Refs = map[string]RefConfiguration{}
			}
			config.Refs[ref.Name] = params.RefConfiguration
			return nil
		})
		if err != nil {
			if e, ok := err.(*jsonrpc2.Error); ok {
				e.Message = fmt.Sprintf("configure ref %s: %s", params.RefIdentifier.Ref, err)
			}
			return nil, err
		}

		if err := c.server.applyRepoConfiguration(ctx, log, params.RefIdentifier.Repo, repo, oldConfig, newConfig, false, true); err != nil {
			return nil, err
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
		if err := c.server.handleRefUpdateFromDownstream(ctx, log, repo, params, c, true, true); err != nil {
			level.Error(log).Log("params", params, "err", err)
			return nil, err
		}
		return nil, nil

	case "ref/updateSymbolic":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefUpdateSymbolicParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		log = log.With("repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, log, params.RefIdentifier.Repo)
		if err != nil {
			return nil, err
		}
		if err := c.server.handleSymbolicRefUpdate(ctx, log, c, repo, params); err != nil {
			return nil, err
		}
		return nil, nil

	case "shutdown":
		level.Debug(log).Log("message", "client "+req.Method)
		return nil, nil

	case "exit":
		level.Debug(log).Log("message", "client "+req.Method)
		if err := c.Close(); err != nil {
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
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()
	c.server.deleteConn(c)
	close(c.toSend)
	return c.conn.Close()
}
