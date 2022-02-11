package graphqlbackend

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCheckMirrorRepositoryConnection(t *testing.T) {
	const repoName = api.RepoName("my/repo")

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	repos := database.NewMockRepoStore()

	db := database.NewMockDB()
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
		defer func() {
			backend.Mocks = backend.MockServices{}
			gitserver.MockIsRepoCloneable = nil
		}()

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

	t.Run("name arg", func(t *testing.T) {
		backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
			t.Fatal("want GetByName to not be called")
			return nil, nil
		}

		calledIsRepoCloneable := false
		gitserver.MockIsRepoCloneable = func(repo api.RepoName) error {
			calledIsRepoCloneable = true
			if want := repoName; !reflect.DeepEqual(repo, want) {
				t.Errorf("got %+v, want %+v", repo, want)
			}
			return nil
		}
		defer func() {
			backend.Mocks = backend.MockServices{}
			gitserver.MockIsRepoCloneable = nil
		}()

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
				mutation {
					checkMirrorRepositoryConnection(name: "my/repo") {
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
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			backend.Mocks.Repos.GetByName = func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
				return &types.Repo{
					Name:      repoName,
					CreatedAt: time.Now(),
					Sources:   map[string]*types.SourceInfo{"1": {CloneURL: tc.repoURL}},
				}, nil
			}
			defer func() {
				backend.Mocks = backend.MockServices{}
			}()

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
