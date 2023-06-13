package resolvers

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/rbac"

	owntypes "github.com/sourcegraph/sourcegraph/enterprise/internal/own/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"

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
	_ graphqlbackend.AssignedOwnerResolver                    = &assignedOwner{}
	_ graphqlbackend.SimpleOwnReasonResolver                  = &assignedOwner{}
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

type ownershipReason struct {
	codeownersRule           *codeownerspb.Rule
	codeownersSource         codeowners.RulesetSource
	recentContributionsCount int
	recentViewsCount         int
	assignedOwnerPath        []string
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

func (o *ownershipReasonResolver) ToAssignedOwner() (res graphqlbackend.AssignedOwnerResolver, ok bool) {
	res, ok = o.resolver.(*assignedOwner)
	return
}

func (o *ownershipReasonResolver) makesAnOwner() bool {
	_, makesAnOwner := o.resolver.(*codeownersFileEntryResolver)
	_, makesAnAssignedOwner := o.resolver.(*assignedOwner)
	return makesAnOwner || makesAnAssignedOwner
}

func (r *ownResolver) GitBlobOwnership(
	ctx context.Context,
	blob *graphqlbackend.GitTreeEntryResolver,
	args graphqlbackend.ListOwnershipArgs,
) (graphqlbackend.OwnershipConnectionResolver, error) {
	if err := areOwnEndpointsAvailable(ctx); err != nil {
		return nil, err
	}

	var rrs []reasonAndReference

	// Evaluate CODEOWNERS rules.
	if args.IncludeReason(graphqlbackend.CodeownersFileEntry) {
		codeowners, err := r.computeCodeowners(ctx, blob)
		if err != nil {
			return nil, err
		}
		rrs = append(rrs, codeowners...)
	}

	repoID := blob.Repository().IDInt32()

	// Retrieve recent contributors signals.
	if args.IncludeReason(graphqlbackend.RecentContributorOwnershipSignal) {
		contribResolvers, err := computeRecentContributorSignals(ctx, r.db, blob.Path(), repoID)
		if err != nil {
			return nil, err
		}
		rrs = append(rrs, contribResolvers...)
	}

	// Retrieve recent view signals.
	if args.IncludeReason(graphqlbackend.RecentViewOwnershipSignal) {
		viewerResolvers, err := computeRecentViewSignals(ctx, r.db, blob.Path(), repoID)
		if err != nil {
			return nil, err
		}
		rrs = append(rrs, viewerResolvers...)
	}

	// Retrieve assigned owners.
	assignedOwners, err := r.computeAssignedOwners(ctx, blob, repoID)
	if err != nil {
		return nil, err
	}
	rrs = append(rrs, assignedOwners...)

	// Retrieve assigned teams.
	assignedTeams, err := r.computeAssignedTeams(ctx, blob, repoID)
	if err != nil {
		return nil, err
	}
	rrs = append(rrs, assignedTeams...)

	return r.ownershipConnection(ctx, args, rrs, blob.Repository(), blob.Path())
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
	rrs, err := computeRecentContributorSignals(ctx, r.db, repoRootPath, repoID)
	if err != nil {
		return nil, err
	}

	// Retrieve recent view signals.
	viewerResolvers, err := computeRecentViewSignals(ctx, r.db, repoRootPath, repoID)
	if err != nil {
		return nil, err
	}
	rrs = append(rrs, viewerResolvers...)

	return r.ownershipConnection(ctx, args, rrs, commit.Repository(), "")
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
	rrs, err := computeRecentContributorSignals(ctx, r.db, tree.Path(), repoID)
	if err != nil {
		return nil, err
	}

	// Retrieve recent view signals.
	viewerResolvers, err := computeRecentViewSignals(ctx, r.db, tree.Path(), repoID)
	if err != nil {
		return nil, err
	}
	rrs = append(rrs, viewerResolvers...)

	// Retrieve assigned owners.
	assignedOwners, err := r.computeAssignedOwners(ctx, tree, repoID)
	if err != nil {
		return nil, err
	}
	rrs = append(rrs, assignedOwners...)

	// Retrieve assigned teams.
	assignedTeams, err := r.computeAssignedTeams(ctx, tree, repoID)
	if err != nil {
		return nil, err
	}
	rrs = append(rrs, assignedTeams...)

	return r.ownershipConnection(ctx, args, rrs, tree.Repository(), tree.Path())
}

func (r *ownResolver) GitTreeOwnershipStats(ctx context.Context, tree *graphqlbackend.GitTreeEntryResolver) (graphqlbackend.OwnershipStatsResolver, error) {
	return &ownStatsResolver{
		db: r.db,
		opts: database.TreeLocationOpts{
			RepoID: tree.Repository().IDInt32(),
			Path:   tree.Path(),
		},
	}, nil
}

func (r *ownResolver) InstanceOwnershipStats(ctx context.Context) (graphqlbackend.OwnershipStatsResolver, error) {
	return &ownStatsResolver{db: r.db}, nil
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

type reasonAndReference struct {
	reason    ownershipReason
	reference own.Reference
}

// computeCodeowners evaluates the codeowners file (if any) against given file (blob)
// and returns resolvers for identified owners.
func (r *ownResolver) computeCodeowners(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver) ([]reasonAndReference, error) {
	repo := blob.Repository()
	repoID, repoName := repo.IDInt32(), repo.RepoName()
	commitID := api.CommitID(blob.Commit().OID())
	// Find ruleset which represents CODEOWNERS file at given revision.
	ruleset, err := r.ownService().RulesetForRepo(ctx, repoName, repoID, commitID)
	if err != nil {
		return nil, err
	}
	var rule *codeownerspb.Rule
	if ruleset != nil {
		rule = ruleset.Match(blob.Path())
	}
	// Compute repo context if possible to allow better unification of references.
	var repoContext *own.RepoContext
	if len(rule.GetOwner()) > 0 {
		spec, err := repo.ExternalRepo(ctx)
		fmt.Println("EXTERNAL REPO", spec, spec.ServiceType)
		// Best effort resolution. We still want to serve the reason if external service cannot be resolved here.
		if err == nil {
			repoContext = &own.RepoContext{
				Name:         repoName,
				CodeHostKind: spec.ServiceType,
			}
		}
	}
	// Return references
	var rrs []reasonAndReference
	for _, o := range rule.GetOwner() {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{
				codeownersRule:   rule,
				codeownersSource: ruleset.GetSource(),
			},
			reference: own.Reference{
				RepoContext: repoContext,
				Handle:      o.Handle,
				Email:       o.Email,
			},
		})
		fmt.Println("RRS", rrs[len(rrs)-1].reference.String())
	}
	return rrs, nil
	// bag := own.EmptyBag()
	// ruleToRefs := make(map[*codeownerspb.Rule][]own.Reference, 0)
	// if rule != nil {
	// 	owners := rule.GetOwner()
	// 	externalRepo, err := repo.ExternalRepo(ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	// For each owner from CODEOWNERS file, we create a reference and put it into the
	// 	// bag.
	// 	for _, owner := range owners {
	// 		ref := own.Reference{
	// 			RepoContext: &own.RepoContext{
	// 				Name:         repoName,
	// 				CodeHostKind: externalRepo.ServiceType,
	// 			},
	// 			Handle: owner.Handle,
	// 			Email:  owner.Email,
	// 		}
	// 		bag.Add(ref)
	// 		if _, ok := ruleToRefs[rule]; !ok {
	// 			ruleToRefs[rule] = make([]own.Reference, 0)
	// 		}
	// 		ruleToRefs[rule] = append(ruleToRefs[rule], ref)
	// 	}
	// }
	// bag.Resolve(ctx, r.db)
	// for rule, refs := range ruleToRefs {
	// 	for _, ref := range refs {
	// 		resolvedOwner, found := bag.FindResolved(ref)
	// 		if !found {
	// 			if guess := ref.ResolutionGuess(); guess != nil {
	// 				resolvedOwner = guess
	// 			}
	// 		}
	// 		if resolvedOwner != nil {
	// 			res := &codeownersFileEntryResolver{
	// 				db:              r.db,
	// 				gitserverClient: r.gitserver,
	// 				source:          rs.GetSource(),
	// 				repo:            blob.Repository(),
	// 				matchLineNumber: rule.GetLineNumber(),
	// 			}
	// 			ownerships = append(ownerships, &ownershipResolver{
	// 				db:            r.db,
	// 				resolvedOwner: resolvedOwner,
	// 				reasons: []*ownershipReasonResolver{
	// 					{
	// 						res,
	// 					},
	// 				},
	// 			})
	// 		}
	// 	}
	// }
}

