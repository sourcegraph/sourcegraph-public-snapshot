package wrapper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromCiphertext(t *testing.T) {
	sek := &StorableEncryptedKey{
		Mechanism:  "test",
		KeyName:    "testkey",
		WrappedKey: []byte("wrappedkey"),
		Ciphertext: []byte("ciphertext"),
		Nonce:      []byte("nonce"),
	}

	serialized, err := sek.Serialize()
	require.NoError(t, err)

	recoveredSek, err := FromCiphertext(serialized)
	require.NoError(t, err)

	require.Equal(t, sek, recoveredSek)
}
