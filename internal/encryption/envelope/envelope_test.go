package envelope

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvelopeEncryptionRoundTrip(t *testing.T) {
	t.Run("basic case", func(t *testing.T) {
		plaintext := []byte("some data to encrypt")
		ev, err := Encrypt(plaintext)
		if err != nil {
			t.Fatal(err)
		}
		decryptedPlaintext, err := Decrypt(ev)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, string(plaintext), string(decryptedPlaintext))
	})
	t.Run("large payload", func(t *testing.T) {
		plaintext := make([]byte, 100000)
		if _, err := io.ReadFull(rand.Reader, plaintext); err != nil {
			t.Fatal(err)
		}
		ev, err := Encrypt(plaintext)
		if err != nil {
			t.Fatal(err)
		}
		decryptedPlaintext, err := Decrypt(ev)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, string(plaintext), string(decryptedPlaintext))
	})
}
