package envelope

import (
	"crypto/rand"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/encryption/aeshelper"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Envelope contains all necessary information to decrypt the encrypted ciphertext
// that was encrypted using envelope encryption. It is used to encrypt large
// payloads of data with a small key.
type Envelope struct {
	Key        []byte
	Nonce      []byte
	Ciphertext []byte
}

// Encrypt encrypts the plaintext using envelope encryption.
func Encrypt(plaintext []byte) (*Envelope, error) {
	// Generate a 32 byte secret for AES-256 encryption.
	key, err := generateSecret()
	if err != nil {
		return nil, errors.Wrap(err, "generate encryption key")
	}

	ciphertext, nonce, err := aeshelper.Encrypt(key, plaintext)
	if err != nil {
		return nil, errors.Wrap(err, "encrypt plaintext")
	}

	return &Envelope{
		Key:        key,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}, nil
}

var MockGenerateSecret func() ([]byte, error)

// generateSecret generates a 32 byte secret for AES-256 encryption.
func generateSecret() ([]byte, error) {
	if MockGenerateSecret != nil {
		return MockGenerateSecret()
	}

	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	return key, err
}

// Encrypt decrypts the data from the given encrypted envelope.
func Decrypt(envelope *Envelope) ([]byte, error) {
	plaintext, err := aeshelper.Decrypt(envelope.Key, envelope.Ciphertext, envelope.Nonce)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt ciphertext")
	}
	return plaintext, nil
}
