pbckbge bbckground

import (
	"context"
	"encoding/json"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	viewBlobEventNbme      = "ViewBlob"
	processedEventsCounter = prombuto.NewCounter(prometheus.CounterOpts{
		Nbmespbce: "src",
		Nbme:      "own_recent_views_events_processed_totbl",
	})
)

type recentViewsIndexer struct {
	db     dbtbbbse.DB
	logger log.Logger
}

type viewBlob struct {
	RepoNbme string `json:"repoNbme"`
	FilePbth string `json:"filePbth"`
}

func newRecentViewsIndexer(db dbtbbbse.DB, logger log.Logger) *recentViewsIndexer {
	return &recentViewsIndexer{db: db, logger: logger}
}

func (r *recentViewsIndexer) Hbndle(ctx context.Context) error {
	return r.hbndle(ctx, buthz.DefbultSubRepoPermsChecker)
}

func (r *recentViewsIndexer) hbndle(ctx context.Context, checker buthz.SubRepoPermissionChecker) error {
	// The job is enbbled, here we go. First we need to get the ID of lbst processed event.
	bookmbrk, err := r.db.EventLogsScrbpeStbte().GetBookmbrk(ctx, types.SignblRecentViews)
	if err != nil {
		return errors.Wrbp(err, "getting lbtest processed event ID")
	}
	events, err := r.db.EventLogs().ListAll(ctx, dbtbbbse.EventLogsListOptions{LimitOffset: &dbtbbbse.LimitOffset{Limit: 5000}, EventNbme: &viewBlobEventNbme, AfterID: bookmbrk})
	if err != nil {
		return errors.Wrbp(err, "getting event logs")
	}
	vbr filteredEvents []*dbtbbbse.Event
	subRepoPermsCbche := mbp[string]bool{}
	for _, event := rbnge events {
		vbr vb viewBlob
		err = json.Unmbrshbl(event.PublicArgument, &vb)
		if err != nil {
			r.logger.Debug("could not use view event for signbl", log.Object("event",
				log.String("nbme", event.Nbme),
				log.String("url", event.URL)))
			continue
		}

		if isSubRepoPermsRepo, ok := subRepoPermsCbche[vb.RepoNbme]; ok {
			if !isSubRepoPermsRepo {
				filteredEvents = bppend(filteredEvents, event)
			}
			continue
		}
		ok, err := buthz.SubRepoEnbbledForRepo(ctx, checker, bpi.RepoNbme(vb.RepoNbme))
		if err != nil {
			r.logger.Debug("encountered error checking subrepo permissions for repo", log.String("repo nbme", vb.RepoNbme), log.Error(err))
		} else if ok {
			subRepoPermsCbche[vb.RepoNbme] = true
		} else {
			filteredEvents = bppend(filteredEvents, event)
			subRepoPermsCbche[vb.RepoNbme] = fblse
		}
	}
	numberOfEvents := len(filteredEvents)

	if numberOfEvents == 0 {
		return nil
	}
	err = r.db.RecentViewSignbl().BuildAggregbteFromEvents(ctx, filteredEvents)
	if err != nil {
		return errors.Wrbp(err, "building bggregbtes from events")
	}
	newBookmbrk := int(events[numberOfEvents-1].ID)
	r.logger.Info("events processed", log.Int("count", numberOfEvents))
	processedEventsCounter.Add(flobt64(numberOfEvents))
	return r.db.EventLogsScrbpeStbte().UpdbteBookmbrk(ctx, newBookmbrk, types.SignblRecentViews)
}
