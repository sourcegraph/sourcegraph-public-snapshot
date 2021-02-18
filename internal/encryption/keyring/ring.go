package keyring

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/encryption/cloudkms"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewRing creates a keyring.Ring containing all the keys configured in site config
func NewRing(ctx context.Context, keyConfig *schema.EncryptionKeys) (*Ring, error) {
	extsvc, err := NewKey(ctx, keyConfig.ExternalServiceKey)
	if err != nil {
		return nil, err
	}
	return &Ring{
		ExternalServiceKey: extsvc,
	}, nil
}

type Ring struct {
	ExternalServiceKey encryption.Key
}

func NewKey(ctx context.Context, k *schema.EncryptionKey) (encryption.Key, error) {
	if k == nil {
		return nil, fmt.Errorf("cannot configure nil key")
	}
	switch {
	case k.Cloudkms != nil:
		return cloudkms.NewKey(ctx, k.Cloudkms.Keyname)
	case k.Noop != nil:
		return &encryption.NoopKey{}, nil
	default:
		return nil, fmt.Errorf("couldn't configure key: %v", *k)
	}
}
