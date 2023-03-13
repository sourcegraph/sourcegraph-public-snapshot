// Ownership resolvers are currently just returning fake data to support development.
// The actual resolver implementation is landing with #46592.
package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func New(db database.DB, gitserver gitserver.Client, ownService own.Service) graphqlbackend.OwnResolver {
	return &ownResolver{
		db:         edb.NewEnterpriseDB(db),
		gitserver:  gitserver,
		ownService: ownService,
	}
}

var (
	_ graphqlbackend.OwnResolver = &ownResolver{}
)

// ownResolver is a dummy graphqlbackend.OwnResolver that returns a single owner
// that is the author of currently viewed commit, and fake ownership reason
// pointing at line 42 of the CODEOWNERS file.
type ownResolver struct {
	db         edb.EnterpriseDB
	gitserver  gitserver.Client
	ownService own.Service
}

func (r *ownResolver) GitBlobOwnership(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, args graphqlbackend.ListOwnershipArgs) (graphqlbackend.OwnershipConnectionResolver, error) {
	repoName := blob.Repository().RepoName()
	commitID := api.CommitID(blob.Commit().OID())
	rs, err := r.ownService.RulesetForRepo(ctx, repoName, commitID)
	if err != nil {
		return nil, err
	}
	// No data found.
	if rs == nil {
		return &ownershipConnectionResolver{db: r.db}, nil
	}

	owners := rs.FindOwners(blob.Path())
	resolvedOwners, err := r.ownService.ResolveOwnersWithType(ctx, owners)
	if err != nil {
		return nil, err
	}
	return &ownershipConnectionResolver{r.db, resolvedOwners}, nil
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
	return map[string]graphqlbackend.NodeByIDFunc{}
}

// ownershipConnectionResolver is a fake graphqlbackend.OwnershipConnectionResolver
// connection with a single dummy item.
type ownershipConnectionResolver struct {
	db             database.DB
	resolvedOwners []codeowners.ResolvedOwner
}

func (r *ownershipConnectionResolver) TotalCount(_ context.Context) (int32, error) {
	return int32(len(r.resolvedOwners)), nil
}

func (r *ownershipConnectionResolver) PageInfo(_ context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (r *ownershipConnectionResolver) Nodes(_ context.Context) ([]graphqlbackend.OwnershipResolver, error) {
	var resolvers []graphqlbackend.OwnershipResolver
	for _, resolvedOwner := range r.resolvedOwners {
		resolvers = append(resolvers, &ownershipResolver{
			db:            r.db,
			resolvedOwner: resolvedOwner,
		})
	}
	return resolvers, nil
}

// ownershipResolver provides a dummy implementation of graphqlbackend.OwnershipResolver
// which just claims the author of given GitTreeEntryResolver Commit is the owner
// and is supports it by pointing at line 42 of the CODEOWNERS file.
type ownershipResolver struct {
	db            database.DB
	resolvedOwner codeowners.ResolvedOwner
}

func (r *ownershipResolver) Owner(ctx context.Context) (graphqlbackend.OwnerResolver, error) {
	return &ownerResolver{
		db:            r.db,
		resolvedOwner: r.resolvedOwner,
	}, nil
}

func (r *ownershipResolver) Reasons(_ context.Context) ([]graphqlbackend.OwnershipReasonResolver, error) {
	return []graphqlbackend.OwnershipReasonResolver{&codeownersFileEntryResolver{}}, nil
}

type ownerResolver struct {
	db            database.DB
	resolvedOwner codeowners.ResolvedOwner
}

func (r *ownerResolver) OwnerField(_ context.Context) (string, error) { return "owner", nil }

func (r *ownerResolver) ToPerson() (*graphqlbackend.PersonResolver, bool) {
	if r.resolvedOwner.Type() != codeowners.OwnerTypePerson {
		return nil, false
	}
	person, ok := r.resolvedOwner.(*codeowners.Person)
	if !ok {
		return nil, false
	}
	includeUserInfo := true
	return graphqlbackend.NewPersonResolver(r.db, person.Handle, person.GetEmail(), includeUserInfo), true
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
	return "Owner is associated with a rule in code owners file.", nil
}

func (r *codeownersFileEntryResolver) CodeownersFile(_ context.Context) (graphqlbackend.FileResolver, error) {
	return nil, nil
}

func (r *codeownersFileEntryResolver) RuleLineMatch(_ context.Context) (int32, error) { return 42, nil }
