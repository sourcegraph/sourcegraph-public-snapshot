package gitlab

import (
	"context"
	"flag"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
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
	provider := NewClientProvider("Test", baseURL, doer)
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
