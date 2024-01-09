package limiter

import (
	"fmt"
	"strings"

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
	// Delete removes a key
	Del(key string) error
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

func (m MockRedisStore) Del(key string) error {
	delete(m, key)

	return nil
}

// RecordingRedisStoreFake is a fake for the RedisStore interface, but it
// also records every operation. This allows tests to confirm the exact
// operations that were performed on the store when processing a request.
type RecordingRedisStoreFake struct {
	Data    map[string]MockRedisEntry
	History []string
}

var _ RedisStore = &RecordingRedisStoreFake{}

func NewRecordingRedisStoreFake() *RecordingRedisStoreFake {
	return &RecordingRedisStoreFake{
		Data:    make(map[string]MockRedisEntry),
		History: make([]string, 0, 10),
	}
}

func (rrs *RecordingRedisStoreFake) record(op string, args ...interface{}) {
	argStrings := make([]string, len(args))
	for idx, arg := range args {
		argStrings[idx] = fmt.Sprintf("%v", arg)
	}
	line := fmt.Sprintf("%s(%s)", op, strings.Join(argStrings, ","))
	rrs.History = append(rrs.History, line)
}

func (rrs *RecordingRedisStoreFake) Incrby(key string, val int) (int, error) {
	rrs.record("Incrby", key, val)
	entry, ok := rrs.Data[key]
	if !ok {
		entry = MockRedisEntry{}
	}
	entry.Value += val
	rrs.Data[key] = entry
	return entry.Value, nil
}

func (rrs *RecordingRedisStoreFake) GetInt(key string) (int, error) {
	rrs.record("GetInt", key)
	entry, ok := rrs.Data[key]
	if !ok {
		return 0, nil
	}
	return entry.Value, nil
}

func (rrs *RecordingRedisStoreFake) TTL(key string) (int, error) {
	rrs.record("TTL", key)
	entry, ok := rrs.Data[key]
	if !ok {
		return -1, errors.New("unknown key")
	}
	return entry.TTL, nil
}

func (rrs *RecordingRedisStoreFake) Expire(key string, ttlSeconds int) error {
	rrs.record("Expire", key, ttlSeconds)
	entry, ok := rrs.Data[key]
	if !ok {
		return errors.New("unknown key")
	}
	entry.TTL = ttlSeconds
	rrs.Data[key] = entry
	return nil
}

func (rrs *RecordingRedisStoreFake) Del(key string) error {
	rrs.record("Del", key)
	delete(rrs.Data, key)
	return nil
}
