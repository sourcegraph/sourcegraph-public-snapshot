package conf

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthPasswordResetLinkDuration(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want int
	}{{
		name: "password link expiry has a default value if null",
		sc:   &Unified{},
		want: defaultPasswordLinkExpiry,
	}, {
		name: "password link expiry has a default value if blank",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{AuthPasswordResetLinkExpiry: 0}},
		want: defaultPasswordLinkExpiry,
	}, {
		name: "password link expiry can be customized",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{AuthPasswordResetLinkExpiry: 60}},
		want: 60,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := AuthPasswordResetLinkExpiry(), test.want; got != want {
				t.Fatalf("AuthPasswordResetLinkExpiry() = %v, want %v", got, want)
			}
		})
	}
}

func TestGitLongCommandTimeout(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want time.Duration
	}{{
		name: "Git long command timeout has a default value if null",
		sc:   &Unified{},
		want: defaultGitLongCommandTimeout,
	}, {
		name: "Git long command timeout has a default value if blank",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitLongCommandTimeout: 0}},
		want: defaultGitLongCommandTimeout,
	}, {
		name: "Git long command timeout can be customized",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitLongCommandTimeout: 60}},
		want: time.Duration(60) * time.Second,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := GitLongCommandTimeout(), test.want; got != want {
				t.Fatalf("GitLongCommandTimeout() = %v, want %v", got, want)
			}
		})
	}
}

func TestGitMaxCodehostRequestsPerSecond(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want int
	}{
		{
			name: "not set should return default",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want: -1,
		},
		{
			name: "bad value should return default",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitMaxCodehostRequestsPerSecond: pointers.Ptr(-100)}},
			want: -1,
		},
		{
			name: "set 0 should return 0",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitMaxCodehostRequestsPerSecond: pointers.Ptr(0)}},
			want: 0,
		},
		{
			name: "set non-0 should return non-0",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitMaxCodehostRequestsPerSecond: pointers.Ptr(100)}},
			want: 100,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := GitMaxCodehostRequestsPerSecond(), test.want; got != want {
				t.Fatalf("GitMaxCodehostRequestsPerSecond() = %v, want %v", got, want)
			}
		})
	}
}

func TestGitMaxConcurrentClones(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want int
	}{
		{
			name: "not set should return default",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want: 5,
		},
		{
			name: "bad value should return default",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					GitMaxConcurrentClones: -100,
				},
			},
			want: 5,
		},
		{
			name: "set non-zero should return non-zero",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					GitMaxConcurrentClones: 100,
				},
			},
			want: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := GitMaxConcurrentClones(), test.want; got != want {
				t.Fatalf("GitMaxConcurrentClones() = %v, want %v", got, want)
			}
		})
	}
}

func TestAuthLockout(t *testing.T) {
	defer Mock(nil)

	tests := []struct {
		name string
		mock *schema.AuthLockout
		want *schema.AuthLockout
	}{
		{
			name: "missing entire config",
			mock: nil,
			want: &schema.AuthLockout{
				ConsecutivePeriod:      3600,
				FailedAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
		{
			name: "missing all fields",
			mock: &schema.AuthLockout{},
			want: &schema.AuthLockout{
				ConsecutivePeriod:      3600,
				FailedAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
		{
			name: "missing some fields",
			mock: &schema.AuthLockout{
				ConsecutivePeriod: 7200,
			},
			want: &schema.AuthLockout{
				ConsecutivePeriod:      7200,
				FailedAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(&Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthLockout: test.mock,
				},
			})

			got := AuthLockout()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestIsAccessRequestEnabled(t *testing.T) {
	falseVal, trueVal := false, true
	tests := []struct {
		name string
		sc   *Unified
		want bool
	}{
		{
			name: "not set should return default true",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want: true,
		},
		{
			name: "parent object set should return default true",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthAccessRequest: &schema.AuthAccessRequest{},
				},
			},
			want: true,
		},
		{
			name: "explicitly set enabled=true should return true",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthAccessRequest: &schema.AuthAccessRequest{Enabled: &trueVal},
				},
			},
			want: true,
		},
		{
			name: "explicitly set enabled=false should return false",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthAccessRequest: &schema.AuthAccessRequest{
						Enabled: &falseVal,
					},
				},
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			have := IsAccessRequestEnabled()
			assert.Equal(t, test.want, have)
		})
	}
}

