package zap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/ot"
	"github.com/sourcegraph/zap/server/refdb"
	"github.com/sourcegraph/zap/ws"
)

// watchers returns a list of connections that are watching this
// ref.
func (s *Server) watchers(ref RefIdentifier) (conns []*serverConn) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	for c := range s.conns {
		c.mu.Lock()
		if c.isWatching(ref) {
			conns = append(conns, c)
		}
		c.mu.Unlock()
	}
	return conns
}

type serverRef struct {
	gitBase, gitBranch string // values given when created via (ServerBackend).Create
	ot                 *ws.Proxy
}

// TODO(sqs): make immutable
func (r serverRef) rev() int {
	if r.ot == nil {
		return 0
	}
	return r.ot.Rev()
}

// TODO(sqs): make immutable
func (r serverRef) history() []ot.WorkspaceOp {
	return r.ot.History()
}

// doApplyRefUpstreamConfiguration sets up the given ref to track a
// remote branch on the given upstream (which is the name of a
// remote). It unwatches any previous tracked branches and watches the
// new tracked branch (if any).
//
// It does not persist the configuration; callers must use
// (*Server).updateRepoConfiguration for that.
//
// If force is true, the changes are applied even if oldConfig ==
// newConfig.
//
// If sendCurrentState, the server immediately sends the current state
// of the ref upstream.
//
// The caller MUST hold the repo.acquireRef lock for the given ref.
func (s *Server) doApplyRefConfiguration(ctx context.Context, logger log.Logger, repo *serverRepo, refID RefIdentifier, ref *refdb.Ref, oldRepoConfig, newRepoConfig RepoConfiguration, force, sendCurrentState, acquireRef bool) error {
	oldConfig := oldRepoConfig.Refs[refID.Ref]
	newConfig := newRepoConfig.Refs[refID.Ref]

	if acquireRef {
		defer repo.acquireRef(refID.Ref)()
	}

	var refObj *serverRef
	if ref != nil && ref.Object != nil {
		v := ref.Object.(serverRef)
		refObj = &v
	}

	upstreamChanged := force || (oldConfig.Upstream != newConfig.Upstream)

	var newRemote RepoRemoteConfiguration
	if newConfig.Upstream != "" {
		var ok bool
		newRemote, ok = newRepoConfig.Remotes[newConfig.Upstream]
		if !ok {
			return &jsonrpc2.Error{
				Code:    int64(ErrorCodeRemoteNotExists),
				Message: fmt.Sprintf("no remote found for new upstream: %s", newConfig.Upstream),
			}
		}
	}

	if upstreamChanged && oldConfig.Upstream != "" {
		// Stop tracking old upstream.
		if refObj != nil {
			refObj.ot.SendToUpstream = nil
		}
	}

	if upstreamChanged && newConfig.Upstream != "" && refObj != nil {
		// Start tracking new upstream.
		logger := log.With(logger, "set-upstream", newConfig.Upstream, "overwrite", newConfig.Overwrite)
		level.Debug(logger).Log()

		upstreamRefID := RefIdentifier{Repo: newRemote.Repo, Ref: ref.Name}
		if sendCurrentState {
			if newConfig.Overwrite {
				cl, err := s.remotes.getOrCreateClient(ctx, logger, newRemote.Endpoint)
				if err != nil {
					return err
				}

				newRefState := &RefState{
					RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
					History:     refObj.history(),
				}
				level.Info(logger).Log("overwrite-ref-on-upstream", newRefState)
				debugSimulateLatency()
				if err := cl.RefUpdate(ctx, RefUpdateUpstreamParams{
					RefIdentifier: upstreamRefID,
					Force:         true,
					State:         newRefState,
				}); err != nil {
					return fmt.Errorf("during overwrite-ref-on-upstream of %q with %s: %s", upstreamRefID, newRefState, err)
				}
				refObj.ot.UpstreamRevNumber = len(newRefState.History)
			} else {
				trackingRefName := "refs/remotes/" + newConfig.Upstream + "/" + refID.Ref
				trackingRef := repo.refdb.Lookup(trackingRefName)
				if trackingRef == nil {
					return &jsonrpc2.Error{
						Code:    int64(ErrorCodeRefNotExists),
						Message: fmt.Sprintf("unable to set local ref %q to track nonexistent remote tracking branch %q (set overwrite=true on the ref to create it)", refID.Ref, trackingRefName),
					}
				}

				if err := s.reconcileRefWithTrackingRef(ctx, logger, repo, refID, *ref, *trackingRef, newConfig); err != nil {
					return err
				}
			}
		}

		refObj.ot.SendToUpstream = func(logger log.Logger, upstreamRev int, op ot.WorkspaceOp) {
			// TODO(sqs): need to get latest GitBase/GitBranch, not
			// use the ones captured by the closure?
			upstreamRefID := RefIdentifier{Repo: newRemote.Repo, Ref: refID.Ref}

			logger = log.With(logger, "send-to-upstream", upstreamRefID, "endpoint", newRemote.Endpoint, "remote", newConfig.Upstream, "rev", upstreamRev, "op", op)

			cl, err := s.remotes.getOrCreateClient(ctx, logger, newRemote.Endpoint)
			if err != nil {
				level.Error(logger).Log("send-to-upstream-failed-to-create-client", err)
				return
			}
			level.Debug(logger).Log()
			debugSimulateLatency()

			DebugMu.Lock()
			simulateError := TestSimulateResetAfterErrorInSendToUpstream
			DebugMu.Unlock()
			if simulateError {
				err = errors.New("TestSimulateResetAfterErrorInSendToUpstream")
			} else {
				err = cl.RefUpdate(ctx, RefUpdateUpstreamParams{
					RefIdentifier: upstreamRefID,
					Current: &RefPointer{
						RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
						Rev:         upstreamRev,
					},
					Op: &op,
				})
			}
			if err != nil {
				level.Error(logger).Log("send-to-upstream-failed", err, "op", op)
				if os.Getenv("SEND_TO_UPSTREAM_ERRORS_FATAL") != "" {
					panic(fmt.Sprintf("SendToUpstream(%d, %v) failed: %s", upstreamRev, op, err))
				}

				// If the ref is configured to overwrite and we
				// get an error, wipe the upstream and overwrite
				// it.
				if newConfig.Overwrite {
					if ref := repo.refdb.Lookup(refID.Ref); ref != nil {
						refObj := ref.Object.(serverRef)
						newRefState := &RefState{
							RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
							History:     refObj.history(),
						}
						logger = log.With(logger, "overwrite-ref-after-error", newRefState)
						level.Info(logger).Log()
						if err := cl.RefUpdate(ctx, RefUpdateUpstreamParams{
							RefIdentifier: upstreamRefID,
							Force:         true,
							State:         newRefState,
						}); err == nil {
							if err := refObj.ot.AckFromUpstream(logger); err != nil {
								level.Error(logger).Log("ack-after-overwrite-ref-after-error-failed", err)
							}
						} else {
							level.Error(logger).Log("overwrite-ref-after-error-failed", err)
						}
					}
				}

				return
			}
		}
	}
	return nil
}

