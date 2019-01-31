package gitlab

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
)

func Test_GitLab_RepoPerms(t *testing.T) {
	// Mock the following scenario:
	// - public projects begin with 99
	// - internal projects begin with 98
	// - private projects begin with the digit of the user that owns them (other users may have access)
	// - u1 owns its own repositories and nothing else
	// - u2 owns its own repos and has guest access to u1's
	// - u3 owns its own repos and has full access to u1's and guest access to u2's
	gitlabMock := newMockGitLab(t,
		[]int{ // public projects
			991,
		},
		[]int{ // internal projects
			981,
		},
		map[int][2][]string{ // private projects
			10: [2][]string{
				[]string{ // guests
					"u2",
				},
				[]string{ // content ("full access")
					"u1",
					"u3",
				},
			},
			20: [2][]string{
				[]string{
					"u3",
				},
				[]string{
					"u2",
				},
			},
			30: [2][]string{
				[]string{},
				[]string{"u3"},
			},
		},
		map[string]string{
			"oauth-u1": "u1",
			"oauth-u2": "u2",
			"oauth-u3": "u3",
		},
	)
	gitlab.MockGetProject = gitlabMock.GetProject
	gitlab.MockListTree = gitlabMock.ListTree

	tests := []GitLab_RepoPerms_Test{
		{
			description: "standard config",
			op: GitLabOAuthAuthzProviderOp{
				BaseURL: mustURL(t, "https://gitlab.mine"),
			},
			calls: []GitLab_RepoPerms_call{
				{
					description: "u1 user has expected perms",
					account:     acct(1, "gitlab", "https://gitlab.mine/", "u1", "oauth-u1"),
					repos: map[authz.Repo]struct{}{
						repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"):        {},
						repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"):        {},
						repo("u3/repo1", gitlab.ServiceType, "https://gitlab.mine/", "30"):        {},
						repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"): {},
						repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"):   {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"u1/repo1":       {authz.Read: true},
						"internal/repo1": {authz.Read: true},
						"public/repo1":   {authz.Read: true},
					},
				},
				{
					description: "u2 user has expected perms",
					account:     acct(2, "gitlab", "https://gitlab.mine/", "u2", "oauth-u2"),
					repos: map[authz.Repo]struct{}{
						repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"):        {},
						repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"):        {},
						repo("u3/repo1", gitlab.ServiceType, "https://gitlab.mine/", "30"):        {},
						repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"): {},
						repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"):   {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"u2/repo1":       {authz.Read: true},
						"internal/repo1": {authz.Read: true},
						"public/repo1":   {authz.Read: true},
					},
				},
				{
					description: "other user has expected perms (internal and public)",
					account:     acct(4, "gitlab", "https://gitlab.mine/", "other", "oauth-other"),
					repos: map[authz.Repo]struct{}{
						repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"):        {},
						repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"):        {},
						repo("u3/repo1", gitlab.ServiceType, "https://gitlab.mine/", "30"):        {},
						repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"): {},
						repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"):   {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"internal/repo1": {authz.Read: true},
						"public/repo1":   {authz.Read: true},
					},
				},
				{
					description: "no token means only public repos",
					account:     acct(4, "gitlab", "https://gitlab.mine/", "no-token", ""),
					repos: map[authz.Repo]struct{}{
						repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"):        {},
						repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"):        {},
						repo("u3/repo1", gitlab.ServiceType, "https://gitlab.mine/", "30"):        {},
						repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"): {},
						repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"):   {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"public/repo1": {authz.Read: true},
					},
				},
				{
					description: "unauthenticated means only public repos",
					account:     nil,
					repos: map[authz.Repo]struct{}{
						repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"):        {},
						repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"):        {},
						repo("u3/repo1", gitlab.ServiceType, "https://gitlab.mine/", "30"):        {},
						repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"): {},
						repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"):   {},
					},
					expPerms: map[api.RepoName]map[authz.Perm]bool{
						"public/repo1": {authz.Read: true},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test.run(t)
	}
}

func Test_GitLab_RepoPerms_cache(t *testing.T) {
	gitlabMock := newMockGitLab(t,
		[]int{ // public projects
			991,
		},
		[]int{ // internal projects
			981,
		},
		map[int][2][]string{ // private projects
			10: [2][]string{
				[]string{ // guests
					"u2",
				},
				[]string{ // content ("full access")
					"u1",
				},
			},
		},
		map[string]string{
			"oauth-u1": "u1",
			"oauth-u2": "u2",
			"oauth-u3": "u3",
		},
	)
	gitlab.MockGetProject = gitlabMock.GetProject
	gitlab.MockListTree = gitlabMock.ListTree

	ctx := context.Background()
	authzProvider := NewProvider(GitLabOAuthAuthzProviderOp{
		BaseURL:   mustURL(t, "https://gitlab.mine"),
		MockCache: make(mockCache),
		CacheTTL:  3 * time.Hour,
	})

	// Initial request for private repo
	if _, err := authzProvider.RepoPerms(ctx,
		acct(1, gitlab.ServiceType, "https://gitlab.mine/", "u1", "oauth-u1"),
		map[authz.Repo]struct{}{
			repo("10", "gitlab", "https://gitlab.mine/", "10"): {},
		},
	); err != nil {
		t.Fatal(err)
	}
	if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
		"oauth-u1": map[gitlab.GetProjectOp]int{{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}
	if actual, exp := gitlabMock.madeListTree, map[string]map[gitlab.ListTreeOp]int{
		"oauth-u1": map[gitlab.ListTreeOp]int{{ProjID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}

	// Exact same request
	if _, err := authzProvider.RepoPerms(ctx,
		acct(1, gitlab.ServiceType, "https://gitlab.mine/", "u1", "oauth-u1"),
		map[authz.Repo]struct{}{
			repo("10", "gitlab", "https://gitlab.mine/", "10"): {},
		},
	); err != nil {
		t.Fatal(err)
	}
	if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
		"oauth-u1": map[gitlab.GetProjectOp]int{{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}
	if actual, exp := gitlabMock.madeListTree, map[string]map[gitlab.ListTreeOp]int{
		"oauth-u1": map[gitlab.ListTreeOp]int{{ProjID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}

	// Different request, on internal repo
	if _, err := authzProvider.RepoPerms(ctx,
		acct(2, gitlab.ServiceType, "https://gitlab.mine/", "u2", "oauth-u2"),
		map[authz.Repo]struct{}{
			repo("981", "gitlab", "https://gitlab.mine/", "981"): {},
		},
	); err != nil {
		t.Fatal(err)
	}
	if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
		"oauth-u1": map[gitlab.GetProjectOp]int{{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
		"oauth-u2": map[gitlab.GetProjectOp]int{{ID: 981, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}

	// Bypass cache when ttl changed
	authzProvider.cacheTTL = 1 * time.Hour

	// Make initial request twice again, expect cache miss the first time around
	if _, err := authzProvider.RepoPerms(ctx,
		acct(1, gitlab.ServiceType, "https://gitlab.mine/", "u1", "oauth-u1"),
		map[authz.Repo]struct{}{
			repo("10", "gitlab", "https://gitlab.mine/", "10"): {},
		},
	); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
			"oauth-u1": map[gitlab.GetProjectOp]int{{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 2},
			"oauth-u2": map[gitlab.GetProjectOp]int{{ID: 981, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
		}; !reflect.DeepEqual(exp, actual) {
			t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
		}
		if actual, exp := gitlabMock.madeListTree, map[string]map[gitlab.ListTreeOp]int{
			"oauth-u1": map[gitlab.ListTreeOp]int{{ProjID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 2},
		}; !reflect.DeepEqual(exp, actual) {
			t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
		}
	}
}

type GitLab_RepoPerms_Test struct {
	description string

	op GitLabOAuthAuthzProviderOp

	calls []GitLab_RepoPerms_call
}

type GitLab_RepoPerms_call struct {
	description string
	account     *extsvc.ExternalAccount
	repos       map[authz.Repo]struct{}
	expPerms    map[api.RepoName]map[authz.Perm]bool
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

func Test_GitLab_Repos(t *testing.T) {
	tests := []GitLab_Repos_Test{
		{
			description: "standard config",
			op: GitLabOAuthAuthzProviderOp{
				BaseURL: mustURL(t, "https://gitlab.mine"),
			},
			calls: []GitLab_Repos_call{
				{
					repos: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""):                   {},
						repo("gitlab.mine/kl/repo-1", "", "", ""):                   {},
						repo("another.host/bl/repo-1", "", "", ""):                  {},
						repo("a", gitlab.ServiceType, "https://gitlab.mine/", "23"): {},
						repo("b", gitlab.ServiceType, "https://not-mine/", "34"):    {},
						repo("c", "not-gitlab", "https://gitlab.mine/", "45"):       {},
					},
					expMine: map[authz.Repo]struct{}{
						repo("a", gitlab.ServiceType, "https://gitlab.mine/", "23"): {},
					},
					expOthers: map[authz.Repo]struct{}{
						repo("gitlab.mine/bl/repo-1", "", "", ""):                {},
						repo("gitlab.mine/kl/repo-1", "", "", ""):                {},
						repo("another.host/bl/repo-1", "", "", ""):               {},
						repo("b", gitlab.ServiceType, "https://not-mine/", "34"): {},
						repo("c", "not-gitlab", "https://gitlab.mine/", "45"):    {},
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
	op          GitLabOAuthAuthzProviderOp
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
