package perforce

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestProvider_FetchAccount(t *testing.T) {
	ctx := context.Background()
	user := &types.User{
		ID:       1,
		Username: "alice",
	}

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(`
alice <alice@example.com> (Alice) accessed 2020/12/04
cindy <cindy@example.com> (Cindy) accessed 2020/12/04
`))
			if err != nil {
				t.Fatal(err)
			}
		}),
	)

	// Strip the protocol from the URI while patching the gitserver client's
	// addresses, since the gitserver implementation does not want the protocol in
	// the address.
	gitserver.DefaultClient.Addrs = func() []string {
		return []string{strings.TrimPrefix(server.URL, "http://")}
	}

	t.Run("no matching account", func(t *testing.T) {
		p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
		got, err := p.FetchAccount(ctx, user, nil, []string{"bob@example.com"})
		if err != nil {
			t.Fatal(err)
		}

		if got != nil {
			t.Fatalf("Want nil but got %v", got)
		}
	})

	t.Run("found matching account", func(t *testing.T) {
		p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
		got, err := p.FetchAccount(ctx, user, nil, []string{"alice@example.com"})
		if err != nil {
			t.Fatal(err)
		}

		accountData, err := jsoniter.Marshal(
			perforce.AccountData{
				Username: "alice",
				Email:    "alice@example.com",
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		want := &extsvc.Account{
			UserID: user.ID,
			AccountSpec: extsvc.AccountSpec{
				ServiceType: p.codeHost.ServiceType,
				ServiceID:   p.codeHost.ServiceID,
				AccountID:   "alice@example.com",
			},
			AccountData: extsvc.AccountData{
				Data: (*json.RawMessage)(&accountData),
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("Mismatch (-want got):\n%s", diff)
		}
	})
}

func TestProvider_FetchUserPerms(t *testing.T) {
	ctx := context.Background()

	t.Run("nil account", func(t *testing.T) {
		p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
		_, err := p.FetchUserPerms(ctx, nil)
		want := "no account provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the account", func(t *testing.T) {
		p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
				},
			},
		)
		want := `not a code host of the account: want "https://gitlab.com/" but have "ssl:111.222.333.444:1666"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("no user found in account data", func(t *testing.T) {
		p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
		_, err := p.FetchUserPerms(ctx,
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypePerforce,
					ServiceID:   "ssl:111.222.333.444:1666",
				},
				AccountData: extsvc.AccountData{},
			},
		)
		want := `no user found in the external account data`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	accountData, err := jsoniter.Marshal(
		perforce.AccountData{
			Username: "alice",
			Email:    "alice@example.com",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var mockResponse string
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(mockResponse))
			if err != nil {
				t.Fatal(err)
			}
		}),
	)

	// Strip the protocol from the URI while patching the gitserver client's
	// addresses, since the gitserver implementation does not want the protocol in
	// the address.
	gitserver.DefaultClient.Addrs = func() []string {
		return []string{strings.TrimPrefix(server.URL, "http://")}
	}

	tests := []struct {
		name      string
		response  string
		wantPerms *authz.ExternalUserPermissions
	}{
		{
			name: "include only",
			response: `
list user alice * //Sourcegraph/Security/... ## "list" can't grant read access
read user alice * //Sourcegraph/Engineering/...
owner user alice * //Sourcegraph/Engineering/Backend/...
open user alice * //Sourcegraph/Engineering/Frontend/...
review user alice * //Sourcegraph/Handbook/...
`,
			wantPerms: &authz.ExternalUserPermissions{
				IncludePrefixes: []extsvc.RepoID{
					"//Sourcegraph/Engineering/",
					"//Sourcegraph/Engineering/Backend/",
					"//Sourcegraph/Engineering/Frontend/",
					"//Sourcegraph/Handbook/",
				},
			},
		},
		{
			name: "exclude only",
			response: `
list user alice * -//Sourcegraph/Security/...
read user alice * -//Sourcegraph/Engineering/...
owner user alice * -//Sourcegraph/Engineering/Backend/...
open user alice * -//Sourcegraph/Engineering/Frontend/...
review user alice * -//Sourcegraph/Handbook/...
`,
			wantPerms: &authz.ExternalUserPermissions{},
		},
		{
			name: "include and exclude",
			response: `
read user alice * //Sourcegraph/Security/...
read user alice * //Sourcegraph/Engineering/...
owner user alice * //Sourcegraph/Engineering/Backend/...
open user alice * //Sourcegraph/Engineering/Frontend/...
review user alice * //Sourcegraph/Handbook/...

list user alice * -//Sourcegraph/Security/...                        ## "list" can revoke read access
=read user alice * -//Sourcegraph/Engineering/Frontend/...           ## exact match of a previous include
open user alice * -//Sourcegraph/Engineering/Backend/Credentials/... ## sub-match of a previous include
`,
			wantPerms: &authz.ExternalUserPermissions{
				IncludePrefixes: []extsvc.RepoID{
					"//Sourcegraph/Engineering/",
					"//Sourcegraph/Engineering/Backend/",
					"//Sourcegraph/Engineering/Frontend/",
					"//Sourcegraph/Handbook/",
				},
				ExcludePrefixes: []extsvc.RepoID{
					"//Sourcegraph/Engineering/Frontend/",
					"//Sourcegraph/Engineering/Backend/Credentials/",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockResponse = test.response

			p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
			got, err := p.FetchUserPerms(ctx,
				&extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
					AccountData: extsvc.AccountData{
						Data: (*json.RawMessage)(&accountData),
					},
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantPerms, got); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	ctx := context.Background()

	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
		_, err := p.FetchRepoPerms(ctx, nil)
		want := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
		_, err := p.FetchRepoPerms(ctx,
			&extsvc.Repository{
				URI: "gitlab.com/user/repo",
				ExternalRepoSpec: api.ExternalRepoSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
				},
			},
		)
		want := `not a code host of the repository: want "https://gitlab.com/" but have "ssl:111.222.333.444:1666"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req protocol.P4ExecRequest
			err := jsoniter.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				t.Fatal(err)
			}

			var response string
			switch req.Args[0] {
			case "protects":
				response = `
## The actual depot prefix does not matter, the "-" sign does
list user * * -//...
write user alice * //Sourcegraph/...
write user bob * //Sourcegraph/...
admin group Backend * //Sourcegraph/...   ## includes "alice" and "cindy"

admin group Frontend * -//Sourcegraph/... ## excludes "bob", "david" and "frank"
read user cindy * -//Sourcegraph/...

list user david * //Sourcegraph/...       ## "list" can't grant read access
`
			case "users":
				response = `
alice <alice@example.com> (Alice) accessed 2020/12/04
bob <bob@example.com> (Bob) accessed 2020/12/04
cindy <cindy@example.com> (Cindy) accessed 2020/12/04
david <david@example.com> (David) accessed 2020/12/04
frank <frank@example.com> (Frank) accessed 2020/12/04
`
			case "group":
				switch req.Args[2] {
				case "Backend":
					response = `
Users:
	alice
	cindy
`
				case "Frontend":
					response = `
Users:
	bob
	david
	frank
`
				}
			}

			_, err = w.Write([]byte(response))
			if err != nil {
				t.Fatal(err)
			}
		}),
	)

	// Strip the protocol from the URI while patching the gitserver client's
	// addresses, since the gitserver implementation does not want the protocol in
	// the address.
	gitserver.DefaultClient.Addrs = func() []string {
		return []string{strings.TrimPrefix(server.URL, "http://")}
	}

	p := NewProvider("", "ssl:111.222.333.444:1666", "admin", "password")
	got, err := p.FetchRepoPerms(ctx,
		&extsvc.Repository{
			URI: "gitlab.com/user/repo",
			ExternalRepoSpec: api.ExternalRepoSpec{
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []extsvc.AccountID{"alice@example.com"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}
