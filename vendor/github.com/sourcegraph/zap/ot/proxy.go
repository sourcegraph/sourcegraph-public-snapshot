package ot

import (
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const extraDebug = true

// Proxy sits between an upstream server and any number of downstream
// clients. It proxies the workspace state. To its upstream server,
// the proxy is an OT client. To its downstream clients, the proxy is
// an OT server.
//
// It exists so that downstream clients (such as an editor extension)
// can be built without needing full OT implementation (specifically
// composition and transformation of workspace ops), which makes the
// job of coding it much easier. When run on the same host as a
// downstream client, the latency is low enough that the downstream
// client can assume that its receives and sends are immediate, which
// means it never needs to compose or transform operations.
type Proxy struct {
	// SendToUpstream sends an operation to the upstream.
	SendToUpstream func(logger log.Logger, upstreamRev int, op WorkspaceOp)

	// Apply immediately applies op to the workspace. It is assumed
	// that op will apply cleanly and has been properly transformed
	// with concurrent ops.
	Apply func(logger log.Logger, op WorkspaceOp) error

	History           []WorkspaceOp // all ops
	Wait              *WorkspaceOp  // pending upstream acknowledgment
	Buf               *WorkspaceOp  // buffered to send upstream when Wait ops are acked
	UpstreamRevNumber int           // upstream revision number of last upstream-acknowledged revision
}

// Rev returns the current revision number for downstream clients.
func (p *Proxy) Rev() int {
	return len(p.History)
}

// Record records a change that has already been applied. It adds op
// to the history and sends it (or buffers it to be sent)
// upstream. The caller is responsible for broadcasting it to all
// downstream clients.
func (p *Proxy) Record(logger log.Logger, op WorkspaceOp) error {
	// if op.Noop() {
	// 	panic("noop")
	// }

	p.History = append(p.History, op)
	if err := p.bufferedSendToUpstream(logger, op); err != nil {
		return err
	}
	return nil
}

// RecvFromDownstream receives ops from downstream clients. The caller
// is responsible for acking op to its sender and sending op to all
// other downstream clients.
func (p *Proxy) RecvFromDownstream(logger log.Logger, rev int, op WorkspaceOp) (WorkspaceOp, error) {
	// if op.Noop() {
	// 	panic("noop")
	// }
	logger = log.With(logger, "recv-from-downstream", fmt.Sprintf("@%d", rev))

	// Transform it so it can be appended to our view of the history.
	if rev < 0 || len(p.History) < rev {
		return WorkspaceOp{}, fmt.Errorf("revision %d not in history", rev)
	}

	if extraDebug {
		level.Debug(logger).Log("op", op, "transform-against-history", fmt.Sprint(p.History[rev:]))
	}

	var err error
	for _, other := range p.History[rev:] {
		if op, _, err = TransformWorkspaceOps(op, other); err != nil {
			return WorkspaceOp{}, err
		}
	}
	if len(p.History[rev:]) > 0 {
		if extraDebug {
			level.Debug(logger).Log("transformed-op", op)
		}
	}

	if p.Apply != nil {
		if err := p.Apply(logger, op); err != nil {
			return WorkspaceOp{}, err
		}
	}
	p.History = append(p.History, op)
	if err := p.bufferedSendToUpstream(logger, op); err != nil {
		return WorkspaceOp{}, err
	}

	return op, nil
}

// bufferedSendToUpstream is called when an op should be sent
// upstream. It sends op to the upstream if there is no pending op;
// otherwise it buffers it and will send it after the server acks the
// pending op.
func (p *Proxy) bufferedSendToUpstream(logger log.Logger, op WorkspaceOp) error {
	if p.SendToUpstream == nil {
		return nil
	}

	switch {
	case p.Buf != nil:
		var err error
		buf, err := ComposeWorkspaceOps(*p.Buf, op)
		if err != nil {
			return err
		}
		p.Buf = &buf
	case p.Wait != nil:
		p.Buf = &op
	default:
		p.Wait = &op
		p.SendToUpstream(logger, p.UpstreamRevNumber, op)
	}
	return nil
}

// AckFromUpstream acknowledges a pending upstream op and sends the
// buffered op (if any) to the upstream.
func (p *Proxy) AckFromUpstream(logger log.Logger) error {
	switch {
	case p.Buf != nil: // ops buffered to send to upstream (AND ops pending upstream ack)
		// Now the upstream is up-to-date with our history at the time
		// we sent this ack'd op to the server.
		p.Wait = p.Buf
		p.Buf = nil
		p.SendToUpstream(logger, p.UpstreamRevNumber+1, *p.Wait)
	case p.Wait != nil: // ops pending upstream ack (NO buffered ops)
		// Now the upstream is up-to-date with us.
		p.Wait = nil
	default:
		return ErrNoPendingOperation
	}
	p.UpstreamRevNumber++
	return nil
}

var ErrNoPendingOperation = errors.New("no pending operation")

// RecvFromUpstream receives ops from the upstream. The caller is
// responsible for sending the returned op to all downstreams.
func (p *Proxy) RecvFromUpstream(logger log.Logger, op WorkspaceOp) (WorkspaceOp, error) {
	if p.UpstreamRevNumber > len(p.History) {
		level.Error(logger).Log("PANIC-BELOW", "")
		panic(fmt.Sprintf("invalid p.UpstreamRevNumber > len(p.History) (%d > %d)", p.UpstreamRevNumber, len(p.History)))
	}

	logger = log.With(logger, "recv-from-upstream", fmt.Sprintf("@%d(upstream)", p.UpstreamRevNumber))
	if extraDebug {
		level.Debug(logger).Log("op", op, "wait", p.Wait, "buf", p.Buf, "history-length", len(p.History), "history", fmt.Sprint(p.History))
	}

	// Transform it so it can be appended to our view of the history.
	var err error
	if p.Wait != nil {
		var wait WorkspaceOp
		if wait, op, err = TransformWorkspaceOps(*p.Wait, op); err != nil {
			return WorkspaceOp{}, err
		}
		p.Wait = &wait
	}
	if p.Buf != nil {
		var buf WorkspaceOp
		if buf, op, err = TransformWorkspaceOps(*p.Buf, op); err != nil {
			return WorkspaceOp{}, err
		}
		p.Buf = &buf
	}

	if p.Wait != nil || p.Buf != nil {
		if extraDebug {
			level.Debug(logger).Log("transformed-op", op, "wait", p.Wait, "buf", p.Buf)
		}
	}

	{
		// DEBUG: Sanity check
		allHistory, _ := ComposeAllWorkspaceOps(p.History)
		if _, err := ComposeWorkspaceOps(allHistory, op); err != nil {
			// See if op is consecutive with the last rev.
			if len(p.History) > 1 {
				allHistoryButLast, _ := ComposeAllWorkspaceOps(p.History[:len(p.History)-1])
				if _, err := ComposeWorkspaceOps(allHistoryButLast, op); err == nil {
					level.Error(logger).Log("op-is-not-consecutive-with-history-but-is-with-last-rev", "", "op", op, "rev", p.Rev(), "last-rev", p.Rev()-1, "history", fmt.Sprint(p.History), "composed-history", allHistory, "composed-history-but-last", allHistoryButLast, "err", err)
					return WorkspaceOp{}, fmt.Errorf("op is consecutive with rev %d not current rev %d: %s", p.Rev()-1, p.Rev(), err)
				}
			}

			level.Error(logger).Log("op-is-not-consecutive-with-history", "", "op", op, "history", fmt.Sprint(p.History), "composed-history", allHistory, "err", err)
			return WorkspaceOp{}, err
		}
	}

	// TODO(sqs): this is bad because if Apply fails, p.Wait and p.Buf
	// have already been overwritten.
	if p.Apply != nil {
		if err := p.Apply(logger, op); err != nil {
			return WorkspaceOp{}, err
		}
	}

	p.History = append(p.History, op)
	p.UpstreamRevNumber++

	return op, nil
}
