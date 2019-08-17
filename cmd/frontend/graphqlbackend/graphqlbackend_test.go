package graphqlbackend

import (
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func TestRepository(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockGetByName(t, "github.com/gorilla/mux", 2)
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"name": "github.com/gorilla/mux"
					}
				}
			`,
		},
	})
}

func TestNodeResolverTo(t *testing.T) {
	// This test exists purely to remove some non determinism in our tests
	// run. The To* resolvers are stored in a map in our graphql
	// implementation => the order we call them is non deterministic =>
	// codecov coverage reports are noisy.
	nodes := []Node{
		&accessTokenResolver{},
		&discussionCommentResolver{},
		&discussionThreadResolver{},
		&externalAccountResolver{},
		&externalServiceResolver{},
		&GitRefResolver{},
		&RepositoryResolver{},
		&UserResolver{},
		&OrgResolver{},
		&organizationInvitationResolver{},
		&GitCommitResolver{},
		&savedSearchResolver{},
		&siteResolver{},
	}

	for _, n := range nodes {
		r := &NodeResolver{n}
		if _, b := r.ToAccessToken(); b {
			continue
		}
		if _, b := r.ToDiscussionComment(); b {
			continue
		}
		if _, b := r.ToDiscussionThread(); b {
			continue
		}
		if _, b := r.ToProductLicense(); b {
			continue
		}
		if _, b := r.ToProductSubscription(); b {
			continue
		}
		if _, b := r.ToExternalAccount(); b {
			continue
		}
		if _, b := r.ToExternalService(); b {
			continue
		}
		if _, b := r.ToGitRef(); b {
			continue
		}
		if _, b := r.ToRepository(); b {
			continue
		}
		if _, b := r.ToUser(); b {
			continue
		}
		if _, b := r.ToOrg(); b {
			continue
		}
		if _, b := r.ToOrganizationInvitation(); b {
			continue
		}
		if _, b := r.ToGitCommit(); b {
			continue
		}
		if _, b := r.ToRegistryExtension(); b {
			continue
		}
		if _, b := r.ToSavedSearch(); b {
			continue
		}
		if _, b := r.ToSite(); b {
			continue
		}
		t.Fatalf("unexpected node %#+v", n)
	}
}

func init() {
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
	}
}
