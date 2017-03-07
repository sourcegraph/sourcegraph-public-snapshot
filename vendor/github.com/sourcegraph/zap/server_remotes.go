package zap

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap/internal/pkg/backoff"
)

// UpstreamClient is the subset of the Client interface that the
// server uses to communicate with the upstream remote server.
type UpstreamClient interface {
	RepoWatch(context.Context, RepoWatchParams) error
	RefInfo(context.Context, RefIdentifier) (*RefInfoResult, error)
	RefUpdate(context.Context, RefUpdateUpstreamParams) error
	SetRefUpdateCallback(func(context.Context, RefUpdateDownstreamParams) error)
	SetRefUpdateSymbolicCallback(f func(context.Context, RefUpdateSymbolicParams) error)
	DisconnectNotify() <-chan struct{}
	Close() error
}

// ConfigureRemoteClientFunc sets the func that this server calls to
// connect to upstream servers.
func (s *Server) ConfigureRemoteClientFunc(newClient func(ctx context.Context, endpoint string) (UpstreamClient, error)) {
	if s.remotes.newClient != nil {
		panic("(serverRemotes).newClient is already set")
	}
	s.remotes.newClient = newClient
}

type serverRemotes struct {
	parent *Server

	mu   sync.Mutex
	conn map[string]UpstreamClient // remote endpoint -> client

	newClient func(ctx context.Context, endpoint string) (UpstreamClient, error)
}

func (sr *serverRemotes) getOrCreateClient(ctx context.Context, logger log.Logger, endpoint string) (UpstreamClient, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if sr.newClient == nil {
		panic("(serverRemotes).newClient must be set with (*Server).ConfigureRemoteClientFunc")
	}
	cl, ok := sr.conn[endpoint]
	if !ok {
		ctx := sr.parent.bgCtx // use background context since this isn't tied to a specific request

		var err error
		debugSimulateLatency()
		cl, err = sr.newClient(ctx, endpoint)
		if err != nil {
			return nil, err
		}
		level.Info(logger).Log("connected-to-remote", endpoint)
		go func() {
			<-cl.DisconnectNotify()
			// If server is closed, do not attempt to connect. This
			// prevents an infinite loop when the server closes and
			// its connections are closed (which usually would cause
			// them all to try to reconnect).
			if sr.parent.isClosed() || ctx.Err() != nil {
				return
			}

			logger := log.With(sr.parent.baseLogger(), "remote-client-monitor", endpoint)
			level.Warn(logger).Log("disconnected", "")
			sr.mu.Lock()
			delete(sr.conn, endpoint)
			sr.mu.Unlock()
			if err := backoff.RetryNotifyWithContext(context.Background(), func(ctx context.Context) error {
				return sr.tryReconnect(ctx, logger, endpoint)
			}, remoteBackOff(), func(err error, d time.Duration) {
				level.Debug(logger).Log("retry-reconnect-after-error", err)
			}); err != nil {
				level.Error(logger).Log("reconnect-failed-after-retries", err)
			}
		}()
		cl.SetRefUpdateCallback(func(ctx context.Context, params RefUpdateDownstreamParams) error {
			debugSimulateLatency()

			// Create a clean logger, because it will be used in
			// requests other than the initial one.
			logger := log.With(sr.parent.baseLogger(), "callback-from-remote-endpoint", endpoint)
			if err := sr.parent.handleRefUpdateFromUpstream(ctx, logger, params, endpoint); err != nil {
				level.Error(logger).Log("params", params, "err", err)
				return err
			}
			return nil
		})
		cl.SetRefUpdateSymbolicCallback(func(context.Context, RefUpdateSymbolicParams) error {
			// Nothing to do here; symbolic refs are not shared between servers.
			return nil
		})
		if sr.conn == nil {
			sr.conn = map[string]UpstreamClient{}
		}
		sr.conn[endpoint] = cl
	}
	return cl, nil
}

func (sr *serverRemotes) closeAndRemoveClient(endpoint string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	cl, ok := sr.conn[endpoint]
	if !ok {
		panic("no remote client for endpoint " + endpoint)
	}
	delete(sr.conn, endpoint)
	return cl.Close()
}

func (sr *serverRemotes) tryReconnect(ctx context.Context, logger log.Logger, endpoint string) error {
	level.Debug(logger).Log("try-reconnect", "")
	cl, err := sr.getOrCreateClient(ctx, logger, endpoint)
	if err != nil {
		return err
	}
	level.Debug(logger).Log("reconnect-ok", "")

	reestablishRepo := func(repoName string, repo *serverRepo) error {
		repoConfig, err := repo.getConfig()
		if err != nil {
			return err
		}
		for remoteName, remote := range repoConfig.Remotes {
			if remote.Endpoint == endpoint {
				level.Debug(logger).Log("reestablish-watch-repo", repoName, "remote", remoteName)
				if err := cl.RepoWatch(ctx, RepoWatchParams{Repo: remote.Repo, Refspecs: remote.Refspecs}); err != nil {
					return err
				}
				for refName, refConfig := range repoConfig.Refs {
					do := func() error {
						// Wrap in func so we can defer here.
						defer repo.acquireRef(refName)()

						ref := repo.refdb.Lookup(refName)
						if refConfig.Overwrite && refConfig.Upstream == remoteName && ref != nil {
							o := ref.Object.(serverRef)
							return cl.RefUpdate(ctx, RefUpdateUpstreamParams{
								RefIdentifier: RefIdentifier{Repo: remote.Repo, Ref: refName},
								Force:         true,
								State: &RefState{
									RefBaseInfo: RefBaseInfo{GitBase: o.gitBase, GitBranch: o.gitBranch},
									History:     o.history(),
								},
							})
						}
						return nil
					}
					if err := do(); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// Briefly hold repo lock; no need to wait for the reestablishment
	// operations to all finish before releasing it.
	sr.parent.reposMu.Lock()
	reposCopy := make(map[string]*serverRepo, len(sr.parent.repos))
	for repoName, repo := range sr.parent.repos {
		reposCopy[repoName] = repo
	}
	sr.parent.reposMu.Unlock()
	for repoName, repo := range reposCopy {
		if err := reestablishRepo(repoName, repo); err != nil {
			return err
		}
	}

	return nil
}

func remoteBackOff() backoff.BackOff {
	p := backoff.NewExponentialBackOff()
	p.InitialInterval = 500 * time.Millisecond
	p.Multiplier = 2
	p.MaxElapsedTime = 1 * time.Minute
	return p
}
