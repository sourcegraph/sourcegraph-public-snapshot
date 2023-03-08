// Ownership resolvers are currently just returning fake data to support development.
// The actual resolver implementation is landing with #46592.
package resolvers

import (
	"context"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
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

func ownerText(o *codeownerspb.Owner) string {
	if o == nil {
		return ""
	}
	if o.Handle != "" {
		return o.Handle
	}
	return o.Email
}

func (r *ownResolver) GitBlobOwnership(
	ctx context.Context,
	blob *graphqlbackend.GitTreeEntryResolver,
	args graphqlbackend.ListOwnershipArgs,
) (graphqlbackend.OwnershipConnectionResolver, error) {
	if err := areOwnEndpointsAvailable(ctx); err != nil {
		return nil, err
	}
	cursor, err := graphqlutil.DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}
	repo := blob.Repository()
	repoID, repoName := repo.IDInt32(), repo.RepoName()
	commitID := api.CommitID(blob.Commit().OID())
	rs, err := r.ownService.RulesetForRepo(ctx, repoName, repoID, commitID)
	if err != nil {
		return nil, err
	}
	// No data found.
	if rs == nil {
		return &ownershipConnectionResolver{db: r.db}, nil
	}
	owners := rs.FindOwners(blob.Path())
	sort.Slice(owners, func(i, j int) bool {
		iText := ownerText(owners[i])
		jText := ownerText(owners[j])
		return iText < jText
	})
	total := len(owners)
	for cursor != "" && len(owners) > 0 && ownerText(owners[0]) != cursor {
		owners = owners[1:]
	}
	var next *string
	if args.First != nil && len(owners) > int(*args.First) {
		cursor := ownerText(owners[*args.First])
		next = &cursor
		owners = owners[:*args.First]
	}
	resolvedOwners, err := r.ownService.ResolveOwnersWithType(ctx, owners)
	if err != nil {
		return nil, err
	}
	return &ownershipConnectionResolver{
		db:             r.db,
		total:          total,
		next:           next,
		resolvedOwners: resolvedOwners,
	}, nil
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
	return map[string]graphqlbackend.NodeByIDFunc{
		codeownersIngestedFileKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			// codeowners ingested files are identified by repo ID at the moment.
			var repoID api.RepoID
			if err := relay.UnmarshalSpec(id, &repoID); err != nil {
				return nil, errors.Wrap(err, "could not unmarshal repository ID")
			}
			return r.RepoIngestedCodeowners(ctx, repoID)
		},
	}
}

// ownershipConnectionResolver is a fake graphqlbackend.OwnershipConnectionResolver
// connection with a single dummy item.
type ownershipConnectionResolver struct {
	db             database.DB
	total          int
	next           *string
	resolvedOwners []codeowners.ResolvedOwner
}

func (r *ownershipConnectionResolver) TotalCount(_ context.Context) (int32, error) {
	return int32(r.total), nil
}

func (r *ownershipConnectionResolver) PageInfo(_ context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.EncodeCursor(r.next), nil
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
	if err := areOwnEndpointsAvailable(ctx); err != nil {
		return nil, err
	}
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
	return graphqlbackend.NewPersonResolver(r.db, person.Handle, person.Email, includeUserInfo), true
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

func areOwnEndpointsAvailable(ctx context.Context) error {
	if !featureflag.FromContext(ctx).GetBoolOr("search-ownership", false) {
		return errors.New("own is not available yet")
	}
	return nil
}
