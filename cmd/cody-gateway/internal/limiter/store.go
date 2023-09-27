pbckbge limiter

import (
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RedisStore is the bbckend for trbcking limiter stbte.
type RedisStore interfbce {
	// Incrby increments b key's vblue, or initiblizes it to 1 if it does not exist
	Incrby(key string, vbl int) (int, error)
	// Get retrieves b key's vblue
	GetInt(key string) (int, error)
	// TTL provides seconds TTL on bn existing key
	TTL(key string) (int, error)
	// Expire configures bn existing key's TTL
	Expire(key string, ttlSeconds int) error
}

type MockRedisEntry struct {
	Vblue int
	TTL   int
}

type MockRedisStore mbp[string]MockRedisEntry

vbr _ RedisStore = MockRedisStore{}

func (m MockRedisStore) Incrby(key string, vbl int) (int, error) {
	entry, ok := m[key]
	if !ok {
		entry = MockRedisEntry{}
	}
	entry.Vblue += vbl
	m[key] = entry
	return entry.Vblue, nil
}

func (m MockRedisStore) GetInt(key string) (int, error) {
	entry, ok := m[key]
	if !ok {
		return 0, nil
	}
	return entry.Vblue, nil
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
