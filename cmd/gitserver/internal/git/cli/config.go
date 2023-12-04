package cli

import (
	"context"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) Config() git.GitConfigBackend {
	return g
}

func (g *gitCLIBackend) Get(ctx context.Context, key string) (string, error) {
	cmd, cancel, err := g.gitCommand(ctx, "config", "--get", key)
	defer cancel()
	if err != nil {
		return "", err
	}

	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 means the key is not set.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
			return "", nil
		}
		return "", commandFailedError(err, cmd, out)
	}

	return strings.TrimSpace(string(out)), nil
}

func (g *gitCLIBackend) Set(ctx context.Context, key, value string) error {
	cmd, cancel, err := g.gitCommand(ctx, "config", key, value)
	defer cancel()
	if err != nil {
		return err
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return commandFailedError(err, cmd, out)
	}

	return nil
}

func (g *gitCLIBackend) Unset(ctx context.Context, key string) error {
	cmd, cancel, err := g.gitCommand(ctx, "config", "--unset-all", key)
	defer cancel()
	if err != nil {
		return err
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		// Exit code 5 means the key is not set.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 5 {
			return nil
		}

		return commandFailedError(err, cmd, out)
	}

	return nil
}
