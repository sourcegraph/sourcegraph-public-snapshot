package server

import (
	"context"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/internal/debugutil"
	"github.com/sourcegraph/zap/server/refstate"
)

// UpstreamClient is the subset of the Client interface that the
// server uses to communicate with the upstream remote server.
type UpstreamClient interface {
	RepoWatch(context.Context, zap.RepoWatchParams) error
	RefInfo(context.Context, zap.RefInfoParams) (*zap.RefInfo, error)
	RefUpdate(context.Context, zap.RefUpdateUpstreamParams) error
	SetRefUpdateCallback(func(context.Context, zap.RefUpdateDownstreamParams))
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

// remoteConn holds the client and connection state associated with a
// connection to a remote upstream server.
type remoteConn struct {
	UpstreamClient // the client

	// refUpdates contains enqueued ref updates that are sent to the
	// remote server using (UpstreamClient).RefUpdate. The refstate package sends updates on this channel.
	//
	// Callers (such as the refstate package) that don't/can't handle
	// errors and only care about ensuring ordering can use this
	// channel instead of calling (UpstreamClient).RefUpdate directly.
	refUpdates chan<- zap.RefUpdateUpstreamParams
}

type serverRemotes struct {
	parent *Server

	mu   sync.Mutex
	conn map[string]remoteConn // remote endpoint -> remote conn info

	newClient func(ctx context.Context, endpoint string) (UpstreamClient, error)
}

func (sr *serverRemotes) getOrCreateClient(logger log.Logger, endpoint string) (*remoteConn, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if sr.newClient == nil {
		panic("(serverRemotes).newClient must be set with (*Server).ConfigureRemoteClientFunc")
	}
	rc, ok := sr.conn[endpoint]
	if !ok {
		ctx := sr.parent.Background // use background context since this isn't tied to a specific request

		var err error
		debugutil.SimulateLatency()
		cl, err := sr.newClient(ctx, endpoint)
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
			logger := sr.parent.BaseLogger()
			level.Error(logger).Log("disconnected-from-remote", endpoint)
			sr.mu.Lock()
			delete(sr.conn, endpoint)
			sr.mu.Unlock()
		}()

		// TODO(nick): Figure out error handling for when we send a
		// bad update to the server or when we receive a bad update
		// from the server.

		// Set callback for receiving updates from the upstream.
		cl.SetRefUpdateCallback(func(ctx context.Context, params zap.RefUpdateDownstreamParams) {
			debugutil.SimulateLatency()

			// Create a clean logger, because it will be used in
			// requests other than the initial one.
			logger := log.With(sr.parent.BaseLogger(), "from-upstream", endpoint, "ref", params.RefIdentifier, "params", params)

			if err := params.Validate(); err != nil {
				level.Error(logger).Log("invalid-params", err)
				return
			}

			repo, localRepoName, _, err := sr.parent.findLocalRepo(ctx, logger, params.RefIdentifier.Repo, endpoint)
			if err != nil {
				level.Error(logger).Log("findLocalRepo", err)
				return
			}
			defer repo.Unlock()
			if repo.Repo == nil {
				level.Error(logger).Log("no-local-repo", "")
				return
			}

			if repo.WorkspaceRef != "" && params.RefIdentifier.Ref == repo.WorkspaceRef && !params.Ack && (params.State != nil || params.Delete) {
				// Ignore upstream resets/deletes to our head ref,
				// because that is exclusively controlled locally.
				return
			}

			params.RefIdentifier.Repo = localRepoName
			logger = log.With(logger, "local-ref", params.RefIdentifier, "update", params)

			ref := repo.RefDB.Lookup(params.RefIdentifier.Ref)
			defer ref.Unlock()

			// Receive the update and apply to our internal ref state.
			var u refstate.RefUpdate
			u.FromUpdateDownstream(params)
			if err := sr.parent.RefUpdate(ctx, logger, nil, repo, &ref, u); err != nil {
				level.Error(logger).Log("update-error", err)
			}
		})

		// Handle refUpdates channel.
		refUpdates := make(chan zap.RefUpdateUpstreamParams) // it is safe to increase the channel buffer size
		go func() {
			ctx := sr.parent.Background
			logger := log.With(sr.parent.BaseLogger(), "send-upstream", endpoint)
			for {
				select {
				case <-cl.DisconnectNotify():
				// TODO(nick): how to quit?
				case params, ok := <-refUpdates:
					if !ok {
						return
					}
					logger := log.With(logger, "local-ref", params.RefIdentifier)
					if err := cl.RefUpdate(ctx, params); err != nil {
						level.Error(logger).Log("upstream-ref-update-error", err, "update", params)
					}
				}
			}
		}()

		rc = remoteConn{
			UpstreamClient: cl,
			refUpdates:     refUpdates,
		}
		if sr.conn == nil {
			sr.conn = map[string]remoteConn{}
		}
		sr.conn[endpoint] = rc
	}
	return &rc, nil
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
