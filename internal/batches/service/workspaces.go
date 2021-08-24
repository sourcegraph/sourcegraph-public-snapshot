package service

import (
	"context"
	"fmt"

	"github.com/gobwas/glob"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

type RepoWorkspaces struct {
	RepoID             string
	Paths              []string
	OnlyFetchWorkspace bool
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
) ([]RepoWorkspaces, error) {
	// Pre-compile all globs.
	workspaceMatchers := make(map[batcheslib.WorkspaceConfiguration]glob.Glob)
	for _, conf := range spec.Workspaces {
		g, err := glob.Compile(conf.In)
		if err != nil {
			return nil, batches.ValidationError{Reason: fmt.Sprintf("failed to compile glob %q: %v", conf.In, err)}
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
				return nil, batches.ValidationError{Reason: fmt.Sprintf("repository %s matches multiple workspaces.in globs in the batch spec. glob: %q", repo.Name, conf.In)}
			}

			matched[idx] = append(matched[idx], repo)
			found = true
		}

		if !found {
			root = append(root, repo)
		}
	}

	workspacesByID := map[string]RepoWorkspaces{}
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
			workspacesByID[repo.ID] = RepoWorkspaces{
				RepoID:             repo.ID,
				Paths:              dirs,
				OnlyFetchWorkspace: conf.OnlyFetchWorkspace,
			}
		}
	}

	// And add the root for repos.
	for _, repo := range root {
		conf, ok := workspacesByID[repo.ID]
		if !ok {
			workspacesByID[repo.ID] = RepoWorkspaces{
				RepoID:             repo.ID,
				Paths:              []string{""},
				OnlyFetchWorkspace: false,
			}
			continue
		}
		conf.Paths = append(workspacesByID[repo.ID].Paths, "")
	}

	workspaces := make([]RepoWorkspaces, 0, len(workspacesByID))
	for _, workspace := range workspacesByID {
		workspaces = append(workspaces, workspace)
	}
	return workspaces, nil
}
