package gitlab

import (
	"context"
	"flag"
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

func TestGetAuthenticatedUserOAuthScopes(t *testing.T) {
	provider := createTestProvider(t)
	// We expect the GitLab token stored in 1Password under gitlab@sourcegraph.com
	token := os.Getenv("GITLAB_TOKEN")
	client := provider.GetOAuthClient(token)
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
	provider := NewClientProvider(baseURL, doer)
	return provider
}

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}
