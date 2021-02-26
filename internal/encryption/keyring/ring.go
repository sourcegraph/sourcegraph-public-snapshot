package keyring

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/encryption/cloudkms"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	mu          sync.RWMutex
	defaultRing Ring
)

func Default() Ring {
	mu.RLock()
	defer mu.RUnlock()
	return defaultRing
}

func Init(ctx context.Context) error {
	config := conf.Get().EncryptionKeys
	ring, err := NewRing(ctx, config)
	if err != nil {
		return err
	}
	if ring != nil {
		defaultRing = *ring
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
