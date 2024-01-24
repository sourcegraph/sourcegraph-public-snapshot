package codygateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

type attributionHandler struct {
	t        *testing.T
	requests []codygateway.AttributionRequest
}

func (h *attributionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request codygateway.AttributionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.requests = append(h.requests, request)
	w.WriteHeader(http.StatusOK)
	require.NoError(h.t, json.NewEncoder(w).Encode(codygateway.AttributionResponse{
		Repositories: []codygateway.AttributionRepository{{"repo1"}, {"repo2"}},
		LimitHit:     false,
	}))
}

func TestAttribution(t *testing.T) {
	h := &attributionHandler{
		t: t,
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	client := codygateway.NewClient(http.DefaultClient, srv.URL, "token")
	attribution, err := client.Attribution(context.Background(), "snippet", 3)
	require.NoError(t, err)
	require.Equal(t, []codygateway.AttributionRequest{{
		Snippet: "snippet",
		Limit:   3,
	}}, h.requests)
	require.Equal(t, codygateway.Attribution{
		Repositories: []string{"repo1", "repo2"},
		LimitHit:     false,
	}, attribution)
}
