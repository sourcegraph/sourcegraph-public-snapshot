package client

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/llmproxy"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Get(endpoint, provider, accessToken, model string) (types.CompletionsClient, error) {
	switch provider {
	case anthropic.ProviderName:
		return anthropic.NewClient(httpcli.ExternalDoer, accessToken, model), nil
	case openai.ProviderName:
		return openai.NewClient(httpcli.ExternalDoer, accessToken, model), nil
	case dotcom.ProviderName:
		return dotcom.NewClient(httpcli.ExternalDoer, accessToken, model), nil
	case llmproxy.ProviderName:
		return llmproxy.NewClient(httpcli.ExternalDoer, endpoint, accessToken, model)
	default:
		return nil, errors.Newf("unknown completion stream provider: %s", provider)
	}
}

func GetCompletionsConfig() *schema.Completions {
	completionsConfig := conf.Get().Completions

	// When the Completions is present always use it
	if completionsConfig != nil {
		if completionsConfig.ChatModel == "" {
			completionsConfig.ChatModel = completionsConfig.Model
		}

		// TODO: Temporary workaround to fix instances where no completion model is set.
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "claude-instant-v1.0"
		}

		if completionsConfig.Provider == llmproxy.ProviderName && completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = llmproxy.DefaultEndpoint
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
		}
	}
	return nil
}
