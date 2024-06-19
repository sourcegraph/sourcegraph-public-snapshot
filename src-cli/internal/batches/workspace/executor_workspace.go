package workspace

import (
	"context"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
)

func NewExecutorWorkspaceCreator(tempDir, repoDir string) Creator {
	return &executorWorkspaceCreator{
		TempDir: tempDir,
		RepoDir: repoDir,
	}
}

type executorWorkspaceCreator struct {
	TempDir string
	RepoDir string
}

var _ Creator = &executorWorkspaceCreator{}

func (wc *executorWorkspaceCreator) Create(ctx context.Context, repo *graphql.Repository, steps []batcheslib.Step, archive repozip.Archive) (Workspace, error) {
	return &dockerBindExecutorWorkspace{
		dockerBindWorkspace: dockerBindWorkspace{
			tempDir: wc.TempDir,
			dir:     wc.RepoDir,
		},
	}, nil
}

// dockerBindExecutorWorkspace implements a workspace that operates on the host FS
// and is mounted into the docker containers using a bind mount in the end.
// It is based on the dockerBindWorkspace implementation, but does no cleanup
// as that's handled by the executor, and we want to honor it's `keepWorkspaces`
// setting.
type dockerBindExecutorWorkspace struct {
	dockerBindWorkspace
}

var _ Workspace = &dockerBindExecutorWorkspace{}

func (w *dockerBindExecutorWorkspace) Close(ctx context.Context) error {
	// Nothing to do here, executor cleanup will handle this.
	return nil
}
