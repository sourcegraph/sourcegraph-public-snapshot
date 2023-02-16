package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func New() graphqlbackend.OwnResolver {
	return &resolver{}
}

type resolver struct{}

func (r *resolver) GitBlobOwnership(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, args graphqlbackend.ListOwnershipArgs) (graphqlbackend.OwnershipConnectionResolver, error) {
	return &testingConnectionResolver{blob}, nil
}

type testingConnectionResolver struct {
	blob *graphqlbackend.GitTreeEntryResolver
}

func (r *testingConnectionResolver) TotalCount() int32 {
	return 47
}
func (r *testingConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}
func (r *testingConnectionResolver) Nodes() []graphqlbackend.OwnershipResolver {
	return []graphqlbackend.OwnershipResolver{&ownershipResolver{r.blob}}
}

type ownershipResolver struct {
	blob *graphqlbackend.GitTreeEntryResolver
}

func (r *ownershipResolver) Owner() graphqlbackend.OwnerResolver { return &ownerResolver{} }
func (r *ownershipResolver) Reasons() []graphqlbackend.OwnershipReasonResolver {
	return []graphqlbackend.OwnershipReasonResolver{&codeownersFileEntryResolver{}}
}

type ownerResolver struct{}

func (r *ownerResolver) OwnerField() string { return "owner" }
func (r *ownerResolver) ToPerson() (*graphqlbackend.PersonResolver, bool) {
	return graphqlbackend.NewPersonResolver(nil, "Foo Bar", "foo.bar@sourcegraph.com", true), true
}
func (r *ownerResolver) ToUser() (*graphqlbackend.UserResolver, bool) { return nil, false }
func (r *ownerResolver) ToTeam() (*graphqlbackend.TeamResolver, bool) { return nil, false }

type codeownersFileEntryResolver struct{}

func (r *codeownersFileEntryResolver) ToCodeownersFileEntry() (graphqlbackend.CodeownersFileEntryResolver, bool) {
	return r, true
}
func (r *codeownersFileEntryResolver) Title() string                               { return "CodeOwners" }
func (r *codeownersFileEntryResolver) Description() string                         { return "Dummy code owners file entry." }
func (r *codeownersFileEntryResolver) CodeownersFile() graphqlbackend.FileResolver { return nil }
func (r *codeownersFileEntryResolver) RuleLineMatch() int32                        { return 476 }

func (r *resolver) PersonOwnerField(person *graphqlbackend.PersonResolver) string {
	return "owner"
}
func (r *resolver) UserOwnerField(user *graphqlbackend.UserResolver) string {
	return "owner"
}
func (r *resolver) TeamOwnerField(team *graphqlbackend.TeamResolver) string {
	return "owner"
}

func (r *resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error){}
}
