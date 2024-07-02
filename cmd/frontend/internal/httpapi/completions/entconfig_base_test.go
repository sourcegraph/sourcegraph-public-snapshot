package completions

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrytest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

type mockRateLimiter struct{}

func (*mockRateLimiter) TryAcquire(ctx context.Context) error {
	return nil
}

// Mock of the httpcli.Doer interface.
type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

// apiProviderTestInfra bundles the various mocks and things necessary for
// running an API provider test.
type apiProviderTestInfra struct {
	// Don't use these directly. Tests should just use the exported fields and functions.
	chatCompletionHandler http.Handler
	codeCompletionHandler http.Handler
	mockGetModelFn        *mockGetModelFn
}

// PushGetModelResult sets what gets returned on the next call to getModelFn, which is invoked
// on every HTTP request to the completions endpoint. So you'll always need to call this before
// invoking the completions API.
func (ti *apiProviderTestInfra) PushGetModelResult(model string, err error) {
	ti.mockGetModelFn.PushResult(model, err)
}

func (ti *apiProviderTestInfra) SetSiteConfig(siteConfig schema.SiteConfiguration) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: siteConfig,
	})
}

func (ti *apiProviderTestInfra) CallChatCompletionAPI(t *testing.T, reqObj types.CodyCompletionRequestParameters) (int, string) {
	return ti.makeCompletionRequest(t, ti.chatCompletionHandler, reqObj)
}
func (ti *apiProviderTestInfra) CallCodeCompletionAPI(t *testing.T, reqObj types.CodyCompletionRequestParameters) (int, string) {
	return ti.makeCompletionRequest(t, ti.codeCompletionHandler, reqObj)
}

// Issues an HTTP request to the given HTTP handler with the supplied payload. Returns the HTTP status and response body.
func (ti *apiProviderTestInfra) makeCompletionRequest(t *testing.T, handler http.Handler, reqObj types.CodyCompletionRequestParameters) (int, string) {
	t.Helper()

	// Convert the request into JSON.
	reqBody, err := json.Marshal(reqObj)
	require.NoError(t, err)

	// Build the request.
	req := httptest.NewRequest(
		http.MethodPost,
		"/cody/completions/handler/api",
		strings.NewReader(string(reqBody)))

	mockUser := actor.FromMockUser(1337)
	ctx := actor.WithActor(context.Background(), mockUser)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Make the request
	handler.ServeHTTP(w, req)
	resp := w.Result()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(respBody)
}

// AssertCompletionsResponse verifies that the JSON text matches the supplied
// CompletionResponse object.
func (ti *apiProviderTestInfra) AssertCompletionsResponse(t *testing.T, rawResponseJSON string, wantResponse types.CompletionResponse) {
	t.Helper()

	var gotResponse types.CompletionResponse
	err := json.Unmarshal([]byte(rawResponseJSON), &gotResponse)
	require.NoError(t, err, "error unmarshalling JSON, full response:\n%s", rawResponseJSON)

	assert.Equal(t, wantResponse, gotResponse)
}

type assertLLMRequestOptions struct {
	// WantRequestObj is what we expect the outbound HTTP request's JSON body
	// to be equal to. Required.
	WantRequestObj any
	// OutResponseObj is serialized to JSON and sent to the caller, i.e. our
	// LLM API Provider which is making the API request. Required.
	OutResponseObj any

	// WantRequestPath is the URL Path we expect in the outbound HTTP request.
	// No check is done if empty.
	WantRequestPath string

	// WantHeaders are HTTP header key/value pairs that must be present.
	WantHeaders map[string]string
}

// See comment on `assertDoerReceivesRequestAndSendResponse`.
func (ti *apiProviderTestInfra) AssertCodyGatewayReceivesRequestWithResponse(
	t *testing.T, opts assertLLMRequestOptions) {
	initialDoer := httpcli.CodyGatewayDoer
	t.Cleanup(func() {
		httpcli.CodyGatewayDoer = initialDoer
	})
	httpcli.CodyGatewayDoer = &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			return ti.assertDoerReceivesRequestAndSendResponse(t, r, opts)
		},
	}
}

// See comment on `assertDoerReceivesRequestAndSendResponse`.
func (ti *apiProviderTestInfra) AssertGenericExternalAPIRequestWithResponse(
	t *testing.T, opts assertLLMRequestOptions) {
	initialDoer := httpcli.UncachedExternalDoer
	t.Logf("Saving initialDoer which was %v", initialDoer)
	t.Cleanup(func() {
		t.Logf("Resetting initialDoer for generic external %v", initialDoer)
		httpcli.UncachedExternalDoer = initialDoer
	})
	httpcli.UncachedExternalDoer = &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			return ti.assertDoerReceivesRequestAndSendResponse(t, r, opts)
		},
	}
	t.Logf("Replaced initialDoer for generic external to %v", httpcli.UncachedExternalDoer)
}

