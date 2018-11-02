package gitlab

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_GitLab_FetchAccount(t *testing.T) {
	tests := []GitLab_FetchAccount_Test{
		{
			description: "1 authn provider, basic authz provider",
			authnProviders: []auth.Provider{
				mockAuthnProvider{
					configID:  auth.ProviderConfigID{ID: "okta.mine", Type: "saml"},
					serviceID: "https://okta.mine/",
				},
			},
			op: GitLabAuthzProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				AuthnConfigID:     auth.ProviderConfigID{ID: "okta.mine", Type: "saml"},
				GitLabProvider:    "okta.mine",
				UseNativeUsername: false,
			},
			calls: []GitLab_FetchAccount_Test_call{
				{
					description: "1 account, matches",
					user:        &types.User{ID: 123},
					current:     []*extsvc.ExternalAccount{acct(1, "saml", "https://okta.mine/", "bl")},
					expMine:     acct(123, gitlab.GitLabServiceType, "https://gitlab.mine/", "101"),
				},
				{
					description: "many accounts, none match",
					user:        &types.User{ID: 123},
					current: []*extsvc.ExternalAccount{
						acct(1, "saml", "https://okta.mine/", "nomatch"),
						acct(1, "saml", "nomatch", "bl"),
						acct(1, "nomatch", "https://okta.mine/", "bl"),
					},
					expMine: nil,
				},
				{
					description: "many accounts, 1 match",
					user:        &types.User{ID: 123},
					current: []*extsvc.ExternalAccount{
						acct(1, "saml", "nomatch", "bl"),
						acct(1, "nomatch", "https://okta.mine/", "bl"),
						acct(1, "saml", "https://okta.mine/", "bl"),
					},
					expMine: acct(123, gitlab.GitLabServiceType, "https://gitlab.mine/", "101"),
				},
				{
					description: "no user",
					user:        nil,
					current:     nil,
					expMine:     nil,
				},
			},
		},
		{
			description:    "0 authn providers, native username",
			authnProviders: nil,
			op: GitLabAuthzProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				UseNativeUsername: true,
			},
			calls: []GitLab_FetchAccount_Test_call{
				{
					description: "username match",
					user:        &types.User{ID: 123, Username: "b.l"},
					expMine:     acct(123, gitlab.GitLabServiceType, "https://gitlab.mine/", "101"),
				},
				{
					description: "no username match",
					user:        &types.User{ID: 123, Username: "nomatch"},
					expMine:     nil,
				},
			},
		},
		{
			description:    "0 authn providers, basic authz provider",
			authnProviders: nil,
			op: GitLabAuthzProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				AuthnConfigID:     auth.ProviderConfigID{ID: "okta.mine", Type: "saml"},
				GitLabProvider:    "okta.mine",
				UseNativeUsername: false,
			},
			calls: []GitLab_FetchAccount_Test_call{
				{
					description: "no matches",
					user:        &types.User{ID: 123, Username: "b.l"},
					expMine:     nil,
				},
			},
		},
		{
			description: "2 authn providers, basic authz provider",
			authnProviders: []auth.Provider{
				mockAuthnProvider{
					configID:  auth.ProviderConfigID{ID: "okta.mine", Type: "saml"},
					serviceID: "https://okta.mine/",
				},
				mockAuthnProvider{
					configID:  auth.ProviderConfigID{ID: "onelogin.mine", Type: "openidconnect"},
					serviceID: "https://onelogin.mine/",
				},
			},
			op: GitLabAuthzProviderOp{
				BaseURL:           mustURL(t, "https://gitlab.mine"),
				AuthnConfigID:     auth.ProviderConfigID{ID: "onelogin.mine", Type: "openidconnect"},
				GitLabProvider:    "onelogin.mine",
				UseNativeUsername: false,
			},
			calls: []GitLab_FetchAccount_Test_call{
				{
					description: "1 authn provider matches",
					user:        &types.User{ID: 123},
					current:     []*extsvc.ExternalAccount{acct(1, "openidconnect", "https://onelogin.mine/", "bl")},
					expMine:     acct(123, gitlab.GitLabServiceType, "https://gitlab.mine/", "101"),
				},
				{
					description: "0 authn providers match",
					user:        &types.User{ID: 123},
					current:     []*extsvc.ExternalAccount{acct(1, "openidconnect", "https://onelogin.mine/", "nomatch")},
					expMine:     nil,
				},
			},
		},
	}

	gitlabMock := mockGitLab{
		t:          t,
		maxPerPage: 1,
		users: []*gitlab.User{
			{
				ID:       101,
				Username: "b.l",
				Identities: []gitlab.Identity{
					{Provider: "okta.mine", ExternUID: "bl"},
					{Provider: "onelogin.mine", ExternUID: "bl"},
				},
			},
			{
				ID:         102,
				Username:   "k.l",
				Identities: []gitlab.Identity{{Provider: "okta.mine", ExternUID: "kl"}},
			},
			{
				ID:         199,
				Username:   "user-without-extern-id",
				Identities: nil,
			},
		},
	}
	gitlab.MockListUsers = gitlabMock.ListUsers

	for _, test := range tests {
		test.run(t)
	}
}

