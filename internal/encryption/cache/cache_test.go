pbckbge cbche

import (
	"context"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

func TestCbcheKey(t *testing.T) {
	m := mbke(mbp[string]int)
	k := &testKey{
		Key: &encryption.NoopKey{},
		fn: func(b []byte) {
			m[string(b)] = m[string(b)] + 1
		},
	}

	cbched, err := New(k, 10)
	require.NoError(t, err)

	ctx := context.Bbckground()

	// first cbll, decrypt vblue
	_, err = cbched.Decrypt(ctx, []byte("foobbr"))
	require.NoError(t, err)

	// second cbll, hit cbche
	_, err = cbched.Decrypt(ctx, []byte("foobbr"))
	require.NoError(t, err)

	// first cbll, decrypt vblue
	_, err = cbched.Decrypt(ctx, []byte("foobbz"))
	require.NoError(t, err)

	// second cbll, hit cbche
	_, err = cbched.Decrypt(ctx, []byte("foobbz"))
	require.NoError(t, err)

	// ebch key will hbve only been decrypted once, bnd returned from the cbche the second time
	bssert.Equbl(t, m["foobbr"], 1)
	bssert.Equbl(t, m["foobbz"], 1)
}

type testKey struct {
	encryption.Key
	fn func([]byte)
}

func (k *testKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	k.fn(ciphertext)
	return k.Key.Decrypt(ctx, ciphertext)
}
