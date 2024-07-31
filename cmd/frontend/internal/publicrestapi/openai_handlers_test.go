package publicrestapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/completions"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
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

func newTest(t *testing.T) *publicrestTest {
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
			CodyPermissions: &falsePtr, // disable RBAC Cody permissions
			Completions: &schema.Completions{
				AccessToken: licenseAccessToken,
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })

	enterpriseServices := enterprise.DefaultServices()
	rateLimitStore, _ := memstore.NewCtx(1024)
	rateLimiter := graphqlbackend.NewBasicLimitWatcher(logger, rateLimitStore)

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
		req, err := http.NewRequest("POST", "/api/openai/v1/chat/completions", strings.NewReader(`{"model": "gpt-4o-mini-2024-07-18", "messages": [{"role": "user", "content": "respond with 'yes' and nothing else"}]}`))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.AccessToken))

		rr := httptest.NewRecorder()
		c.Handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		expectedResponse := `{
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
			"created": 1722438858,
			"model": "anthropic::unknown::claude-3-sonnet-20240229",
			"system_fingerprint": "",
			"object": "chat.completion",
			"usage": {
				"completion_tokens": 0,
				"prompt_tokens": 0,
				"total_tokens": 0
			}
		}`

		assert.JSONEq(t, expectedResponse, rr.Body.String())
		assert.NoError(t, err)
	})
}
