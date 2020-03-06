package gitlab

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// Test_GitLab_RepoPerms checks for correctness and cache-correctness
func Test_GitLab_RepoPerms(t *testing.T) {
	type call struct {
		description string
		account     *extsvc.ExternalAccount
		repos       []*types.Repo
		expPerms    []authz.RepoPerms
	}
	type test struct {
		description string

		// Configures the provider. Do NOT set the MaxBatchRequests and MinBatchThreshold fields in
		// the test struct declarations, as these are set to a range of values in the test logic,
		// itself.
		op OAuthAuthzProviderOp

		calls []call
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
	gitlab.MockListProjects = gitlabMock.ListProjects
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
				for _, batchingThreshold := range [][2]int{
					{0, 0},     // should NOT trigger batch-fetching visibility
					{1, 300},   // should trigger batch-fetching visibility
					{1, 1},     // should trigger batch-fetching visibility, but hit maximum
					{200, 300}, // should NOT trigger batch-fetching visibility
				} {
					t.Logf("Call %q, batchingThreshold %d", c.description, batchingThreshold)

					// Recreate the authz provider cache every time, before running twice (once uncached, once cached)
					ctx := context.Background()
					op := test.op
					op.MinBatchThreshold, op.MaxBatchRequests = batchingThreshold[0], batchingThreshold[1]
					op.MockCache = make(mockCache)
					authzProvider := newOAuthProvider(op)
					for i := 0; i < 2; i++ {
						t.Logf("iter %d", i)
						perms, err := authzProvider.RepoPerms(ctx, c.account, c.repos)
						if err != nil {
							t.Errorf("unexpected error: %v", err)
							continue
						}
						sort.Sort(authz.RepoPermsSort(perms))
						sort.Sort(authz.RepoPermsSort(c.expPerms))
						if diff := cmp.Diff(perms, c.expPerms); diff != "" {
							t.Errorf("perms != c.expPerms:\n%s", diff)
						}
					}
				}
			}
		})
	}
}

// Test_GitLab_RepoPerms_cache tests for cache-effectiveness (we cache what we expect to cache).
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

