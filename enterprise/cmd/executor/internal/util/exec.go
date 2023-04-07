package util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CmdRunner is an interface for running commands.
type CmdRunner interface {
	// CommandContext returns the Cmd struct to execute the named program with the given arguments.
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
	// CombinedOutput runs the command and returns its combined standard output and standard error.
	CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)
	// LookPath looks for an executable named file in the directories named by the PATH environment variable.
	LookPath(file string) (string, error)
	// Stat returns a FileInfo describing the named file.
	Stat(filename string) (os.FileInfo, error)
}

// RealCmdRunner is a CmdRunner that actually runs commands.
type RealCmdRunner struct{}

var _ CmdRunner = &RealCmdRunner{}

func (r *RealCmdRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

func (r *RealCmdRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	return r.CommandContext(ctx, name, args...).CombinedOutput()
}

func (r *RealCmdRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (r *RealCmdRunner) Stat(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

func execOutput(ctx context.Context, runner CmdRunner, name string, args ...string) (string, error) {
	b, err := runner.CombinedOutput(ctx, name, args...)
	if err != nil {
		cmdLine := strings.Join(append([]string{name}, args...), " ")
		return "", errors.Wrap(err, fmt.Sprintf("'%s': %s", cmdLine, string(b)))
	}
	return strings.TrimSpace(string(b)), nil
}

// ExistsPath returns true if the given path exists.
func ExistsPath(runner CmdRunner, name string) (bool, error) {
	if _, err := runner.LookPath(name); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