type reasonsAndOwner struct {
	reasons []ownershipReason
	owner   codeowners.ResolvedOwner
}

func (ro reasonsAndOwner) String() string {
	var b bytes.Buffer
	fmt.Fprint(&b, ro.owner.Identifier())
	for _, r := range ro.reasons {
		if r.codeownersRule != nil {
			fmt.Fprint(&b, " CODEOWNERS")
		}
		if len(r.assignedOwnerPath) > 0 {
			fmt.Fprint(&b, " ASSIGNED OWNER")
		}
		if r.recentContributionsCount > 0 {
			fmt.Fprint(&b, " RECENT CONTRIBUTOR")
		}
		if r.recentViewsCount > 0 {
			fmt.Fprint(&b, " RECENT VIEWER")
		}
	}
	return b.String()
}

func (ro reasonsAndOwner) order() int {
	var ownershipReasons, reasons, contributions, views int
	for _, r := range ro.reasons {
		if len(r.assignedOwnerPath) > 0 || r.codeownersRule != nil {
			ownershipReasons++
		}
		reasons++
		contributions += r.recentContributionsCount
		views += r.recentViewsCount
	}
	// Smaller numbers are ordered in front, so take negative score.
	return -(100000*ownershipReasons +
		1000*reasons +
		10*contributions +
		views)
}

