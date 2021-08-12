package executorqueue

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func init() {
	sharedConfig.FrontendUsername = "test"
	sharedConfig.FrontendPassword = "hunter2"
}

func TestInternalProxyAuthTokenMiddleware(t *testing.T) {
	ts := httptest.NewServer(basicAuthMiddleware(
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusUnauthorized, resp.StatusCode)
	}
	if value := resp.Header.Get("WWW-Authenticate"); value != `Basic realm="Sourcegraph"` {
		t.Errorf("unexpected www-authenticate header. want=%q have=%q", `Basic realm="Sourcegraph"`, value)
	}

	// wrong username
	req.SetBasicAuth(strings.ToUpper(sharedConfig.FrontendUsername), sharedConfig.FrontendPassword)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusForbidden, resp.StatusCode)
	}

	// wrong password
	req.SetBasicAuth(sharedConfig.FrontendUsername, strings.ToUpper(sharedConfig.FrontendPassword))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusForbidden, resp.StatusCode)
	}

	// correct token
	req.SetBasicAuth(sharedConfig.FrontendUsername, sharedConfig.FrontendPassword)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error performing request: %s", err)
	}
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("unexpected status code. want=%d have=%d", http.StatusTeapot, resp.StatusCode)
	}
}
