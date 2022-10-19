package webhooks

import (
	"context"
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
	wh, err := db.Webhooks(keyring.Default().WebhookKey).Create(
		context.Background(),
		extsvc.KindGitHub,
		"http://github.com",
		u.ID,
		nil)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := NewHandler(db)

		handler.ServeHTTP(w, r)
	}))

	t.Run("found webhook returns unimplemented", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, wh.UUID)

		resp, err := http.Post(requestURL, "", nil)
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
}
