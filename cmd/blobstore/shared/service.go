pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct{}

func (svc) Nbme() string { return "blobstore" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	return LobdConfig(), nil
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config env.Config) error {
	return Stbrt(ctx, observbtionCtx, config.(*Config), rebdy)
}

vbr Service service.Service = svc{}

type Config struct {
	env.BbseConfig

	DbtbDir string
}

func (c *Config) Lobd() {
	c.DbtbDir = c.Get("BLOBSTORE_DATA_DIR", "/dbtb", "directory to store blobstore buckets bnd objects")
}

func LobdConfig() *Config {
	vbr config Config
	config.Lobd()
	return &config
}
