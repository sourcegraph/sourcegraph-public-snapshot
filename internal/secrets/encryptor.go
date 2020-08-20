package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"
)

const (
	validKeyLength = 32 // 32 bytes is the required length for AES-256.
	hmacSize       = sha256.Size
)

// EncryptionError is an error about encryption or decryption.
type EncryptionError struct {
	error
}

// TODO: docstring
type Encryptor interface {
	// TODO: docstring
	EncryptBytes(b []byte) ([]byte, error)
	// TODO: docstring
	DecryptBytes(b []byte) ([]byte, error)
}

// encryptor performs encryption and decryption.
type encryptor struct {
	// TODO: docstring
	primaryKey []byte
	// TODO: docstring
	secondaryKey []byte
}

func newEncryptor(primaryKey, secondaryKey []byte) Encryptor {
	return encryptor{
		primaryKey:   primaryKey,
		secondaryKey: secondaryKey,
	}
}

func (e encryptor) EncryptBytes(plaintext []byte) ([]byte, error) {
	// ONLY use the primary key to EncryptBytes
	if len(e.primaryKey) < validKeyLength {
		return nil, &EncryptionError{errors.New("primary key is not available")}
	}

	// Create a one time nonce of standard length, without repetitions
	block, err := aes.NewCipher(e.primaryKey)
	if err != nil {
		return nil, &EncryptionError{errors.Errorf("unable to create new cipher: %v", err)}
	}

	encrypted := make([]byte, aes.BlockSize+len(plaintext))
	nonce := encrypted[:aes.BlockSize]

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, &EncryptionError{errors.Errorf("unable to read nonce: %v", err)}
	}

	stream := cipher.NewCFBEncrypter(block, nonce)
	stream.XORKeyStream(encrypted[aes.BlockSize:], plaintext)

	// Compute HMAC checksum then append to encrypted bytes
	// TODO(Dax): We should stretch the key above rather than try to reuse this
	mac := hmac.New(sha256.New, e.primaryKey)
	_, _ = mac.Write(encrypted)
	macSum := mac.Sum(nil)
	encrypted = append(encrypted, macSum...)

	// return base64.StdEncoding.EncodeToString(encrypted), nil
	return encrypted, nil
}

func (e encryptor) DecryptBytes(ciphertext []byte) ([]byte, error) {
	if len(e.primaryKey) < validKeyLength && len(e.secondaryKey) < validKeyLength {
		return nil, &EncryptionError{errors.New("no valid keys available")}
	}

	// encrypted, err := base64.StdEncoding.DecodeString(encodedValue)
	// if err != nil {
	// 	return "", &EncryptionError{Err: ErrNotEncoded}
	// }

	// Extract HAMC from cipher text
	extractedMac := ciphertext[len(ciphertext)-hmacSize:]
	encryptedValue := ciphertext[:len(ciphertext)-hmacSize]
	if len(encryptedValue) < aes.BlockSize {
		return nil, &EncryptionError{errors.New("invalid block size.")}
	}

	// TODO(Dax): We should stretch the key above rather than try to reuse

	// validate hmac
	validate := func(key []byte) bool {
		mac := hmac.New(sha256.New, key)
		_, _ = mac.Write(encryptedValue)
		expectedMac := mac.Sum(nil)
		return hmac.Equal(extractedMac, expectedMac)
	}

	var keyToDecrypt []byte
	if validate(e.primaryKey) {
		keyToDecrypt = e.primaryKey
	} else if validate(e.secondaryKey) {
		keyToDecrypt = e.secondaryKey
	} else {
		return nil, &EncryptionError{errors.New("no valid key")}
	}

	block, err := aes.NewCipher(keyToDecrypt)
	if err != nil {
		return nil, &EncryptionError{errors.Errorf("unable to create new cipher: %v", err)}
	}

	nonce := encryptedValue[:aes.BlockSize]
	value := encryptedValue[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, nonce)
	stream.XORKeyStream(value, value)
	return value, nil
}

// noOpEncryptor always returns original content and does no encryption or decryption.
type noOpEncryptor struct{}

func (noOpEncryptor) EncryptBytes(b []byte) ([]byte, error) {
	return b, nil
}

func (noOpEncryptor) DecryptBytes(b []byte) ([]byte, error) {
	return b, nil
}

// EncryptBytes encrypts the plaintxt and returns the encrypted value.
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
