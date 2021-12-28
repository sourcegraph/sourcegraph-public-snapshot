package migrators

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

// invalidKey is an encryption.Key that just base64 encodes the plaintext,
// but silently fails to decrypt the secret.
type invalidKey struct{}

func (k invalidKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(plaintext)), nil
}

func (k invalidKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	s := encryption.NewSecret(string(ciphertext))
	return &s, nil
}

func (k invalidKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "invalidkey"}, nil
}