func TestCodyEnabled(t *testing.T) {
	tests := []struct {
		name string
		sc   schema.SiteConfiguration
		want bool
	}{
		{
			name: "nothing set",
			sc:   schema.SiteConfiguration{},
			want: false,
		},
		{
			name: "cody enabled",
			sc:   schema.SiteConfiguration{CodyEnabled: pointers.Ptr(true)},
			want: true,
		},
		{
			name: "cody disabled",
			sc:   schema.SiteConfiguration{CodyEnabled: pointers.Ptr(false)},
			want: false,
		},
		{
			name: "cody enabled, completions configured",
			sc:   schema.SiteConfiguration{CodyEnabled: pointers.Ptr(true), Completions: &schema.Completions{Model: "foobar"}},
			want: true,
		},
		{
			name: "cody disabled, completions enabled",
			sc:   schema.SiteConfiguration{CodyEnabled: pointers.Ptr(false), Completions: &schema.Completions{Enabled: pointers.Ptr(true), Model: "foobar"}},
			want: false,
		},
		{
			name: "cody disabled, completions configured",
			sc:   schema.SiteConfiguration{CodyEnabled: pointers.Ptr(false), Completions: &schema.Completions{Model: "foobar"}},
			want: false,
		},
		{
			// Legacy support: remove this once completions.enabled is removed
			name: "cody.enabled not set, completions configured but not enabled",
			sc:   schema.SiteConfiguration{Completions: &schema.Completions{Model: "foobar"}},
			want: false,
		},
		{
			// Legacy support: remove this once completions.enabled is removed
			name: "cody.enabled not set, completions configured and enabled",
			sc:   schema.SiteConfiguration{Completions: &schema.Completions{Enabled: pointers.Ptr(true), Model: "foobar"}},
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(&Unified{SiteConfiguration: test.sc})
			have := CodyEnabled()
			assert.Equal(t, test.want, have)
		})
	}
}

