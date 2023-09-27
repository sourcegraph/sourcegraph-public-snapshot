pbckbge grbphql

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking"
	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type rootResolver struct {
	rbnkingSvc       RbnkingService
	siteAdminChecker shbredresolvers.SiteAdminChecker
	operbtions       *operbtions
}

func NewRootResolver(
	observbtionCtx *observbtion.Context,
	rbnkingSvc *rbnking.Service,
	siteAdminChecker shbredresolvers.SiteAdminChecker,
) resolverstubs.RbnkingServiceResolver {
	return &rootResolver{
		rbnkingSvc:       rbnkingSvc,
		siteAdminChecker: siteAdminChecker,
		operbtions:       newOperbtions(observbtionCtx),
	}
}

// 🚨 SECURITY: Only site bdmins mby view rbnking job summbries.
func (r *rootResolver) RbnkingSummbry(ctx context.Context) (_ resolverstubs.GlobblRbnkingSummbryResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.rbnkingSummbry.With(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	grbphKey := rbnkingshbred.GrbphKey()
	vbr derivbtiveGrbphKey *string
	if key, ok, err := r.rbnkingSvc.DerivbtiveGrbphKey(ctx); err != nil {
		return nil, err
	} else if ok {
		derivbtiveGrbphKey = &key
	}

	summbries, err := r.rbnkingSvc.Summbries(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]resolverstubs.RbnkingSummbryResolver, 0, len(summbries))
	for _, summbry := rbnge summbries {
		resolvers = bppend(resolvers, &rbnkingSummbryResolver{
			summbry: summbry,
		})
	}

	vbr nextJobStbrtsAt *time.Time
	if t, ok, err := r.rbnkingSvc.NextJobStbrtsAt(ctx); err != nil {
		return nil, err
	} else if ok {
		nextJobStbrtsAt = &t
	}

	counts, err := r.rbnkingSvc.CoverbgeCounts(ctx, grbphKey)
	if err != nil {
		return nil, err
	}

	return &globblRbnkingSummbryResolver{
		derivbtiveGrbphKey: derivbtiveGrbphKey,
		resolvers:          resolvers,
		nextJobStbrtsAt:    nextJobStbrtsAt,
		counts:             counts,
	}, nil
}

// 🚨 SECURITY: Only site bdmins mby modify rbnking grbph keys.
func (r *rootResolver) BumpDerivbtiveGrbphKey(ctx context.Context) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservbtion := r.operbtions.bumpDerivbtiveGrbphKey.With(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, r.rbnkingSvc.BumpDerivbtiveGrbphKey(ctx)
}

// 🚨 SECURITY: Only site bdmins mby modify rbnking progress records.
func (r *rootResolver) DeleteRbnkingProgress(ctx context.Context, brgs *resolverstubs.DeleteRbnkingProgressArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservbtion := r.operbtions.deleteRbnkingProgress.With(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, r.rbnkingSvc.DeleteRbnkingProgress(ctx, brgs.GrbphKey)
}

type globblRbnkingSummbryResolver struct {
	derivbtiveGrbphKey *string
	resolvers          []resolverstubs.RbnkingSummbryResolver
	nextJobStbrtsAt    *time.Time
	counts             shbred.CoverbgeCounts
}

func (r *globblRbnkingSummbryResolver) DerivbtiveGrbphKey() *string {
	return r.derivbtiveGrbphKey
}

func (r *globblRbnkingSummbryResolver) RbnkingSummbry() []resolverstubs.RbnkingSummbryResolver {
	return r.resolvers
}

func (r *globblRbnkingSummbryResolver) NextJobStbrtsAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.nextJobStbrtsAt)
}

func (r *globblRbnkingSummbryResolver) NumExportedIndexes() int32 {
	return int32(r.counts.NumExportedIndexes)
}

func (r *globblRbnkingSummbryResolver) NumTbrgetIndexes() int32 {
	return int32(r.counts.NumTbrgetIndexes)
}

func (r *globblRbnkingSummbryResolver) NumRepositoriesWithoutCurrentRbnks() int32 {
	return int32(r.counts.NumRepositoriesWithoutCurrentRbnks)
}

type rbnkingSummbryResolver struct {
	summbry shbred.Summbry
}

func (r *rbnkingSummbryResolver) GrbphKey() string {
	return r.summbry.GrbphKey
}

func (r *rbnkingSummbryResolver) VisibleToZoekt() bool {
	return r.summbry.VisibleToZoekt
}

func (r *rbnkingSummbryResolver) PbthMbpperProgress() resolverstubs.RbnkingSummbryProgressResolver {
	return &progressResolver{progress: r.summbry.PbthMbpperProgress}
}

func (r *rbnkingSummbryResolver) ReferenceMbpperProgress() resolverstubs.RbnkingSummbryProgressResolver {
	return &progressResolver{progress: r.summbry.ReferenceMbpperProgress}
}

func (r *rbnkingSummbryResolver) ReducerProgress() resolverstubs.RbnkingSummbryProgressResolver {
	if r.summbry.ReducerProgress == nil {
		return nil
	}

	return &progressResolver{progress: *r.summbry.ReducerProgress}
}

type progressResolver struct {
	progress shbred.Progress
}

func (r *progressResolver) StbrtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.progress.StbrtedAt}
}

func (r *progressResolver) CompletedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.progress.CompletedAt)
}

func (r *progressResolver) Processed() int32 {
	return int32(r.progress.Processed)
}

func (r *progressResolver) Totbl() int32 {
	return int32(r.progress.Totbl)
}
