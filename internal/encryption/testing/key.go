package testing

import (
	"context"
	"encoding/base64"

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