func (s *Server) automaticallyConfigureRefUpstream(ctx context.Context, logger log.Logger, repoName string, repo *serverRepo, ref string) error {
	config, err := repo.getConfigNoLock()
	if err != nil {
		return err
	}
	if len(config.Remotes) == 0 {
		return nil // unable to autoconfigure
	}
	if _, ok := config.Refs[ref]; ok {
		return nil // ref already has configuration; do nothing
	}
	upstream, err := config.defaultUpstream()
	if err != nil {
		return err
	}
	oldConfig, newConfig, err := s.updateRepoConfiguration(ctx, repo, func(config *RepoConfiguration) error {
		if config.Refs == nil {
			config.Refs = map[string]RefConfiguration{}
		}
		config.Refs[ref] = RefConfiguration{Upstream: upstream, Overwrite: true}
		return nil
	})
	if err != nil {
		return err
	}
	finalConfig, err := repo.mergedConfigNoLock(newConfig)
	if err != nil {
		return err
	}
	return s.doApplyRefConfiguration(ctx, logger, repo, RefIdentifier{Repo: repoName, Ref: ref}, repo.refdb.Lookup(ref), oldConfig, finalConfig, false, false, false)
}

// TODO(sqs): hack to "safely" determine a default upstream, until we
// have a full config for this.
func (c *RepoConfiguration) defaultUpstream() (string, error) {
	if len(c.Remotes) == 1 {
		for k := range c.Remotes {
			return k, nil
		}
	}
	return "", errors.New("unable to determine branch's default upstream: more than 1 remote exists")
}

