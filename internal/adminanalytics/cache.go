package adminanalytics

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	pool     = redispool.Store
	scopeKey = "adminanalytics:"
)

func getArrayFromCache[K interface{}](cacheKey string) ([]*K, error) {
	rdb := pool.Get()
	defer rdb.Close()

	data, err := redis.String(rdb.Do("GET", scopeKey+cacheKey))
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
	rdb := pool.Get()
	defer rdb.Close()

	data, err := redis.String(rdb.Do("GET", scopeKey+cacheKey))
	if err != nil {
		return nil, err
	}

	var summary T

	if err = json.Unmarshal([]byte(data), &summary); err != nil {
		return &summary, err
	}

	return &summary, nil
}

func setDataToCache(key string, data string) (bool, error) {
	rdb := pool.Get()
	defer rdb.Close()

	if _, err := rdb.Do("SET", key, data); err != nil {
		return false, err
	}

	if _, err := rdb.Do("EXPIRE", key, int64(24*time.Hour/time.Second)); err != nil {
		return false, err
	}

	return true, nil
}

func setArrayToCache[T interface{}](cacheKey string, nodes []*T) (bool, error) {
	data, err := json.Marshal(nodes)
	if err != nil {
		return false, err
	}

	return setDataToCache(scopeKey+cacheKey, string(data))
}

func setItemToCache[T interface{}](cacheKey string, summary *T) (bool, error) {
	data, err := json.Marshal(summary)
	if err != nil {
		return false, err
	}

	return setDataToCache(scopeKey+cacheKey, string(data))
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
				&Search{DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&Users{DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&Notebooks{DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&CodeIntel{DateRange: dateRange, Grouping: groupBy, DB: db, Cache: true},
				&Repos{DB: db, Cache: true},
				&BatchChanges{Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
				&Extensions{Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
				&CodeInsights{Grouping: groupBy, DateRange: dateRange, DB: db, Cache: true},
			}
			for _, store := range stores {
				if err := store.CacheAll(ctx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

var started bool

func StartAnalyticsCacheRefresh(ctx context.Context, db database.DB) {
	logger := log.Scoped("adminanalytics:cache-refresh", "admin analytics cache refresh")

	if started {
		panic("already started")
	}

	started = true
	ctx = featureflag.WithFlags(ctx, db.FeatureFlags())

	const delay = 24 * time.Hour
	for {
		if !featureflag.FromContext(ctx).GetBoolOr("admin-analytics-disabled", false) {
			if err := refreshAnalyticsCache(ctx, db); err != nil {
				logger.Error("Error refreshing admin analytics cache", log.Error(err))
			}
		}

		// Randomize sleep to prevent thundering herds.
		randomDelay := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(delay + randomDelay)
	}
}
