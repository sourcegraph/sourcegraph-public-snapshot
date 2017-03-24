package zap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/ot"
	"github.com/sourcegraph/zap/pkg/config"
	"github.com/sourcegraph/zap/pkg/errorlist"
	"github.com/sourcegraph/zap/pkg/fpath"
	"github.com/sourcegraph/zap/server/refdb"
)

// InitWorkspaceServer creates a workspace server on this server to
// handle workspace/* requests.
func (s *Server) InitWorkspaceServer(newWorkspace func(ctx context.Context, logger log.Logger, dir string) (Workspace, *RepoConfiguration, error)) {
	s.workspaceServer = &workspaceServer{
		parent:       s,
		NewWorkspace: newWorkspace,
	}
}

var errNotHandled = errors.New("method not handled by server extension")

// Workspace represents a watched directory tree.
type Workspace interface {
	// Apply applies an operation to the workspace.
	Apply(context.Context, log.Logger, ot.WorkspaceOp) error

	// Checkout checks out a new Zap branch in the workspace. If a
	// conflict occurs, an error of type
	// *WorkspaceCheckoutConflictError is returned describing the
	// conflict.
	//
	// It returns the op equivalent to a1 =
	// transform(composed(history), currentWorkspaceState). The
	// returned op should be recorded on the ref by the caller to
	// Checkout. If keepLocalChanges is false, the op is a noop.
	//
	// The updateExternal func is called AFTER ensuring that the
	// checkout will succeed (i.e., there is no conflict between the
	// workspace's current worktree and the new branch) and BEFORE
	// making any modifications to files on disk or Git state.
	Checkout(ctx context.Context, logger log.Logger, keepLocalChanges bool, ref, gitBase, gitBranch string, history []ot.WorkspaceOp, updateExternal func(ctx context.Context) error) (ot.WorkspaceOp, error)

	// ResetToCurrentState returns a sequence of ops that, when
	// applied to the gitBaseCommit, would yield the exact current
	// workspace state plus the state of buffered files in
	// bufferFiles. (If bufferFiles is nil, the workspace's existing
	// known buffer file contents are used.)
	ResetToCurrentState(ctx context.Context, logger log.Logger, gitBaseCommit string, bufferFiles map[string]string) ([]ot.WorkspaceOp, error)

	// Configure updates the configuration for the repository and
	// workspace.
	Configure(context.Context, RepoConfiguration) error

	// WillSaveFile indicates that the file will be saved by the
	// editor soon. The workspace should ignore edit ops from file
	// system changes until after the next save op. This lets us avoid
	// double-applying after an editor save (the Zap editor extension
	// sends a "save" op and the file system watcher notices an "edit"
	// op).
	WillSaveFile(relativePath string)

	// DefaultRepoName returns the default name that should be used to
	// represent the repository for the given remote.
	DefaultRepoName(remote string) (string, error)

	// Op returns a channel that receives ops describing changes made
	// to the workspace's file system or git HEAD branch tip.
	Op() <-chan ot.WorkspaceOp

	// Reset returns a channel that receives an op whenever the Zap
	// ref is reset (e.g., after the user runs "git checkout" to
	// switch to a new git branch).
	Reset() <-chan WorkspaceResetInfo

	// ConfigChange returns a channel that receives an op whenever the
	// workspace configuration changes. The value sent on the channel
	// is the new configuration.
	//
	// NOTE: Bare (non-workspace) repos also could have configuration,
	// but right now the configuration options only make sense for
	// workspaces, so configuration change listening is coupled to
	// workspaces.
	ConfigChange() <-chan RepoConfiguration

	// Close stops the workspace.
	Close() error

	// CloseNotify returns a channel that is closed when the workspace
	// is closed (either because Close() was called or because the
	// workspace was deleted or became inaccessible on disk).
	CloseNotify() <-chan struct{}

	// Err returns the error, if any, that caused the workspace to
	// close. Err panics if the workspace is not yet closed. Callers
	// should ensure the workspace is closed by waiting to receive on
	// the CloseNotify() channel before calling Err.
	Err() error
}

