package releasecache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"testing/iotest"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestHandler_HandleBranch(t *testing.T) {
	logger, _ := logtest.Captured(t)

	t.Run("no branch in version cache", func(t *testing.T) {
		rc := NewMockReleaseCache()
		rc.CurrentFunc.SetDefaultHook(func(branch string) (string, error) {
			assert.Equal(t, "3.43", branch)
			return "", branchNotFoundError(branch)
		})
		handler := &handler{logger: logger, rc: rc}

		rec := httptest.NewRecorder()

		handler.handleBranch(rec, "3.43")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("other error from version cache", func(t *testing.T) {
		rc := NewMockReleaseCache()
		rc.CurrentFunc.SetDefaultHook(func(branch string) (string, error) {
			assert.Equal(t, "3.43", branch)
			return "", errors.New("error!")
		})
		handler := &handler{logger: logger, rc: rc}

		rec := httptest.NewRecorder()

		handler.handleBranch(rec, "3.43")
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		rc := NewMockReleaseCache()
		rc.CurrentFunc.SetDefaultHook(func(branch string) (string, error) {
			assert.Equal(t, "3.43", branch)
			return "3.43.9", nil
		})
		handler := &handler{logger: logger, rc: rc}

		rec := httptest.NewRecorder()

		handler.handleBranch(rec, "3.43")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "\"3.43.9\"", rec.Body.String())
	})
}

func TestHandler_HandleWebhook(t *testing.T) {
	logger, _ := logtest.Captured(t)

	t.Run("payload error", func(t *testing.T) {
		handler := &handler{logger: logger}

		rec := httptest.NewRecorder()
		body := iotest.ErrReader(errors.New("error!"))
		req := httptest.NewRequest("POST", "/.api/src-cli/versions/webhook", body)

		handler.doHandleWebhook(rec, req, nil)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "invalid payload\n", rec.Body.String())
	})

	t.Run("signature error", func(t *testing.T) {
		handler := &handler{logger: logger, webhookSecret: "secret"}

		rec := httptest.NewRecorder()
		body := bytes.NewBufferString("body")
		req := httptest.NewRequest("POST", "/.api/src-cli/versions/webhook", body)
		req.Header.Add("X-Hub-Signature", "signature")

		handler.doHandleWebhook(rec, req, func(signature string, payload, secret []byte) error {
			assert.Equal(t, "signature", signature)
			assert.Equal(t, "body", string(payload))
			assert.Equal(t, handler.webhookSecret, string(secret))

			return errors.New("error!")
		})
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "invalid signature\n", rec.Body.String())
	})

	t.Run("update error", func(t *testing.T) {
		rc := NewMockReleaseCache()
		rc.UpdateNowFunc.SetDefaultReturn(errors.New("error!"))
		handler := &handler{logger: logger, rc: rc, webhookSecret: "secret"}

		rec := httptest.NewRecorder()
		body := bytes.NewBufferString("body")
		req := httptest.NewRequest("POST", "/.api/src-cli/versions/webhook", body)
		req.Header.Add("X-Hub-Signature", "signature")

		handler.doHandleWebhook(rec, req, func(signature string, payload, secret []byte) error {
			assert.Equal(t, "signature", signature)
			assert.Equal(t, "body", string(payload))
			assert.Equal(t, handler.webhookSecret, string(secret))

			return nil
		})
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("valid", func(t *testing.T) {
		rc := NewMockReleaseCache()
		rc.UpdateNowFunc.SetDefaultReturn(nil)
		handler := &handler{logger: logger, rc: rc, webhookSecret: "secret"}

		rec := httptest.NewRecorder()
		body := bytes.NewBufferString("body")
		req := httptest.NewRequest("POST", "/.api/src-cli/versions/webhook", body)
		req.Header.Add("X-Hub-Signature", "signature")

		handler.doHandleWebhook(rec, req, func(signature string, payload, secret []byte) error {
			assert.Equal(t, "signature", signature)
			assert.Equal(t, "body", string(payload))
			assert.Equal(t, handler.webhookSecret, string(secret))

			return nil
		})
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}

func mustParseUrl(t *testing.T, uri string) *url.URL {
	u, err := url.Parse(uri)
	require.NoError(t, err)
	return u
}
