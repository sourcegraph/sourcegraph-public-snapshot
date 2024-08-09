package modelconfig

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/embedded"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// verifyProviderByID will lookup the Provider from the ModelConfig and call the supplied function.
// If the provider isn't found, will report an error.
func verifyProviderByID(t *testing.T, cfg *types.ModelConfiguration, id string, fn func(t *testing.T, p types.Provider)) {
	t.Helper()
	t.Run(fmt.Sprintf("Provider=%s", id), func(t *testing.T) {
		for _, provider := range cfg.Providers {
			if provider.ID == types.ProviderID(id) {
				fn(t, provider)
				return
			}
		}

		t.Errorf("no provider with ID %q found", id)
	})
}

// verifyModelByMRef will lookup the Model from the ModelConfig and call the supplied function.
// If the model isn't found, will report an error.
func verifyModelByMRef(t *testing.T, cfg *types.ModelConfiguration, id string, fn func(t *testing.T, p types.Model)) {
	t.Helper()
	t.Run(fmt.Sprintf("Model=%s", id), func(t *testing.T) {
		var ranTest bool
		for _, model := range cfg.Models {
			if model.ModelRef == types.ModelRef(id) {
				if ranTest {
					t.Errorf("two models were found with the same ID %q", id)
				}
				fn(t, model)
				return
			}
		}

		t.Errorf("no model with ID %q found", id)
	})
}

