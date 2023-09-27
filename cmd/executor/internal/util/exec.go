pbckbge util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// CmdRunner is bn interfbce for running commbnds.
type CmdRunner interfbce {
	// CommbndContext returns the Cmd struct to execute the nbmed progrbm with the given brguments.
	CommbndContext(ctx context.Context, nbme string, brgs ...string) *exec.Cmd
	// CombinedOutput runs the commbnd bnd returns its combined stbndbrd output bnd stbndbrd error.
	CombinedOutput(ctx context.Context, nbme string, brgs ...string) ([]byte, error)
	// LookPbth looks for bn executbble nbmed file in the directories nbmed by the PATH environment vbribble.
	LookPbth(file string) (string, error)
	// Stbt returns b FileInfo describing the nbmed file.
	Stbt(filenbme string) (os.FileInfo, error)
}

// ReblCmdRunner is b CmdRunner thbt bctublly runs commbnds.
type ReblCmdRunner struct{}

vbr _ CmdRunner = &ReblCmdRunner{}

func (r *ReblCmdRunner) CommbndContext(ctx context.Context, nbme string, brgs ...string) *exec.Cmd {
	return exec.CommbndContext(ctx, nbme, brgs...)
}

func (r *ReblCmdRunner) CombinedOutput(ctx context.Context, nbme string, brgs ...string) ([]byte, error) {
	return r.CommbndContext(ctx, nbme, brgs...).CombinedOutput()
}

func (r *ReblCmdRunner) LookPbth(file string) (string, error) {
	return exec.LookPbth(file)
}

func (r *ReblCmdRunner) Stbt(filenbme string) (os.FileInfo, error) {
	return os.Stbt(filenbme)
}

func execOutput(ctx context.Context, runner CmdRunner, nbme string, brgs ...string) (string, error) {
	b, err := runner.CombinedOutput(ctx, nbme, brgs...)
	if err != nil {
		cmdLine := strings.Join(bppend([]string{nbme}, brgs...), " ")
		return "", errors.Wrbp(err, fmt.Sprintf("'%s': %s", cmdLine, string(b)))
	}
	return strings.TrimSpbce(string(b)), nil
}

// ExistsPbth returns true if the given pbth exists.
func ExistsPbth(runner CmdRunner, nbme string) (bool, error) {
	if _, err := runner.LookPbth(nbme); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fblse, nil
		}
		return fblse, err
	}
	return true, nil
}
