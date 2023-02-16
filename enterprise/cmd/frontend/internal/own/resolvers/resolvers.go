// Ownership resolvers are currently just returning fake data to support development.
// The actual resolver implementation is landing with #46592.
package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func New() *ownResolver {
	return &ownResolver{}
}

var (
	_ graphqlbackend.OwnResolver = &ownResolver{}
)

// ownResolver is a dummy graphqlbackend.OwnResolver that reutns a single owner
// that is the author of currently viewed commit, and fake ownership reason
// pointing at line 42 of the CODEOWNERS file.
type ownResolver struct{}

func (r *ownResolver) GitBlobOwnership(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, args graphqlbackend.ListOwnershipArgs) (graphqlbackend.OwnershipConnectionResolver, error) {
	return &ownershipConnectionResolver{blob}, nil
}
func (r *ownResolver) PersonOwnerField(person *graphqlbackend.PersonResolver) string {
	return "owner"
}
func (r *ownResolver) UserOwnerField(user *graphqlbackend.UserResolver) string {
	return "owner"
}
func (r *ownResolver) TeamOwnerField(team *graphqlbackend.TeamResolver) string {
	return "owner"
}
func (r *ownResolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error){}
}

// ownershipConnectionResolver is a fake graphqlbackend.OwnershipConnectionResolver
// connection with a single dummy item.
type ownershipConnectionResolver struct {
	blob *graphqlbackend.GitTreeEntryResolver
}

func (r *ownershipConnectionResolver) TotalCount(_ context.Context) (int32, error) {
	return 1, nil
}
func (r *ownershipConnectionResolver) PageInfo(_ context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
func (r *ownershipConnectionResolver) Nodes(_ context.Context) ([]graphqlbackend.OwnershipResolver, error) {
	return []graphqlbackend.OwnershipResolver{&ownershipResolver{r.blob}}, nil
}

// ownershipResolver provides a dummy implementation of graphqlbackend.OwnershipResolver
// which just claims the the auhthor of given GitTreeEntryResolver Commit is the owner
// and is supports it by pointing at line 42 of the CODEOWNERS file.
type ownershipResolver struct {
	blob *graphqlbackend.GitTreeEntryResolver
}

func (r *ownershipResolver) Owner(ctx context.Context) (graphqlbackend.OwnerResolver, error) {
	return &ownerResolver{ctx, r.blob}, nil
}
func (r *ownershipResolver) Reasons(_ context.Context) ([]graphqlbackend.OwnershipReasonResolver, error) {
	return []graphqlbackend.OwnershipReasonResolver{&codeownersFileEntryResolver{}}, nil
}

type ownerResolver struct {
	ctx  context.Context
	blob *graphqlbackend.GitTreeEntryResolver
}

func (r *ownerResolver) OwnerField(_ context.Context) (string, error) { return "owner", nil }

// ToPerson is just a dummy implementation that returns the author of the viewed commit
// at this point. This is to aid UI development and will be replaced with actual implementation
// with landing #46592.
func (r *ownerResolver) ToPerson() (*graphqlbackend.PersonResolver, bool) {
	author, err := r.blob.Commit().Author(r.ctx)
	if err != nil {
		return nil, false
	}
	return author.Person(), true
}
func (r *ownerResolver) ToUser() (*graphqlbackend.UserResolver, bool) {
	return nil, false
}
func (r *ownerResolver) ToTeam() (*graphqlbackend.TeamResolver, bool) {
	return nil, false
}

type codeownersFileEntryResolver struct{}

func (r *codeownersFileEntryResolver) ToCodeownersFileEntry() (graphqlbackend.CodeownersFileEntryResolver, bool) {
	return r, true
}
func (r *codeownersFileEntryResolver) Title(_ context.Context) (string, error) {
	return "CodeOwners", nil
}
func (r *codeownersFileEntryResolver) Description(_ context.Context) (string, error) {
	return "Dummy code owners file entry.", nil
}
func (r *codeownersFileEntryResolver) CodeownersFile(_ context.Context) (graphqlbackend.FileResolver, error) {
	return nil, nil
}
func (r *codeownersFileEntryResolver) RuleLineMatch(_ context.Context) (int32, error) { return 42, nil }
