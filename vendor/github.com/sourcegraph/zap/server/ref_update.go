package server

import (
	"context"
	"sort"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/internal/debugutil"
	"github.com/sourcegraph/zap/server/refdb"
	"github.com/sourcegraph/zap/server/refstate"
	"github.com/sourcegraph/zap/server/repodb"
)

// RefUpdate processes a ref update.
//
// It accepts a ref update (either upstream, downstream, or internal),
// recording and applying it to the ref's state. It then calls the
// AfterRefUpdate hook for all server extensions. Finally, it
// broadcasts the update to downstream watchers.
func (s *Server) RefUpdate(ctx context.Context, logger log.Logger, sender *Conn, repo *repodb.OwnedRepo, ref *refdb.OwnedRef, params refstate.RefUpdate) error {
	// Record the update in the ref's state.
	if err := refstate.Update(logger, repo, ref, &params); err != nil {
		return err
	}

	// Call server extensions' AfterRefUpdate hooks.
	for _, ext := range s.ext {
		if ext.AfterRefUpdate != nil {
			if err := ext.AfterRefUpdate(ctx, logger, *repo, *ref, params); err != nil {
				return err
			}
		}
	}

	// Broadcast to downstream watchers.
	//
	// Don't re-broadcast acks; the ack is only meaningful to the
	// immediate sender (this server), not to downstreams of the
	// sender.
	if !params.Ack {
		if err := s.broadcastRefUpdate(ctx, logger, sender, params.ToUpdateDownstream()); err != nil {
			return err
		}
	}

	return nil
}

// broadcastRefUpdate enqueues a ref update to be sent to all watchers
// and acked to the sender.
func (s *Server) broadcastRefUpdate(ctx context.Context, logger log.Logger, sender *Conn, params zap.RefUpdateDownstreamParams) error {
	debugutil.SimulateLatency()

	if watchers := s.watchers(params.RefIdentifier); len(watchers) > 0 {
		level.Debug(logger).Log("broadcast-ref-update", params.RefIdentifier, "watchers", strings.Join(clientIDs(watchers), " "))

		for _, c := range watchers {
			// Set Ack = true if this is being sent to the
			// original sender.
			params.Ack = c == sender
			c.send(ctx, logger, params)
		}
	}
	return nil
}

func clientIDs(conns []*Conn) (ids []string) {
	ids = make([]string, len(conns))
	for i, c := range conns {
		c.mu.Lock()
		if c.init != nil {
			ids[i] = c.init.ID
		}
		c.mu.Unlock()
	}
	sort.Strings(ids)
	return ids
}

type sortableRefs []refdb.Ref

func (v sortableRefs) Len() int           { return len(v) }
func (v sortableRefs) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v sortableRefs) Less(i, j int) bool { return v[i].Name < v[j].Name }
