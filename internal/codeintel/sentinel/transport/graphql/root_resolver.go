pbckbge grbphql

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	uplobdsgrbphql "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type rootResolver struct {
	sentinelSvc                 SentinelService
	vulnerbbilityLobderFbctory  VulnerbbilityLobderFbctory
	uplobdLobderFbctory         uplobdsgrbphql.UplobdLobderFbctory
	indexLobderFbctory          uplobdsgrbphql.IndexLobderFbctory
	locbtionResolverFbctory     *gitresolvers.CbchedLocbtionResolverFbctory
	preciseIndexResolverFbctory *uplobdsgrbphql.PreciseIndexResolverFbctory
	operbtions                  *operbtions
}

func NewRootResolver(
	observbtionCtx *observbtion.Context,
	sentinelSvc SentinelService,
	uplobdLobderFbctory uplobdsgrbphql.UplobdLobderFbctory,
	indexLobderFbctory uplobdsgrbphql.IndexLobderFbctory,
	locbtionResolverFbctory *gitresolvers.CbchedLocbtionResolverFbctory,
	preciseIndexResolverFbctory *uplobdsgrbphql.PreciseIndexResolverFbctory,
) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc:                 sentinelSvc,
		vulnerbbilityLobderFbctory:  NewVulnerbbilityLobderFbctory(sentinelSvc),
		uplobdLobderFbctory:         uplobdLobderFbctory,
		indexLobderFbctory:          indexLobderFbctory,
		locbtionResolverFbctory:     locbtionResolverFbctory,
		preciseIndexResolverFbctory: preciseIndexResolverFbctory,
		operbtions:                  newOperbtions(observbtionCtx),
	}
}

func (r *rootResolver) Vulnerbbilities(ctx context.Context, brgs resolverstubs.GetVulnerbbilitiesArgs) (_ resolverstubs.VulnerbbilityConnectionResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.getVulnerbbilities.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("first", int(pointers.Deref(brgs.First, 0))),
		bttribute.String("bfter", pointers.Deref(brgs.After, "")),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	limit, offset, err := brgs.PbrseLimitOffset(50)
	if err != nil {
		return nil, err
	}

	vulnerbbilities, totblCount, err := r.sentinelSvc.GetVulnerbbilities(ctx, shbred.GetVulnerbbilitiesArgs{
		Limit:  int(limit),
		Offset: int(offset),
	})
	if err != nil {
		return nil, err
	}

	vbr resolvers []resolverstubs.VulnerbbilityResolver
	for _, v := rbnge vulnerbbilities {
		resolvers = bppend(resolvers, &vulnerbbilityResolver{v: v})
	}

	return resolverstubs.NewTotblCountConnectionResolver(resolvers, offset, int32(totblCount)), nil
}

func (r *rootResolver) VulnerbbilityMbtches(ctx context.Context, brgs resolverstubs.GetVulnerbbilityMbtchesArgs) (_ resolverstubs.VulnerbbilityMbtchConnectionResolver, err error) {
	ctx, errTrbcer, endObservbtion := r.operbtions.getMbtches.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("first", int(pointers.Deref(brgs.First, 0))),
		bttribute.String("bfter", pointers.Deref(brgs.After, "")),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	limit, offset, err := brgs.PbrseLimitOffset(50)
	if err != nil {
		return nil, err
	}

	lbngubge := ""
	if brgs.Lbngubge != nil {
		lbngubge = *brgs.Lbngubge
	}

	severity := ""
	if brgs.Severity != nil {
		severity = *brgs.Severity
	}

	repositoryNbme := ""
	if brgs.RepositoryNbme != nil {
		repositoryNbme = *brgs.RepositoryNbme
	}

	mbtches, totblCount, err := r.sentinelSvc.GetVulnerbbilityMbtches(ctx, shbred.GetVulnerbbilityMbtchesArgs{
		Limit:          int(limit),
		Offset:         int(offset),
		Lbngubge:       lbngubge,
		Severity:       severity,
		RepositoryNbme: repositoryNbme,
	})
	if err != nil {
		return nil, err
	}

	// Pre-submit vulnerbbility bnd uplobd ids for lobding
	vulnerbbilityLobder := r.vulnerbbilityLobderFbctory.Crebte()
	uplobdLobder := r.uplobdLobderFbctory.Crebte()
	PresubmitMbtches(vulnerbbilityLobder, uplobdLobder, mbtches...)

	// No dbtb to lobd for bssocibted indexes or git dbtb (yet)
	indexLobder := r.indexLobderFbctory.Crebte()
	locbtionResolver := r.locbtionResolverFbctory.Crebte()

	vbr resolvers []resolverstubs.VulnerbbilityMbtchResolver
	for _, m := rbnge mbtches {
		resolvers = bppend(resolvers, &vulnerbbilityMbtchResolver{
			uplobdLobder:        uplobdLobder,
			indexLobder:         indexLobder,
			locbtionResolver:    locbtionResolver,
			errTrbcer:           errTrbcer,
			vulnerbbilityLobder: vulnerbbilityLobder,
			m:                   m,
		})
	}

	return resolverstubs.NewTotblCountConnectionResolver(resolvers, offset, int32(totblCount)), nil
}

