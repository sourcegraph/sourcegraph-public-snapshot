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

type modelconfigResolver struct {
	logger log.Logger
}

func newResolver(logger log.Logger) graphqlbackend.ModelconfigResolver {
	return &modelconfigResolver{logger: logger}
}

var _ = (*graphqlbackend.CodyLLMConfigurationResolver)(nil)

func (r *modelconfigResolver) CodyLLMConfiguration(ctx context.Context) (graphqlbackend.CodyLLMConfigurationResolver, error) {

	siteConfig := conf.Get().SiteConfig()

	modelCfgSvc := Get()
	modelconfig, err := modelCfgSvc.Get()
	if err != nil {
		r.logger.Warn("error obtaining model configuration data", log.Error(err))
		return nil, errors.New("error fetching model configuration data")
	}

	// Create a new instance of the codyLLMConfigurationResolver per-request, so that
	// we always pick up the latest site config, rather than using a stale version from
	// when the Sourcegraph instance was initialized.
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

func (r *codyLLMConfigurationResolver) ChatModel() (string, error) {
	defaultChatModelRef := r.modelconfig.DefaultModels.Chat
	model := r.modelconfig.GetModelByMRef(defaultChatModelRef)
	if model == nil {
		return "", errors.Errorf("default chat model %q not found", defaultChatModelRef)
	}
	return toLegacyModelIdentifier(model.ModelRef), nil
}

func (r *codyLLMConfigurationResolver) ChatModelMaxTokens() (*int32, error) {
	defaultChatModelRef := r.modelconfig.DefaultModels.Chat
	model := r.modelconfig.GetModelByMRef(defaultChatModelRef)
	if model == nil {
		return nil, errors.Errorf("default chat model %q not found", defaultChatModelRef)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}
func (r *codyLLMConfigurationResolver) SmartContextWindow() string {
	if r.doNotUseCompletionsConfig != nil {
		if r.doNotUseCompletionsConfig.SmartContextWindow == "disabled" {
			return "disabled"
		} else {
			return "enabled"
		}
	}

	// If the admin has explicitly provided the newer "modelConfiguration" site config
	// data, disable SmartContextWindow. We may want to re-enable this capability, but
	// in some other way. (e.g. passing this flag on a per-model basis, or just having
	// a more nuanced view of a model's specific context window.)
	return "disabled"
}
func (r *codyLLMConfigurationResolver) DisableClientConfigAPI() bool {
	if r.doNotUseCompletionsConfig != nil {
		if val := r.doNotUseCompletionsConfig.DisableClientConfigAPI; val != nil {
			return *val
		}
	}
	return false
}

func (r *codyLLMConfigurationResolver) FastChatModel() (string, error) {
	defaultFastChatModelRef := r.modelconfig.DefaultModels.FastChat
	model := r.modelconfig.GetModelByMRef(defaultFastChatModelRef)
	if model == nil {
		return "", errors.Errorf("default fast chat model %q not found", defaultFastChatModelRef)
	}
	return toLegacyModelIdentifier(model.ModelRef), nil

}

func (r *codyLLMConfigurationResolver) FastChatModelMaxTokens() (*int32, error) {
	defaultFastChatModelRef := r.modelconfig.DefaultModels.FastChat
	model := r.modelconfig.GetModelByMRef(defaultFastChatModelRef)
	if model == nil {
		return nil, errors.Errorf("default fast chat model %q not found", defaultFastChatModelRef)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}

func (r *codyLLMConfigurationResolver) Provider() string {
	if len(r.modelconfig.Providers) != 1 {
		return "various"
	}
	return r.modelconfig.Providers[0].DisplayName
}

func (r *codyLLMConfigurationResolver) CompletionModel() (string, error) {
	defaultCompletionModel := r.modelconfig.DefaultModels.CodeCompletion
	model := r.modelconfig.GetModelByMRef(defaultCompletionModel)
	if model == nil {
		return "", errors.Errorf("default code completion model %q not found", defaultCompletionModel)
	}
	return toLegacyModelIdentifier(model.ModelRef), nil
}

func (r *codyLLMConfigurationResolver) CompletionModelMaxTokens() (*int32, error) {
	defaultCompletionModel := r.modelconfig.DefaultModels.CodeCompletion
	model := r.modelconfig.GetModelByMRef(defaultCompletionModel)
	if model == nil {
		return nil, errors.Errorf("default code completion model %q not found", defaultCompletionModel)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}
