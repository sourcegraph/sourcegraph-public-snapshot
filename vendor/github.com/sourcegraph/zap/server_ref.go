package zap

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
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

func (s *Server) doUpdateBulkRefConfiguration(ctx context.Context, log *log.Context, repoName string, repo *serverRepo, oldRefs, newRefs map[string]RefConfiguration) error {
	for name, config := range oldRefs {
		// Remove refs that are in oldRefs but not newRefs.
		if _, ok := newRefs[name]; !ok {
			if err := s.doUpdateRefConfiguration(ctx, log, repo, RefIdentifier{Repo: repoName, Ref: name}, repo.refdb.Lookup(name), config, RefConfiguration{}); err != nil {
				return err
			}
		}
	}
	for name, newConfig := range newRefs {
		if oldConfig := oldRefs[name]; oldConfig != newConfig {
			if err := s.doUpdateRefConfiguration(ctx, log, repo, RefIdentifier{Repo: repoName, Ref: name}, repo.refdb.Lookup(name), oldConfig, newConfig); err != nil {
				return err
			}
		}
	}
	return nil
}

// doUpdateRefUpstreamConfiguration sets up the given ref to track a remote
// branch on the given upstream (which is the name of a remote). It
// unwatches any previous tracked branches and watches the new tracked
// branch (if any).
func (s *Server) doUpdateRefConfiguration(ctx context.Context, log *log.Context, repo *serverRepo, refID RefIdentifier, ref *refdb.Ref, oldConfig, newConfig RefConfiguration) error {
	// Assumes caller holds repo.mu.

	var refObj *serverRef
	if ref != nil && ref.Object != nil {
		v := ref.Object.(serverRef)
		refObj = &v
	}

	var oldRemote, newRemote RepoRemoteConfiguration
	if oldConfig.Upstream != "" {
		var ok bool
		oldRemote, ok = repo.config.Remotes[oldConfig.Upstream]
		if !ok {
			level.Warn(log).Log("old-upstream-refers-to-missing-remote", oldConfig.Upstream)
		}
	}
	if newConfig.Upstream != "" {
		var ok bool
		newRemote, ok = repo.config.Remotes[newConfig.Upstream]
		if !ok {
			return &jsonrpc2.Error{
				Code:    int64(ErrorCodeRemoteNotExists),
				Message: fmt.Sprintf("no remote found for new upstream: %s", newConfig.Upstream),
			}
		}
	}

	if oldConfig == newConfig {
		// Unchanged.
		if oldRemote != newRemote {
			panic("oldRemote != newRemote, expected remotes to stay in sync with upstreams")
		}
		return nil
	}

	if oldConfig.Upstream != "" {
		// Stop tracking old upstream.
		if refObj != nil {
			refObj.ot.SendToUpstream = nil
		}
	}

	if newConfig.Upstream != "" && refObj != nil {
		// Start tracking new upstream.
		log := log.With("set-upstream", newConfig.Upstream, "overwrite", newConfig.Overwrite)
		level.Debug(log).Log()

		upstreamRefID := RefIdentifier{Repo: newRemote.Repo, Ref: ref.Name}
		if newConfig.Overwrite {
			cl, ok := s.remotes.getClient(newRemote.Endpoint)
			if !ok {
				panic(fmt.Sprintf("no client for remote %q", newConfig.Upstream))
			}

			newRefState := &RefState{
				RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
				History:     refObj.history(),
			}
			level.Info(log).Log("overwrite-ref-on-upstream", newRefState)
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

			if err := s.reconcileRefWithTrackingRef(ctx, log, repo, refID, *ref, *trackingRef, newConfig); err != nil {
				return err
			}
		}

		refObj.ot.SendToUpstream = func(upstreamRev int, op ot.WorkspaceOp) {
			go func() {
				// TODO(sqs): need to get latest GitBase/GitBranch, not
				// use the ones captured by the closure?
				upstreamRefID := RefIdentifier{Repo: newRemote.Repo, Ref: refID.Ref}

				log := s.baseLogger()
				log = log.With("send-to-upstream", upstreamRefID, "endpoint", newRemote.Endpoint, "remote", newConfig.Upstream, "rev", upstreamRev, "op", op)

				cl, err := s.remotes.getOrCreateClient(ctx, log, newRemote.Endpoint)
				if err != nil {
					level.Error(log).Log("send-to-upstream-failed-to-create-client", err)
					return
				}
				level.Debug(log).Log()
				if err := cl.RefUpdate(ctx, RefUpdateUpstreamParams{
					RefIdentifier: upstreamRefID,
					Current: &RefPointer{
						RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
						Rev:         upstreamRev,
					},
					Op: &op,
				}); err != nil {
					level.Error(log).Log("send-to-upstream-failed", err)
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
							level.Info(log).Log("overwrite-ref-after-error", newRefState)
							if err := cl.RefUpdate(ctx, RefUpdateUpstreamParams{
								RefIdentifier: upstreamRefID,
								Force:         true,
								State:         newRefState,
							}); err == nil {
								if err := refObj.ot.AckFromUpstream(); err != nil {
									level.Error(log).Log("ack-after-overwrite-ref-after-error-failed", err)
								}
							} else {
								level.Error(log).Log("overwrite-ref-after-error-failed", err)
							}
						}
					}

					return
				}
			}()
		}
	}

	if repo.config.Refs == nil {
		repo.config.Refs = map[string]RefConfiguration{}
	}
	repo.config.Refs[refID.Ref] = newConfig

	return nil
}

func (s *Server) reconcileRefWithTrackingRef(ctx context.Context, log *log.Context, repo *serverRepo, localRefID RefIdentifier, local, tracking refdb.Ref, refConfig RefConfiguration) error {
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
			log.With("fast-forward-to-upstream", newRefState),
			repo,
			local,
			RefUpdateDownstreamParams{
				RefIdentifier: localRefID,
				Current:       &RefBaseInfo{GitBase: localObj.gitBase, GitBranch: localObj.gitBranch},
				State:         newRefState,
			},
			refConfig,
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

func (s *Server) findLocalRepo(remoteRepoName, endpoint string) (repo *serverRepo, localRepoName, remoteName string) {
	// TODO(sqs) HACK: this is indicative of a design flaw
	s.reposMu.Lock()
	defer s.reposMu.Unlock()
	var matchingLocalRepos []*serverRepo
	var matchingRemoteNames []string
	for localRepoName2, localRepo := range s.repos {
		localRepo.mu.Lock()
		for remoteName, config := range localRepo.config.Remotes {
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
		return nil, "", ""
	}
	return matchingLocalRepos[0], localRepoName, matchingRemoteNames[0]
}
