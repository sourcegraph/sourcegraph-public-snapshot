package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// NewInsightsDataPrunerJob will periodically delete recorded data series that have been marked `deleted`.
func NewInsightsDataPrunerJob(ctx context.Context, base dbutil.DB, insights dbutil.DB) goroutine.BackgroundRoutine {
	interval := time.Minute * 60

	return goroutine.NewPeriodicGoroutine(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("insights_data_prune", func(ctx context.Context) error {
			// select the series that need to be deleted

			// delete them!
			return nil
		}))
}

const selectDeletedSeriesSql = `
select series_id from insight_series i
left join insight_view_series ivs ON i.id = ivs.insight_series_id
where i.deleted_at is not null
  and i.deleted_at < %s
and ivs.insight_series_id is null;
`
