package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOldDefsRedirect(t *testing.T) {
	router := New(nil)

	tests := map[string]string{
		"/r@c/.GoPackage/u/.def/p":     "/r@c/-/def/GoPackage/u/-/p",
		"/r@c/.GoPackage/u1/u2/.def/p": "/r@c/-/def/GoPackage/u1/u2/-/p",
		"/r@c/.GoPackage/u/.def/p1/p2": "/r@c/-/def/GoPackage/u/-/p1/p2",

		"/r@c/-/def/GoPackage/u/-/p/-/refs?repo=r2": "/r@c/-/info/GoPackage/u/-/p",
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
