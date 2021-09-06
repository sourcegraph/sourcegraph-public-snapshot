package service

import (
	"context"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"

	"github.com/sourcegraph/src-cli/internal/batches/executor"
)

// buildTasks returns *executor.Tasks for all the workspaces determined for the given spec.
func buildTasks(ctx context.Context, spec *batcheslib.BatchSpec, workspaces []RepoWorkspace) []*executor.Task {
	tasks := make([]*executor.Task, 0, len(workspaces))

	for _, ws := range workspaces {
		task := &executor.Task{
			Repository:         ws.Repo,
			Path:               ws.Path,
			Steps:              ws.Steps,
			OnlyFetchWorkspace: ws.OnlyFetchWorkspace,

			TransformChanges: spec.TransformChanges,
			Template:         spec.ChangesetTemplate,
			BatchChangeAttributes: &template.BatchChangeAttributes{
				Name:        spec.Name,
				Description: spec.Description,
			},
		}
		tasks = append(tasks, task)
	}

	return tasks
}
