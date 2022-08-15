package encryption

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Encryptable struct {
	mutex     sync.Mutex
	decrypted *decryptedValue
	encrypted *EncryptedValue
	key       Key
}

type decryptedValue struct {
	value string
	err   error
}

type EncryptedValue struct {
	Cipher string
	KeyID  string
}

func NewUnencrypted(value string) *Encryptable {
	return NewUnencryptedWithKey(value, nil)
}

func NewUnencryptedWithKey(value string, key Key) *Encryptable {
	return &Encryptable{
		decrypted: &decryptedValue{value, nil},
		key:       key,
	}
}

func NewEncrypted(cipher, keyID string, key Key) *Encryptable {
	return &Encryptable{
		encrypted: &EncryptedValue{cipher, keyID},
		key:       key,
	}
}

func (e *Encryptable) Decrypted(ctx context.Context) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.decryptedLocked(ctx)
}

func (e *Encryptable) decryptedLocked(ctx context.Context) (string, error) {
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

func (e *Encryptable) Encrypted(ctx context.Context, key Key) (string, string, error) {
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

	e.encrypted = &EncryptedValue{cipher, keyID}
	return cipher, keyID, err
}

func (e *Encryptable) Set(value string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.decrypted = &decryptedValue{value, nil}
	e.encrypted = nil
}

func (e *Encryptable) SetKey(ctx context.Context, key Key) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.key == key {
		return nil
	}

	if _, err := e.decryptedLocked(ctx); err != nil {
		return err
	}

	e.key = key
	e.encrypted = nil
	return nil
}