// Test_GitLab_RepoPerms_batchVisibility tests that Project visibility is fetched in batch (for
// performance reasons)
func Test_GitLab_RepoPerms_batchVisibility(t *testing.T) {
	gitlabMockOp := mockGitLabOp{
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
		},
		oauthToks: map[string]int32{
			"oauth-u1": 1,
			"oauth-u2": 2,
			"oauth-u3": 3,
		},
	}
	repos := map[string]*types.Repo{
		"u1/repo1":       repo("u1/repo1", gitlab.ServiceType, "https://gitlab.mine/", "10"),
		"u2/repo1":       repo("u2/repo1", gitlab.ServiceType, "https://gitlab.mine/", "20"),
		"internal/repo1": repo("internal/repo1", gitlab.ServiceType, "https://gitlab.mine/", "981"),
		"public/repo1":   repo("public/repo1", gitlab.ServiceType, "https://gitlab.mine/", "991"),
	}
	repoSlice := make([]*types.Repo, 0, len(repos))
	for _, r := range repos {
		repoSlice = append(repoSlice, r)
	}

	{
		// Test case 1: batching threshold hit, { MinBatchThreshold: 1, MaxBatchRequests: 300 }
		gitlabMock := newMockGitLab(gitlabMockOp)
		gitlab.MockGetProject = gitlabMock.GetProject
		gitlab.MockListProjects = gitlabMock.ListProjects
		gitlab.MockListTree = gitlabMock.ListTree

		ctx := context.Background()
		authzProvider := newOAuthProvider(OAuthAuthzProviderOp{
			BaseURL:           mustURL(t, "https://gitlab.mine"),
			MockCache:         make(mockCache),
			CacheTTL:          3 * time.Hour,
			MinBatchThreshold: 1,
			MaxBatchRequests:  300,
		})
		expPerms := []authz.RepoPerms{
			{Repo: repos["u1/repo1"], Perms: authz.Read},
			{Repo: repos["internal/repo1"], Perms: authz.Read},
			{Repo: repos["public/repo1"], Perms: authz.Read},
		}
		expMadeGetProject := map[string]map[gitlab.GetProjectOp]int{
			"oauth-u1": {
				{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}: 1,
				{ID: 20, CommonOp: gitlab.CommonOp{NoCache: true}}: 1,
			},
		}
		expMadeListProjects := map[string]map[string]int{
			"oauth-u1": {"projects?per_page=100": 1},
		}

		perms, err := authzProvider.RepoPerms(ctx, acct(t, 1, "gitlab", "https://gitlab.mine/", "1", "oauth-u1"), repoSlice)
		if err != nil {
			t.Fatal(err)
		}
		sort.Sort(authz.RepoPermsSort(perms))
		sort.Sort(authz.RepoPermsSort(expPerms))
		if diff := cmp.Diff(expPerms, perms); diff != "" {
			t.Errorf("expPerms != perms:\n%s", diff)
		}
		if diff := cmp.Diff(expMadeGetProject, gitlabMock.madeGetProject); diff != "" {
			t.Errorf("expMadeGetProject != gitlabMock.madeGetProject:\n%s", diff)
		}
		if diff := cmp.Diff(expMadeListProjects, gitlabMock.madeListProjects); diff != "" {
			t.Errorf("expMadeListProjects != gitlabMock.madeListProjects:\n%s", diff)
		}

		gitlab.MockGetProject = nil
		gitlab.MockListProjects = nil
		gitlab.MockListTree = nil
	}

	{
		// Test case 2: batching threshold NOT hit { MinBatchThreshold: 200, MaxBatchRequests: 300 }
		gitlabMock := newMockGitLab(gitlabMockOp)
		gitlab.MockGetProject = gitlabMock.GetProject
		gitlab.MockListProjects = gitlabMock.ListProjects
		gitlab.MockListTree = gitlabMock.ListTree

		ctx := context.Background()
		authzProvider := newOAuthProvider(OAuthAuthzProviderOp{
			BaseURL:           mustURL(t, "https://gitlab.mine"),
			MockCache:         make(mockCache),
			CacheTTL:          3 * time.Hour,
			MinBatchThreshold: 200,
			MaxBatchRequests:  300,
		})
		expPerms := []authz.RepoPerms{
			{Repo: repos["u1/repo1"], Perms: authz.Read},
			{Repo: repos["internal/repo1"], Perms: authz.Read},
			{Repo: repos["public/repo1"], Perms: authz.Read},
		}
		expMadeGetProject := map[string]map[gitlab.GetProjectOp]int{
			"oauth-u1": {
				{ID: 10, CommonOp: gitlab.CommonOp{NoCache: true}}:  1,
				{ID: 20, CommonOp: gitlab.CommonOp{NoCache: true}}:  1,
				{ID: 981, CommonOp: gitlab.CommonOp{NoCache: true}}: 1,
				{ID: 991, CommonOp: gitlab.CommonOp{NoCache: true}}: 1,
			},
		}
		expMadeListProjects := map[string]map[string]int{}

		perms, err := authzProvider.RepoPerms(ctx, acct(t, 1, "gitlab", "https://gitlab.mine/", "1", "oauth-u1"), repoSlice)
		if err != nil {
			t.Fatal(err)
		}
		sort.Sort(authz.RepoPermsSort(perms))
		sort.Sort(authz.RepoPermsSort(expPerms))
		if diff := cmp.Diff(perms, expPerms); diff != "" {
			t.Errorf("perms != expPerms:\n%s", diff)
		}
		if diff := cmp.Diff(expMadeGetProject, gitlabMock.madeGetProject); diff != "" {
			t.Errorf("expMadeGetProject != gitlabMock.madeGetProject:\n%s", diff)
		}
		if diff := cmp.Diff(expMadeListProjects, gitlabMock.madeListProjects); diff != "" {
			t.Errorf("expMadeListProjects != gitlabMock.madeListProjects:\n%s", diff)
		}

		gitlab.MockGetProject = nil
		gitlab.MockListProjects = nil
		gitlab.MockListTree = nil
	}
}
