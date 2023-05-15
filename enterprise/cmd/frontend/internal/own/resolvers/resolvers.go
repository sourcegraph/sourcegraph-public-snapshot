package resolvers

import (
	"context"
	"fmt"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/errcode"

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
	_ graphqlbackend.OwnResolver                              = &ownResolver{}
	_ graphqlbackend.OwnershipReasonResolver                  = &ownershipReasonResolver{}
	_ graphqlbackend.RecentContributorOwnershipSignalResolver = &recentContributorOwnershipSignal{}
	_ graphqlbackend.SimpleOwnReasonResolver                  = &recentContributorOwnershipSignal{}
	_ graphqlbackend.RecentViewOwnershipSignalResolver        = &recentViewOwnershipSignal{}
	_ graphqlbackend.SimpleOwnReasonResolver                  = &recentViewOwnershipSignal{}
	_ graphqlbackend.SimpleOwnReasonResolver                  = &codeownersFileEntryResolver{}
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
	return o.resolver.Title()
}

func (o *ownershipReasonResolver) Description() (string, error) {
	return o.resolver.Description()
}

func (o *ownershipReasonResolver) ToCodeownersFileEntry() (res graphqlbackend.CodeownersFileEntryResolver, ok bool) {
	res, ok = o.resolver.(*codeownersFileEntryResolver)
	return
}

func (o *ownershipReasonResolver) ToRecentContributorOwnershipSignal() (res graphqlbackend.RecentContributorOwnershipSignalResolver, ok bool) {
	res, ok = o.resolver.(*recentContributorOwnershipSignal)
	return
}

func (o *ownershipReasonResolver) ToRecentViewOwnershipSignal() (res graphqlbackend.RecentViewOwnershipSignalResolver, ok bool) {
	res, ok = o.resolver.(*recentViewOwnershipSignal)
	return
}

func (r *ownResolver) GitBlobOwnership(
	ctx context.Context,
	blob *graphqlbackend.GitTreeEntryResolver,
	args graphqlbackend.ListOwnershipArgs,
) (graphqlbackend.OwnershipConnectionResolver, error) {
	if err := areOwnEndpointsAvailable(ctx); err != nil {
		return nil, err
	}

	// Evaluate CODEOWNERS rules.
	ownerships, err := r.computeCodeowners(ctx, blob)
	if err != nil {
		return nil, err
	}

	// Retrieve recent contributors signals.
	repoID := blob.Repository().IDInt32()
	contribResolvers, err := computeRecentContributorSignals(ctx, r.db, blob.Path(), repoID)
	if err != nil {
		return nil, err
	}
	ownerships = append(ownerships, contribResolvers...)

	// Retrieve recent view signals.
	viewerResolvers, err := computeRecentViewSignals(ctx, r.logger, r.db, blob.Path(), repoID)
	if err != nil {
		return nil, err
	}
	ownerships = append(ownerships, viewerResolvers...)

	return r.ownershipConnection(args, ownerships)
}

// repoRootPath is the path that designates all the aggregate signals
// for a repository.
const repoRootPath = ""

// GitCommitOwnership retrieves ownership signals (not CODEOWNERS data)
// aggregated for the whole repository.
//
// It's a commit ownership rather than repo ownership because
// from the resolution point of view repo needs to be versioned
// at a certain commit to compute signals. At this point, however
// signals are not versioned yet, so every commit gets the same data.
func (r *ownResolver) GitCommitOwnership(
	ctx context.Context,
	commit *graphqlbackend.GitCommitResolver,
	args graphqlbackend.ListOwnershipArgs,
) (graphqlbackend.OwnershipConnectionResolver, error) {
	if err := areOwnEndpointsAvailable(ctx); err != nil {
		return nil, err
	}
	repoID := commit.Repository().IDInt32()

	// Retrieve recent contributors signals.
	ownerships, err := computeRecentContributorSignals(ctx, r.db, repoRootPath, repoID)
	if err != nil {
		return nil, err
	}

	// Retrieve recent view signals.
	viewerResolvers, err := computeRecentViewSignals(ctx, r.logger, r.db, repoRootPath, repoID)
	if err != nil {
		return nil, err
	}
	ownerships = append(ownerships, viewerResolvers...)

	return r.ownershipConnection(args, ownerships)
}

