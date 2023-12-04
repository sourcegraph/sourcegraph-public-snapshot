package cli

import (
	"bytes"
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func (g *gitCLIBackend) MergeBase(ctx context.Context, baseRevspec, headRevspec string) (api.CommitID, error) {
	cmd, cancel, err := g.gitCommand(ctx, "merge-base", "--", baseRevspec, headRevspec)
	defer cancel()
	if err != nil {
		return "", err
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", commandFailedError(err, cmd, out)
	}

	return api.CommitID(bytes.TrimSpace(out)), nil
}
