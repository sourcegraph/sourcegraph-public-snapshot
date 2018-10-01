package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

func TestCheckMirrorRepositoryConnection(t *testing.T) {
	resetMocks()

	const repoURI = "my/repo"

	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}

	t.Run("repository arg", func(t *testing.T) {
		backend.Mocks.Repos.Get = func(ctx context.Context, repoID api.RepoID) (*types.Repo, error) {
			return &types.Repo{URI: repoURI}, nil
		}

		calledRepoLookup := false
		repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			calledRepoLookup = true
			if args.Repo != repoURI {
				t.Errorf("got %q, want %q", args.Repo, repoURI)
			}
			return &protocol.RepoLookupResult{
				Repo: &protocol.RepoInfo{URI: repoURI, VCS: protocol.VCSInfo{URL: "http://example.com/my/repo"}},
			}, nil
		}
		defer func() { repoupdater.MockRepoLookup = nil }()

		calledIsRepoCloneable := false
		gitserver.MockIsRepoCloneable = func(repo gitserver.Repo) error {
			calledIsRepoCloneable = true
			if want := (gitserver.Repo{Name: repoURI, URL: "http://example.com/my/repo"}); !reflect.DeepEqual(repo, want) {
				t.Errorf("got %+v, want %+v", repo, want)
			}
			return nil
		}
		defer func() { gitserver.MockIsRepoCloneable = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
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

		if !calledRepoLookup {
			t.Error("!calledRepoLookup")
		}
		if !calledIsRepoCloneable {
			t.Error("!calledIsRepoCloneable")
		}
	})

	t.Run("name arg", func(t *testing.T) {
		backend.Mocks.Repos.GetByURI = func(ctx context.Context, uri api.RepoURI) (*types.Repo, error) {
			t.Fatal("want GetByURI to not be called")
			return nil, nil
		}

		calledRepoLookup := false
		repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			calledRepoLookup = true
			if args.Repo != repoURI {
				t.Errorf("got %q, want %q", args.Repo, repoURI)
			}
			return &protocol.RepoLookupResult{
				Repo: &protocol.RepoInfo{URI: repoURI, VCS: protocol.VCSInfo{URL: "http://example.com/my/repo"}},
			}, nil
		}
		defer func() { repoupdater.MockRepoLookup = nil }()

		calledIsRepoCloneable := false
		gitserver.MockIsRepoCloneable = func(repo gitserver.Repo) error {
			calledIsRepoCloneable = true
			if want := (gitserver.Repo{Name: repoURI, URL: "http://example.com/my/repo"}); !reflect.DeepEqual(repo, want) {
				t.Errorf("got %+v, want %+v", repo, want)
			}
			return nil
		}
		defer func() { gitserver.MockIsRepoCloneable = nil }()

		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
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

		if !calledRepoLookup {
			t.Error("!calledRepoLookup")
		}
		if !calledIsRepoCloneable {
			t.Error("!calledIsRepoCloneable")
		}
	})
}
