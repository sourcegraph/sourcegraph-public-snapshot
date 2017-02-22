package zap

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/sourcegraph/jsonrpc2"
)

// getConfig returns a deep copy of the repo configuration.
func (s *serverRepo) getConfig() RepoConfiguration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config.deepCopy()
}

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
func (s *Server) updateRepoConfiguration(ctx context.Context, repo *serverRepo, updateFunc func(config *RepoConfiguration) error) (old, new RepoConfiguration, err error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	old = repo.config.deepCopy()

	new = repo.config.deepCopy()
	if err := updateFunc(&new); err != nil {
		return RepoConfiguration{}, RepoConfiguration{}, err
	}

	// Persist the config.
	if repo.workspace != nil {
		if err := repo.workspace.Configure(ctx, new); err != nil {
			return RepoConfiguration{}, RepoConfiguration{}, err
		}
	}
	repo.config = new

	return old, new, nil
}

func (s *Server) applyRepoConfiguration(ctx context.Context, log *log.Context, repoName string, repo *serverRepo, oldConfig, newConfig RepoConfiguration, force, sendCurrentState bool) error {
	// TODO(sqs): remove force and sendCurrentState options

	if err := s.doApplyBulkRepoRemoteConfiguration(ctx, log, repoName, repo, oldConfig, newConfig); err != nil {
		return err
	}
	if err := s.doApplyBulkRefConfiguration(ctx, log, repoName, repo, oldConfig, newConfig); err != nil {
		return err
	}
	return nil
}

// doApplyBulkRefConfiguration should only be called from
// applyRepoConfiguration.
func (s *Server) doApplyBulkRepoRemoteConfiguration(ctx context.Context, log *log.Context, repoName string, repo *serverRepo, oldRepoConfig, newRepoConfig RepoConfiguration) error {
	oldConfig := oldRepoConfig.Remotes
	newConfig := newRepoConfig.Remotes

	// Forbid multiple remotes having the same endpoint.
	seenEndpoints := make(map[string][]string, len(oldConfig))
	for name, config := range newConfig {
		seenEndpoints[config.Endpoint] = append(seenEndpoints[config.Endpoint], name)
	}
	for endpoint, remoteNames := range seenEndpoints {
		if len(remoteNames) > 1 {
			return &jsonrpc2.Error{
				Code:    int64(ErrorCodeInvalidConfig),
				Message: fmt.Sprintf("invalid configuration: remote endpoint %q used by more than one remote: %v", endpoint, remoteNames),
			}
		}
	}

	// Disconnect from removed remote endpoints and connect/reconfigure added or changed remotes.
	for oldName, oldRemote := range oldConfig {
		if newRemote, ok := newConfig[oldName]; ok && newRemote.Endpoint == oldRemote.Endpoint {
			continue // connection remains the same
		}

		// TODO(sqs): this needs to check whether this client is being
		// used by any other repos. It should use reference counting
		// and only remove/close when it gets to 0.
		level.Info(log).Log("rm-remote", oldName)
		if err := s.remotes.closeAndRemoveClient(oldRemote.Endpoint); err != nil {
			return err
		}
	}
	for newName, newRemote := range newConfig {
		oldRemote, ok := oldConfig[newName]
		if ok && oldRemote.EquivalentTo(newRemote) {
			continue // unchanged
		}
		log := log.With("add-or-update-remote", newName)
		level.Debug(log).Log("new", newRemote, "old", oldConfig[newName])

		cl, err := s.remotes.getOrCreateClient(ctx, log, newRemote.Endpoint)
		if err != nil {
			return err
		}

		// TODO(sqs): does not correctly clean up repo watches
		// established on previous endpoints or repo names. Kind of an edge case.
		if len(oldRemote.Refspecs) != 0 || len(newRemote.Refspecs) != 0 {
			if err := cl.RepoWatch(ctx, RepoWatchParams{Repo: newRemote.Repo, Refspecs: newRemote.Refspecs}); err != nil {
				return err
			}
		}

		// Apply existing ref configuration against new upstream, if
		// the endpoint or repo changed. If just the refspec changed,
		// we don't need to do anything as that would not change the
		// desired state on the upstream.
		if oldRemote.Endpoint != newRemote.Endpoint || oldRemote.Repo != newRemote.Repo {
			for ref, refConfig := range newRepoConfig.Refs {
				if refConfig.Upstream == newName {
					if err := s.doApplyRefConfiguration(ctx, log, repo, RefIdentifier{Repo: repoName, Ref: ref}, repo.refdb.Lookup(ref), oldRepoConfig, newRepoConfig, true /* force */, true, true); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// doApplyBulkRefConfiguration should only be called from
// applyRepoConfiguration.
func (s *Server) doApplyBulkRefConfiguration(ctx context.Context, log *log.Context, repoName string, repo *serverRepo, oldConfig, newConfig RepoConfiguration) error {
	for name := range oldConfig.Refs {
		// Remove refs that were removed from the config.
		if _, ok := newConfig.Refs[name]; !ok {
			if err := s.doApplyRefConfiguration(ctx, log, repo, RefIdentifier{Repo: repoName, Ref: name}, repo.refdb.Lookup(name), oldConfig, newConfig, false, true, true); err != nil {
				return err
			}
		}
	}
	for name, newRefConfig := range newConfig.Refs {
		if oldRefConfig := oldConfig.Refs[name]; oldRefConfig != newRefConfig {
			if err := s.doApplyRefConfiguration(ctx, log, repo, RefIdentifier{Repo: repoName, Ref: name}, repo.refdb.Lookup(name), oldConfig, newConfig, false, true, true); err != nil {
				return err
			}
		}
	}
	return nil
}
