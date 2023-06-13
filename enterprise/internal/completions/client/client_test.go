package client

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetCompletionsConfig(t *testing.T) {
	truePtr := true

	for _, tc := range []struct {
		name       string
		deployType string
		config     schema.SiteConfiguration
		want       autogold.Value
	}{
		{
			name: "cody not enabled",
			config: schema.SiteConfiguration{
				CodyEnabled: nil,
			},
			want: autogold.Expect((*schema.Completions)(nil)),
		},
		{
			name: "anthropic completions",
			config: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				Completions: &schema.Completions{
					Enabled:         true,
					Provider:        "anthropic",
					ChatModel:       "claude-v1",
					FastChatModel:   "claude-instant-v1",
					CompletionModel: "claude-instant-v1",
				},
			},
			want: autogold.Expect(&schema.Completions{
				Enabled:         true,
				ChatModel:       "claude-v1",
				FastChatModel:   "claude-instant-v1",
				CompletionModel: "claude-instant-v1",
				Provider:        "anthropic",
			}),
		},
		{
			name: "anthropic completions, with cody.enabled taking precedence over completions.enabled",
			config: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				Completions: &schema.Completions{
					Enabled:         false,
					Provider:        "anthropic",
					ChatModel:       "claude-v1",
					CompletionModel: "claude-instant-v1",
				},
			},
			want: autogold.Expect(&schema.Completions{
				ChatModel: "claude-v1", CompletionModel: "claude-instant-v1",
				Enabled:  true,
				Provider: "anthropic",
			}),
		},
		{
			name: "zero-config cody gateway completions without license key",
			config: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				LicenseKey:  "",
			},
			want: autogold.Expect(&schema.Completions{
				ChatModel:       "anthropic/claude-v1",
				CompletionModel: "anthropic/claude-instant-v1",
				Enabled:         true,
				Endpoint:        "https://cody-gateway.sourcegraph.com",
				Provider:        "sourcegraph",
			}),
		},
		{
			name: "zero-config cody gateway completions with license key",
			config: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				LicenseKey:  "foobar",
			},
			want: autogold.Expect(&schema.Completions{
				Enabled:         true,
				AccessToken:     "slk_c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2",
				ChatModel:       "anthropic/claude-v1",
				CompletionModel: "anthropic/claude-instant-v1",
				Endpoint:        "https://cody-gateway.sourcegraph.com",
				Provider:        "sourcegraph",
			}),
		},
		{
			// Legacy support for completions.enabled
			name: "legacy field completions.enabled: zero-config cody gateway completions without license key",
			config: schema.SiteConfiguration{
				Completions: &schema.Completions{Enabled: true},
				LicenseKey:  "",
			},
			want: autogold.Expect(&schema.Completions{
				Enabled:         true,
				ChatModel:       "anthropic/claude-v1",
				CompletionModel: "anthropic/claude-instant-v1",
				Endpoint:        "https://cody-gateway.sourcegraph.com",
				Provider:        "sourcegraph",
			}),
		},
		{
			// Legacy support for completions.enabled
			name: "legacy field completions.enabled: zero-config cody gateway completions with license key",
			config: schema.SiteConfiguration{
				Completions: &schema.Completions{Enabled: true},
				LicenseKey:  "foobar",
			},
			want: autogold.Expect(&schema.Completions{
				Enabled:         true,
				AccessToken:     "slk_c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2",
				ChatModel:       "anthropic/claude-v1",
				CompletionModel: "anthropic/claude-instant-v1",
				Endpoint:        "https://cody-gateway.sourcegraph.com",
				Provider:        "sourcegraph",
			}),
		},
		{
			name:       "app zero-config cody gateway completions with dotcom token",
			deployType: deploy.App,
			config: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				App: &schema.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			want: autogold.Expect(&schema.Completions{
				AccessToken:     "sgd_5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456",
				ChatModel:       "anthropic/claude-v1",
				CompletionModel: "anthropic/claude-instant-v1",
				Enabled:         true,
				Endpoint:        "https://cody-gateway.sourcegraph.com",
				Provider:        "sourcegraph",
			}),
		},
		{
			name:       "app with custom configuration",
			deployType: deploy.App,
			config: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				Completions: &schema.Completions{
					AccessToken:     "CUSTOM_TOKEN",
					Provider:        "anthropic",
					ChatModel:       "claude-v1",
					FastChatModel:   "claude-instant-v1",
					CompletionModel: "claude-instant-v1",
				},
				App: &schema.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			want: autogold.Expect(&schema.Completions{
				AccessToken: "CUSTOM_TOKEN", ChatModel: "claude-v1",
				CompletionModel: "claude-instant-v1",
				Enabled:         true,
				FastChatModel:   "claude-instant-v1",
				Provider:        "anthropic",
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defaultDeploy := deploy.Type()
			if tc.deployType != "" {
				deploy.Mock(tc.deployType)
			}
			defer deploy.Mock(defaultDeploy)
			got := GetCompletionsConfig(tc.config)
			tc.want.Equal(t, got)
		})
	}
}
