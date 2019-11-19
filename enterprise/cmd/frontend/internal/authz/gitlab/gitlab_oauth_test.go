package gitlab

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

func Test_GitLab_RepoPerms(t *testing.T) {
	type call struct {
		description string
		account     *extsvc.ExternalAccount
		repos       []*types.Repo
		expPerms    []authz.RepoPerms
	}
	type test struct {
		description string
		op          OAuthAuthzProviderOp
		calls       []call
	}

	// Mock the following scenario:
	// - public projects begin with 99
	// - internal projects begin with 98
	// - private projects begin with the digit of the user that owns them (other users may have access)
	// - u1 owns its own repositories and nothing else
	// - u2 owns its own repos and has guest access to u1's
	// - u3 owns its own repos and has full access to u1's and guest access to u2's
	gitlabMock := newMockGitLab(mockGitLabOp{
		t: t,
		publicProjs: []int{ // public projects
			991,
		},
		internalProjs: []int{ // internal projects
			981,
		},
		privateProjs: map[int][2][]int32{ // private projects
			10: {
				{ // guests
					2,
				},
				{ // content ("full access")
					1,
					3,
				},
			},
			20: {
				{
					3,
				},
				{
					2,
				},
			},
			30: {
				{},
				{3},
			},
		},
		oauthToks: map[string]int32{
			"oauth-u1": 1,
			"oauth-u2": 2,
			"oauth-u3": 3,
		},
	})
	gitlab.MockGetProject = gitlabMock.GetProject
	gitlab.MockListTree = gitlabMock.ListTree

	repos := map[string]*types.Repo{
		"u1/repo1":       repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"),
		"u2/repo1":       repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"),
		"u3/repo1":       repo("u3/repo1", gitlab.ServiceType, "https://gitlab.mine/", "30"),
		"internal/repo1": repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"),
		"public/repo1":   repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"),
	}

	tests := []test{
		{
			description: "standard config",
			op: OAuthAuthzProviderOp{
				BaseURL: mustURL(t, "https://gitlab.mine"),
			},
			calls: []call{
				{
					description: "u1 user has expected perms",
					account:     acct(t, 1, "gitlab", "https://gitlab.mine/", "1", "oauth-u1"),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["u1/repo1"], Perms: authz.Read},
						{Repo: repos["internal/repo1"], Perms: authz.Read},
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "u2 user has expected perms",
					account:     acct(t, 2, "gitlab", "https://gitlab.mine/", "2", "oauth-u2"),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["u2/repo1"], Perms: authz.Read},
						{Repo: repos["internal/repo1"], Perms: authz.Read},
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "other user has expected perms (internal and public)",
					account:     acct(t, 4, "gitlab", "https://gitlab.mine/", "555", "oauth-other"),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["internal/repo1"], Perms: authz.Read},
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "no token means only public repos",
					account:     acct(t, 4, "gitlab", "https://gitlab.mine/", "555", ""),
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
				{
					description: "unauthenticated means only public repos",
					account:     nil,
					repos: []*types.Repo{
						repos["u1/repo1"],
						repos["u2/repo1"],
						repos["u3/repo1"],
						repos["internal/repo1"],
						repos["public/repo1"],
					},
					expPerms: []authz.RepoPerms{
						{Repo: repos["public/repo1"], Perms: authz.Read},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			for _, c := range test.calls {
				t.Logf("Call %q", c.description)

				// Recreate the authz provider cache every time, before running twice (once uncached, once cached)
				ctx := context.Background()
				op := test.op
				op.MockCache = make(mockCache)
				authzProvider := newOAuthProvider(op)

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
		})
	}
}

func Test_GitLab_RepoPerms_cache(t *testing.T) {
	gitlabMock := newMockGitLab(mockGitLabOp{
		t: t,
		publicProjs: []int{ // public projects
			991,
		},
		internalProjs: []int{ // internal projects
			981,
		},
		privateProjs: map[int][2][]int32{ // private projects
			10: {
				{ // guests
					2,
				},
				{ // content ("full access")
					1,
				},
			},
		},
		oauthToks: map[string]int32{
			"oauth-u1": 1,
			"oauth-u2": 2,
			"oauth-u3": 3,
		},
	})
	gitlab.MockGetProject = gitlabMock.GetProject
	gitlab.MockListTree = gitlabMock.ListTree

	ctx := context.Background()
	authzProvider := newOAuthProvider(OAuthAuthzProviderOp{
		BaseURL:   mustURL(t, "https://gitlab.mine"),
		MockCache: make(mockCache),
		CacheTTL:  3 * time.Hour,
	})

	// Initial request for private repo
	if _, err := authzProvider.RepoPerms(ctx,
		acct(t, 1, gitlab.ServiceType, "https://gitlab.mine/", "1", "oauth-u1"),
		[]*types.Repo{
			repo("10", "gitlab", "https://gitlab.mine/", "10"),
		},
	); err != nil {
		t.Fatal(err)
	}
	if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
		"oauth-u1": {{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}
	if actual, exp := gitlabMock.madeListTree, map[string]map[gitlab.ListTreeOp]int{
		"oauth-u1": {{ProjID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}

	// Exact same request
	if _, err := authzProvider.RepoPerms(ctx,
		acct(t, 1, gitlab.ServiceType, "https://gitlab.mine/", "1", "oauth-u1"),
		[]*types.Repo{
			repo("10", "gitlab", "https://gitlab.mine/", "10"),
		},
	); err != nil {
		t.Fatal(err)
	}
	if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
		"oauth-u1": {{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}
	if actual, exp := gitlabMock.madeListTree, map[string]map[gitlab.ListTreeOp]int{
		"oauth-u1": {{ProjID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}

	// Different request, on internal repo
	if _, err := authzProvider.RepoPerms(ctx,
		acct(t, 2, gitlab.ServiceType, "https://gitlab.mine/", "2", "oauth-u2"),
		[]*types.Repo{
			repo("981", "gitlab", "https://gitlab.mine/", "981"),
		},
	); err != nil {
		t.Fatal(err)
	}
	if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
		"oauth-u1": {{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
		"oauth-u2": {{ID: 981, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
	}; !reflect.DeepEqual(exp, actual) {
		t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
	}

	// Bypass cache when ttl changed
	authzProvider.cacheTTL = 1 * time.Hour

	// Make initial request twice again, expect cache miss the first time around
	if _, err := authzProvider.RepoPerms(ctx,
		acct(t, 1, gitlab.ServiceType, "https://gitlab.mine/", "1", "oauth-u1"),
		[]*types.Repo{
			repo("10", "gitlab", "https://gitlab.mine/", "10"),
		},
	); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if actual, exp := gitlabMock.madeGetProject, map[string]map[gitlab.GetProjectOp]int{
			"oauth-u1": {{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 2},
			"oauth-u2": {{ID: 981, CommonOp: gitlab.CommonOp{NoCache: true}}: 1},
		}; !reflect.DeepEqual(exp, actual) {
			t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
		}
		if actual, exp := gitlabMock.madeListTree, map[string]map[gitlab.ListTreeOp]int{
			"oauth-u1": {{ProjID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 2},
		}; !reflect.DeepEqual(exp, actual) {
			t.Errorf("Unexpected cache behavior. Expected %v, but got %v", exp, actual)
		}
	}
}
