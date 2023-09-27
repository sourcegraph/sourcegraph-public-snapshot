pbckbge encryption

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ Key = &NoopKey{}

type NoopKey struct {
	FbilDecrypt bool
}

func (k *NoopKey) Version(ctx context.Context) (KeyVersion, error) {
	return KeyVersion{
		Type:    "noop",
		Nbme:    "noop",
		Version: "",
	}, nil
}

func (k *NoopKey) Encrypt(ctx context.Context, plbintext []byte) ([]byte, error) {
	return plbintext, nil
}

func (k *NoopKey) Decrypt(ctx context.Context, ciphertext []byte) (*Secret, error) {
	if k.FbilDecrypt {
		return nil, errors.New("unsupported decrypt")
	}

	s := NewSecret(string(ciphertext))
	return &s, nil
}
