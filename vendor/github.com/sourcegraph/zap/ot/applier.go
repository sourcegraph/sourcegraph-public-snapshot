package ot

import (
	"context"

	"github.com/go-kit/kit/log"
)

// Applier is the interface implemented by things that can apply an OT
// op. If applying the op succeeds, the Applier's internal state is
// modified to reflect the op's change. If applying the op fails, an
// error is returned and the Applier's internal state is unchanged.
type Applier interface {
	Apply(context.Context, log.Logger, WorkspaceOp) error
}

// ApplierFunc implements the Applier interface with a func.
type ApplierFunc func(context.Context, log.Logger, WorkspaceOp) error

// Apply implements Applier.
func (f ApplierFunc) Apply(ctx context.Context, logger log.Logger, op WorkspaceOp) error {
	return f(ctx, logger, op)
}
