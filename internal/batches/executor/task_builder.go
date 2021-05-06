package executor

import (
	"context"
	"fmt"

	"github.com/gobwas/glob"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

type DirectoryFinder interface {
	FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*graphql.Repository) (map[*graphql.Repository][]string, error)
}

type TaskBuilder struct {
	spec   *batches.BatchSpec
	finder DirectoryFinder

	initializedWorkspaceConfigs []batches.WorkspaceConfiguration
}

func NewTaskBuilder(spec *batches.BatchSpec, finder DirectoryFinder) (*TaskBuilder, error) {
	tb := &TaskBuilder{spec: spec, finder: finder}

	for _, conf := range tb.spec.Workspaces {
		g, err := glob.Compile(conf.In)
		if err != nil {
			return nil, err
		}
		conf.SetGlob(g)
		tb.initializedWorkspaceConfigs = append(tb.initializedWorkspaceConfigs, conf)
	}

	return tb, nil
}

func (tb *TaskBuilder) buildTask(r *graphql.Repository, path string, onlyWorkspace bool) (*Task, bool, error) {
	stepCtx := &StepContext{
		Repository: *r,
		BatchChange: BatchChangeAttributes{
			Name:        tb.spec.Name,
			Description: tb.spec.Description,
		},
	}

	var taskSteps []batches.Step
	for _, step := range tb.spec.Steps {
		if step.IfCondition() == "" {
			taskSteps = append(taskSteps, step)
			continue
		}

		static, boolVal, err := isStaticBool(step.IfCondition(), stepCtx)
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

	return &Task{
		Repository:         r,
		Path:               path,
		Steps:              taskSteps,
		OnlyFetchWorkspace: onlyWorkspace,

		TransformChanges: tb.spec.TransformChanges,
		Template:         tb.spec.ChangesetTemplate,
		BatchChangeAttributes: &BatchChangeAttributes{
			Name:        tb.spec.Name,
			Description: tb.spec.Description,
		},
	}, true, nil
}

func (tb *TaskBuilder) BuildAll(ctx context.Context, repos []*graphql.Repository) ([]*Task, error) {
	// Find workspaces in repositories, if configured
	workspaces, root, err := tb.findWorkspaces(ctx, repos, tb.initializedWorkspaceConfigs)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for repo, ws := range workspaces {
		for _, path := range ws.paths {
			t, ok, err := tb.buildTask(repo, path, ws.onlyFetchWorkspace)
			if err != nil {
				return nil, err
			}

			if ok {
				tasks = append(tasks, t)
			}
		}
	}

	for _, repo := range root {
		t, ok, err := tb.buildTask(repo, "", false)
		if err != nil {
			return nil, err
		}

		if ok {
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}

type repoWorkspaces struct {
	paths              []string
	onlyFetchWorkspace bool
}

// findWorkspaces matches the given repos to the workspace configs and
// searches, via the Sourcegraph instance, the locations of the workspaces in
// each repository.
// The repositories that were matched by a workspace config are returned in
// workspaces. root contains the repositories that didn't match a config.
// If the user didn't specify any workspaces, the repositories are returned as
// root repositories.
func (tb *TaskBuilder) findWorkspaces(
	ctx context.Context,
	repos []*graphql.Repository,
	configs []batches.WorkspaceConfiguration,
) (workspaces map[*graphql.Repository]repoWorkspaces, root []*graphql.Repository, err error) {
	if len(configs) == 0 {
		return nil, repos, nil
	}

	matched := map[int][]*graphql.Repository{}

	for _, repo := range repos {
		found := false

		for idx, conf := range configs {
			if !conf.Matches(repo.Name) {
				continue
			}

			if found {
				return nil, nil, fmt.Errorf("repository %s matches multiple workspaces.in globs in the batch spec. glob: %q", repo.Name, conf.In)
			}

			matched[idx] = append(matched[idx], repo)
			found = true
		}

		if !found {
			root = append(root, repo)
		}
	}

	workspaces = map[*graphql.Repository]repoWorkspaces{}
	for idx, repos := range matched {
		conf := configs[idx]
		repoDirs, err := tb.finder.FindDirectoriesInRepos(ctx, conf.RootAtLocationOf, repos...)
		if err != nil {
			return nil, nil, err
		}

		for repo, dirs := range repoDirs {
			workspaces[repo] = repoWorkspaces{
				paths:              dirs,
				onlyFetchWorkspace: conf.OnlyFetchWorkspace,
			}
		}
	}

	return workspaces, root, nil
}
