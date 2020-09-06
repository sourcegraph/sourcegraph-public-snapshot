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
	ConfiguredToEncrypt() bool
	ConfiguredToRotate() bool
	DecryptBytes(b []byte) ([]byte, error)
	EncryptBytes(b []byte) ([]byte, error)
	EncryptWithKey(b, k []byte) ([]byte, error)
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

// EncryptBytes is the general purpose encryption function used to
// return the encrypted versions of bytes.
// EncryptBytes uses 256-bit AES-GCM. This both hides the content of
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
	if len(e.primaryKey) < validKeyLength {
		return nil, &EncryptionError{errors.New("no valid keys available")}
	}

	return gcmDecrypt(e.primaryKey, ciphertext)
}

func (e encryptor) EncryptWithKey(plaintext, key []byte) ([]byte, error) {
	return gcmEncrypt(plaintext, key)
}

// RotateEncryption rotates the encryption on a ciphertext by
// decrypting the byte array using the primaryKey, and then reencrypting
// it using the secondaryKey.
func (e encryptor) RotateEncryption(ciphertext []byte) ([]byte, error) {
	if !e.ConfiguredToRotate() {
		return nil, &EncryptionError{errors.New("key rotatation not configured")}
	}
	plaintext, err := gcmDecrypt(ciphertext, e.primaryKey)
	if err != nil { // perhaps it's already encrypted?
		_, err = gcmDecrypt(ciphertext, e.secondaryKey)
		if err == nil {
			return ciphertext, nil
		}
		return ciphertext, err
	}

	return e.EncryptWithKey(plaintext, e.secondaryKey)
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

func (noOpEncryptor) EncryptWithKey(b, k []byte) ([]byte, error) {
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
