package aeshelper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Encrypt encrypts the plaintext using the given secret.
// secret should be exactly 32 bytes to encrypt using AES-256.
func Encrypt(secret, plaintext []byte) (ciphertext, nonce []byte, _ error) {
	if len(secret) != 32 {
		return nil, nil, errors.New("secret is not 32 characters long")
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, nil, errors.Wrap(err, "generate new AES cipher")
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, errors.Wrap(err, "generate new GCM cipher")
	}

	nonce = make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, errors.Wrap(err, "generate nonce")
	}

	ciphertext = aesGCM.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// Decrypt decrypts the data from the given aes encrypted value.
// secret should be exactly 32 bytes.
func Decrypt(secret, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, errors.Wrap(err, "generate new AES cipher")
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "generate new GCM cipher")
	}

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
