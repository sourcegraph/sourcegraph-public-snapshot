package graphqlbackend

import (
	"context"
	"fmt"
	"sort"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/graph-gophers/graphql-go"
	"github.com/hexops/autogold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log/logtest"
)

const exampleCommitSHA1 = "1234567890123456789012345678901234567890"

func TestRepository_Commit(t *testing.T) {
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		assert.Equal(t, api.RepoID(2), repo.ID)
		assert.Equal(t, "abc", rev)
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &gitdomain.Commit{ID: exampleCommitSHA1})
	defer func() {
		backend.Mocks = backend.MockServices{}
	}()

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)

	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						commit(rev: "abc") {
							oid
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"commit": {
							"oid": "` + exampleCommitSHA1 + `"
						}
					}
				}
			`,
		},
	})
}

func TestRepositoryHydration(t *testing.T) {
	t.Parallel()

	makeRepos := func() (*types.Repo, *types.Repo) {
		const id = 42
		name := fmt.Sprintf("repo-%d", id)

		minimal := types.Repo{
			ID:   api.RepoID(id),
			Name: api.RepoName(name),
		}

		hydrated := minimal
		hydrated.ExternalRepo = api.ExternalRepoSpec{
			ID:          name,
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		}
		hydrated.URI = fmt.Sprintf("github.com/foobar/%s", name)
		hydrated.Description = "This is a description of a repository"
		hydrated.Fork = false

		return &minimal, &hydrated
	}

	ctx := context.Background()

	t.Run("hydrated without errors", func(t *testing.T) {
		minimalRepo, hydratedRepo := makeRepos()

		rs := database.NewMockRepoStore()
		rs.GetFunc.SetDefaultReturn(hydratedRepo, nil)
		db := database.NewMockDB()
		db.ReposFunc.SetDefaultReturn(rs)

		repoResolver := NewRepositoryResolver(db, minimalRepo)
		assertRepoResolverHydrated(ctx, t, repoResolver, hydratedRepo)
		mockrequire.CalledOnce(t, rs.GetFunc)
	})

	t.Run("hydration results in errors", func(t *testing.T) {
		minimalRepo, _ := makeRepos()

		dbErr := errors.New("cannot load repo")

		rs := database.NewMockRepoStore()
		rs.GetFunc.SetDefaultReturn(nil, dbErr)
		db := database.NewMockDB()
		db.ReposFunc.SetDefaultReturn(rs)

		repoResolver := NewRepositoryResolver(db, minimalRepo)
		_, err := repoResolver.Description(ctx)
		require.ErrorIs(t, err, dbErr)

		// Another call to make sure err does not disappear
		_, err = repoResolver.URI(ctx)
		require.ErrorIs(t, err, dbErr)

		mockrequire.CalledOnce(t, rs.GetFunc)
	})
}

func assertRepoResolverHydrated(ctx context.Context, t *testing.T, r *RepositoryResolver, hydrated *types.Repo) {
	t.Helper()

	description, err := r.Description(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if description != hydrated.Description {
		t.Fatalf("wrong Description. want=%q, have=%q", hydrated.Description, description)
	}

	uri, err := r.URI(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if uri != hydrated.URI {
		t.Fatalf("wrong URI. want=%q, have=%q", hydrated.URI, uri)
	}
}

func TestRepositoryLabel(t *testing.T) {
	test := func(name string) string {
		r := &RepositoryResolver{
			logger: logtest.Scoped(t),
			RepoMatch: result.RepoMatch{
				Name: api.RepoName(name),
				ID:   api.RepoID(0),
			},
		}
		result, _ := r.Label()
		return result.HTML()
	}

	autogold.Want("encodes spaces for URL in HTML", `<p><a href="/repo%20with%20spaces" rel="nofollow">repo with spaces</a></p>
