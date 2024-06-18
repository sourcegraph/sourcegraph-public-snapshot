package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"

	gh "github.com/google/go-github/v55/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestWebhooksHandler(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	u, err := db.Users().Create(context.Background(), database.NewUser{
		Email:           "test@user.com",
		Username:        "testuser",
		EmailIsVerified: true,
	})
	require.NoError(t, err)
	dbWebhooks := db.Webhooks(keyring.Default().WebhookKey)
	gitLabWH, err := dbWebhooks.Create(
		context.Background(),
		"gitLabWH",
		extsvc.KindGitLab,
		"http://gitlab.com",
		u.ID,
		types.NewUnencryptedSecret("somesecret"))
	require.NoError(t, err)

	gitHubWH, err := dbWebhooks.Create(
		context.Background(),
		"gitHubWH",
		extsvc.KindGitHub,
		"http://github.com",
		u.ID,
		types.NewUnencryptedSecret("githubsecret"),
	)
	require.NoError(t, err)

	gitHubWHNoSecret, err := dbWebhooks.Create(
		context.Background(),
		"gitHubWHNoSecret",
		extsvc.KindGitHub,
		"http://github.com",
		u.ID,
		nil,
	)
	require.NoError(t, err)

	bbServerWH, err := dbWebhooks.Create(
		context.Background(),
		"bbServerWH",
		extsvc.KindBitbucketServer,
		"http://bitbucket.com",
		u.ID,
		types.NewUnencryptedSecret("bbsecret"),
	)
	require.NoError(t, err)

	bbCloudWH, err := dbWebhooks.Create(
		context.Background(),
		"bb webhook",
		extsvc.KindBitbucketCloud,
		"http://bitbucket.com",
		u.ID,
		types.NewUnencryptedSecret("supersecretstring"),
	)
	require.NoError(t, err)

	bbCloudWHOtherSecret, err := dbWebhooks.Create(
		context.Background(),
		"bb webhook",
		extsvc.KindBitbucketCloud,
		"http://bitbucket.com",
		u.ID,
		types.NewUnencryptedSecret("othersecret"),
	)
	require.NoError(t, err)

	bbCloudWHNoSecret, err := dbWebhooks.Create(
		context.Background(),
		"bb webhook",
		extsvc.KindBitbucketCloud,
		"http://bitbucket.com",
		u.ID,
		nil,
	)
	require.NoError(t, err)

	azureDevOpsWH, err := dbWebhooks.Create(
		context.Background(),
		"ado webhook",
		extsvc.KindAzureDevOps,
		"https://dev.azure.com",
		u.ID,
		types.NewUnencryptedSecret("adosecret"),
	)
	require.NoError(t, err)

	wr := Router{Logger: logger, DB: db}
	gwh := GitHubWebhook{Router: &wr}

	webhookMiddleware := NewLogMiddleware(
		db.WebhookLogs(keyring.Default().WebhookLogKey),
	)

	base := mux.NewRouter()
	base.Path("/.api/webhooks/{webhook_uuid}").Methods("POST").Handler(webhookMiddleware.Logger(NewHandler(logger, db, gwh.Router)))
	srv := httptest.NewServer(base)

	t.Run("ping event from Github returns 200", func(t *testing.T) {
		wh := fakeWebhookHandler{}
		// need to call wr.Register to initialize the default handlers. Any eventType/codeHostKind will work.
		wr.Register(wh.handleEvent, extsvc.KindBitbucketCloud, "push")
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, gitHubWHNoSecret.UUID)
		payload := []byte(`{"body": "text"}`)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Github-Event", "ping")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("found GitLab webhook with correct secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, gitLabWH.UUID)

		event := webhooks.EventCommon{
			ObjectKind: "pipeline",
		}
		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindGitLab: {
				"pipeline": []Handler{wh.handleEvent},
			},
		}
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Add("X-GitLab-Token", "somesecret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, gitLabWH.CodeHostURN, wh.codeHostURNReceived)
		expectedEvent := webhooks.PipelineEvent{
			EventCommon: event,
		}
		assert.Equal(t, &expectedEvent, wh.eventReceived)
	})

	t.Run("not-found webhook returns 404", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, uuid.New())

		resp, err := http.Post(requestURL, "", nil)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("malformed UUID returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/SomeInvalidUUID", srv.URL)

		resp, err := http.Post(requestURL, "", nil)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("incorrect GitLab secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, gitLabWH.UUID)

		req, err := http.NewRequest("POST", requestURL, nil)
		require.NoError(t, err)
		req.Header.Add("X-GitLab-Token", "someothersecret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("correct GitHub secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, gitHubWH.UUID)

		h := hmac.New(sha1.New, []byte("githubsecret"))
		event := gh.PublicEvent{}
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		h.Write(payload)
		res := h.Sum(nil)

		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindGitHub: {
				"member": []Handler{wh.handleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Hub-Signature", "sha1="+hex.EncodeToString(res))
		req.Header.Set("X-Github-Event", "member")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		logs, _, err := db.WebhookLogs(keyring.Default().WebhookLogKey).List(context.Background(), database.WebhookLogListOpts{
			WebhookID: &gitHubWH.ID,
		})
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
		for _, log := range logs {
			assert.Equal(t, gitHubWH.ID, *log.WebhookID)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, gitHubWH.CodeHostURN, wh.codeHostURNReceived)
		expectedEvent := &gh.MemberEvent{}
		assert.Equal(t, expectedEvent, wh.eventReceived)
	})

	t.Run("not found handler returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, gitHubWH.UUID)

		h := hmac.New(sha1.New, []byte("githubsecret"))
		payload := []byte(`{"body": "text"}`)
		h.Write(payload)
		res := h.Sum(nil)

		wr.handlers = map[string]eventHandlers{}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Hub-Signature", "sha1="+hex.EncodeToString(res))
		req.Header.Set("X-Github-Event", "member")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GitHub with no secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, gitHubWHNoSecret.UUID)

		payload := []byte(`{"body": "text"}`)

		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindGitHub: {
				"member": []Handler{wh.handleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Github-Event", "member")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("incorrect GitHub secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, gitHubWH.UUID)

		h := hmac.New(sha1.New, []byte("wrongsecret"))
		payload := []byte(`{"body": "text"}`)
		h.Write(payload)
		res := h.Sum(nil)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Hub-Signature", "sha1="+hex.EncodeToString(res))
		req.Header.Set("X-Github-Event", "member")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("correct Bitbucket Server secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, bbServerWH.UUID)

		h := hmac.New(sha1.New, []byte("bbsecret"))
		event := bitbucketserver.PingEvent{}
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		h.Write(payload)
		res := h.Sum(nil)

		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindBitbucketServer: {
				"ping": []Handler{wh.handleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Hub-Signature", "sha1="+hex.EncodeToString(res))
		req.Header.Set("X-Event-Key", "ping")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		logs, _, err := db.WebhookLogs(keyring.Default().WebhookLogKey).List(context.Background(), database.WebhookLogListOpts{
			WebhookID: &bbServerWH.ID,
		})
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
		for _, log := range logs {
			assert.Equal(t, bbServerWH.ID, *log.WebhookID)
		}

		assert.Equal(t, bbServerWH.CodeHostURN, wh.codeHostURNReceived)
		assert.Equal(t, event, wh.eventReceived)
	})

	t.Run("incorrect Bitbucket server secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, bbServerWH.UUID)

		h := hmac.New(sha1.New, []byte("wrongsecret"))
		payload := []byte(`{"body": "text"}`)
		h.Write(payload)
		res := h.Sum(nil)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Hub-Signature", "sha1="+hex.EncodeToString(res))
		req.Header.Set("X-Event-Key", "ping")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Bitbucket Cloud returns 200 without secret", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, bbCloudWHNoSecret.UUID)

		event := bitbucketcloud.PullRequestCommentCreatedEvent{}
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindBitbucketCloud: {
				"pullrequest:comment_created": []Handler{wh.handleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Event-Key", "pullrequest:comment_created")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		logs, _, err := db.WebhookLogs(keyring.Default().WebhookLogKey).List(context.Background(), database.WebhookLogListOpts{
			WebhookID: &bbCloudWHNoSecret.ID,
		})
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
		for _, log := range logs {
			assert.Equal(t, bbCloudWHNoSecret.ID, *log.WebhookID)
		}
		assert.Equal(t, bbCloudWHNoSecret.CodeHostURN, wh.codeHostURNReceived)
		assert.Equal(t, &event, wh.eventReceived)
	})

	t.Run("Bitbucket Cloud returns 200 with correct secret", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, bbCloudWH.UUID)

		payload, err := os.ReadFile("testdata/bitbucketcloud_body.json")
		require.NoError(t, err)
		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindBitbucketCloud: {
				"repo:push": []Handler{wh.handleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Event-Key", "repo:push")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Time", "Tue, 11 Jun 2024 10:19:00 GMT")
		req.Header.Set("X-Hub-Signature", "sha256=4d0940b0224f927c796e6568837f4af66bd86845fcafab60a357feb2abc5561a")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		logs, _, err := db.WebhookLogs(keyring.Default().WebhookLogKey).List(context.Background(), database.WebhookLogListOpts{
			WebhookID: &bbCloudWH.ID,
		})
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
		for _, log := range logs {
			assert.Equal(t, bbCloudWH.ID, *log.WebhookID)
		}
		assert.Equal(t, bbCloudWH.CodeHostURN, wh.codeHostURNReceived)
		assert.Equal(t, "{dfeaae25-8168-466f-ade4-d07e6837dedf}", wh.eventReceived.(*bitbucketcloud.PushEvent).Repository.UUID)
	})

	t.Run("Bitbucket Cloud returns 400 with wrong secret", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, bbCloudWHOtherSecret.UUID)

		payload, err := os.ReadFile("testdata/bitbucketcloud_body.json")
		require.NoError(t, err)
		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindBitbucketCloud: {
				"repo:push": []Handler{wh.handleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Event-Key", "repo:push")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Time", "Tue, 11 Jun 2024 10:19:00 GMT")
		req.Header.Set("X-Hub-Signature", "sha256=4d0940b0224f927c796e6568837f4af66bd86845fcafab60a357feb2abc5561a")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Bitbucket Cloud returns 404 not found if webhook event type unknown", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, bbCloudWHNoSecret.UUID)

		payload := []byte(`{"body": "text"}`)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("X-Event-Key", "unknown_event")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Azure DevOps returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, azureDevOpsWH.UUID)

		event := azuredevops.PullRequestUpdatedEvent{EventType: "git.pullrequest.updated"}
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		wh := &fakeWebhookHandler{}
		wr.handlers = map[string]eventHandlers{
			extsvc.KindAzureDevOps: {
				"git.pullrequest.updated": []Handler{wh.handleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		logs, _, err := db.WebhookLogs(keyring.Default().WebhookLogKey).List(context.Background(), database.WebhookLogListOpts{
			WebhookID: &azureDevOpsWH.ID,
		})
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
		for _, log := range logs {
			assert.Equal(t, azureDevOpsWH.ID, *log.WebhookID)
		}
		assert.Equal(t, azureDevOpsWH.CodeHostURN, wh.codeHostURNReceived)
		assert.Equal(t, &event, wh.eventReceived)
	})

	t.Run("Azure DevOps returns 404 not found if webhook event type unknown", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.api/webhooks/%v", srv.URL, azureDevOpsWH.UUID)

		payload := []byte(`{"body": "text"}`)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

type fakeWebhookHandler struct {
	eventReceived       any
	codeHostURNReceived extsvc.CodeHostBaseURL
}

func (wh *fakeWebhookHandler) handleEvent(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, event any) error {
	wh.eventReceived = event
	wh.codeHostURNReceived = codeHostURN
	return nil
}
