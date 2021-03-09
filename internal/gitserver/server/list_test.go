package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_handleList(t *testing.T) {
	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()
	_, ok := s.locker.TryAcquire("a", "test status")
	if !ok {
		t.Fatal("could not acquire lock")
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/list", nil)
	h.ServeHTTP(rr, req)

	body := strings.TrimSpace(rr.Body.String())
	if want := `[]`; body != want {
		t.Errorf("got %q, want %q", body, want)
	}
}