func (s *Server) reconcileRefWithTrackingRef(ctx context.Context, logger log.Logger, repo *serverRepo, localRefID RefIdentifier, local, tracking refdb.Ref, refConfig RefConfiguration) error {
	// Assumes caller holds the repo.refAcquire(ref.Ref) lock.

	localObj := local.Object.(serverRef)
	trackingObj := tracking.Object.(serverRef)
	// Check that the localObj tracking ref can be fast-forwarded to the tracking ref.
	if localObj.gitBase != trackingObj.gitBase {
		return &jsonrpc2.Error{
			Code:    int64(ErrorCodeRefConflict),
			Message: fmt.Sprintf("local ref %s git base %q != remote tracking branch git base %q", local.Name, localObj.gitBase, trackingObj.gitBase),
		}
	}
	if localObj.gitBranch != trackingObj.gitBranch {
		return &jsonrpc2.Error{
			Code:    int64(ErrorCodeRefConflict),
			Message: fmt.Sprintf("local ref %s git branch %q != remote tracking branch git branch %q", local.Name, localObj.gitBranch, trackingObj.gitBranch),
		}
	}
	// TODO(sqs): allow if local tracking ref is strictly behind
	// tracking ref.
	if len(trackingObj.history()) != 0 && localObj.rev() != 0 && !reflect.DeepEqual(trackingObj.history(), localObj.history()) {
		return &jsonrpc2.Error{
			Code:    int64(ErrorCodeRefConflict),
			Message: fmt.Sprintf("local ref %s history conflicts with remote tracking branch's history", local.Name),
		}
	}

	// Fast forward our ref and downstream refs.
	if localObj.rev() == 0 && trackingObj.rev() > 0 {
		newRefState := &RefState{
			RefBaseInfo: RefBaseInfo{GitBase: trackingObj.gitBase, GitBranch: trackingObj.gitBranch},
			History:     trackingObj.history(),
		}
		if err := s.updateLocalTrackingRefAfterUpstreamUpdate(ctx,
			log.With(logger, "fast-forward-to-upstream", newRefState),
			repo,
			local,
			RefUpdateDownstreamParams{
				RefIdentifier: localRefID,
				Current:       &RefBaseInfo{GitBase: localObj.gitBase, GitBranch: localObj.gitBranch},
				State:         newRefState,
			},
			refConfig,
			false,
		); err != nil {
			return err
		}
	}

	return nil
}

// isLikelySymbolicRef returns true if name is all uppercase ASCII
// characters.
func isLikelySymbolicRef(name string) bool {
	for _, r := range name {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func remoteTrackingRef(remote, ref string) string {
	return "refs/remotes/" + remote + "/" + ref
}

func (s *Server) findLocalRepo(remoteRepoName, endpoint string) (repo *serverRepo, localRepoName, remoteName string, err error) {
	// TODO(sqs) HACK: this is indicative of a design flaw
	s.reposMu.Lock()
	defer s.reposMu.Unlock()
	var matchingLocalRepos []*serverRepo
	var matchingRemoteNames []string
	for localRepoName2, localRepo := range s.repos {
		localRepo.mu.Lock()
		localRepoConfig, err := localRepo.getConfigNoLock()
		if err != nil {
			return nil, "", "", err
		}
		for remoteName, config := range localRepoConfig.Remotes {
			if config.Endpoint == endpoint && config.Repo == remoteRepoName {
				matchingLocalRepos = append(matchingLocalRepos, localRepo)
				matchingRemoteNames = append(matchingRemoteNames, remoteName)
				localRepoName = localRepoName2
			}
		}
		localRepo.mu.Unlock()
	}
	if len(matchingLocalRepos) > 1 {
		panic(fmt.Sprintf("more than 1 local repo is tracking a remote repo %q at endpoint %q", remoteRepoName, endpoint))
	}
	if len(matchingLocalRepos) == 0 {
		return nil, "", "", nil
	}
	return matchingLocalRepos[0], localRepoName, matchingRemoteNames[0], nil
}