func TestGetCompletionsConfig(t *testing.T) {
	licenseKey := "theasdfkey"
	licenseAccessToken := license.GenerateLicenseKeyBasedAccessToken(licenseKey)
	zeroConfigDefaultWithLicense := &conftypes.CompletionsConfig{
		ChatModel:                "anthropic/claude-2",
		ChatModelMaxTokens:       12000,
		FastChatModel:            "anthropic/claude-instant-1",
		FastChatModelMaxTokens:   9000,
		CompletionModel:          "anthropic/claude-instant-1",
		CompletionModelMaxTokens: 9000,
		AccessToken:              licenseAccessToken,
		Provider:                 "sourcegraph",
		Endpoint:                 "https://cody-gateway.sourcegraph.com",
	}

	testCases := []struct {
		name         string
		siteConfig   schema.SiteConfiguration
		deployType   string
		wantConfig   *conftypes.CompletionsConfig
		wantDisabled bool
	}{
		{
			name: "Completions disabled",
			siteConfig: schema.SiteConfiguration{
				LicenseKey: licenseKey,
				Completions: &schema.Completions{
					Enabled: pointers.Ptr(false),
				},
			},
			wantDisabled: true,
		},
		{
			name: "Completions disabled, but Cody enabled",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Enabled: pointers.Ptr(false),
				},
			},
			// cody.enabled=true and completions.enabled=false, the newer
			// cody.enabled takes precedence and completions is enabled.
			wantConfig: zeroConfigDefaultWithLicense,
		},
		{
			name: "cody.enabled and empty completions object",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{},
			},
			wantConfig: zeroConfigDefaultWithLicense,
		},
		{
			name: "cody.enabled set false",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(false),
				Completions: &schema.Completions{},
			},
			wantDisabled: true,
		},
		{
			name: "no cody config",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: nil,
				Completions: nil,
			},
			wantDisabled: true,
		},
		{
			name: "Invalid provider",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Provider: "invalid",
				},
			},
			wantDisabled: true,
		},
		{
			name: "anthropic completions",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Enabled:     pointers.Ptr(true),
					Provider:    "anthropic",
					AccessToken: "asdf",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				ChatModel:                "claude-2",
				ChatModelMaxTokens:       12000,
				FastChatModel:            "claude-instant-1",
				FastChatModelMaxTokens:   9000,
				CompletionModel:          "claude-instant-1",
				CompletionModelMaxTokens: 9000,
				AccessToken:              "asdf",
				Provider:                 "anthropic",
				Endpoint:                 "https://api.anthropic.com/v1/complete",
			},
		},
		{
			name: "anthropic completions, with only completions.enabled",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Enabled:         pointers.Ptr(true),
					Provider:        "anthropic",
					AccessToken:     "asdf",
					ChatModel:       "claude-v1",
					CompletionModel: "claude-instant-1",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				ChatModel:                "claude-v1",
				ChatModelMaxTokens:       9000,
				FastChatModel:            "claude-instant-1",
				FastChatModelMaxTokens:   9000,
				CompletionModel:          "claude-instant-1",
				CompletionModelMaxTokens: 9000,
				AccessToken:              "asdf",
				Provider:                 "anthropic",
				Endpoint:                 "https://api.anthropic.com/v1/complete",
			},
		},
		{
			name: "soucregraph completions defaults",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Provider: "sourcegraph",
				},
			},
			wantConfig: zeroConfigDefaultWithLicense,
		},
		{
			name: "OpenAI completions completions",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Provider:    "openai",
					AccessToken: "asdf",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				ChatModel:                "gpt-4",
				ChatModelMaxTokens:       8000,
				FastChatModel:            "gpt-3.5-turbo",
				FastChatModelMaxTokens:   4000,
				CompletionModel:          "gpt-3.5-turbo-instruct",
				CompletionModelMaxTokens: 4000,
				AccessToken:              "asdf",
				Provider:                 "openai",
				Endpoint:                 "https://api.openai.com",
			},
		},
		{
			name: "Azure OpenAI completions completions",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Provider:        "azure-openai",
					AccessToken:     "asdf",
					Endpoint:        "https://acmecorp.openai.azure.com",
					ChatModel:       "gpt4-deployment",
					FastChatModel:   "gpt35-turbo-deployment",
					CompletionModel: "gpt35-turbo-deployment",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				ChatModel:                "gpt4-deployment",
				ChatModelMaxTokens:       8000,
				FastChatModel:            "gpt35-turbo-deployment",
				FastChatModelMaxTokens:   8000,
				CompletionModel:          "gpt35-turbo-deployment",
				CompletionModelMaxTokens: 8000,
				AccessToken:              "asdf",
				Provider:                 "azure-openai",
				Endpoint:                 "https://acmecorp.openai.azure.com",
			},
		},
		{
			name: "Fireworks completions completions",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Provider:    "fireworks",
					AccessToken: "asdf",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				ChatModel:                "accounts/fireworks/models/llama-v2-7b",
				ChatModelMaxTokens:       3000,
				FastChatModel:            "accounts/fireworks/models/llama-v2-7b",
				FastChatModelMaxTokens:   3000,
				CompletionModel:          "accounts/fireworks/models/starcoder-7b-w8a16",
				CompletionModelMaxTokens: 6000,
				AccessToken:              "asdf",
				Provider:                 "fireworks",
				Endpoint:                 "https://api.fireworks.ai/inference/v1/completions",
			},
		},
		{
			name: "AWS Bedrock completions completions",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					Provider: "aws-bedrock",
					Endpoint: "us-west-2",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				ChatModel:                "anthropic.claude-v2",
				ChatModelMaxTokens:       12000,
				FastChatModel:            "anthropic.claude-instant-v1",
				FastChatModelMaxTokens:   9000,
				CompletionModel:          "anthropic.claude-instant-v1",
				CompletionModelMaxTokens: 9000,
				AccessToken:              "",
				Provider:                 "aws-bedrock",
				Endpoint:                 "us-west-2",
			},
		},
		{
			name: "zero-config cody gateway completions without license key",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  "",
			},
			wantDisabled: true,
		},
		{
			name: "zero-config cody gateway completions with license key",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
			},
			wantConfig: zeroConfigDefaultWithLicense,
		},
		{
			name: "zero-config cody gateway completions without provider",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schema.Completions{
					ChatModel:       "anthropic/claude-v1.3",
					FastChatModel:   "anthropic/claude-instant-1.3",
					CompletionModel: "anthropic/claude-instant-1.3",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				ChatModel:                "anthropic/claude-v1.3",
				ChatModelMaxTokens:       9000,
				FastChatModel:            "anthropic/claude-instant-1.3",
				FastChatModelMaxTokens:   9000,
				CompletionModel:          "anthropic/claude-instant-1.3",
				CompletionModelMaxTokens: 9000,
				AccessToken:              licenseAccessToken,
				Provider:                 "sourcegraph",
				Endpoint:                 "https://cody-gateway.sourcegraph.com",
			},
		},
		{
			// Legacy support for completions.enabled
			name: "legacy field completions.enabled: zero-config cody gateway completions without license key",
			siteConfig: schema.SiteConfiguration{
				Completions: &schema.Completions{Enabled: pointers.Ptr(true)},
				LicenseKey:  "",
			},
			wantDisabled: true,
		},
		{
			name: "legacy field completions.enabled: zero-config cody gateway completions with license key",
			siteConfig: schema.SiteConfiguration{
				Completions: &schema.Completions{
					Enabled: pointers.Ptr(true),
				},
				LicenseKey: licenseKey,
			},
			// Not supported, zero-config is new and should be using the new
			// config.
			wantDisabled: true,
		},
		{
			name:       "app zero-config cody gateway completions with dotcom token",
			deployType: deploy.App,
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				App: &schema.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				AccessToken:              "sgd_5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456",
				ChatModel:                "anthropic/claude-2",
				ChatModelMaxTokens:       12000,
				FastChatModel:            "anthropic/claude-instant-1",
				FastChatModelMaxTokens:   9000,
				CompletionModel:          "anthropic/claude-instant-1",
				CompletionModelMaxTokens: 9000,
				Endpoint:                 "https://cody-gateway.sourcegraph.com",
				Provider:                 "sourcegraph",
			},
		},
		{
			name:       "app with custom configuration",
			deployType: deploy.App,
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				Completions: &schema.Completions{
					AccessToken:     "CUSTOM_TOKEN",
					Provider:        "anthropic",
					ChatModel:       "claude-v1",
					FastChatModel:   "claude-instant-1",
					CompletionModel: "claude-instant-1",
				},
				App: &schema.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wantConfig: &conftypes.CompletionsConfig{
				AccessToken:              "CUSTOM_TOKEN",
				ChatModel:                "claude-v1",
				ChatModelMaxTokens:       9000,
				CompletionModel:          "claude-instant-1",
				FastChatModelMaxTokens:   9000,
				FastChatModel:            "claude-instant-1",
				CompletionModelMaxTokens: 9000,
				Provider:                 "anthropic",
				Endpoint:                 "https://api.anthropic.com/v1/complete",
			},
		},
		{
			name:       "App but no dotcom username",
			deployType: deploy.App,
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				App: &schema.App{
					DotcomAuthToken: "",
				},
			},
			wantDisabled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaultDeploy := deploy.Type()
			if tc.deployType != "" {
				deploy.Mock(tc.deployType)
			}
			t.Cleanup(func() {
				deploy.Mock(defaultDeploy)
			})
			conf := GetCompletionsConfig(tc.siteConfig)
			if tc.wantDisabled {
				if conf != nil {
					t.Fatalf("expected nil config but got non-nil: %+v", conf)
				}
			} else {
				if conf == nil {
					t.Fatal("unexpected nil config returned")
				}
				if diff := cmp.Diff(tc.wantConfig, conf); diff != "" {
					t.Fatalf("unexpected config computed: %s", diff)
				}
			}
		})
	}
}

