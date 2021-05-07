package cache

import (
	"context"

	lru "github.com/hashicorp/golang-lru"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

func New(k encryption.Key, size int) (*Key, error) {
	c, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &Key{
		Key:   k,
		cache: c,
	}, nil
}

type Key struct {
	encryption.Key

	cache *lru.Cache
}

func (k *Key) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	v, found := k.cache.Get(string(ciphertext))
	s, ok := v.(encryption.Secret)
	if !ok || !found {
		s, err := k.Key.Decrypt(ctx, ciphertext)
		if err != nil {
			return nil, err
		}
		k.cache.Add(string(ciphertext), *s)
		return s, err
	}
	return &s, nil
}
