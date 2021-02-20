package background

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func newSpecExpireWorker(ctx context.Context, campaignsStore *store.Store) goroutine.BackgroundRoutine {
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
