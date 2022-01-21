package executorqueue

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInternalProxyAuthTokenMiddleware(t *testing.T) {
	accessToken := "hunter2"

	ts := httptest.NewServer(authMiddleware(
		func() string { return accessToken },
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}),
	))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}

	// no auth
	req.Header.Del("Authorization")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusUnauthorized, resp.StatusCode)
	}

	// malformed token
	req.Header.Set("Authorization", fmt.Sprintf("token-unknown %s", strings.ToUpper(accessToken)))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusUnauthorized, resp.StatusCode)
	}

	// wrong token
	req.Header.Set("Authorization", fmt.Sprintf("token-executor %s", strings.ToUpper(accessToken)))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusForbidden, resp.StatusCode)
	}

	// correct token
	req.Header.Set("Authorization", fmt.Sprintf("token-executor %s", accessToken))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusTeapot, resp.StatusCode)
	}
}
