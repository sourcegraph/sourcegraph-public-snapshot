package ws

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/zap/ot"
)

type Server struct {
	Apply func(ot.WorkspaceOp) error

	mu      sync.Mutex
	history []ot.WorkspaceOp
}

// Rev reports the server's revision number.
func (s *Server) Rev() int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.history)
}

// History returns a copy of the server's history.
func (s *Server) History() []ot.WorkspaceOp {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.history
}

// Recv transforms, applies, and returns a client op and its revision.
// An error is returned if the op could not be applied. Sending the
// derived op to connected clients and acking the sender are the
// caller's responsibility.
func (s *Server) Recv(rev int, op ot.WorkspaceOp) (ot.WorkspaceOp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rev < 0 || len(s.history) < rev {
		return ot.WorkspaceOp{}, fmt.Errorf("revision %d not in history", rev)

	}

	// Transform ops against all operations that happened since rev.
	for _, sop := range s.history[rev:] {
		var err error
		op, _, err = ot.TransformWorkspaceOps(op, sop)
		if err != nil {
			return ot.WorkspaceOp{}, err
		}
	}
	if err := s.Apply(op); err != nil {
		return ot.WorkspaceOp{}, err
	}
	s.history = append(s.history, op)
	return op, nil
}
