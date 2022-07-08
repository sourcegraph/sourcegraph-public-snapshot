package gitlab

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"
	"github.com/stretchr/testify/require"

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

func TestClient_doWithBaseURLWithOAuthContext(t *testing.T) {
	baseURL, err := url.Parse("https://gitlab.com/")
	require.NoError(t, err)

	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			fmt.Println("header", r.Header.Get("Authorization"))
			return &http.Response{
				Status:     http.StatusText(http.StatusUnauthorized),
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}`))),
			}, nil
		},
	}
	client := NewClientProvider("Test", baseURL, doer,
		func(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.Context) (string, error) {
			fmt.Println("trying to refresh")
			return "refreshed-token", nil
		},
	).getClient(&auth.OAuthBearerToken{Token: "bad token"})

	ctx := context.Background()
	req, err := http.NewRequest(http.MethodGet, "todo", nil)
	require.NoError(t, err)

	var result interface{}
	_, _, err = client.doWithBaseURLWithOAuthContext(ctx, req, result, oauthutil.Context{})
	require.NoError(t, err)
}
