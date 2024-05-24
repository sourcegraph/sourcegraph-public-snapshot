package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/packagefilters"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type packagesFilterApplicatorJob struct {
	store       store.Store
	extsvcStore ExternalServiceStore
	operations  *operations
}

func NewPackagesFilterApplicator(
	obsctx *observation.Context,
	db database.DB,
) goroutine.BackgroundRoutine {
	job := packagesFilterApplicatorJob{
		store:       store.New(obsctx, db),
		extsvcStore: db.ExternalServices(),
		operations:  newOperations(obsctx),
	}

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(job.handle),
		goroutine.WithName("codeintel.package-filter-applicator"),
		goroutine.WithDescription("applies package repo filters to all package repo references to precompute their blocked status"),
		goroutine.WithInterval(time.Second*30),
	)
}

func (j *packagesFilterApplicatorJob) handle(ctx context.Context) (err error) {
	ctx, _, endObservation := j.operations.packagesFilterApplicator.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if needsFiltering, err := j.store.ShouldRefilterPackageRepoRefs(ctx); !needsFiltering || err != nil {
		// returns nil if err is nil, so its fine
		return errors.Wrap(err, "failed to check whether package repo filters need applying to anything")
	}

	filters, _, err := j.store.ListPackageRepoRefFilters(ctx, store.ListPackageRepoRefFiltersOpts{})
	if err != nil {
		return errors.Wrap(err, "failed to list package repo filters")
	}

	packageFilters, err := packagefilters.NewFilterLists(filters)
	if err != nil {
		return err
	}

	var (
		totalPkgsUpdated     int
		totalVersionsUpdated int
		startTime            = time.Now()
	)

	for lastID := 0; ; {
		pkgs, _, _, err := j.store.ListPackageRepoRefs(ctx, store.ListDependencyReposOpts{
			After:          lastID,
			Limit:          1000,
			IncludeBlocked: true,
		})
		if err != nil {
			return errors.Wrap(err, "failed to list package repos")
		}

		if len(pkgs) == 0 {
			break
		}

		lastID = pkgs[len(pkgs)-1].ID

		for i, pkg := range pkgs {
			pkg.Blocked = !packagefilters.IsPackageAllowed(pkg.Scheme, pkg.Name, packageFilters)
			for j, version := range pkg.Versions {
				pkg.Versions[j].Blocked = !packagefilters.IsVersionedPackageAllowed(pkg.Scheme, pkg.Name, version.Version, packageFilters)
			}
			pkgs[i] = pkg
		}

		pkgsUpdated, versionsUpdated, err := j.store.UpdateAllBlockedStatuses(ctx, pkgs, startTime)
		if err != nil {
			return errors.Wrap(err, "failed to update blocked statuses")
		}
		totalPkgsUpdated += pkgsUpdated
		totalVersionsUpdated += versionsUpdated
	}

	j.operations.versionsUpdated.Add(float64(totalVersionsUpdated))
	j.operations.packagesUpdated.Add(float64(totalPkgsUpdated))

	// now we want to mark all package repo extsvcs to sync so any (un)blocked pacakge repo references will be picked up

	nextSyncAt := time.Now()

	extsvcs, err := j.extsvcStore.List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{extsvc.KindJVMPackages, extsvc.KindNpmPackages, extsvc.KindGoPackages, extsvc.KindRustPackages, extsvc.KindRubyPackages, extsvc.KindPythonPackages},
	})
	if err != nil {
		return errors.Wrap(err, "failed to list package repo external services")
	}

	for _, extsvc := range extsvcs {
		extsvc.NextSyncAt = nextSyncAt
	}
	if err := j.extsvcStore.Upsert(ctx, extsvcs...); err != nil {
		return errors.Wrap(err, "failed to update next_sync_at for package repo external services")
	}

	return nil
}
