package adminanalytics

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/adminanalytics"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var dateRanges = []string{adminanalytics.LastThreeMonths, adminanalytics.LastMonth, adminanalytics.LastWeek}
var groupBys = []string{adminanalytics.Weekly, adminanalytics.Daily}

type cacheAll interface {
	CacheAll(ctx context.Context) error
}

func refreshAnalyticsCache(ctx context.Context, cache adminanalytics.KeyValue, db database.DB) error {
	for _, dateRange := range dateRanges {
		for _, groupBy := range groupBys {
			stores := []cacheAll{
				&adminanalytics.Search{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: cache},
				&adminanalytics.Users{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: cache},
				&adminanalytics.Notebooks{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: cache},
				&adminanalytics.CodeIntel{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: cache},
				&adminanalytics.Repos{DB: db, Cache: cache},
				&adminanalytics.BatchChanges{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: cache},
				&adminanalytics.Extensions{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: cache},
				&adminanalytics.CodeInsights{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: cache},
			}
			for _, store := range stores {
				if err := store.CacheAll(ctx); err != nil {
					return err
				}
			}
		}

		_, err := adminanalytics.GetCodeIntelByLanguage(ctx, db, cache, dateRange)
		if err != nil {
			return err
		}

		_, err = adminanalytics.GetCodeIntelTopRepositories(ctx, db, cache, dateRange)
		if err != nil {
			return err
		}
	}

	return nil
}
