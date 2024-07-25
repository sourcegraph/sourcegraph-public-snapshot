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

func refreshAnalyticsCache(ctx context.Context, db database.DB) error {
	for _, dateRange := range dateRanges {
		for _, groupBy := range groupBys {
			stores := []cacheAll{
				&adminanalytics.Search{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&adminanalytics.Users{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&adminanalytics.Notebooks{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&adminanalytics.CodeIntel{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&adminanalytics.Repos{DB: db, Cache: true},
				&adminanalytics.BatchChanges{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
				&adminanalytics.Extensions{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
				&adminanalytics.CodeInsights{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
			}
			for _, store := range stores {
				if err := store.CacheAll(ctx); err != nil {
					return err
				}
			}
		}

		_, err := adminanalytics.GetCodeIntelByLanguage(ctx, db, true, dateRange)
		if err != nil {
			return err
		}

		_, err = adminanalytics.GetCodeIntelTopRepositories(ctx, db, true, dateRange)
		if err != nil {
			return err
		}
	}

	return nil
}
