// Ownership resolvers are currently just returning fake data to support development.
// The actual resolver implementation is landing with #46592.
package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func New(db database.DB, ownService own.Service) *ownResolver {
	return &ownResolver{
		db:         db,
		ownService: ownService,
	}
}

var (
	_ graphqlbackend.OwnResolver = &ownResolver{}
)

type ownResolver struct {
	db         database.DB
	ownService own.Service
}

func (r *ownResolver) GitBlobOwnership(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, args graphqlbackend.ListOwnershipArgs) (graphqlbackend.OwnershipConnectionResolver, error) {
	repoName := blob.Repository().RepoName()
	commitID := api.CommitID(blob.Commit().OID())
	rs, source, err := r.ownService.RulesetForRepo(ctx, repoName, commitID)
	if err != nil {
		return nil, err
	}

	// No data found.
	if rs == nil {
		return &ownershipConnectionResolver{db: r.db}, nil
	}

	rule := rs.FindMatchingRule(blob.Path())
	if rule == nil {
		return &ownershipConnectionResolver{db: r.db}, nil
	}
	resolvedOwners, err := r.ownService.ResolveOwnersWithType(ctx, rule.GetOwner())
	if err != nil {
		return nil, err
	}

	if args.Reasons != nil && len(*args.Reasons) > 0 {
		return nil, errors.New("filtering by reasons is not yet implemented")
	}

	after, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cursor")
	}

	first := 0
	if args.First != nil {
		first = int(*args.First)
	}

	return &ownershipConnectionResolver{
		db:             r.db,
		resolvedOwners: resolvedOwners,
		source:         source,
		first:          first,
		after:          after,
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
	return map[string]graphqlbackend.NodeByIDFunc{}
}

type ownershipConnectionResolver struct {
	db             database.DB
	resolvedOwners []codeowners.ResolvedOwner
	source         own.RulesetSource
	first          int
	after          int
}

func (r *ownershipConnectionResolver) TotalCount(_ context.Context) (int32, error) {
	return int32(len(r.resolvedOwners)), nil
}

func (r *ownershipConnectionResolver) PageInfo(_ context.Context) (*graphqlutil.PageInfo, error) {
	if r.first == 0 {
		return graphqlutil.HasNextPage(false), nil
	}
	if r.after+r.first >= len(r.resolvedOwners) {
		return graphqlutil.HasNextPage(false), nil
	}
	idx := int32(r.first + r.after)
	return graphqlutil.EncodeIntCursor(&idx), nil
}

func (r *ownershipConnectionResolver) Nodes(_ context.Context) ([]graphqlbackend.OwnershipResolver, error) {
	var resolvers []graphqlbackend.OwnershipResolver
	for i, resolvedOwner := range r.resolvedOwners[r.after:] {
		if r.first != 0 && i == r.first-1 {
			break
		}
		resolvers = append(resolvers, &ownershipResolver{
			db:            r.db,
			resolvedOwner: resolvedOwner,
			source:        r.source,
		})
	}
	return resolvers, nil
}

type ownershipResolver struct {
	db            database.DB
	resolvedOwner codeowners.ResolvedOwner
	source        own.RulesetSource
}

func (r *ownershipResolver) Owner(ctx context.Context) (graphqlbackend.OwnerResolver, error) {
	return &ownerResolver{
		db:            r.db,
		resolvedOwner: r.resolvedOwner,
	}, nil
}

func (r *ownershipResolver) Reasons(_ context.Context) ([]graphqlbackend.OwnershipReasonResolver, error) {
	return []graphqlbackend.OwnershipReasonResolver{&codeownersFileEntryResolver{source: r.source}}, nil
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

type codeownersFileEntryResolver struct {
	source own.RulesetSource
}

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
	s, ok := r.source.(own.RulesetSourceCommittedFile)
	if ok {
		commit := graphqlbackend.NewGitCommitResolver(r.db, r.gc, r.repo, r.commit, nil)
		// commit.inputRev = r.inputRev
		return graphqlbackend.NewGitTreeEntryResolver(r.db, r.gc, graphqlbackend.GitTreeEntryResolverOpts{
			Commit: commit,
			Stat:   graphqlbackend.CreateFileInfo(s.Path, false),
		}), nil
	}
	return nil, nil
}

func (r *codeownersFileEntryResolver) RuleLineMatch(_ context.Context) (int32, error) { return 42, nil }
