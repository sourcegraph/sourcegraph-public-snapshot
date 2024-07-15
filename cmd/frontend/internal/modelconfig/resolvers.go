package modelconfig

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func newResolver(logger log.Logger) graphqlbackend.ModelconfigResolver {
	return &modelconfigResolver{logger: logger}
}

type modelconfigResolver struct {
	logger log.Logger
}

func (r *modelconfigResolver) CodyLLMConfiguration(ctx context.Context) (graphqlbackend.CodyLLMConfigurationResolver, error) {

	siteConfig := conf.Get().SiteConfig()

	modelCfgSvc := Get()
	modelconfig, err := modelCfgSvc.Get()
	if err != nil {
		r.logger.Warn("error obtaining model configuration data", log.Error(err))
		return nil, errors.New("error fetching model configuration data")
	}

	resolver := &codyLLMConfigurationResolver{
		modelconfig:               modelconfig,
		doNotUseCompletionsConfig: siteConfig.Completions,
	}
	return resolver, nil

}

type codyLLMConfigurationResolver struct {
	// modelconfig is the LLM model configuration data for this Sourcegraph instance.
	// This is the source of truth and accurately reflects the site configuration.
	modelconfig *modelconfigSDK.ModelConfiguration

	// doNotUseCompletionsConfig is the older-style configuration data for Cody
	// Enterprise, and is only passed along for backwards compatibility.
	//
	// DO NOT USE IT.
	//
	// The information it returns is only looking at the "completions" site config
	// data, which may not even be provided. Only read from this value if you really
	// know what you are doing.
	doNotUseCompletionsConfig *schema.Completions
}

// toLegacyModelIdentifier converts the "new style" model identity into the old style
// expected by Cody Clients.
//
// This is dangerous, as it will only work if this Sourcegraph backend AND Cody Gateway
// can correctly map the legacy identifier into the correct ModelRef.
//
// Once Cody Clients are capable of natively using the modelref format, we should remove
// this function and have all of our GraphQL APIs only refer to models using a ModelRef.
func toLegacyModelIdentifier(mref modelconfigSDK.ModelRef) string {
	return fmt.Sprintf("%s/%s", mref.ProviderID(), mref.ModelID())
}

func (c *codyLLMConfigurationResolver) ChatModel() (string, error) {
	defaultChatModelRef := c.modelconfig.DefaultModels.Chat
	model := c.modelconfig.GetModelByMRef(defaultChatModelRef)
	if model == nil {
		return "", errors.Errorf("default chat model %q not found", defaultChatModelRef)
	}
	return toLegacyModelIdentifier(model.ModelRef), nil
}

func (c *codyLLMConfigurationResolver) ChatModelMaxTokens() (*int32, error) {
	defaultChatModelRef := c.modelconfig.DefaultModels.Chat
	model := c.modelconfig.GetModelByMRef(defaultChatModelRef)
	if model == nil {
		return nil, errors.Errorf("default chat model %q not found", defaultChatModelRef)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}
func (c *codyLLMConfigurationResolver) SmartContextWindow() string {
	if c.doNotUseCompletionsConfig != nil {
		if c.doNotUseCompletionsConfig.SmartContextWindow == "disabled" {
			return "disabled"
		} else {
			return "enabled"
		}
	}

	// If the admin has explicitly provided the newer "modelConfiguration" site config
	// data, disable SmartContextWindow.
	//
	// BUG: This probably should be "enabled", but it isn't clear what this actually
	// means relative to LLM model configuration.
	return "disabled"
}
func (c *codyLLMConfigurationResolver) DisableClientConfigAPI() bool {
	if c.doNotUseCompletionsConfig != nil {
		if val := c.doNotUseCompletionsConfig.DisableClientConfigAPI; val != nil {
			return *val
		}
	}
	return false
}

func (c *codyLLMConfigurationResolver) FastChatModel() (string, error) {
	defaultFastChatModelRef := c.modelconfig.DefaultModels.FastChat
	model := c.modelconfig.GetModelByMRef(defaultFastChatModelRef)
	if model == nil {
		return "", errors.Errorf("default fast chat model %q not found", defaultFastChatModelRef)
	}
	return toLegacyModelIdentifier(model.ModelRef), nil

}

func (c *codyLLMConfigurationResolver) FastChatModelMaxTokens() (*int32, error) {
	defaultFastChatModelRef := c.modelconfig.DefaultModels.FastChat
	model := c.modelconfig.GetModelByMRef(defaultFastChatModelRef)
	if model == nil {
		return nil, errors.Errorf("default fast chat model %q not found", defaultFastChatModelRef)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}

func (c *codyLLMConfigurationResolver) Provider() string {
	if len(c.modelconfig.Providers) != 1 {
		return "various"
	}
	return c.modelconfig.Providers[0].DisplayName
}

func (c *codyLLMConfigurationResolver) CompletionModel() (string, error) {
	defaultCompletionModel := c.modelconfig.DefaultModels.CodeCompletion
	model := c.modelconfig.GetModelByMRef(defaultCompletionModel)
	if model == nil {
		return "", errors.Errorf("default code completion model %q not found", defaultCompletionModel)
	}
	return toLegacyModelIdentifier(model.ModelRef), nil
}

func (c *codyLLMConfigurationResolver) CompletionModelMaxTokens() (*int32, error) {
	defaultCompletionModel := c.modelconfig.DefaultModels.CodeCompletion
	model := c.modelconfig.GetModelByMRef(defaultCompletionModel)
	if model == nil {
		return nil, errors.Errorf("default code completion model %q not found", defaultCompletionModel)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}
