package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/adminanalytics"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type siteAnalyticsResolver struct {
	db database.DB
}

func (r *siteResolver) Analytics(ctx context.Context) (*siteAnalyticsResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	return &siteAnalyticsResolver{r.db}, nil
}

type siteAnalyticsStatItemResolver struct {
	fetcher *adminanalytics.AnalyticsFetcher
}

func (r *siteAnalyticsStatItemResolver) Nodes(ctx context.Context) ([]*adminanalytics.AnalyticsNode, error) {
	nodes, err := r.fetcher.GetNodes(ctx)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (r *siteAnalyticsStatItemResolver) Summary(ctx context.Context) (*adminanalytics.AnalyticsSummary, error) {
	summary, err := r.fetcher.GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *siteAnalyticsResolver) Search(ctx context.Context, args *struct {
	DateRange *string
}) (*siteAnalyticsSearchResolver, error) {
	return &siteAnalyticsSearchResolver{store: &adminanalytics.Search{DateRange: *args.DateRange, DB: r.db}}, nil
}

type siteAnalyticsSearchResolver struct {
	store *adminanalytics.Search
}

func (r *siteAnalyticsSearchResolver) Searches(ctx context.Context) (*siteAnalyticsStatItemResolver, error) {
	fetcher, err := r.store.Searches()

	if err != nil {
		return nil, err
	}

	return &siteAnalyticsStatItemResolver{fetcher}, nil
}

func (r *siteAnalyticsSearchResolver) FileViews(ctx context.Context) (*siteAnalyticsStatItemResolver, error) {
	fetcher, err := r.store.FileViews()

	if err != nil {
		return nil, err
	}

	return &siteAnalyticsStatItemResolver{fetcher}, nil
}

func (r *siteAnalyticsSearchResolver) FileOpens(ctx context.Context) (*siteAnalyticsStatItemResolver, error) {
	fetcher, err := r.store.FileOpens()

	if err != nil {
		return nil, err
	}

	return &siteAnalyticsStatItemResolver{fetcher}, nil
}
