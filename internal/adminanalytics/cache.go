package adminanalytics

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	store               = redispool.Store
	scopeKey            = "adminanalytics:"
	cacheDisabledInTest = false
)

func getArrayFromCache[K interface{}](cacheKey string) ([]*K, error) {
	data, err := store.Get(scopeKey + cacheKey).String()
	if err != nil {
		return nil, err
	}

	nodes := make([]*K, 0)

	if err = json.Unmarshal([]byte(data), &nodes); err != nil {
		return nodes, err
	}

	return nodes, nil
}

func getItemFromCache[T interface{}](cacheKey string) (*T, error) {
	data, err := store.Get(scopeKey + cacheKey).String()
	if err != nil {
		return nil, err
	}

	var summary T

	if err = json.Unmarshal([]byte(data), &summary); err != nil {
		return &summary, err
	}

	return &summary, nil
}

func setDataToCache(key string, data string, expireSeconds int) error {
	if cacheDisabledInTest {
		return nil
	}

	if expireSeconds == 0 {
		expireSeconds = 24 * 60 * 60 // 1 day
	}

	return store.SetEx(scopeKey+key, expireSeconds, data)
}

func setArrayToCache[T interface{}](cacheKey string, nodes []*T) error {
	data, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	return setDataToCache(cacheKey, string(data), 0)
}

func setItemToCache[T interface{}](cacheKey string, summary *T) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}

	return setDataToCache(cacheKey, string(data), 0)
}

var dateRanges = []string{LastThreeMonths, LastMonth, LastWeek}
var groupBys = []string{Weekly, Daily}

type CacheAll interface {
	CacheAll(ctx context.Context) error
}

func refreshAnalyticsCache(ctx context.Context, db database.DB) error {
	for _, dateRange := range dateRanges {
		for _, groupBy := range groupBys {
			stores := []CacheAll{
				&Search{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&Users{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&Notebooks{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&CodeIntel{Ctx: ctx, DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&Repos{DB: db, Cache: true},
				&BatchChanges{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
				&Extensions{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
				&CodeInsights{Ctx: ctx, Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
			}
			for _, store := range stores {
				if err := store.CacheAll(ctx); err != nil {
					return err
				}
			}
		}

		_, err := GetCodeIntelByLanguage(ctx, db, true, dateRange)
		if err != nil {
			return err
		}

		_, err = GetCodeIntelTopRepositories(ctx, db, true, dateRange)
		if err != nil {
			return err
		}
	}

	return nil
}

var started bool

func StartAnalyticsCacheRefresh(ctx context.Context, db database.DB) {
	logger := log.Scoped("adminanalytics:cache-refresh")

	if started {
		panic("already started")
	}

	started = true
	ctx = featureflag.WithFlags(ctx, db.FeatureFlags())

	const delay = 24 * time.Hour
	for {
		if err := refreshAnalyticsCache(ctx, db); err != nil {
			logger.Error("Error refreshing admin analytics cache", log.Error(err))
		}

		// Randomize sleep to prevent thundering herds.
		randomDelay := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(delay + randomDelay)
	}
}
