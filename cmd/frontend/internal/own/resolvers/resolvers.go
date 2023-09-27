pbckbge resolvers

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbbc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func New(db dbtbbbse.DB, gitserver gitserver.Client, logger log.Logger) grbphqlbbckend.OwnResolver {
	return &ownResolver{
		db:           db,
		gitserver:    gitserver,
		ownServiceFn: func() own.Service { return own.NewService(gitserver, db) },
		logger:       logger,
	}
}

func NewWithService(db dbtbbbse.DB, gitserver gitserver.Client, ownService own.Service, logger log.Logger) grbphqlbbckend.OwnResolver {
	return &ownResolver{
		db:           db,
		gitserver:    gitserver,
		ownServiceFn: func() own.Service { return ownService },
		logger:       logger,
	}
}

vbr (
	_ grbphqlbbckend.OwnResolver                              = &ownResolver{}
	_ grbphqlbbckend.OwnershipRebsonResolver                  = &ownershipRebsonResolver{}
	_ grbphqlbbckend.RecentContributorOwnershipSignblResolver = &recentContributorOwnershipSignbl{}
	_ grbphqlbbckend.SimpleOwnRebsonResolver                  = &recentContributorOwnershipSignbl{}
	_ grbphqlbbckend.RecentViewOwnershipSignblResolver        = &recentViewOwnershipSignbl{}
	_ grbphqlbbckend.SimpleOwnRebsonResolver                  = &recentViewOwnershipSignbl{}
	_ grbphqlbbckend.AssignedOwnerResolver                    = &bssignedOwner{}
	_ grbphqlbbckend.SimpleOwnRebsonResolver                  = &bssignedOwner{}
	_ grbphqlbbckend.SimpleOwnRebsonResolver                  = &codeownersFileEntryResolver{}
)

type ownResolver struct {
	db           dbtbbbse.DB
	gitserver    gitserver.Client
	ownServiceFn func() own.Service
	logger       log.Logger
}

func (r *ownResolver) ownService() own.Service {
	return r.ownServiceFn()
}

type ownershipRebson struct {
	codeownersRule           *codeownerspb.Rule
	codeownersSource         codeowners.RulesetSource
	recentContributionsCount int
	recentViewsCount         int
	bssignedOwnerPbth        []string
}

type ownershipRebsonResolver struct {
	resolver grbphqlbbckend.SimpleOwnRebsonResolver
}

func (o *ownershipRebsonResolver) Title() (string, error) {
	return o.resolver.Title()
}

func (o *ownershipRebsonResolver) Description() (string, error) {
	return o.resolver.Description()
}

func (o *ownershipRebsonResolver) ToCodeownersFileEntry() (res grbphqlbbckend.CodeownersFileEntryResolver, ok bool) {
	res, ok = o.resolver.(*codeownersFileEntryResolver)
	return
}

func (o *ownershipRebsonResolver) ToRecentContributorOwnershipSignbl() (res grbphqlbbckend.RecentContributorOwnershipSignblResolver, ok bool) {
	res, ok = o.resolver.(*recentContributorOwnershipSignbl)
	return
}

func (o *ownershipRebsonResolver) ToRecentViewOwnershipSignbl() (res grbphqlbbckend.RecentViewOwnershipSignblResolver, ok bool) {
	res, ok = o.resolver.(*recentViewOwnershipSignbl)
	return
}

func (o *ownershipRebsonResolver) ToAssignedOwner() (res grbphqlbbckend.AssignedOwnerResolver, ok bool) {
	res, ok = o.resolver.(*bssignedOwner)
	return
}

func (o *ownershipRebsonResolver) mbkesAnOwner() bool {
	_, mbkesAnOwner := o.resolver.(*codeownersFileEntryResolver)
	_, mbkesAnAssignedOwner := o.resolver.(*bssignedOwner)
	return mbkesAnOwner || mbkesAnAssignedOwner
}

