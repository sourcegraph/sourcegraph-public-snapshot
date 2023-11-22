package perforce

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestProvider_FetchAccount(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	user := &types.User{
		ID:       1,
		Username: "alice",
	}

	gitserverClient := gitserver.NewStrictMockClient()
	gitserverClient.PerforceUsersFunc.SetDefaultReturn([]*p4types.User{
		{Username: "alice", Email: "alice@example.com"},
		{Username: "cindy", Email: "cindy@example.com"},
	}, nil)

	t.Run("no matching account", func(t *testing.T) {
		p := NewProvider(logger, gitserverClient, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
		got, err := p.FetchAccount(ctx, user, nil, []string{"bob@example.com"})
		if err != nil {
			t.Fatal(err)
		}

		if got != nil {
			t.Fatalf("Want nil but got %v", got)
		}
	})

	t.Run("found matching account", func(t *testing.T) {
		p := NewProvider(logger, gitserverClient, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
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

	t.Run("nil account", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, gitserver.NewTestClient(t), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
		_, err := p.FetchUserPerms(ctx, nil, authz.FetchPermsOptions{})
		want := "no account provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the account", func(t *testing.T) {
		logger := logtest.Scoped(t)
		p := NewProvider(logger, gitserver.NewTestClient(t), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
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
		p := NewProvider(logger, gitserver.NewTestClient(t), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
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
		protects  []*p4types.Protect
		wantPerms *authz.ExternalUserPermissions
	}{
		{
			name: "include only",
			protects: testParseP4ProtectsRaw(t, strings.NewReader(`
list user alice * //Sourcegraph/Security/... ## "list" can't grant read access
read user alice * //Sourcegraph/Engineering/...
owner user alice * //Sourcegraph/Engineering/Backend/...
open user alice * //Sourcegraph/Engineering/Frontend/...
review user alice * //Sourcegraph/Handbook/...
review user alice * //Sourcegraph/*/Handbook/...
review user alice * //Sourcegraph/.../Handbook/...
`)),
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
			protects: testParseP4ProtectsRaw(t, strings.NewReader(`
list user alice * -//Sourcegraph/Security/...
read user alice * -//Sourcegraph/Engineering/...
owner user alice * -//Sourcegraph/Engineering/Backend/...
open user alice * -//Sourcegraph/Engineering/Frontend/...
review user alice * -//Sourcegraph/Handbook/...
review user alice * -//Sourcegraph/*/Handbook/...
review user alice * -//Sourcegraph/.../Handbook/...
`)), wantPerms: &authz.ExternalUserPermissions{
				ExcludeContains: []extsvc.RepoID{
					"//Sourcegraph/[^/]+/Handbook/%",
					"//Sourcegraph/%/Handbook/%",
				},
			},
		},
		{
			name: "include and exclude",
			protects: testParseP4ProtectsRaw(t, strings.NewReader(`
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
`)),
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
		{
			name: "include and exclude, then include again",
			protects: testParseP4ProtectsRaw(t, strings.NewReader(`
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

read user alice * //Sourcegraph/Security/... 						 ## give access to alice again after revoking
`)),
			wantPerms: &authz.ExternalUserPermissions{
				IncludeContains: []extsvc.RepoID{
					"//Sourcegraph/Engineering/%",
					"//Sourcegraph/Engineering/Backend/%",
					"//Sourcegraph/Engineering/Frontend/%",
					"//Sourcegraph/Handbook/%",
					"//Sourcegraph/Engineering/%/Frontend/%",
					"//Sourcegraph/%/Handbook/%",
					"//Sourcegraph/Security/%",
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
			gc := gitserver.NewStrictMockClient()
			gc.PerforceProtectsForUserFunc.SetDefaultReturn(test.protects, nil)

			p := NewProvider(logger, gc, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
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

		gitserverClient := gitserver.NewStrictMockClient()
		ps := []*p4types.Protect{
			{Level: "read", EntityType: "user", EntityName: "alice", Host: "*", Match: "//Sourcegraph/Engineering/..."},
			{Level: "read", EntityType: "user", EntityName: "alice", Host: "*", Match: "//Sourcegraph/Security/...", IsExclusion: true},
		}
		gitserverClient.PerforceProtectsForDepotFunc.SetDefaultReturn(ps, nil)
		gitserverClient.PerforceProtectsForUserFunc.SetDefaultReturn(ps, nil)

		p := NewProvider(logger, gitserverClient, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
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
					Paths: []string{
						mustGlobPattern(t, "/Engineering/..."),
						mustGlobPattern(t, "-/Security/..."),
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

	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider(logger, gitserver.NewTestClient(t), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
		_, err := p.FetchRepoPerms(ctx, nil, authz.FetchPermsOptions{})
		want := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := NewProvider(logger, gitserver.NewTestClient(t), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
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

	gitserverClient := gitserver.NewStrictMockClient()
	gitserverClient.PerforceUsersFunc.SetDefaultReturn([]*p4types.User{
		{Username: "alice", Email: "alice@example.com"},
		{Username: "bob", Email: "bob@example.com"},
		{Username: "cindy", Email: "cindy@example.com"},
		{Username: "david", Email: "david@example.com"},
		{Username: "frank", Email: "frank@example.com"},
	}, nil)
	gitserverClient.PerforceGroupMembersFunc.SetDefaultHook(func(ctx context.Context, conn protocol.PerforceConnectionDetails, group string) ([]string, error) {
		switch group {
		case "Backend":
			return []string{"alice", "cindy"}, nil
		case "Frontend":
			return []string{"bob", "david", "frank"}, nil
		default:
			return nil, errors.New("invalid group")
		}
	})
	gitserverClient.PerforceProtectsForDepotFunc.SetDefaultReturn([]*p4types.Protect{
		{Level: "list", EntityType: "user", EntityName: "*", Host: "*", Match: "//...", IsExclusion: true},
		{Level: "write", EntityType: "user", EntityName: "alice", Host: "*", Match: "//Sourcegraph/..."},
		{Level: "write", EntityType: "user", EntityName: "bob", Host: "*", Match: "//Sourcegraph/..."},
		{Level: "admin", EntityType: "group", EntityName: "Backend", Host: "*", Match: "//Sourcegraph/..."},                     // includes "alice" and "cindy"
		{Level: "admin", EntityType: "group", EntityName: "Frontend", Host: "*", Match: "//Sourcegraph/...", IsExclusion: true}, // excludes "bob", "david" and "frank"
		{Level: "read", EntityType: "user", EntityName: "cindy", Host: "*", Match: "//Sourcegraph/...", IsExclusion: true},
		{Level: "list", EntityType: "user", EntityName: "david", Host: "*", Match: "//Sourcegraph/..."}, // "list" can't grant read access
	}, nil)

	p := NewProvider(logger, gitserverClient, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
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
