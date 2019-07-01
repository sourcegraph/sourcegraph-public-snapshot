package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

const exampleCommitSHA1 = "1234567890123456789012345678901234567890"

func TestRepository_Commit(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockGetByName(t, "github.com/gorilla/mux", 2)
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo types.RepoIdentifier, rev string) (api.CommitID, error) {
		if repo.RepoID() != 2 || rev != "abc" {
			t.Error("wrong arguments to ResolveRev")
		}
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &git.Commit{ID: exampleCommitSHA1})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
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
	id := 42
	name := fmt.Sprintf("repo-%d", id)

	minimalRepo := &db.MinimalRepo{
		ID:   api.RepoID(id),
		Name: api.RepoName(name),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          name,
			ServiceType: "github",
			ServiceID:   "https://github.com",
		},
	}
	hydratedRepo := &types.Repo{
		ID:           minimalRepo.ID,
		ExternalRepo: &(minimalRepo.ExternalRepo),
		Name:         minimalRepo.Name,
		URI:          fmt.Sprintf("github.com/foobar/%s", name),
		Description:  "This is a description of a repository",
		Language:     "monkey",
		Fork:         false,
	}

	db.Mocks.Repos.Get = func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
		return hydratedRepo, nil
	}

	defer func() { db.Mocks = db.MockStores{} }()

	ctx := context.Background()

	repoResolver := &repositoryResolver{repo: minimalRepo}
	if have, want := repoResolver.Description(ctx), hydratedRepo.Description; have != want {
		t.Fatalf("wrong Description. want=%q, have=%q", want, have)
	}
	if have, want := repoResolver.URI(ctx), hydratedRepo.URI; have != want {
		t.Fatalf("wrong URI. want=%q, have=%q", want, have)
	}
	if have, want := repoResolver.Language(ctx), hydratedRepo.Language; have != want {
		t.Fatalf("wrong Language. want=%q, have=%q", want, have)
	}
}