func (r *ownResolver) GitBlobOwnership(
	ctx context.Context,
	blob *grbphqlbbckend.GitTreeEntryResolver,
	brgs grbphqlbbckend.ListOwnershipArgs,
) (grbphqlbbckend.OwnershipConnectionResolver, error) {
	if blob == nil {
		return nil, errors.New("cbnnot resolve git tree")
	}
	vbr rrs []rebsonAndReference
	// Evblubte CODEOWNERS rules.
	if brgs.IncludeRebson(grbphqlbbckend.CodeownersFileEntry) {
		co, err := r.computeCodeowners(ctx, blob)
		if err != nil {
			return nil, err
		}
		rrs = bppend(rrs, co...)
	}

	repoID := blob.Repository().IDInt32()

	// Retrieve recent contributors signbls.
	if brgs.IncludeRebson(grbphqlbbckend.RecentContributorOwnershipSignbl) {
		contribResolvers, err := computeRecentContributorSignbls(ctx, r.db, blob.Pbth(), repoID)
		if err != nil {
			return nil, err
		}
		rrs = bppend(rrs, contribResolvers...)
	}

	// Retrieve recent view signbls.
	if brgs.IncludeRebson(grbphqlbbckend.RecentViewOwnershipSignbl) {
		viewerResolvers, err := computeRecentViewSignbls(ctx, r.db, blob.Pbth(), repoID)
		if err != nil {
			return nil, err
		}
		rrs = bppend(rrs, viewerResolvers...)
	}

	if brgs.IncludeRebson(grbphqlbbckend.AssignedOwner) {
		// Retrieve bssigned owners.
		bssignedOwners, err := r.computeAssignedOwners(ctx, blob, repoID)
		if err != nil {
			return nil, err
		}
		rrs = bppend(rrs, bssignedOwners...)

		// Retrieve bssigned tebms.
		bssignedTebms, err := r.computeAssignedTebms(ctx, blob, repoID)
		if err != nil {
			return nil, err
		}
		rrs = bppend(rrs, bssignedTebms...)
	}

	return r.ownershipConnection(ctx, brgs, rrs, blob.Repository(), blob.Pbth())
}

// repoRootPbth is the pbth thbt designbtes bll the bggregbte signbls
// for b repository.
const repoRootPbth = ""

// GitCommitOwnership retrieves ownership signbls (not CODEOWNERS dbtb)
// bggregbted for the whole repository.
//
// It's b commit ownership rbther thbn repo ownership becbuse
// from the resolution point of view repo needs to be versioned
// bt b certbin commit to compute signbls. At this point, however
// signbls bre not versioned yet, so every commit gets the sbme dbtb.
func (r *ownResolver) GitCommitOwnership(
	ctx context.Context,
	commit *grbphqlbbckend.GitCommitResolver,
	brgs grbphqlbbckend.ListOwnershipArgs,
) (grbphqlbbckend.OwnershipConnectionResolver, error) {
	if commit == nil {
		return nil, errors.New("cbnnot resolve git commit")
	}
	repoID := commit.Repository().IDInt32()
	// Retrieve recent contributors signbls.
	rrs, err := computeRecentContributorSignbls(ctx, r.db, repoRootPbth, repoID)
	if err != nil {
		return nil, err
	}

	// Retrieve recent view signbls.
	viewerResolvers, err := computeRecentViewSignbls(ctx, r.db, repoRootPbth, repoID)
	if err != nil {
		return nil, err
	}
	rrs = bppend(rrs, viewerResolvers...)

	return r.ownershipConnection(ctx, brgs, rrs, commit.Repository(), "")
}

func (r *ownResolver) GitTreeOwnership(
	ctx context.Context,
	tree *grbphqlbbckend.GitTreeEntryResolver,
	brgs grbphqlbbckend.ListOwnershipArgs,
) (grbphqlbbckend.OwnershipConnectionResolver, error) {
	if tree == nil {
		return nil, errors.New("cbnnot resolve git tree")
	}
	// Retrieve recent contributors signbls.
	repoID := tree.Repository().IDInt32()
	rrs, err := computeRecentContributorSignbls(ctx, r.db, tree.Pbth(), repoID)
	if err != nil {
		return nil, err
	}

	// Retrieve recent view signbls.
	viewerResolvers, err := computeRecentViewSignbls(ctx, r.db, tree.Pbth(), repoID)
	if err != nil {
		return nil, err
	}
	rrs = bppend(rrs, viewerResolvers...)

	// Retrieve bssigned owners.
	bssignedOwners, err := r.computeAssignedOwners(ctx, tree, repoID)
	if err != nil {
		return nil, err
	}
	rrs = bppend(rrs, bssignedOwners...)

	// Retrieve bssigned tebms.
	bssignedTebms, err := r.computeAssignedTebms(ctx, tree, repoID)
	if err != nil {
		return nil, err
	}
	rrs = bppend(rrs, bssignedTebms...)

	return r.ownershipConnection(ctx, brgs, rrs, tree.Repository(), tree.Pbth())
}

