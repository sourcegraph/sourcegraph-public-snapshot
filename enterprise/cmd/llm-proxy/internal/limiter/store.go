package limiter

import "github.com/sourcegraph/sourcegraph/lib/errors"

// RedisStore is the backend for tracking limiter state.
type RedisStore interface {
	// Incr increments a key's value, or initializes it to 1 if it does not exist
	Incr(key string) (int, error)
	// TTL provides seconds TTL on an existing key
	TTL(key string) (int, error)
	// Expire configures an existing key's TTL
	Expire(key string, ttlSeconds int) error
}

type mockRedisEntry struct {
	value int
	ttl   int
}

type mockStore map[string]mockRedisEntry

var _ RedisStore = mockStore{}

func (m mockStore) Incr(key string) (int, error) {
	entry, ok := m[key]
	if !ok {
		entry = mockRedisEntry{}
	}
	entry.value++
	m[key] = entry
	return entry.value, nil
}

func (m mockStore) TTL(key string) (int, error) {
	entry, ok := m[key]
	if !ok {
		return -1, errors.New("unknown key")
	}
	return entry.ttl, nil
}

func (m mockStore) Expire(key string, ttlSeconds int) error {
	entry, ok := m[key]
	if !ok {
		return errors.New("unknown key")
	}
	entry.ttl = ttlSeconds
	m[key] = entry
	return nil
}
