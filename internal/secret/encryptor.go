package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"

	"github.com/pkg/errors"
)

const (
	validKeyLength = 32 // 32 bytes is the required length for AES-256.
	separator      = ":"
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
	// Decrypts a cipher, attempting all keys
	Decrypt(b string) (string, error)
	// Encrypt encrypts a plaintext with the primary key
	Encrypt(b string) (string, error)
	// EncryptBytes encrypts a plaintext with the primary key
	EncryptWithKey(b, key []byte) ([]byte, error)
	// Return hash of the primary key to be used when filtering encoding
	PrimaryKeyHash() string
	// Return hash of secondary key to be used in decrypting
	SecondaryKeyHash() string
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
	// primaryKeyHash is prepended to base64-encoded ciphertext with `separator`.
	primaryKeyHash string
	// secondaryKeyHash is the previously hash that was prepened to ciphertext with `seperator`
	secondaryKeyHash string
}

// sliceKeyHash returns a string which is the first 6 bytes of given key's SHA256 checksum in hexadecimal.
func sliceKeyHash(k []byte) string {
	if k == nil {
		return ""
	}

	h := sha256.New()
	h.Write(k)
	return hex.EncodeToString(h.Sum(nil))[:6]
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
		primaryKey:       primaryKey,
		secondaryKey:     secondaryKey,
		primaryKeyHash:   sliceKeyHash(primaryKey),
		secondaryKeyHash: sliceKeyHash(secondaryKey),
	}
}

// PrimaryKeyHash returns the keyHash for the primary key.
func (e encryptor) PrimaryKeyHash() string {
	return e.primaryKeyHash
}

// SecondaryKeyHash returns the keyHash for the secondary key.
func (e encryptor) SecondaryKeyHash() string {
	return e.secondaryKeyHash
}

// ConfiguredToEncrypt returns the status of our encryptor, whether or not
// it has a key specified, and can thus encrypt.
func (e encryptor) ConfiguredToEncrypt() bool {
	return len(e.primaryKey) == validKeyLength
}

// ConfiguredToRotate returns the status of our encryptor. If it contains two keys it
// is configured to rotate.
func (e encryptor) ConfiguredToRotate() bool {
	return len(e.primaryKey) == validKeyLength && len(e.secondaryKey) == validKeyLength
}

// Encrypt encrypts the plaintext using the primaryKey of the encryptor. This
// relies on the AES-GCM encryption defined in encrypt, within this package.
func (e encryptor) Encrypt(plaintext string) (ciphertext string, err error) {
	if len(e.primaryKey) < validKeyLength {
		return "", &EncryptionError{errors.New("primary key is unavailable")}
	}

	cipherbytes, err := gcmEncrypt([]byte(plaintext), e.primaryKey)
	if err != nil {
		return "", &EncryptionError{errors.Errorf("unable to encrypt: %v", err)}
	}

	ciphertext = base64.StdEncoding.EncodeToString(cipherbytes)
	return e.PrimaryKeyHash() + separator + ciphertext, nil
}

// Decrypt decrypts the plaintext using the primaryKey of the encryptor.
// This relies on AES-GCM.
func (e encryptor) Decrypt(ciphertext string) (plaintext string, err error) {
	if len(e.primaryKey) < validKeyLength && len(e.secondaryKey) < validKeyLength {
		return "", &EncryptionError{errors.New("no valid keys available")}
	}

	var keyToDecrypt []byte

	// Check the keyHash prefix to determine if we can decrypt it or whether it is encrypted at all.
	if strings.HasPrefix(ciphertext, e.PrimaryKeyHash()+separator) {
		ciphertext = ciphertext[len(e.PrimaryKeyHash()+separator):]
		keyToDecrypt = e.primaryKey
	} else if strings.HasPrefix(ciphertext, e.SecondaryKeyHash()+separator) {
		ciphertext = ciphertext[len(e.SecondaryKeyHash()+separator):]
		keyToDecrypt = e.secondaryKey
	} else {
		return ciphertext, nil
	}

	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", &EncryptionError{err}
	}

	plainbytes, err := gcmDecrypt([]byte(decodedCiphertext), keyToDecrypt)
	if err != nil {
		return "", &EncryptionError{err}
	}
	return string(plainbytes), nil
}

// EncryptWithKey encrypts the plaintext byte array with a specific key
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

func (noOpEncryptor) PrimaryKeyHash() string {
	return ""
}

func (noOpEncryptor) SecondaryKeyHash() string {
	return ""
}

func (noOpEncryptor) Encrypt(s string) (string, error) {
	return s, nil
}

func (noOpEncryptor) Decrypt(s string) (string, error) {
	return s, nil
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

// Encrypt encrypts the plaintext and returns the encrypted value.
// An error is returned when encryption fails. It returns the original
// content if encryption is not configured.
func Encrypt(plaintext string) (string, error) {
	return defaultEncryptor.Encrypt(plaintext)
}

// Decrypt decrypts the ciphertext and returns the decrypted value.
// An error is returned when decryption fails. It returns the original
// content if encryption is not configured.
func Decrypt(ciphertext string) (string, error) {
	return defaultEncryptor.Decrypt(ciphertext)
}

// EncryptWithKey encrypts the plaintext byte array with a specific key
func EncryptWithKey(ciphertext, key []byte) ([]byte, error) {
	return defaultEncryptor.EncryptWithKey(ciphertext, key)
}

// RotateEncryption rotates the encryption on a ciphertext by
// decrypting the byte array using the primaryKey, and then reencrypting
// it using the secondaryKey.
func RotateEncryption(ciphertext []byte) ([]byte, error) {
	return defaultEncryptor.RotateEncryption(ciphertext)
}

// KeyHash returns the private keyHash variable.
func KeyHash() string {
	return defaultEncryptor.PrimaryKeyHash()
}