// WorkspaceCheckoutConflictError indicates that a workspace checkout
// failed because the branch to check out has changes that conflict
// with the workspace's state on disk. This happens, for example, when
// both the branch and the workspace have different changes to the
// same file.
type WorkspaceCheckoutConflictError struct {
	Ref         string         `json:"ref"`         // the ref name (e.g., "branch/mybranch")
	GitBase     string         `json:"gitBase"`     // the branch's Git base commit
	BranchState ot.WorkspaceOp `json:"branchState"` // the composed branch history ops vs. the branch's Git base commit
	Diff        ot.WorkspaceOp `json:"diff"`        // op describing workspace state vs. the branch state
}

func (e *WorkspaceCheckoutConflictError) Error() string {
	return fmt.Sprintf("conflict checking out ref %q to workspace (branch %v and diff %v)", e.Ref, e.BranchState, e.Diff)
}

// WorkspaceResetInfo describes a workspace reset that occurred. A
// workspace reset is when the workspace's HEAD changes (i.e., it
// points to a different Zap ref).
type WorkspaceResetInfo struct {
	Ref         string `json:"ref"` // new Zap ref; might be the same as the old branch if the history is just being reset
	RefBaseInfo        // new base to reset the ref to
	Overwrite   bool   // whether a new Zap ref should be created (possibly clobbering what exists)
}

func (w WorkspaceResetInfo) String() string {
	s := fmt.Sprintf("%s git(%s:%s)", w.Ref, w.GitBranch, abbrevGitOID(w.GitBase))
	if w.Overwrite {
		s += " overwrite"
	}
	return s
}

type workspaceServer struct {
	parent *Server

	NewWorkspace func(ctx context.Context, logger log.Logger, dir string) (Workspace, *RepoConfiguration, error)
}

var mockWorkspaceHandled chan struct{}

