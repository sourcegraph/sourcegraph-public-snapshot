package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestPersonEmailIsAnonymized(t *testing.T) {
	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(nil, nil)

	repos := dbmocks.NewMockRepoStore()
	repos.GetByNameFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)

	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo api.RepoName, rev string) (api.CommitID, error) {
		assert.Equal(t, exampleCommitSHA1, rev)
		return exampleCommitSHA1, nil
	}

	t.Cleanup(func() {
		backend.Mocks = backend.MockServices{}
	})

	db := dbmocks.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(repos)

	users := dbmocks.NewMockUserStore()

	author := gitdomain.Signature{
		Email: "nono@sourcegraph.com",
		Name:  "Ñoñó Sourcegraph", // special characters should be stripped from anonymized emails
		Date:  time.Time{},
	}

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetCommitFunc.SetDefaultReturn(&gitdomain.Commit{Author: author, ID: exampleCommitSHA1}, nil)

	testcases := []struct {
		name            string
		currentAuthUser *types.User
		context         context.Context
		dotcom          bool
		expectedEmail   string
	}{{
		name:          "anonymous user cannot see real email",
		context:       actor.WithActor(context.Background(), actor.FromAnonymousUser("anonymous")),
		dotcom:        true,
		expectedEmail: "nonosourcegraph@noreply.sourcegraph.com",
	}, {
		name: "authenticated user can see real email",
		currentAuthUser: &types.User{
			ID:        1,
			Username:  "nono",
			SiteAdmin: false,
		},
		context:       actor.WithActor(context.Background(), actor.FromUser(1)),
		dotcom:        true,
		expectedEmail: "nono@sourcegraph.com",
	}, {
		name:          "works as usual with dotcom disabled",
		dotcom:        false,
		expectedEmail: "nono@sourcegraph.com",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			dotcom.MockSourcegraphDotComMode(t, tc.dotcom)
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(tc.currentAuthUser, nil)

			db.UsersFunc.SetDefaultReturn(users)
			RunTest(t, &Test{
				Context: tc.context,
				Schema:  mustParseGraphQLSchemaWithClient(t, db, gitserverClient),
				Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						commit(rev: "` + exampleCommitSHA1 + `") {
							author {
								person {
									email
								}
							}
						}
					}
				}
			`,
				ExpectedResult: `
				{
				  "repository": {
				    "commit": {
				      "author": {
				        "person": {
				          "email": "` + tc.expectedEmail + `"
				        }
				      }
				    }
				  }
				}
			`,
			})
		})
	}
}
