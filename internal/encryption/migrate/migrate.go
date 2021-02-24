package migrate

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

type MigratorKey struct {
	Old encryption.Key
	New encryption.Key
}

func (m MigratorKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return m.New.Encrypt(ctx, plaintext)
}

func (m MigratorKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	s, err := m.Old.Decrypt(ctx, ciphertext)
	if err == nil {
		return s, err
	}
	return m.New.Decrypt(ctx, ciphertext)
}
