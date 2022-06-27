package adminanalytics

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	pool = redispool.Store
)

func getCacheKey(f *AnalyticsFetcher, data string) string {
	return fmt.Sprintf("adminanalytics:%s:%s:%s", f.group, f.dateRange, data)
}

func getNodesFromCache(f *AnalyticsFetcher) ([]*AnalyticsNode, error) {
	rdb := pool.Get()
	defer rdb.Close()

	data, err := redis.String(rdb.Do("GET", getCacheKey(f, "nodes")))
	if err != nil {
		return nil, err
	}

	nodes := make([]*AnalyticsNode, 0)

	if err = json.Unmarshal([]byte(data), &nodes); err != nil {
		return nodes, err
	}

	return nodes, nil
}

func getSummaryFromCache(f *AnalyticsFetcher) (*AnalyticsSummary, error) {
	rdb := pool.Get()
	defer rdb.Close()

	data, err := redis.String(rdb.Do("GET", getCacheKey(f, "summary")))
	if err != nil {
		return nil, err
	}

	var summary AnalyticsSummary

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

func setNodesToCache(f *AnalyticsFetcher, nodes []*AnalyticsNode) (bool, error) {
	data, err := json.Marshal(nodes)
	if err != nil {
		return false, err
	}

	return setDataToCache(getCacheKey(f, "nodes"), string(data))
}

func setSummaryToCache(f *AnalyticsFetcher, summary *AnalyticsSummary) (bool, error) {
	data, err := json.Marshal(summary)
	if err != nil {
		return false, err
	}

	return setDataToCache(getCacheKey(f, "summary"), string(data))
}
