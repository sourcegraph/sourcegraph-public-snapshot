package completions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/azureopenai"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrytest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"

	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
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
	mockDB                database.DB
	chatCompletionHandler http.Handler
	codeCompletionHandler http.Handler
	mockGetModelFn        *mockGetModelFn
}

// PushGetModelResult sets what gets returned on the next call to getModelFn, which is invoked
// on every HTTP request to the completions endpoint. So you'll always need to call this before
// invoking the completions API.
func (ti *apiProviderTestInfra) PushGetModelResult(mref modelconfigSDK.ModelRef, err error) {
	ti.mockGetModelFn.PushResult(mref, err)
}

func (ti *apiProviderTestInfra) SetSiteConfig(t *testing.T, siteConfig schema.SiteConfiguration) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: siteConfig,
	})
	if err := modelconfig.ResetMock(); err != nil {
		require.NoError(t, err)
	}
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
	if err != nil {
		if strings.Contains(rawResponseJSON, "dial tcp: lookup") {
			// If you see this error, it is because a "real" HTTP client was used. And not one which
			// mas mocked out. So double check which HTTP client the LLM API Provider is using, and
			// that the apiProviderTestInfra mocked it out correctly. (e.g. as part of the
			// `AssertCodyGatewayReceivesRequestWithResponse` function.)
			t.Log("HINT: It looks like the HTTP request wasn't mocked out. Did you mock out the right httpcli.Doer?")
		}
		t.Errorf("error unmarshalling JSON\nError: %s\nFull response body:\n%s", err.Error(), rawResponseJSON)
	}

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
	t.Cleanup(func() {
		httpcli.UncachedExternalDoer = initialDoer
	})
	httpcli.UncachedExternalDoer = &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			return ti.assertDoerReceivesRequestAndSendResponse(t, r, opts)
		},
	}
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
		"Difference from the expected HTTP request body sent to 3rd party LLM provider:\nWant: %s\nGot : %s",
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
	require.NoError(t, modelconfig.InitMock())

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
		mockDB:                mockDB,
		mockGetModelFn:        &mockGetModelFn,
	}

	// Run the set of tests using that infra.
	testSuites := []struct {
		Name   string
		TestFn func(t *testing.T, infra *apiProviderTestInfra)
	}{
		{"BasicConfigChecks", testBasicConfiguration},
		{"APIProvider-Anthropic", testAPIProviderAnthropic},
		{"APIProvider-AWSBedrock", testAPIProviderAWSBedrock},
		{"APIProvider-AzureOpenAI", testAPIProviderAzureOpenAI},
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
			infra.SetSiteConfig(t, schema.SiteConfiguration{})

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
	})

	t.Run("WithDefaultModels", func(t *testing.T) {
		// Set the site configuration to have Cody enabled (from the previous test,
		// we were just missing the LicenseKey) but do not specify any completions.
		infra.SetSiteConfig(t, schema.SiteConfiguration{
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
			t.Run("InvalidModelFormat", func(t *testing.T) {
				// Verify we have a check to ensure that will fail a request if
				// the internal getModelFn returns an invalid ModelRef.
				t.Run("InvalidMRefReturned", func(t *testing.T) {
					for _, invalidMRef := range []string{"foo", "foo/bar", "a::b::c::d"} {
						infra.PushGetModelResult(modelconfigSDK.ModelRef(invalidMRef), nil)
						status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{})
						assert.Equal(t, http.StatusBadRequest, status)
						wantError := fmt.Sprintf("getModelFn(%q) returned invalid mref: modelRef syntax error", invalidMRef)
						assert.Contains(t, respBody, wantError)
					}
				})
				t.Run("ProviderNotFound", func(t *testing.T) {
					infra.PushGetModelResult("provider::api::model", nil)
					status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{})
					assert.Equal(t, http.StatusBadRequest, status)
					assert.Contains(t, respBody, `unable to find provider for mref "provider::api::model"`)
				})

				t.Run("UnknownModel", func(t *testing.T) {
					infra.PushGetModelResult("anthropic::api::model", nil)
					status, respBody := infra.CallCodeCompletionAPI(t, types.CodyCompletionRequestParameters{
						CompletionRequestParameters: types.CompletionRequestParameters{
							Stream: pointers.Ptr(false),
						},
					})
					assert.Equal(t, http.StatusBadRequest, status)
					assert.Contains(t, respBody, `unable to find model "anthropic::api::model"`)
				})
			})
		})
	})
}

