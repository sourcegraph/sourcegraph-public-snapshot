package ws

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/sourcegraph/zap/ot"
)

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
	SendToUpstream func(upstreamRev int, op ot.WorkspaceOp)

	// Apply immediately applies op to the workspace. It is assumed
	// that op will apply cleanly and has been properly transformed
	// with concurrent ops.
	Apply func(log *log.Context, op ot.WorkspaceOp) error

	history           []ot.WorkspaceOp // all ops
	Wait              *ot.WorkspaceOp  // pending upstream acknowledgment
	Buf               *ot.WorkspaceOp  // buffered to send upstream when Wait ops are acked
	UpstreamRevNumber int              // upstream revision number of last upstream-acknowledged revision

	mu sync.Mutex
}

// Rev returns the current revision number for downstream clients.
func (p *Proxy) Rev() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.history)
}

// History returns all acknowledged ops.
func (p *Proxy) History() []ot.WorkspaceOp {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.history
}

// Record records a change that has already been applied. It adds op
// to the history and sends it (or buffers it to be sent)
// upstream. The caller is responsible for broadcasting it to all
// downstream clients.
func (p *Proxy) Record(op ot.WorkspaceOp) error {
	// if op.Noop() {
	// 	panic("noop")
	// }
	p.mu.Lock()
	defer p.mu.Unlock()

	p.history = append(p.history, op)
	if err := p.bufferedSendToUpstream(op); err != nil {
		return err
	}
	return nil
}

// RecvFromDownstream receives ops from downstream clients. The caller
// is responsible for acking op to its sender and sending op to all
// other downstream clients.
func (p *Proxy) RecvFromDownstream(log *log.Context, rev int, op ot.WorkspaceOp) (ot.WorkspaceOp, error) {
	// if op.Noop() {
	// 	panic("noop")
	// }
	p.mu.Lock()
	defer p.mu.Unlock()

	log = log.With("recv-from-downstream", fmt.Sprintf("@%d", rev))

	// Transform it so it can be appended to our view of the history.
	if rev < 0 || len(p.history) < rev {
		return ot.WorkspaceOp{}, fmt.Errorf("revision %d not in history", rev)
	}

	level.Debug(log).Log("op", op, "transform-against-history", fmt.Sprint(p.history[rev:]))

	var err error
	for _, other := range p.history[rev:] {
		if op, _, err = ot.TransformWorkspaceOps(op, other); err != nil {
			return ot.WorkspaceOp{}, err
		}
	}
	if len(p.history[rev:]) > 0 {
		level.Debug(log).Log("transformed-op", op)
	}

	if p.Apply != nil {
		if err := p.Apply(log, op); err != nil {
			return ot.WorkspaceOp{}, err
		}
	}
	if err := p.bufferedSendToUpstream(op); err != nil {
		return ot.WorkspaceOp{}, err
	}

	p.history = append(p.history, op)

	return op, nil
}

// bufferedSendToUpstream is called when an op should be sent
// upstream. It sends op to the upstream if there is no pending op;
// otherwise it buffers it and will send it after the server acks the
// pending op.
func (p *Proxy) bufferedSendToUpstream(op ot.WorkspaceOp) error {
	if p.SendToUpstream == nil {
		return nil
	}

	switch {
	case p.Buf != nil:
		var err error
		buf, err := ot.ComposeWorkspaceOps(*p.Buf, op)
		if err != nil {
			return err
		}
		p.Buf = &buf
	case p.Wait != nil:
		p.Buf = &op
	default:
		p.Wait = &op
		p.SendToUpstream(p.UpstreamRevNumber, op)
	}
	return nil
}

// AckFromUpstream acknowledges a pending upstream op and sends the
// buffered op (if any) to the upstream.
func (p *Proxy) AckFromUpstream() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch {
	case p.Buf != nil: // ops buffered to send to upstream (AND ops pending upstream ack)
		p.SendToUpstream(p.UpstreamRevNumber+1, *p.Buf)
		// Now the upstream is up-to-date with our history at the time
		// we sent this ack'd op to the server.
		p.Wait = p.Buf
		p.Buf = nil
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
func (p *Proxy) RecvFromUpstream(log *log.Context, op ot.WorkspaceOp) (ot.WorkspaceOp, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	log = log.With("recv-from-upstream", fmt.Sprintf("@%d(upstream)", p.UpstreamRevNumber))
	level.Debug(log).Log("op", op, "wait", p.Wait, "buf", p.Buf, "history-length", len(p.history), "history", fmt.Sprint(p.history))

	// Transform it so it can be appended to our view of the history.
	var err error
	if p.Wait != nil {
		var wait ot.WorkspaceOp
		if wait, op, err = ot.TransformWorkspaceOps(*p.Wait, op); err != nil {
			return ot.WorkspaceOp{}, err
		}
		p.Wait = &wait
	}
	if p.Buf != nil {
		var buf ot.WorkspaceOp
		if buf, op, err = ot.TransformWorkspaceOps(*p.Buf, op); err != nil {
			return ot.WorkspaceOp{}, err
		}
		p.Buf = &buf
	}

	if p.Wait != nil || p.Buf != nil {
		level.Debug(log).Log("transformed-op", op, "wait", p.Wait, "buf", p.Buf)
	}

	// TODO(sqs): this is bad because if Apply fails, p.Wait and p.Buf
	// have already been overwritten.
	if p.Apply != nil {
		if err := p.Apply(log, op); err != nil {
			return ot.WorkspaceOp{}, err
		}
	}

	p.history = append(p.history, op)
	p.UpstreamRevNumber++

	return op, nil
}
