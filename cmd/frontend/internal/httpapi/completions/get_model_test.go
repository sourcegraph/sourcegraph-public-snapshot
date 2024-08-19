package completions

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/embedded"
	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

// getModelResult is a mock implementation of the getModelFn, one of the parameters
// to newCompletionsHandler.
type getModelResult struct {
	ModelRef modelconfigSDK.ModelRef
	Err      error
}

type mockGetModelFn struct {
	results []getModelResult
}

func (m *mockGetModelFn) PushResult(mref modelconfigSDK.ModelRef, err error) {
	result := getModelResult{
		ModelRef: mref,
		Err:      err,
	}
	m.results = append(m.results, result)
}

func (m *mockGetModelFn) ToFunc() getModelFn {
	return func(
		ctx context.Context, requestParams types.CodyCompletionRequestParameters, c *modelconfigSDK.ModelConfiguration) (
		modelconfigSDK.ModelRef, error) {
		if len(m.results) == 0 {
			panic("no call registered to getModels")
		}
		v := m.results[0]
		m.results = m.results[1:]
		return v.ModelRef, v.Err
	}
}

func TestLegacyModelRef(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		tests := []struct {
			Input        string
			WantProvider string
			WantModel    string
		}{
			{
				Input:        "openai/gpt-4o",
				WantProvider: "openai",
				WantModel:    "gpt-4o",
			},

			// Awkward case with more than one slash.
			{
				Input:        "a/b/c",
				WantProvider: "a",
				WantModel:    "b/c",
			},
			{
				Input:        "fireworks/accounts/fireworks/models/mixtral-8x22b-instruct",
				WantProvider: "fireworks",
				WantModel:    "accounts/fireworks/models/mixtral-8x22b-instruct",
			},
			{
				Input:        "a//b",
				WantProvider: "a",
				WantModel:    "/b",
			},

			// Any unrecognized formats just return an empty provider.
			{
				Input:        "just-model",
				WantProvider: "",
				WantModel:    "just-model",
			},
			{
				Input:        "a::b::c",
				WantProvider: "",
				WantModel:    "a::b::c",
			},
		}

		for _, test := range tests {
			testLegacyRef := legacyModelRef(test.Input)
			gotProvider, gotModel := testLegacyRef.Parse()
			assert.Equal(t, test.WantProvider, gotProvider, "legacyModelRef(%q)", test.Input)
			assert.Equal(t, test.WantModel, gotModel, "legacyModelRef(%q)", test.Input)

			// BONUS TEST! Verify the ToModelRef function works as expected.
			wantMRef := fmt.Sprintf("%s::unknown::%s", gotProvider, gotModel)
			gotMRef := testLegacyRef.ToModelRef()
			assert.Equal(t, wantMRef, string(gotMRef))
		}
	})

	t.Run("EqualToIgnoringAPIVersion", func(t *testing.T) {
		tests := []struct {
			InLegacyRef string
			InMRef      string
			Want        bool
		}{
			// Matches.
			{
				InLegacyRef: "openai/gpt-4o",
				InMRef:      "openai::v1::gpt-4o",
				Want:        true,
			},
			{
				InLegacyRef: "openai/gpt-4o",
				InMRef:      "openai::v2::gpt-4o",
				Want:        true,
			},
			// Only the model name is supplied, no provider at all.
			{
				InLegacyRef: "gpt-4o",
				InMRef:      "openai::v2::gpt-4o",
				Want:        true,
			},
			{
				InLegacyRef: "gpt-4o",
				InMRef:      "something-other-than-openai::v2::gpt-4o",
				Want:        true,
			},

			// Non matches.
			{
				InLegacyRef: "openai/gpt-4o",
				InMRef:      "openai::v1::gpt-4o-turbo",
				Want:        false,
			},
			{
				InLegacyRef: "openai/gpt-4o",
				InMRef:      "openai::v1::gpt-4o-turbo",
				Want:        false,
			},
			{
				InLegacyRef: "gpt-4o",
				InMRef:      "openai::v1::gpt-4o-turbo",
				Want:        false,
			},
		}

		for _, test := range tests {
			legacyMRef := legacyModelRef(test.InLegacyRef)
			mref := modelconfigSDK.ModelRef(test.InMRef)
			got := legacyMRef.EqualToIgnoringAPIVersion(mref)
			assert.Equal(t, test.Want, got, "%q.EqualToIgnoringAPIVersion(%q)", test.InLegacyRef, test.InMRef)
		}
	})
}

