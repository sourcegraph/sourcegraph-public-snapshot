package backend

import (
	"encoding/json"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

// Channel implements a live communication channel between a client
// and the server.
//
// It relies on the Channel store.
//
// A sender can check whether the channel they just sent to has a
// listener by using the CheckForListeners option in
// Channel.Send. This is implemented by having Listen acknowledge each
// receipt of a message by (in turn) sending an ack message on another
// channel that Channel.Send listens on. This is essentially role
// reversal for the purpose of acking.
var Channel sourcegraph.ChannelServer = &channel{}

type channel struct{}

var _ sourcegraph.ChannelServer = (*channel)(nil)

var listenersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "channel",
	Name:      "listeners",
	Help:      "Number of currently active channel listeners (Channel.Listen).",
})

func init() {
	prometheus.MustRegister(listenersGauge)
}

const ackPayload = "ACK"

func (s *channel) Listen(op *sourcegraph.ChannelListenOp, stream sourcegraph.Channel_ListenServer) error {
	ctx := stream.Context()
	l, unlisten, err := store.ChannelFromContext(ctx).Listen(ctx, op.Channel)
	if err != nil {
		return err
	}
	defer unlisten()
	listenersGauge.Inc()
	defer listenersGauge.Dec()

	for {
		const timeout = 8 * time.Hour
		select {
		case <-time.After(timeout):
			return grpc.Errorf(codes.DeadlineExceeded, "channel listen idle timed out after %s to prevent resource exhaustion; can restart listening immediately", timeout)

		case <-stream.Context().Done():
			// The client closed the stream.
			return nil

		case n, ok := <-l:
			if !ok {
				// Channel could be closed due to the PostgreSQL
				// connection being lost or killed.
				return grpc.Errorf(codes.Unavailable, "backend channel closed (possibly due to transient error); retry after a short delay")
			}

			if n.Payload == ackPayload {
				// Don't pass along ack payloads to the caller; they are
				// an internal implementation detail.
				continue
			}

			var action *sourcegraph.ChannelAction
			if err := json.Unmarshal([]byte(n.Payload), &action); err != nil {
				return err
			}

			if err := stream.Send(action); err != nil {
				return err
			}

			// Acknowledge (might be ignored) with an ack payload
			// notification.
			if err := store.ChannelFromContext(ctx).Notify(ctx, op.Channel, ackPayload); err != nil {
				return err
			}
		}
	}
}

func (s *channel) Send(ctx context.Context, op *sourcegraph.ChannelSendOp) (*sourcegraph.ChannelSendResult, error) {
	payload, err := json.Marshal(op.Action)
	if err != nil {
		return nil, err
	}

	// Support waiting for listener acknowledgement.
	var wait func(ctx context.Context) error
	if op.CheckForListeners {
		l, unlisten, err := store.ChannelFromContext(ctx).Listen(ctx, op.Channel)
		if err != nil {
			return nil, err
		}
		defer unlisten()
		wait = func(ctx context.Context) error {
			// Wait until we get an ack payload.
			for {
				select {
				case n, ok := <-l:
					if !ok {
						return grpc.Errorf(codes.Aborted, "channel %q closed while waiting for listeners", op.Channel)
					}
					if n.Payload == ackPayload {
						// Got the ack payload!
						return nil
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}

	if err := store.ChannelFromContext(ctx).Notify(ctx, op.Channel, string(payload)); err != nil {
		return nil, err
	}

	if wait != nil {
		ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
		defer cancel()
		if err := wait(ctx); err != nil {
			return nil, err
		}
	}

	return &sourcegraph.ChannelSendResult{}, nil
}