func (ro reasonsAndOwner) isOwner() bool {
	for _, r := range ro.reasons {
		if len(r.assignedOwnerPath) > 0 || r.codeownersRule != nil {
			return true
		}
	}
	return false
}

// ownershipConnection handles ordering and pagination of given set of ownerships.
func (r *ownResolver) ownershipConnection(
	ctx context.Context,
	args graphqlbackend.ListOwnershipArgs,
	ownerships []reasonAndReference,
	repo *graphqlbackend.RepositoryResolver,
	path string,
) (*ownershipConnectionResolver, error) {
	fmt.Println("REASON & REFERENCE", ownerships)
	// 1. Resolve ownership references
	bag := own.EmptyBag()
	for _, r := range ownerships {
		bag.Add(r.reference)
	}
	bag.Resolve(ctx, r.db)
	fmt.Println(bag)
	// 2. Group reasons by resolved owners
	ownersByKey := map[string]*reasonsAndOwner{}
	for _, r := range ownerships {
		resolvedOwner, found := bag.FindResolved(r.reference)
		fmt.Println("RESOLVED OWNER", r.reference, resolvedOwner, found)
		if !found {
			if guess := r.reference.ResolutionGuess(); guess != nil {
				resolvedOwner = guess
				fmt.Println("GUESSED OWNER", r.reference, resolvedOwner, found)
			} else {
				fmt.Println("NOT GUESSED")
			}
		}
		if resolvedOwner == nil {
			// TODO: Log - this is a bug
			continue
		}
		ownerKey := resolvedOwner.Identifier()
		ownerInfo, ok := ownersByKey[ownerKey]
		if !ok {
			ownerInfo = &reasonsAndOwner{owner: resolvedOwner}
			ownersByKey[ownerKey] = ownerInfo
		}
		ownerInfo.reasons = append(ownerInfo.reasons, r.reason)
	}
	var owners []reasonsAndOwner
	for _, o := range ownersByKey {
		owners = append(owners, *o)
	}

	// 3. Order ownerships for deterministic pagination:
	sort.Slice(owners, func(i, j int) bool {
		o, p := owners[i], owners[j]
		if x, y := o.order(), p.order(); x != y {
			return x < y
		}
		return o.owner.Identifier() < p.owner.Identifier()
	})

	fmt.Println("OWNERS FINAL", owners)

	// 4. Compute counts
	total := len(owners)
	var totalOwners int
	for _, o := range owners {
		if o.isOwner() {
			totalOwners++
		}
	}

	// 5. Apply pagination from the parameter & compute next cursor:
	cursor, err := graphqlutil.DecodeCursor(args.After)
	if err != nil {
		return nil, err
	}
	for cursor != "" && len(owners) > 0 && owners[0].owner.Identifier() != cursor {
		owners = owners[1:]
	}
	var next *string
	if args.First != nil && len(owners) > int(*args.First) {
		c := owners[*args.First].owner.Identifier()
		next = &c
		owners = owners[:*args.First]
	}

	// 6. Assemble the connection resolver object:
	return &ownershipConnectionResolver{
		db:          r.db,
		total:       total,
		totalOwners: totalOwners,
		next:        next,
		owners:      owners,
		gitserver:   r.gitserver,
		repo:        repo,
		path:        path,
	}, nil
}