func TestCodyProModelAllowlists(t *testing.T) {
	t.Run("CompletionModels", func(t *testing.T) {
		tests := []struct {
			LegacyMRef string
			Want       bool
		}{
			{
				LegacyMRef: "fireworks/starcoder2-7b",
				Want:       true,
			},
			{
				LegacyMRef: "google/gemini-pro",
				Want:       true,
			},

			// Unknown or disallowed models.
			{
				LegacyMRef: "google/gemini-pro-v999-max-R3",
				Want:       false,
			},
		}

		for _, test := range tests {
			lmref := legacyModelRef(test.LegacyMRef)
			got := isAllowedCodyProCompletionModel(lmref)
			assert.Equal(t, test.Want, got, "isAllowedCodyProCompletionModel(%q)", lmref)
		}
	})
	t.Run("ChatModels", func(t *testing.T) {
		tests := []struct {
			LegacyMRef string
			IsPro      bool
			Want       bool
		}{
			// Pro-only models.
			{
				LegacyMRef: "openai/gpt-4-turbo-preview",
				IsPro:      true,
				Want:       true,
			},
			{
				LegacyMRef: "openai/gpt-4-turbo-preview",
				IsPro:      false,
				Want:       false,
			},

			// Pro and Free models.
			{
				LegacyMRef: "openai/gpt-3.5-turbo",
				IsPro:      true,
				Want:       true,
			},
			{
				LegacyMRef: "openai/gpt-3.5-turbo",
				IsPro:      false,
				Want:       true,
			},

			{
				LegacyMRef: "fireworks/accounts/fireworks/models/mixtral-8x22b-instruct",
				IsPro:      true,
				Want:       true,
			},
			{
				LegacyMRef: "fireworks/accounts/fireworks/models/mixtral-8x22b-instruct",
				IsPro:      false,
				Want:       true,
			},

			// Unknown models.
			{
				LegacyMRef: "openai/gpt-6-omega",
				IsPro:      true,
				Want:       false,
			},
			{
				LegacyMRef: "openai/gpt-6-omega",
				IsPro:      false,
				Want:       false,
			},
		}

		for _, test := range tests {
			lmref := legacyModelRef(test.LegacyMRef)
			got := isAllowedCodyProChatModel(lmref, test.IsPro)
			assert.Equal(t, test.Want, got, "isAllowedCodyProChatModel(%q, %v)", lmref, test.IsPro)
		}
	})

	// Confirm that all Sourcegraph-supplied LLM models resolve to the correct models.
	t.Run("SourcegraphSuppliedModels", func(t *testing.T) {
		staticConfig, err := embedded.GetCodyGatewayModelConfig()
		require.NoError(t, err)

		for _, sourcegraphSuppliedModel := range staticConfig.Models {
			t.Run(string(sourcegraphSuppliedModel.ModelRef), func(t *testing.T) {
				var supportsChat bool
				for _, capability := range sourcegraphSuppliedModel.Capabilities {
					if capability == modelconfigSDK.ModelCapabilityChat {
						supportsChat = true
						break
					}
				}
				if !supportsChat {
					t.Skipf("NA. Skipping model %q as it does not support chat.", sourcegraphSuppliedModel.ModelRef)
				}

				legacyModelRef := toLegacyMRef(sourcegraphSuppliedModel.ModelRef)
				got := isAllowedCodyProChatModel(legacyModelRef, true)
				assert.True(t, got)
			})
		}
	})

	// Confirm that every Sourcegraph-supplied LLM model with a virtualized ID is
	// found in our lookup.
	t.Run("VirutalizedModels", func(t *testing.T) {
		staticConfig, err := embedded.GetCodyGatewayModelConfig()
		require.NoError(t, err)

		for _, sourcegraphSuppliedModel := range staticConfig.Models {
			t.Run(string(sourcegraphSuppliedModel.ModelRef), func(t *testing.T) {
				modelID := string(sourcegraphSuppliedModel.ModelRef.ModelID())
				modelName := sourcegraphSuppliedModel.ModelName

				gotModelName, ok := virutalizedModelRefLookup[modelID]
				if modelID == modelName {
					// Not virtualized.
					assert.False(t, ok)
				} else {
					// Is virtualized.
					assert.True(t, ok)
					assert.Equal(t, modelName, gotModelName)
				}
			})
		}
	})
}

