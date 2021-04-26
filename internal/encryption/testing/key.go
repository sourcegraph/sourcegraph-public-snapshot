package testing

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
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

// TransparentDecrypter is a decrypter that returns its own string value, but
// errors if not called within a test. This allows mocking the decrypter when
// it's only important that it's called, and not what it actually does.
type TransparentDecrypter struct{ called int }

var _ encryption.Decrypter = &TransparentDecrypter{}

func NewTransparentDecrypter(t *testing.T) *TransparentDecrypter {
	dec := &TransparentDecrypter{}
	t.Cleanup(func() {
		if dec.called == 0 {
			t.Error("transparent decrypter was never called")
		}
	})

	return dec
}

func (dec *TransparentDecrypter) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	dec.called++

	secret := encryption.NewSecret(string(ciphertext))
	return &secret, nil
}