func (r *ownResolver) GitTreeOwnership(
	ctx context.Context,
	tree *graphqlbackend.GitTreeEntryResolver,
	args graphqlbackend.ListOwnershipArgs,
) (graphqlbackend.OwnershipConnectionResolver, error) {
	if err := areOwnEndpointsAvailable(ctx); err != nil {
		return nil, err
	}

	// Retrieve recent contributors signals.
	repoID := tree.Repository().IDInt32()
	ownerships, err := computeRecentContributorSignals(ctx, r.db, tree.Path(), repoID)
	if err != nil {
		return nil, err
	}

	// Retrieve recent view signals.
	viewerResolvers, err := computeRecentViewSignals(ctx, r.logger, r.db, tree.Path(), repoID)
	if err != nil {
		return nil, err
	}
	ownerships = append(ownerships, viewerResolvers...)

	return r.ownershipConnection(args, ownerships)
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

// computeCodeowners evaluates the codeowners file (if any) against given file (blob)
// and returns resolvers for identified owners.
func (r *ownResolver) computeCodeowners(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver) ([]*ownershipResolver, error) {
	repo := blob.Repository()
	repoID, repoName := repo.IDInt32(), repo.RepoName()
	commitID := api.CommitID(blob.Commit().OID())
	var ownerships []*ownershipResolver
	rs, err := r.ownService().RulesetForRepo(ctx, repoName, repoID, commitID)
	if err != nil {
		return nil, err
	}
	var rule *codeownerspb.Rule
	if rs != nil {
		rule = rs.Match(blob.Path())
	}
	if rule != nil {
		owners := rule.GetOwner()
		resolvedOwners, err := r.ownService().ResolveOwnersWithType(ctx, owners)
		if err != nil {
			return nil, err
		}
		for _, ro := range resolvedOwners {
			res := &codeownersFileEntryResolver{
				db:              r.db,
				gitserverClient: r.gitserver,
				source:          rs.GetSource(),
				repo:            blob.Repository(),
				matchLineNumber: rule.GetLineNumber(),
			}
			ownerships = append(ownerships, &ownershipResolver{
				db:            r.db,
				resolvedOwner: ro,
				reasons: []graphqlbackend.OwnershipReasonResolver{
					&ownershipReasonResolver{res},
				},
			})
		}
	}
	return ownerships, nil
}

// ownershipConnection handles ordering and pagination of given set of ownerships.
func (r *ownResolver) ownershipConnection(
	args graphqlbackend.ListOwnershipArgs,
	ownerships []*ownershipResolver,
) (*ownershipConnectionResolver, error) {
	// 1. Order ownerships for deterministic pagination:

	// TODO(#51636): Introduce deterministic ordering based on priority of signals.
	sort.Slice(ownerships, func(i, j int) bool {
		o, p := ownerships[i], ownerships[j]
		if x, y := o.order(), p.order(); x != y {
			return x < y
		}
		iText := o.resolvedOwner.Identifier()
		jText := p.resolvedOwner.Identifier()
		return iText < jText
	})
	total := len(ownerships)

	// 2. Apply pagination from the parameter & compute next cursor:
	cursor, err := graphqlutil.DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}
	for cursor != "" && len(ownerships) > 0 && ownerships[0].resolvedOwner.Identifier() != cursor {
		ownerships = ownerships[1:]
	}
	var next *string
	if args.First != nil && len(ownerships) > int(*args.First) {
		c := ownerships[*args.First].resolvedOwner.Identifier()
		next = &c
		ownerships = ownerships[:*args.First]
	}

	// 3. Assemble the connection resolver object:
	var rs []graphqlbackend.OwnershipResolver
	for _, o := range ownerships {
		rs = append(rs, o)
	}
	return &ownershipConnectionResolver{
		db:         r.db,
		total:      total,
		next:       next,
		ownerships: rs,
	}, nil
}