type ownStatsResolver struct {
	db   edb.EnterpriseDB
	opts database.TreeLocationOpts
}

func (r *ownStatsResolver) TotalFiles(ctx context.Context) (int32, error) {
	return r.db.RepoPaths().AggregateFileCount(ctx, r.opts)
}

func (r *ownStatsResolver) TotalCodeownedFiles(ctx context.Context) (int32, error) {
	counts, err := r.db.OwnershipStats().QueryAggregateCounts(ctx, r.opts)
	if err != nil {
		return 0, err
	}
	return int32(counts.CodeownedFileCount), nil
}

type ownershipConnectionResolver struct {
	db          edb.EnterpriseDB
	total       int
	totalOwners int
	next        *string
	owners      []reasonsAndOwner
	gitserver   gitserver.Client
	repo        *graphqlbackend.RepositoryResolver
	path        string
}

func (r *ownershipConnectionResolver) TotalCount(_ context.Context) (int32, error) {
	return int32(r.total), nil
}

func (r *ownershipConnectionResolver) TotalOwners(_ context.Context) (int32, error) {
	return int32(r.totalOwners), nil
}

func (r *ownershipConnectionResolver) PageInfo(_ context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.EncodeCursor(r.next), nil
}

func (r *ownershipConnectionResolver) Nodes(_ context.Context) ([]graphqlbackend.OwnershipResolver, error) {
	var rs []graphqlbackend.OwnershipResolver
	for _, o := range r.owners {
		rs = append(rs, &ownershipResolver{
			db:            r.db,
			gitserver:     r.gitserver,
			repo:          r.repo,
			resolvedOwner: o.owner,
			reasons:       o.reasons,
			path:          r.path,
		})
	}
	return rs, nil
}

type ownershipResolver struct {
	db            edb.EnterpriseDB
	gitserver     gitserver.Client
	resolvedOwner codeowners.ResolvedOwner
	path          string
	repo          *graphqlbackend.RepositoryResolver
	reasons       []ownershipReason
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
	var rs []graphqlbackend.OwnershipReasonResolver
	for _, reason := range r.reasons {
		if reason.codeownersRule != nil {
			rs = append(rs, &ownershipReasonResolver{
				resolver: &codeownersFileEntryResolver{
					db:              r.db,
					gitserverClient: r.gitserver,
					source:          reason.codeownersSource,
					repo:            r.repo,
					matchLineNumber: reason.codeownersRule.GetLineNumber(),
				},
			})

		}
		for _, p := range reason.assignedOwnerPath {
			rs = append(rs, &ownershipReasonResolver{
				resolver: &assignedOwner{
					directMatch: r.path == p,
				},
			})
		}
		if reason.recentContributionsCount > 0 {
			rs = append(rs, &ownershipReasonResolver{
				resolver: &recentContributorOwnershipSignal{
					total: int32(reason.recentContributionsCount),
				},
			})
		}
		if reason.recentViewsCount > 0 {
			rs = append(rs, &ownershipReasonResolver{
				resolver: &recentViewOwnershipSignal{
					total: int32(reason.recentViewsCount),
				},
			})
		}
	}
	return rs, nil
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
		fmt.Println("NOT TYPE TEAM")
		return nil, false
	}
	resolvedTeam, ok := r.resolvedOwner.(*codeowners.Team)
	if !ok {
		fmt.Println("NOT TEAM", resolvedTeam)
		return nil, false
	}
	if resolvedTeam.Team != nil {
		fmt.Println("GOT TEAM", resolvedTeam.Team)
		return graphqlbackend.NewTeamResolver(r.db, resolvedTeam.Team), true
	}
	fmt.Println("GUESSSSS")
	// The sourcegraph instance does not have that team in the database.
	// It might be an unimported code-host team, a guess or both.
	return graphqlbackend.NewTeamResolver(r.db, &types.Team{
		Name:        resolvedTeam.Identifier(),
		DisplayName: resolvedTeam.Identifier(),
	}), true
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

