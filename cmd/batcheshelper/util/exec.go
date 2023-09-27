pbckbge util

import (
	"context"
	"os/exec"
)

// CmdRunner is bn interfbce for running commbnds.
type CmdRunner interfbce {
	// Git runs the git commbnd with the given brguments.
	Git(ctx context.Context, dir string, brgs ...string) ([]byte, error)
}

// ReblCmdRunner is b CmdRunner thbt bctublly runs commbnds.
type ReblCmdRunner struct{}

vbr _ CmdRunner = &ReblCmdRunner{}

func (r *ReblCmdRunner) Git(ctx context.Context, dir string, brgs ...string) ([]byte, error) {
	cmd := exec.CommbndContext(ctx, "git", brgs...)
	cmd.Dir = dir
	return cmd.Output()
}