// completionsRequestTestData bundles the test data we wish to validate when exercising one of our
// completion APIs. Tests are of the form:
//
// 1. Configure the local Sourcegraph instance to reflect the provided `SiteConfig`.
// 2. Invoke the UserCompletionRequest, with the `GetModelFn` returning the supplied string.
// 3. Verify that the outbound `WantRequestToLLMProvider` matches what is provided.
// 4. Simulate `ResponseFromLLMProvider` being returned.
// 5. Verify the completions handler returned the `WantCompletionsResponse` payload.
type completionsRequestTestData struct {
	// SiteConfig of the Sourcegraph instance.
	SiteConfig schema.SiteConfiguration

	// HTTP request the user sent to invoke the completions endpoint.
	UserCompletionRequest types.CodyCompletionRequestParameters

	// RequestToLLMProvider is the HTTP request that Sourcegraph will send to the
	// 3rd party LLM Provider. e.g. AWS Bedrock, Fireworks, or even Cody Gateway.
	//
	// BUG: We describe this as a map[string]any instead of a generic type because in
	// many cases the actual type isn't exported from the `internal/completions/client/...`
	// package. But if we were to actually export types, this would be much easier to maintain.
	WantRequestToLLMProvider map[string]any
	// URL path we expect the request to be sent to, e.g. "/v1/chat/completions"
	WantRequestToLLMProviderPath string
	// ResponseFromLLMProvider is the HTTP response sent from the 3rd party LLM Provider.
	ResponseFromLLMProvider map[string]any

	// WantCompletionsResponse is the result we expect the Sourcegraph Completions API to
	// return.
	WantCompletionResponse types.CompletionResponse
}

func runCompletionsTest(t *testing.T, infra *apiProviderTestInfra, data completionsRequestTestData) {
	ctx := context.Background()

	// Set the site config.
	infra.SetSiteConfig(t, data.SiteConfig)

	// Reset the modelconfig service mock too, so that it will pick up the updated
	// site config and make it available for tests.
	err := modelconfig.ResetMock()
	require.NoError(t, err)

	modelconfig, err := modelconfig.Get().Get()
	require.NoError(t, err)

	// When we make the completion request, the mocked getModelFn will be called.
	// Assuming the test request is a chat completion, we just use the "real" function
	// here.
	{
		realChatModelFn := getChatModelFn(infra.mockDB)
		getModelResult, getModelErr := realChatModelFn(ctx, data.UserCompletionRequest, modelconfig)
		infra.PushGetModelResult(getModelResult, getModelErr)
	}

	// Register our hook to verify that the expected LLM request was sent out.
	// The HTTP Client Doer we mock out is dependent on the API provider used.
	//
	// Since the mocked Doers are all cleaned up at the end of the testcase,
	// we just register every possible one so things (hopefully) just work.
	{
		require.NotEmpty(t, data.WantRequestToLLMProviderPath)

		infra.AssertCodyGatewayReceivesRequestWithResponse(
			t, assertLLMRequestOptions{
				WantRequestPath: data.WantRequestToLLMProviderPath,
				WantRequestObj:  &data.WantRequestToLLMProvider,
				OutResponseObj:  &data.ResponseFromLLMProvider,
			})
		infra.AssertGenericExternalAPIRequestWithResponse(
			t, assertLLMRequestOptions{
				WantRequestPath: data.WantRequestToLLMProviderPath,
				WantRequestObj:  &data.WantRequestToLLMProvider,
				OutResponseObj:  &data.ResponseFromLLMProvider,
			})

		// The Azure OpenAI SDK is a bit tricky to use. But we
		// hook into the AzureAPI client that gets created and
		// set it to the `UncachedExternalDoer` which was replaced
		// with our mock in the earlier call.
		azureopenai.MockAzureAPIClientTransport = httpcli.UncachedExternalDoer
		t.Cleanup(func() {
			azureopenai.MockAzureAPIClientTransport = nil
		})
	}

	// With all of our mocks in-place, we now simulate making the HTTP request.
	status, responseBody := infra.CallChatCompletionAPI(t, data.UserCompletionRequest)
	assert.Equal(t, http.StatusOK, status)
	infra.AssertCompletionsResponse(t, responseBody, data.WantCompletionResponse)
}
