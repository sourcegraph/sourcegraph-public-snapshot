package svc

import (
	"context"
	"sync"
)

// Pending manages a map of all pending requests to a rpc.Service for a
// connection (an rpc.ServerCodec).
type Pending struct {
	mu sync.Mutex
	m  map[uint64]context.CancelFunc // seq -> cancel
}

func NewPending() *Pending {
	return &Pending{m: make(map[uint64]context.CancelFunc)}
}

func (s *Pending) Start(seq uint64) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	// we assume seq is not already in map. If not, the client is broken.
	s.m[seq] = cancel
	s.mu.Unlock()
	return ctx
}

func (s *Pending) Cancel(seq uint64) {
	s.mu.Lock()
	cancel, ok := s.m[seq]
	if ok {
		delete(s.m, seq)
	}
	s.mu.Unlock()
	if ok {
		cancel()
	}
}

type CancelArgs struct {
	// Seq is the sequence number for the rpc.Call to cancel.
	Seq uint64

	// pending is the DS used by rpc.Server to track the ongoing calls for
	// this connection. It should not be set by the client, the Service will
	// set it.
	pending *Pending
}

// SetPending sets the pending map for the server to use. Do not use on the
// client.
func (a *CancelArgs) SetPending(p *Pending) {
	a.pending = p
}

// GoRPC is an internal service used by rpc.
type GoRPC struct{}

func (s *GoRPC) Cancel(ctx context.Context, args *CancelArgs, reply *bool) error {
	args.pending.Cancel(args.Seq)
	return nil
}
