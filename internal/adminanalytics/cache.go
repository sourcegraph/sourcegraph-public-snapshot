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
	pool                = redispool.Store
	scopeKey            = "adminanalytics:"
	cacheDisabledInTest = false
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

func setDataToCache(key string, data string, expire int64) (bool, error) {
	if cacheDisabledInTest {
		return true, nil
	}

	rdb := pool.Get()
	defer rdb.Close()

	if _, err := rdb.Do("SET", scopeKey+key, data); err != nil {
		return false, err
	}

	if expire == 0 {
		expire = int64((24 * time.Hour).Seconds())
	}

	if _, err := rdb.Do("EXPIRE", scopeKey+key, expire); err != nil {
		return false, err
	}

	return true, nil
}

func setArrayToCache[T interface{}](cacheKey string, nodes []*T) (bool, error) {
	data, err := json.Marshal(nodes)
	if err != nil {
		return false, err
	}

	return setDataToCache(cacheKey, string(data), 0)
}

func setItemToCache[T interface{}](cacheKey string, summary *T) (bool, error) {
	data, err := json.Marshal(summary)
	if err != nil {
		return false, err
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