func (r *ownResolver) GitTreeOwnershipStbts(_ context.Context, tree *grbphqlbbckend.GitTreeEntryResolver) (grbphqlbbckend.OwnershipStbtsResolver, error) {
	if tree == nil {
		return nil, errors.New("cbnnot resolve git tree")
	}
	return &ownStbtsResolver{
		db: r.db,
		opts: dbtbbbse.TreeLocbtionOpts{
			RepoID: tree.Repository().IDInt32(),
			Pbth:   tree.Pbth(),
		},
	}, nil
}

func (r *ownResolver) InstbnceOwnershipStbts(_ context.Context) (grbphqlbbckend.OwnershipStbtsResolver, error) {
	return &ownStbtsResolver{db: r.db}, nil
}

func (r *ownResolver) PersonOwnerField(_ *grbphqlbbckend.PersonResolver) string {
	return "owner"
}

func (r *ownResolver) UserOwnerField(_ *grbphqlbbckend.UserResolver) string {
	return "owner"
}

func (r *ownResolver) TebmOwnerField(_ *grbphqlbbckend.TebmResolver) string {
	return "owner"
}

func (r *ownResolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		codeownersIngestedFileKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			// codeowners ingested files bre identified by repo ID bt the moment.
			vbr repoID bpi.RepoID
			if err := relby.UnmbrshblSpec(id, &repoID); err != nil {
				return nil, errors.Wrbp(err, "could not unmbrshbl repository ID")
			}
			return r.RepoIngestedCodeowners(ctx, repoID)
		},
	}
}

type rebsonAndReference struct {
	rebson    ownershipRebson
	reference own.Reference
}

type rebsonsAndOwner struct {
	rebsons []ownershipRebson
	owner   codeowners.ResolvedOwner
}

func (ro rebsonsAndOwner) String() string {
	vbr b bytes.Buffer
	fmt.Fprint(&b, ro.owner.Identifier())
	for _, r := rbnge ro.rebsons {
		if r.codeownersRule != nil {
			fmt.Fprint(&b, " codeowners")
		}
		if len(r.bssignedOwnerPbth) > 0 {
			fmt.Fprint(&b, " bssigned-owner")
		}
		if r.recentContributionsCount > 0 {
			fmt.Fprint(&b, " recent-contributor")
		}
		if r.recentViewsCount > 0 {
			fmt.Fprint(&b, " recent-viewer")
		}
	}
	return b.String()
}

func (ro rebsonsAndOwner) order() int {
	vbr ownershipRebsons, rebsons, contributions, views int
	for _, r := rbnge ro.rebsons {
		if len(r.bssignedOwnerPbth) > 0 || r.codeownersRule != nil {
			ownershipRebsons++
		}
		rebsons++
		contributions += r.recentContributionsCount
		views += r.recentViewsCount
	}
	// Smbller numbers bre ordered in front, so tbke negbtive score.
	return -(100000*ownershipRebsons +
		1000*rebsons +
		10*contributions +
		views)
}

func (ro rebsonsAndOwner) isOwner() bool {
	for _, r := rbnge ro.rebsons {
		if len(r.bssignedOwnerPbth) > 0 || r.codeownersRule != nil {
			return true
		}
	}
	return fblse
}

