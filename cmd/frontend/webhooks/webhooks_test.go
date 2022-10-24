package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhooksHandler(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	u, err := db.Users().Create(context.Background(), database.NewUser{
		Email:           "test@user.com",
		Username:        "testuser",
		EmailIsVerified: true,
	})
	require.NoError(t, err)
	dbWebhooks := db.Webhooks(keyring.Default().WebhookKey)
	gitLabWH, err := dbWebhooks.Create(
		context.Background(),
		extsvc.KindGitLab,
		"http://gitlab.com",
		u.ID,
		types.NewUnencryptedSecret("somesecret"))
	require.NoError(t, err)

	gitHubWH, err := dbWebhooks.Create(
		context.Background(),
		extsvc.KindGitHub,
		"http://github.com",
		u.ID,
		types.NewUnencryptedSecret("githubsecret"),
	)
	require.NoError(t, err)

	gitHubWHNoSecret, err := dbWebhooks.Create(
		context.Background(),
		extsvc.KindGitHub,
		"http://github.com",
		u.ID,
		nil,
	)

	require.NoError(t, err)
	gh := GitHubWebhook{
		DB: db,
	}

	srv := httptest.NewServer(NewHandler(logger, db, &gh))

	t.Run("found GitLab webhook with correct secret returns unimplemented", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, gitLabWH.UUID)

		req, err := http.NewRequest("POST", requestURL, nil)
		require.NoError(t, err)
		req.Header.Add("X-GitLab-Token", "somesecret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	})

	t.Run("not-found webhook returns 404", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, uuid.New())

		resp, err := http.Post(requestURL, "", nil)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("malformed UUID returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/SomeInvalidUUID", srv.URL)

		resp, err := http.Post(requestURL, "", nil)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("incorrect GitLab secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, gitLabWH.UUID)

		req, err := http.NewRequest("POST", requestURL, nil)
		require.NoError(t, err)
		req.Header.Add("X-GitLab-Token", "someothersecret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("correct GitHub secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, gitHubWH.UUID)

		h := hmac.New(sha1.New, []byte("githubsecret"))
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

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GitHub with no secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, gitHubWHNoSecret.UUID)

		payload := []byte(`{"body": "text"}`)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Github-Event", "member")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("incorrect GitHub secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, gitHubWH.UUID)

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
}
