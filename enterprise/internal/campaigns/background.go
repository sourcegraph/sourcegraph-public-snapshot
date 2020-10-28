package campaigns

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func StartBackgroundJobs(ctx context.Context, db *sql.DB, campaignsStore *Store, repoStore repos.Store, cf *httpcli.Factory) {
	sourcer := repos.NewSourcer(cf)

	reconcilerWorker := NewWorker(ctx, campaignsStore, gitserver.DefaultClient, sourcer)

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

	// Set up expired spec deletion.
	expireWorker := goroutine.NewPeriodicGoroutine(ctx, 2*time.Minute, expireSpecs)

	goroutine.MonitorBackgroundRoutines(expireWorker, reconcilerWorker)
}