func (s *workspaceServer) handleWorkspaceTasks(ctx context.Context, repo *serverRepo, w WorkspaceIdentifier, workspace Workspace, ready chan error) {
	// Create a clean logger, because it will be reused across
	// requests.
	logger := log.With(s.parent.baseLogger(), "watching-workspace", w.Dir)

	loop := func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				level.Debug(logger).Log("done", ctx.Err())
				return ctx.Err()

			case op, ok := <-workspace.Op():
				if !ok {
					return workspace.Err()
				}
				logger := log.With(logger, "op", op.String())
				level.Info(logger).Log()

				ref, err := repo.refdb.Resolve("HEAD")
				if err != nil {
					return err
				}
				do := func() error {
					// Wrap in func so we can defer here.
					defer repo.acquireRef(ref.Name)()

					refObj := ref.Object.(serverRef)
					err = refObj.ot.Record(logger, op)
					if err != nil {
						return err
					}
					if err := s.parent.broadcastRefUpdate(ctx, logger, []refdb.Ref{*ref}, nil, &RefUpdateDownstreamParams{
						RefIdentifier: w.Ref("HEAD"),
						Current:       &RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
						Op:            &op,
					}, nil); err != nil {
						return fmt.Errorf("after workspace op, broadcasting ref update: %s", err)
					}
					return nil
				}
				if err := do(); err != nil {
					return err
				}

			case resetInfo, ok := <-workspace.Reset():
				if !ok {
					return workspace.Err()
				}
				logger := log.With(logger, "reset", resetInfo)
				level.Info(logger).Log()

				if resetInfo.Ref == "" {
					panic("empty ref")
				}
				if resetInfo.RefBaseInfo == (RefBaseInfo{}) {
					panic("empty ref base info")
				}

				do := func() error {
					// Wrap in func so we can defer here.
					defer repo.acquireRef(resetInfo.Ref)()

					ref := repo.refdb.Lookup(resetInfo.Ref)
					var current *RefPointer
					if ref != nil {
						refObj := ref.Object.(serverRef)
						current = &RefPointer{
							RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
							Rev:         refObj.rev(),
						}
						// if refObj.ot.Apply != nil {
						// 	panic("Apply func is already set")
						// }
						refObj.ot.Apply = func(logger log.Logger, op ot.WorkspaceOp) error {
							return workspace.Apply(ctx, logger, op)
						}
					}

					// Automatically configure new workspace HEADs
					// with an upstream.
					if err := s.parent.automaticallyConfigureRefUpstream(ctx, logger, w.Dir, repo, resetInfo.Ref); err != nil {
						return err
					}
					level.Debug(logger).Log("POST-resetInfo.Overwrite", resetInfo.Overwrite, "config", repo.config)
					if resetInfo.Overwrite {
						newRefState := &RefState{
							RefBaseInfo: resetInfo.RefBaseInfo,
						}
						if err := s.parent.handleRefUpdateFromDownstream(ctx, logger, repo, RefUpdateUpstreamParams{
							RefIdentifier: w.Ref(resetInfo.Ref),
							Current:       current,
							State:         newRefState,
						}, nil, false, false /* do not double-acquire ref */); err != nil {
							return fmt.Errorf("after workspace reset, updating ref %q (current %v) to %v: %s", resetInfo.Ref, current, newRefState, err)
						}
					}
					if ref == nil {
						ref := repo.refdb.Lookup(resetInfo.Ref)
						if ref == nil {
							panic("ref was not created: " + resetInfo.Ref)
						}
						refObj := ref.Object.(serverRef)
						// if refObj.ot.Apply != nil {
						// 	panic("Apply func is already set")
						// }
						refObj.ot.Apply = func(logger log.Logger, op ot.WorkspaceOp) error {
							return workspace.Apply(ctx, logger, op)
						}
					}

					var oldTarget string
					if oldRef := repo.refdb.Lookup("HEAD"); oldRef != nil {
						oldTarget = oldRef.Target
					}
					if resetInfo.Ref == oldTarget {
						level.Warn(logger).Log("HEAD-symbolic-ref-already-points-to", resetInfo.Ref)
					}
					if err := s.parent.handleSymbolicRefUpdate(ctx, logger, nil, repo, RefUpdateSymbolicParams{
						RefIdentifier: w.Ref("HEAD"),
						Target:        resetInfo.Ref,
						OldTarget:     oldTarget,
					}); err != nil {
						return fmt.Errorf("after workspace reset, updating HEAD from %q to %q: %s", oldTarget, resetInfo.Ref, err)
					}

					if ready != nil {
						close(ready)
						ready = nil
					}
					return nil
				}
				if err := do(); err != nil {
					return err
				}

			case newConfig, ok := <-workspace.ConfigChange():
				if !ok {
					return workspace.Err()
				}
				repo.mu.Lock()
				oldConfig := repo.config.deepCopy()
				repo.config = newConfig
				newConfig, err := repo.getConfigNoLock()
				repo.mu.Unlock()
				if err != nil {
					return err
				}

				logger := log.With(logger, "new-config", newConfig, "old-config", oldConfig)
				// level.Info(log).Log()
				if err := s.parent.applyRepoConfiguration(ctx, logger, w.Dir, repo, oldConfig, newConfig); err != nil {
					return err
				}

			case <-workspace.CloseNotify():
				level.Warn(logger).Log("close-notified-with-error", workspace.Err())
				return workspace.Err()
			}

			// For use in tests, to resume running a test when this loop
			// iteration is complete.
			if mockWorkspaceHandled != nil {
				mockWorkspaceHandled <- struct{}{}
			}
		}
	}
	if err := loop(); err != nil {
		if ready != nil {
			ready <- err
			close(ready)
			ready = nil
		}
		if err != context.Canceled {
			level.Error(logger).Log("ended-with-error", err)
		}
	} else if err == nil && ready != nil {
		panic("ready not yet signaled: " + err.Error())
	}
	s.removeWorkspace(logger, w, repo)
}

var (
	errWorkspaceIdentifierRequired = &jsonrpc2.Error{
		Code:    int64(ErrorCodeWorkspaceIdentifierRequired),
		Message: "workspace identifier required",
	}
)

type workspaceServerConn struct {
	parent *serverConn
}