func computeRecentContributorSignals(ctx context.Context, db edb.EnterpriseDB, path string, repoID api.RepoID) ([]reasonAndReference, error) {
	enabled, err := db.OwnSignalConfigurations().IsEnabled(ctx, owntypes.SignalRecentContributors)
	if err != nil {
		return nil, errors.Wrap(err, "IsEnabled")
	}
	if !enabled {
		return nil, nil
	}

	recentAuthors, err := db.RecentContributionSignals().FindRecentAuthors(ctx, repoID, path)
	if err != nil {
		return nil, errors.Wrap(err, "FindRecentAuthors")
	}

	var rrs []reasonAndReference
	for _, a := range recentAuthors {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{recentContributionsCount: a.ContributionCount},
			reference: own.Reference{
				// Just use the email.
				Email: a.AuthorEmail,
			},
		})
	}
	return rrs, nil

	// for _, author := range recentAuthors {
	// 	res := ownershipResolver{
	// 		db: db,
	// 		resolvedOwner: &codeowners.Person{
	// 			Handle: author.AuthorName,
	// 			Email:  author.AuthorEmail,
	// 		},
	// 		reasons: []*ownershipReasonResolver{
	// 			{
	// 				&recentContributorOwnershipSignal{},
	// 			},
	// 		},
	// 	}
	// 	user, err := db.Users().GetByVerifiedEmail(ctx, author.AuthorEmail)
	// 	if err == nil {
	// 		// if we don't get an error (meaning we can match) we will add it to the resolver, otherwise use the contributor data
	// 		em := author.AuthorEmail
	// 		res.resolvedOwner = &codeowners.Person{
	// 			User:         user,
	// 			Email:        em,
	// 			PrimaryEmail: &em,
	// 			Handle:       author.AuthorName,
	// 		}
	// 	}
	// 	results = append(results, &res)
	// }
	// return ors, nil
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

func computeRecentViewSignals(ctx context.Context, db edb.EnterpriseDB, path string, repoID api.RepoID) ([]reasonAndReference, error) {
	enabled, err := db.OwnSignalConfigurations().IsEnabled(ctx, owntypes.SignalRecentViews)
	if err != nil {
		return nil, errors.Wrap(err, "IsEnabled")
	}
	if !enabled {
		return nil, nil
	}

	summaries, err := db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{Path: path, RepoID: repoID})
	if err != nil {
		return nil, errors.Wrap(err, "list recent view signals")
	}

	var rrs []reasonAndReference
	for _, s := range summaries {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{recentViewsCount: s.ViewsCount},
			reference: own.Reference{
				UserID: s.UserID,
			},
		})
	}
	return rrs, nil

	// fetchedUsers := make(map[int32]*types.User)
	// userEmails := make(map[int32]string)

	// for _, summary := range summaries {
	// 	var user *types.User
	// 	var email string
	// 	userID := summary.UserID
	// 	if fetchedUser, found := fetchedUsers[userID]; found {
	// 		user = fetchedUser
	// 		email = userEmails[userID]
	// 	} else {
	// 		userFromDB, err := db.Users().GetByID(ctx, userID)
	// 		if err != nil {
	// 			return nil, errors.Wrap(err, "getting user")
	// 		}
	// 		primaryEmail, _, err := db.UserEmails().GetPrimaryEmail(ctx, userID)
	// 		if err != nil {
	// 			if errcode.IsNotFound(err) {
	// 				logger.Warn("Cannot find a primary email", log.Int32("userID", userID))
	// 			} else {
	// 				return nil, errors.Wrap(err, "getting user primary email")
	// 			}
	// 		}
	// 		user = userFromDB
	// 		email = primaryEmail
	// 		fetchedUsers[userID] = userFromDB
	// 		userEmails[userID] = primaryEmail
	// 	}
	// 	// TODO(sashaostrikov): what to do if email here is empty
	// 	res := ownershipResolver{
	// 		db: db,
	// 		resolvedOwner: &codeowners.Person{
	// 			User:         user,
	// 			PrimaryEmail: &email,
	// 			Handle:       user.Username,
	// 		},
	// 		reasons: []*ownershipReasonResolver{
	// 			{
	// 				&recentViewOwnershipSignal{},
	// 			},
	// 		},
	// 	}
	// 	results = append(results, &res)
	// }
	// return results, nil
}

