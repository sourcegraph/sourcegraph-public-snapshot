package cli

import (
	"context"
	"io"
)

func (g *gitCLIBackend) Exec(ctx context.Context, args ...string) (io.ReadCloser, error) {
	cmd, cancel, err := g.gitCommand(ctx, args...)
	defer cancel()
	if err != nil {
		return nil, err
	}

	return g.runGitCommand(ctx, cmd)
}
