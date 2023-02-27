package encryption

import (
	"context"
	"encoding/json"
)

// JSONEncryptable wraps a value of type T and an encryption key and handles lazily encoding/encrypting
// and decrypting/decoding that value. This struct should be used in all places where a JSON-serialized
// value is encrypted at-rest to maintain a consistent handling of data with security concerns.
//
// This struct should always be passed by reference.
type JSONEncryptable[T any] struct {
	*Encryptable
}

// NewUnencryptedJSON creates a new JSON encryptable from the given value.
func NewUnencryptedJSON[T any](value T) (*JSONEncryptable[T], error) {
	serialized, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return &JSONEncryptable[T]{Encryptable: NewUnencrypted(string(serialized))}, nil
}

// NewEncryptedJSON creates a new JSON encryptable an encrypted value and a relevant encryption key.
func NewEncryptedJSON[T any](cipher, keyID string, key Key) *JSONEncryptable[T] {
	return &JSONEncryptable[T]{Encryptable: NewEncrypted(cipher, keyID, key)}
}

// Decrypt decrypts and returns the underlying value as a T. This method may make an external API call
// to decrypt the underlying encrypted value, but will memoize the result so that subsequent calls will
// be cheap.
func (e *JSONEncryptable[T]) Decrypt(ctx context.Context) (value T, _ error) {
	serialized, err := e.Encryptable.Decrypt(ctx)
	if err != nil {
		return value, err
	}

	if err := json.Unmarshal([]byte(serialized), &value); err != nil {
		return value, err
	}

	return value, nil
}

// Set updates the underlying value.
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

// DecryptJSON decrypts the encryptable value. This method may make an external
// API call to decrypt the underlying encrypted value, but will memoize the result so that subsequent calls
// will be cheap.
func DecryptJSON[T any](ctx context.Context, e *JSONEncryptable[any]) (*T, error) {
	var value T

	serialized, err := e.Encryptable.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(serialized), &value); err != nil {
		return nil, err
	}

	return &value, nil
}
