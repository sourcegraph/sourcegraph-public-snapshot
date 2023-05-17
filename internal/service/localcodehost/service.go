package localcodehost

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/service/servegit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Creates a default external service configuration for the provided pattern.
func ensureExtSVC(observationCtx *observation.Context, config *Config) error {
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, conf.Get().ServiceConnections().PostgresDSN, "servegit")
	if err != nil {
		return errors.Wrap(err, "localcodehost failed to connect to frontend DB")
	}
	store := database.NewDB(observationCtx.Logger, sqlDB).ExternalServices()
	ctx := context.Background()
	serviceConfig, err := json.Marshal(schema.LocalExternalService{
		Repos: config.Repos,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal external service configuration")
	}

	return store.Upsert(ctx, &types.ExternalService{
		ID:          servegit.ExtSVCID,
		Kind:        extsvc.VariantLocalGit.AsKind(),
		DisplayName: "Local repositories",
		Config:      extsvc.NewUnencryptedConfig(string(serviceConfig)),
	})
}

type Config struct {
	env.BaseConfig
	Repos []*schema.LocalRepoPattern
}

func (c *Config) Load() {
	configFilePath := c.Get("SRC_LOCAL_REPOS_CONFIG_FILE", "", "Path to the local repositories configuration file")

	configJSON, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}
	config, err := extsvc.ParseConfig(extsvc.KindLocal, string(configJSON))
	if err != nil {
		return
	}
	c.Repos = config.(*schema.LocalExternalService).Repos
}

type svc struct{}

func (s *svc) Name() string {
	return "localcodehost"
}

func (s *svc) Configure() (env.Config, []debugserver.Endpoint) {
	c := &Config{}
	c.Load()
	return c, nil
}

func (s *svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, envConf env.Config) (err error) {
	config := envConf.(*Config)

	if len(config.Repos) > 0 {
		if err := ensureExtSVC(observationCtx, config); err != nil {
			return errors.Wrap(err, "failed to create external service which imports local repositories")
		}
	}

	return nil
}

var Service = &svc{}
var _ service.Service = Service
