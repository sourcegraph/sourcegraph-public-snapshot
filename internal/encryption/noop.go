package encryption

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ Key = &NoopKey{}

type NoopKey struct {
	FailDecrypt bool
}

func (k *NoopKey) Version(ctx context.Context) (KeyVersion, error) {
	return KeyVersion{
		Type:    "noop",
		Name:    "noop",
		Version: "",
	}, nil
}

func (k *NoopKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (k *NoopKey) Decrypt(ctx context.Context, ciphertext []byte) (*Secret, error) {
	if k.FailDecrypt {
		return nil, errors.New("unsupported decrypt")
	}

	s := NewSecret(string(ciphertext))
	return &s, nil
}
