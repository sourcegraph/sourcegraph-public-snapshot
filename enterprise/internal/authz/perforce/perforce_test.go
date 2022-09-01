package perforce

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestProvider_FetchAccount(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	user := &types.User{
		ID:       1,
		Username: "alice",
	}

	execer := p4ExecFunc(func(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
		data := `
alice <alice@example.com> (Alice) accessed 2020/12/04
cindy <cindy@example.com> (Cindy) accessed 2020/12/04
`
		return io.NopCloser(strings.NewReader(data)), nil, nil
	})

	t.Run("no matching account", func(t *testing.T) {
		p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", execer)
		got, err := p.FetchAccount(ctx, user, nil, []string{"bob@example.com"})
		if err != nil {
			t.Fatal(err)
		}

		if got != nil {
			t.Fatalf("Want nil but got %v", got)
		}
	})

	t.Run("found matching account", func(t *testing.T) {
		p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", execer)
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
				Data: extsvc.NewUnencryptedData(accountData),
			},
		}
		if diff := cmp.Diff(want, got, et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want got):\n%s", diff)
		}
	})
}

func TestProvider_FetchUserPerms(t *testing.T) {
	ctx := context.Background()
	db := database.NewMockDB()

	t.Run("nil account", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", nil, db)
		_, err := p.FetchUserPerms(ctx, nil, authz.FetchPermsOptions{})
		want := "no account provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the account", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, db)
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
				},
			},
			authz.FetchPermsOptions{},
		)
		want := `not a code host of the account: want "https://gitlab.com/" but have "ssl:111.222.333.444:1666"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("no user found in account data", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, db)
		_, err := p.FetchUserPerms(ctx,
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypePerforce,
					ServiceID:   "ssl:111.222.333.444:1666",
				},
				AccountData: extsvc.AccountData{},
			},
			authz.FetchPermsOptions{},
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
review user alice * //Sourcegraph/*/Handbook/...
review user alice * //Sourcegraph/.../Handbook/...
`,
			wantPerms: &authz.ExternalUserPermissions{
				IncludeContains: []extsvc.RepoID{
					"//Sourcegraph/Engineering/%",
					"//Sourcegraph/Engineering/Backend/%",
					"//Sourcegraph/Engineering/Frontend/%",
					"//Sourcegraph/Handbook/%",
					"//Sourcegraph/[^/]+/Handbook/%",
					"//Sourcegraph/%/Handbook/%",
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
review user alice * -//Sourcegraph/*/Handbook/...
review user alice * -//Sourcegraph/.../Handbook/...
`,
			wantPerms: &authz.ExternalUserPermissions{
				ExcludeContains: []extsvc.RepoID{
					"//Sourcegraph/[^/]+/Handbook/%",
					"//Sourcegraph/%/Handbook/%",
				},
			},
		},
		{
			name: "include and exclude",
			response: `
read user alice * //Sourcegraph/Security/...
read user alice * //Sourcegraph/Engineering/...
owner user alice * //Sourcegraph/Engineering/Backend/...
open user alice * //Sourcegraph/Engineering/Frontend/...
review user alice * //Sourcegraph/Handbook/...
open user alice * //Sourcegraph/Engineering/.../Frontend/...
open user alice * //Sourcegraph/.../Handbook/...  ## wildcard A

list user alice * -//Sourcegraph/Security/...                        ## "list" can revoke read access
=read user alice * -//Sourcegraph/Engineering/Frontend/...           ## exact match of a previous include
open user alice * -//Sourcegraph/Engineering/Backend/Credentials/... ## sub-match of a previous include
open user alice * -//Sourcegraph/Engineering/*/Frontend/Folder/...   ## sub-match of a previous include
open user alice * -//Sourcegraph/*/Handbook/...                      ## sub-match of wildcard A include
`,
			wantPerms: &authz.ExternalUserPermissions{
				IncludeContains: []extsvc.RepoID{
					"//Sourcegraph/Engineering/%",
					"//Sourcegraph/Engineering/Backend/%",
					"//Sourcegraph/Engineering/Frontend/%",
					"//Sourcegraph/Handbook/%",
					"//Sourcegraph/Engineering/%/Frontend/%",
					"//Sourcegraph/%/Handbook/%",
				},
				ExcludeContains: []extsvc.RepoID{
					"//Sourcegraph/Engineering/Frontend/%",
					"//Sourcegraph/Engineering/Backend/Credentials/%",
					"//Sourcegraph/Engineering/[^/]+/Frontend/Folder/%",
					"//Sourcegraph/[^/]+/Handbook/%",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logtest.Scoped(t)
			execer := p4ExecFunc(func(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
				return io.NopCloser(strings.NewReader(test.response)), nil, nil
			})

			p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", execer)
			got, err := p.FetchUserPerms(ctx,
				&extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
					AccountData: extsvc.AccountData{
						Data: extsvc.NewUnencryptedData(accountData),
					},
				},
				authz.FetchPermsOptions{},
			)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantPerms, got); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}

	// Specific behaviour is tested in TestScanFullRepoPermissions
	t.Run("SubRepoPermissions", func(t *testing.T) {
		logger := logtest.Scoped(t)
		execer := p4ExecFunc(func(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
			return io.NopCloser(strings.NewReader(`
read user alice * //Sourcegraph/Engineering/...
read user alice * -//Sourcegraph/Security/...
`)), nil, nil
		})
		p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", execer)
		p.depots = append(p.depots, "//Sourcegraph/")

		got, err := p.FetchUserPerms(ctx,
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypePerforce,
					ServiceID:   "ssl:111.222.333.444:1666",
				},
				AccountData: extsvc.AccountData{
					Data: extsvc.NewUnencryptedData(accountData),
				},
			},
			authz.FetchPermsOptions{},
		)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(&authz.ExternalUserPermissions{
			Exacts: []extsvc.RepoID{"//Sourcegraph/"},
			SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissions{
				"//Sourcegraph/": {
					PathIncludes: []string{
						mustGlobPattern(t, "Engineering/..."),
					},
					PathExcludes: []string{
						mustGlobPattern(t, "Security/..."),
					},
				},
			},
		}, got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewMockDB()

	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, db)
		_, err := p.FetchRepoPerms(ctx, nil, authz.FetchPermsOptions{})
		want := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := NewProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, db)
		_, err := p.FetchRepoPerms(ctx,
			&extsvc.Repository{
				URI: "gitlab.com/user/repo",
				ExternalRepoSpec: api.ExternalRepoSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
				},
			},
			authz.FetchPermsOptions{},
		)
		want := `not a code host of the repository: want "https://gitlab.com/" but have "ssl:111.222.333.444:1666"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})
	execer := p4ExecFunc(func(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
		var data string

		switch args[0] {

		case "protects":
			data = `
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
			data = `
alice <alice@example.com> (Alice) accessed 2020/12/04
bob <bob@example.com> (Bob) accessed 2020/12/04
cindy <cindy@example.com> (Cindy) accessed 2020/12/04
david <david@example.com> (David) accessed 2020/12/04
frank <frank@example.com> (Frank) accessed 2020/12/04
`
		case "group":
			switch args[2] {
			case "Backend":
				data = `
Users:
	alice
	cindy
`
			case "Frontend":
				data = `
Users:
	bob
	david
	frank
`
			}
		}

		return io.NopCloser(strings.NewReader(data)), nil, nil
	})

	p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "admin", "password", execer)
	got, err := p.FetchRepoPerms(ctx,
		&extsvc.Repository{
			URI: "gitlab.com/user/repo",
			ExternalRepoSpec: api.ExternalRepoSpec{
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		},
		authz.FetchPermsOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []extsvc.AccountID{"alice@example.com"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func NewTestProvider(logger log.Logger, urn, host, user, password string, execer p4Execer) *Provider {
	p := NewProvider(logger, urn, host, user, password, []extsvc.RepoID{}, database.NewMockDB())
	p.p4Execer = execer
	return p
}

type p4ExecFunc func(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error)

func (p p4ExecFunc) P4Exec(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
	return p(ctx, host, user, password, args...)
}
