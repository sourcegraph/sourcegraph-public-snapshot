package validation

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
)

func init() {
	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		if _, err := keyring.NewRing(context.Background(), cfg.SiteConfig().EncryptionKeys); err != nil {
			return conf.Problems{conf.NewSiteProblem(fmt.Sprintf("Invalid encryption.keys config: %s", err))}
		}
		return nil
	})
}
