package adminanalytics

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	scopeKey = "adminanalytics:"
)

func getArrayFromCache[K interface{}](cache redispool.KeyValue, cacheKey string) ([]*K, error) {
	data, err := cache.Get(scopeKey + cacheKey).String()
	if err != nil {
		return nil, err
	}

	nodes := make([]*K, 0)

	if err = json.Unmarshal([]byte(data), &nodes); err != nil {
		return nodes, err
	}

	return nodes, nil
}

func getItemFromCache[T interface{}](cache redispool.KeyValue, cacheKey string) (*T, error) {
	data, err := cache.Get(scopeKey + cacheKey).String()
	if err != nil {
		return nil, err
	}

	var summary T

	if err = json.Unmarshal([]byte(data), &summary); err != nil {
		return &summary, err
	}

	return &summary, nil
}

func setDataToCache(cache redispool.KeyValue, key string, data string, expireSeconds int) error {
	if expireSeconds == 0 {
		expireSeconds = 24 * 60 * 60 // 1 day
	}

	return cache.SetEx(scopeKey+key, expireSeconds, data)
}

func setArrayToCache[T interface{}](cache redispool.KeyValue, cacheKey string, nodes []*T) error {
	data, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	return setDataToCache(cache, cacheKey, string(data), 0)
}

func setItemToCache[T interface{}](cache redispool.KeyValue, cacheKey string, summary *T) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}

	return setDataToCache(cache, cacheKey, string(data), 0)
}