func TestGetCodeCompletionsModelFn(t *testing.T) {
	ctx := context.Background()
	getModelFn := getCodeCompletionModelFn()

	t.Run("ErrorUnsupportedModel", func(t *testing.T) {
		reqParams := types.CodyCompletionRequestParameters{
			CompletionRequestParameters: types.CompletionRequestParameters{
				RequestedModel: "model-the-user-requested",
			},
		}
		_, err := getModelFn(ctx, reqParams, nil /* modelconfigSDK.ModelConfiguration */)
		require.ErrorContains(t, err, "no configuration data supplied")

		_, err2 := getModelFn(ctx, reqParams, &modelconfigSDK.ModelConfiguration{})
		require.ErrorContains(t, err2, `unsupported code completion model "model-the-user-requested"`)
	})

	t.Run("OverrideSiteConfig", func(t *testing.T) {
		// Empty model config, except that it does contain the expected model.
		modelConfig := modelconfigSDK.ModelConfiguration{
			Models: []modelconfigSDK.Model{
				{ModelRef: "google::xxxx::some-other-model1"},
				{ModelRef: "google::xxxx::gemini-pro"},
				{ModelRef: "google::xxxx::some-other-model2"},
			},
		}
		reqParams := types.CodyCompletionRequestParameters{
			// BUG: This is inconsistent with how user-requested models work for "chats", which
			// totally ignore user preferences. Here we _always_ honor the user's preference.
			//
			// We should reject requests to use models the calling user cannot access, or are not
			// available to the current "Cody Pro Subscription" or "Cody Enterprise config".
			CompletionRequestParameters: types.CompletionRequestParameters{
				RequestedModel: "google/gemini-pro",
			},
		}
		gotMRef, err := getModelFn(ctx, reqParams, &modelConfig)
		require.NoError(t, err)
		assert.EqualValues(t, "google::xxxx::gemini-pro", gotMRef)
	})

	t.Run("Default", func(t *testing.T) {
		// For these tests, the Model field in the request body isn't set.
		// The default Code Completion model should be returned.
		t.Run("NoSiteConfig", func(t *testing.T) {
			reqParams := types.CodyCompletionRequestParameters{}
			_, err := getModelFn(ctx, reqParams, nil)
			assert.ErrorContains(t, err, "no configuration data supplied")
		})
		t.Run("WithSiteConfig", func(t *testing.T) {
			modelConfig := modelconfigSDK.ModelConfiguration{
				Models: []modelconfigSDK.Model{
					{ModelRef: "other-model-1"},
					{ModelRef: "other-model-2"},
					{ModelRef: "code-model-in-config"},
					{ModelRef: "other-model-3"},
				},
				DefaultModels: modelconfigSDK.DefaultModels{
					CodeCompletion: "code-model-in-config",
				},
			}

			reqParams := types.CodyCompletionRequestParameters{}
			model, err := getModelFn(ctx, reqParams, &modelConfig)
			require.NoError(t, err)
			assert.EqualValues(t, "code-model-in-config", model)
		})
	})
}

func TestGetChatModelFn(t *testing.T) {
	ctx := context.Background()
	mockDB := dbmocks.NewMockDB()

	t.Run("CodyEnterprise", func(t *testing.T) {
		t.Run("Chat", func(t *testing.T) {
			getModelFn := getChatModelFn(mockDB)
			modelConfig := modelconfigSDK.ModelConfiguration{
				Models: []modelconfigSDK.Model{
					{ModelRef: "some-other-model"},
					{ModelRef: "model-the-user-requested"},
				},
				DefaultModels: modelconfigSDK.DefaultModels{
					Chat: "default-chat-model",
				},
			}

			t.Run("Found", func(t *testing.T) {
				reqParams := types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						RequestedModel: "model-the-user-requested",
					},
				}
				model, err := getModelFn(ctx, reqParams, &modelConfig)

				require.NoError(t, err)
				assert.EqualValues(t, "model-the-user-requested", model)
			})

			// User requests to use an LLM model not supported by the backend.
			t.Run("NotFound", func(t *testing.T) {
				reqParams := types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						RequestedModel: "some-model-not-in-config",
					},
				}
				_, err := getModelFn(ctx, reqParams, &modelConfig)
				require.ErrorContains(t, err, `unsupported chat model "some-model-not-in-config"`)
			})
		})

		t.Run("FastChat", func(t *testing.T) {
			getModelFn := getChatModelFn(mockDB)
			modelConfig := modelconfigSDK.ModelConfiguration{
				Models: []modelconfigSDK.Model{
					{ModelRef: "some-other-model"},
					{ModelRef: "model-the-user-requested"},
				},
				DefaultModels: modelconfigSDK.DefaultModels{
					Chat:     "default-chat-model",
					FastChat: "default-fastchat-model",
				},
			}

			reqParams := types.CodyCompletionRequestParameters{
				CompletionRequestParameters: types.CompletionRequestParameters{
					RequestedModel: "model-the-user-requested",
				},
				// ... but again, for "fast" chats.
				Fast: true,
			}
			model, err := getModelFn(ctx, reqParams, &modelConfig)

			require.NoError(t, err)
			// We use the FastChat model, regardless of what the user requested.
			assert.EqualValues(t, "default-fastchat-model", model)
		})
	})

	// TODO(PRIME-283): As part of enabling model selection for Cody Enterprise users,
	// add more tests for the Cody Pro path as well. Where we only allow certain models
	// based on the calling user's subscription status, etc.
}