// assertDoerReceivesRequestAndSendResponse is an http.HandlerFunc that we hook into
// the httpcli.Doer's used for sending LLM API requests.
//
// This HTTP handler will verify that the HTTP request being sent matches what is expected.
// e.g. that the outbound URL path and that the HTTP request's body (as unmarshalled
// JSON) matches the provided values.
//
// The handler then returns a generic 200 OK response, with the the JSON response body
// matching the respObj.
func (ti *apiProviderTestInfra) assertDoerReceivesRequestAndSendResponse(
	t *testing.T, r *http.Request, opts assertLLMRequestOptions) (*http.Response, error) {
	t.Helper()

	// Verify aspects of the request metadata.
	if opts.WantRequestPath != "" {
		assert.Equal(t, opts.WantRequestPath, r.URL.Path)
	}
	for header, wantValue := range opts.WantHeaders {
		assert.Equal(t, wantValue, r.Header.Get(header), "all request headers: %+v", r.Header)
	}

	// We don't know what the actual type of the request object is, and even
	// if we did, verifying it matches the incomming JSON isn't straight forward.
	// So we just marshall each to a generic `map[string]any` and compare the two.
	var wantReqPayload map[string]any
	reqObjJSON, err := json.Marshal(opts.WantRequestObj)
	require.NoError(t, err)
	err = json.Unmarshal(reqObjJSON, &wantReqPayload)
	require.NoError(t, err)
	require.True(t, len(wantReqPayload) > 0, "req object was empty? JSON `%s`", string(reqObjJSON))

	var gotReqPayload map[string]any
	reqBodyJSON, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	err = json.Unmarshal(reqBodyJSON, &gotReqPayload)
	require.NoError(t, err)

	// Compare the incomming HTTP request matches what we expect.
	assert.EqualValues(t, wantReqPayload, gotReqPayload,
		"Comparing the raw vs. expected payloads:\nWant: %s\nGot : %s",
		string(reqObjJSON), string(reqBodyJSON))

	respObjJSON, err := json.Marshal(opts.OutResponseObj)
	require.NoError(t, err)

	respBodyReadCloser := io.NopCloser(strings.NewReader(string(respObjJSON)))
	okResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       respBodyReadCloser,
	}
	return okResponse, nil
}

// TestAPIProviders is the big deal, it calls into specific implemntations.
// To cut down on the boilerplate, they inject various things.
func TestAPIProviders(t *testing.T) {
	// Configure all the mocks necessary for testing completion handlers.
	initialMockCheck := licensing.MockCheckFeature
	licensing.MockCheckFeature = func(licensing.Feature) error {
		return nil // Don't fail when checking if Cody is enabled.
	}
	t.Cleanup(func() { licensing.MockCheckFeature = initialMockCheck })
	t.Cleanup(func() { conf.Mock(nil) })

	// Set up mocks.
	logger := logtest.NoOp(t)
	mockDB := dbmocks.NewMockDB()

	mockGetModelFn := mockGetModelFn{}
	eventRecorder := telemetry.NewEventRecorder(telemetrytest.NewMockEventsStore())

	// No idea, but the TokenManager assumes that it has access to Redis.
	rcache.SetupForTest(t)

	// Create the HTTP handlers.
	chatCompletionHandler := newCompletionsHandler(
		logger,
		mockDB,
		nil, // database.UserStore
		nil, // database.AccessTokenStore
		eventRecorder,
		nil, // guardrails.AttributionTest
		types.CompletionsFeatureChat,
		&mockRateLimiter{},
		"trace-family",
		mockGetModelFn.ToFunc())
	codeCompletionHandler := newCompletionsHandler(
		logger,
		mockDB,
		nil, // database.UserStore
		nil, // database.AccessTokenStore
		eventRecorder,
		nil, // guardrails.AttributionTest
		types.CompletionsFeatureCode,
		&mockRateLimiter{},
		"trace-family",
		mockGetModelFn.ToFunc())

	// Bundle into a test infra object, for convenience.
	testInfra := apiProviderTestInfra{
		chatCompletionHandler: chatCompletionHandler,
		codeCompletionHandler: codeCompletionHandler,
		mockGetModelFn:        &mockGetModelFn,
	}

	// Run the set of tests using that infra.
	testSuites := []struct {
		Name   string
		TestFn func(t *testing.T, infra *apiProviderTestInfra)
	}{
		{"BasicConfigChecks", testBasicConfiguration},
		{"APIProviderAnthropic", testAPIProviderAnthropic},
	}
	for _, testSuite := range testSuites {
		t.Run(testSuite.Name, func(t *testing.T) {
			testSuite.TestFn(t, &testInfra)
		})
	}
}

