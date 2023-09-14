package cache

import (
	"context"
	"hash/fnv"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

// New returns a cache.Key with an LRU cache of `size` values, wrapping the passed key.
func New(k encryption.Key, size int) (*Key, error) {
	c, err := lru.NewWithEvict(size, func(key uint64, value encryption.Secret) { evictTotal.WithLabelValues().Inc() })
	if err != nil {
		return nil, err
	}
	return &Key{
		Key:   k,
		cache: c,
	}, nil
}

// Key provides an LRU cache wrapper for any encryption.Key implementation, caching the decrypted
// value based on the ciphertext passed.
type Key struct {
	encryption.Key

	cache *lru.Cache[uint64, encryption.Secret]
}

// Decrypt attempts to find the decrypted ciphertext in the cache, if it is not found, the
// underlying key implementation is used, and the result is added to the cache.
func (k *Key) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	key := hash(ciphertext)
	s, found := k.cache.Get(key)
	if !found {
		missTotal.WithLabelValues().Inc()
		s, err := k.Key.Decrypt(ctx, ciphertext)
		if err != nil {
			loadErrorTotal.WithLabelValues().Inc()
			return nil, err
		}
		loadSuccessTotal.WithLabelValues().Inc()
		k.cache.Add(key, *s)
		return s, err
	} else {
		hitTotal.WithLabelValues().Inc()
	}
	return &s, nil
}

func hash(v []byte) uint64 {
	h := fnv.New64()
	h.Write(v)
	return h.Sum64()
}
