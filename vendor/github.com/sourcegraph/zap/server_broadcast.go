package zap

import (
	"context"
	"sort"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap/internal/debugutil"
	"github.com/sourcegraph/zap/server/refdb"
)

// broadcastRefUpdate enqueues a ref update (either to a symbolic ref
// or non-symbolic ref) to be sent to all watchers and acked to the
// sender.
//
// Exactly 1 of the nonSymbolic and symbolic parameters must be set.
func (s *Server) broadcastRefUpdate(ctx context.Context, logger log.Logger, updatedRefs []refdb.Ref, sender *serverConn, nonSymbolic *RefUpdateDownstreamParams, symbolic *RefUpdateSymbolicParams) error {
	if ctx == nil {
		panic("ctx == nil")
	}

	debugutil.SimulateLatency()

	var repo string
	if nonSymbolic != nil {
		repo = nonSymbolic.RefIdentifier.Repo
	} else {
		repo = symbolic.RefIdentifier.Repo
	}
	makeRefUpdateItem := func(ref string, ack bool) refUpdateItem {
		if nonSymbolic != nil {
			params := *nonSymbolic // copy
			params.RefIdentifier.Ref = ref
			params.Ack = ack
			return refUpdateItem{nonSymbolic: &params}
		}
		params := *symbolic // copy
		params.RefIdentifier.Ref = ref
		params.Ack = ack
		return refUpdateItem{symbolic: &params}
	}

	sort.Sort(sortableRefs(updatedRefs))
	for _, ref := range updatedRefs {
		refID := RefIdentifier{Repo: repo, Ref: ref.Name}
		if watchers := s.watchers(refID); len(watchers) > 0 {
			level.Debug(logger).Log("broadcast-ref-update", refID, "watchers", strings.Join(clientIDs(watchers), " "))

			for _, c := range watchers {
				// Send the update with the ref name that the client
				// is watching as (e.g., "HEAD" not "master" if they
				// are only watching HEAD).
				//
				// Also set Ack = true if this is being sent to the
				// original sender.
				c.send(ctx, logger, makeRefUpdateItem(ref.Name, c == sender))
			}
		}
	}
	return nil
}
