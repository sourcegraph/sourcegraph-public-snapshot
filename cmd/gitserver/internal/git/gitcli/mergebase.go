package gitcli

import (
	"bytes"
	"context"
	"io"
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

	out, err := g.runGitCommand(ctx, cmd)
	if err != nil {
		return "", err
	}
	defer out.Close()

	stdout, err := io.ReadAll(out)
	if err != nil {
		// Exit code 1 and empty output most likely means that no common merge-base was found.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
			if len(e.Stderr) == 0 {
				return "", nil
			}
		}

		return "", err
	}

	return api.CommitID(bytes.TrimSpace(stdout)), nil
}
