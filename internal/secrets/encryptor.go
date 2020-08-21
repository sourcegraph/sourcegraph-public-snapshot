package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/pkg/errors"
)

const (
	validKeyLength = 32 // 32 bytes is the required length for AES-256.
)

// EncryptionError is an error about encryption or decryption.
type EncryptionError struct {
	error
}

// Encryptor is an interface that provides encryption & decryption primitives
type Encryptor interface {
	EncryptBytes(b []byte) ([]byte, error)
	DecryptBytes(b []byte) ([]byte, error)
}

// encryptor performs encryption and decryption.
type encryptor struct {
	// primaryKey is always used for encryption
	primaryKey []byte
	// secondaryKey is used during key rotation to provide decryption during key rotations.
	// It was the primary key that was used for encryption before the key rotation.
	secondaryKey []byte
}

func newEncryptor(primaryKey, secondaryKey []byte) Encryptor {
	return encryptor{
		primaryKey:   primaryKey,
		secondaryKey: secondaryKey,
	}
}

// Encrypt encrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
// From https://github.com/gtank/cryptopasta/blob/1f550f6f2f69009f6ae57347c188e0a67cd4e500/encrypt.go#L37
func (e encryptor) EncryptBytes(plaintext []byte) (ciphertext []byte, err error) {
	// ONLY use the primary key to EncryptBytes
	if len(e.primaryKey) < validKeyLength {
		return nil, &EncryptionError{errors.New("primary key is not available")}
	}

	block, err := aes.NewCipher(e.primaryKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
// From https://github.com/gtank/cryptopasta/blob/1f550f6f2f69009f6ae57347c188e0a67cd4e500/encrypt.go#L60
func (e encryptor) DecryptBytes(ciphertext []byte) (plaintext []byte, err error) {
	if len(e.primaryKey) < validKeyLength && len(e.secondaryKey) < validKeyLength {
		return nil, &EncryptionError{errors.New("no valid keys available")}
	}

	validate := func(key, ciphertext []byte) ([]byte, error) {
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		if len(ciphertext) < gcm.NonceSize() {
			return nil, errors.New("malformed ciphertext")
		}

		return gcm.Open(nil,
			ciphertext[:gcm.NonceSize()],
			ciphertext[gcm.NonceSize():],
			nil,
		)
	}
	if plaintext, err = validate(e.primaryKey, ciphertext); err == nil {
		return plaintext, nil
	}
	if plaintext, err = validate(e.secondaryKey, ciphertext); err == nil {
		return plaintext, nil
	}
	return nil, &EncryptionError{err}
}

// noOpEncryptor always returns original content and does no encryption or decryption.
type noOpEncryptor struct{}

func (noOpEncryptor) EncryptBytes(b []byte) ([]byte, error) {
	return b, nil
}

func (noOpEncryptor) DecryptBytes(b []byte) ([]byte, error) {
	return b, nil
}

// EncryptBytes encrypts the plaintext and returns the encrypted value.
// An error is returned when encryption fails. It returns the original
// content if encryption is not configured.
func EncryptBytes(plaintext []byte) ([]byte, error) {
	return defaultEncryptor.EncryptBytes(plaintext)
}

// DecryptBytes decrypts the ciphertext and returns the decrypted value.
// An error is returned when decryption fails. It returns the original
// content if encryption is not configured.
func DecryptBytes(ciphertext []byte) ([]byte, error) {
	return defaultEncryptor.DecryptBytes(ciphertext)
}
