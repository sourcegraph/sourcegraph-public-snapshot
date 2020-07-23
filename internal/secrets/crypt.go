package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type EncryptionError struct {
	Message string
}

func (err *EncryptionError) Error() string {
	return err.Message
}

// Encrypter contains the encryption key used in encryption and decryption
type Encrypter struct {
	EncryptionKey []byte
}

// Returns an enrypted string
func (e *Encrypter) encrypt(key []byte, value string) (string, error) {
	// create a one time nonce of standard length, without repetitions

	byteVal := []byte(value)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	encrypted := make([]byte, aes.BlockSize+len(byteVal))
	nonce := encrypted[:aes.BlockSize]

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, nonce)
	stream.XORKeyStream(encrypted[aes.BlockSize:], byteVal)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Encrypts the string, returning the encrypted value
func (e *Encrypter) Encrypt(value string) (string, error) {
	return e.encrypt(e.EncryptionKey, value)
}

// Decrypts the string, returning the decrypted value
func (e *Encrypter) Decrypt(encodedValue string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encodedValue)
	if err != nil {
		return "", nil
	}

	if len(encrypted) < aes.BlockSize {
		return "", &EncryptionError{Message: "Invalid block size."}
	}

	block, err := aes.NewCipher(e.EncryptionKey)
	if err != nil {
		return "", nil
	}

	nonce := encrypted[:aes.BlockSize]
	value := encrypted[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, nonce)
	stream.XORKeyStream(value, value)
	return string(value), nil
}

// This function rotates the encryption used on an item by decryping and then recencrypting.
// Rotating keys updates the EncryptionKey within the Encrypter object
func (e *Encrypter) RotateKey(newKey []byte, encryptedValue string) (string, error) {
	decrypted, err := e.Decrypt(encryptedValue)
	if err != nil {
		return "", err
	}

	e.EncryptionKey = newKey
	return e.encrypt(newKey, decrypted)
}
