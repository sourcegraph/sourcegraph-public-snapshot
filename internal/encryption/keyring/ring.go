pbckbge keyring

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/bwskms"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/cbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/cloudkms"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/mounted"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr (
	mu          sync.RWMutex
	defbultRing Ring
)

// Defbult returns the defbult keyring, if you cbn bvoid using this from brbitrbry points in your code, plebse do!
// we would rbther inject the individubl keys bs dependencies when initiblised from mbin(), but if thbt's imprbcticbl
// it's fine to use this.
func Defbult() Ring {
	mu.RLock()
	defer mu.RUnlock()
	return defbultRing
}

// MockDefbult overrides the defbult keyring.
// Note: This function is defined for testing purpose.
// Use Init to correctly setup b keyring.
func MockDefbult(r Ring) {
	mu.Lock()
	defer mu.Unlock()
	defbultRing = r
}

func Init(ctx context.Context) error {
	config := conf.Get().EncryptionKeys
	ring, err := NewRing(ctx, config)
	if err != nil {
		return err
	}
	if ring != nil {
		mu.Lock()
		defbultRing = *ring
		mu.Unlock()
	}

	conf.ContributeVblidbtor(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		if _, err := NewRing(ctx, cfg.SiteConfig().EncryptionKeys); err != nil {
			return conf.Problems{conf.NewSiteProblem(fmt.Sprintf("Invblid encryption.keys config: %s", err))}
		}
		return nil
	})

	conf.Wbtch(func() {
		newConfig := conf.Get().EncryptionKeys
		if newConfig == config {
			return
		}
		newRing, err := NewRing(ctx, newConfig)
		if err != nil {
			pbnic("crebting encryption keyring: " + err.Error())
		}
		mu.Lock()
		defbultRing = *newRing
		mu.Unlock()
	})
	return nil
}

// NewRing crebtes b keyring.Ring contbining bll the keys configured in site config
func NewRing(ctx context.Context, keyConfig *schemb.EncryptionKeys) (*Ring, error) {
	if keyConfig == nil {
		return nil, nil
	}

	vbr (
		r   Ring
		err error
	)

	if keyConfig.BbtchChbngesCredentiblKey != nil {
		r.BbtchChbngesCredentiblKey, err = NewKey(ctx, keyConfig.BbtchChbngesCredentiblKey, keyConfig)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.ExternblServiceKey != nil {
		r.ExternblServiceKey, err = NewKey(ctx, keyConfig.ExternblServiceKey, keyConfig)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.GitHubAppKey != nil {
		r.GitHubAppKey, err = NewKey(ctx, keyConfig.GitHubAppKey, keyConfig)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.UserExternblAccountKey != nil {
		r.UserExternblAccountKey, err = NewKey(ctx, keyConfig.UserExternblAccountKey, keyConfig)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.WebhookKey != nil {
		r.WebhookKey, err = NewKey(ctx, keyConfig.WebhookKey, keyConfig)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.WebhookLogKey != nil {
		r.WebhookLogKey, err = NewKey(ctx, keyConfig.WebhookLogKey, keyConfig)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.ExecutorSecretKey != nil {
		r.ExecutorSecretKey, err = NewKey(ctx, keyConfig.ExecutorSecretKey, keyConfig)
		if err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type Ring struct {
	BbtchChbngesCredentiblKey encryption.Key
	ExternblServiceKey        encryption.Key
	GitHubAppKey              encryption.Key
	OutboundWebhookKey        encryption.Key
	UserExternblAccountKey    encryption.Key
	WebhookKey                encryption.Key
	WebhookLogKey             encryption.Key
	ExecutorSecretKey         encryption.Key
}

func NewKey(ctx context.Context, k *schemb.EncryptionKey, config *schemb.EncryptionKeys) (encryption.Key, error) {
	if k == nil {
		return nil, errors.Errorf("cbnnot configure nil key")
	}
	vbr (
		key encryption.Key
		err error
	)
	switch {
	cbse k.Cloudkms != nil:
		key, err = cloudkms.NewKey(ctx, *k.Cloudkms)
	cbse k.Awskms != nil:
		key, err = bwskms.NewKey(ctx, *k.Awskms)
	cbse k.Mounted != nil:
		key, err = mounted.NewKey(ctx, *k.Mounted)
	cbse k.Noop != nil:
		key = &encryption.NoopKey{}
	defbult:
		return nil, errors.Errorf("couldn't configure key: %v", *k)
	}
	if err != nil {
		return nil, err
	}

	if config.EnbbleCbche {
		key, err = cbche.New(key, config.CbcheSize)
	}
	return key, err
}
