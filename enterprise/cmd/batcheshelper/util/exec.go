package util

import (
	"context"
	"os/exec"
)

// CmdRunner is an interface for running commands.
type CmdRunner interface {
	// Git runs the git command with the given arguments.
	Git(ctx context.Context, dir string, args ...string) ([]byte, error)
}

// RealCmdRunner is a CmdRunner that actually runs commands.
type RealCmdRunner struct{}

var _ CmdRunner = &RealCmdRunner{}

func (r *RealCmdRunner) Git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	return cmd.Output()
}