func TestModelConfigBuilder(t *testing.T) {
	// Mock out the licensing check confirming Cody is enabled.
	initialMockCheck := licensing.MockCheckFeature
	licensing.MockCheckFeature = func(licensing.Feature) error {
		return nil // Don't fail when checking if Cody is enabled.
	}
	t.Cleanup(func() { licensing.MockCheckFeature = initialMockCheck })
	t.Cleanup(func() { conf.Mock(nil) })

	// setSiteCompletionsConfig sets site configuration data
	// to enable Cody Enterprise, and have the supplied completions
	// data.
	setSiteCompletionsConfig := func(userSuppliedCompConfig schema.Completions) *conftypes.CompletionsConfig {
		fauxSiteConfig := schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),
			LicenseKey:                   "license-key",

			Completions: &userSuppliedCompConfig,
		}
		return conf.GetCompletionsConfig(fauxSiteConfig)
	}

	type testData struct {
		codyGatewayConfig     *types.ModelConfiguration
		SiteCompletionsConfig schema.Completions
	}

	// verifyCanBuildModelConfig builds the types.ModelConfiguration based on the supplied data.
	// Will abort the test if any invalid configuration data is produced.
	verifyCanBuildModelConfig := func(t *testing.T, testData testData) *types.ModelConfiguration {
		baseConfig, err := embedded.GetCodyGatewayModelConfig()
		require.NoError(t, err)
		require.NoError(t, modelconfigSDK.ValidateModelConfig(baseConfig), "baseConfig is invalid!")

		// First round-trip the completions config through the `conf` package,
		// so that various hard-coded defualts from `conf/computed.go` get applied.
		// (Otherwise we'd test combinations that would never be possible in prod.)
		completionsConfig := setSiteCompletionsConfig(testData.SiteCompletionsConfig)

		// TODO: We are assuming that the SiteModelConfiguration ONLY comes from the
		// "completions" section of the site config. In the future, we'll add a new
		// format in the site configuration so users can express the SiteModelConfiguration
		// data structure directly. (Since it is more flexible than the completions config
		// that is available today.)
		siteConfigData, err := convertCompletionsConfig(completionsConfig)
		require.NoError(t, err)

		b := builder{
			staticData:      baseConfig,
			codyGatewayData: testData.codyGatewayConfig,
			siteConfigData:  siteConfigData,
		}
		result, err := b.build()
		require.NoError(t, err)
		require.NoError(t, modelconfigSDK.ValidateModelConfig(result), "produced ModelConfiguration was invalid")

		return result
	}

	// Provide the bare minimum of site configuration data. This will enable Cody Pro
	// and return the conftypes.CompletionsConfig as a types.ModelConfiguration.
	t.Run("NoOverrides", func(t *testing.T) {
		cfg := verifyCanBuildModelConfig(t, testData{})
		assert.Equal(t, 2, len(cfg.Providers), "providers: %v", cfg.Providers)
		assert.Equal(t, 3, len(cfg.Models), "models: %v", cfg.Models)

		// Verify Providers are configured as expected.
		verifyProviderByID(t, cfg, "anthropic", func(t *testing.T, prov types.Provider) {
			require.NotNil(t, prov.ServerSideConfig)
			// Use the "sourcegraph" LLM API provider under the hood.
			require.NotNil(t, prov.ServerSideConfig.SourcegraphProvider)
			require.Nil(t, prov.ServerSideConfig.GenericProvider)
		})
		verifyProviderByID(t, cfg, "fireworks", func(t *testing.T, prov types.Provider) {
			require.NotNil(t, prov.ServerSideConfig)
			// Use the "sourcegraph" LLM API provider under the hood.
			require.NotNil(t, prov.ServerSideConfig.SourcegraphProvider)
			require.Nil(t, prov.ServerSideConfig.GenericProvider)
		})

		// Verify DefaultModels.
		assert.Equal(t, types.DefaultModels{
			Chat:           types.ModelRef("anthropic::unknown::claude-3-sonnet-20240229"),
			CodeCompletion: types.ModelRef("fireworks::unknown::starcoder"),
			FastChat:       types.ModelRef("anthropic::unknown::claude-3-haiku-20240307"),
		}, cfg.DefaultModels)

		// Verify Models are configured as expected.
		verifyModelByMRef(t, cfg, "anthropic::unknown::claude-3-haiku-20240307", func(t *testing.T, model types.Model) {
			// The MaxInputTokens can be specified via the site config.
			assert.Equal(t, 12_000, model.ContextWindow.MaxInputTokens)
			assert.Equal(t, 4_000, model.ContextWindow.MaxOutputTokens)
		})

		verifyModelByMRef(t, cfg, "anthropic::unknown::claude-3-sonnet-20240229", func(t *testing.T, model types.Model) {
			// The MaxInputTokens can be specified via the site config.
			assert.Equal(t, 12_000, model.ContextWindow.MaxInputTokens)
			assert.Equal(t, 4_000, model.ContextWindow.MaxOutputTokens)
		})

		verifyModelByMRef(t, cfg, "fireworks::unknown::starcoder", func(t *testing.T, model types.Model) {
			// The MaxInputTokens can be specified via the site config.
			assert.Equal(t, 9_000, model.ContextWindow.MaxInputTokens)
			assert.Equal(t, 4_000, model.ContextWindow.MaxOutputTokens)
		})
	})

	// Verify that the special handling of AWS Provisioned Throughput ARNs is handled correctly.
	t.Run("AWSProvisionedThroughput", func(t *testing.T) {
		const (
			chatModelInConfig     = "anthropic.claude-3-haiku-20240307-v1:0-100k"
			fastChatModelInConfig = "anthropic.claude-v2-fastchat"

			priovisionedThroughputARN = "arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl"

			accessTokenInConfig = "<ACCESS_KEY_ID>:<SECRET_ACCESS_KEY>:<SESSION_TOKEN>"
			endpointInConfig    = "https://vpce-0a10b2345cd67e89f-abc0defg.bedrock-runtime.us-west-2.vpce.amazonaws.com"
		)

		cfg := verifyCanBuildModelConfig(t, testData{
			SiteCompletionsConfig: schema.Completions{
				Provider:    string(conftypes.CompletionsProviderNameAWSBedrock),
				AccessToken: accessTokenInConfig,
				Endpoint:    endpointInConfig,

				// The chat model is using provisioned throughput, hence the odd looking model name.
				ChatModel: chatModelInConfig + "/" + priovisionedThroughputARN,

				// Completions and FastChat are using the same model, one without provisioned throughput.
				CompletionModel: fastChatModelInConfig,
				FastChatModel:   fastChatModelInConfig,
			},
		})
		assert.Equal(t, 1, len(cfg.Providers), "providers: %v", cfg.Providers)
		assert.Equal(t, 2, len(cfg.Models), "models: %v", cfg.Models)

		verifyProviderByID(t, cfg, "anthropic", func(t *testing.T, prov types.Provider) {
			require.NotNil(t, prov.ServerSideConfig)
			require.NotNil(t, prov.ServerSideConfig.AWSBedrock)

			bedrockCfg := prov.ServerSideConfig.AWSBedrock
			require.Equal(t, accessTokenInConfig, bedrockCfg.AccessToken)
			require.Equal(t, endpointInConfig, bedrockCfg.Endpoint)
			require.Equal(t, "", bedrockCfg.Region)
		})

		verifyModelByMRef(t, cfg, "anthropic::unknown::"+fastChatModelInConfig, func(t *testing.T, model types.Model) {
			assert.Equal(t, fastChatModelInConfig, model.ModelName)
			// No server-side config for this one. It should just work.
			assert.Nil(t, model.ServerSideConfig)
		})

		// The mref includes a sanitized version of the chatModelInConfig.
		verifyModelByMRef(t, cfg, "anthropic::unknown::anthropic.claude-3-haiku-20240307-v1_0-100k", func(t *testing.T, model types.Model) {
			assert.Equal(t, chatModelInConfig, model.ModelName)
			//  Confirm the provisioned throughput was applied, too.
			require.NotNil(t, model.ServerSideConfig)
			require.NotNil(t, model.ServerSideConfig.AWSBedrockProvisionedThroughput)

			ptCfg := model.ServerSideConfig.AWSBedrockProvisionedThroughput
			assert.Equal(t, priovisionedThroughputARN, ptCfg.ARN)
		})
	})
}

