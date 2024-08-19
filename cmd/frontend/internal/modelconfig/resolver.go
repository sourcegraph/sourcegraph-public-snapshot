package modelconfig

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type codyLLMConfigResolver struct {
	logger log.Logger
}

func newResolver(logger log.Logger) graphqlbackend.ModelconfigResolver {
	return &codyLLMConfigResolver{logger: logger}
}

var _ = (*graphqlbackend.CodyLLMConfigurationResolver)(nil)

// CodyLLMConfiguration returns the GraphQL resolver which returns the CodyLLMConfiguration data,
// the behavior of which depends on exactly how the current site is configured.
func (llmResolver *codyLLMConfigResolver) CodyLLMConfiguration(ctx context.Context) (graphqlbackend.CodyLLMConfigurationResolver, error) {
	siteConfig := conf.Get().SiteConfig()

	// If the site is configured to only use the older "completions" configuration, then we return a
	// resolver that matches the previous behavior. (Which just returned the provider and models as-is,
	// which is what older Cody clients expect.)
	if siteConfig.Completions != nil && siteConfig.ModelConfiguration == nil {
		completionsCfg := conf.GetCompletionsConfig(siteConfig)
		// e.g. if Cody is not enabled.
		if completionsCfg == nil {
			return nil, nil
		}
		return &completionsConfigResolver{
			config: completionsCfg,
		}, nil
	}

	// Otherwise, use a different GraphQL resolver which needs to go to great lengths to preserve
	// the previous behavior.
	modelCfgSvc := Get()
	modelconfig, err := modelCfgSvc.Get()
	if err != nil {
		llmResolver.logger.Warn("error obtaining model configuration data", log.Error(err))
		return nil, errors.New("error fetching model configuration data")
	}

	// Create a new instance of the codyLLMConfigurationResolver per-request, so that
	// we always pick up the latest site config, rather than using a stale version from
	// when the Sourcegraph instance was initialized.
	resolver := &modelconfigResolver{
		modelconfig: modelconfig,
	}
	return resolver, nil

}

// completionsConfigResolver is the original logic of the CodyLLMConfigurationResolver before the
// modelconfig changes landed. This should be used when the site configuration is only using the
// "completions" section to configure LLM models.
type completionsConfigResolver struct {
	config *conftypes.CompletionsConfig
}

func (c *completionsConfigResolver) ChatModel() (string, error) {
	return convertLegacyModelNameToModelID(c.config.ChatModel), nil
}

func (c *completionsConfigResolver) ChatModelMaxTokens() (*int32, error) {
	if c.config.ChatModelMaxTokens != 0 {
		max := int32(c.config.ChatModelMaxTokens)
		return &max, nil
	}
	return nil, nil
}

func (c *completionsConfigResolver) SmartContextWindow() string {
	if c.config.SmartContextWindow == "disabled" {
		return "disabled"
	}
	return "enabled"
}

func (c *completionsConfigResolver) DisableClientConfigAPI() bool {
	return c.config.DisableClientConfigAPI
}

func (c *completionsConfigResolver) FastChatModel() (string, error) {
	return convertLegacyModelNameToModelID(c.config.FastChatModel), nil
}

func (c *completionsConfigResolver) FastChatModelMaxTokens() (*int32, error) {
	if c.config.FastChatModelMaxTokens != 0 {
		max := int32(c.config.FastChatModelMaxTokens)
		return &max, nil
	}
	return nil, nil
}

func (c *completionsConfigResolver) Provider() string {
	return string(c.config.Provider)
}

func (c *completionsConfigResolver) CompletionModel() (string, error) {
	return convertLegacyModelNameToModelID(c.config.CompletionModel), nil
}

func (c *completionsConfigResolver) CompletionModelMaxTokens() (*int32, error) {
	if c.config.CompletionModelMaxTokens != 0 {
		max := int32(c.config.CompletionModelMaxTokens)
		return &max, nil
	}
	return nil, nil
}

// modelconfigResolver implements the CodyLLMConfigurationResolver, used when the
// site is configured to use the newer-style LLM model configuration format.
type modelconfigResolver struct {
	// modelconfig is the LLM model configuration data for this Sourcegraph instance.
	// This is the source of truth and accurately reflects the site configuration.
	modelconfig *modelconfigSDK.ModelConfiguration
}

func (r *modelconfigResolver) ChatModel() (string, error) {
	defaultChatModelRef := r.modelconfig.DefaultModels.Chat
	model := r.modelconfig.GetModelByMRef(defaultChatModelRef)
	if model == nil {
		return "", errors.Errorf("default chat model %q not found", defaultChatModelRef)
	}
	return r.toLegacyModelRef(*model), nil
}