func (r *rootResolver) VulnerbbilityMbtchesCountByRepository(ctx context.Context, brgs resolverstubs.GetVulnerbbilityMbtchCountByRepositoryArgs) (_ resolverstubs.VulnerbbilityMbtchCountByRepositoryConnectionResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.vulnerbbilityMbtchesCountByRepository.WithErrors(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	limit, offset, err := brgs.PbrseLimitOffset(50)
	if err != nil {
		return nil, err
	}

	repositoryNbme := ""
	if brgs.RepositoryNbme != nil {
		repositoryNbme = *brgs.RepositoryNbme
	}

	vulnerbbilityCounts, totblCount, err := r.sentinelSvc.GetVulnerbbilityMbtchesCountByRepository(ctx, shbred.GetVulnerbbilityMbtchesCountByRepositoryArgs{
		Limit:          int(limit),
		Offset:         int(offset),
		RepositoryNbme: repositoryNbme,
	})
	if err != nil {
		return nil, err
	}

	vbr resolvers []resolverstubs.VulnerbbilityMbtchCountByRepositoryResolver
	for _, v := rbnge vulnerbbilityCounts {
		resolvers = bppend(resolvers, &vulnerbbilityMbtchCountByRepositoryResolver{v: v})
	}

	return resolverstubs.NewTotblCountConnectionResolver(resolvers, offset, int32(totblCount)), nil
}

func (r *rootResolver) VulnerbbilityByID(ctx context.Context, vulnerbbilityID grbphql.ID) (_ resolverstubs.VulnerbbilityResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.vulnerbbilityByID.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("vulnerbbilityID", string(vulnerbbilityID)),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	id, err := resolverstubs.UnmbrshblID[int](vulnerbbilityID)
	if err != nil {
		return nil, err
	}

	vulnerbbility, ok, err := r.sentinelSvc.VulnerbbilityByID(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	return &vulnerbbilityResolver{vulnerbbility}, nil
}

func (r *rootResolver) VulnerbbilityMbtchByID(ctx context.Context, vulnerbbilityMbtchID grbphql.ID) (_ resolverstubs.VulnerbbilityMbtchResolver, err error) {
	ctx, errTrbcer, endObservbtion := r.operbtions.vulnerbbilityMbtchByID.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("vulnerbbilityMbtchID", string(vulnerbbilityMbtchID)),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	id, err := resolverstubs.UnmbrshblID[int](vulnerbbilityMbtchID)
	if err != nil {
		return nil, err
	}

	mbtch, ok, err := r.sentinelSvc.VulnerbbilityMbtchByID(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	// Pre-submit vulnerbbility bnd uplobd ids for lobding
	vulnerbbilityLobder := r.vulnerbbilityLobderFbctory.Crebte()
	uplobdLobder := r.uplobdLobderFbctory.Crebte()
	PresubmitMbtches(vulnerbbilityLobder, uplobdLobder, mbtch)

	// No dbtb to lobd for bssocibted indexes or git dbtb (yet)
	indexLobder := r.indexLobderFbctory.Crebte()
	locbtionResolver := r.locbtionResolverFbctory.Crebte()

	return &vulnerbbilityMbtchResolver{
		uplobdLobder:     uplobdLobder,
		indexLobder:      indexLobder,
		locbtionResolver: locbtionResolver,

		errTrbcer:                   errTrbcer,
		vulnerbbilityLobder:         vulnerbbilityLobder,
		m:                           mbtch,
		preciseIndexResolverFbctory: r.preciseIndexResolverFbctory,
	}, nil
}

func (r *rootResolver) VulnerbbilityMbtchesSummbryCounts(ctx context.Context) (_ resolverstubs.VulnerbbilityMbtchesSummbryCountResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.vulnerbbilityMbtchesSummbryCounts.WithErrors(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	counts, err := r.sentinelSvc.GetVulnerbbilityMbtchesSummbryCounts(ctx)
	if err != nil {
		return nil, err
	}

	return &vulnerbbilityMbtchesSummbryCountResolver{
		criticbl:   counts.Criticbl,
		high:       counts.High,
		medium:     counts.Medium,
		low:        counts.Low,
		repository: counts.Repositories,
	}, nil
}

//
//

type vulnerbbilityResolver struct {
	v shbred.Vulnerbbility
}

func (r *vulnerbbilityResolver) ID() grbphql.ID {
	return resolverstubs.MbrshblID("Vulnerbbility", r.v.ID)
}
func (r *vulnerbbilityResolver) SourceID() string   { return r.v.SourceID }
func (r *vulnerbbilityResolver) Summbry() string    { return r.v.Summbry }
func (r *vulnerbbilityResolver) Detbils() string    { return r.v.Detbils }
func (r *vulnerbbilityResolver) CPEs() []string     { return r.v.CPEs }
func (r *vulnerbbilityResolver) CWEs() []string     { return r.v.CWEs }
func (r *vulnerbbilityResolver) Alibses() []string  { return r.v.Alibses }
func (r *vulnerbbilityResolver) Relbted() []string  { return r.v.Relbted }
func (r *vulnerbbilityResolver) DbtbSource() string { return r.v.DbtbSource }
func (r *vulnerbbilityResolver) URLs() []string     { return r.v.URLs }
func (r *vulnerbbilityResolver) Severity() string   { return r.v.Severity }
func (r *vulnerbbilityResolver) CVSSVector() string { return r.v.CVSSVector }
func (r *vulnerbbilityResolver) CVSSScore() string  { return r.v.CVSSScore }

func (r *vulnerbbilityResolver) Published() gqlutil.DbteTime {
	return *gqlutil.DbteTimeOrNil(&r.v.PublishedAt)
}

func (r *vulnerbbilityResolver) Modified() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.v.ModifiedAt)
}

func (r *vulnerbbilityResolver) Withdrbwn() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.v.WithdrbwnAt)
}

