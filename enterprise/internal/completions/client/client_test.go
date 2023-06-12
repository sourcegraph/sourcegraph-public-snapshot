package client

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetCompletionsConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config schema.SiteConfiguration
		want   autogold.Value
	}{
		{
			name: "completions disabled",
			config: schema.SiteConfiguration{
				Completions: nil,
			},
			want: autogold.Expect((*schema.Completions)(nil)),
		},
		{
			name: "anthropic completions",
			config: schema.SiteConfiguration{
				Completions: &schema.Completions{
					Enabled:         true,
					Provider:        "anthropic",
					ChatModel:       "claude-v1",
					FastChatModel:   "claude-instant-v1",
					CompletionModel: "claude-instant-v1",
				},
			},
			want: autogold.Expect(&schema.Completions{
				ChatModel: "claude-v1", CompletionModel: "claude-instant-v1",
				FastChatModel: "claude-instant-v1",
				Enabled:       true,
				Provider:      "anthropic",
			}),
		},
		{
			name: "zero-config cody gateway completions without license key",
			config: schema.SiteConfiguration{
				Completions: &schema.Completions{
					Enabled: true,
				},
				LicenseKey: "",
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
				Completions: &schema.Completions{
					Enabled: true,
				},
				LicenseKey: "foobar",
			},
			want: autogold.Expect(&schema.Completions{
				AccessToken:     "slk_3f2c7ccae98af81e44c0ec419659f50d8b7d48c681e5d57fc747d0461e42dda1",
				ChatModel:       "anthropic/claude-v1",
				CompletionModel: "anthropic/claude-instant-v1",
				Enabled:         true,
				Endpoint:        "https://cody-gateway.sourcegraph.com",
				Provider:        "sourcegraph",
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := GetCompletionsConfig(tc.config)
			tc.want.Equal(t, got)
		})
	}
}