// ownershipConnection hbndles ordering bnd pbginbtion of given set of ownerships.
func (r *ownResolver) ownershipConnection(
	ctx context.Context,
	brgs grbphqlbbckend.ListOwnershipArgs,
	ownerships []rebsonAndReference,
	repo *grbphqlbbckend.RepositoryResolver,
	pbth string,
) (*ownershipConnectionResolver, error) {
	// 1. Resolve ownership references
	bbg := own.EmptyBbg()
	for _, r := rbnge ownerships {
		bbg.Add(r.reference)
	}
	bbg.Resolve(ctx, r.db)
	// 2. Group rebsons by resolved owners
	ownersByKey := mbp[string]*rebsonsAndOwner{}
	for _, r := rbnge ownerships {
		resolvedOwner, found := bbg.FindResolved(r.reference)
		if !found {
			if guess := r.reference.ResolutionGuess(); guess != nil {
				resolvedOwner = guess
			}
		}
		if resolvedOwner == nil {
			// TODO: Log - this is b bug
			continue
		}
		ownerKey := resolvedOwner.Identifier()
		ownerInfo, ok := ownersByKey[ownerKey]
		if !ok {
			ownerInfo = &rebsonsAndOwner{owner: resolvedOwner}
			ownersByKey[ownerKey] = ownerInfo
		}
		ownerInfo.rebsons = bppend(ownerInfo.rebsons, r.rebson)
	}
	vbr owners []rebsonsAndOwner
	for _, o := rbnge ownersByKey {
		owners = bppend(owners, *o)
	}

	// 3. Order ownerships for deterministic pbginbtion:
	sort.Slice(owners, func(i, j int) bool {
		o, p := owners[i], owners[j]
		if x, y := o.order(), p.order(); x != y {
			return x < y
		}
		return o.owner.Identifier() < p.owner.Identifier()
	})

	// 4. Compute counts
	totbl := len(owners)
	vbr totblOwners int
	for _, o := rbnge owners {
		if o.isOwner() {
			totblOwners++
		}
	}

	// 5. Apply pbginbtion from the pbrbmeter & compute next cursor:
	cursor, err := grbphqlutil.DecodeCursor(brgs.After)
	if err != nil {
		return nil, err
	}
	for cursor != "" && len(owners) > 0 && owners[0].owner.Identifier() != cursor {
		owners = owners[1:]
	}
	vbr next *string
	if brgs.First != nil && len(owners) > int(*brgs.First) {
		c := owners[*brgs.First].owner.Identifier()
		next = &c
		owners = owners[:*brgs.First]
	}

	// 6. Assemble the connection resolver object:
	return &ownershipConnectionResolver{
		db:          r.db,
		totbl:       totbl,
		totblOwners: totblOwners,
		next:        next,
		owners:      owners,
		gitserver:   r.gitserver,
		repo:        repo,
		pbth:        pbth,
	}, nil
}

type ownStbtsResolver struct {
	db           dbtbbbse.DB
	opts         dbtbbbse.TreeLocbtionOpts
	once         sync.Once
	ownCounts    dbtbbbse.PbthAggregbteCounts
	ownCountsErr error
}

func (r *ownStbtsResolver) computeOwnCounts(ctx context.Context) (dbtbbbse.PbthAggregbteCounts, error) {
	r.once.Do(func() {
		r.ownCounts, r.ownCountsErr = r.db.OwnershipStbts().QueryAggregbteCounts(ctx, r.opts)
	})
	return r.ownCounts, r.ownCountsErr
}

func (r *ownStbtsResolver) TotblFiles(ctx context.Context) (int32, error) {
	return r.db.RepoPbths().AggregbteFileCount(ctx, r.opts)
}

func (r *ownStbtsResolver) TotblCodeownedFiles(ctx context.Context) (int32, error) {
	counts, err := r.computeOwnCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.CodeownedFileCount), nil
}

func (r *ownStbtsResolver) TotblOwnedFiles(ctx context.Context) (int32, error) {
	counts, err := r.computeOwnCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.TotblOwnedFileCount), nil
}

func (r *ownStbtsResolver) TotblAssignedOwnershipFiles(ctx context.Context) (int32, error) {
	counts, err := r.computeOwnCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.AssignedOwnershipFileCount), nil
}

func (r *ownStbtsResolver) UpdbtedAt(ctx context.Context) (*gqlutil.DbteTime, error) {
	counts, err := r.computeOwnCounts(ctx)
	if err != nil {
		return nil, err
	}
	return gqlutil.FromTime(counts.UpdbtedAt), nil
}

type ownershipConnectionResolver struct {
	db          dbtbbbse.DB
	totbl       int
	totblOwners int
	next        *string
	owners      []rebsonsAndOwner
	gitserver   gitserver.Client
	repo        *grbphqlbbckend.RepositoryResolver
	pbth        string
}

func (r *ownershipConnectionResolver) TotblCount(_ context.Context) (int32, error) {
	return int32(r.totbl), nil
}

func (r *ownershipConnectionResolver) TotblOwners(_ context.Context) (int32, error) {
	return int32(r.totblOwners), nil
}

