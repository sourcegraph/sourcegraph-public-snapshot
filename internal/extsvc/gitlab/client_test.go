package gitlab

import (
	"bytes"
	"context"
	"flag"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
)

func TestGetAuthenticatedUserOAuthScopes(t *testing.T) {
	// To update this test's fixtures, use the GitLab token stored in
	// 1Password under gitlab@sourcegraph.com.
	client := createTestClient(t)
	ctx := context.Background()
	have, err := client.GetAuthenticatedUserOAuthScopes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"read_user", "read_api", "api"}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

func createTestProvider(t *testing.T) *ClientProvider {
	t.Helper()
	fac, cleanup := httptestutil.NewRecorderFactory(t, update(t.Name()), t.Name())
	t.Cleanup(cleanup)
	doer, err := fac.Doer()
	if err != nil {
		t.Fatal(err)
	}
	baseURL, _ := url.Parse("https://gitlab.com/")
	provider := NewClientProvider("Test", baseURL, doer, nil)
	return provider
}

func createTestClient(t *testing.T) *Client {
	t.Helper()
	token := os.Getenv("GITLAB_TOKEN")
	return createTestProvider(t).GetOAuthClient(token)
}

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestClient_doWithBaseURL(t *testing.T) {
	baseURL, err := url.Parse("https://gitlab.com/")
	require.NoError(t, err)

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

	mockOauthContext := &oauthutil.OAuthContext{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "url/oauth/authorize",
			TokenURL: "url/oauth/token",
		},
		Scopes: []string{"read_user"},
	}

	provider := NewClientProvider("Test", baseURL, doer, func(ctx context.Context, doer httpcli.Doer, oauthCtxt oauthutil.OAuthContext) (string, error) {
		return "refreshed-token", nil
	})

	client := provider.getClient(&auth.OAuthBearerToken{Token: "bad token"})

	req, err := http.NewRequest(http.MethodGet, "url", nil)
	require.NoError(t, err)

	var result map[string]any
	_, _, err = client.doWithBaseURL(ctx, mockOauthContext, req, &result)
	require.NoError(t, err)
}
