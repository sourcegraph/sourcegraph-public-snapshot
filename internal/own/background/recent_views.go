package background

import (
	"context"
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/own/types"
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

type viewBlob struct {
	RepoName string `json:"repoName"`
	FilePath string `json:"filePath"`
}

func newRecentViewsIndexer(db database.DB, logger log.Logger) *recentViewsIndexer {
	return &recentViewsIndexer{db: db, logger: logger}
}

func (r *recentViewsIndexer) Handle(ctx context.Context) error {
	return r.handle(ctx, authz.DefaultSubRepoPermsChecker)
}

func (r *recentViewsIndexer) handle(ctx context.Context, checker authz.SubRepoPermissionChecker) error {
	// The job is enabled, here we go. First we need to get the ID of last processed event.
	bookmark, err := r.db.EventLogsScrapeState().GetBookmark(ctx, types.SignalRecentViews)
	if err != nil {
		return errors.Wrap(err, "getting latest processed event ID")
	}
	events, err := r.db.EventLogs().ListAll(ctx, database.EventLogsListOptions{LimitOffset: &database.LimitOffset{Limit: 5000}, EventName: &viewBlobEventName, AfterID: bookmark})
	if err != nil {
		return errors.Wrap(err, "getting event logs")
	}
	var filteredEvents []*database.Event
	subRepoPermsCache := map[string]bool{}
	for _, event := range events {
		var vb viewBlob
		err = json.Unmarshal(event.PublicArgument, &vb)
		if err != nil {
			r.logger.Debug("could not use view event for signal", log.Object("event",
				log.String("name", event.Name),
				log.String("url", event.URL)))
			continue
		}

		if isSubRepoPermsRepo, ok := subRepoPermsCache[vb.RepoName]; ok {
			if !isSubRepoPermsRepo {
				filteredEvents = append(filteredEvents, event)
			}
			continue
		}
		ok, err := authz.SubRepoEnabledForRepo(ctx, checker, api.RepoName(vb.RepoName))
		if err != nil {
			r.logger.Debug("encountered error checking subrepo permissions for repo", log.String("repo name", vb.RepoName), log.Error(err))
		} else if ok {
			subRepoPermsCache[vb.RepoName] = true
		} else {
			filteredEvents = append(filteredEvents, event)
			subRepoPermsCache[vb.RepoName] = false
		}
	}
	numberOfEvents := len(filteredEvents)

	if numberOfEvents == 0 {
		return nil
	}
	err = r.db.RecentViewSignal().BuildAggregateFromEvents(ctx, filteredEvents)
	if err != nil {
		return errors.Wrap(err, "building aggregates from events")
	}
	newBookmark := int(events[numberOfEvents-1].ID)
	r.logger.Info("events processed", log.Int("count", numberOfEvents))
	processedEventsCounter.Add(float64(numberOfEvents))
	return r.db.EventLogsScrapeState().UpdateBookmark(ctx, newBookmark, types.SignalRecentViews)
}