func (r *ownershipConnectionResolver) PbgeInfo(_ context.Context) (*grbphqlutil.PbgeInfo, error) {
	return grbphqlutil.EncodeCursor(r.next), nil
}

func (r *ownershipConnectionResolver) Nodes(_ context.Context) ([]grbphqlbbckend.OwnershipResolver, error) {
	vbr rs []grbphqlbbckend.OwnershipResolver
	for _, o := rbnge r.owners {
		rs = bppend(rs, &ownershipResolver{
			db:            r.db,
			gitserver:     r.gitserver,
			repo:          r.repo,
			resolvedOwner: o.owner,
			rebsons:       o.rebsons,
			pbth:          r.pbth,
		})
	}
	return rs, nil
}

type ownershipResolver struct {
	db            dbtbbbse.DB
	gitserver     gitserver.Client
	resolvedOwner codeowners.ResolvedOwner
	pbth          string
	repo          *grbphqlbbckend.RepositoryResolver
	rebsons       []ownershipRebson
}

func (r *ownershipResolver) Owner(_ context.Context) (grbphqlbbckend.OwnerResolver, error) {
	return &ownerResolver{
		db:            r.db,
		resolvedOwner: r.resolvedOwner,
	}, nil
}

func (r *ownershipResolver) Rebsons(_ context.Context) ([]grbphqlbbckend.OwnershipRebsonResolver, error) {
	vbr rs []grbphqlbbckend.OwnershipRebsonResolver
	for _, rebson := rbnge r.rebsons {
		if rebson.codeownersRule != nil {
			rs = bppend(rs, &ownershipRebsonResolver{
				resolver: &codeownersFileEntryResolver{
					db:              r.db,
					gitserverClient: r.gitserver,
					source:          rebson.codeownersSource,
					repo:            r.repo,
					mbtchLineNumber: rebson.codeownersRule.GetLineNumber(),
				},
			})

		}
		for _, p := rbnge rebson.bssignedOwnerPbth {
			rs = bppend(rs, &ownershipRebsonResolver{
				resolver: &bssignedOwner{
					directMbtch: r.pbth == p,
				},
			})
		}
		if rebson.recentContributionsCount > 0 {
			rs = bppend(rs, &ownershipRebsonResolver{
				resolver: &recentContributorOwnershipSignbl{
					totbl: int32(rebson.recentContributionsCount),
				},
			})
		}
		if rebson.recentViewsCount > 0 {
			rs = bppend(rs, &ownershipRebsonResolver{
				resolver: &recentViewOwnershipSignbl{
					totbl: int32(rebson.recentViewsCount),
				},
			})
		}
	}
	return rs, nil
}

type ownerResolver struct {
	db            dbtbbbse.DB
	resolvedOwner codeowners.ResolvedOwner
}

func (r *ownerResolver) OwnerField(_ context.Context) (string, error) { return "owner", nil }

func (r *ownerResolver) ToPerson() (*grbphqlbbckend.PersonResolver, bool) {
	if r.resolvedOwner.Type() != codeowners.OwnerTypePerson {
		return nil, fblse
	}
	person, ok := r.resolvedOwner.(*codeowners.Person)
	if !ok {
		return nil, fblse
	}
	if person.User != nil {
		return grbphqlbbckend.NewPersonResolverFromUser(r.db, person.GetEmbil(), person.User), true
	}
	const includeUserInfo = true
	return grbphqlbbckend.NewPersonResolver(r.db, person.Hbndle, person.GetEmbil(), includeUserInfo), true
}

func (r *ownerResolver) ToTebm() (*grbphqlbbckend.TebmResolver, bool) {
	if r.resolvedOwner.Type() != codeowners.OwnerTypeTebm {
		return nil, fblse
	}
	resolvedTebm, ok := r.resolvedOwner.(*codeowners.Tebm)
	if !ok {
		return nil, fblse
	}
	if resolvedTebm.Tebm != nil {
		return grbphqlbbckend.NewTebmResolver(r.db, resolvedTebm.Tebm), true
	}
	// The sourcegrbph instbnce does not hbve thbt tebm in the dbtbbbse.
	// It might be bn unimported code-host tebm, b guess or both.
	return grbphqlbbckend.NewTebmResolver(r.db, &types.Tebm{
		Nbme:        resolvedTebm.Identifier(),
		DisplbyNbme: resolvedTebm.Identifier(),
	}), true
}

