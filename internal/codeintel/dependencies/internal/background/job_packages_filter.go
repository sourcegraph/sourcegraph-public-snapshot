package background

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type packagesFilterApplicatorJob struct {
	store store.Store
}

func NewPackagesFilterApplicator(
	obsctx *observation.Context,
	db database.DB,
) goroutine.BackgroundRoutine {
	job := packagesFilterApplicatorJob{
		store: store.New(obsctx, db),
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.package-filter-applicator", "applies package repo filters to all package repo references to precompute their blocked status",
		time.Second*5,
		goroutine.HandlerFunc(job.handle))
}

func (j *packagesFilterApplicatorJob) handle(ctx context.Context) (err error) {
	var pkgsAffected, versionsAffected int

	filters, err := j.store.ListPackageRepoRefFilters(ctx, store.ListPackageRepoRefFiltersOpts{
		IncludeDeleted: true,
	})
	if err != nil {
		return errors.Wrap(err, "failed to list package repo filters")
	}

	if len(filters) == 0 {
		return nil
	}

	slices.SortFunc(filters, func(a, b shared.PackageFilter) bool {
		if a.DeletedAt.Valid {
			return a.DeletedAt.Time.After(b.DeletedAt.Time)
		}

		return a.UpdatedAt.After(b.UpdatedAt)
	})

	latestUpdated := filters[0].UpdatedAt
	if filters[0].DeletedAt.Valid {
		latestUpdated = filters[0].DeletedAt.Time
	}

	if needsFiltering, err := j.store.ExistsPackageRepoRefLastCheckedBefore(ctx, latestUpdated); !needsFiltering || err != nil {
		// returns nil if err is nil, so its fine
		return errors.Wrap(err, "failed to check whether package repo filters need applying to anything")
	}

	pkgsAffected, versionsAffected, err = j.store.ApplyPackageFilters(ctx)
	fmt.Println("DOING THE THING", pkgsAffected, versionsAffected, err)
	return err
}