type GitLab_FetchAccount_Test struct {
	description string

	authnProviders []auth.Provider
	op             GitLabAuthzProviderOp

	calls []GitLab_FetchAccount_Test_call
}

type GitLab_FetchAccount_Test_call struct {
	description string

	user    *types.User
	current []*extsvc.ExternalAccount

	expMine *extsvc.ExternalAccount
}

func (g GitLab_FetchAccount_Test) run(t *testing.T) {
	t.Logf("Test case %q", g.description)

	provs := make(map[auth.Provider]bool)
	for _, p := range g.authnProviders {
		provs[p] = true
	}
	auth.UpdateProviders(provs)

	ctx := context.Background()
	authzProvider := NewProvider(g.op)
	for _, c := range g.calls {
		t.Logf("Call %q", c.description)
		acct, err := authzProvider.FetchAccount(ctx, c.user, c.current)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if acct != nil {
			// ignore these fields for comparison
			acct.AuthData = nil
			acct.AccountData = nil
		}
		if !reflect.DeepEqual(acct, c.expMine) {
			t.Errorf("expected %+v, but got %+v", c.expMine, acct)
		}
	}
}

func Test_GitLab_RepoPerms(t *testing.T) {
	gitlabMock := newMockGitLab(t,
		[]int{
			11, // gitlab.mine/bl/repo-1
			12, // gitlab.mine/bl/repo-2
			13, // gitlab.mine/bl/repo-3
			21, // gitlab.mine/kl/repo-1
			22, // gitlab.mine/kl/repo-2
			23, // gitlab.mine/kl/repo-3
			31, // gitlab.mine/org/repo-1
			32, // gitlab.mine/org/repo-2
			33, // gitlab.mine/org/repo-3
			41, // gitlab.mine/public/repo-1
		},
		map[string][]int{
			"101":    []int{11, 12, 13, 31, 32, 33, 41},
			"201":    []int{21, 22, 23, 31, 32, 33, 41},
			"PUBLIC": []int{41},
		}, 1)
	gitlab.MockListProjects = gitlabMock.ListProjects

	tests := []GitLab_RepoPerms_Test{
		{
			description: "without match pattern",
			op: GitLabAuthzProviderOp{
				BaseURL: mustURL(t, "https://gitlab.mine"),
			},
			calls: []GitLab_RepoPerms_call{
				{
					description: "bl user has expected perms",
					account:     acct(1, "gitlab", "https://gitlab.mine/", "101"),
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "11"):                 struct{}{},
						repo("bl/repo-2", gitlab.GitLabServiceType, "other", "12"):                                struct{}{},
						repo("gitlab.mine/bl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "999"):    struct{}{},
						repo("kl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "21"):                 struct{}{},
						repo("kl/repo-2", gitlab.GitLabServiceType, "other", "22"):                                struct{}{},
						repo("gitlab.mine/kl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "998"):    struct{}{},
						repo("org/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "31"):                struct{}{},
						repo("org/repo-2", gitlab.GitLabServiceType, "other", "32"):                               struct{}{},
						repo("gitlab.mine/org/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "997"):   struct{}{},
						repo("gitlab.mine/public/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "41"): struct{}{},
					},
					expPerms: map[api.RepoURI]map[authz.Perm]bool{
						"bl/repo-1":                 map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/bl/repo-3":     map[authz.Perm]bool{},
						"kl/repo-1":                 map[authz.Perm]bool{},
						"gitlab.mine/kl/repo-3":     map[authz.Perm]bool{},
						"org/repo-1":                map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/org/repo-3":    map[authz.Perm]bool{},
						"gitlab.mine/public/repo-1": map[authz.Perm]bool{authz.Read: true},
					},
				},
				{
					description: "kl user has expected perms",
					account:     acct(2, "gitlab", "https://gitlab.mine/", "201"),
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "11"):                 struct{}{},
						repo("bl/repo-2", gitlab.GitLabServiceType, "other", "12"):                                struct{}{},
						repo("gitlab.mine/bl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "999"):    struct{}{},
						repo("kl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "21"):                 struct{}{},
						repo("kl/repo-2", gitlab.GitLabServiceType, "other", "22"):                                struct{}{},
						repo("gitlab.mine/kl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "998"):    struct{}{},
						repo("org/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "31"):                struct{}{},
						repo("org/repo-2", gitlab.GitLabServiceType, "other", "32"):                               struct{}{},
						repo("gitlab.mine/org/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "997"):   struct{}{},
						repo("gitlab.mine/public/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "41"): struct{}{},
					},
					expPerms: map[api.RepoURI]map[authz.Perm]bool{
						"bl/repo-1":                 map[authz.Perm]bool{},
						"gitlab.mine/bl/repo-3":     map[authz.Perm]bool{},
						"kl/repo-1":                 map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/kl/repo-3":     map[authz.Perm]bool{},
						"org/repo-1":                map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/org/repo-3":    map[authz.Perm]bool{},
						"gitlab.mine/public/repo-1": map[authz.Perm]bool{authz.Read: true},
					},
				},
				{
					description: "unknown user has no perms",
					account:     acct(3, "gitlab", "https://gitlab.mine/", "999"),
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "11"):                 struct{}{},
						repo("bl/repo-2", gitlab.GitLabServiceType, "other", "12"):                                struct{}{},
						repo("gitlab.mine/bl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "999"):    struct{}{},
						repo("kl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "21"):                 struct{}{},
						repo("kl/repo-2", gitlab.GitLabServiceType, "other", "22"):                                struct{}{},
						repo("gitlab.mine/kl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "998"):    struct{}{},
						repo("org/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "31"):                struct{}{},
						repo("org/repo-2", gitlab.GitLabServiceType, "other", "32"):                               struct{}{},
						repo("gitlab.mine/org/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "997"):   struct{}{},
						repo("gitlab.mine/public/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "41"): struct{}{},
					},
					expPerms: map[api.RepoURI]map[authz.Perm]bool{
						"bl/repo-1":                 map[authz.Perm]bool{},
						"gitlab.mine/bl/repo-3":     map[authz.Perm]bool{},
						"kl/repo-1":                 map[authz.Perm]bool{},
						"gitlab.mine/kl/repo-3":     map[authz.Perm]bool{},
						"org/repo-1":                map[authz.Perm]bool{},
						"gitlab.mine/org/repo-3":    map[authz.Perm]bool{},
						"gitlab.mine/public/repo-1": map[authz.Perm]bool{},
					},
				},
				{
					description: "unauthenticated user has access to public only",
					account:     nil,
					repos: map[authz.Repo]struct{}{
						repo("bl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "11"):                 struct{}{},
						repo("bl/repo-2", gitlab.GitLabServiceType, "other", "12"):                                struct{}{},
						repo("gitlab.mine/bl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "999"):    struct{}{},
						repo("kl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "21"):                 struct{}{},
						repo("kl/repo-2", gitlab.GitLabServiceType, "other", "22"):                                struct{}{},
						repo("gitlab.mine/kl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "998"):    struct{}{},
						repo("org/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "31"):                struct{}{},
						repo("org/repo-2", gitlab.GitLabServiceType, "other", "32"):                               struct{}{},
						repo("gitlab.mine/org/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "997"):   struct{}{},
						repo("gitlab.mine/public/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "41"): struct{}{},
					},
					expPerms: map[api.RepoURI]map[authz.Perm]bool{
						"bl/repo-1":                 map[authz.Perm]bool{},
						"gitlab.mine/bl/repo-3":     map[authz.Perm]bool{},
						"kl/repo-1":                 map[authz.Perm]bool{},
						"gitlab.mine/kl/repo-3":     map[authz.Perm]bool{},
						"org/repo-1":                map[authz.Perm]bool{},
						"gitlab.mine/org/repo-3":    map[authz.Perm]bool{},
						"gitlab.mine/public/repo-1": map[authz.Perm]bool{authz.Read: true},
					},
				},
			},
		},
		{
			description: "with match pattern",
			op: GitLabAuthzProviderOp{
				BaseURL:      mustURL(t, "https://gitlab.mine"),
				MatchPattern: "gitlab.mine/*",
			},
			calls: []GitLab_RepoPerms_call{
				{
					description: "bl user has expected perms, short input repos list",
					account:     acct(1, gitlab.GitLabServiceType, "https://gitlab.mine/", "101"),
					repos: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "11"):                          struct{}{},
						repo("gitlab.mine/kl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "21"):                          struct{}{},
						repo("gitlab.mine/org/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "31"):                         struct{}{},
						repo("part-of-gitlab.mine-but-doesn't-match-pattern", gitlab.GitLabServiceType, "https://gitlab.mine/", "999"): struct{}{},
					},
					expPerms: map[api.RepoURI]map[authz.Perm]bool{
						"gitlab.mine/bl/repo-1":  map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/kl/repo-1":  map[authz.Perm]bool{},
						"gitlab.mine/org/repo-1": map[authz.Perm]bool{authz.Read: true},
					},
				},
				{
					description: "bl user has expected perms, long input repos list",
					account:     acct(1, gitlab.GitLabServiceType, "https://gitlab.mine/", "101"),
					repos: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "11"):  struct{}{},
						repo("gitlab.mine/bl/repo-2", gitlab.GitLabServiceType, "https://gitlab.mine/", "12"):  struct{}{},
						repo("bl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "13"):              struct{}{},
						repo("gitlab.mine/kl/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "21"):  struct{}{},
						repo("gitlab.mine/kl/repo-2", gitlab.GitLabServiceType, "https://gitlab.mine/", "22"):  struct{}{},
						repo("kl/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "23"):              struct{}{},
						repo("gitlab.mine/org/repo-1", gitlab.GitLabServiceType, "https://gitlab.mine/", "31"): struct{}{},
						repo("gitlab.mine/org/repo-2", gitlab.GitLabServiceType, "https://gitlab.mine/", "32"): struct{}{},
						repo("org/repo-3", gitlab.GitLabServiceType, "https://gitlab.mine/", "33"):             struct{}{},
					},
					expPerms: map[api.RepoURI]map[authz.Perm]bool{
						"gitlab.mine/bl/repo-1":  map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/bl/repo-2":  map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/kl/repo-1":  map[authz.Perm]bool{},
						"gitlab.mine/kl/repo-2":  map[authz.Perm]bool{},
						"gitlab.mine/org/repo-1": map[authz.Perm]bool{authz.Read: true},
						"gitlab.mine/org/repo-2": map[authz.Perm]bool{authz.Read: true},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test.run(t)
	}
}

type GitLab_RepoPerms_Test struct {
	description string

	op GitLabAuthzProviderOp

	calls []GitLab_RepoPerms_call
}

type GitLab_RepoPerms_call struct {
	description string
	account     *extsvc.ExternalAccount
	repos       map[authz.Repo]struct{}
	expPerms    map[api.RepoURI]map[authz.Perm]bool
}

func (g GitLab_RepoPerms_Test) run(t *testing.T) {
	t.Logf("Test case %q", g.description)

	for _, c := range g.calls {
		t.Logf("Call %q", c.description)

		// Recreate the authz provider cache every time, before running twice (once uncached, once cached)
		ctx := context.Background()
		op := g.op
		op.MockCache = make(mockCache)
		authzProvider := NewProvider(op)

		for i := 0; i < 2; i++ {
			t.Logf("iter %d", i)
			perms, err := authzProvider.RepoPerms(ctx, c.account, c.repos)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			if !reflect.DeepEqual(perms, c.expPerms) {
				t.Errorf("expected %s, but got %s", asJSON(t, c.expPerms), asJSON(t, perms))
			}
		}
	}
}

func Test_GitLab_RepoPerms_cache(t *testing.T) {
	gitlabMock := newMockGitLab(t,
		[]int{
			11, // gitlab.mine/bl/repo-1
			12, // gitlab.mine/bl/repo-2
			13, // gitlab.mine/bl/repo-3
			21, // gitlab.mine/kl/repo-1
			22, // gitlab.mine/kl/repo-2
			23, // gitlab.mine/kl/repo-3
			31, // gitlab.mine/org/repo-1
			32, // gitlab.mine/org/repo-2
			33, // gitlab.mine/org/repo-3
		},
		map[string][]int{
			"101": []int{11, 12, 13, 31, 32, 33},
			"201": []int{21, 22, 23, 31, 32, 33},
		}, 1)
	gitlab.MockListProjects = gitlabMock.ListProjects

	ctx := context.Background()
	authzProvider := NewProvider(GitLabAuthzProviderOp{
		BaseURL:       mustURL(t, "https://gitlab.mine"),
		AuthnConfigID: auth.ProviderConfigID{ID: "https://gitlab.mine/", Type: gitlab.GitLabServiceType},
		MockCache:     make(mockCache),
	})
	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.GitLabServiceType, "https://gitlab.mine/", "bl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]int{"projects?per_page=100&sudo=bl": 1}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.GitLabServiceType, "https://gitlab.mine/", "bl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]int{"projects?per_page=100&sudo=bl": 1}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.GitLabServiceType, "https://gitlab.mine/", "kl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]int{"projects?per_page=100&sudo=bl": 1, "projects?per_page=100&sudo=kl": 1}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}

	if _, err := authzProvider.RepoPerms(ctx, acct(1, gitlab.GitLabServiceType, "https://gitlab.mine/", "kl"), nil); err != nil {
		t.Fatal(err)
	}
	if exp := map[string]int{"projects?per_page=100&sudo=bl": 1, "projects?per_page=100&sudo=kl": 1}; !reflect.DeepEqual(gitlabMock.madeProjectReqs, exp) {
		t.Errorf("Unexpected cache behavior. Expected underying requests to be %v, but got %v", exp, gitlabMock.madeProjectReqs)
	}
}

func Test_GitLab_Repos(t *testing.T) {
	tests := []GitLab_Repos_Test{
		{
			description: "with match pattern",
			op: GitLabAuthzProviderOp{
				BaseURL:      mustURL(t, "https://gitlab.mine"),
				MatchPattern: "gitlab.mine/*",
			},
			calls: []GitLab_Repos_call{
				{
					repos: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""):         struct{}{},
						repo("gitlab.mine/kl/repo-1", "", "", ""):         struct{}{},
						repo("another.host/bl/repo-1", "", "", ""):        struct{}{},
						repo("a", "gitlab", "https://gitlab.mine/", "23"): struct{}{},
					},
					expMine: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""): struct{}{},
						repo("gitlab.mine/kl/repo-1", "", "", ""): struct{}{},
					},
					expOthers: map[authz.Repo]struct{}{
						repo("another.host/bl/repo-1", "", "", ""):        struct{}{},
						repo("a", "gitlab", "https://gitlab.mine/", "23"): struct{}{},
					},
				},
			},
		},
		{
			description: "without match pattern",
			op: GitLabAuthzProviderOp{
				BaseURL: mustURL(t, "https://gitlab.mine"),
			},
			calls: []GitLab_Repos_call{
				{
					repos: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""):                         struct{}{},
						repo("gitlab.mine/kl/repo-1", "", "", ""):                         struct{}{},
						repo("another.host/bl/repo-1", "", "", ""):                        struct{}{},
						repo("a", gitlab.GitLabServiceType, "https://gitlab.mine/", "23"): struct{}{},
						repo("b", gitlab.GitLabServiceType, "https://not-mine/", "34"):    struct{}{},
						repo("c", "not-gitlab", "https://gitlab.mine/", "45"):             struct{}{},
					},
					expMine: map[authz.Repo]struct{}{
						repo("a", gitlab.GitLabServiceType, "https://gitlab.mine/", "23"): struct{}{},
					},
					expOthers: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""):                      struct{}{},
						repo("gitlab.mine/kl/repo-1", "", "", ""):                      struct{}{},
						repo("another.host/bl/repo-1", "", "", ""):                     struct{}{},
						repo("b", gitlab.GitLabServiceType, "https://not-mine/", "34"): struct{}{},
						repo("c", "not-gitlab", "https://gitlab.mine/", "45"):          struct{}{},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test.run(t)
	}
}

type GitLab_Repos_Test struct {
	description string
	op          GitLabAuthzProviderOp
	calls       []GitLab_Repos_call
}

type GitLab_Repos_call struct {
	repos     map[authz.Repo]struct{}
	expMine   map[authz.Repo]struct{}
	expOthers map[authz.Repo]struct{}
}

func (g GitLab_Repos_Test) run(t *testing.T) {
	t.Logf("Test case %q", g.description)
	for _, c := range g.calls {
		ctx := context.Background()
		op := g.op
		op.MockCache = make(mockCache)
		authzProvider := NewProvider(op)

		mine, others := authzProvider.Repos(ctx, c.repos)
		if !reflect.DeepEqual(mine, c.expMine) {
			t.Errorf("For input %v, expected mine to be %v, but got %v", c.repos, c.expMine, mine)
		}
		if !reflect.DeepEqual(others, c.expOthers) {
			t.Errorf("For input %v, expected others to be %v, but got %v", c.repos, c.expOthers, others)
		}
	}
}

// mockGitLab is a mock for the GitLab client that can be used by tests. Instantiating a mockGitLab
// instance itself does nothing, but its methods can be used to replace the mock functions (e.g.,
// MockListProjects).
//
// We prefer to do it this way, instead of defining an interface for the GitLab client, because this
// preserves the ability to jump-to-def around the actual implementation.
type mockGitLab struct {
	t *testing.T

	// acls is a map from GitLab user ID to list of accessible project IDs on GitLab
	// Publicly accessible projects are keyed by the special "PUBLIC" key
	acls map[string][]int

	// projs is a map of all projects on the instance, keyed by project ID
	projs map[int]*gitlab.Project

	// users is a list of all users
	users []*gitlab.User

	// maxPerPage returns the max per_page value for the instance
	maxPerPage int

	// madeUserReqs and madeProjectReqs record how many ListUsers and ListProjects requests were
	// made by url string
	madeUserReqs    map[string]int
	madeProjectReqs map[string]int
}

func newMockGitLab(t *testing.T, projIDs []int, acls map[string][]int, maxPerPage int) mockGitLab {
	projs := make(map[int]*gitlab.Project)
	for _, p := range projIDs {
		projs[p] = &gitlab.Project{ProjectCommon: gitlab.ProjectCommon{ID: p}}
	}
	return mockGitLab{
		t:          t,
		projs:      projs,
		acls:       acls,
		maxPerPage: maxPerPage,
	}
}

func (m *mockGitLab) ListUsers(ctx context.Context, urlStr string) (users []*gitlab.User, nextPageURL *string, err error) {
	if m.madeUserReqs == nil {
		m.madeUserReqs = make(map[string]int)
	}
	m.madeUserReqs[urlStr]++

	u, err := url.Parse(urlStr)
	if err != nil {
		m.t.Fatalf("could not parse ListUsers urlStr %q: %s", urlStr, err)
	}

	var matchingUsers []*gitlab.User
	for _, user := range m.users {
		userMatches := true
		if qExternUID := u.Query().Get("extern_uid"); qExternUID != "" {
			qProvider := u.Query().Get("provider")

			match := false
			for _, identity := range user.Identities {
				if identity.ExternUID == qExternUID && identity.Provider == qProvider {
					match = true
					break
				}
			}
			if !match {
				userMatches = false
				break
			}
		}
		if qUsername := u.Query().Get("username"); qUsername != "" {
			if user.Username != qUsername {
				userMatches = false
				break
			}
		}
		if userMatches {
			matchingUsers = append(matchingUsers, user)
		}
	}

	// pagination
	perPage, err := getIntOrDefault(u.Query().Get("per_page"), m.maxPerPage)
	if err != nil {
		return nil, nil, err
	}
	page, err := getIntOrDefault(u.Query().Get("page"), 1)
	if err != nil {
		return nil, nil, err
	}
	p := page - 1
	var (
		pagedUsers []*gitlab.User
	)
	if perPage*p > len(matchingUsers)-1 {
		pagedUsers = nil
	} else if perPage*(p+1) > len(matchingUsers)-1 {
		pagedUsers = matchingUsers[perPage*p:]
	} else {
		pagedUsers = matchingUsers[perPage*p : perPage*(p+1)]
		if perPage*(p+1) <= len(matchingUsers)-1 {
			newU := *u
			q := u.Query()
			q.Set("page", strconv.Itoa(page+1))
			newU.RawQuery = q.Encode()
			s := newU.String()
			nextPageURL = &s
		}
	}
	return pagedUsers, nextPageURL, nil
}

func (m *mockGitLab) ListProjects(ctx context.Context, urlStr string) (proj []*gitlab.Project, nextPageURL *string, err error) {
	if m.madeProjectReqs == nil {
		m.madeProjectReqs = make(map[string]int)
	}
	m.madeProjectReqs[urlStr]++

	u, err := url.Parse(urlStr)
	if err != nil {
		m.t.Fatalf("could not parse ListProjects urlStr %q: %s", urlStr, err)
	}
	var repoIDs []int
	if sudo := u.Query().Get("sudo"); sudo != "" {
		repoIDs = m.acls[sudo]
	} else if visibility := u.Query().Get("visibility"); visibility == "public" {
		repoIDs = m.acls["PUBLIC"]
	} else {
		m.t.Fatalf("mockGitLab unable to handle urlStr %q", urlStr)
	}
	allProjs := make([]*gitlab.Project, len(repoIDs))
	for i, repoID := range repoIDs {
		proj, ok := m.projs[repoID]
		if !ok {
			m.t.Fatalf("Dangling project reference in mockGitLab: %d", repoID)
		}
		allProjs[i] = proj
	}

	// pagination
	perPage, err := getIntOrDefault(u.Query().Get("per_page"), m.maxPerPage)
	if err != nil {
		return nil, nil, err
	}
	if perPage > m.maxPerPage {
		perPage = m.maxPerPage
	}
	page, err := getIntOrDefault(u.Query().Get("page"), 1)
	if err != nil {
		return nil, nil, err
	}
	p := page - 1
	var (
		pagedProjs []*gitlab.Project
	)
	if perPage*p > len(allProjs)-1 {
		pagedProjs = nil
	} else if perPage*(p+1) > len(allProjs)-1 {
		pagedProjs = allProjs[perPage*p:]
	} else {
		pagedProjs = allProjs[perPage*p : perPage*(p+1)]
		if perPage*(p+1) <= len(allProjs)-1 {
			newU := *u
			q := u.Query()
			q.Set("page", strconv.Itoa(page+1))
			newU.RawQuery = q.Encode()
			s := newU.String()
			nextPageURL = &s
		}
	}
	return pagedProjs, nextPageURL, nil
}

type mockCache map[string]string

func (m mockCache) Get(key string) ([]byte, bool) {
	v, ok := m[key]
	return []byte(v), ok
}
func (m mockCache) Set(key string, b []byte) {
	m[key] = string(b)
}
func (m mockCache) Delete(key string) {
	delete(m, key)
}

func getIntOrDefault(str string, def int) (int, error) {
	if str == "" {
		return def, nil
	}
	return strconv.Atoi(str)
}

func acct(userID int32, serviceType, serviceID, accountID string) *extsvc.ExternalAccount {
	return &extsvc.ExternalAccount{
		UserID: userID,
		ExternalAccountSpec: extsvc.ExternalAccountSpec{
			ServiceType: serviceType,
			ServiceID:   serviceID,
			AccountID:   accountID,
		},
	}
}

func repo(uri, serviceType, serviceID, id string) authz.Repo {
	return authz.Repo{
		URI: api.RepoURI(uri),
		ExternalRepoSpec: api.ExternalRepoSpec{
			ID:          id,
			ServiceType: serviceType,
			ServiceID:   serviceID,
		},
	}
}

type mockAuthnProvider struct {
	configID  auth.ProviderConfigID
	serviceID string
}

func (m mockAuthnProvider) ConfigID() auth.ProviderConfigID {
	return m.configID
}

func (m mockAuthnProvider) Config() schema.AuthProviders {
	panic("should not be called")
}

func (m mockAuthnProvider) CachedInfo() *auth.ProviderInfo {
	return &auth.ProviderInfo{ServiceID: m.serviceID}
}

func (m mockAuthnProvider) Refresh(ctx context.Context) error {
	panic("should not be called")
}

func mustURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
