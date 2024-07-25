package adminanalytics

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	MockStore           redispool.KeyValue
	scopeKey            = "adminanalytics:"
	cacheDisabledInTest = false
)

func store() redispool.KeyValue {
	if MockStore != nil {
		return MockStore
	}
	return redispool.Store
}

func getArrayFromCache[K interface{}](cacheKey string) ([]*K, error) {
	data, err := store().Get(scopeKey + cacheKey).String()
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
	data, err := store().Get(scopeKey + cacheKey).String()
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

	return store().SetEx(scopeKey+key, expireSeconds, data)
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
