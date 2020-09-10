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
	// ConfiguredToEncrypt returns true if the encryptor is able to encrypt
	ConfiguredToEncrypt() bool
	// ConfiguredToRotate returns if primary and secondary keys are valid keys
	ConfiguredToRotate() bool
	// DecryptBytes decrypts a ciphertext with the keys available to the encryptor
	DecryptBytes(b []byte) ([]byte, error)
	// EncryptBytes encrypts a plaintext with the primary key
	EncryptBytes(b []byte) ([]byte, error)
	// EncryptWithKey encrypts plaintext with the given key
	EncryptWithKey(b, key []byte) ([]byte, error)
	// RotateEncryption decrypts given byte array and then re-encrypts with the primary key
	RotateEncryption(b []byte) ([]byte, error)
}

// encryptor performs encryption and decryption.
type encryptor struct {
	// primaryKey is always used for encryption
	primaryKey []byte
	// secondaryKey is used during key rotation to provide decryption during key rotations.
	// It was the primary key that was used for encryption before the key rotation.
	secondaryKey []byte
}

// gcmEncrypt is the general purpose encryption function used to
// return the encrypted versions of bytes.
// gcmEncrypt uses 256-bit AES-GCM. This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the form
// `nonce|ciphertext|tag` where '|' indicates concatenation. It is a modified version of
// https://github.com/gtank/cryptopasta/blob/1f550f6f2f69009f6ae57347c188e0a67cd4e500/encrypt.go#L37
func gcmEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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

// gcmDecrypt decrypts data using 256-bit AES-GCM. This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input form
// `nonce|ciphertext|tag` where '|' indicates concatenation. It is a modified version of
// https://github.com/gtank/cryptopasta/blob/1f550f6f2f69009f6ae57347c188e0a67cd4e500/encrypt.go#L60
func gcmDecrypt(ciphertext, key []byte) ([]byte, error) {
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

	return gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], nil)
}

func newEncryptor(primaryKey, secondaryKey []byte) Encryptor {
	return encryptor{
		primaryKey:   primaryKey,
		secondaryKey: secondaryKey,
	}
}

// ConfiguredToEncrypt returns the statue of our encryptor, whether or not
// it has a key specified, and can thus encrypt.
func (e encryptor) ConfiguredToEncrypt() bool {
	return len(e.primaryKey) == validKeyLength
}

func (e encryptor) ConfiguredToRotate() bool {
	return len(e.primaryKey) == validKeyLength && len(e.secondaryKey) == validKeyLength
}

// EncryptBytes encrypts the plaintext using the primaryKey of the encryptor. This
// relies on the AES-GCM encryption defined in encrypt, within this package.
func (e encryptor) EncryptBytes(plaintext []byte) (ciphertext []byte, err error) {
	if len(e.primaryKey) < validKeyLength {
		return nil, &EncryptionError{errors.New("primary key is unavailable")}
	}

	return gcmEncrypt(plaintext, e.primaryKey)
}

// DecryptBytes decrypts the plaintext using the primaryKey of the encryptor.
// This relies on AES-GCM.
func (e encryptor) DecryptBytes(ciphertext []byte) (plaintext []byte, err error) {
	if len(e.primaryKey) < validKeyLength && len(e.secondaryKey) < validKeyLength {
		return nil, &EncryptionError{errors.New("no valid keys available")}
	}

	if plaintext, err = gcmDecrypt(ciphertext, e.primaryKey); err == nil {
		return plaintext, nil
	}
	if plaintext, err = gcmDecrypt(ciphertext, e.secondaryKey); err == nil {
		return plaintext, nil
	}
	return nil, &EncryptionError{err}

}

func (e encryptor) EncryptWithKey(plaintext, key []byte) ([]byte, error) {
	return gcmEncrypt(plaintext, key)
}

// RotateEncryption rotates the encryption on a ciphertext by
// decrypting the byte array using the primaryKey, and then reencrypting
// it using the secondaryKey.
func (e encryptor) RotateEncryption(ciphertext []byte) ([]byte, error) {
	if !e.ConfiguredToRotate() {
		return nil, &EncryptionError{errors.New("key rotation not configured")}
	}
	// try previous key first
	plaintext, err := gcmDecrypt(ciphertext, e.secondaryKey)
	if err != nil {
		_, err = gcmDecrypt(ciphertext, e.primaryKey)
		if err != nil {
			return nil, err
		}
		return ciphertext, nil
	}

	return e.EncryptWithKey(plaintext, e.primaryKey)
}

// noOpEncryptor always returns original content and does no encryption or decryption.
type noOpEncryptor struct{}

func (noOpEncryptor) EncryptBytes(b []byte) ([]byte, error) {
	return b, nil
}

func (noOpEncryptor) DecryptBytes(b []byte) ([]byte, error) {
	return b, nil
}

func (noOpEncryptor) ConfiguredToEncrypt() bool {
	return false
}

func (noOpEncryptor) ConfiguredToRotate() bool {
	return false
}

func (noOpEncryptor) RotateEncryption(b []byte) ([]byte, error) {
	return b, nil
}

func (noOpEncryptor) EncryptWithKey(b, key []byte) ([]byte, error) {
	return b, nil
}

// ConfiguredToEncrypt returns a boolean indicating whether this type of
// encryption was configured to encrypt. This is effectively the status of
// a struct having an encryption key specified.
func ConfiguredToEncrypt() bool {
	return defaultEncryptor.ConfiguredToEncrypt()
}

// ConfiguredToRotate returns a boolean indicating whether this type of
// encryption was configured to rotate keys. This is effectively the status of
// a struct having two encryption keys specified.
func ConfiguredToRotate() bool {
	return defaultEncryptor.ConfiguredToRotate()
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

func EncryptWithKey(ciphertext, key []byte) ([]byte, error) {
	return defaultEncryptor.EncryptWithKey(ciphertext, key)
}

func RotateEncryption(ciphertext []byte) ([]byte, error) {
	return defaultEncryptor.RotateEncryption(ciphertext)
}
