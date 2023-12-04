package testing

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TestKey is an encryption.Key that just base64 encodes the plaintext, to make
// sure the data is actually transformed, so as to be unreadable by
// misconfigured Stores, but doesn't do any encryption.
type TestKey struct{}

var _ encryption.Key = TestKey{}

func (k TestKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(plaintext)), nil
}

func (k TestKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(ciphertext))
	s := encryption.NewSecret(string(decoded))
	return &s, err
}

func (k TestKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "testkey"}, nil
}

// ByteaTestKey is an encryption.Key that wraps TestKey in a way that adds arbitrary
// non-ascii characters. This ensures that we do not try to insert illegal text into
// encrypted bytea columns.
type ByteaTestKey struct{}

var _ encryption.Key = TestKey{}

func (k ByteaTestKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return []byte("\\x20" + base64.StdEncoding.EncodeToString(plaintext)), nil
}

func (k ByteaTestKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	if len(ciphertext) < 4 || string(ciphertext[:4]) != "\\x20" {
		return nil, errors.New("incorrect prefix")
	}

	decoded, err := base64.StdEncoding.DecodeString(string(ciphertext[4:]))
	s := encryption.NewSecret(string(decoded))
	return &s, err
}

func (k ByteaTestKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "byteatestkey"}, nil
}

// BadKey is an encryption.Key that always returns an error when any of its
// methods are invoked.
type BadKey struct{ Err error }

var _ encryption.Key = &BadKey{}

func (k *BadKey) Encrypt(context.Context, []byte) ([]byte, error) {
	return nil, k.Err
}

func (k *BadKey) Decrypt(context.Context, []byte) (*encryption.Secret, error) {
	return nil, k.Err
}

func (k *BadKey) Version(context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{}, k.Err
}

// TransparentKey is a key that performs no encryption or decryption, but errors
// if not called within a test. This allows mocking the decrypter when it's only
// important that it's called, and not what it actually does.
type TransparentKey struct{ called int }

var _ encryption.Key = &TransparentKey{}

func NewTransparentKey(t *testing.T) *TransparentKey {
	dec := &TransparentKey{}
	t.Cleanup(func() {
		if dec.called == 0 {
			t.Error("transparent key was never called")
		}
	})

	return dec
}

func (dec *TransparentKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	dec.called++

	secret := encryption.NewSecret(string(ciphertext))
	return &secret, nil
}

func (dec *TransparentKey) Encrypt(ctx context.Context, value []byte) ([]byte, error) {
	dec.called++
	return value, nil
}

func (dec *TransparentKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "transparent"}, nil
}
