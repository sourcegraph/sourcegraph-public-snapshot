package encryption

import (
	"context"
	"encoding/json"
)

type JSONEncryptable[T any] struct {
	*Encryptable
}

func NewUnencryptedJSON[T any](value T) (*JSONEncryptable[T], error) {
	return NewUnencryptedJSONWithKey(value, nil)
}

func NewUnencryptedJSONWithKey[T any](value T, key Key) (*JSONEncryptable[T], error) {
	serialized, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return &JSONEncryptable[T]{Encryptable: NewUnencryptedWithKey(string(serialized), key)}, nil
}

func NewEncryptedJSON[T any](cipher, keyID string, key Key) *JSONEncryptable[T] {
	return &JSONEncryptable[T]{Encryptable: NewEncrypted(cipher, keyID, key)}
}

func (e *JSONEncryptable[T]) Decrypted(ctx context.Context) (value T, _ error) {
	serialized, err := e.Encryptable.Decrypted(ctx)
	if err != nil {
		return value, err
	}

	if err := json.Unmarshal([]byte(serialized), &value); err != nil {
		return value, err
	}

	return value, nil
}

func (e *JSONEncryptable[T]) DecryptedInto(ctx context.Context, value T) error {
	serialized, err := e.Encryptable.Decrypted(ctx)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(serialized), &value); err != nil {
		return err
	}

	return nil
}

func (e *JSONEncryptable[T]) Set(value T) error {
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.decrypted = &decryptedValue{string(serialized), nil}
	e.encrypted = nil
	return nil
}