func testBasicConfiguration(t *testing.T, infra *apiProviderTestInfra) {
	t.Run("Errors", func(t *testing.T) {
		t.Run("CodyNotEnabled", func(t *testing.T) {
			infra.SetSiteConfig(schema.SiteConfiguration{})

			{
				status, respBody := infra.CallChatCompletionAPI(t, types.CodyCompletionRequestParameters{})
				assert.Equal(t, http.StatusUnauthorized, status)
				assert.Equal(t, "cody is not enabled: cody is disabled\n", respBody)
			}
			{
				status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{})
				assert.Equal(t, http.StatusUnauthorized, status)
				assert.Equal(t, "cody is not enabled: cody is disabled\n", respBody)
			}
		})

		t.Run("NoCompletionsConfig", func(t *testing.T) {
			infra.SetSiteConfig(schema.SiteConfiguration{
				CodyEnabled:                  pointers.Ptr(true),
				CodyPermissions:              pointers.Ptr(false),
				CodyRestrictUsersFeatureFlag: pointers.Ptr(false),

				Completions: nil,
			})

			t.Run("Complete", func(t *testing.T) {
				status, respBody := infra.CallChatCompletionAPI(t, types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						Stream: pointers.Ptr(true),
					},
				})
				assert.Equal(t, http.StatusInternalServerError, status)
				assert.Equal(t, "completions are not configured or disabled\n", respBody)
			})
			t.Run("Streaming", func(t *testing.T) {
				status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						Stream: pointers.Ptr(true),
					},
				})
				assert.Equal(t, http.StatusInternalServerError, status)
				assert.Equal(t, "completions are not configured or disabled\n", respBody)
			})
		})
	})

	t.Run("WithDefaultModels", func(t *testing.T) {
		// Set the site configuration to have Cody enabled (from the previous test,
		// we were just missing the LicenseKey) but do not specify any completions.
		infra.SetSiteConfig(schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),

			// LicenseKey is required in order to use Cody.
			LicenseKey:  "license-key",
			Completions: nil,
		})

		t.Run("ConfirmDefaultsSet", func(t *testing.T) {
			modelConfig := conf.GetCompletionsConfig(conf.Get().SiteConfiguration)

			assert.Equal(t, "anthropic/claude-3-sonnet-20240229", modelConfig.ChatModel)
			assert.Equal(t, "fireworks/starcoder", modelConfig.CompletionModel)
			assert.Equal(t, "anthropic/claude-3-haiku-20240307", modelConfig.FastChatModel)

			assert.Greater(t, modelConfig.ChatModelMaxTokens, 3000)
			assert.Greater(t, modelConfig.CompletionModelMaxTokens, 3000)
			assert.Greater(t, modelConfig.FastChatModelMaxTokens, 3000)
		})

		// When the call to getModelFn -- how the HTTP endpoint knows which model to
		// use for serving the request -- returns an error, we serve it directly to
		// the end user. As usually these contain user-facing messages.
		t.Run("ErrorGettingModel", func(t *testing.T) {
			t.Run("Sync", func(t *testing.T) {
				infra.PushGetModelResult("NA", errors.New("error-from-getModelFn"))
				status, respBody := infra.CallChatCompletionAPI(t, types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						Stream: pointers.Ptr(false),
					},
				})
				assert.Equal(t, http.StatusBadRequest, status)
				assert.Equal(t, "error-from-getModelFn\n", respBody)
			})
			t.Run("Streaming", func(t *testing.T) {
				infra.PushGetModelResult("NA", errors.New("error-from-getModelFn"))
				status, respBody := infra.CallChatCompletionAPI(t, types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						Stream: pointers.Ptr(false),
					},
				})
				assert.Equal(t, http.StatusBadRequest, status)
				assert.Equal(t, "error-from-getModelFn\n", respBody)
			})
		})

		t.Run("ErrorOnUnknownModel", func(t *testing.T) {
			// Cody Gateway will reject the model with a different error name if it is sent something
			// without any slashes.
			t.Run("InvalidModelFormat", func(t *testing.T) {
				infra.PushGetModelResult("model-name-no-slashes", nil)
				status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{})
				assert.Equal(t, http.StatusInternalServerError, status)
				assert.Equal(t,
					"no provider provided in model model-name-no-slashes - a model in the format '$PROVIDER/$MODEL_NAME' is expected",
					respBody)
			})

			// In these tests, we resolve the request to use an unknown model.
			// The error originates from Cody Gateway which is serving a 400.
			// BUG: We serve these as 500s on our side, but they should be 4xx.
			t.Run("Sync", func(t *testing.T) {
				infra.PushGetModelResult("acmeco/llm-tron-9k", nil)
				status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						Stream: pointers.Ptr(false),
					},
				})
				assert.Equal(t, http.StatusInternalServerError, status)
				assert.Equal(t, "no client known for upstream provider acmeco", respBody)
			})
			t.Run("Streaming", func(t *testing.T) {
				infra.PushGetModelResult("acmeco/llm-tron-9k", nil)
				status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{
					CompletionRequestParameters: types.CompletionRequestParameters{
						// If nil, defaults to streaming.
					},
				})
				assert.Equal(t, http.StatusInternalServerError, status)
				assert.Equal(t, "no client known for upstream provider acmeco", respBody)
			})
		})
	})
}