`).Equal(t, test("repo with spaces"))
}

func TestRepository_DefaultBranch(t *testing.T) {
	ctx := context.Background()
	ts := []struct {
		name                    string
		getDefaultBranchRefName string
		getDefaultBranchErr     error
		wantBranch              *GitRefResolver
		wantErr                 error
	}{
		{
			name:                    "ref exists",
			getDefaultBranchRefName: "refs/heads/main",
			wantBranch:              &GitRefResolver{name: "refs/heads/main"},
		},
		{
			// When clone is in progress GetDefaultBranch returns "", nil
			name: "clone in progress",
			// Expect it to not fail and not return a resolver.
			wantBranch: nil,
			wantErr:    nil,
		},
		{
			name:                "symbolic ref fails",
			getDefaultBranchErr: errors.New("bad git error"),
			wantErr:             errors.New("bad git error"),
		},
	}
	for _, tt := range ts {
		t.Run(tt.name, func(t *testing.T) {
			gitserver.Mocks.GetDefaultBranch = func(repo api.RepoName) (refName string, commit api.CommitID, err error) {
				return tt.getDefaultBranchRefName, "", tt.getDefaultBranchErr
			}
			t.Cleanup(func() {
				gitserver.Mocks.ResolveRevision = nil
			})

			res := &RepositoryResolver{RepoMatch: result.RepoMatch{Name: "repo"}, logger: logtest.Scoped(t)}
			branch, err := res.DefaultBranch(ctx)
			if tt.wantErr != nil && err != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Fatalf("incorrect error message, want=%q have=%q", tt.wantErr.Error(), err.Error())
				}
			} else if tt.wantErr != err {
				t.Fatalf("incorrect error, want=%v have=%v", tt.wantErr, err)
			}
			if branch == nil && tt.wantBranch != nil {
				t.Fatal("invalid nil resolver returned")
			}
			if branch != nil && tt.wantBranch == nil {
				t.Fatalf("expected nil resolver but got %q", branch.name)
			}
			if tt.wantBranch != nil && branch.name != tt.wantBranch.name {
				t.Fatalf("wrong resolver returned, want=%q have=%q", branch.name, tt.wantBranch.name)
			}
		})
	}
}

func TestRepository_KVPs(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewMockDBFrom(database.NewDB(logger, dbtest.NewDB(logger, t)))
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	db.UsersFunc.SetDefaultReturn(users)

	err := db.Repos().Create(ctx, &types.Repo{
		Name: "testrepo",
	})
	require.NoError(t, err)
	repo, err := db.Repos().GetByName(ctx, "testrepo")
	require.NoError(t, err)

	schema := newSchemaResolver(db)
	gqlID := MarshalRepositoryID(repo.ID)

	strPtr := func(s string) *string { return &s }

	t.Run("add", func(t *testing.T) {
		_, err = schema.AddRepoKeyValuePair(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: strPtr("val1"),
		})
		require.NoError(t, err)

		_, err = schema.AddRepoKeyValuePair(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "tag1",
			Value: nil,
		})
		require.NoError(t, err)

		repoResolver, err := schema.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.KeyValuePairs(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Equal(t, []KeyValuePair{{
			key:   "key1",
			value: strPtr("val1"),
		}, {
			key:   "tag1",
			value: nil,
		}}, kvps)
	})

	t.Run("update", func(t *testing.T) {
		_, err = schema.UpdateRepoKeyValuePair(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: strPtr("val2"),
		})
		require.NoError(t, err)

		_, err = schema.UpdateRepoKeyValuePair(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "tag1",
			Value: strPtr("val3"),
		})
		require.NoError(t, err)

		repoResolver, err := schema.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.KeyValuePairs(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Equal(t, []KeyValuePair{{
			key:   "key1",
			value: strPtr("val2"),
		}, {
			key:   "tag1",
			value: strPtr("val3"),
		}}, kvps)
	})

	t.Run("delete", func(t *testing.T) {
		_, err = schema.DeleteRepoKeyValuePair(ctx, struct {
			Repo graphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "key1",
		})
		require.NoError(t, err)

		_, err = schema.DeleteRepoKeyValuePair(ctx, struct {
			Repo graphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "tag1",
		})
		require.NoError(t, err)

		repoResolver, err := schema.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.KeyValuePairs(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Empty(t, kvps)
	})
}
