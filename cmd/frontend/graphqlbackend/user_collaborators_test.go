package graphqlbackend

import (
	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"

	"context"
	"sort"
	"sync"
	"testing"
)

func TestUserCollaborators_gitserverParallelRecentCommitters(t *testing.T) {
	ctx := context.Background()

	type args struct {
		repoName api.RepoName
		opt      git.CommitsOptions
	}
	var (
		callsMu sync.Mutex
		calls   []args
	)
	gitCommitsFunc := func(ctx context.Context, repoName api.RepoName, opt git.CommitsOptions, perms authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error) {
		callsMu.Lock()
		calls = append(calls, args{repoName, opt})
		callsMu.Unlock()

		return []*gitdomain.Commit{
			{
				Author: gitdomain.Signature{
					Name: string(repoName) + "-joe",
				},
			},
			{
				Author: gitdomain.Signature{
					Name: string(repoName) + "-jane",
				},
			},
			{
				Author: gitdomain.Signature{
					Name: string(repoName) + "-janet",
				},
			},
		}, nil
	}

	repos := []*types.Repo{
		{Name: "gorilla/mux"},
		{Name: "golang/go"},
		{Name: "sourcegraph/sourcegraph"},
	}
	recentCommitters, err := gitserverParallelRecentCommitters(ctx, repos, gitCommitsFunc)
	if err != nil {
		t.Fatal(err)
	}

	sort.Slice(calls, func(i, j int) bool {
		return calls[i].repoName < calls[j].repoName
	})
	sort.Slice(recentCommitters, func(i, j int) bool {
		return recentCommitters[i].name < recentCommitters[j].name
	})

	autogold.Want("calls", nil).Equal(t, calls)

	autogold.Want("recentCommitters", nil).Equal(t, recentCommitters)
}
