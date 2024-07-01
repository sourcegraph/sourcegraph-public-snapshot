package encryption

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Encryptable wraps a value and an encryption key and handles lazily encrypting and
// decrypting that value. This struct should be used in all places where a value is
// encrypted at-rest to maintain a consistent handling of data with security concerns.
//
// This struct should always be passed by reference.
type Encryptable struct {
	mutex     sync.Mutex
	decrypted *decryptedValue
	encrypted *encryptedValue
	key       Key
}

type decryptedValue struct {
	value string
	err   error
}

// encryptedValue wraps an encrypted value and serialized metadata about that key that
// encrypted it.
type encryptedValue struct {
	Cipher string
	KeyID  string
}

// NewUnencrypted creates a new encryptable from a plaintext value.
func NewUnencrypted(value string) *Encryptable {
	return &Encryptable{
		decrypted: &decryptedValue{value, nil},
	}
}

// NewEncrypted creates a new encryptable from an encrypted value and a relevant encryption key.
func NewEncrypted(cipher, keyID string, key Key) *Encryptable {
	return &Encryptable{
		encrypted: &encryptedValue{cipher, keyID},
		key:       key,
	}
}

// Decrypt returns the underlying plaintext value. This method may make an external API call to
// decrypt the underlying encrypted value, but will memoize the result so that subsequent calls
// will be cheap.
func (e *Encryptable) Decrypt(ctx context.Context) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.decryptLocked(ctx)
}

func (e *Encryptable) decryptLocked(ctx context.Context) (string, error) {
	if e.decrypted != nil {
		return e.decrypted.value, e.decrypted.err
	}
	if e.encrypted == nil {
		return "", errors.New("no encrypted value")
	}

	value, err := MaybeDecrypt(ctx, e.key, e.encrypted.Cipher, e.encrypted.KeyID)
	e.decrypted = &decryptedValue{value, err}
	return value, err
}

// Encrypt returns the underlying encrypted value. This method may make an external API call to
// encrypt the underlying plaintext value, but will memoize the result so that subsequent calls
// will be cheap.
func (e *Encryptable) Encrypt(ctx context.Context, key Key) (string, string, error) {
	if err := e.SetKey(ctx, key); err != nil {
		return "", "", err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.encrypted != nil {
		return e.encrypted.Cipher, e.encrypted.KeyID, nil
	}
	if e.decrypted == nil {
		return "", "", errors.New("nothing to encrypt")
	}

	cipher, keyID, err := MaybeEncrypt(ctx, e.key, e.decrypted.value)
	if err != nil {
		return "", "", err
	}

	e.encrypted = &encryptedValue{cipher, keyID}
	return cipher, keyID, err
}

// Set updates the underlying plaintext value.
func (e *Encryptable) Set(value string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.decrypted = &decryptedValue{value, nil}
	e.encrypted = nil
}

// SetKey updates the encryption key used with the encrypted value. This method may trigger an
// external API call to decrypt the current value.
func (e *Encryptable) SetKey(ctx context.Context, key Key) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.key == key {
		return nil
	}

	if _, err := e.decryptLocked(ctx); err != nil {
		return err
	}

	e.key = key
	e.encrypted = nil
	return nil
}
