package server

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/server/refstate"
	"github.com/sourcegraph/zap/server/repodb"
)

// updateRepoConfiguration is called with a callback to persist new
// configuration. It does not actually apply the configuration update;
// you must call applyRepoConfig to apply it.
//
// This is intentional. We want to avoid a situation where some of the
// config is updated but not all of it, and it ends up in a
// configuration state that the user never specified. This way at
// least the saved config will always reflect explicit user input. (We
// don't yet support (and probably never will support) atomically
// setting the configuration, as we would need to wait on many remote
// things to truly know if the new configuration is good.)
func (s *Server) updateRepoConfiguration(ctx context.Context, repo repodb.OwnedRepo, updateFunc func(config *zap.RepoConfiguration) error) (old, new zap.RepoConfiguration, err error) {
	old = repo.Config.DeepCopy()
	new = repo.Config.DeepCopy()
	if err := updateFunc(&new); err != nil {
		return zap.RepoConfiguration{}, zap.RepoConfiguration{}, err
	}

	// Call extensions.
	//
	// TODO(sqs): roll back if any ext.ConfigureRepo calls fail.
	for _, ext := range s.ext {
		if ext.ConfigureRepo != nil {
			if err := ext.ConfigureRepo(ctx, &repo, old, new); err != nil {
				return old, new, err
			}
		}
	}

	// Save the new config in the repodb.
	repo.Config = new

	return old, new, nil
}

func (s *Server) ApplyRepoConfiguration(ctx context.Context, logger log.Logger, repo repodb.OwnedRepo, oldConfig, newConfig zap.RepoConfiguration) error {
	// Forbid multiple remotes having the same endpoint.
	seenEndpoints := make(map[string][]string, len(oldConfig.Remotes))
	for name, config := range newConfig.Remotes {
		seenEndpoints[config.Endpoint] = append(seenEndpoints[config.Endpoint], name)
	}
	for endpoint, remoteNames := range seenEndpoints {
		if len(remoteNames) > 1 {
			return &jsonrpc2.Error{
				Code:    int64(zap.ErrorCodeInvalidConfig),
				Message: fmt.Sprintf("invalid configuration: remote endpoint %q used by more than one remote: %v", endpoint, remoteNames),
			}
		}
	}

	// Disconnect from removed remote endpoints and connect/reconfigure added or changed remotes.
	for oldName, oldRemote := range oldConfig.Remotes {
		if newRemote, ok := newConfig.Remotes[oldName]; ok && newRemote.Endpoint == oldRemote.Endpoint {
			continue // connection remains the same
		}

		// TODO(sqs): this needs to check whether this client is being
		// used by any other repos. It should use reference counting
		// and only remove/close when it gets to 0.
		level.Info(logger).Log("rm-remote", oldName)
		if err := s.remotes.closeAndRemoveClient(oldRemote.Endpoint); err != nil {
			return err
		}
	}
	for newName, newRemote := range newConfig.Remotes {
		oldRemote, ok := oldConfig.Remotes[newName]
		if ok && oldRemote.EquivalentTo(newRemote) {
			continue // unchanged
		}
		logger := log.With(logger, "add-or-update-remote", newName)
		level.Debug(logger).Log("new", newRemote, "old", oldConfig.Remotes[newName])

		if refs := repo.RefDB.List("*"); len(refs) > 1 || (len(refs) == 1 && (repo.WorkspaceRef == "" || refs[0].Name != repo.WorkspaceRef)) {
			// Disallow adding a remote to a repo that already has
			// refs locally. This is because we don't yet want to have
			// to support merging local branches with their
			// upstreams. (It is fine if the repo has exactly 1 ref
			// that is the local head/$CLIENT ref, though, since we
			// always clobber the upstream for that particular ref.)
			//
			// TODO(nick8): By the time we reach this point, the
			// configuration has already been locally
			// persisted. Failing here does not roll back the
			// config. So, we will be left in an inconsistent state
			// after we return an error here. Need to fix this
			// problem.
			return zap.Errorf(zap.ErrorCodeInvalidConfig, "not yet implemented: adding a remote to a repo that already has 1 or more non-head refs")
		}

		cl, err := s.remotes.getOrCreateClient(logger, newRemote.Endpoint)
		if err != nil {
			return err
		}

		repo.SendRefUpdateUpstream = cl.refUpdates

		// Push our head ref upstream if this is a workspace server.
		if repo.WorkspaceRef != "" {
			level.Debug(logger).Log("send-head-ref-upstream", repo.WorkspaceRef)
			workspaceRef := repo.RefDB.Lookup(repo.WorkspaceRef)
			defer workspaceRef.Unlock()
			if workspaceRef.Ref != nil {
				refID := zap.RefIdentifier{
					Repo: newRemote.Repo,
					Ref:  repo.WorkspaceRef, // This does not support following someone else's head ref.
				}
				refState := workspaceRef.Ref.Data.(refstate.RefState)
				upstreamParams := zap.RefUpdateUpstreamParams{
					RefIdentifier: refID,
					Force:         true,
					State:         &refState.RefState,
				}

				// We don't receive an ack for the RefUpdate because we have not started watching it yet.
				// This pretends that the server immediately acks the RefUpdate.
				// This only works because RefUpdate blocks until the server acks the update.
				refState.Upstream = &refstate.Upstream{
					Rev:        uint(len(refState.RefState.Data.History)),
					Send:       cl.refUpdates,
					RemoteRepo: newRemote.Repo,
				}
				if refState.Upstream.RemoteRepo == "" {
					refState.Upstream.RemoteRepo = repo.Path
				}
				workspaceRef.Ref.Data = refState
				if err := repo.RefDB.Write(workspaceRef); err != nil {
					return err
				}

				if err := cl.RefUpdate(ctx, upstreamParams); err != nil {
					return err
				}
			}
		}

		// TODO(sqs): does not correctly clean up repo watches
		// established on previous endpoints or repo names. Kind of an edge case.
		if len(oldRemote.Refspecs) != 0 || len(newRemote.Refspecs) != 0 {
			if err := cl.RepoWatch(ctx, zap.RepoWatchParams{Repo: newRemote.Repo, Refspecs: newRemote.Refspecs}); err != nil {
				return err
			}
		}
	}

	return nil
}
