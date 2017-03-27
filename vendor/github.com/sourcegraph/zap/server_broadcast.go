package zap

import (
	"context"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap/internal/debugutil"
)

// broadcastRefUpdate enqueues a ref update (either to a symbolic ref
// or non-symbolic ref) to be sent to all watchers and acked to the
// sender.
//
// Exactly 1 of the nonSymbolic and symbolic parameters must be set.
func (s *Server) broadcastRefUpdate(ctx context.Context, logger log.Logger, sender *serverConn, nonSymbolic *RefUpdateDownstreamParams, symbolic *RefUpdateSymbolicParams) error {
	debugutil.SimulateLatency()

	if (nonSymbolic == nil) == (symbolic == nil) {
		panic("exactly 1 of nonSymbolic and symbolic must be set")
	}

	var refID RefIdentifier
	if nonSymbolic != nil {
		refID = nonSymbolic.RefIdentifier
	} else {
		refID = symbolic.RefIdentifier
	}

	if watchers := s.watchers(refID); len(watchers) > 0 {
		level.Debug(logger).Log("broadcast-ref-update", refID, "watchers", strings.Join(clientIDs(watchers), " "))

		for _, c := range watchers {
			// Set Ack = true if this is being sent to the
			// original sender.
			ack := c == sender

			var item refUpdateItem
			if nonSymbolic != nil {
				item.nonSymbolic = nonSymbolic // copy
				item.nonSymbolic.Ack = ack
			} else {
				item.symbolic = symbolic // copy
				item.symbolic.Ack = ack
			}
			c.send(ctx, logger, item)
		}
	}
	return nil
}