func (r *modelconfigResolver) ChatModelMaxTokens() (*int32, error) {
	defaultChatModelRef := r.modelconfig.DefaultModels.Chat
	model := r.modelconfig.GetModelByMRef(defaultChatModelRef)
	if model == nil {
		return nil, errors.Errorf("default chat model %q not found", defaultChatModelRef)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}

func (r *modelconfigResolver) SmartContextWindow() string {
	return "disabled"
}

func (r *modelconfigResolver) DisableClientConfigAPI() bool {
	// There is no way to disable the client config API if the site
	// is using the "modelConfiguration" format.
	return false
}

func (r *modelconfigResolver) FastChatModel() (string, error) {
	defaultFastChatModelRef := r.modelconfig.DefaultModels.FastChat
	model := r.modelconfig.GetModelByMRef(defaultFastChatModelRef)
	if model == nil {
		return "", errors.Errorf("default fast chat model %q not found", defaultFastChatModelRef)
	}
	return r.toLegacyModelRef(*model), nil

}

func (r *modelconfigResolver) FastChatModelMaxTokens() (*int32, error) {
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
func (r *modelconfigResolver) Provider() string {
	completionProviderID := r.modelconfig.DefaultModels.CodeCompletion.ProviderID()
	return r.convertProviderID(completionProviderID)
}

func (r *modelconfigResolver) CompletionModel() (string, error) {
	defaultCompletionModel := r.modelconfig.DefaultModels.CodeCompletion
	model := r.modelconfig.GetModelByMRef(defaultCompletionModel)
	if model == nil {
		return "", errors.Errorf("default code completion model %q not found", defaultCompletionModel)
	}
	return r.toLegacyModelRef(*model), nil
}

func (r *modelconfigResolver) CompletionModelMaxTokens() (*int32, error) {
	defaultCompletionModel := r.modelconfig.DefaultModels.CodeCompletion
	model := r.modelconfig.GetModelByMRef(defaultCompletionModel)
	if model == nil {
		return nil, errors.Errorf("default code completion model %q not found", defaultCompletionModel)
	}
	maxTokens := int32(model.ContextWindow.MaxInputTokens)
	return &maxTokens, nil
}

// toLegacyModelRef converts the newer-style model (e.g. mref anthropic::v1::claude-3-sonnet) to
// the older style (anthropic/claude-3-sonnet-20240229). However, this function will rename
// the provider as needed to match older behavior. (See unit tests and convertProviderID for
// more information.)
func (r *modelconfigResolver) toLegacyModelRef(model modelconfigSDK.Model) string {
	modelID := model.ModelRef.ModelID()
	providerID := model.ModelRef.ProviderID()
	legacyProviderName := r.convertProviderID(providerID)

	// Potential issue: Older Cody clients calling the GraphQL may expect to see the model **name**
	// such as "claude-3-sonnet-20240229". But it is important that we only return the model **ID**
	// because that is what the HTTP completions API is expecting to see from the client.
	//
	// So when using older Cody clients, unaware of the newer modelconfig system, this could lead
	// to some errors. (But newer clients won't be using this GraphQL endpoint at all and instead
	// just use the newer modelconfig system, so hopefully this won't be a major concern.)
	return fmt.Sprintf("%s/%s", legacyProviderName, modelID)
}

// convertProviderID returns the _API Provider_ for the referenced modelconfig provider.
// The provider ID may be admin-defined, or typically just the _model_ provider like "anthropic".
// But the Cody client needs to know the _API provider_, so it can build API requests in the
// right format.
//
// Now, client's _shouldn't_ do this, because it doesn't account for LLM providers shipping breaking
// changes with their models. But we still want to keep older Cody clients working.
func (r *modelconfigResolver) convertProviderID(id modelconfigSDK.ProviderID) string {
	// Lookup the provider that was referenced.
	var referencedProvider *modelconfigSDK.Provider
	for _, provider := range r.modelconfig.Providers {
		if provider.ID == id {
			referencedProvider = &provider
			break
		}
	}

	if referencedProvider == nil || referencedProvider.ServerSideConfig == nil {
		return "sourcegraph"
	}

	// Inspect the server-side configuration to see what API provider is being used
	// under the hood.
	ssConfig := referencedProvider.ServerSideConfig
	if ssConfig.AWSBedrock != nil {
		return "aws-bedrock"
	}
	if ssConfig.AzureOpenAI != nil {
		return "azure-openai"
	}
	if ssConfig.GenericProvider != nil {
		// e.g. "anthropic", "fireworks", "google", etc.
		return string(ssConfig.GenericProvider.ServiceName)
	}
	if ssConfig.OpenAICompatible != nil {
		return "openaicompatible"
	}
	if ssConfig.SourcegraphProvider != nil {
		return "sourcegraph"
	}

	return "unknown"
}