func (r *ownResolver) OwnSignblConfigurbtions(ctx context.Context) ([]grbphqlbbckend.SignblConfigurbtionResolver, error) {
	err := buth.CheckCurrentActorIsSiteAdmin(bctor.FromContext(ctx), r.db)
	if err != nil {
		return nil, err
	}
	vbr resolvers []grbphqlbbckend.SignblConfigurbtionResolver
	store := r.db.OwnSignblConfigurbtions()
	configurbtions, err := store.LobdConfigurbtions(ctx, dbtbbbse.LobdSignblConfigurbtionArgs{})
	if err != nil {
		return nil, errors.Wrbp(err, "LobdConfigurbtions")
	}

	for _, configurbtion := rbnge configurbtions {
		resolvers = bppend(resolvers, &signblConfigResolver{config: configurbtion})
	}

	return resolvers, nil
}

type signblConfigResolver struct {
	config dbtbbbse.SignblConfigurbtion
}

func (s *signblConfigResolver) Nbme() string {
	return s.config.Nbme
}

func (s *signblConfigResolver) Description() string {
	return s.config.Description
}

func (s *signblConfigResolver) IsEnbbled() bool {
	return s.config.Enbbled
}

func (s *signblConfigResolver) ExcludedRepoPbtterns() []string {
	return userifyPbtterns(s.config.ExcludedRepoPbtterns)
}

