package zap

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/go-kit/kit/log/level"
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

	*workspaceServerConn
}

const (
	sendBufferSize = 50
	sendTimeout    = 20 * time.Second
)

var (
	DebugMu sync.Mutex // guards all vars in this block

	DebugSimulatedLatency, _          = time.ParseDuration(os.Getenv("SIMULATED_LATENCY"))
	DebugRandomizeSimulatedLatency, _ = strconv.ParseBool(os.Getenv("RANDOMIZE_SIMULATED_LATENCY"))

	// TestSimulateResetAfterErrorInSendToUpstream is used in tests only
	// and causes SendToUpstream to simulate an error condition that
	// triggers this server to reset the ref on its upstream.
	TestSimulateResetAfterErrorInSendToUpstream bool
)

func debugSimulateLatency() {
	DebugMu.Lock()
	if DebugSimulatedLatency == 0 && !DebugRandomizeSimulatedLatency {
		DebugMu.Unlock()
		return
	}
	d := DebugSimulatedLatency
	if DebugRandomizeSimulatedLatency {
		if d == 0 {
			d = 10 * time.Millisecond // default
		}
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
	}
	if server.workspaceServer != nil {
		c.workspaceServerConn = &workspaceServerConn{parent: c}
	}
	c.conn = jsonrpc2.NewConn(ctx, stream, jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(c.handle)), server.ConnOpt...)
	close(c.ready)
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

