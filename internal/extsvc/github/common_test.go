package github

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSplitRepositoryNameWithOwner(t *testing.T) {
	owner, name, err := SplitRepositoryNameWithOwner("a/b")
	if err != nil {
		t.Fatal(err)
	}
	if want := "a"; owner != want {
		t.Errorf("got owner %q, want %q", owner, want)
	}
	if want := "b"; name != want {
		t.Errorf("got name %q, want %q", name, want)
	}
}

type mockHTTPResponseBody struct {
	count        int
	responseBody string
	status       int
}

func (s *mockHTTPResponseBody) Do(req *http.Request) (*http.Response, error) {
	s.count++
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		Request:    req,
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(s.responseBody)),
	}, nil
}

func stringForRepoList(repos []*Repository) string {
	repoStrings := []string{}
	for _, repo := range repos {
		repoStrings = append(repoStrings, fmt.Sprintf("%#v", repo))
	}
	return "{\n" + strings.Join(repoStrings, ",\n") + "}\n"
}

func repoListsAreEqual(a []*Repository, b []*Repository) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if *a[i] != *b[i] {
			return false
		}
	}
	return true
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestClient_doRequestWithV3Client(t *testing.T) {
	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Authorization") == "Bearer bad token" {
				return &http.Response{
					Status:     http.StatusText(http.StatusUnauthorized),
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}`))),
				}, nil
			}

			body := `{"access_token": "refreshed-token", "token_type": "Bearer", "expires_in":3600, "refresh_token":"refresh-now", "scope":"create"}`
			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			}, nil

		},
	}

	ctx := context.Background()

	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	bearerToken := &auth.OAuthBearerToken{Token: "bad token"}

	v3Client := NewV3Client(logtest.Scoped(t), "Test", uri, bearerToken, doer)
	req, err := http.NewRequest(http.MethodGet, "url", nil)
	require.NoError(t, err)

	var result map[string]any
	_, err = doRequest(ctx, logtest.Scoped(t), v3Client.apiURL, v3Client.auth, v3Client.rateLimitMonitor, doer, req, &result)

	require.NoError(t, err)
}

func TestClient_doRequestWithV4Client(t *testing.T) {
	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Authorization") == "Bearer fafdfd token" {
				return &http.Response{
					Status:     http.StatusText(http.StatusUnauthorized),
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}`))),
				}, nil
			}

			body := `{"access_token": "refreshed-token", "token_type": "Bearer", "expires_in":3600, "refresh_token":"refresh-now", "scope":"create"}`
			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			}, nil

		},
	}

	ctx := context.Background()
	tokenRefresherFunc := func(ctx context.Context, doer httpcli.Doer, oauthCtxt oauthutil.OAuthContext) (*auth.OAuthBearerToken, error) {
		return &auth.OAuthBearerToken{Token: "refreshed-token"}, nil
	}

	v4Client, save := newV4Client(t, "GetAuthenticatedUserV4", tokenRefresherFunc)
	defer save()

	req, err := http.NewRequest(http.MethodGet, "url", nil)
	require.NoError(t, err)

	var result map[string]any
	_, err = doRequest(ctx, logtest.Scoped(t), v4Client.apiURL, v4Client.auth, v4Client.rateLimitMonitor, doer, req, &result)

	require.NoError(t, err)
}

func TestClient_doRequestWithoutARefresher(t *testing.T) {
	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Authorization") == "Bearer bad token" {
				return &http.Response{
					Status:     http.StatusText(http.StatusUnauthorized),
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}`))),
				}, nil
			}

			body := `{"access_token": "refreshed-token", "token_type": "Bearer", "expires_in":3600, "refresh_token":"refresh-now", "scope":"create"}`
			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			}, nil

		},
	}

	ctx := context.Background()

	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	bearerToken := &auth.OAuthBearerToken{AccessToken: "bad token"}

	v3Client := NewV3Client(logtest.Scoped(t), "Test", uri, bearerToken, doer)
	req, err := http.NewRequest(http.MethodGet, "url", nil)
	require.NoError(t, err)

	expectedError := "do request with retry and refresh: could not refresh token. Refresher is missing *oauthutil.oauthError"
	var result map[string]any
	_, err = doRequest(ctx, logtest.Scoped(t), v3Client.apiURL, v3Client.auth, v3Client.rateLimitMonitor, doer, req, &result)

	if err == nil || err.Error() != expectedError {
		t.Fatalf("received error: %v, want %s", err, expectedError)
	}
}
