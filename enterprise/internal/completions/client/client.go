package client

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
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
	case codygateway.ProviderName:
		return codygateway.NewClient(httpcli.ExternalDoer, endpoint, accessToken)
	default:
		return nil, errors.Newf("unknown completion stream provider: %s", provider)
	}
}

// GetCompletionsConfig evaluates a complete completions configuration based on
// site configuration. The configuration may be nil if completions is disabled.
func GetCompletionsConfig(siteConfig schema.SiteConfiguration) *schema.Completions {
	codyEnabled := siteConfig.CodyEnabled
	completionsConfig := siteConfig.Completions

	// If App is running and there wasn't a completions config
	// use a provider that sends the request to dotcom
	if deploy.IsApp() {
		// If someone explicitly turned Cody off, no config
		if codyEnabled != nil && !*codyEnabled {
			return nil
		}

		// If Cody is on or not explicitly turned off and no config, assume default
		if completionsConfig == nil {
			appConfig := conf.Get().App
			if appConfig == nil {
				return nil
			}
			// Only the Provider, Access Token and Enabled required to forward the request to dotcom
			return &schema.Completions{
				AccessToken: appConfig.DotcomAuthToken,
				Provider:    dotcom.ProviderName,
				// TODO: These are not required right now as upstream overwrites this,
				// but should we switch to Cody Gateway they will be.
				ChatModel:       "claude-v1",
				FastChatModel:   "claude-instant-v1",
				CompletionModel: "claude-instant-v1",
			}
		}
	}

	// If `cody.enabled` is used but no completions config, we assume defaults
	if codyEnabled != nil && *codyEnabled {
		if completionsConfig == nil {
			completionsConfig = &schema.Completions{}
		}
		// Since `cody.enabled` is true, we override the `completions.enabled` value
		completionsConfig.Enabled = true
	}

	// Otherwise, if we don't have a config, or it's disabled, we return nil
	if completionsConfig == nil || (completionsConfig != nil && !completionsConfig.Enabled) {
		return nil
	}

	// If a provider is not set, or if the provider is Cody Gateway, set up
	// magic defaults. Note that we do NOT enable completions for the user -
	// that still needs to be explicitly configured.
	if completionsConfig.Provider == "" || completionsConfig.Provider == codygateway.ProviderName {
		// Set provider to Cody Gateway in case it's empty.
		completionsConfig.Provider = codygateway.ProviderName

		// Configure accessToken. We don't validate the license here because
		// Cody Gateway will check and reject the request.
		if completionsConfig.AccessToken == "" && siteConfig.LicenseKey != "" {
			completionsConfig.AccessToken = licensing.GenerateLicenseKeyBasedAccessToken(siteConfig.LicenseKey)
		}

		// Configure endpoint
		if completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = codygateway.DefaultEndpoint
		}
		// Configure chatModel
		if completionsConfig.ChatModel == "" {
			completionsConfig.ChatModel = "anthropic/claude-v1"
		}
		// Configure completionModel
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "anthropic/claude-instant-v1"
		}

		// NOTE: We explicitly aren't adding back-compat for completions.model
		// because Cody Gateway disallows the use of most chat models for
		// code completions, so in most cases the back-compat wouldn't work
		// anyway.

		return completionsConfig
	}

	if completionsConfig.ChatModel == "" {
		// If no model for chat is configured, nothing we can do.
		if completionsConfig.Model == "" {
			return nil
		}
		completionsConfig.ChatModel = completionsConfig.Model
	}

	// TODO: Temporary workaround to fix instances where no completion model is set.
	if completionsConfig.CompletionModel == "" {
		completionsConfig.CompletionModel = "claude-instant-v1"
	}

	return completionsConfig
}