type ownershipConnectionResolver struct {
	db         edb.EnterpriseDB
	total      int
	next       *string
	ownerships []graphqlbackend.OwnershipResolver
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

func (r *ownershipResolver) order() int {
	reasonsCount := 0
	codeownersCount := 0
	for _, r := range r.reasons {
		reasonsCount++
		if _, ok := r.ToCodeownersFileEntry(); ok {
			codeownersCount++
		}
	}
	// Smaller numbers are ordered in front, so take negative score.
	return -10*codeownersCount + reasonsCount
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
	if person.User != nil {
		return graphqlbackend.NewPersonResolverFromUser(r.db, person.GetEmail(), person.User), true
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
	return "codeowners", nil
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

type recentContributorOwnershipSignal struct {
	total int32
}

func (g *recentContributorOwnershipSignal) Title() (string, error) {
	return "recent contributor", nil
}

func (g *recentContributorOwnershipSignal) Description() (string, error) {
	return "Owner is associated because they have contributed to this file in the last 90 days.", nil
}

func computeRecentContributorSignals(ctx context.Context, db edb.EnterpriseDB, path string, repoID api.RepoID) (results []*ownershipResolver, err error) {
	recentAuthors, err := db.RecentContributionSignals().FindRecentAuthors(ctx, repoID, path)
	if err != nil {
		return nil, errors.Wrap(err, "FindRecentAuthors")
	}

	for _, author := range recentAuthors {
		res := ownershipResolver{
			db: db,
			resolvedOwner: &codeowners.Person{
				Handle: author.AuthorName,
				Email:  author.AuthorEmail,
			},
			reasons: []graphqlbackend.OwnershipReasonResolver{&ownershipReasonResolver{&recentContributorOwnershipSignal{}}},
		}
		user, err := db.Users().GetByVerifiedEmail(ctx, author.AuthorEmail)
		if err == nil {
			// if we don't get an error (meaning we can match) we will add it to the resolver, otherwise use the contributor data
			em := author.AuthorEmail
			res.resolvedOwner = &codeowners.Person{
				User:         user,
				Email:        em,
				PrimaryEmail: &em,
				Handle:       author.AuthorName,
			}
		}
		results = append(results, &res)
	}
	return results, nil
}

type recentViewOwnershipSignal struct {
	total int32
}

func (v *recentViewOwnershipSignal) Title() (string, error) {
	return "recent view", nil
}

func (v *recentViewOwnershipSignal) Description() (string, error) {
	return "Owner is associated because they have viewed this file in the last 90 days.", nil
}

func computeRecentViewSignals(ctx context.Context, logger log.Logger, db edb.EnterpriseDB, path string, repoID api.RepoID) (results []*ownershipResolver, err error) {
	summaries, err := db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{Path: path, RepoID: repoID})
	if err != nil {
		return nil, errors.Wrap(err, "list recent view signals")
	}

	fetchedUsers := make(map[int32]*types.User)
	userEmails := make(map[int32]string)

	for _, summary := range summaries {
		var user *types.User
		var email string
		userID := summary.UserID
		if fetchedUser, found := fetchedUsers[userID]; found {
			user = fetchedUser
			email = userEmails[userID]
		} else {
			userFromDB, err := db.Users().GetByID(ctx, userID)
			if err != nil {
				return nil, errors.Wrap(err, "getting user")
			}
			primaryEmail, _, err := db.UserEmails().GetPrimaryEmail(ctx, userID)
			if err != nil {
				if errcode.IsNotFound(err) {
					logger.Warn("Cannot find a primary email", log.Int32("userID", userID))
				} else {
					return nil, errors.Wrap(err, "getting user primary email")
				}
			}
			user = userFromDB
			email = primaryEmail
			fetchedUsers[userID] = userFromDB
			userEmails[userID] = primaryEmail
		}
		// TODO(sashaostrikov): what to do if email here is empty
		res := ownershipResolver{
			db: db,
			resolvedOwner: &codeowners.Person{
				User:         user,
				PrimaryEmail: &email,
				Handle:       user.Username,
			},
			reasons: []graphqlbackend.OwnershipReasonResolver{&ownershipReasonResolver{&recentViewOwnershipSignal{}}},
		}
		results = append(results, &res)
	}
	return results, nil
}
