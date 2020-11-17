package background

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func newSpecExpireWorker(ctx context.Context, campaignsStore *campaigns.Store) goroutine.BackgroundRoutine {
	expireSpecs := goroutine.NewHandlerWithErrorMessage("expire campaigns specs", func(ctx context.Context) error {
		// We first need to delete expired ChangesetSpecs...
		if err := campaignsStore.DeleteExpiredChangesetSpecs(ctx); err != nil {
			return errors.Wrap(err, "DeleteExpiredChangesetSpecs")
		}
		// ... and then the CampaignSpecs, due to the campaign_spec_id
		// foreign key on changeset_specs.
		if err := campaignsStore.DeleteExpiredCampaignSpecs(ctx); err != nil {
			return errors.Wrap(err, "DeleteExpiredCampaignSpecs")
		}
		return nil
	})
	return goroutine.NewPeriodicGoroutine(ctx, 2*time.Minute, expireSpecs)
}

func newHistoryMigrator(ctx context.Context, campaignsStore *campaigns.Store) goroutine.BackgroundRoutine {
	migrateHistory := goroutine.NewHandlerWithErrorMessage("migrate changeset histories", func(ctx context.Context) error {
		println("Migrating changeset histories")
		for {
			cs, next, err := campaignsStore.ListChangesets(ctx, campaigns.ListChangesetsOpts{WithoutHistory: true, LimitOpts: campaigns.LimitOpts{Limit: 50}})
			if err != nil {
				return err
			}
			for _, c := range cs {
				fmt.Printf("Migrating changeset %d\n", c.ID)
				events, _, err := campaignsStore.ListChangesetEvents(ctx, campaigns.ListChangesetEventsOpts{ChangesetIDs: []int64{c.ID}})
				if err != nil {
					return err
				}
				campaigns.SetDerivedState(ctx, c, events)
				if err := campaignsStore.UpdateChangeset(ctx, c); err != nil {
					return err
				}
				fmt.Printf("Migrated changeset %d\n", c.ID)
			}
			if next == 0 {
				time.Sleep(1 * time.Minute)
			}
		}
	})
	return goroutine.NewPeriodicGoroutine(ctx, 2*time.Minute, migrateHistory)
}
