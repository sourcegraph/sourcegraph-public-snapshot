package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/adminanalytics"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type siteAnalyticsResolver struct {
	db database.DB
}

var cache = true

/* Analytics root resolver */
func (r *siteResolver) Analytics(ctx context.Context) (*siteAnalyticsResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("admin-analytics-enabled", false) {
		return nil, errors.New("'admin-analytics-enabled' feature flag is not enabled")
	}
	return &siteAnalyticsResolver{r.db}, nil
}

/* Search */

func (r *siteAnalyticsResolver) Search(ctx context.Context, args *struct {
	DateRange *string
}) *adminanalytics.Search {
	return &adminanalytics.Search{DateRange: *args.DateRange, DB: r.db, Cache: cache}
}

/* Notebooks */

func (r *siteAnalyticsResolver) Notebooks(ctx context.Context, args *struct {
	DateRange *string
}) *adminanalytics.Notebooks {
	return &adminanalytics.Notebooks{DateRange: *args.DateRange, DB: r.db, Cache: cache}
}

/* Users */

func (r *siteAnalyticsResolver) Users(ctx context.Context, args *struct {
	DateRange *string
}) (*adminanalytics.Users, error) {
	return &adminanalytics.Users{DateRange: *args.DateRange, DB: r.db, Cache: cache}, nil
}

/* Code-intel */

func (r *siteAnalyticsResolver) CodeIntel(ctx context.Context, args *struct {
	DateRange *string
}) *adminanalytics.CodeIntel {
	return &adminanalytics.CodeIntel{DateRange: *args.DateRange, DB: r.db, Cache: cache}
}

/* Repos */

func (r *siteAnalyticsResolver) Repos(ctx context.Context) (*adminanalytics.ReposSummary, error) {
	repos := adminanalytics.Repos{DB: r.db, Cache: cache}

	return repos.Summary(ctx)
}
