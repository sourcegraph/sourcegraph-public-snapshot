package limiter

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

func NewPrefixRedisStore(prefix string, store RedisStore) RedisStore {
	return &prefixRedisStore{
		prefix: prefix,
		store:  store,
	}
}

// featurePrefix is the prefix used by redis store for the given
// feature because we need to rate limit by feature.
func featurePrefix(feature codygateway.Feature) string {
	return fmt.Sprintf("%s:", feature)
}

func NewFeatureUsageStore(store RedisStore, feature codygateway.Feature) RedisStore {
	return &prefixRedisStore{
		prefix: featurePrefix(feature),
		store:  store,
	}
}

type prefixRedisStore struct {
	prefix string
	store  RedisStore
}

func (s *prefixRedisStore) Incrby(key string, val int) (int, error) {
	return s.store.Incrby(s.prefix+key, val)
}

func (s *prefixRedisStore) GetInt(key string) (int, error) {
	return s.store.GetInt(s.prefix + key)
}

func (s *prefixRedisStore) TTL(key string) (int, error) {
	return s.store.TTL(s.prefix + key)
}

func (s *prefixRedisStore) Expire(key string, ttlSeconds int) error {
	return s.store.Expire(s.prefix+key, ttlSeconds)
}

func (s *prefixRedisStore) Del(key string) error {
	return s.store.Del(s.prefix + key)
}
