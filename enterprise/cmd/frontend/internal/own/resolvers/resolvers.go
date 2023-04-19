package resolvers

import (
	"context"
	"fmt"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func New(db database.DB, gitserver gitserver.Client, logger log.Logger) graphqlbackend.OwnResolver {
	return &ownResolver{
		db:           edb.NewEnterpriseDB(db),
		gitserver:    gitserver,
		ownServiceFn: func() own.Service { return own.NewService(gitserver, db) },
		logger:       logger,
	}
}

func NewWithService(db database.DB, gitserver gitserver.Client, ownService own.Service, logger log.Logger) graphqlbackend.OwnResolver {
	return &ownResolver{
		db:           edb.NewEnterpriseDB(db),
		gitserver:    gitserver,
		ownServiceFn: func() own.Service { return ownService },
		logger:       logger,
	}
}

var (
	_ graphqlbackend.OwnResolver                      = &ownResolver{}
	_ graphqlbackend.OwnershipReasonResolver          = &ownershipReasonResolver{}
	_ graphqlbackend.GitCommitOwnershipSignalResolver = &gitCommitOwnershipSignalResolver{}
	_ graphqlbackend.SimpleOwnReasonResolver          = &gitCommitOwnershipSignalResolver{}
	_ graphqlbackend.SimpleOwnReasonResolver          = &codeownersFileEntryResolver{}
)

type ownResolver struct {
	db           edb.EnterpriseDB
	gitserver    gitserver.Client
	ownServiceFn func() own.Service
	logger       log.Logger
}

func (r *ownResolver) ownService() own.Service {
	return r.ownServiceFn()
}

type ownershipReasonResolver struct {
	resolver graphqlbackend.SimpleOwnReasonResolver
}

func (o *ownershipReasonResolver) Title() (string, error) {
	if res, ok := o.resolver.(graphqlbackend.SimpleOwnReasonResolver); ok {
		return res.Title()
	}
	return "", errors.New("invalid resolver")
}

func (o *ownershipReasonResolver) Description() (string, error) {
	if res, ok := o.resolver.(graphqlbackend.SimpleOwnReasonResolver); ok {
		return res.Description()
	}
	return "", errors.New("invalid resolver")
}

func (o *ownershipReasonResolver) ToCodeownersFileEntry() (graphqlbackend.CodeownersFileEntryResolver, bool) {
	if res, ok := o.resolver.(*codeownersFileEntryResolver); ok {
		return res, true
	}
	return nil, false
}

func (o *ownershipReasonResolver) ToGitCommitOwnershipSignal() (graphqlbackend.GitCommitOwnershipSignalResolver, bool) {
	if res, ok := o.resolver.(*gitCommitOwnershipSignalResolver); ok {
		return res, true
	}
	return nil, false
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
	ownService := r.ownService()
	rs, err := ownService.RulesetForRepo(ctx, repoName, repoID, commitID)
	if err != nil {
		return nil, err
	}
	// No ruleset found.
	if rs == nil {
		return &ownershipConnectionResolver{db: r.db}, nil
	}
	rule := rs.Match(blob.Path())
	// No match found.
	if rule == nil {
		return &ownershipConnectionResolver{db: r.db}, nil
	}
	owners := rule.GetOwner()
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
	resolvedOwners, err := ownService.ResolveOwnersWithType(ctx, owners)
	if err != nil {
		return nil, err
	}
	ownerships := make([]graphqlbackend.OwnershipResolver, 0, len(resolvedOwners))
	for _, ro := range resolvedOwners {

		res := &codeownersFileEntryResolver{
			db:              r.db,
			gitserverClient: r.gitserver,
			source:          rs.GetSource(),
			repo:            blob.Repository(),
			matchLineNumber: rule.GetLineNumber(),
		}

		reasons := []graphqlbackend.OwnershipReasonResolver{
			&ownershipReasonResolver{res},
			// &ownershipReasonResolver{&gitCommitOwnershipSignalResolver{repo: repoID, rev: commitID.Short(), path: blob.Path(), repoName: repoName}},
		}

		ownerships = append(ownerships, &ownershipResolver{
			db:            r.db,
			resolvedOwner: ro,
			reasons:       reasons,
		})
	}

	name, email, count, err := computeCommitOwner(blob.Path(), commitID.Short(), repoName)
	if err != nil {
		return nil, err
	}

	contribResolver := &ownershipResolver{
		db: r.db,
		resolvedOwner: &codeowners.Person{
			User:         nil,
			PrimaryEmail: nil,
			Handle:       name,
			Email:        email,
		},
		reasons: []graphqlbackend.OwnershipReasonResolver{&ownershipReasonResolver{&gitCommitOwnershipSignalResolver{total: count}}},
	}

	user, err := identifyUser(r.db, email)
	if err == nil {
		// if we don't get an error (meaning we can match) we will add it to the resolver
		contribResolver.resolvedOwner = &codeowners.Person{
			User:   user,
			Handle: name,
			Email:  email,
		}
	}

	ownerships = append(ownerships, contribResolver)

	return &ownershipConnectionResolver{
		db:             r.db,
		total:          total,
		next:           next,
		resolvedOwners: resolvedOwners,
		ownerships:     ownerships,
	}, nil
}

func (r *ownResolver) PersonOwnerField(_ *graphqlbackend.PersonResolver) string {
	return "owner"
}

func (r *ownResolver) UserOwnerField(_ *graphqlbackend.UserResolver) string {
	return "owner"
}

func (r *ownResolver) TeamOwnerField(_ *graphqlbackend.TeamResolver) string {
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

type ownershipConnectionResolver struct {
	db             edb.EnterpriseDB
	total          int
	next           *string
	resolvedOwners []codeowners.ResolvedOwner
	ownerships     []graphqlbackend.OwnershipResolver
}

func (r *ownershipConnectionResolver) TotalCount(_ context.Context) (int32, error) {
	return int32(r.total), nil
}

func (r *ownershipConnectionResolver) PageInfo(_ context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.EncodeCursor(r.next), nil
}

func (r *ownershipConnectionResolver) Nodes(_ context.Context) ([]graphqlbackend.OwnershipResolver, error) {
	return r.ownerships, nil
}

type ownershipResolver struct {
	db            edb.EnterpriseDB
	resolvedOwner codeowners.ResolvedOwner
	reasons       []graphqlbackend.OwnershipReasonResolver
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
	return r.reasons, nil
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
	if r.resolvedOwner.Type() != codeowners.OwnerTypeTeam {
		return nil, false
	}
	resolvedTeam, ok := r.resolvedOwner.(*codeowners.Team)
	if !ok {
		return nil, false
	}
	return graphqlbackend.NewTeamResolver(r.db, resolvedTeam.Team), true
}

type codeownersFileEntryResolver struct {
	db              edb.EnterpriseDB
	source          codeowners.RulesetSource
	matchLineNumber int32
	repo            *graphqlbackend.RepositoryResolver
	gitserverClient gitserver.Client
}

func (r *codeownersFileEntryResolver) Title() (string, error) {
	return "CODEOWNERS", nil
}

func (r *codeownersFileEntryResolver) Description() (string, error) {
	return "Owner is associated with a rule in a CODEOWNERS file.", nil
}

func (r *codeownersFileEntryResolver) CodeownersFile(ctx context.Context) (graphqlbackend.FileResolver, error) {
	switch src := r.source.(type) {
	case codeowners.IngestedRulesetSource:
		// For ingested, create a virtual file resolver that loads the raw contents
		// on demand.
		stat := graphqlbackend.CreateFileInfo("CODEOWNERS", false)
		return graphqlbackend.NewVirtualFileResolver(stat, func(ctx context.Context) (string, error) {
			f, err := r.db.Codeowners().GetCodeownersForRepo(ctx, api.RepoID(src.ID))
			if err != nil {
				return "", err
			}
			return f.Contents, nil
		}, graphqlbackend.VirtualFileResolverOptions{
			URL: fmt.Sprintf("%s/-/own", r.repo.URL()),
		}), nil
	case codeowners.GitRulesetSource:
		// For committed, we can return a GitTreeEntry, as it implements File2.
		c := graphqlbackend.NewGitCommitResolver(r.db, r.gitserverClient, r.repo, src.Commit, nil)
		return c.File(ctx, &struct{ Path string }{Path: src.Path})
	default:
		return nil, errors.New("unknown ownership file source")
	}
}

func (r *codeownersFileEntryResolver) RuleLineMatch(_ context.Context) (int32, error) {
	return r.matchLineNumber, nil
}

func areOwnEndpointsAvailable(ctx context.Context) error {
	if !featureflag.FromContext(ctx).GetBoolOr("search-ownership", false) {
		return errors.New("own is not available yet")
	}
	return nil
}

type gitCommitOwnershipSignalResolver struct {
	total int32
}

func (g *gitCommitOwnershipSignalResolver) Title() (string, error) {
	return "TOP CONTRIBUTOR", nil
}

func (g *gitCommitOwnershipSignalResolver) Description() (string, error) {
	return "Owner is associated because they are the top committer in the file", nil
}

func (g *gitCommitOwnershipSignalResolver) TotalCount() (int32, error) {
	return g.total, nil
}

func (g *gitCommitOwnershipSignalResolver) RecentCount() (int32, error) {
	return 0, nil
}

func computeCommitOwner(path string, rev string, repoName api.RepoName) (name string, email string, total int32, err error) {
	fmt.Println("asdfasdf")
	// commits, err := gitserver.NewClient().Commits(context.Background(), authz.DefaultSubRepoPermsChecker, repoName, gitserver.CommitsOptions{
	// 	Path:  path,
	// 	Range: rev,
	// })
	// if err != nil {
	// 	return "", 0, err
	// }
	//
	// counts := make(map[string]int)
	// for _, commit := range commits {
	// 	counts[commit.Author.Name]++
	// }
	//
	// max := ""
	// maxVal := 0
	// for key, val := range counts {
	// 	if val > maxVal {
	// 		max = key
	// 		maxVal = val
	// 	}
	// }

	contribs, err := gitserver.NewClient().ContributorCount(context.Background(), repoName, gitserver.ContributorOptions{
		Path:  path,
		Range: rev,
	})

	if len(contribs) == 0 {
		return "", "", 0, nil
	}

	var max int
	for i := range contribs {
		if contribs[i].Count > contribs[max].Count {
			max = i
		}
	}

	top := contribs[max]

	return top.Name, top.Email, top.Count, nil
}

func identifyUser(db database.DB, email string) (*types.User, error) {
	return db.Users().GetByVerifiedEmail(context.Background(), email)
}