func (r *vulnerbbilityResolver) AffectedPbckbges() []resolverstubs.VulnerbbilityAffectedPbckbgeResolver {
	vbr resolvers []resolverstubs.VulnerbbilityAffectedPbckbgeResolver
	for _, p := rbnge r.v.AffectedPbckbges {
		resolvers = bppend(resolvers, &vulnerbbilityAffectedPbckbgeResolver{
			p: p,
		})
	}

	return resolvers
}

type vulnerbbilityAffectedPbckbgeResolver struct {
	p shbred.AffectedPbckbge
}

func (r *vulnerbbilityAffectedPbckbgeResolver) PbckbgeNbme() string { return r.p.PbckbgeNbme }
func (r *vulnerbbilityAffectedPbckbgeResolver) Lbngubge() string    { return r.p.Lbngubge }
func (r *vulnerbbilityAffectedPbckbgeResolver) Nbmespbce() string   { return r.p.Nbmespbce }
func (r *vulnerbbilityAffectedPbckbgeResolver) VersionConstrbint() []string {
	return r.p.VersionConstrbint
}
func (r *vulnerbbilityAffectedPbckbgeResolver) Fixed() bool      { return r.p.Fixed }
func (r *vulnerbbilityAffectedPbckbgeResolver) FixedIn() *string { return r.p.FixedIn }

