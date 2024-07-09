package modelconfig

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/embedded"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Init is the initalization function wired into the `frontend` application startup.
// This registers the necessary watchers and hooks so that the `Service` can always
// have an up-to-date view of this Sourcegraph instance's configuration data.
func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	codeIntelServices codeintel.Services,
	initialConf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	// Ensure that we only regsiter this once.
	if singletonConfigService != nil {
		return errors.New("the Init function has already been called")
	}

	logger := log.Scoped("modelconfig")

	initialSiteConfig := initialConf.SiteConfig()
	// If Cody isn't enabled on startup, we don't bother registering anything.
	// Because there is no need to.
	if codyEnabled := initialSiteConfig.CodyEnabled; codyEnabled == nil || !*codyEnabled {
		logger.Info("Cody is not enabled, not registering ModelConfigService")
		return nil
	}

	// The base model configuration data is what is embedded in this binary.
	staticModelConfig, err := embedded.GetCodyGatewayModelConfig()
	if err != nil {
		return errors.New("embedded model data is missing or corrupt")
	}

	siteModelConfig, err := getSiteModelConfigurationOrNil(logger, initialSiteConfig)
	if err != nil {
		logger.Error("error loading LLM model configuration data via site config", log.Error(err))
		return errors.Wrap(err, "loading LLM configuration info")
	}

	// Now build the initial view of the Sourcegraph instance's model configuration using this data.
	configBuilder := builder{
		staticData:      staticModelConfig,
		codyGatewayData: nil,
		siteConfigData:  siteModelConfig,
	}
	initialConfig, err := configBuilder.build()
	if err != nil {
		return errors.Wrap(err, "building initial model configuration")
	}

	// Register the initial singletonConfigService.
	initialConfigSvc := service{
		currentConfig: initialConfig,
	}
	singletonConfigService = &initialConfigSvc

	// Register a watcher to pull in updates whenever the site config changes.
	conf.Watch(func() {
		logger.Info("site config updated, recalculating model configuration")

		latestSiteConfig := conf.Get().SiteConfiguration

		// TODO(chrsmith): Load the newer form of LLM model configuration data. For now, we just
		// load the older-stype completions configuration data if available.
		latestSiteModelConfiguration, err := getSiteModelConfigurationOrNil(logger, latestSiteConfig)
		if err != nil {
			// BUG: If the site configuration data is somehow bad, we silently ignore
			// the changes. This is bad, because there is no signal to the Sourcegraph
			// admin that their configuration is invalid. Hopefully we can put the necessary
			// validation logic inside the site configuration validation checks, so
			// that they will be prevented from saving invalid config data.
			logger.Error("error loading updated site configuration", log.Error(err))
			latestSiteModelConfiguration = nil
		}

		// Update and rebuild the LLM model configuration.
		//
		// NOTE: We currently don't need a mutex on `configBuilder` because there
		// is only one writer (this callback). But when we wire up the Cody Gateway
		// polling, we'll need to address that.
		configBuilder.siteConfigData = latestSiteModelConfiguration
		updatedConfig, err := configBuilder.build()
		if err != nil {
			logger.Error("error regenerating model config based on site config update", log.Error(err))
			return
		}

		singletonConfigService.set(updatedConfig)
	})

	// TODO(chrsmith): When the enhanced model configuration data is available, if configured to do so
	// we will spawn a background job that will poll Cody Gateway for any updated model information. This
	// will be tricky, because we want to honor any dynamic changes to the site config. e.g. the `conf.Watch`
	// callback needs to be able to stop/restart/update the way the poller works in response to changes.
	return nil
}
