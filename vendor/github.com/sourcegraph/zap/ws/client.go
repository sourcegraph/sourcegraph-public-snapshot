package ws

import (
	"errors"
	"fmt"
	"os"

	"github.com/sourcegraph/zap/ot"
)

// Client communicates with the server about the state of the
// workspace. It sends the server operations that were applied
// locally, and it applies operations that it receives from the server
// (which originate from other clients connected to the server).
//
// The local client workspace has three states:
//
// 1. A synchronized workspace sends applied ops immediately and...
//
// 2. waits for an acknowledgement from the remote workspace,
//    meanwhile buffering applied ops.
//
// 3. The buffer is composed with new ops and sent immediately when
//    the pending ack arrives.
type Client struct {
	// Send is called by the client to send a locally applied
	// operation to the server, to be distributed to other clients.
	Send func(rev int, op ot.WorkspaceOp)

	// Apply immediately applies op to the workspace. It is assumed
	// that op will apply cleanly and has been properly transformed
	// with concurrent ops.
	Apply func(op ot.WorkspaceOp) error

	Rev int // last server-acknowledged revision (sequential ID)

	Wait *ot.WorkspaceOp // pending ops
	Buf  *ot.WorkspaceOp // buffered ops
}

// Record records an operation that was already applied locally and
// buffers or sends it to the server.
func (c *Client) Record(op ot.WorkspaceOp) error {
	// Use this to make sure we aren't needlessly recording noops,
	// which pollutes logs and makes debugging a bit harder.
	//
	// TODO(sqs): can remove in production
	if op.Noop() {
		panic(fmt.Sprintf("Record op is noop: %s", op))
	}

	switch {
	case c.Buf != nil:
		buf, err := ot.ComposeWorkspaceOps(*c.Buf, op)
		if err != nil {
			return err
		}
		c.Buf = &buf
	case c.Wait != nil:
		c.Buf = &op
	default:
		c.Wait = &op
		c.Send(c.Rev, op)
	}
	return nil
}

// Ack acknowledges a pending server operation and sends buffered
// updates (if any).
func (c *Client) Ack() error {
	switch {
	case c.Buf != nil:
		c.Send(c.Rev+1, *c.Buf)
		c.Wait = c.Buf
		c.Buf = nil
	case c.Wait != nil:
		c.Wait = nil
	default:
		return errors.New("no pending operation")
	}
	c.Rev++
	return nil
}

// Recv receives operations from the server originating from other
// clients. It applies them locally.
func (c *Client) Recv(op ot.WorkspaceOp) error {
	origStr := op.String()

	var err error
	if c.Wait != nil {
		var wait ot.WorkspaceOp
		wait, op, err = ot.TransformWorkspaceOps(*c.Wait, op)
		if err != nil {
			return err
		}
		c.Wait = &wait
	}
	if c.Buf != nil {
		var buf ot.WorkspaceOp
		buf, op, err = ot.TransformWorkspaceOps(*c.Buf, op)
		if err != nil {
			return err
		}
		c.Buf = &buf
	}
	if op.String() != origStr {
		fmt.Fprintf(os.Stderr, "# remote op transformed to: %s  wait=%v  buf=%v\n", op, c.Wait, c.Buf)
	}
	if !op.Noop() {
		if err := c.Apply(op); err != nil {
			return err
		}
	}
	c.Rev++
	return nil
}
