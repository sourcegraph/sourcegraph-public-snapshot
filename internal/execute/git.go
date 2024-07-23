package execute

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// HandleGitCommandExec There's a weird behavior that occurs where an error isn't accessible in the err variable
// from a *Cmd executing a git command after calling CombinedOutput().
// This occurs due to how Git handles errors and how the exec package in Go interprets the command's output.
// Git often writes error messages to stderr, but it might still exit with a status code of 0 (indicating success).
// In this case, CombinedOutput() won't return an error, but the error message will be in the out variable.
func handleGitCommandExec(cmd *exec.Cmd) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		maybeErrMessage := strings.Trim(stderr.String(), "\n")
		if strings.HasPrefix(maybeErrMessage, "fatal:") || strings.HasPrefix(maybeErrMessage, "error:") {
			return nil, errors.New(maybeErrMessage)
		}
		return nil, err
	}

	return stdout.Bytes(), nil
}

func Git(ctx context.Context, args ...string) ([]byte, error) {
	return handleGitCommandExec(GitCmd(ctx, args...))
}

func GitCmd(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "git", args...)
}

func GHCmd(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "gh", args...)
}

func GH(ctx context.Context, args ...string) ([]byte, error) {
	cmd := GHCmd(ctx, args...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}
