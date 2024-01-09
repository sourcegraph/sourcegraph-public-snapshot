package cli

import (
	"bytes"
	"context"
	"os/exec"
	"syscall"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) MergeBase(ctx context.Context, baseRevspec, headRevspec string) (api.CommitID, error) {
	cmd, cancel, err := g.gitCommand(ctx, "merge-base", "--", baseRevspec, headRevspec)
	defer cancel()
	if err != nil {
		return "", err
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		// Exit code 1 and empty output most likely means that no common merge-base was found.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
			if len(out) == 0 {
				return "", nil
			}
		}

		return "", commandFailedError(err, cmd, out)
	}

	return api.CommitID(bytes.TrimSpace(out)), nil
}
