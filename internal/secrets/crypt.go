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
	validKeyLength = 32
	hmacSize       = sha256.Size
	primaryKey     = 0
	secondaryKey   = 1
)

type EncryptionError struct {
	Message string
}

// NotEncodedError means we can test for whether or not a string is encoded, prior to attempting decryption
var NotEncodedError = errors.New("object is not encoded")

// Generate a valid key for AES-256 encryption
func GenerateRandomAESKey() ([]byte, error) {
	b := make([]byte, validKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (err *EncryptionError) Error() string {
	return err.Message
}

// Encryptor contains the encryption key used in encryption and decryption.
type Encryptor struct {
	// the first key is always used to encrypt, attempt to decrypt with every key
	EncryptionKeys [][]byte
}

// Returns an encrypted string.
func (e *Encryptor) encrypt(b []byte) (string, error) {
	// create a one time nonce of standard length, without repetitions
	// ONLY use the primary key to encrypt
	block, err := aes.NewCipher(e.EncryptionKeys[primaryKey])
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

	// encrypt-then-MAC
	// TODO(Dax): We should stretch the key above rather than try to reuse this
	mac := hmac.New(sha256.New, e.EncryptionKeys[primaryKey])
	n, err := mac.Write(encrypted)
	if err != nil {
		return "", err
	}
	fmt.Println("bytes written: ", n)
	macSum := mac.Sum(nil)
	encrypted = append(encrypted, macSum...)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Encrypts the string, returning the encrypted value.
func (e *Encryptor) EncryptBytes(b []byte) (string, error) {
	return e.encrypt(b)
}

// Encrypts the string, returning the encrypted value.
func (e *Encryptor) Encrypt(value string) (string, error) {
	return e.encrypt([]byte(value))
}

func (e *Encryptor) EncryptBytesIfPossible(b []byte) (string, error) {
	if configuredToEncrypt {
		return e.EncryptBytes(b)
	}
	return string(b), nil
}

func EncryptBytesIfPossible(b []byte) (string, error) {
	if configuredToEncrypt {
		return CryptObject.encrypt(b)
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

// EncryptIfPossible encrypts  the string if encryption is configured.
// Returns an error only when encryption is enabled, and encryption fails.
func (e *Encryptor) EncryptIfPossible(value string) (string, error) {
	if configuredToEncrypt {
		return e.Encrypt(value)
	}
	return value, nil
}

func (e *Encryptor) Decrypt(value string) (string, error) {
	return e.decrypt(value)
}

func (e *Encryptor) DecryptBytes(b []byte) (string, error) {
	return e.decrypt(string(b))
}

// Decrypts the string, returning the decrypted value.
func (e *Encryptor) decrypt(encodedValue string) (string, error) {

	// handle plaintext use case
	if len(e.EncryptionKeys) == 0 {
		return encodedValue, nil
	}

	for _, key := range e.EncryptionKeys {
		encrypted, err := base64.StdEncoding.DecodeString(encodedValue)
		if err != nil {
			return "", NotEncodedError
		}

		//remove hmac
		extractedMac := encrypted[len(encrypted)-hmacSize:]
		encrypted = encrypted[:len(encrypted)-hmacSize]

		// validate hmac
		// TODO(Dax): We should stretch the key above rather than try to reuse
		mac := hmac.New(sha256.New, key)
		n, err := mac.Write(encrypted)
		if err != nil {
			return "", err
		}
		fmt.Printf("wrote %d bytes \n", n)
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
	return "", fmt.Errorf("unable to decrypt")

}

// DecryptIfPossible decrypts the string if encryption is configured.
// It returns an error only if encryption is enabled and it cannot decrypt the string
func (e *Encryptor) DecryptIfPossible(value string) (string, error) {
	if configuredToEncrypt {
		return e.Decrypt(value)
	}
	return value, nil
}

func (e *Encryptor) DecryptBytesIfPossible(b []byte) (string, error) {
	if configuredToEncrypt {
		return e.DecryptBytes(b)
	}
	return string(b), nil
}

// This function rotates the encryption used on an item by decrypting and then re-encrypting.
// Rotating keys updates the EncryptionKey within the Encryptor object
//func (e *Encryptor) RotateKey(newKey []byte, encryptedValue string) (string, error) {
//	decrypted, err := e.Decrypt(encryptedValue)
//	if err != nil {
//		return "", err
//	}
//
//	e.EncryptionKey = newKey
//	return e.encrypt([]byte(decrypted))
//}
