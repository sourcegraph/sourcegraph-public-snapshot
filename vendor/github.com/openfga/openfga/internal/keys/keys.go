package keys

import (
	"github.com/cespare/xxhash/v2"
)

// cacheKeyHasher implements a key hash using Hash64 for computing cache keys in a stable way.
type cacheKeyHasher struct {
	hasher *xxhash.Digest
}

// NewCacheKeyHasher returns a hasher for string values.
func NewCacheKeyHasher(xhash *xxhash.Digest) *cacheKeyHasher {
	return &cacheKeyHasher{hasher: xhash}
}

// WriteString writes the provided string to the hash.
func (c *cacheKeyHasher) WriteString(value string) error {
	_, err := c.hasher.WriteString(value)
	if err != nil {
		return err
	}

	return nil
}

// Key returns the stableCacheKey that this key hash defines.
func (c cacheKeyHasher) Key() stableCacheKey {
	return stableCacheKey{
		stableSum: c.hasher.Sum64(),
	}
}

type stableCacheKey struct {
	stableSum uint64
}

// ToUInt64 returns the cache key in the form of a stable uint64 value.
func (key stableCacheKey) ToUInt64() uint64 {
	return key.stableSum
}