type assignedOwner struct {
	directMatch bool
}

func (a *assignedOwner) Title() (string, error) {
	return "assigned owner", nil
}

func (a *assignedOwner) Description() (string, error) {
	return "Owner is manually assigned.", nil
}

func (a *assignedOwner) IsDirectMatch() bool {
	return a.directMatch
}

func (r *ownResolver) computeAssignedOwners(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, repoID api.RepoID) ([]reasonAndReference, error) {
	assignedOwnership, err := r.ownService().AssignedOwnership(ctx, repoID, api.CommitID(blob.Commit().OID()))
	if err != nil {
		return nil, errors.Wrap(err, "computing assigned ownership")
	}
	var rrs []reasonAndReference
	for _, o := range assignedOwnership.Match(blob.Path()) {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{
				assignedOwnerPath: []string{o.FilePath},
			},
			reference: own.Reference{
				UserID: o.OwnerUserID,
			},
		})
	}
	return rrs, nil

	// fetchedUsers := make(map[int32]*types.User)
	// userEmails := make(map[int32]string)

	// isDirectMatch := false
	// for _, summary := range assignedOwnerSummaries {
	// 	var user *types.User
	// 	var email string
	// 	userID := summary.OwnerUserID
	// 	if fetchedUser, found := fetchedUsers[userID]; found {
	// 		user = fetchedUser
	// 		email = userEmails[userID]
	// 	} else {
	// 		userFromDB, err := db.Users().GetByID(ctx, userID)
	// 		if err != nil {
	// 			return nil, errors.Wrap(err, "getting user")
	// 		}
	// 		primaryEmail, _, err := db.UserEmails().GetPrimaryEmail(ctx, userID)
	// 		if err != nil {
	// 			if errcode.IsNotFound(err) {
	// 				logger.Warn("Cannot find a primary email", log.Int32("userID", userID))
	// 			} else {
	// 				return nil, errors.Wrap(err, "getting user primary email")
	// 			}
	// 		}
	// 		user = userFromDB
	// 		email = primaryEmail
	// 		fetchedUsers[userID] = userFromDB
	// 		userEmails[userID] = primaryEmail
	// 	}
	// 	if blob.Path() == summary.FilePath {
	// 		isDirectMatch = true
	// 	}
	// 	res := ownershipResolver{
	// 		db: db,
	// 		resolvedOwner: &codeowners.Person{
	// 			User:         user,
	// 			PrimaryEmail: &email,
	// 			Handle:       user.Username,
	// 		},
	// 		reasons: []*ownershipReasonResolver{
	// 			{
	// 				&assignedOwner{directMatch: isDirectMatch},
	// 			},
	// 		},
	// 	}
	// 	results = append(results, &res)
	// }
	// return results, nil
}

func (r *ownResolver) computeAssignedTeams(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, repoID api.RepoID) ([]reasonAndReference, error) {
	assignedTeams, err := r.ownService().AssignedTeams(ctx, repoID, api.CommitID(blob.Commit().OID()))
	if err != nil {
		return nil, errors.Wrap(err, "computing assigned ownership")
	}
	var rrs []reasonAndReference
	for _, summary := range assignedTeams.Match(blob.Path()) {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{
				assignedOwnerPath: []string{summary.FilePath},
			},
			reference: own.Reference{
				TeamID: summary.OwnerTeamID,
			},
		})
	}
	return rrs, nil
}

