package completions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"

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
			// Any unrecognized formats just return an empty provider.
			{
				Input:        "just-model",
				WantProvider: "",
				WantModel:    "just-model",
			},
			{
				Input:        "a//b",
				WantProvider: "",
				WantModel:    "a//b",
			},
			{
				Input:        "a::b::c",
				WantProvider: "",
				WantModel:    "a::b::c",
			},
		}

		for _, test := range tests {
			gotProvider, gotModel := legacyModelRef(test.Input).Parse()
			assert.Equal(t, test.WantProvider, gotProvider, "legacyModelRef(%q)", test.Input)
			assert.Equal(t, test.WantModel, gotModel, "legacyModelRef(%q)", test.Input)
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
			assert.Equal(t, test.Want, got, "isAllowedCodyProChatModel(%q)", lmref)
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
				require.ErrorContains(t, err, `unsupported code completion model "some-model-not-in-config"`)
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
