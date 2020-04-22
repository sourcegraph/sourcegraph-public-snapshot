package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"golang.org/x/oauth2"
)

type Provider_RepoPerms_Test struct {
	description string
	githubURL   *url.URL
	cacheTTL    time.Duration
	calls       []Provider_RepoPerms_call
}

type Provider_RepoPerms_call struct {
	description string
	userAccount *extsvc.Account
	repos       []*types.Repo
	wantPerms   []authz.RepoPerms
	wantErr     error
}

func (p *Provider_RepoPerms_Test) run(t *testing.T) {
	githubMock := newMockGitHub([]*github.Repository{
		{ID: "u0/private", IsPrivate: true},
		{ID: "u0/public"},
		{ID: "u1/private", IsPrivate: true},
		{ID: "u1/public"},
		{ID: "u99/private", IsPrivate: true},
		{ID: "u99/public"},
	}, map[string][]string{
		"t0": {"u0/private", "u0/public"},
		"t1": {"u1/private", "u1/public"},
	})
	github.GetRepositoryByNodeIDMock = githubMock.GetRepositoryByNodeID
	defer func() { github.GetRepositoryByNodeIDMock = nil }()
	github.GetRepositoriesByNodeIDFromAPIMock = githubMock.GetRepositoriesByNodeIDFromAPI
	defer func() { github.GetRepositoriesByNodeIDFromAPIMock = nil }()

	provider := NewProvider(p.githubURL, "base-token", p.cacheTTL, make(authz.MockCache))
	for j := 0; j < 2; j++ { // run twice for cache coherency
		for _, c := range p.calls {
			t.Run(fmt.Sprintf("%s: run %d", c.description, j), func(t *testing.T) {
				c := c
				ctx := context.Background()
				githubMock.getRepositoryByNodeIDCount = 0

				gotPerms, gotErr := provider.RepoPerms(ctx, c.userAccount, c.repos)
				if gotErr != c.wantErr {
					t.Errorf("expected err %v, got err %v", c.wantErr, gotErr)
				}

				for _, perms := range [][]authz.RepoPerms{gotPerms, c.wantPerms} {
					sort.Slice(perms, func(i, j int) bool {
						return perms[i].Repo.Name <= perms[j].Repo.Name
					})
				}

				if !reflect.DeepEqual(gotPerms, c.wantPerms) {
					dmp := diffmatchpatch.New()
					t.Errorf("expected perms did not equal actual, diff:\n%s",
						dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(c.wantPerms), spew.Sdump(gotPerms), false)))
				}

				if j == 1 && githubMock.getRepositoryByNodeIDCount > 0 {
					t.Errorf("expected entries to be fully cached")
				}
			})
		}
	}
}

