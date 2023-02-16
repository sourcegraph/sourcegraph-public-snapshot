package graphqlbackend

import (
	"github.com/hexops/autogold"

	"context"
	"sort"
	"sync"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserCollaborators_gitserverParallelRecentCommitters(t *testing.T) {
	ctx := context.Background()

	type args struct {
		repoName api.RepoName
		opt      gitserver.CommitsOptions
	}
	var (
		callsMu sync.Mutex
		calls   []args
	)
	gitCommitsFunc := func(ctx context.Context, perms authz.SubRepoPermissionChecker, repoName api.RepoName, opt gitserver.CommitsOptions) ([]*gitdomain.Commit, error) {
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
	recentCommitters := gitserverParallelRecentCommitters(ctx, repos, gitCommitsFunc)

	sort.Slice(calls, func(i, j int) bool {
		return calls[i].repoName < calls[j].repoName
	})
	sort.Slice(recentCommitters, func(i, j int) bool {
		return recentCommitters[i].name < recentCommitters[j].name
	})

	autogold.Want("calls", []args{
		{
			repoName: "golang/go",
			opt: gitserver.CommitsOptions{
				N:                200,
				NoEnsureRevision: true,
				NameOnly:         true,
			},
		},
		{
			repoName: "gorilla/mux",
			opt: gitserver.CommitsOptions{
				N:                200,
				NoEnsureRevision: true,
				NameOnly:         true,
			},
		},
		{
			repoName: "sourcegraph/sourcegraph",
			opt: gitserver.CommitsOptions{
				N:                200,
				NoEnsureRevision: true,
				NameOnly:         true,
			},
		},
	}).Equal(t, calls)

	autogold.Want("recentCommitters", []*invitableCollaboratorResolver{
		{
			name:      "golang/go-jane",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "golang/go-janet",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "golang/go-joe",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "gorilla/mux-jane",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "gorilla/mux-janet",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "gorilla/mux-joe",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "sourcegraph/sourcegraph-jane",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "sourcegraph/sourcegraph-janet",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
		{
			name:      "sourcegraph/sourcegraph-joe",
			avatarURL: "https://www.gravatar.com/avatar/d41d8cd98f00b204e9800998ecf8427e?d=mp",
		},
	}).Equal(t, recentCommitters)
}
