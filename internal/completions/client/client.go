package client

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/awsbedrock"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/azureopenai"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/fireworks"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/google"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/openaicompatible"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

func Get(
	logger log.Logger,
	events *telemetry.EventRecorder,
	modelConfigInfo types.ModelConfigInfo) (types.CompletionsClient, error) {
	client, err := getAPIProvider(modelConfigInfo)
	if err != nil {
		return nil, err
	}
	return newObservedClient(logger, events, client), nil
}

// getAPIProvider returns the CompletionsClient needed to serve the given LLM request. There is
// a super-important detail here! Typically, "Provider" is referring to the "Model Provider". Or
// a 3rd party that built various models. The "API Provider" on the other hand is the actual API
// or service we use to make LLM requests. And the two may be very different.
//
// For example, the Model Provider may be "openai" but the API Provider could be Cody Gateway,
// AzureOpenAI, or even AWS Bedrock.
//
// So the `modelConfigInfo.Provider` isn't strictly necessary. We don't care _who built_ the model,
// instead we care about the API we need to call in order to serve the request. So we look at the
// `modelConfigInfo.Provider.ServerSideConfig` to determine which CompletionsClient should be used
// to make the actual LLM request.
func getAPIProvider(modelConfigInfo types.ModelConfigInfo) (types.CompletionsClient, error) {
	tokenManager := tokenusage.NewManager()

	ssConfig := modelConfigInfo.Provider.ServerSideConfig
	if ssConfig == nil {
		return nil, errors.Errorf("no server-side config available for provider %q", modelConfigInfo.Provider.ID)
	}

	// AWS Bedrock.
	if awsBedrockCfg := ssConfig.AWSBedrock; awsBedrockCfg != nil {
		client := awsbedrock.NewClient(
			httpcli.UncachedExternalDoer, awsBedrockCfg.Endpoint, awsBedrockCfg.AccessToken, *tokenManager)
		return client, nil
	}

	// Azure OpenAI
	if azureOpenAICfg := ssConfig.AzureOpenAI; azureOpenAICfg != nil {
		client, err := azureopenai.NewClient(
			azureopenai.GetAPIClient, azureOpenAICfg.Endpoint, azureOpenAICfg.AccessToken, *tokenManager)
		return client, errors.Wrap(err, "getting api provider")
	}

	// OpenAI Compatible
	if openAICompatibleCfg := ssConfig.OpenAICompatible; openAICompatibleCfg != nil {
		return openaicompatible.NewClient(httpcli.UncachedExternalClient, *tokenManager), nil
	}

	// The "GenericProvider" is an escape hatch for a set of API Providers not needing any additional configuration.
	if genProviderCfg := ssConfig.GenericProvider; genProviderCfg != nil {
		token := genProviderCfg.AccessToken
		endpoint := genProviderCfg.Endpoint

		switch genProviderCfg.ServiceName {
		case modelconfigSDK.GenericServiceProviderAnthropic:
			client := anthropic.NewClient(httpcli.UncachedExternalDoer, endpoint, token, false, *tokenManager)
			return client, nil
		case modelconfigSDK.GenericServiceProviderFireworks:
			client := fireworks.NewClient(httpcli.UncachedExternalDoer, endpoint, token)
			return client, nil
		case modelconfigSDK.GenericServiceProviderGoogle:
			// Don't resolve Google LLM requests via Cody Gateway, contact Google directly.
			viaGateway := false
			return google.NewClient(httpcli.UncachedExternalDoer, endpoint, token, viaGateway)
		case modelconfigSDK.GenericServiceProviderOpenAI:
			client := openai.NewClient(httpcli.UncachedExternalDoer, endpoint, token, *tokenManager)
			return client, nil
		default:
			return nil, errors.Errorf("unknown GeneralProvider %q", genProviderCfg.ServiceName)
		}
	}

	// The "Sourcegraph" provider, AKA Cody Gateway.
	if sgProviderCfg := ssConfig.SourcegraphProvider; sgProviderCfg != nil {
		client, err := codygateway.NewClient(
			httpcli.CodyGatewayDoer, sgProviderCfg.Endpoint, sgProviderCfg.AccessToken, *tokenManager)
		return client, err
	}

	return nil, errors.Newf("no LLM API provider available for %q", modelConfigInfo.Provider.ID)
}
