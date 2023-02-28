package worker

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
)

// prepareWorkspace creates and returns a temporary directory in which acts the workspace
// while processing a single job. It is up to the caller to ensure that this directory is
// removed after the job has finished processing. If a repository name is supplied, then
// that repository will be cloned (through the frontend API) into the workspace.
func (h *handler) prepareWorkspace(
	ctx context.Context,
	cmd command.Command,
	job types.Job,
	commandLogger command.Logger,
) (workspace.Workspace, error) {
	if h.options.RunnerOptions.FirecrackerOptions.Enabled {
		return workspace.NewFirecrackerWorkspace(
			ctx,
			h.filesStore,
			job,
			h.options.RunnerOptions.DockerOptions.Resources.DiskSpace,
			h.options.KeepWorkspaces,
			cmd,
			commandLogger,
			h.cloneOptions,
			h.operations,
		)
	}

	return workspace.NewDockerWorkspace(
		ctx,
		h.filesStore,
		job,
		cmd,
		commandLogger,
		h.cloneOptions,
		h.operations,
	)
}
