package aeshelper

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAESEncryptionRoundTrip(t *testing.T) {
	secret := []byte("_a_secret_that_is_32_bytes_long_")
	t.Run("basic case", func(t *testing.T) {
		plaintext := []byte("some data to encrypt")
		ciphertext, nonce, err := Encrypt(secret, plaintext)
		if err != nil {
			t.Fatal(err)
		}
		decryptedPlaintext, err := Decrypt(secret, ciphertext, nonce)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, string(plaintext), string(decryptedPlaintext))
	})
	t.Run("empty plaintext", func(t *testing.T) {
		plaintext := []byte("")
		ciphertext, nonce, err := Encrypt(secret, plaintext)
		if err != nil {
			t.Fatal(err)
		}
		decryptedPlaintext, err := Decrypt(secret, ciphertext, nonce)
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
		ciphertext, nonce, err := Encrypt(secret, plaintext)
		if err != nil {
			t.Fatal(err)
		}
		decryptedPlaintext, err := Decrypt(secret, ciphertext, nonce)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, string(plaintext), string(decryptedPlaintext))
	})
}
