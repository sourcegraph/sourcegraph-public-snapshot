package zap

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/sourcegraph/jsonrpc2"
)

func (s *Server) doUpdateRepoConfiguration(ctx context.Context, log *log.Context, repoName string, repo *serverRepo, config RepoConfiguration) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if err := s.doUpdateBulkRepoRemoteConfiguration(ctx, log, repoName, repo, repo.config.Remotes, config.Remotes); err != nil {
		return err
	}
	if err := s.doUpdateBulkRefConfiguration(ctx, log, repoName, repo, repo.config.Refs, config.Refs); err != nil {
		return err
	}
	return nil
}

// doUpdateRepoRemoteConfiguration assumes the caller holds repo.mu.
func (s *Server) doUpdateBulkRepoRemoteConfiguration(ctx context.Context, log *log.Context, repoName string, repo *serverRepo, oldRemotes, newRemotes map[string]RepoRemoteConfiguration) error {
	// Forbid multiple remotes having the same endpoint.
	seenEndpoints := make(map[string][]string, len(oldRemotes))
	for name, config := range newRemotes {
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
	for oldName, oldRemote := range oldRemotes {
		if newRemote, ok := newRemotes[oldName]; ok && newRemote.Endpoint == oldRemote.Endpoint {
			continue // connection remains the same
		}

		// TODO(sqs): this needs to check whether this client is being
		// used by any other repos. It should use reference counting
		// and only remove/close when it gets to 0.
		level.Info(log).Log("rm-remote", oldName)
		if err := s.remotes.closeAndRemoveClient(oldRemote.Endpoint); err != nil {
			return err
		}
		delete(repo.config.Remotes, oldName)
	}
	for newName, newRemote := range newRemotes {
		oldRemote, ok := oldRemotes[newName]
		if ok && oldRemote == newRemote {
			continue // unchanged
		}
		log := log.With("add-or-update-remote", newName)
		level.Debug(log).Log("new", newRemote, "old", oldRemotes[newName])
		cl, err := s.remotes.getOrCreateClient(ctx, log, newRemote.Endpoint)
		if err != nil {
			return err
		}

		// TODO(sqs): does not correctly clean up repo watches
		// established on previous endpoints or repo names. Kind of an edge case.
		if oldRemote.Refspec != "" || newRemote.Refspec != "" {
			if err := cl.RepoWatch(ctx, RepoWatchParams{Repo: newRemote.Repo, Refspec: newRemote.Refspec}); err != nil {
				return err
			}
		}

		if repo.config.Remotes == nil {
			repo.config.Remotes = map[string]RepoRemoteConfiguration{}
		}
		repo.config.Remotes[newName] = newRemote

		// Apply existing ref configuration against new upstream, if
		// the endpoint or repo changed. If just the refspec changed,
		// we don't need to do anything as that would not change the
		// desired state on the upstream.
		if oldRemote.Endpoint != newRemote.Endpoint || oldRemote.Repo != newRemote.Repo {
			for ref, refConfig := range repo.config.Refs {
				if refConfig.Upstream == newName {
					if err := s.doUpdateRefConfiguration(ctx, log, repo, RefIdentifier{Repo: repoName, Ref: ref}, repo.refdb.Lookup(ref), RefConfiguration{}, refConfig, true /* force */, true); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
