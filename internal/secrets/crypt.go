package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/inconshreveable/log15"
)

const (
	validKeyLength    = 32
	hmacSize          = sha256.Size
	primaryKeyIndex   = 0
	secondaryKeyIndex = 1
)

type EncryptionError struct {
	Message string
}

// ErrNotEncoded means we can test for whether or not a string is encoded, prior to attempting decryption
var ErrNotEncoded = errors.New("object is not encoded")

func (err *EncryptionError) Error() string {
	return err.Message
}

// Generate a valid key for AES-256 encryption
func GenerateRandomAESKey() ([]byte, error) {
	b := make([]byte, validKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// EncryptBytesIfPossible encrypts the byte array if encryption is configured.
// Returns an error only when encryption is enabled, and encryption fails.
func EncryptBytesIfPossible(b []byte) (string, error) {
	if configuredToEncrypt {
		return CryptObject.EncryptBytes(b)
	}
	return string(b), nil
}

// EncryptIfPossible encrypts the string if encryption is configured.
// Returns an error only when encryption is enabled, and encryption fails.
func EncryptIfPossible(value string) (string, error) {
	if configuredToEncrypt {
		return CryptObject.Encrypt(value)
	}
	return value, nil
}

// DecryptIfPossible decrypts the string if encryption is configured.
// It returns an error only if encryption is enabled and it cannot Decrypt the string
func DecryptIfPossible(value string) (string, error) {
	if configuredToEncrypt {
		return CryptObject.Decrypt(value)
	}
	return value, nil
}

// DecryptBytesIfPossible decrypts the byte array if encryption is configured.
// It returns an error only if encryption is enabled and it cannot Decrypt the string
func DecryptBytesIfPossible(b []byte) (string, error) {
	if configuredToEncrypt {
		return CryptObject.DecryptBytes(b)
	}
	return string(b), nil
}

// Encryptor contains the encryption key used in encryption and decryption.
type Encryptor struct {
	// the first key is always used to EncryptBytes, attempt to Decrypt with every key
	EncryptionKeys [][]byte
}

// Returns an encrypted string.
func (e *Encryptor) EncryptBytes(b []byte) (string, error) {

	if len(e.EncryptionKeys) == 0 || e.EncryptionKeys[primaryKeyIndex] == nil {
		return string(b), nil
	}

	// create a one time nonce of standard length, without repetitions
	// ONLY use the primary key to EncryptBytes
	block, err := aes.NewCipher(e.EncryptionKeys[primaryKeyIndex])
	if err != nil {
		return "", err
	}

	encrypted := make([]byte, aes.BlockSize+len(b))
	nonce := encrypted[:aes.BlockSize]

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, nonce)
	stream.XORKeyStream(encrypted[aes.BlockSize:], b)

	// EncryptBytes-then-MAC
	// TODO(Dax): We should stretch the key above rather than try to reuse this
	mac := hmac.New(sha256.New, e.EncryptionKeys[primaryKeyIndex])
	_, err = mac.Write(encrypted)
	if err != nil {
		return "", err
	}
	macSum := mac.Sum(nil)
	encrypted = append(encrypted, macSum...)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Encrypts the string, returning the encrypted value.
func (e *Encryptor) Encrypt(value string) (string, error) {
	return e.EncryptBytes([]byte(value))
}

// EncryptBytesIfPossible encrypts the byte array if encryption is configured.
// Returns an error only when encryption is enabled, and encryption fails.
func (e *Encryptor) EncryptBytesIfPossible(b []byte) (string, error) {
	if configuredToEncrypt {
		return e.EncryptBytes(b)
	}
	return string(b), nil
}

// EncryptIfPossible encrypts  the string if encryption is configured.
// Returns an error only when encryption is enabled, and encryption fails.
func (e *Encryptor) EncryptIfPossible(value string) (string, error) {
	if configuredToEncrypt {
		return e.Encrypt(value)
	}
	return value, nil
}

func (e *Encryptor) DecryptBytes(b []byte) (string, error) {
	return e.Decrypt(string(b))
}

// Decrypts the string, returning the decrypted value.
func (e *Encryptor) Decrypt(encodedValue string) (string, error) {

	// handle plaintext use case
	if len(e.EncryptionKeys) == 0 {
		return encodedValue, nil
	}

	for _, key := range e.EncryptionKeys {
		encrypted, err := base64.StdEncoding.DecodeString(encodedValue)
		if err != nil {
			return "", ErrNotEncoded
		}

		//remove hmac
		extractedMac := encrypted[len(encrypted)-hmacSize:]
		encrypted = encrypted[:len(encrypted)-hmacSize]

		// validate hmac
		// TODO(Dax): We should stretch the key above rather than try to reuse
		mac := hmac.New(sha256.New, key)
		_, err = mac.Write(encrypted)
		if err != nil {
			return "", err
		}
		expectedMac := mac.Sum(nil)
		if !hmac.Equal(extractedMac, expectedMac) {
			log15.Warn("mac doesn't match, may retry")
			continue
		}

		if len(encrypted) < aes.BlockSize {
			return "", &EncryptionError{Message: "Invalid block size."}
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			return "", nil
		}

		nonce := encrypted[:aes.BlockSize]
		value := encrypted[aes.BlockSize:]
		stream := cipher.NewCFBDecrypter(block, nonce)
		stream.XORKeyStream(value, value)
		return string(value), nil
	}
	return "", fmt.Errorf("unable to Decrypt")

}

// DecryptIfPossible decrypts the string if encryption is configured.
// It returns an error only if encryption is enabled and it cannot Decrypt the string
func (e *Encryptor) DecryptIfPossible(value string) (string, error) {
	if configuredToEncrypt {
		return e.Decrypt(value)
	}
	return value, nil
}

// DecryptBytesIfPossible decrypts the byte array if encryption is configured.
// It returns an error only if encryption is enabled and it cannot Decrypt the string
func (e *Encryptor) DecryptBytesIfPossible(b []byte) (string, error) {
	if configuredToEncrypt {
		return e.DecryptBytes(b)
	}
	return string(b), nil
}

func (e *Encryptor) Raw(crypt string) (string, error) {
	plaintext, err := e.Decrypt(crypt)
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

// Return a masked version of the decrypted secret, for when a token-like string needs to be displayed in the UI
func (e *Encryptor) Mask(crypt string) (string, error) {
	plaintext, err := e.Decrypt(crypt)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s***********", plaintext[0:0]), nil
}
