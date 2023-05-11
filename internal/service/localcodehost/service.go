package localcodehost

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

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
	configFilePath := c.Get("LOCAL_REPOS_CONFIG_FILE", "", "Path to the local repositories configuration file")

	file, err := os.Open(configFilePath)
	if err != nil {
		return
	}
	c.Repos = parseConfig(file)
}

func parseConfig(input io.Reader) []*schema.LocalRepoPattern {
	repos := make([]*schema.LocalRepoPattern, 0)
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		pattern := line
		group := ""

		separatorIndex := -1

		if line[0] == '"' {
			line = line[1:]
			separatorIndex = strings.Index(line, "\"")
		} else {
			separatorIndex = strings.Index(line, " ")
		}

		if separatorIndex > -1 {
			pattern = line[0:separatorIndex]
			group = line[separatorIndex+1:]
		}

		repos = append(repos, &schema.LocalRepoPattern{Pattern: strings.TrimSpace(pattern), Group: strings.TrimSpace(group)})
	}
	return repos
}

type svc struct {
	srvReady chan (any)
}

func (s *svc) Name() string {
	return "localcodehost"
}

func (s *svc) Configure() (env.Config, []debugserver.Endpoint) {
	c := &Config{}
	c.Load()
	return c, nil
}

func (s *svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, configI env.Config) (err error) {

	config := configI.(*Config)

	if len(config.Repos) > 0 {
		if err := ensureExtSVC(observationCtx, config); err != nil {
			return errors.Wrap(err, "failed to create external service which imports local repositories")
		}
	}

	return nil
}

var Service = &svc{
	srvReady: make(chan any),
}
var _ service.Service = Service
