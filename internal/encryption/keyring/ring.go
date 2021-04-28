package keyring

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/cloudkms"
	"github.com/sourcegraph/sourcegraph/internal/encryption/mounted"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	mu          sync.RWMutex
	defaultRing Ring
)

// Default returns the default keyring, if you can avoid using this from arbitrary points in your code, please do!
// we would rather inject the individual keys as dependencies when initialised from main(), but if that's impractical
// it's fine to use this.
func Default() Ring {
	mu.RLock()
	defer mu.RUnlock()
	return defaultRing
}

// MockDefault overrides the default keyring.
// Note: This function is defined for testing purpose.
// Use Init to correctly setup a keyring.
func MockDefault(r Ring) {
	mu.Lock()
	defer mu.Unlock()
	defaultRing = r
}

func Init(ctx context.Context) error {
	config := conf.Get().EncryptionKeys
	ring, err := NewRing(ctx, config)
	if err != nil {
		return err
	}
	if ring != nil {
		mu.Lock()
		defaultRing = *ring
		mu.Unlock()
	}

	conf.ContributeValidator(func(cfg conf.Unified) conf.Problems {
		if _, err := NewRing(ctx, cfg.EncryptionKeys); err != nil {
			return conf.Problems{conf.NewSiteProblem(fmt.Sprintf("Invalid encryption.keys config: %s", err))}
		}
		return nil
	})

	conf.Watch(func() {
		newConfig := conf.Get().EncryptionKeys
		if newConfig == config {
			return
		}
		newRing, err := NewRing(ctx, newConfig)
		if err != nil {
			panic("creating encryption keyring: " + err.Error())
		}
		mu.Lock()
		defaultRing = *newRing
		mu.Unlock()
	})
	return nil
}

// NewRing creates a keyring.Ring containing all the keys configured in site config
func NewRing(ctx context.Context, keyConfig *schema.EncryptionKeys) (*Ring, error) {
	if keyConfig == nil {
		return nil, nil
	}

	var r Ring
	var err error

	if keyConfig.ExternalServiceKey != nil {
		r.ExternalServiceKey, err = NewKey(ctx, keyConfig.ExternalServiceKey)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.UserExternalAccountKey != nil {
		r.UserExternalAccountKey, err = NewKey(ctx, keyConfig.UserExternalAccountKey)
		if err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type Ring struct {
	ExternalServiceKey     encryption.Key
	UserExternalAccountKey encryption.Key
}

func NewKey(ctx context.Context, k *schema.EncryptionKey) (encryption.Key, error) {
	if k == nil {
		return nil, fmt.Errorf("cannot configure nil key")
	}
	switch {
	case k.Cloudkms != nil:
		return cloudkms.NewKey(ctx, *k.Cloudkms)
	case k.Mounted != nil:
		return mounted.NewKey(ctx, *k.Mounted)
	case k.Noop != nil:
		return &encryption.NoopKey{}, nil
	default:
		return nil, fmt.Errorf("couldn't configure key: %v", *k)
	}
}
