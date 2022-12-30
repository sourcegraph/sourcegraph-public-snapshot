package bitbucketcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"golang.org/x/oauth2"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func mustURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func createTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/repositories") {
			json.NewEncoder(w).Encode(struct {
				Values []bitbucket.Repository `json:"values"`
			}{
				Values: []bitbucket.Repository{
					{Uuid: "1"},
					{Uuid: "2"},
					{Uuid: "3"},
				},
			})
			return
		}
	}))
}

func TestProvider_FetchUserPerms(t *testing.T) {
	t.Run("nil account", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://bitbucket.org"), "", nil)
		_, err := p.FetchUserPerms(context.Background(), nil, authz.FetchPermsOptions{})
		want := "no account provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the account", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://bitbucket.org"), "", nil)
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
			},
			authz.FetchPermsOptions{},
		)
		want := `not a code host of the account: want "https://bitbucket.org/" but have "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("no account data provided", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://bitbucket.org"), "", nil)
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeBitbucketCloud,
					ServiceID:   "https://bitbucket.org/",
				},
			},
			authz.FetchPermsOptions{},
		)
		want := `no account data provided`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	server := createTestServer()
	defer server.Close()

	t.Run("fetch user permissions", func(t *testing.T) {
		p := NewProvider(mustURL(t, server.URL), "", nil)

		var acctData extsvc.AccountData
		err := bitbucketcloud.SetExternalAccountData(&acctData, &bitbucket.User{}, &oauth2.Token{AccessToken: "my-access-token"})
		if err != nil {
			t.Fatal(err)
		}

		account := &extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: extsvc.TypeBitbucketCloud,
				ServiceID:   extsvc.NormalizeBaseURL(mustURL(t, server.URL)).String(),
			},
			AccountData: acctData,
		}
		userPerms, err := p.FetchUserPerms(context.Background(), account, authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		expRepoIDs := []extsvc.RepoID{"1", "2", "3"}
		if diff := cmp.Diff(expRepoIDs, userPerms.Exacts); diff != "" {
			t.Fatal(diff)
		}
	})
}
