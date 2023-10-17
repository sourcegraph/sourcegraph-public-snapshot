package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCheckMirrorRepositoryConnection(t *testing.T) {
	const repoName = api.RepoName("my/repo")

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	repos := dbmocks.NewMockRepoStore()

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ReposFunc.SetDefaultReturn(repos)

	t.Run("repository arg", func(t *testing.T) {
		backend.Mocks.Repos.Get = func(ctx context.Context, repoID api.RepoID) (*types.Repo, error) {
			return &types.Repo{Name: repoName}, nil
		}

		calledIsRepoCloneable := false
		gitserver.MockIsRepoCloneable = func(repo api.RepoName) error {
			calledIsRepoCloneable = true
			if want := repoName; !reflect.DeepEqual(repo, want) {
				t.Errorf("got %+v, want %+v", repo, want)
			}
			return nil
		}
		t.Cleanup(func() {
			backend.Mocks = backend.MockServices{}
			gitserver.MockIsRepoCloneable = nil
		})

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
				mutation {
					checkMirrorRepositoryConnection(repository: "UmVwb3NpdG9yeToxMjM=") {
					    error
					}
				}
			`,
				ExpectedResult: `
				{
					"checkMirrorRepositoryConnection": {
						"error": null
					}
				}
			`,
			},
		})

		if !calledIsRepoCloneable {
			t.Error("!calledIsRepoCloneable")
		}
	})
}

func TestCheckMirrorRepositoryRemoteURL(t *testing.T) {
	const repoName = "my/repo"

	cases := []struct {
		repoURL string
		want    string
	}{
		{
			repoURL: "git@github.com:gorilla/mux.git",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"github.com:gorilla/mux.git"}}}`,
		},
		{
			repoURL: "git+https://github.com/gorilla/mux.git",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"git+https://github.com/gorilla/mux.git"}}}`,
		},
		{
			repoURL: "https://github.com/gorilla/mux.git",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"https://github.com/gorilla/mux.git"}}}`,
		},
		{
			repoURL: "https://github.com/gorilla/mux",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"https://github.com/gorilla/mux"}}}`,
		},
		{
			repoURL: "ssh://git@github.com/gorilla/mux",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://github.com/gorilla/mux"}}}`,
		},
		{
			repoURL: "ssh://github.com/gorilla/mux.git",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://github.com/gorilla/mux.git"}}}`,
		},
		{
			repoURL: "ssh://git@github.com:/my/repo.git",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://github.com:/my/repo.git"}}}`,
		},
		{
			repoURL: "git://git@github.com:/my/repo.git",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"git://github.com:/my/repo.git"}}}`,
		},
		{
			repoURL: "user@host.xz:/path/to/repo.git/",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"host.xz:/path/to/repo.git/"}}}`,
		},
		{
			repoURL: "host.xz:/path/to/repo.git/",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"host.xz:/path/to/repo.git/"}}}`,
		},
		{
			repoURL: "ssh://user@host.xz:1234/path/to/repo.git/",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://host.xz:1234/path/to/repo.git/"}}}`,
		},
		{
			repoURL: "host.xz:~user/path/to/repo.git/",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"host.xz:~user/path/to/repo.git/"}}}`,
		},
		{
			repoURL: "ssh://host.xz/~/path/to/repo.git",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://host.xz/~/path/to/repo.git"}}}`,
		},
		{
			repoURL: "git://host.xz/~user/path/to/repo.git/",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"git://host.xz/~user/path/to/repo.git/"}}}`,
		},
		{
			repoURL: "file:///path/to/repo.git/",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"file:///path/to/repo.git/"}}}`,
		},
		{
			repoURL: "file://~/path/to/repo.git/",
			want:    `{"repository":{"mirrorInfo":{"remoteURL":"file://~/path/to/repo.git/"}}}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.repoURL, func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
				return &types.Repo{
					Name:      repoName,
					CreatedAt: time.Now(),
					Sources:   map[string]*types.SourceInfo{"1": {CloneURL: tc.repoURL}},
				}, nil
			}
			t.Cleanup(func() {
				backend.Mocks = backend.MockServices{}
			})

			RunTests(t, []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
					{
						repository(name: "my/repo") {
							mirrorInfo {
								remoteURL
							}
						}
					}
				`,
					ExpectedResult: tc.want,
				},
			})
		})
	}
}

type fakeGitserverClient struct {
	gitserver.Client
}

func (f *fakeGitserverClient) RepoCloneProgress(_ context.Context, repoName ...api.RepoName) (*protocol.RepoCloneProgressResponse, error) {
	results := map[api.RepoName]*protocol.RepoCloneProgress{}
	for _, n := range repoName {
		results[n] = &protocol.RepoCloneProgress{
			CloneInProgress: true,
			CloneProgress:   fmt.Sprintf("cloning fake %s...", n),
			Cloned:          false,
		}
	}
	return &protocol.RepoCloneProgressResponse{
		Results: results,
	}, nil
}

func TestRepositoryMirrorInfoCloneProgressCallsGitserver(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{
			Name:      "repo-name",
			CreatedAt: time.Now(),
			Sources:   map[string]*types.SourceInfo{"1": {}},
		}, nil
	}
	t.Cleanup(func() {
		backend.Mocks = backend.MockServices{}
	})

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchemaWithClient(t, db, &fakeGitserverClient{}),
		Query: `
			{
				repository(name: "my/repo") {
					mirrorInfo {
						cloneProgress
					}
				}
			}
		`,
		ExpectedResult: `
			{
				"repository": {
					"mirrorInfo": {
						"cloneProgress": "cloning fake repo-name..."
					}
				}
			}
		`,
	})
}

func TestRepositoryMirrorInfoCloneProgressFetchedFromDatabase(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	featureFlags := dbmocks.NewMockFeatureFlagStore()
	featureFlags.GetGlobalFeatureFlagsFunc.SetDefaultReturn(map[string]bool{"clone-progress-logging": true}, nil)

	gitserverRepos := dbmocks.NewMockGitserverRepoStore()
	gitserverRepos.GetByIDFunc.SetDefaultReturn(&types.GitserverRepo{
		CloneStatus:     types.CloneStatusCloning,
		CloningProgress: "cloning progress from the db",
	}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
	db.GitserverReposFunc.SetDefaultReturn(gitserverRepos)

	backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{
			ID:        4752134,
			Name:      "repo-name",
			CreatedAt: time.Now(),
			Sources:   map[string]*types.SourceInfo{"1": {}},
		}, nil
	}
	t.Cleanup(func() {
		backend.Mocks = backend.MockServices{}
	})

	ctx := featureflag.WithFlags(context.Background(), db.FeatureFlags())

	RunTest(t, &Test{
		Context: ctx,
		Schema:  mustParseGraphQLSchemaWithClient(t, db, &fakeGitserverClient{}),
		Query: `
			{
				repository(name: "my/repo") {
					mirrorInfo {
						cloneProgress
					}
				}
			}
		`,
		ExpectedResult: `
			{
				"repository": {
					"mirrorInfo": {
						"cloneProgress": "cloning progress from the db"
					}
				}
			}
		`,
	})
}
