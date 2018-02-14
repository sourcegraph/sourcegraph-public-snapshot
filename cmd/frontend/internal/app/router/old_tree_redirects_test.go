package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOldTreesRedirect(t *testing.T) {
	router := Router()

	tests := map[string]string{
		"/r@c/.tree":       "/r@c/-/tree",
		"/r@c/.tree/p":     "/r@c/-/tree/p",
		"/r@c/.tree/p1/p2": "/r@c/-/tree/p1/p2",
	}
	for oldURL, wantNewURL := range tests {
		rw := httptest.NewRecorder()
		req, err := http.NewRequest("GET", oldURL, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(rw, req)

		if got := rw.Header().Get("location"); got != wantNewURL {
			t.Errorf("%s: got %s, want %s", oldURL, got, wantNewURL)
		}
	}
}