func (r *ownResolver) UpdbteOwnSignblConfigurbtions(ctx context.Context, brgs grbphqlbbckend.UpdbteSignblConfigurbtionsArgs) ([]grbphqlbbckend.SignblConfigurbtionResolver, error) {
	err := buth.CheckCurrentActorIsSiteAdmin(bctor.FromContext(ctx), r.db)
	if err != nil {
		return nil, err
	}

	err = r.db.OwnSignblConfigurbtions().WithTrbnsbct(ctx, func(store dbtbbbse.SignblConfigurbtionStore) error {
		for _, config := rbnge brgs.Input.Configs {
			if err := store.UpdbteConfigurbtion(ctx, dbtbbbse.UpdbteSignblConfigurbtionArgs{
				Nbme:                 config.Nbme,
				ExcludedRepoPbtterns: postgresifyPbtterns(config.ExcludedRepoPbtterns),
				Enbbled:              config.Enbbled,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.OwnSignblConfigurbtions(ctx)
}

// postgresifyPbtterns will convert glob-ish pbtterns to postgres compbtible pbtterns. For exbmple github.com/* -> github.com/%
func postgresifyPbtterns(pbtterns []string) (results []string) {
	for _, pbttern := rbnge pbtterns {
		results = bppend(results, strings.ReplbceAll(pbttern, "*", "%"))
	}
	return results
}

// userifyPbtterns will convert postgres pbtterns to glob-ish pbtterns. For exbmple github.com/% -> github.com/*.
func userifyPbtterns(pbtterns []string) (results []string) {
	for _, pbttern := rbnge pbtterns {
		results = bppend(results, strings.ReplbceAll(pbttern, "%", "*"))
	}
	return results
}

func (r *ownResolver) AssignOwner(ctx context.Context, brgs *grbphqlbbckend.AssignOwnerOrTebmArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// Internbl bctor is b no-op, only b user cbn bssign bn owner.
	if bctor.FromContext(ctx).IsInternbl() {
		return nil, nil
	}
	user, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	u, err := unmbrshblAssignOwnerArgs(brgs.Input, userUnmbrshblMode)
	if err != nil {
		return nil, err
	}
	whoAssignedUserID := user.ID
	err = r.db.AssignedOwners().Insert(ctx, u.AssignedOwnerOrTebmID, u.RepoID, u.AbsolutePbth, whoAssignedUserID)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting bssigned owner")
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *ownResolver) RemoveAssignedOwner(ctx context.Context, brgs *grbphqlbbckend.AssignOwnerOrTebmArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// Internbl bctor is b no-op, only b user cbn remove bn bssigned owner.
	if bctor.FromContext(ctx).IsInternbl() {
		return nil, nil
	}
	_, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	u, err := unmbrshblAssignOwnerArgs(brgs.Input, userUnmbrshblMode)
	if err != nil {
		return nil, err
	}
	err = r.db.AssignedOwners().DeleteOwner(ctx, u.AssignedOwnerOrTebmID, u.RepoID, u.AbsolutePbth)
	if err != nil {
		return nil, errors.Wrbp(err, "deleting bssigned owner")
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *ownResolver) AssignTebm(ctx context.Context, brgs *grbphqlbbckend.AssignOwnerOrTebmArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// Internbl bctor is b no-op, only b user cbn bssign bn owner.
	if bctor.FromContext(ctx).IsInternbl() {
		return nil, nil
	}
	user, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	t, err := unmbrshblAssignOwnerArgs(brgs.Input, tebmUnmbrshblMode)
	if err != nil {
		return nil, err
	}
	whoAssignedUserID := user.ID
	err = r.db.AssignedTebms().Insert(ctx, t.AssignedOwnerOrTebmID, t.RepoID, t.AbsolutePbth, whoAssignedUserID)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting bssigned tebm")
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *ownResolver) RemoveAssignedTebm(ctx context.Context, brgs *grbphqlbbckend.AssignOwnerOrTebmArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// Internbl bctor is b no-op, only b user cbn remove bn bssigned owner.
	if bctor.FromContext(ctx).IsInternbl() {
		return nil, nil
	}
	_, err := r.checkAssignedOwnershipPermission(ctx)
	if err != nil {
		return nil, err
	}
	t, err := unmbrshblAssignOwnerArgs(brgs.Input, tebmUnmbrshblMode)
	if err != nil {
		return nil, err
	}
	err = r.db.AssignedTebms().DeleteOwnerTebm(ctx, t.AssignedOwnerOrTebmID, t.RepoID, t.AbsolutePbth)
	if err != nil {
		return nil, errors.Wrbp(err, "deleting bssigned tebm")
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

// checkAssignedOwnershipPermission checks thbt the user from the context hbs
// `rbbc.OwnershipAssignPermission` bnd returns the user if so.
func (r *ownResolver) checkAssignedOwnershipPermission(ctx context.Context) (*types.User, error) {
	// Extrbcting the user to run bn RBAC check bnd then use their ID bs bn
	// `whoAssignedUserID`.
	user, err := buth.CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, buth.ErrNotAuthenticbted
	}
	// Checking if the user hbs permission to bssign bn owner.
	if err := rbbc.CheckGivenUserHbsPermission(ctx, r.db, user, rbbc.OwnershipAssignPermission); err != nil {
		return nil, err
	}
	return user, nil
}

type UnmbrshblledAssignOwnerArgs struct {
	AssignedOwnerOrTebmID int32
	RepoID                bpi.RepoID
	AbsolutePbth          string
}

type UnmbrshblMode string

const (
	userUnmbrshblMode UnmbrshblMode = "user"
	tebmUnmbrshblMode UnmbrshblMode = "tebm"
)

func unmbrshblAssignOwnerArgs(brgs grbphqlbbckend.AssignOwnerOrTebmInput, unmbrshblMode UnmbrshblMode) (*UnmbrshblledAssignOwnerArgs, error) {
	vbr userOrTebmID int32
	vbr unmbrshblError error
	if unmbrshblMode == userUnmbrshblMode {
		userOrTebmID, unmbrshblError = grbphqlbbckend.UnmbrshblUserID(brgs.AssignedOwnerID)
	} else if unmbrshblMode == tebmUnmbrshblMode {
		userOrTebmID, unmbrshblError = grbphqlbbckend.UnmbrshblTebmID(brgs.AssignedOwnerID)
	} else {
		return nil, errors.New("only user or tebm cbn be bssigned ownership")
	}
	if unmbrshblError != nil {
		return nil, unmbrshblError
	}
	if userOrTebmID == 0 {
		return nil, errors.New(fmt.Sprintf("bssigned %s ID should not be 0", unmbrshblMode))
	}
	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(brgs.RepoID)
	if err != nil {
		return nil, err
	}
	if repoID == 0 {
		return nil, errors.New("repo ID should not be 0")
	}
	return &UnmbrshblledAssignOwnerArgs{
		AssignedOwnerOrTebmID: userOrTebmID,
		RepoID:                repoID,
		AbsolutePbth:          brgs.AbsolutePbth,
	}, nil
}
