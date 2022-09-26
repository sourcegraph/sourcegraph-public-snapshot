package worker

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
)

// prepareWorkspace creates and returns a temporary directory in which acts the workspace
// while processing a single job. It is up to the caller to ensure that this directory is
// removed after the job has finished processing. If a repository name is supplied, then
// that repository will be cloned (through the frontend API) into the workspace.
func (h *handler) prepareWorkspace(
	ctx context.Context,
	commandRunner command.Runner,
	job executor.Job,
	commandLogger command.Logger,
) (workspace.Workspace, error) {
	if h.options.FirecrackerOptions.Enabled {
		return workspace.NewFirecrackerWorkspace(
			ctx,
			h.filesStore,
			job,
			h.options.ResourceOptions.DiskSpace,
			h.options.KeepWorkspaces,
			commandRunner,
			commandLogger,
			workspace.CloneOptions{
				EndpointURL:    h.options.QueueOptions.BaseClientOptions.EndpointOptions.URL,
				GitServicePath: h.options.GitServicePath,
				ExecutorToken:  h.options.QueueOptions.BaseClientOptions.EndpointOptions.Token,
			},
			h.operations,
		)
	}

	return workspace.NewDockerWorkspace(
		ctx,
		h.filesStore,
		job,
		commandRunner,
		commandLogger,
		workspace.CloneOptions{
			EndpointURL:    h.options.QueueOptions.BaseClientOptions.EndpointOptions.URL,
			GitServicePath: h.options.GitServicePath,
			ExecutorToken:  h.options.QueueOptions.BaseClientOptions.EndpointOptions.Token,
		},
		h.operations,
	)
}