func TestGetEmbeddingsConfig(t *testing.T) {
	licenseKey := "theasdfkey"
	licenseAccessToken := license.GenerateLicenseKeyBasedAccessToken(licenseKey)
	defaultQdrantConfig := conftypes.QdrantConfig{
		QdrantHNSWConfig: conftypes.QdrantHNSWConfig{
			OnDisk: true,
		},
		QdrantOptimizersConfig: conftypes.QdrantOptimizersConfig{
			IndexingThreshold: 0,
			MemmapThreshold:   100,
		},
		QdrantQuantizationConfig: conftypes.QdrantQuantizationConfig{
			Enabled:  true,
			Quantile: 0.98,
		},
	}
	zeroConfigDefaultWithLicense := &conftypes.EmbeddingsConfig{
		Provider:                   "sourcegraph",
		AccessToken:                licenseAccessToken,
		Model:                      "openai/text-embedding-ada-002",
		Endpoint:                   "https://cody-gateway.sourcegraph.com/v1/embeddings",
		Dimensions:                 1536,
		Incremental:                true,
		MinimumInterval:            24 * time.Hour,
		MaxCodeEmbeddingsPerRepo:   3_072_000,
		MaxTextEmbeddingsPerRepo:   512_000,
		PolicyRepositoryMatchLimit: pointers.Ptr(5000),
		FileFilters: conftypes.EmbeddingsFileFilters{
			MaxFileSizeBytes: 1000000,
		},
		ExcludeChunkOnError: true,
		Qdrant:              defaultQdrantConfig,
	}

	testCases := []struct {
		name         string
		siteConfig   schema.SiteConfiguration
		deployType   string
		wantConfig   *conftypes.EmbeddingsConfig
		wantDisabled bool
	}{
		{
			name: "Embeddings disabled",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Enabled: pointers.Ptr(false),
				},
			},
			wantDisabled: true,
		},
		{
			name: "cody.enabled and empty embeddings object",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings:  &schema.Embeddings{},
			},
			wantConfig: zeroConfigDefaultWithLicense,
		},
		{
			name: "cody.enabled set false",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(false),
				Embeddings:  &schema.Embeddings{},
			},
			wantDisabled: true,
		},
		{
			name: "no cody config",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: nil,
				Embeddings:  nil,
			},
			wantDisabled: true,
		},
		{
			name: "Invalid provider",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Provider: "invalid",
				},
			},
			wantDisabled: true,
		},
		{
			name: "Implicit config with cody.enabled",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
			},
			wantConfig: zeroConfigDefaultWithLicense,
		},
		{
			name: "Sourcegraph provider",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Provider: "sourcegraph",
				},
			},
			wantConfig: zeroConfigDefaultWithLicense,
		},
		{
			name: "File filters",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Provider: "sourcegraph",
					FileFilters: &schema.FileFilters{
						MaxFileSizeBytes:         200,
						IncludedFilePathPatterns: []string{"*.go"},
						ExcludedFilePathPatterns: []string{"*.java"},
					},
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegraph",
				AccessToken:                licenseAccessToken,
				Model:                      "openai/text-embedding-ada-002",
				Endpoint:                   "https://cody-gateway.sourcegraph.com/v1/embeddings",
				Dimensions:                 1536,
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes:         200,
					IncludedFilePathPatterns: []string{"*.go"},
					ExcludedFilePathPatterns: []string{"*.java"},
				},
				ExcludeChunkOnError: true,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name: "Disable exclude failed chunk during indexing",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Provider: "sourcegraph",
					FileFilters: &schema.FileFilters{
						MaxFileSizeBytes:         200,
						IncludedFilePathPatterns: []string{"*.go"},
						ExcludedFilePathPatterns: []string{"*.java"},
					},
					ExcludeChunkOnError: pointers.Ptr(false),
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegraph",
				AccessToken:                licenseAccessToken,
				Model:                      "openai/text-embedding-ada-002",
				Endpoint:                   "https://cody-gateway.sourcegraph.com/v1/embeddings",
				Dimensions:                 1536,
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes:         200,
					IncludedFilePathPatterns: []string{"*.go"},
					ExcludedFilePathPatterns: []string{"*.java"},
				},
				ExcludeChunkOnError: false,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name: "No provider and no token, assume Sourcegraph",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Model: "openai/text-embedding-bobert-9000",
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegraph",
				AccessToken:                licenseAccessToken,
				Model:                      "openai/text-embedding-bobert-9000",
				Endpoint:                   "https://cody-gateway.sourcegraph.com/v1/embeddings",
				Dimensions:                 0, // unknown model used for test case
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name: "Sourcegraph provider without license",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  "",
				Embeddings: &schema.Embeddings{
					Provider: "sourcegraph",
				},
			},
			wantDisabled: true,
		},
		{
			name: "OpenAI provider",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Provider:    "openai",
					AccessToken: "asdf",
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "openai",
				AccessToken:                "asdf",
				Model:                      "text-embedding-ada-002",
				Endpoint:                   "https://api.openai.com/v1/embeddings",
				Dimensions:                 1536,
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name: "OpenAI provider without access token",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Provider: "openai",
				},
			},
			wantDisabled: true,
		},
		{
			name: "Azure OpenAI provider",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schema.Embeddings{
					Provider:    "azure-openai",
					AccessToken: "asdf",
					Endpoint:    "https://acmecorp.openai.azure.com",
					Dimensions:  1536,
					Model:       "the-model",
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "azure-openai",
				AccessToken:                "asdf",
				Model:                      "the-model",
				Endpoint:                   "https://acmecorp.openai.azure.com",
				Dimensions:                 1536,
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name:       "App default config",
			deployType: deploy.App,
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				App: &schema.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegraph",
				AccessToken:                "sgd_5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456",
				Model:                      "openai/text-embedding-ada-002",
				Endpoint:                   "https://cody-gateway.sourcegraph.com/v1/embeddings",
				Dimensions:                 1536,
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name: "App but no dotcom username",
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				App: &schema.App{
					DotcomAuthToken: "",
				},
			},
			wantDisabled: true,
		},
		{
			name:       "App with dotcom token",
			deployType: deploy.App,
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				Embeddings: &schema.Embeddings{
					Provider: "sourcegraph",
				},
				App: &schema.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegraph",
				AccessToken:                "sgd_5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456",
				Model:                      "openai/text-embedding-ada-002",
				Endpoint:                   "https://cody-gateway.sourcegraph.com/v1/embeddings",
				Dimensions:                 1536,
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name:       "App with user token",
			deployType: deploy.App,
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				Embeddings: &schema.Embeddings{
					Provider:    "sourcegraph",
					AccessToken: "TOKEN",
				},
			},
			wantConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegraph",
				AccessToken:                "TOKEN",
				Model:                      "openai/text-embedding-ada-002",
				Endpoint:                   "https://cody-gateway.sourcegraph.com/v1/embeddings",
				Dimensions:                 1536,
				Incremental:                true,
				MinimumInterval:            24 * time.Hour,
				MaxCodeEmbeddingsPerRepo:   3_072_000,
				MaxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMatchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MaxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
				Qdrant:              defaultQdrantConfig,
			},
		},
		{
			name:       "App without dotcom or user token",
			deployType: deploy.App,
			siteConfig: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
				Embeddings: &schema.Embeddings{
					Provider: "sourcegraph",
				},
			},
			wantDisabled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaultDeploy := deploy.Type()
			if tc.deployType != "" {
				deploy.Mock(tc.deployType)
			}
			t.Cleanup(func() {
				deploy.Mock(defaultDeploy)
			})
			conf := GetEmbeddingsConfig(tc.siteConfig)
			if tc.wantDisabled {
				if conf != nil {
					t.Fatalf("expected nil config but got non-nil: %+v", conf)
				}
			} else {
				if conf == nil {
					t.Fatal("unexpected nil config returned")
				}
				if diff := cmp.Diff(tc.wantConfig, conf); diff != "" {
					t.Fatalf("unexpected config computed: %s", diff)
				}
			}
		})
	}
}

func TestEmailSenderName(t *testing.T) {
	testCases := []struct {
		name       string
		siteConfig schema.SiteConfiguration
		want       string
	}{
		{
			name:       "nothing set",
			siteConfig: schema.SiteConfiguration{},
			want:       "Sourcegraph",
		},
		{
			name: "value set",
			siteConfig: schema.SiteConfiguration{
				EmailSenderName: "Horsegraph",
			},
			want: "Horsegraph",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Mock(&Unified{SiteConfiguration: tc.siteConfig})
			t.Cleanup(func() { Mock(nil) })

			if got, want := EmailSenderName(), tc.want; got != want {
				t.Fatalf("EmailSenderName() = %v, want %v", got, want)
			}
		})
	}
}