func (c *serverConn) handleWorkspaceServerMethod(ctx context.Context, logger log.Logger, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	ws := c.server.workspaceServer

	if req.Params != nil {
		var partialParams WorkspaceIdentifier
		if err := json.Unmarshal(*req.Params, &partialParams); err != nil {
			panic(err)
		}
		logger = log.With(logger, "ws", partialParams.Dir)
	}

	switch req.Method {
	case "workspace/add":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params WorkspaceAddParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if params.WorkspaceIdentifier == (WorkspaceIdentifier{}) {
			return nil, errWorkspaceIdentifierRequired
		}
		if err := ws.addWorkspace(logger, params); err != nil {
			return nil, err
		}
		return WorkspaceAddResult{}, nil

	case "workspace/remove":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params WorkspaceRemoveParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if params.WorkspaceIdentifier == (WorkspaceIdentifier{}) {
			return nil, errWorkspaceIdentifierRequired
		}
		repo, err := ws.parent.getWorkspaceRepo(ctx, logger, params.WorkspaceIdentifier)
		if err != nil {
			return nil, err
		}
		err = ws.removeWorkspace(logger, params.WorkspaceIdentifier, repo)
		if err != nil {
			return nil, err
		}
		if err := config.EnsureWorkspaceNotInGlobalConfig(params.WorkspaceIdentifier.Dir); err != nil {
			return nil, err
		}
		return WorkspaceRemoveResult{}, nil

	case "workspace/status":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params WorkspaceStatusParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if params.WorkspaceIdentifier == (WorkspaceIdentifier{}) {
			return nil, errWorkspaceIdentifierRequired
		}
		if _, err := ws.parent.getWorkspaceRepo(ctx, logger, params.WorkspaceIdentifier); err != nil {
			return nil, err
		}
		// TODO(sqs): this is not fully implemented or useful yet
		return &ShowStatusParams{Message: "Monitoring for file system changes", Type: StatusTypeOK}, nil

	case "workspace/checkout":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params WorkspaceCheckoutParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if params.WorkspaceIdentifier == (WorkspaceIdentifier{}) {
			return nil, errWorkspaceIdentifierRequired
		}

		logger = log.With(logger, "checkout-ref", params.Ref)

		repo, err := ws.parent.getWorkspaceRepo(ctx, logger, params.WorkspaceIdentifier)
		if err != nil {
			return nil, err
		}

		var oldTarget string
		if oldRef := repo.refdb.Lookup("HEAD"); oldRef != nil {
			oldTarget = oldRef.Target
		}

		defer repo.acquireRef(params.Ref)()

		ref := repo.refdb.Lookup(params.Ref)
		if ref == nil {
			return nil, &jsonrpc2.Error{
				Code:    int64(ErrorCodeRefNotExists),
				Message: fmt.Sprintf("checkout of nonexistent ref %q", params.Ref),
			}
		}
		repo.mu.Lock()
		workspace := repo.workspace
		repo.mu.Unlock()
		refObj := ref.Object.(serverRef)
		gitBase := refObj.gitBase
		gitBranch := refObj.gitBranch
		history := refObj.history()
		refObj.ot.Apply = func(logger log.Logger, op ot.WorkspaceOp) error {
			return workspace.Apply(ctx, logger, op)
		}

		updateExternal := func(ctx context.Context) error {
			if oldTarget != params.Ref {
				if err := ws.parent.handleSymbolicRefUpdate(ctx, logger, c, repo, RefUpdateSymbolicParams{
					RefIdentifier: params.WorkspaceIdentifier.Ref("HEAD"),
					Target:        params.Ref,
					OldTarget:     oldTarget,
				}); err != nil {
					return err
				}
			}
			return nil
		}
		recordOp, err := workspace.Checkout(ctx, logger, true, params.Ref, gitBase, gitBranch, history, updateExternal)
		if err != nil {
			return nil, err
		}

		if !recordOp.Noop() {
			if err := refObj.ot.Record(logger, recordOp); err != nil {
				return nil, err
			}
			if err := ws.parent.broadcastRefUpdate(ctx, logger, []refdb.Ref{*ref}, nil, &RefUpdateDownstreamParams{
				RefIdentifier: params.WorkspaceIdentifier.Ref("HEAD"),
				Current:       &RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
				Op:            &recordOp,
			}, nil); err != nil {
				return nil, fmt.Errorf("after checkout, broadcasting ref update: %s", err)
			}
		}

		// Ensure local ref matches tracking ref if no history is passed.
		if len(history) == 0 {
			repoConfig, err := repo.getConfig()
			if err != nil {
				return nil, err
			}
			config := repoConfig.Refs[params.Ref]
			trackingBranchRefName := remoteTrackingBranchRef(config.Upstream, params.Ref)
			trackingRef := repo.refdb.Lookup(trackingBranchRefName)
			if trackingRef != nil {
				if err := ws.parent.reconcileRefWithTrackingRef(ctx, logger, repo, params.WorkspaceIdentifier.Ref("HEAD"), *ref, *trackingRef, config); err != nil {
					return nil, err
				}
			}
		}

		return nil, nil

	case "workspace/willSaveFile":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params WorkspaceWillSaveFileParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		_, workspace, repoName, relPath, err := c.server.getWorkspaceForFileURI(params.URI)
		if err != nil {
			return nil, err
		}
		level.Debug(logger).Log("ws", repoName, "will-save-file", relPath)
		workspace.WillSaveFile(relPath)
		return nil, nil

	case "workspace/reset":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params WorkspaceResetParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if params.WorkspaceIdentifier == (WorkspaceIdentifier{}) {
			return nil, errWorkspaceIdentifierRequired
		}

		logger = log.With(logger, "reset-ref", params.Ref)
		level.Info(logger).Log()

		repo, err := ws.parent.getWorkspaceRepo(ctx, logger, params.WorkspaceIdentifier)
		if err != nil {
			return nil, err
		}

		ref, err := repo.refdb.Resolve("HEAD")
		if err != nil {
			return nil, &jsonrpc2.Error{
				Code:    int64(ErrorCodeSymbolicRefInvalid),
				Message: fmt.Sprintf("symbolic ref resolution error: %s", err),
			}
		}
		// Consistency check.
		if ref.Name != params.Ref {
			return nil, &jsonrpc2.Error{
				Code:    int64(ErrorCodeRefConflict),
				Message: fmt.Sprintf("invalid ref %q for workspace reset (expected %q, which is the current HEAD)", params.Ref, ref.Name),
			}
		}

		repo.mu.Lock()
		workspace := repo.workspace
		repo.mu.Unlock()

		defer repo.acquireRef(ref.Name)()
		serverRef := ref.Object.(serverRef)

		// Synthesize a ref/update operation that resets the workspace.
		resetOps, err := workspace.ResetToCurrentState(ctx, logger, serverRef.gitBase, params.BufferFiles)
		if err != nil {
			return nil, err
		}
		refState := &RefState{
			RefBaseInfo: RefBaseInfo{GitBase: serverRef.gitBase, GitBranch: serverRef.gitBranch},
			History:     resetOps,
		}
		if err := ws.parent.handleRefUpdateFromDownstream(ctx, logger, repo, RefUpdateUpstreamParams{
			RefIdentifier: params.WorkspaceIdentifier.Ref(params.Ref),
			Force:         true,
			State:         refState,
		}, c.parent, false, false); err != nil {
			return nil, err
		}
		return refState, nil
	}
	return nil, errNotHandled
}

