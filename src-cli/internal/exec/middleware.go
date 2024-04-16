package exec

import (
	"context"
	goexec "os/exec"
)

// CmdCreatorMiddleware creates *exec.Cmd instances that delegate command
// creation to a provided callback.
type CmdCreatorMiddleware struct{ previous CmdCreator }

// NewMiddleware adds a middleware to the command creation stack.
func NewMiddleware(mock func(context.Context, CmdCreator, string, ...string) *goexec.Cmd) CmdCreatorMiddleware {
	mc := CmdCreatorMiddleware{previous: creator}
	creator = func(ctx context.Context, name string, arg ...string) *goexec.Cmd {
		return mock(ctx, mc.previous, name, arg...)
	}
	return mc
}

// Remove removes the command creation middleware from the stack.
func (mc CmdCreatorMiddleware) Remove() {
	creator = mc.previous
}
