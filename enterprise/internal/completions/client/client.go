package client

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/dotcomuser"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
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

	// When the Completions is present always use it
	if completionsConfig != nil {
		// If a provider is not set, or if the provider is Cody Gateway, set up
		// magic defaults. Note that we do NOT enable completions for the user -
		// that still needs to be explicitly configured.
		if completionsConfig.Provider == "" || completionsConfig.Provider == codygateway.ProviderName {
			// Set provider to Cody Gateway in case it's empty.
			completionsConfig.Provider = codygateway.ProviderName

			// Configure accessToken. We don't validate the license here because
			// Cody Gateway will check and reject the request.
			if completionsConfig.AccessToken == "" {
				switch deploy.Type() {
				case deploy.App:
					if siteConfig.App != nil && siteConfig.App.DotcomAuthToken != "" {
						completionsConfig.AccessToken = dotcomuser.GenerateDotcomUserGatewayAccessToken(siteConfig.App.DotcomAuthToken)
					}
				default:
					if siteConfig.LicenseKey != "" {
						completionsConfig.AccessToken = licensing.GenerateLicenseKeyBasedAccessToken(siteConfig.LicenseKey)
					}
				}

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

	return nil
}
