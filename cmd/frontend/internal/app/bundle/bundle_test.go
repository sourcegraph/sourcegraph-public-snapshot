package bundle

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	r := httptest.NewRequest("GET", entrypointPath, nil)
	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, r)
	body := string(w.Body.Bytes())
	if !strings.Contains(body, "not enabled") {
		t.Errorf("expected to get response 'not enabled', got %s", body)
	}
}
