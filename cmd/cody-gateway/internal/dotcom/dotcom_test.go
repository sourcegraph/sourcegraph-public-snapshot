package dotcom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpInQuery(t *testing.T) {
	var requestReceived bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		assert.Equal(t, r.URL.RawQuery, "CheckDotcomUserAccessToken")
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token", "random", "dev")
	// We don't care about the actual result of the call
	_, _ = CheckDotcomUserAccessToken(context.Background(), c, "slk_foobar")
	// But we do care that we did get through to the handler
	assert.True(t, requestReceived)
}
