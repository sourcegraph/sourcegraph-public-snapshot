package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type RecentViewSignalStore interface {
	List(ctx context.Context, opts ListRecentViewSignalOpts) ([]RecentViewSummary, error)
	ProcessEventsBetweenDates(ctx context.Context, from, to *time.Time)
}

type ListRecentViewSignalOpts struct {
	ViewerID    int
	ViewerEmail string
	ViewerName  string
	RepoID      api.RepoID
	path        string
}

type RecentViewSummary struct {
	ViewerEmail string
	ViewerName  string
	ViewsCount  int
}

type recentViewSignalStore struct {
	*basestore.Store
}

// TODO update query with more opts
const listRecentViewSignalsFmtstr = `
	SELECT viewer_email, viewer_name, views_count
	FROM own_aggregate_recent_view
	WHERE viewer_email = %s
`

func (s *recentViewSignalStore) List(ctx context.Context, opts ListRecentViewSignalOpts) ([]RecentViewSummary, error) {
	q := sqlf.Sprintf(listRecentViewSignalsFmtstr, opts.ViewerEmail)

	viewsScanner := basestore.NewSliceScanner(func(scanner dbutil.Scanner) (RecentViewSummary, error) {
		var summary RecentViewSummary
		if err := scanner.Scan(&summary.ViewerEmail, &summary.ViewerName, &summary.ViewsCount); err != nil {
			return RecentViewSummary{}, err
		}
		return summary, nil
	})

	return viewsScanner(s.Query(ctx, q))
}

func (s *recentViewSignalStore) ProcessEventsBetweenDates(ctx context.Context, from, to *time.Time) {
	// TODO
	// 1) Query all "ViewBlob" events from `event_logs` between the dates
	// 2) Make similar thing to `recentContributionSignalStore.ensureRepoPaths`
	// 3) Insert all the values to `own_aggregate_recent_view`
}
