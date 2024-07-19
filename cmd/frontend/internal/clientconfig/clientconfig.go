package clientconfig

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/clientconfig"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func GetForActor(ctx context.Context, logger log.Logger, db database.DB, actor *actor.Actor) (*clientconfig.ClientConfig, error) {
	c := clientconfig.ClientConfig{
		// If the site config has "modelConfiguration" specified / non-null, then the site admin
		// has opted into the new model configuration system, wants to use the new /.api/supported-llms
		// endpoint for models, etc.
		ModelsAPIEnabled: conf.UseExperimentalModelConfiguration(),
	}

	// 🚨 SECURITY: This code lets site admins restrict who has access to Cody at all via RBAC.
	// https://sourcegraph.com/docs/cody/clients/enable-cody-enterprise#enable-cody-only-for-some-users
	c.CodyEnabled, _ = cody.IsCodyEnabled(ctx, db)

	// 🚨 SECURITY: This code enforces that users do not have access to Cody features which
	// site admins do not want them to have access to.
	//
	// Legacy admin-control configuration which should be moved to RBAC, not globally in site
	// config. e.g. we should do it like https://github.com/sourcegraph/sourcegraph/pull/58831
	features := conf.GetConfigFeatures(conf.Get().SiteConfig())
	if features != nil { // nil -> Cody not enabled
		c.ChatEnabled = features.Chat
		c.AutoCompleteEnabled = features.AutoComplete
		c.CustomCommandsEnabled = features.Commands
		c.AttributionEnabled = features.Attribution
	}

	// Legacy feature-enablement configuration which should be moved to featureflag or RBAC,
	// not exist in site config.
	completionConfig := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	if completionConfig != nil { // nil -> Cody not enabled
		c.SmartContextWindowEnabled = completionConfig.SmartContextWindow != "disabled"
	}

	return &c, nil
}
