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
	sek.SetChecksum([]byte("plaintext"))

	serialized, err := sek.Serialize()
	require.NoError(t, err)

	recoveredSek, err := FromCiphertext(serialized)
	require.NoError(t, err)

	require.Equal(t, sek, recoveredSek)
}

func TestChecksum(t *testing.T) {
	sek := &StorableEncryptedKey{}
	plaintext := []byte("plaintext")
	sek.SetChecksum(plaintext)

	require.NoError(t, sek.VerifyChecksum(plaintext))

	require.Error(t, sek.VerifyChecksum([]byte("wrong plaintext")))
}
