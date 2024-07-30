package adminanalytics

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const scopeKey = "adminanalytics:"

// KeyValue is the subset of redispool.KeyValue that we use in adminanalytics.
type KeyValue interface {
	Get(key string) redispool.Value
	SetEx(key string, ttlSeconds int, val any) error
}

type NoopCache struct{}

var err = errors.New("NoopCache cache miss")

func (n NoopCache) Get(_ string) redispool.Value {
	// Return an error to simulate a cache miss.
	return redispool.NewValue(nil, err)
}

func (n NoopCache) SetEx(_ string, _ int, _ any) error {
	return nil
}

func getArrayFromCache[K interface{}](cache KeyValue, cacheKey string) ([]*K, error) {
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

func getItemFromCache[T interface{}](cache KeyValue, cacheKey string) (*T, error) {
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

func setDataToCache(cache KeyValue, key string, data string, expireSeconds int) error {
	if expireSeconds == 0 {
		expireSeconds = 24 * 60 * 60 // 1 day
	}

	return cache.SetEx(scopeKey+key, expireSeconds, data)
}

func setArrayToCache[T interface{}](cache KeyValue, cacheKey string, nodes []*T) error {
	data, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	return setDataToCache(cache, cacheKey, string(data), 0)
}

func setItemToCache[T interface{}](cache KeyValue, cacheKey string, summary *T) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}

	return setDataToCache(cache, cacheKey, string(data), 0)
}
