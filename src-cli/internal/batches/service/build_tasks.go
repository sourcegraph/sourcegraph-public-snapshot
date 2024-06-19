package service

import (
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"

	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

type RepoWorkspace struct {
	Repo               *graphql.Repository
	Path               string
	OnlyFetchWorkspace bool
}

// buildTasks returns *executor.Tasks for all the workspaces determined for the given spec.
func buildTasks(attributes *template.BatchChangeAttributes, steps []batcheslib.Step, workspaces []RepoWorkspace) []*executor.Task {
	tasks := make([]*executor.Task, 0, len(workspaces))

	for _, ws := range workspaces {
		task := &executor.Task{
			Repository:         ws.Repo,
			Path:               ws.Path,
			Steps:              steps,
			OnlyFetchWorkspace: ws.OnlyFetchWorkspace,

			BatchChangeAttributes: attributes,
		}
		tasks = append(tasks, task)
	}

	return tasks
}
