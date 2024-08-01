package publicrestapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/completions"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	txemail.DisableSilently()
}

type publicrestTest struct {
	t           *testing.T
	Handler     http.Handler
	AccessToken string
}

// Mock of the httpcli.Doer interface.
type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func newTest(t *testing.T) *publicrestTest {
	rcache.SetupForTest(t)

	assert.NoError(t, modelconfig.InitMock())
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	// Enable Cody (and all other license features)
	oldLicensingMock := licensing.MockCheckFeature
	licensing.MockCheckFeature = func(feature licensing.Feature) error {
		return nil
	}
	t.Cleanup(func() { licensing.MockCheckFeature = oldLicensingMock })

	// Mock the site configuration
	truePtr := true
	falsePtr := false
	licenseKey := "theasdfkey"
	licenseAccessToken := license.GenerateLicenseKeyBasedAccessToken(licenseKey)
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			CodyEnabled:     &truePtr,
			CodyPermissions: &falsePtr,
			Completions: &schema.Completions{
				AccessToken: licenseAccessToken,
			},
			ModelConfiguration: &schema.SiteModelConfiguration{
				ProviderOverrides: []*schema.ProviderOverride{
					{
						DisplayName: "Anthropic",
						Id:          "anthropic",
					},
				},
				ModelOverrides: []*schema.ModelOverride{
					{
						ModelRef:     "anthropic::xxxx::claude-sonnet-3.5-20240728",
						ModelName:    "claude-sonnet-3.5-20240728",
						Capabilities: []string{"chat", "completion", "fastChat"},
					},
				},
				DefaultModels: &schema.DefaultModels{
					Chat:           "anthropic::xxxx::claude-sonnet-3.5-20240728",
					CodeCompletion: "anthropic::xxxx::claude-sonnet-3.5-20240728",
					FastChat:       "anthropic::xxxx::claude-sonnet-3.5-20240728",
				},
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })

	// needs to happen after conf.Mock(...) to pick up model config
	assert.NoError(t, modelconfig.ResetMock())

	cfg, err := modelconfig.Get().Get()
	assert.NoError(t, err)
	if len(cfg.Models) == 0 {
		t.Fatal("expected model overrides")
	}

	enterpriseServices := enterprise.DefaultServices()
	rateLimitStore, _ := memstore.NewCtx(1024)
	rateLimiter := graphqlbackend.NewBasicLimitWatcher(logger, rateLimitStore)

	initialDoer := httpcli.CodyGatewayDoer
	t.Cleanup(func() {
		httpcli.CodyGatewayDoer = initialDoer
	})
	httpcli.CodyGatewayDoer = &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			fmt.Println("DOER_REQUEST:", r.Method, r.URL.String(), r.Header.Get("Authorization"))
			response := http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"id":"msg_01159fxVenqQDniWYomhMFYj","type":"message","role":"assistant","model":"claude-sonnet-3.5-20240728","content":[{"type":"text","text":"yes"}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":25,"output_tokens":32}}`)),
				Header:     make(http.Header),
			}
			response.Header.Add("Content-Type", "application/json")
			return &response, nil
		},
	}

	chatCompletionsStreamHandler := func() http.Handler {
		return completions.NewChatCompletionsStreamHandler(logger, db)
	}

	apiHandler, err := httpapi.NewHandler(db,
		nil,
		rateLimiter,
		&httpapi.Handlers{
			BatchesGitHubWebhook:            enterpriseServices.BatchesGitHubWebhook,
			BatchesGitLabWebhook:            enterpriseServices.BatchesGitLabWebhook,
			GitHubSyncWebhook:               enterpriseServices.ReposGithubWebhook,
			GitLabSyncWebhook:               enterpriseServices.ReposGitLabWebhook,
			BitbucketServerSyncWebhook:      enterpriseServices.ReposBitbucketServerWebhook,
			BitbucketCloudSyncWebhook:       enterpriseServices.ReposBitbucketCloudWebhook,
			BatchesBitbucketServerWebhook:   enterpriseServices.BatchesBitbucketServerWebhook,
			BatchesBitbucketCloudWebhook:    enterpriseServices.BatchesBitbucketCloudWebhook,
			BatchesAzureDevOpsWebhook:       enterpriseServices.BatchesAzureDevOpsWebhook,
			SCIMHandler:                     enterpriseServices.SCIMHandler,
			NewCodeIntelUploadHandler:       enterpriseServices.NewCodeIntelUploadHandler,
			NewComputeStreamHandler:         enterpriseServices.NewComputeStreamHandler,
			PermissionsGitHubWebhook:        enterpriseServices.PermissionsGitHubWebhook,
			NewChatCompletionsStreamHandler: chatCompletionsStreamHandler,
			NewCodeCompletionsHandler:       enterpriseServices.NewCodeCompletionsHandler,
		},
	)
	require.NoError(t, err)
	publicrestHandler := NewHandler(apiHandler)
	return &publicrestTest{
		t:           t,
		Handler:     publicrestHandler,
		AccessToken: licenseAccessToken,
	}
}

func TestAPI(t *testing.T) {
	c := newTest(t)

	t.Run("hello world", func(t *testing.T) {
		modelConfig := conf.GetCompletionsConfig(conf.Get().SiteConfiguration)
		fmt.Println("chatModel:", modelConfig.ChatModel)
		req, err := http.NewRequest("POST",
			"/api/openai/v1/chat/completions",
			strings.NewReader(`{
			    "model": "anthropic::xxxx::claude-sonnet-3.5-20240728",
			    "messages": [{"role": "user", "content": "respond with 'yes' and nothing else"}]
			}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
		assert.NoError(t, err)

		ctx := context.Background()
		req = req.WithContext(
			actor.WithActor(ctx, &actor.Actor{
				UID: 99,
			}),
		)

		rr := httptest.NewRecorder()
		c.Handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, []byte(body), "", "    ")
		if err != nil {
			t.Fatalf("Failed to format JSON: %v", err)
		}
		body = prettyJSON.String()

		autogold.Expect(`{
    "id": "chat-12345678-1234-1234-1234-123456789012",
    "choices": [
        {
            "finish_reason": "end_turn",
            "index": 0,
            "message": {
                "content": "yes",
                "role": "assistant"
            }
        }
    ],
    "created": 1722516460,
    "model": "anthropic::xxxx::claude-sonnet-3.5-20240728",
    "system_fingerprint": "",
    "object": "chat.completion",
    "usage": {
        "completion_tokens": 0,
        "prompt_tokens": 0,
        "total_tokens": 0
    }
}
`).Equal(t, body)
	})
}
