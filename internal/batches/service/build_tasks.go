package service

import (
	"context"

	"github.com/pkg/errors"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

// buildTasks returns tasks for all the workspaces determined for the given spec.
func buildTasks(ctx context.Context, spec *batcheslib.BatchSpec, repos []*graphql.Repository, workspaces []RepoWorkspace) ([]*executor.Task, error) {
	repoByID := make(map[string]*graphql.Repository)
	for _, repo := range repos {
		repoByID[repo.ID] = repo
	}

	tasks := []*executor.Task{}
	for _, ws := range workspaces {
		repo, ok := repoByID[ws.RepoID]
		if !ok {
			return nil, errors.New("invalid state, didn't find repo for workspace definition")
		}

		t, err := buildTask(spec, repo, ws.Path, ws.OnlyFetchWorkspace, ws.Steps)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func buildTask(spec *batcheslib.BatchSpec, r *graphql.Repository, path string, onlyWorkspace bool, steps []batcheslib.Step) (*executor.Task, error) {
	batchChange := template.BatchChangeAttributes{
		Name:        spec.Name,
		Description: spec.Description,
	}

	// "." means the path is root, but in the executor we use "" to signify root
	if path == "." {
		path = ""
	}

	return &executor.Task{
		Repository:         r,
		Path:               path,
		Steps:              steps,
		OnlyFetchWorkspace: onlyWorkspace,

		TransformChanges:      spec.TransformChanges,
		Template:              spec.ChangesetTemplate,
		BatchChangeAttributes: &batchChange,
	}, nil
}
