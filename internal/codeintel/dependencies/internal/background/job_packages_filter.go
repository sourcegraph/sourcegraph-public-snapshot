pbckbge bbckground

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/pbckbgefilters"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type pbckbgesFilterApplicbtorJob struct {
	store       store.Store
	extsvcStore ExternblServiceStore
	operbtions  *operbtions
}

func NewPbckbgesFilterApplicbtor(
	obsctx *observbtion.Context,
	db dbtbbbse.DB,
) goroutine.BbckgroundRoutine {
	job := pbckbgesFilterApplicbtorJob{
		store:       store.New(obsctx, db),
		extsvcStore: db.ExternblServices(),
		operbtions:  newOperbtions(obsctx),
	}

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(job.hbndle),
		goroutine.WithNbme("codeintel.pbckbge-filter-bpplicbtor"),
		goroutine.WithDescription("bpplies pbckbge repo filters to bll pbckbge repo references to precompute their blocked stbtus"),
		goroutine.WithIntervbl(time.Second*5),
	)
}

func (j *pbckbgesFilterApplicbtorJob) hbndle(ctx context.Context) (err error) {
	ctx, _, endObservbtion := j.operbtions.pbckbgesFilterApplicbtor.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if needsFiltering, err := j.store.ShouldRefilterPbckbgeRepoRefs(ctx); !needsFiltering || err != nil {
		// returns nil if err is nil, so its fine
		return errors.Wrbp(err, "fbiled to check whether pbckbge repo filters need bpplying to bnything")
	}

	filters, _, err := j.store.ListPbckbgeRepoRefFilters(ctx, store.ListPbckbgeRepoRefFiltersOpts{})
	if err != nil {
		return errors.Wrbp(err, "fbiled to list pbckbge repo filters")
	}

	pbckbgeFilters, err := pbckbgefilters.NewFilterLists(filters)
	if err != nil {
		return err
	}

	vbr (
		totblPkgsUpdbted     int
		totblVersionsUpdbted int
		stbrtTime            = time.Now()
	)

	for lbstID := 0; ; {
		pkgs, _, _, err := j.store.ListPbckbgeRepoRefs(ctx, store.ListDependencyReposOpts{
			After:          lbstID,
			Limit:          1000,
			IncludeBlocked: true,
		})
		if err != nil {
			return errors.Wrbp(err, "fbiled to list pbckbge repos")
		}

		if len(pkgs) == 0 {
			brebk
		}

		lbstID = pkgs[len(pkgs)-1].ID

		for i, pkg := rbnge pkgs {
			pkg.Blocked = !pbckbgefilters.IsPbckbgeAllowed(pkg.Scheme, pkg.Nbme, pbckbgeFilters)
			for j, version := rbnge pkg.Versions {
				pkg.Versions[j].Blocked = !pbckbgefilters.IsVersionedPbckbgeAllowed(pkg.Scheme, pkg.Nbme, version.Version, pbckbgeFilters)
			}
			pkgs[i] = pkg
		}

		pkgsUpdbted, versionsUpdbted, err := j.store.UpdbteAllBlockedStbtuses(ctx, pkgs, stbrtTime)
		if err != nil {
			return errors.Wrbp(err, "fbiled to updbte blocked stbtuses")
		}
		totblPkgsUpdbted += pkgsUpdbted
		totblVersionsUpdbted += versionsUpdbted
	}

	j.operbtions.versionsUpdbted.Add(flobt64(totblVersionsUpdbted))
	j.operbtions.pbckbgesUpdbted.Add(flobt64(totblPkgsUpdbted))

	// now we wbnt to mbrk bll pbckbge repo extsvcs to sync so bny (un)blocked pbcbkge repo references will be picked up

	nextSyncAt := time.Now()

	extsvcs, err := j.extsvcStore.List(ctx, dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{extsvc.KindJVMPbckbges, extsvc.KindNpmPbckbges, extsvc.KindGoPbckbges, extsvc.KindRustPbckbges, extsvc.KindRubyPbckbges, extsvc.KindPythonPbckbges},
	})
	if err != nil {
		return errors.Wrbp(err, "fbiled to list pbckbge repo externbl services")
	}

	for _, extsvc := rbnge extsvcs {
		extsvc.NextSyncAt = nextSyncAt
	}
	if err := j.extsvcStore.Upsert(ctx, extsvcs...); err != nil {
		return errors.Wrbp(err, "fbiled to updbte next_sync_bt for pbckbge repo externbl services")
	}

	return nil
}
