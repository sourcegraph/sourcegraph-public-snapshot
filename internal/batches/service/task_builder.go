package service

import (
	"context"

	"github.com/pkg/errors"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/template"
)

type directoryFinder interface {
	FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*graphql.Repository) (map[*graphql.Repository][]string, error)
}

type taskBuilder struct {
	spec   *batcheslib.BatchSpec
	finder directoryFinder
}

// buildTasks returns tasks for all the workspaces determined for the given spec.
func buildTasks(ctx context.Context, spec *batcheslib.BatchSpec, finder directoryFinder, repos []*graphql.Repository, workspaces []RepoWorkspaces) ([]*executor.Task, error) {
	tb := &taskBuilder{spec: spec, finder: finder}
	return tb.buildAll(ctx, repos, workspaces)
}

func (tb *taskBuilder) buildTask(r *graphql.Repository, path string, onlyWorkspace bool) (*executor.Task, bool, error) {
	stepCtx := &template.StepContext{
		Repository: *r,
		BatchChange: template.BatchChangeAttributes{
			Name:        tb.spec.Name,
			Description: tb.spec.Description,
		},
	}

	var taskSteps []batcheslib.Step
	for _, step := range tb.spec.Steps {
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

		TransformChanges: tb.spec.TransformChanges,
		Template:         tb.spec.ChangesetTemplate,
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        tb.spec.Name,
			Description: tb.spec.Description,
		},
	}, true, nil
}

func (tb *taskBuilder) buildAll(ctx context.Context, repos []*graphql.Repository, workspaces []RepoWorkspaces) ([]*executor.Task, error) {
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
			t, ok, err := tb.buildTask(repo, path, fetchWorkspace)
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
