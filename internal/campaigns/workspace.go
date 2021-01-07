package campaigns

import (
	"context"

	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

// WorkspaceCreator implementations are used to create workspaces, which manage
// per-changeset persistent storage when executing campaign steps and are
// responsible for ultimately generating a diff.
type WorkspaceCreator interface {
	// Create creates a new workspace for the given repository and ZIP file.
	Create(ctx context.Context, repo *graphql.Repository, zip string) (Workspace, error)

	// DockerImages returns any Docker images required to use workspaces created
	// by this creator.
	DockerImages() []string
}

// Workspace implementations manage per-changeset storage when executing
// campaign step.
type Workspace interface {
	// DockerRunOpts provides the options that should be given to `docker run`
	// in order to use this workspace. Generally, this will be a set of mount
	// options.
	DockerRunOpts(ctx context.Context, target string) ([]string, error)

	// WorkDir allows workspaces to specify the working directory that should be
	// used when running Docker. If no specific working directory is needed,
	// then the function should return nil.
	WorkDir() *string

	// Close is called once, after all steps have been executed and the diff has
	// been calculated and stored outside the workspace. Implementations should
	// delete the workspace when Close is called.
	Close(ctx context.Context) error

	// Changes is called after each step is executed, and should return the
	// cumulative file changes that have occurred since Prepare was called.
	Changes(ctx context.Context) (*StepChanges, error)

	// Diff should return the total diff for the workspace. This may be called
	// multiple times in the life of a workspace.
	Diff(ctx context.Context) ([]byte, error)
}
