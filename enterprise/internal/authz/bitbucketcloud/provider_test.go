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

	// The OAuthProvider uses the gitlab.Client under the hood,
	// which uses rcache, a caching layer that uses Redis.
	// We need to clear the cache before we run the tests
	//	rcache.SetupForTest(t)
	//
	//	p := newOAuthProvider(
	//		OAuthProviderOp{
	//			BaseURL: mustURL(t, "https://gitlab.com"),
	//			Token:   "admin_token",
	//		},
	//		&mockDoer{
	//			do: func(r *http.Request) (*http.Response, error) {
	//				visibility := r.URL.Query().Get("visibility")
	//				if visibility != "private" && visibility != "internal" {
	//					return nil, errors.Errorf("URL visibility: want private or internal, got %s", visibility)
	//				}
	//				want := fmt.Sprintf("https://gitlab.com/api/v4/projects?min_access_level=20&per_page=100&visibility=%s", visibility)
	//				if r.URL.String() != want {
	//					return nil, errors.Errorf("URL: want %q but got %q", want, r.URL)
	//				}
	//
	//				want = "Bearer my_access_token"
	//				got := r.Header.Get("Authorization")
	//				if got != want {
	//					return nil, errors.Errorf("HTTP Authorization: want %q but got %q", want, got)
	//				}
	//
	//				body := `[{"id": 1}, {"id": 2}]`
	//				if visibility == "internal" {
	//					body = `[{"id": 3}]`
	//				}
	//				return &http.Response{
	//					Status:     http.StatusText(http.StatusOK),
	//					StatusCode: http.StatusOK,
	//					Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	//				}, nil
	//			},
	//		},
	//	)
	//
	//	gitlab.MockGetOAuthContext = func() *oauthutil.OAuthContext {
	//		return &oauthutil.OAuthContext{
	//			ClientID:     "client",
	//			ClientSecret: "client_sec",
	//			Endpoint: oauth2.Endpoint{
	//				AuthURL:  "url/oauth/authorize",
	//				TokenURL: "url/oauth/token",
	//			},
	//			Scopes: []string{"read_user"},
	//		}
	//	}
	//	defer func() { gitlab.MockGetOAuthContext = nil }()
	//
	//	authData := json.RawMessage(`{"access_token": "my_access_token"}`)
	//	repoIDs, err := p.FetchUserPerms(context.Background(),
	//		&extsvc.Account{
	//			AccountSpec: extsvc.AccountSpec{
	//				ServiceType: extsvc.TypeGitLab,
	//				ServiceID:   "https://gitlab.com/",
	//			},
	//			AccountData: extsvc.AccountData{
	//				AuthData: extsvc.NewUnencryptedData(authData),
	//			},
	//		},
	//		authz.FetchPermsOptions{},
	//	)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	expRepoIDs := []extsvc.RepoID{"1", "2", "3"}
	//	if diff := cmp.Diff(expRepoIDs, repoIDs.Exacts); diff != "" {
	//		t.Fatal(diff)
	//	}
}
