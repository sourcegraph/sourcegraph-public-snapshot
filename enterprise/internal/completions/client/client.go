package client

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Get(endpoint, provider, accessToken string) (types.CompletionsClient, error) {
	switch provider {
	case anthropic.ProviderName:
		return anthropic.NewClient(httpcli.ExternalDoer, endpoint, accessToken), nil
	case openai.ProviderName:
		return openai.NewClient(httpcli.ExternalDoer, endpoint, accessToken), nil
	case dotcom.ProviderName:
		return dotcom.NewClient(httpcli.ExternalDoer, accessToken), nil
	case codygateway.ProviderName, "llmproxy": // temporary back-compat
		return codygateway.NewClient(httpcli.ExternalDoer, endpoint, accessToken)
	default:
		return nil, errors.Newf("unknown completion stream provider: %s", provider)
	}
}

func GetCompletionsConfig() *schema.Completions {
	completionsConfig := conf.Get().Completions

	// When the Completions is present always use it
	if completionsConfig != nil {
		if completionsConfig.ChatModel == "" {
			// If no model for chat is configured, nothing we can do.
			if completionsConfig.Model == "" {
				return nil
			}
			completionsConfig.ChatModel = completionsConfig.Model
		}

		// TODO: Temporary workaround to fix instances where no completion model is set.
		if completionsConfig.CompletionModel == "" {
			if completionsConfig.Provider == codygateway.ProviderName {
				completionsConfig.CompletionModel = "anthropic/claude-instant-v1"
			}
			completionsConfig.CompletionModel = "claude-instant-v1"
		}

		// Set a default for the Cody Gateway provider, so users don't have to specify it.
		if completionsConfig.Provider == codygateway.ProviderName && completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = codygateway.DefaultEndpoint
		}

		return completionsConfig
	}

	// If App is running and there wasn't a completions config
	// use a provider that sends the request to dotcom
	if deploy.IsApp() {
		appConfig := conf.Get().App
		if appConfig == nil {
			return nil
		}
		// Only the Provider, Access Token and Enabled required to forward the request to dotcom
		return &schema.Completions{
			AccessToken: appConfig.DotcomAuthToken,
			Enabled:     len(appConfig.DotcomAuthToken) > 0,
			Provider:    dotcom.ProviderName,
			// TODO: These are not required right now as upstream overwrites this,
			// but should we switch to Cody Gateway they will be.
			ChatModel:       "claude-v1",
			CompletionModel: "claude-instant-v1",
		}
	}
	return nil
}