func (s *workspaceServer) addWorkspace(logger log.Logger, params WorkspaceAddParams) error {
	// Allow upgrading a bare repo to a workspace repo.
	s.parent.reposMu.Lock()
	repo, exists := s.parent.repos[fpath.Key(params.Dir)]
	s.parent.reposMu.Unlock()
	if exists {
		level.Info(logger).Log("create-workspace-in-existing-repo", "")
	} else {
		repo = &serverRepo{repoDir: params.Dir, refdb: refdb.NewMemoryRefDB()}
		level.Info(logger).Log("create-workspace-in-new-repo", "")
	}

	ctx, cancel := context.WithCancel(s.parent.bgCtx)
	workspace, cfg, err := s.NewWorkspace(ctx, log.With(s.parent.baseLogger(), "workspace", params.Dir), params.Dir)
	if err != nil {
		cancel()
		return err
	}
	{
		repo.mu.Lock()
		if repo.workspace != nil {
			repo.mu.Unlock()
			cancel()
			return &jsonrpc2.Error{
				Code:    int64(ErrorCodeWorkspaceExists),
				Message: fmt.Sprintf("already added workspace %v", params.WorkspaceIdentifier),
			}
		}
		repo.config = *cfg
		repo.workspace = workspace
		repo.workspaceCancel = cancel
		repo.mu.Unlock()
	}

	{
		// Add repo to server.
		s.parent.reposMu.Lock()
		if _, exists := s.parent.repos[fpath.Key(params.Dir)]; exists {
			s.parent.reposMu.Unlock()
			cancel()
			return &jsonrpc2.Error{
				Code:    int64(ErrorCodeRepoNotExists),
				Message: fmt.Sprintf("during workspace initialization, repo was added by someone else: %v", params.WorkspaceIdentifier),
			}
		}
		s.parent.repos[fpath.Key(params.Dir)] = repo
		s.parent.reposMu.Unlock()
	}

	newConfig, err := repo.getConfig()
	if err != nil {
		return err
	}
	if err := s.parent.applyRepoConfiguration(ctx, logger, params.Dir, repo, RepoConfiguration{}, newConfig); err != nil {
		level.Error(logger).Log("apply-initial-workspace-configuration-error", err)
		cancel()
		return err
	}

	// Block until workspace is ready (or fails to become ready).
	ready := make(chan error)
	go s.handleWorkspaceTasks(s.parent.bgCtx, repo, params.WorkspaceIdentifier, workspace, ready)
	select {
	case err := <-ready:
		if err != nil {
			cancel()
			return fmt.Errorf("workspace %q failed to become ready: %s", params.Dir, err)
		}
	case <-time.After(15 * time.Second):
		cancel()
		return fmt.Errorf("workspace %q failed to become ready after 15s", params.Dir)
	}

	return config.EnsureWorkspaceInGlobalConfig(params.Dir)
}

