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
		tasks = append(tasks, buildTask(spec, ws))
	}

	return tasks
}

func buildTask(spec *batcheslib.BatchSpec, workspace RepoWorkspace) *executor.Task {
	batchChange := template.BatchChangeAttributes{
		Name:        spec.Name,
		Description: spec.Description,
	}

	// "." means the path is root, but in the executor we use "" to signify root.
	path := workspace.Path
	if path == "." {
		path = ""
	}

	return &executor.Task{
		Repository:         workspace.Repo,
		Path:               path,
		Steps:              workspace.Steps,
		OnlyFetchWorkspace: workspace.OnlyFetchWorkspace,

		TransformChanges:      spec.TransformChanges,
		Template:              spec.ChangesetTemplate,
		BatchChangeAttributes: &batchChange,
	}
}
