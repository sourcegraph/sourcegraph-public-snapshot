package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

func TestCacheKey(t *testing.T) {
	m := make(map[string]int)
	k := &testKey{
		Key: &encryption.NoopKey{},
		fn: func(b []byte) {
			m[string(b)] = m[string(b)] + 1
		},
	}

	cached, err := New(k, 10)
	require.NoError(t, err)

	ctx := context.Background()

	// first call, decrypt value
	_, err = cached.Decrypt(ctx, []byte("foobar"))
	require.NoError(t, err)

	// second call, hit cache
	_, err = cached.Decrypt(ctx, []byte("foobar"))
	require.NoError(t, err)

	// first call, decrypt value
	_, err = cached.Decrypt(ctx, []byte("foobaz"))
	require.NoError(t, err)

	// second call, hit cache
	_, err = cached.Decrypt(ctx, []byte("foobaz"))
	require.NoError(t, err)

	// each key will have only been decrypted once, and returned from the cache the second time
	assert.Equal(t, m["foobar"], 1)
	assert.Equal(t, m["foobaz"], 1)
}

type testKey struct {
	encryption.Key
	fn func([]byte)
}

func (k *testKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	k.fn(ciphertext)
	return k.Key.Decrypt(ctx, ciphertext)
}
