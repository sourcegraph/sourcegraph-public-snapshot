package client

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/llmproxy"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
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
	case llmproxy.ProviderName:
		return llmproxy.NewClient(httpcli.ExternalDoer, endpoint, accessToken)
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
			if completionsConfig.Provider == llmproxy.ProviderName {
				completionsConfig.CompletionModel = "anthropic/claude-instant-v1.0"
			}
			completionsConfig.CompletionModel = "claude-instant-v1.0"
		}

		// Set a default for the llmproxy provider, so users don't have to specify it.
		if completionsConfig.Provider == llmproxy.ProviderName && completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = llmproxy.DefaultEndpoint
		}

		return completionsConfig
	}

	return nil
}
