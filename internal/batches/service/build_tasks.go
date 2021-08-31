package service

import (
	"context"

	"github.com/pkg/errors"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/template"
)

// buildTasks returns tasks for all the workspaces determined for the given spec.
func buildTasks(ctx context.Context, spec *batcheslib.BatchSpec, repos []*graphql.Repository, workspaces []RepoWorkspaces) ([]*executor.Task, error) {
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
		for _, path := range ws.Paths {
			fetchWorkspace := ws.OnlyFetchWorkspace
			if path == "" {
				fetchWorkspace = false
			}
			t, ok, err := buildTask(spec, repo, path, fetchWorkspace)
			if err != nil {
				return nil, err
			}

			if ok {
				tasks = append(tasks, t)
			}
		}
	}

	return tasks, nil
}

func buildTask(spec *batcheslib.BatchSpec, r *graphql.Repository, path string, onlyWorkspace bool) (*executor.Task, bool, error) {
	stepCtx := &template.StepContext{
		Repository: *r,
		BatchChange: template.BatchChangeAttributes{
			Name:        spec.Name,
			Description: spec.Description,
		},
	}

	var taskSteps []batcheslib.Step
	for _, step := range spec.Steps {
		if step.IfCondition() == "" {
			taskSteps = append(taskSteps, step)
			continue
		}

		static, boolVal, err := template.IsStaticBool(step.IfCondition(), stepCtx)
		if err != nil {
			return nil, false, err
		}

		// If we could evaluate the condition statically and the resulting
		// boolean is false, we don't add that step.
		if !static {
			taskSteps = append(taskSteps, step)
		} else if boolVal {
			taskSteps = append(taskSteps, step)
		}
	}

	// If the task doesn't have any steps we don't need to execute it
	if len(taskSteps) == 0 {
		return nil, false, nil
	}

	// "." means the path is root, but in the executor we use "" to signify root
	if path == "." {
		path = ""
	}

	return &executor.Task{
		Repository:         r,
		Path:               path,
		Steps:              taskSteps,
		OnlyFetchWorkspace: onlyWorkspace,

		TransformChanges: spec.TransformChanges,
		Template:         spec.ChangesetTemplate,
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        spec.Name,
			Description: spec.Description,
		},
	}, true, nil
}
