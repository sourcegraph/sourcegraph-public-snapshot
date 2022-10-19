package webhooks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhooksHandler(t *testing.T) {
	whUUID := uuid.New()
	db := database.NewMockDB()
	mockWebhooks := database.NewMockWebhookStore()
	mockWebhooks.GetByUUIDFunc.SetDefaultReturn(
		&types.Webhook{
			ID:           1,
			UUID:         whUUID,
			CodeHostKind: extsvc.KindGitHub,
		},
		nil)
	db.WebhooksFunc.SetDefaultReturn(mockWebhooks)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := NewHandler(db)

		handler.ServeHTTP(w, r)
	}))

	requestURL := fmt.Sprintf("%s/webhooks/%v", srv.URL, whUUID)

	resp, err := http.Post(requestURL, "", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}
