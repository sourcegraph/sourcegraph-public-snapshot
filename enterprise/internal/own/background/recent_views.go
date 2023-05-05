package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	viewBlobEventName      = "ViewBlob"
	processedEventsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "own_recent_views_events_processed_total",
	})
	indexInterval     = time.Minute * 5
	mockIndexInterval time.Duration
)

func NewOwnRecentViewsIndexer(db database.DB, observationCtx *observation.Context) goroutine.BackgroundRoutine {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"own_background_recent_views_indexer",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)
	operation := observationCtx.Operation(observation.Op{
		Name:              "own.background.index.recent-views",
		MetricLabelValues: []string{"recent-views"},
		Metrics:           redMetrics,
	})
	handler := newRecentViewsIndexer(db, operation.Logger)
	interval := indexInterval
	if mockIndexInterval != 0 {
		interval = mockIndexInterval
	}
	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), "own.recent-views", "", interval, handler, operation)
}

type recentViewsIndexer struct {
	db     database.DB
	logger log.Logger
}

func newRecentViewsIndexer(db database.DB, logger log.Logger) *recentViewsIndexer {
	return &recentViewsIndexer{db: db, logger: logger}
}

func (r *recentViewsIndexer) Handle(ctx context.Context) error {
	logJobDisabled := func() {
		r.logger.Info("skipping own background job, job disabled", log.String("job-name", "recent-views"))
	}
	flag, err := r.db.FeatureFlags().GetFeatureFlag(ctx, "own-background-index-repo-recent-views")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logJobDisabled()
			return nil
		} else {
			return errors.Wrap(err, "database.FeatureFlagsWith")
		}
	}
	res, ok := flag.EvaluateGlobal()
	if !ok || !res {
		logJobDisabled()
		return nil
	}

	// The job is enabled, here we go. First we need to get the ID of last processed event.
	bookmark, err := r.db.EventLogsScrapeState().GetBookmark(ctx, int(RecentViews))
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
	return r.db.EventLogsScrapeState().UpdateBookmark(ctx, newBookmark, int(RecentViews))
}
