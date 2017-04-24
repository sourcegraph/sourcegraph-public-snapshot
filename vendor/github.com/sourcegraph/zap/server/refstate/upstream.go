package refstate

import (
	"errors"

	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/ot"
)

// Upstream is the ephemeral state associated with a connection to a
// remote upstream server for the ref. It is only valid for the
// lifetime of the connection.
type Upstream struct {
	Wait *zap.RefUpdateUpstreamParams // update pending upstream acknowledgment
	Buf  *zap.RefUpdateUpstreamParams // buffered update to send upstream when Wait is acked
	Rev  uint                         // upstream revision number of last upstream-acked revision

	// Send enqueues updates to send to this upstream server.
	Send chan<- zap.RefUpdateUpstreamParams

	// RemoteRepo is the path of the repo on the upstream (or "" if it
	// has the same path as the local repo). The RefIdentifier.Repo
	// fields of all values sent via Send must be updated to equal
	// RemoteRepo (using the withRemoteRepo method).
	RemoteRepo string
}

func (u *Upstream) withRemoteRepo(params zap.RefUpdateUpstreamParams) zap.RefUpdateUpstreamParams {
	if u.RemoteRepo == "" {
		panic("empty RemoteRepo")
	}
	params.RefIdentifier.Repo = u.RemoteRepo
	return params
}

// sendUpdate is called when a full ref update should be sent
// upstream. It clobbers any existing buffered update but it still
// waits for the pending update to be acked.
func (u *Upstream) sendUpdate(params zap.RefUpdateUpstreamParams) error {
	if params.Op != nil {
		params.Rev = 0 // We determine the rev when we send it, so zero it out for now.
	}
	switch {
	case u.Buf != nil:
		switch {
		case params.State != nil:
			u.Buf = &params
		case u.Buf.State != nil:
			u.Buf.State.Data.History = append(u.Buf.State.Data.History, *params.Op)
		case u.Buf.Op != nil:
			var err error
			buf, err := ot.ComposeWorkspaceOps(*u.Buf.Op, *params.Op)
			if err != nil {
				return err
			}
			u.Buf.Op = &buf
		default:
			panic("unable to append/compose op to RefUpdateUpstreamParams with State == nil and Op == nil")
		}

	case u.Wait != nil:
		u.Buf = &params

	default:
		u.Wait = &params
		if u.Wait.Op != nil {
			u.Wait.Rev = u.Rev
		}
		u.Send <- u.withRemoteRepo(u.Wait.DeepCopy())
	}
	return nil
}

// recvFromUpstream receives ops from the upstream. The caller is
// responsible for sending the returned op to all downstreams.
//
// The params value is modified to transform them against the pending
// and buffered ops so that params is consecutive with our current
// history (if it is a single op update).
func (u *Upstream) recvFromUpstream(params *RefUpdate) error {
	if u.Wait == nil && u.Buf != nil {
		panic("u.Wait == nil && u.Buf != nil")
	}

	switch {
	case params.State != nil:
		if params.State.Data == nil && params.State.Target == "" {
			return errors.New("invalid RefState received from upstream (contains neither Data nor Target)")
		}

		// The upstream's reset will win over anything we have in wait
		// or buf.
		u.Wait = nil
		u.Buf = nil
		if params.State.IsSymbolic() {
			u.Rev = 0
		} else {
			u.Rev = uint(len(params.State.Data.History))
		}

	case params.Op != nil:
		if (u.Wait != nil && u.Wait.State != nil) || (u.Buf != nil && u.Buf.State != nil) {
			// Our local wait or buf reset will win over the
			// upstream's op, so we want to make the upstream's update
			// a noop.
			params.Op = nil
			return nil
		}

		// Otherwise transform the server's op.
		tmp := params.Op.DeepCopy()
		params.Op = &tmp
		if u.Wait != nil && u.Wait.Op != nil {
			var err error
			if *u.Wait.Op, *params.Op, err = ot.TransformWorkspaceOps(*u.Wait.Op, *params.Op); err != nil {
				return err
			}
		}
		if u.Buf != nil && u.Buf.Op != nil {
			var err error
			if *u.Buf.Op, *params.Op, err = ot.TransformWorkspaceOps(*u.Buf.Op, *params.Op); err != nil {
				return err
			}
		}
		u.Rev++
	}
	return nil
}

// ackFromUpstream acknowledges a pending upstream update and sends the
// buffered update (if any) to the upstream.
func (u *Upstream) ackFromUpstream() error {
	if u.Wait == nil {
		return ErrNoPendingOperation
	}

	// Update our revision number to match what the upstream has acked.
	switch {
	case u.Wait.State != nil:
		if u.Wait.State.IsSymbolic() {
			u.Rev = 0
		} else {
			u.Rev = uint(len(u.Wait.State.Data.History))
		}
	case u.Wait.Op != nil:
		u.Rev++
	default:
		panic("unsure how revision should be incremented for waiting op")
	}

	switch {
	case u.Buf != nil:
		// This means there are updates buffered to send upstream (AND
		// updates pending upstream ack).
		u.Wait = u.Buf
		if u.Wait.Op != nil {
			u.Wait.Rev = u.Rev
		}
		u.Buf = nil
		u.Send <- u.withRemoteRepo(u.Wait.DeepCopy())

	case u.Wait != nil:
		// Now the upstream is up-to-date with us.
		u.Wait = nil
	}
	return nil
}

// ErrNoPendingOperation occurs when an ack is received from the
// upstream but there is no pending operation that was waiting to be
// acked.
var ErrNoPendingOperation = errors.New("no pending operation")
