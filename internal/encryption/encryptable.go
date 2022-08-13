package encryption

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Encryptable struct {
	mutex          sync.Mutex
	decryptedValue *decryptedValue
	encryptedValue *EncryptedValue
	key            Key
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
		decryptedValue: &decryptedValue{value, nil},
		key:            key,
	}
}

func NewEncrypted(cipher, keyID string, key Key) *Encryptable {
	return &Encryptable{
		encryptedValue: &EncryptedValue{cipher, keyID},
		key:            key,
	}
}

func (e *Encryptable) Decrypted(ctx context.Context) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.decrypted(ctx)
}

func (e *Encryptable) decrypted(ctx context.Context) (string, error) {
	if e.decryptedValue != nil {
		return e.decryptedValue.value, e.decryptedValue.err
	}
	if e.encryptedValue == nil {
		return "", errors.New("no encrypted value")
	}

	value, err := MaybeDecrypt(ctx, e.key, e.encryptedValue.Cipher, e.encryptedValue.KeyID)
	e.decryptedValue = &decryptedValue{value, err}
	return value, err
}

func (e *Encryptable) Encrypted(ctx context.Context, key Key) (string, string, error) {
	if err := e.SetKey(ctx, key); err != nil {
		return "", "", err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.encryptedValue != nil {
		return e.encryptedValue.Cipher, e.encryptedValue.KeyID, nil
	}
	if e.decryptedValue == nil {
		return "", "", errors.New("nothing to encrypt")
	}

	cipher, keyID, err := MaybeEncrypt(ctx, e.key, e.decryptedValue.value)
	if err != nil {
		return "", "", err
	}

	e.encryptedValue = &EncryptedValue{cipher, keyID}
	return cipher, keyID, err
}

func (e *Encryptable) Set(value string) {
	e.mutex.Lock()
	e.decryptedValue = &decryptedValue{value, nil}
	e.encryptedValue = nil
	e.mutex.Unlock()
}

func (e *Encryptable) SetKey(ctx context.Context, key Key) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.key == key {
		return nil
	}

	if _, err := e.decrypted(ctx); err != nil {
		return err
	}

	e.key = key
	e.encryptedValue = nil
	return nil
}
