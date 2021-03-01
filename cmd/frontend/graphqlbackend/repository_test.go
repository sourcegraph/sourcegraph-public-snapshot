package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

const exampleCommitSHA1 = "1234567890123456789012345678901234567890"

func TestRepository_Commit(t *testing.T) {
	resetMocks()
	database.Mocks.Repos.MockGetByName(t, "github.com/gorilla/mux", 2)
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		if repo.ID != 2 || rev != "abc" {
			t.Error("wrong arguments to ResolveRev")
		}
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &git.Commit{ID: exampleCommitSHA1})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t),
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
	db := new(dbtesting.MockDB)
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
		database.Mocks.Repos.Get = func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
			return hydratedRepo, nil
		}
		defer func() { database.Mocks = database.MockStores{} }()

		repoResolver := NewRepositoryResolver(db, minimalRepo)
		assertRepoResolverHydrated(ctx, t, repoResolver, hydratedRepo)
	})

	t.Run("hydration results in errors", func(t *testing.T) {
		minimalRepo, _ := makeRepos()

		dbErr := errors.New("cannot load repo")

		database.Mocks.Repos.Get = func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
			return nil, dbErr
		}
		defer func() { database.Mocks = database.MockStores{} }()

		repoResolver := NewRepositoryResolver(db, minimalRepo)
		_, err := repoResolver.Description(ctx)
		if err == nil {
			t.Fatal("err is unexpected nil")
		}

		if err != dbErr {
			t.Fatalf("wrong err. want=%q, have=%q", dbErr, err)
		}

		// Another call to make sure err does not disappear
		_, err = repoResolver.URI(ctx)
		if err == nil {
			t.Fatal("err is unexpected nil")
		}

		if err != dbErr {
			t.Fatalf("wrong err. want=%q, have=%q", dbErr, err)
		}
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
			name: api.RepoName(name),
			id:   api.RepoID(0),
		}
		result, _ := r.Label()
		return result.HTML()
	}

	autogold.Want("encodes spaces for URL in HTML", `<p><a href="/repo%20with%20spaces" rel="nofollow">repo with spaces</a></p>
`).Equal(t, test("repo with spaces"))
}