// send sends a ref update to the peer client. If the update fails to
// send, then it must be a connection error, and we will
// disconnect/remove the client.
func (c *serverConn) send(ctx context.Context, logger log.Logger, item refUpdateItem) {
	c.mu.Lock()
	var clientID string
	if c.init != nil {
		clientID = c.init.ID
	}
	logger = log.With(logger, "client", clientID)
	c.mu.Unlock()

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
	defer cancel()

	debugSimulateLatency()

	// No wait for the client to reply (we use Notify not Call). We
	// are the server and we are the source of truth. If the client
	// experiences an error applying our op, then it is the client's
	// fault and there is nothing we can do to help resolve it; the
	// recovery must be done on the client.
	if err := c.conn.Notify(ctx, method, params); err != nil {
		level.Warn(logger).Log("send-error", err, "id", c.init.ID)
		if err := c.conn.Close(); err != nil {
			level.Warn(logger).Log("send-close-error", err, "id", c.init.ID)
		}
		return
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
	logger := c.server.baseLogger()
	if c.init != nil {
		logger = log.With(logger, "client", c.init.ID)
	}
	logger = log.With(logger, "method", req.Method)
	c.mu.Unlock()
	if !inited && req.Method != "initialize" {
		return nil, &jsonrpc2.Error{Code: int64(ErrorCodeNotInitialized), Message: "connection is not initialized (client must send initialize request)"}
	}

	switch req.Method {
	case "initialize":
		// Hold lock since we check and modify c.init
		c.mu.Lock()
		defer c.mu.Unlock()
		inited := c.init != nil
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
		c.init = &params
		level.Debug(log.With(logger, "client", c.init.ID)).Log("msg", "new client connected")
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
		logger = log.With(logger, "repo", params.Repo)
		repo, err := c.server.getRepo(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}
		return repo.getConfig()

	case "repo/configure":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RepoConfigureParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		logger = log.With(logger, "repo", params.Repo)
		repo, err := c.server.getRepo(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}

		var res RepoConfigureResult

		oldConfig, newConfig, err := c.server.updateRepoConfiguration(ctx, repo, func(config *RepoConfiguration) error {
			finalConfig, err := repo.mergedConfigNoLock(*config)
			if err != nil {
				return err
			}

			var removedRemotes []string
			for remoteName := range finalConfig.Remotes {
				if _, ok := params.Remotes[remoteName]; !ok {
					removedRemotes = append(removedRemotes, remoteName)
				}
			}
			sort.Strings(removedRemotes)

			// Handle refs whose upstream remotes will be removed.
			for _, removedRemote := range removedRemotes {
				for refName, ref := range finalConfig.Refs {
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

		newConfig, err = repo.getConfig()
		if err != nil {
			return nil, err
		}
		if err := c.server.applyRepoConfiguration(ctx, logger, params.Repo, repo, oldConfig, newConfig); err != nil {
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
		logger = log.With(logger, "repo", params.Repo, "refspecs", fmt.Sprint(params.Refspecs))
		repo, err := c.server.getRepo(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}
		if err := c.handleRepoWatch(ctx, logger, repo, params); err != nil {
			return nil, err
		}
		return nil, nil

	case "ref/list":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefListParams
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

		repo, err := c.server.getRepo(ctx, logger, params.Repo)
		if err != nil {
			return nil, err
		}

		refs := c.refsInRepo(params.Repo, repo)
		if refs == nil {
			refs = []RefInfo{} // marshal to [] JSON even if empty
		}

		sort.Sort(sortableRefInfos(refs))
		return refs, nil

	case "repo/list":
		if req.Params != nil && string(*req.Params) != "null" {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		if !c.server.IsPrivate {
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: "repo/list not allowed on shared server",
			}
		}
		c.server.reposMu.Lock()
		defer c.server.reposMu.Unlock()
		reposMap := map[string]struct{}{}
		for repoName := range c.server.repos {
			reposMap[repoName] = struct{}{}
		}
		repos := make([]string, 0, len(reposMap))
		for repo := range reposMap {
			repos = append(repos, repo)
		}
		sort.Strings(repos)
		return repos, nil

	case "ref/info":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefIdentifier
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		logger = log.With(logger, "repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, logger, params.Repo)
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
			repoConfig, err := repo.getConfig()
			if err != nil {
				return nil, err
			}
			if remoteConfig, ok := repoConfig.Remotes[remote]; !ok {
				return nil, fmt.Errorf("HINT: requested RefInfo for ref %q but there is no remote configured with name %q (remotes: %+v)", params.Ref, remote, repoConfig.Remotes)
			} else if !matchAnyRefspec(remoteConfig.Refspecs, branch) {
				return nil, fmt.Errorf("HINT: requested RefInfo for ref %q but no remote %q refspecs %q match the branch name", params.Ref, remote, remoteConfig.Refspecs)
			}
		}

		ref := repo.refdb.Lookup(params.Ref)
		if ref == nil {
			return nil, &jsonrpc2.Error{
				Code: int64(ErrorCodeRefNotExists),
				// TODO(slimsag): do not use CLI instructions here, instead
				// type assert to *jsonrpc2.Error in the CLI and then print
				// instructions.
				Message: fmt.Sprintf("ref not found: %s\n\nHINT: Maybe you need to run 'zap auth'?", params.Ref),
			}
		}
		res := RefInfo{Target: ref.Target}
		if ref.IsSymbolic() {
			ref, err = repo.refdb.Resolve(ref.Name)
			if err != nil {
				return nil, &jsonrpc2.Error{
					Code:    int64(ErrorCodeSymbolicRefInvalid),
					Message: fmt.Sprintf("symbolic ref resolution error: %s", err),
				}
			}
		}

		// TODO(sqs): it is slightly racy (in a logical sense, not in
		// a memory corruption sense) to not hold BOTH the symbolic
		// ref and its target here (we only lock the target ref). we
		// might return a state of the target ref that was never in
		// existence when the symbolic ref pointed to it.
		defer repo.acquireRef(ref.Name)()

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

		// Add watchers.
		if watchers := c.server.watchers(params); len(watchers) > 0 {
			res.Watchers = make([]string, len(watchers))
			for i, wc := range watchers {
				res.Watchers[i] = wc.init.ID
			}
			sort.Strings(res.Watchers)
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
		logger = log.With(logger, "repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, logger, params.RefIdentifier.Repo)
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
			finalConfig, err := repo.mergedConfigNoLock(*config)
			if err != nil {
				return err
			}
			if remoteName := params.RefConfiguration.Upstream; remoteName != "" {
				remote, ok := finalConfig.Remotes[remoteName]
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

		newConfig, err = repo.getConfig()
		if err != nil {
			return nil, err
		}
		if err := c.server.applyRepoConfiguration(ctx, logger, params.RefIdentifier.Repo, repo, oldConfig, newConfig); err != nil {
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
		logger = log.With(logger, "repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, logger, params.RefIdentifier.Repo)
		if err != nil {
			return nil, err
		}
		if err := c.server.handleRefUpdateFromDownstream(ctx, logger, repo, params, c, true, true); err != nil {
			level.Error(logger).Log("params", params, "err", err)
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
		logger = log.With(logger, "repo", params.Repo, "ref", params.Ref)
		repo, err := c.server.getRepo(ctx, logger, params.RefIdentifier.Repo)
		if err != nil {
			return nil, err
		}
		if err := c.server.handleSymbolicRefUpdate(ctx, logger, c, repo, params); err != nil {
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
		var params DebugLogParams
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
		res, err := c.handleWorkspaceServerMethod(ctx, logger, conn, req)
		if err != errNotHandled {
			return res, err
		}

		if extHandler, present := serverExtensions[req.Method]; present {
			return extHandler(ctx, conn, req)
		}

		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not found: %q", req.Method)}
	}
}

// refsInRepo returns all references in the given repo.
func (c *serverConn) refsInRepo(repoName string, repo *serverRepo) []RefInfo {
	var refs []RefInfo
	for _, ref := range repo.refdb.List("*") {
		info := RefInfo{RefIdentifier: RefIdentifier{Repo: repoName, Ref: ref.Name}}
		if ref.IsSymbolic() {
			info.Target = ref.Target
		} else {
			release := repo.acquireRef(ref.Name)
			refObj := ref.Object.(serverRef)
			info.State = &RefState{
				RefBaseInfo: RefBaseInfo{
					GitBase:   refObj.gitBase,
					GitBranch: refObj.gitBranch,
				},
				History: refObj.history(),
			}
			release()
		}

		watchers := c.server.watchers(RefIdentifier{Repo: repoName, Ref: ref.Name})
		info.Watchers = make([]string, len(watchers))
		for i, wc := range watchers {
			info.Watchers[i] = wc.init.ID
		}
		sort.Strings(info.Watchers)

		refs = append(refs, info)
	}
	return refs
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
	return c.conn.Close()
}