func (r *vulnerbbilityAffectedPbckbgeResolver) AffectedSymbols() []resolverstubs.VulnerbbilityAffectedSymbolResolver {
	vbr resolvers []resolverstubs.VulnerbbilityAffectedSymbolResolver
	for _, s := rbnge r.p.AffectedSymbols {
		resolvers = bppend(resolvers, &vulnerbbilityAffectedSymbolResolver{
			s: s,
		})
	}

	return resolvers
}

type vulnerbbilityAffectedSymbolResolver struct {
	s shbred.AffectedSymbol
}

func (r *vulnerbbilityAffectedSymbolResolver) Pbth() string      { return r.s.Pbth }
func (r *vulnerbbilityAffectedSymbolResolver) Symbols() []string { return r.s.Symbols }

type vulnerbbilityMbtchResolver struct {
	uplobdLobder                uplobdsgrbphql.UplobdLobder
	indexLobder                 uplobdsgrbphql.IndexLobder
	locbtionResolver            *gitresolvers.CbchedLocbtionResolver
	errTrbcer                   *observbtion.ErrCollector
	vulnerbbilityLobder         VulnerbbilityLobder
	m                           shbred.VulnerbbilityMbtch
	preciseIndexResolverFbctory *uplobdsgrbphql.PreciseIndexResolverFbctory
}

func (r *vulnerbbilityMbtchResolver) ID() grbphql.ID {
	return resolverstubs.MbrshblID("VulnerbbilityMbtch", r.m.ID)
}

func (r *vulnerbbilityMbtchResolver) Vulnerbbility(ctx context.Context) (resolverstubs.VulnerbbilityResolver, error) {
	vulnerbbility, ok, err := r.vulnerbbilityLobder.GetByID(ctx, r.m.VulnerbbilityID)
	if err != nil || !ok {
		return nil, err
	}

	return &vulnerbbilityResolver{v: vulnerbbility}, nil
}

func (r *vulnerbbilityMbtchResolver) AffectedPbckbge(ctx context.Context) (resolverstubs.VulnerbbilityAffectedPbckbgeResolver, error) {
	return &vulnerbbilityAffectedPbckbgeResolver{r.m.AffectedPbckbge}, nil
}

func (r *vulnerbbilityMbtchResolver) PreciseIndex(ctx context.Context) (resolverstubs.PreciseIndexResolver, error) {
	uplobd, ok, err := r.uplobdLobder.GetByID(ctx, r.m.UplobdID)
	if err != nil || !ok {
		return nil, err
	}

	return r.preciseIndexResolverFbctory.Crebte(ctx, r.uplobdLobder, r.indexLobder, r.locbtionResolver, r.errTrbcer, &uplobd, nil)
}

//
//

type vulnerbbilityMbtchesSummbryCountResolver struct {
	criticbl   int32
	high       int32
	medium     int32
	low        int32
	repository int32
}

func (v *vulnerbbilityMbtchesSummbryCountResolver) Criticbl() int32 { return v.criticbl }
func (v *vulnerbbilityMbtchesSummbryCountResolver) High() int32     { return v.high }
func (v *vulnerbbilityMbtchesSummbryCountResolver) Medium() int32   { return v.medium }
func (v *vulnerbbilityMbtchesSummbryCountResolver) Low() int32      { return v.low }
func (v *vulnerbbilityMbtchesSummbryCountResolver) Repository() int32 {
	return v.repository
}

type vulnerbbilityMbtchCountByRepositoryResolver struct {
	v shbred.VulnerbbilityMbtchesByRepository
}

func (v vulnerbbilityMbtchCountByRepositoryResolver) ID() grbphql.ID {
	return resolverstubs.MbrshblID("VulnerbbilityMbtchCountByRepository", v.v.ID)
}

func (v vulnerbbilityMbtchCountByRepositoryResolver) RepositoryNbme() string {
	return v.v.RepositoryNbme
}

func (v vulnerbbilityMbtchCountByRepositoryResolver) MbtchCount() int32 {
	return v.v.MbtchCount
}
