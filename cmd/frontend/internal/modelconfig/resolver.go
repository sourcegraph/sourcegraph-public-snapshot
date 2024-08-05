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

// Here Be Dragons (written July 30 2024)
//
// Cody clients currently rely on CodyLLMConfiguration, they shouldn't, they should
// use the information provided by the ModelsService and the /.api/client-config -
// both of which supersede this information. However, they use it today.
//
// Clients currently rely on this CodyLLMConfiguration.provider field ONLY to
// determine which **autocomplete provider implementation** to use. That is, to
// control context limits, prompting behavior, and other aspects of autocomplete.
// This is not a great way to handle this (provider/model-specific behavior being
// fundamentally tied together), but again, it's how it works today.
//
// The 'autocomplete provider name string' is determined by the following logic:
// 1) If the server has the new `modelConfiguration` in their site config, then the
// client will use the autocomplete `Model.provider` (provider ID, not name) field
// as the string.
// 2) Else `if (authStatus.configOverwrites?.provider)` -- i.e. the string returned by this
// function -- will be used.
// 3) Else, the default string 'anthropic' will be used.
//
// The 'autocomplete provider name string' is then entered into a switch statement
// (see create-provider.ts, `createProviderConfig` - which can be summarized as:
//
// 'openai', 'azure-openai'   - createUnstableOpenAIProviderConfig
// 'fireworks'                - createFireworksProviderConfig
// 'aws-bedrock', 'anthropic' - createAnthropicProviderConfig
// 'google'                   - createAnthropicProviderConfig or createGeminiProviderConfig dep. on model
//
// Note that all other cases are irrelevant:
//
// 'experimental-openaicompatible' - deprecated; client-side only option; does not need to be returned by this function.
// 'openaicompatible'              - does not need to be returned by this function (uses new Models service instead of CodyLLMConfiguration.provider)
// Ollama and other options        - are client-side only
//
// Finally, it is worth noting that Sourcegraph instance versions prior to Aug 7th 2024
// using Cody Gateway would return 'sourcegraph' as the provider name here. This is expected
// on the client side in some places.
//
// Lastly, remember that we only ever use the default autocomplete model. There is no UI
// currently in the product to choose the autocomplete model as a Cody user. However,
// when we choose to add that feature, note that this return value is NOT used when
// `modelConfiguration` is present in the site config: `Model.provider`.
//
// TL;DR: this function is currently expected to return strings that satisfy these conditions:
//
// 'openai', 'azure-openai'   - createUnstableOpenAIProviderConfig
// 'fireworks'                - createFireworksProviderConfig
// 'aws-bedrock', 'anthropic' - createAnthropicProviderConfig
// 'google'                   - createAnthropicProviderConfig or createGeminiProviderConfig depending on model
func (r *codyLLMConfigurationResolver) Provider() string {
	// Special case. If the older completions is used to configure this Sourcegraph instance,
	// just use that. This matches the behavior prior to modelconfig being used, and will support
	// the case where "sourcegraph" is the configured provider.
	if r.doNotUseCompletionsConfig != nil {
		return r.doNotUseCompletionsConfig.Provider
	}

	// Otherwise, just return the provider ID of the code completion model.
	// NOTE: In situations where the site admin has configured their LLM models, the Provider ID
	// could be something unknown to the client. So there are situations where this value cannot
	// be used correctly on the client, without additional information from the server. (Such as
	// knowing the specific API service the provider is using, e.g. OpenAI, AWS Bedrock, etc.)
	return string(r.modelconfig.DefaultModels.CodeCompletion.ProviderID())
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