func (r *ownResolver) OwnSignalConfigurations(ctx context.Context) ([]graphqlbackend.SignalConfigurationResolver, error) {
	err := auth.CheckCurrentActorIsSiteAdmin(actor.FromContext(ctx), r.db)
	if err != nil {
		return nil, err
	}
	var resolvers []graphqlbackend.SignalConfigurationResolver
	store := r.db.OwnSignalConfigurations()
	configurations, err := store.LoadConfigurations(ctx, database.LoadSignalConfigurationArgs{})
	if err != nil {
		return nil, errors.Wrap(err, "LoadConfigurations")
	}

	for _, configuration := range configurations {
		resolvers = append(resolvers, &signalConfigResolver{config: configuration})
	}

	return resolvers, nil
}

type signalConfigResolver struct {
	config database.SignalConfiguration
}

func (s *signalConfigResolver) Name() string {
	return s.config.Name
}

func (s *signalConfigResolver) Description() string {
	return s.config.Description
}

func (s *signalConfigResolver) IsEnabled() bool {
	return s.config.Enabled
}

func (s *signalConfigResolver) ExcludedRepoPatterns() []string {
	return userifyPatterns(s.config.ExcludedRepoPatterns)
}

func (r *ownResolver) UpdateOwnSignalConfigurations(ctx context.Context, args graphqlbackend.UpdateSignalConfigurationsArgs) ([]graphqlbackend.SignalConfigurationResolver, error) {
	err := auth.CheckCurrentActorIsSiteAdmin(actor.FromContext(ctx), r.db)
	if err != nil {
		return nil, err
	}

	err = r.db.OwnSignalConfigurations().WithTransact(ctx, func(store database.SignalConfigurationStore) error {
		for _, config := range args.Input.Configs {
			if err := store.UpdateConfiguration(ctx, database.UpdateSignalConfigurationArgs{
				Name:                 config.Name,
				ExcludedRepoPatterns: postgresifyPatterns(config.ExcludedRepoPatterns),
				Enabled:              config.Enabled,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.OwnSignalConfigurations(ctx)
}

// postgresifyPatterns will convert glob-ish patterns to postgres compatible patterns. For example github.com/* -> github.com/%
func postgresifyPatterns(patterns []string) (results []string) {
	for _, pattern := range patterns {
		results = append(results, strings.ReplaceAll(pattern, "*", "%"))
	}
	return results
}

// userifyPatterns will convert postgres patterns to glob-ish patterns. For example github.com/% -> github.com/*.
func userifyPatterns(patterns []string) (results []string) {
	for _, pattern := range patterns {
		results = append(results, strings.ReplaceAll(pattern, "%", "*"))
	}
	return results
}

func (r *ownResolver) AssignOwner(ctx context.Context, args *graphqlbackend.AssignOwnerOrTeamArgs) (*graphqlbackend.EmptyResponse, error) {
	// Internal actor is a no-op, only a user can assign an owner.
	if actor.FromContext(ctx).IsInternal() {
		return nil, nil
	}
	user, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	u, err := unmarshalAssignOwnerArgs(args.Input, userUnmarshalMode)
	if err != nil {
		return nil, err
	}
	whoAssignedUserID := user.ID
	err = r.db.AssignedOwners().Insert(ctx, u.AssignedOwnerOrTeamID, u.RepoID, u.AbsolutePath, whoAssignedUserID)
	if err != nil {
		return nil, errors.Wrap(err, "creating assigned owner")
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *ownResolver) RemoveAssignedOwner(ctx context.Context, args *graphqlbackend.AssignOwnerOrTeamArgs) (*graphqlbackend.EmptyResponse, error) {
	// Internal actor is a no-op, only a user can remove an assigned owner.
	if actor.FromContext(ctx).IsInternal() {
		return nil, nil
	}
	_, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	u, err := unmarshalAssignOwnerArgs(args.Input, userUnmarshalMode)
	if err != nil {
		return nil, err
	}
	err = r.db.AssignedOwners().DeleteOwner(ctx, u.AssignedOwnerOrTeamID, u.RepoID, u.AbsolutePath)
	if err != nil {
		return nil, errors.Wrap(err, "deleting assigned owner")
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *ownResolver) AssignTeam(ctx context.Context, args *graphqlbackend.AssignOwnerOrTeamArgs) (*graphqlbackend.EmptyResponse, error) {
	// Internal actor is a no-op, only a user can assign an owner.
	if actor.FromContext(ctx).IsInternal() {
		return nil, nil
	}
	user, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	t, err := unmarshalAssignOwnerArgs(args.Input, teamUnmarshalMode)
	if err != nil {
		return nil, err
	}
	whoAssignedUserID := user.ID
	err = r.db.AssignedTeams().Insert(ctx, t.AssignedOwnerOrTeamID, t.RepoID, t.AbsolutePath, whoAssignedUserID)
	if err != nil {
		return nil, errors.Wrap(err, "creating assigned team")
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *ownResolver) RemoveAssignedTeam(ctx context.Context, args *graphqlbackend.AssignOwnerOrTeamArgs) (*graphqlbackend.EmptyResponse, error) {
	// Internal actor is a no-op, only a user can remove an assigned owner.
	if actor.FromContext(ctx).IsInternal() {
		return nil, nil
	}
	_, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	t, err := unmarshalAssignOwnerArgs(args.Input, teamUnmarshalMode)
	if err != nil {
		return nil, err
	}
	err = r.db.AssignedTeams().DeleteOwnerTeam(ctx, t.AssignedOwnerOrTeamID, t.RepoID, t.AbsolutePath)
	if err != nil {
		return nil, errors.Wrap(err, "deleting assigned team")
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

// checkAssignedOwnershipPermission checks that the user from the context has
// `rbac.OwnershipAssignPermission` and returns the user if so.
func (r *ownResolver) checkAssignedOwnershipPermission(ctx context.Context) (*types.User, error) {
	// Extracting the user to run an RBAC check and then use their ID as an
	// `whoAssignedUserID`.
	user, err := auth.CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, auth.ErrNotAuthenticated
	}
	// Checking if the user has permission to assign an owner.
	if err := rbac.CheckGivenUserHasPermission(ctx, r.db, user, rbac.OwnershipAssignPermission); err != nil {
		return nil, err
	}
	return user, nil
}

type UnmarshalledAssignOwnerArgs struct {
	AssignedOwnerOrTeamID int32
	RepoID                api.RepoID
	AbsolutePath          string
}

type UnmarshalMode string

const (
	userUnmarshalMode UnmarshalMode = "user"
	teamUnmarshalMode UnmarshalMode = "team"
)

func unmarshalAssignOwnerArgs(args graphqlbackend.AssignOwnerOrTeamInput, unmarshalMode UnmarshalMode) (*UnmarshalledAssignOwnerArgs, error) {
	var userOrTeamID int32
	var unmarshalError error
	if unmarshalMode == userUnmarshalMode {
		userOrTeamID, unmarshalError = graphqlbackend.UnmarshalUserID(args.AssignedOwnerID)
	} else if unmarshalMode == teamUnmarshalMode {
		userOrTeamID, unmarshalError = graphqlbackend.UnmarshalTeamID(args.AssignedOwnerID)
	} else {
		return nil, errors.New("only user or team can be assigned ownership")
	}
	if unmarshalError != nil {
		return nil, unmarshalError
	}
	if userOrTeamID == 0 {
		return nil, errors.New(fmt.Sprintf("assigned %s ID should not be 0", unmarshalMode))
	}
	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.RepoID)
	if err != nil {
		return nil, err
	}
	if repoID == 0 {
		return nil, errors.New("repo ID should not be 0")
	}
	return &UnmarshalledAssignOwnerArgs{
		AssignedOwnerOrTeamID: userOrTeamID,
		RepoID:                repoID,
		AbsolutePath:          args.AbsolutePath,
	}, nil
}
