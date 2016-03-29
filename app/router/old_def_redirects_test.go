package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOldDefsRedirect(t *testing.T) {
	router := New(nil)

	tests := map[string]string{
		"/r@c/.GoPackage/u/.def/p":     "/r@c/-/def/GoPackage/u/p",
		"/r@c/.GoPackage/u1/u2/.def/p": "/r@c/-/def/GoPackage/u1--u2/p",
		"/r@c/.GoPackage/u/.def/p1/p2": "/r@c/-/def/GoPackage/u/p1--p2",

		// These routes will 404. See addOldDefRedirectRoute doc for more info.
		"/r@c/.GoPackage/.def/p": "/r@c/-/def/GoPackage/p",
		"/r@c/.GoPackage/u/.def": "/r@c/-/def/GoPackage/u",
		"/r@c/.GoPackage/.def":   "/r@c/-/def/GoPackage",
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