func TestApplySiteConfig(t *testing.T) {

	// validModelWith returns a new, unique Model and applies the given override.
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	validModelWith := func(override types.ModelOverride) types.Model {
		modelID := fmt.Sprintf("test-model-%x", rng.Uint64())
		m := types.Model{
			ModelRef:     types.ModelRef(fmt.Sprintf("test-provider::v1::%s", modelID)),
			ModelName:    modelID,
			Category:     types.ModelCategoryBalanced,
			Capabilities: []types.ModelCapability{types.ModelCapabilityChat, types.ModelCapabilityAutocomplete},
			ContextWindow: types.ContextWindow{
				MaxInputTokens:  1,
				MaxOutputTokens: 1,
			},
		}

		err := modelconfig.ApplyModelOverride(&m, override)
		require.NoError(t, err)
		return m
	}

	toModelOverride := func(m types.Model) types.ModelOverride {
		return types.ModelOverride{
			ModelRef:      m.ModelRef,
			ModelName:     m.ModelName,
			Capabilities:  m.Capabilities,
			ContextWindow: m.ContextWindow,
			Category:      m.Category,
		}
	}

	t.Run("SourcegraphSuppliedModels", func(t *testing.T) {
		t.Run("StatusFilter", func(t *testing.T) {
			// The source config contains four models, one with each status.
			sourcegraphSuppliedConfig := types.ModelConfiguration{
				Providers: []types.Provider{
					{
						ID: types.ProviderID("test-provider"),
					},
				},
				Models: []types.Model{
					validModelWith(types.ModelOverride{
						Status: types.ModelStatusExperimental,
					}),
					validModelWith(types.ModelOverride{
						Status: types.ModelStatusBeta,
					}),
					validModelWith(types.ModelOverride{
						Status: types.ModelStatusStable,
					}),
					validModelWith(types.ModelOverride{
						Status: types.ModelStatusDeprecated,
					}),
				},
			}

			// The site configuration filters out all but "beta" and "stable".
			siteConfig := types.SiteModelConfiguration{
				SourcegraphModelConfig: &types.SourcegraphModelConfig{
					ModelFilters: &types.ModelFilters{
						StatusFilter: []string{"beta", "stable"},
					},
				},
			}

			gotConfig, err := applySiteConfig(&sourcegraphSuppliedConfig, &siteConfig)
			require.NoError(t, err)

			// Count the final models after the filter was applied.
			statusCounts := map[types.ModelStatus]int{}
			for _, model := range gotConfig.Models {
				statusCounts[model.Status]++
			}

			assert.Equal(t, 2, len(gotConfig.Models))
			assert.Equal(t, 1, statusCounts[types.ModelStatusBeta])
			assert.Equal(t, 1, statusCounts[types.ModelStatusStable])
		})
	})

	t.Run("FilterOutDeprecatedModels", func(t *testing.T) {
		// The source config contains four models, one with each status.
		sourcegraphSuppliedConfig := types.ModelConfiguration{
			Providers: []types.Provider{
				{
					ID: types.ProviderID("test-provider"),
				},
			},
			Models: []types.Model{
				validModelWith(types.ModelOverride{
					Status: types.ModelStatusExperimental,
				}),
				validModelWith(types.ModelOverride{
					Status: types.ModelStatusBeta,
				}),
				validModelWith(types.ModelOverride{
					Status: types.ModelStatusStable,
				}),
				validModelWith(types.ModelOverride{
					Status: types.ModelStatusDeprecated,
				}),
			},
		}

		// The default Cody configuration: use default settings with no customization.
		siteConfig := types.SiteModelConfiguration{
			SourcegraphModelConfig: &types.SourcegraphModelConfig{},
		}

		gotConfig, err := applySiteConfig(&sourcegraphSuppliedConfig, &siteConfig)
		require.NoError(t, err)

		assert.Equal(t, 3, len(gotConfig.Models))
		for _, model := range gotConfig.Models {
			if model.Status == types.ModelStatusDeprecated {
				t.Error("Deprecated models must be filtered out")
			}
		}
	})

	// This test covers the situation where the the default models from the base configuration
	// are removed due to model filter, but the site config doesn't provide valid values.
	t.Run("ReplacedDefaultModels", func(t *testing.T) {
		testModel := func(id string, capabilities []types.ModelCapability, category types.ModelCategory) types.Model {
			m := validModelWith(types.ModelOverride{
				Capabilities: capabilities,
				Category:     category,
			})
			m.ModelRef = types.ModelRef(fmt.Sprintf("test-provider::v1::%s", id))
			return m
		}

		getValidBaseConfig := func() types.ModelConfiguration {
			chatModel := testModel("chat", []types.ModelCapability{types.ModelCapabilityChat}, types.ModelCategoryAccuracy)
			codeModel := testModel("code", []types.ModelCapability{types.ModelCapabilityAutocomplete}, types.ModelCategorySpeed)
			return types.ModelConfiguration{
				Providers: []types.Provider{
					{
						ID: types.ProviderID("test-provider"),
					},
				},
				Models: []types.Model{
					chatModel,
					codeModel,
				},
				DefaultModels: types.DefaultModels{
					Chat:           chatModel.ModelRef,
					CodeCompletion: codeModel.ModelRef,
					FastChat:       chatModel.ModelRef,
				},
			}
		}

		t.Run("Base", func(t *testing.T) {
			baseConfig := getValidBaseConfig()
			_, err := applySiteConfig(&baseConfig, &types.SiteModelConfiguration{
				SourcegraphModelConfig: &types.SourcegraphModelConfig{}, // i.e. use the baseconfig.
			})
			require.NoError(t, err)
		})

		t.Run("ErrorNoChatModelAvail", func(t *testing.T) {
			// Now have the site config reject the chat model that was used as the default model.
			// This will now fail because there is nothing suitable.
			baseConfig := getValidBaseConfig()
			_, err := applySiteConfig(&baseConfig, &types.SiteModelConfiguration{
				SourcegraphModelConfig: &types.SourcegraphModelConfig{
					ModelFilters: &types.ModelFilters{
						Deny: []string{"*chat"},
					},
				},
			})
			assert.ErrorContains(t, err, "no suitable model found for Chat (1 candidates)")
		})

		t.Run("AlternativeUsed", func(t *testing.T) {
			t.Run("ErrorUnsuitableCandidate", func(t *testing.T) {
				// We add a new model from the site config, but the capability and category
				// make it unsuitable as the default chat model.
				modelInSiteConfig := testModel(
					"err-from-site-config", []types.ModelCapability{types.ModelCapabilityAutocomplete}, types.ModelCategorySpeed)

				baseConfig := getValidBaseConfig()
				_, err := applySiteConfig(&baseConfig, &types.SiteModelConfiguration{
					SourcegraphModelConfig: &types.SourcegraphModelConfig{
						ModelFilters: &types.ModelFilters{
							Deny: []string{"*chat"},
						},
					},
					ModelOverrides: []types.ModelOverride{
						toModelOverride(modelInSiteConfig),
					},
				})
				assert.ErrorContains(t, err, "no suitable model found for Chat (2 candidates)")
			})

			t.Run("ErrorStillNoSuitableCandidate", func(t *testing.T) {
				// This time it works, because the model's capabilities and category.
				//
				// However, we still get an error because there is no valid model for
				// the *fast chat*. Because "accuracy" isn't viable for fast chat,
				// it needs to be "speed" or "balanced".
				fromSiteConfig1 := testModel(
					"from-site-config1", []types.ModelCapability{types.ModelCapabilityChat}, types.ModelCategoryAccuracy)
				fromSiteConfig2 := testModel(
					"from-site-config2", []types.ModelCapability{types.ModelCapabilityAutocomplete}, types.ModelCategoryBalanced)

				baseConfig := getValidBaseConfig()
				_, err := applySiteConfig(&baseConfig, &types.SiteModelConfiguration{
					SourcegraphModelConfig: &types.SourcegraphModelConfig{
						ModelFilters: &types.ModelFilters{
							Deny: []string{"*chat"},
						},
					},
					ModelOverrides: []types.ModelOverride{
						toModelOverride(fromSiteConfig1),
						toModelOverride(fromSiteConfig2),
					},
				})
				assert.ErrorContains(t, err, "no suitable model found for FastChat (3 candidates)")
			})

			t.Run("Works", func(t *testing.T) {
				// This time it all works, because the new model is "balanced".
				modelInSiteConfig := testModel(
					"from-site-config", []types.ModelCapability{types.ModelCapabilityChat}, types.ModelCategoryBalanced)

				baseConfig := getValidBaseConfig()
				gotConfig, err := applySiteConfig(&baseConfig, &types.SiteModelConfiguration{
					SourcegraphModelConfig: &types.SourcegraphModelConfig{
						ModelFilters: &types.ModelFilters{
							Deny: []string{"*chat"},
						},
					},
					ModelOverrides: []types.ModelOverride{
						toModelOverride(modelInSiteConfig),
					},
				})
				require.NoError(t, err)
				assert.EqualValues(t, modelInSiteConfig.ModelRef, gotConfig.DefaultModels.Chat)
				assert.EqualValues(t, modelInSiteConfig.ModelRef, gotConfig.DefaultModels.FastChat)
			})
		})
	})
}