func (s *workspaceServer) loadWorkspacesFromConfig(logger log.Logger) error {
	// Read the current config file, if it exists.
	cfg, err := config.ReadGlobalFile()
	if err != nil {
		return err
	}

	var (
		wg   sync.WaitGroup
		errs errorlist.Errors
	)
	for _, opt := range cfg.Section("workspaces").Options {
		if opt.Key != "workspace" {
			continue
		}
		if opt.Value == "" {
			level.Warn(logger).Log("skip-invalid-workspace", "")
			continue
		}
		opt := opt
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if _, err := os.Stat(path); err != nil {
				level.Warn(logger).Log("skip-inaccessible-workspace", path, "err", err)
				return
			}
			if err := s.addWorkspace(logger, WorkspaceAddParams{WorkspaceIdentifier{Dir: path}}); err != nil {
				errs.Add(err)
			}
		}(opt.Value)
	}
	wg.Wait()
	return errs.Error()
}

func (s *workspaceServer) removeWorkspace(logger log.Logger, workspace WorkspaceIdentifier, repo *serverRepo) error {
	level.Info(logger).Log("rm-workspace", workspace.Dir)

	s.parent.reposMu.Lock()
	delete(s.parent.repos, fpath.Key(workspace.Dir))
	s.parent.reposMu.Unlock()

	repo.mu.Lock()
	if repo.workspaceCancel != nil {
		repo.workspaceCancel()
		repo.workspaceCancel = nil
	}
	repo.workspace = nil
	repo.config = RepoConfiguration{}
	repo.mu.Unlock()
	return nil
}

func (s *Server) getWorkspaceForFileURI(uriStr string) (repo *serverRepo, workspace Workspace, repoName string, relPath string, err error) {
	uri, err := url.Parse(uriStr)
	if uri.Scheme != "file" {
		return nil, nil, "", "", fmt.Errorf("only file URIs are supported: %q", uriStr)
	}
	if !filepath.IsAbs(uri.Path) {
		return nil, nil, "", "", fmt.Errorf("file URI must be absolute: %q", uriStr)
	}
	s.reposMu.Lock()
	defer s.reposMu.Unlock()
	for _, repo := range s.repos {
		if strings.HasPrefix(uri.Path, repo.repoDir+string(os.PathSeparator)) {
			repo.mu.Lock()
			workspace := repo.workspace
			repo.mu.Unlock()
			if workspace != nil {
				return repo, workspace, repo.repoDir, strings.TrimPrefix(uri.Path, repo.repoDir+string(os.PathSeparator)), nil
			}
		}
	}
	return nil, nil, "", "", &jsonrpc2.Error{
		Code:    int64(ErrorCodeWorkspaceNotExists),
		Message: fmt.Sprintf("no workspace found for file %q", uriStr),
	}
}

func (s *Server) getWorkspaceRepo(ctx context.Context, logger log.Logger, w WorkspaceIdentifier) (*serverRepo, error) {
	repo, err := s.getRepoIfExists(ctx, logger, w.Dir)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, &jsonrpc2.Error{
			Code:    int64(ErrorCodeWorkspaceNotExists),
			Message: fmt.Sprintf("workspace not found: %s (add it with 'zap init')", w.Dir),
		}
	}
	repo.mu.Lock()
	isWorkspace := repo.workspace != nil
	repo.mu.Unlock()
	if !isWorkspace {
		return nil, &jsonrpc2.Error{
			Code:    int64(ErrorCodeWorkspaceNotExists),
			Message: fmt.Sprintf("repo at %q is not configured as a workspace (run 'zap init')", w.Dir),
		}
	}
	return repo, nil
}