func TestProvider_RepoPerms(t *testing.T) {
	repos := map[string]*types.Repo{
		"r0":  rp("r0", "u0/private", "https://github.com/"),
		"r1":  rp("r1", "u0/public", "https://github.com/"),
		"r2":  rp("r2", "u1/private", "https://github.com/"),
		"r3":  rp("r3", "u1/public", "https://github.com/"),
		"r4":  rp("r4", "u99/private", "https://github.com/"),
		"r5":  rp("r5", "u99/public", "https://github.com/"),
		"r00": rp("r00", "404", "https://github.com/"),
	}

	tests := []Provider_RepoPerms_Test{
		{
			description: "common_case",
			githubURL:   mustURL(t, "https://github.com"),
			cacheTTL:    3 * time.Hour,
			calls: []Provider_RepoPerms_call{
				{
					description: "t0_repos",
					userAccount: ua("u0", "t0"),
					repos: []*types.Repo{
						repos["r0"],
						repos["r1"],
						repos["r2"],
						repos["r3"],
						repos["r4"],
						repos["r5"],
					},
					wantPerms: []authz.RepoPerms{
						{Repo: repos["r0"], Perms: authz.Read},
						{Repo: repos["r1"], Perms: authz.Read},
						{Repo: repos["r2"], Perms: authz.None},
						{Repo: repos["r3"], Perms: authz.Read},
						{Repo: repos["r4"], Perms: authz.None},
						{Repo: repos["r5"], Perms: authz.Read},
					},
				},
				{
					description: "t1_repos",
					userAccount: ua("u1", "t1"),
					repos: []*types.Repo{
						repos["r0"],
						repos["r1"],
						repos["r2"],
						repos["r3"],
						repos["r4"],
						repos["r5"],
					},
					wantPerms: []authz.RepoPerms{
						{Repo: repos["r0"], Perms: authz.None},
						{Repo: repos["r1"], Perms: authz.Read},
						{Repo: repos["r2"], Perms: authz.Read},
						{Repo: repos["r3"], Perms: authz.Read},
						{Repo: repos["r4"], Perms: authz.None},
						{Repo: repos["r5"], Perms: authz.Read},
					},
				},
				{
					description: "repos_with_unknown_token_(only_public_repos)",
					userAccount: ua("unknown-user", "unknown-token"),
					repos: []*types.Repo{
						repos["r0"],
						repos["r1"],
						repos["r2"],
						repos["r3"],
						repos["r4"],
						repos["r5"],
					},
					wantPerms: []authz.RepoPerms{
						{Repo: repos["r0"], Perms: authz.None},
						{Repo: repos["r1"], Perms: authz.Read},
						{Repo: repos["r2"], Perms: authz.None},
						{Repo: repos["r3"], Perms: authz.Read},
						{Repo: repos["r4"], Perms: authz.None},
						{Repo: repos["r5"], Perms: authz.Read},
					},
				},
				{
					description: "public repos",
					userAccount: nil,
					repos: []*types.Repo{
						repos["r0"],
						repos["r1"],
						repos["r2"],
						repos["r3"],
						repos["r4"],
						repos["r5"],
					},
					wantPerms: []authz.RepoPerms{
						{Repo: repos["r1"], Perms: authz.Read},
						{Repo: repos["r3"], Perms: authz.Read},
						{Repo: repos["r5"], Perms: authz.Read},
					},
				},
				{
					description: "t0 select",
					userAccount: ua("u0", "t0"),
					repos: []*types.Repo{
						repos["r2"],
					},
					wantPerms: []authz.RepoPerms{
						{Repo: repos["r2"], Perms: authz.None},
					},
				},
				{
					description: "t0 missing",
					userAccount: ua("u0", "t0"),
					repos: []*types.Repo{
						repos["r00"],
					},
					wantPerms: []authz.RepoPerms{
						{Repo: repos["r00"], Perms: authz.None},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, test.run)
	}
}

func Test_fetchUserRepos(t *testing.T) {
	githubMock := newMockGitHub([]*github.Repository{
		{ID: "u0/private", IsPrivate: true},
		{ID: "u0/public"},
		{ID: "u1/private", IsPrivate: true},
		{ID: "u1/public"},
		{ID: "u99/private", IsPrivate: true},
		{ID: "u99/public"},
	}, map[string][]string{
		"t0": {"u0/private", "u0/public"},
		"t1": {"u1/private", "u1/public"},
	})
	github.GetRepositoriesByNodeIDFromAPIMock = githubMock.GetRepositoriesByNodeIDFromAPI
	defer func() { github.GetRepositoriesByNodeIDFromAPIMock = nil }()
	oldMaxNodeIDs := github.MaxNodeIDs
	github.MaxNodeIDs = 2
	defer func() { github.MaxNodeIDs = oldMaxNodeIDs }()

	provider := NewProvider(mustURL(t, "https://github.com"), "base-token", 0, make(authz.MockCache))
	canAccess, isPublic, err := provider.fetchUserRepos(context.Background(), ua("u0", "t0"), []string{
		"u0/private",
		"u0/public",
		"u1/private",
		"u1/public",
		"u99/private",
		"u99/public",
	})
	if err != nil {
		t.Fatal(err)
	}

	wantCanAccess := map[string]bool{
		"u0/private":  true,
		"u0/public":   true,
		"u1/private":  false,
		"u1/public":   true,
		"u99/private": false,
		"u99/public":  true,
	}
	wantIsPublic := map[string]bool{
		"u0/private": false,
		"u0/public":  true,
		"u1/public":  true,
		"u99/public": true,
	}

	if !reflect.DeepEqual(canAccess, wantCanAccess) {
		t.Errorf("canAccess %+v != wantCanAccess %+v", canAccess, wantCanAccess)
	}
	if !reflect.DeepEqual(isPublic, wantIsPublic) {
		t.Errorf("isPublic %+v != wantIsPublic %+v", isPublic, wantIsPublic)
	}
}

func mustURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func ua(accountID, token string) *extsvc.Account {
	var a extsvc.Account
	a.AccountID = accountID
	github.SetExternalAccountData(&a.AccountData, nil, &oauth2.Token{
		AccessToken: token,
	})
	return &a
}

func rp(name, ghid, serviceID string) *types.Repo {
	return &types.Repo{
		Name: api.RepoName(name),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          ghid,
			ServiceType: github.ServiceType,
			ServiceID:   serviceID,
		},
	}
}

type mockGitHub struct {
	// Repos is a map from repo ID to repository
	Repos map[string]*github.Repository

	// TokenRepos is a map from auth token to list of repo IDs that are explicitly readable with that token
	TokenRepos map[string]map[string]struct{}

	// PublicRepos is the set of repo IDs corresponding to public repos
	PublicRepos map[string]struct{}

	// getRepositoryByNodeIDCount tracks the number of times GetRepositoryByNodeID is called
	getRepositoryByNodeIDCount int

	// getRepositoriesByNodeIDCount tracks the number of times GetRepositoriesByNodeIDFromAPI is called
	getRepositoriesByNodeIDCount int
}

func newMockGitHub(repos []*github.Repository, tokenRepos map[string][]string) *mockGitHub {
	rp := make(map[string]*github.Repository)
	for _, r := range repos {
		rp[r.ID] = r
	}
	tr := make(map[string]map[string]struct{})
	for t, rps := range tokenRepos {
		tr[t] = make(map[string]struct{})
		for _, r := range rps {
			tr[t][r] = struct{}{}
		}
	}
	pr := make(map[string]struct{})
	for _, r := range repos {
		if !r.IsPrivate {
			pr[r.ID] = struct{}{}
		}
	}
	return &mockGitHub{
		Repos:       rp,
		TokenRepos:  tr,
		PublicRepos: pr,
	}
}

func (m *mockGitHub) GetRepositoryByNodeID(ctx context.Context, token, id string) (repo *github.Repository, err error) {
	m.getRepositoryByNodeIDCount++
	if _, isPublic := m.PublicRepos[id]; isPublic {
		r, ok := m.Repos[id]
		if !ok {
			return nil, github.ErrNotFound
		}
		return r, nil
	}

	if token == "" {
		return nil, github.ErrNotFound
	}

	tr := m.TokenRepos[token]
	if tr == nil {
		return nil, github.ErrNotFound
	}
	if _, explicit := tr[id]; !explicit {
		return nil, github.ErrNotFound
	}
	r, ok := m.Repos[id]
	if !ok {
		return nil, github.ErrNotFound
	}
	return r, nil
}

func (m *mockGitHub) GetRepositoriesByNodeIDFromAPI(ctx context.Context, token string, nodeIDs []string) (map[string]*github.Repository, error) {
	m.getRepositoriesByNodeIDCount++

	repos := make(map[string]*github.Repository)
	for rid := range m.PublicRepos {
		repos[rid] = m.Repos[rid]
	}
	for rid := range m.TokenRepos[token] {
		repos[rid] = m.Repos[rid]
	}
	return repos, nil
}

func TestProvider_FetchUserPerms(t *testing.T) {
	t.Run("nil account", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://github.com"), "admin_token", 3*time.Hour, nil)
		_, err := p.FetchUserPerms(context.Background(), nil)
		want := "no account provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the account", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://github.com"), "admin_token", 3*time.Hour, nil)
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gitlab",
					ServiceID:   "https://gitlab.com/",
				},
			},
		)
		want := `not a code host of the account: want "https://gitlab.com/" but have "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("no token found in account data", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://github.com"), "admin_token", 3*time.Hour, nil)
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				},
				AccountData: extsvc.AccountData{},
			},
		)
		want := `no token found in the external account data`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	mockClient := &mockClient{
		MockListAffiliatedRepositories: func(ctx context.Context, page int) ([]*github.Repository, bool, int, error) {
			switch page {
			case 1:
				return []*github.Repository{
					{ID: "MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE="},
					{ID: "MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY="},
				}, true, 1, nil
			case 2:
				return []*github.Repository{
					{ID: "MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA="},
				}, false, 1, nil
			}

			return []*github.Repository{}, false, 1, nil
		},
	}
	calledWithToken := false
	mockClient.MockWithToken = func(token string) client {
		calledWithToken = true
		return mockClient
	}

	p := NewProvider(mustURL(t, "https://github.com"), "admin_token", 3*time.Hour, nil)
	p.client = mockClient

	authData := json.RawMessage(`{"access_token": "my_access_token"}`)
	repoIDs, err := p.FetchUserPerms(context.Background(),
		&extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
			AccountData: extsvc.AccountData{
				AuthData: &authData,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !calledWithToken {
		t.Fatal("!calledWithToken")
	}

	expRepoIDs := []extsvc.RepoID{
		"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
		"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
		"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
	}
	if diff := cmp.Diff(expRepoIDs, repoIDs); diff != "" {
		t.Fatal(diff)
	}
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://github.com"), "admin_token", 3*time.Hour, nil)
		_, err := p.FetchRepoPerms(context.Background(), nil)
		want := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := NewProvider(mustURL(t, "https://github.com"), "admin_token", 3*time.Hour, nil)
		_, err := p.FetchRepoPerms(context.Background(),
			&extsvc.Repository{
				URI: "gitlab.com/user/repo",
				ExternalRepoSpec: api.ExternalRepoSpec{
					ServiceType: "gitlab",
					ServiceID:   "https://gitlab.com/",
				},
			},
		)
		want := `not a code host of the repository: want "https://gitlab.com/" but have "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	p := NewProvider(mustURL(t, "https://github.com"), "admin_token", 3*time.Hour, nil)
	p.client = &mockClient{
		MockListRepositoryCollaborators: func(ctx context.Context, owner, repo string, page int) ([]*github.Collaborator, bool, error) {
			switch page {
			case 1:
				return []*github.Collaborator{
					{ID: "MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE="},
					{ID: "MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY="},
				}, true, nil
			case 2:
				return []*github.Collaborator{
					{ID: "MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA="},
				}, false, nil
			}

			return []*github.Collaborator{}, false, nil
		},
	}

	accountIDs, err := p.FetchRepoPerms(context.Background(),
		&extsvc.Repository{
			URI: "github.com/user/repo",
			ExternalRepoSpec: api.ExternalRepoSpec{
				ID:          "github_project_id",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	expAccountIDs := []extsvc.AccountID{
		"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
		"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
		"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
	}
	if diff := cmp.Diff(expAccountIDs, accountIDs); diff != "" {
		t.Fatal(diff)
	}
}
