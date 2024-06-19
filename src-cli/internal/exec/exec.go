// Package exec provides wrapped implementations of os/exec's Command and
// CommandContext functions that allow for command creation to be overridden,
// thereby allowing commands to be mocked.
package exec

import (
	"context"
	goexec "os/exec"
)

// CmdCreator instances are used to create commands. os/exec.CommandContext is a
// valid CmdCreator.
type CmdCreator func(context.Context, string, ...string) *goexec.Cmd

// creator is the singleton used to create a new command.
var creator CmdCreator = goexec.CommandContext

// Command wraps os/exec.Command, and implements the same behaviour.
func Command(name string, arg ...string) *goexec.Cmd {
	return CommandContext(context.TODO(), name, arg...)
}

// CommandContext wraps os/exec.CommandContext, and implements the same
// behaviour.
func CommandContext(ctx context.Context, name string, arg ...string) *goexec.Cmd {
	// TODO: if we add global logging infrastructure to cmd/src, we could
	// leverage it here to log all commands that are executed in an appropriate
	// verbose mode.

	return creator(ctx, name, arg...)
}
