package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/own/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func computeRecentViewSignals(ctx context.Context, db database.DB, path string, repoID api.RepoID) ([]reasonAndReference, error) {
	enabled, err := db.OwnSignalConfigurations().IsEnabled(ctx, types.SignalRecentViews)
	if err != nil {
		return nil, errors.Wrap(err, "IsEnabled")
	}
	if !enabled {
		return nil, nil
	}

	summaries, err := db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{Path: path, RepoID: repoID})
	if err != nil {
		return nil, errors.Wrap(err, "list recent view signals")
	}

	var rrs []reasonAndReference
	for _, s := range summaries {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{recentViewsCount: s.ViewsCount},
			reference: own.Reference{
				UserID: s.UserID,
			},
		})
	}
	return rrs, nil
}

type recentViewOwnershipSignal struct {
	total int32
}

func (v *recentViewOwnershipSignal) Title() (string, error) {
	return "recent view", nil
}

func (v *recentViewOwnershipSignal) Description() (string, error) {
	return "Associated because they have viewed this file in the last 90 days.", nil
}
