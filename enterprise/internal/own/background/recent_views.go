package background

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	viewBlobEventName      = "ViewBlob"
	processedEventsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "own_recent_views_events_processed_total",
	})
)

type recentViewsIndexer struct {
	db     database.DB
	logger log.Logger
}

func newRecentViewsIndexer(db database.DB, logger log.Logger) *recentViewsIndexer {
	return &recentViewsIndexer{db: db, logger: logger}
}

func (r *recentViewsIndexer) Handle(ctx context.Context) error {
	// The job is enabled, here we go. First we need to get the ID of last processed event.
	bookmark, err := r.db.EventLogsScrapeState().GetBookmark(ctx, types.SignalRecentViews)
	if err != nil {
		return errors.Wrap(err, "getting latest processed event ID")
	}
	events, err := r.db.EventLogs().ListAll(ctx, database.EventLogsListOptions{LimitOffset: &database.LimitOffset{Limit: 5000}, EventName: &viewBlobEventName, AfterID: bookmark})
	if err != nil {
		return errors.Wrap(err, "getting event logs")
	}
	numberOfEvents := len(events)
	if numberOfEvents == 0 {
		return nil
	}
	err = r.db.RecentViewSignal().BuildAggregateFromEvents(ctx, events)
	if err != nil {
		return errors.Wrap(err, "building aggregates from events")
	}
	newBookmark := int(events[numberOfEvents-1].ID)
	r.logger.Info("events processed", log.Int("count", numberOfEvents))
	processedEventsCounter.Add(float64(numberOfEvents))
	return r.db.EventLogsScrapeState().UpdateBookmark(ctx, newBookmark, types.SignalRecentViews)
}
