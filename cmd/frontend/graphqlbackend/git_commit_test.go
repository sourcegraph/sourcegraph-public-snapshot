package graphqlbackend

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCommitResolver(t *testing.T) {
	ctx := context.Background()

	xyzRepo := &types.Repo{ID: 2, Name: "xyz"}
	npmPkgRepo := &types.Repo{ID: 3, Name: "npm/pkg"}
	bobRepo := &types.Repo{ID: 7, Name: "bob-repo"}
	repoMap := map[api.RepoID]*types.Repo{xyzRepo.ID: xyzRepo, npmPkgRepo.ID: npmPkgRepo, bobRepo.ID: bobRepo}

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		repo, found := repoMap[id]
		if !found {
			return nil, errors.Errorf("failed to find repo")
		}
		return repo, nil
	})
	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)

	commit := &gitdomain.Commit{
		ID:      "c1",
		Message: "subject: Changes things\nBody of changes",
		Parents: []api.CommitID{"p1", "p2"},
		Author: gitdomain.Signature{
			Name:  "Bob",
			Email: "bob@alice.com",
			Date:  time.Now(),
		},
		Committer: &gitdomain.Signature{
			Name:  "Alice",
			Email: "alice@bob.com",
			Date:  time.Now(),
		},
		Tags: []string{"v1.0", "v2.0"},
	}

	url := func(t *testing.T, res *GitCommitResolver) string {
		url, err := res.URL(ctx)
		require.NoError(t, err)
		return url
	}
	canonicalURL := func(t *testing.T, res *GitCommitResolver) string {
		url, err := res.CanonicalURL(ctx)
		require.NoError(t, err)
		return url
	}

	t.Run("URLs", func(t *testing.T) {
		repoResolver := NewRepositoryResolver(db, xyzRepo)
		commitResolver := NewGitCommitResolver(db, repoResolver, "c1", commit)
		{
			require.Equal(t, "/xyz/-/commit/c1", url(t, commitResolver))
			require.Equal(t, "/xyz/-/commit/c1", canonicalURL(t, commitResolver))

			inputRev := "master^1"
			commitResolver.inputRev = &inputRev
			require.Equal(t, "/xyz/-/commit/master%5E1", url(t, commitResolver))
			require.Equal(t, "/xyz/-/commit/c1", canonicalURL(t, commitResolver))

			treeResolver := NewGitTreeEntryResolver(db, commitResolver, CreateFileInfo("a/b", false))
			url, err := treeResolver.URL(ctx)
			require.Nil(t, err)
			require.Equal(t, "/xyz@master%5E1/-/blob/a/b", url)
		}
		{
			inputRev := "refs/heads/main"
			commitResolver.inputRev = &inputRev
			require.Equal(t, "/xyz/-/commit/refs/heads/main", url(t, commitResolver))
		}
	})

	t.Run("Tags", func(t *testing.T) {
		repoResolver := NewRepositoryResolver(db, npmPkgRepo)
		commitResolver := NewGitCommitResolver(db, repoResolver, "c1", commit)

		require.Equal(t, "/npm/pkg/-/commit/v2.0", url(t, commitResolver))
		commitish, err := commitResolver.PreferredCanonicalCommitish(ctx)
		require.NoError(t, err)
		require.Equal(t, "v2.0", commitish)

		inputRev := "master"
		commitResolver.inputRev = &inputRev

		require.Equal(t, "/npm/pkg/-/commit/master", url(t, commitResolver))
		commitish, err = commitResolver.PreferredCanonicalCommitish(ctx)
		require.NoError(t, err)
		require.Equal(t, "v2.0", commitish)
		require.Equal(t, "/npm/pkg/-/commit/v2.0", canonicalURL(t, commitResolver))
	})

	t.Run("Lazy loading", func(t *testing.T) {
		git.Mocks.GetCommit = func(api.CommitID) (*gitdomain.Commit, error) {
			return commit, nil
		}
		t.Cleanup(func() {
			git.Mocks.GetCommit = nil
		})

		for _, tc := range []struct {
			name string
			want interface{}
			have func(*GitCommitResolver) (interface{}, error)
		}{{
			name: "author",
			want: toSignatureResolver(db, &commit.Author, true),
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Author(ctx)
			},
		}, {
			name: "committer",
			want: toSignatureResolver(db, commit.Committer, true),
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Committer(ctx)
			},
		}, {
			name: "message",
			want: string(commit.Message),
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Message(ctx)
			},
		}, {
			name: "subject",
			want: "subject: Changes things",
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Subject(ctx)
			},
		}, {
			name: "body",
			want: "Body of changes",
			have: func(r *GitCommitResolver) (interface{}, error) {
				s, err := r.Body(ctx)
				return *s, err
			},
		}, {
			name: "url",
			want: "/bob-repo/-/commit/c1",
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.URL(ctx)
			},
		}, {
			name: "canonical-url",
			want: "/bob-repo/-/commit/c1",
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.CanonicalURL(ctx)
			},
		}} {
			t.Run(tc.name, func(t *testing.T) {
				repo := NewRepositoryResolver(db, bobRepo)
				// We pass no commit here to test that it gets lazy loaded via
				// the git.GetCommit mock above.
				r := NewGitCommitResolver(db, repo, "c1", nil)

				have, err := tc.have(r)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(have, tc.want) {
					t.Errorf("\nhave: %s\nwant: %s", spew.Sprint(have), spew.Sprint(tc.want))
				}
			})
		}
	})
}

func TestGitCommitFileNames(t *testing.T) {
	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(nil, nil)

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)

	db := database.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(repos)

	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		assert.Equal(t, api.RepoID(2), repo.ID)
		assert.Equal(t, exampleCommitSHA1, rev)
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &gitdomain.Commit{ID: exampleCommitSHA1})
	git.Mocks.LsFiles = func(repo api.RepoName, commit api.CommitID) ([]string, error) {
		return []string{"a", "b"}, nil
	}
	defer func() {
		backend.Mocks = backend.MockServices{}
		git.ResetMocks()
	}()

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						commit(rev: "` + exampleCommitSHA1 + `") {
							fileNames
						}
					}
				}
			`,
			ExpectedResult: `
{
  "repository": {
    "commit": {
		"fileNames": ["a", "b"]
    }
  }
}
			`,
		},
	})
}
