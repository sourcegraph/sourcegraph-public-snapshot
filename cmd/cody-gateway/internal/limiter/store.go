package limiter

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RedisStore is the backend for tracking limiter state.
type RedisStore interface {
	// Incrby increments a key's value, or initializes it to 1 if it does not exist
	Incrby(key string, val int) (int, error)
	// Get retrieves a key's value
	GetInt(key string) (int, error)
	// TTL provides seconds TTL on an existing key
	TTL(key string) (int, error)
	// Expire configures an existing key's TTL
	Expire(key string, ttlSeconds int) error
}

type MockRedisEntry struct {
	Value int
	TTL   int
}

type MockRedisStore map[string]MockRedisEntry

var _ RedisStore = MockRedisStore{}

func (m MockRedisStore) Incrby(key string, val int) (int, error) {
	entry, ok := m[key]
	if !ok {
		entry = MockRedisEntry{}
	}
	entry.Value += val
	m[key] = entry
	return entry.Value, nil
}

func (m MockRedisStore) GetInt(key string) (int, error) {
	entry, ok := m[key]
	if !ok {
		return 0, nil
	}
	return entry.Value, nil
}

func (m MockRedisStore) TTL(key string) (int, error) {
	entry, ok := m[key]
	if !ok {
		return -1, errors.New("unknown key")
	}
	return entry.TTL, nil
}

func (m MockRedisStore) Expire(key string, ttlSeconds int) error {
	entry, ok := m[key]
	if !ok {
		return errors.New("unknown key")
	}
	entry.TTL = ttlSeconds
	m[key] = entry
	return nil
}
