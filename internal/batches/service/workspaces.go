package service

import (
	"context"
	"sort"

	"github.com/gobwas/glob"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/util"
)

type RepoWorkspace struct {
	Repo               *graphql.Repository
	Path               string
	OnlyFetchWorkspace bool
}

type directoryFinder interface {
	FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*graphql.Repository) (map[*graphql.Repository][]string, error)
}

// findWorkspaces matches the given repos to the workspace configs and
// searches, via the Sourcegraph instance, the locations of the workspaces in
// each repository.
// The repositories that were matched by a workspace config and all repos that didn't
// match a config are returned as workspaces.
func findWorkspaces(
	ctx context.Context,
	spec *batcheslib.BatchSpec,
	finder directoryFinder,
	repos []*graphql.Repository,
) ([]RepoWorkspace, error) {
	// Pre-compile all globs.
	workspaceMatchers := make(map[batcheslib.WorkspaceConfiguration]glob.Glob)
	for _, conf := range spec.Workspaces {
		in := conf.In
		// Empty `in` should fall back to matching all, instead of nothing.
		if in == "" {
			in = "*"
		}
		g, err := glob.Compile(in)
		if err != nil {
			return nil, batcheslib.NewValidationError(errors.Errorf("failed to compile glob %q: %v", in, err))
		}
		workspaceMatchers[conf] = g
	}

	root := []*graphql.Repository{}

	// Maps workspace config indexes to repositories matching them.
	matched := map[int][]*graphql.Repository{}

	for _, repo := range repos {
		found := false

		// Try to find a workspace configuration matching this repo.
		for idx, conf := range spec.Workspaces {
			if !workspaceMatchers[conf].Match(repo.Name) {
				continue
			}

			// Don't allow duplicate matches.
			if found {
				return nil, batcheslib.NewValidationError(errors.Errorf("repository %s matches multiple workspaces.in globs in the batch spec. glob: %q", repo.Name, conf.In))
			}

			matched[idx] = append(matched[idx], repo)
			found = true
		}

		if !found {
			root = append(root, repo)
		}
	}

	type repoWorkspaces struct {
		Repo               *graphql.Repository
		Paths              []string
		OnlyFetchWorkspace bool
	}
	type workspaceKey struct {
		repo   string
		branch string
	}
	workspacesByKey := map[workspaceKey]repoWorkspaces{}
	for idx, repos := range matched {
		conf := spec.Workspaces[idx]
		repoDirs, err := finder.FindDirectoriesInRepos(ctx, conf.RootAtLocationOf, repos...)
		if err != nil {
			return nil, err
		}

		for repo, dirs := range repoDirs {
			// Don't add repos that don't have any matched workspaces.
			if len(dirs) == 0 {
				continue
			}
			key := workspaceKey{
				repo:   repo.ID,
				branch: repo.Branch.Name,
			}
			workspacesByKey[key] = repoWorkspaces{
				Repo:               repo,
				Paths:              dirs,
				OnlyFetchWorkspace: conf.OnlyFetchWorkspace,
			}
		}
	}

	// And add the root for repos.
	for _, repo := range root {
		key := workspaceKey{
			repo:   repo.ID,
			branch: repo.Branch.Name,
		}
		conf, ok := workspacesByKey[key]
		if !ok {
			workspacesByKey[key] = repoWorkspaces{
				Repo:               repo,
				Paths:              []string{""},
				OnlyFetchWorkspace: false,
			}
			continue
		}
		conf.Paths = append(workspacesByKey[key].Paths, "")
	}

	workspaces := make([]RepoWorkspace, 0, len(workspacesByKey))
	for _, workspace := range workspacesByKey {
		for _, path := range workspace.Paths {
			fetchWorkspace := workspace.OnlyFetchWorkspace
			if path == "" {
				fetchWorkspace = false
			}

			steps, err := stepsForRepo(spec, util.NewTemplatingRepo(workspace.Repo.Name, workspace.Repo.FileMatches))
			if err != nil {
				return nil, err
			}

			// If the workspace doesn't have any steps we don't need to include it.
			if len(steps) == 0 {
				continue
			}

			workspaces = append(workspaces, RepoWorkspace{
				Repo:               workspace.Repo,
				Path:               path,
				OnlyFetchWorkspace: fetchWorkspace,
			})
		}
	}

	// Stable sorting.
	sort.Slice(workspaces, func(i, j int) bool {
		if workspaces[i].Repo.Name == workspaces[j].Repo.Name {
			return workspaces[i].Path < workspaces[j].Path
		}
		return workspaces[i].Repo.Name < workspaces[j].Repo.Name
	})

	return workspaces, nil
}

// stepsForRepo calculates the steps required to run on the given repo.
func stepsForRepo(spec *batcheslib.BatchSpec, repo template.Repository) ([]batcheslib.Step, error) {
	taskSteps := []batcheslib.Step{}
	for _, step := range spec.Steps {
		// If no if condition is given, just go ahead and add the step to the list.
		if step.IfCondition() == "" {
			taskSteps = append(taskSteps, step)
			continue
		}

		batchChange := template.BatchChangeAttributes{
			Name:        spec.Name,
			Description: spec.Description,
		}
		stepCtx := &template.StepContext{
			Repository:  repo,
			BatchChange: batchChange,
		}
		static, boolVal, err := template.IsStaticBool(step.IfCondition(), stepCtx)
		if err != nil {
			return nil, err
		}

		// If we could evaluate the condition statically and the resulting
		// boolean is false, we don't add that step.
		if !static {
			taskSteps = append(taskSteps, step)
		} else if boolVal {
			taskSteps = append(taskSteps, step)
		}
	}
	return taskSteps, nil
}
