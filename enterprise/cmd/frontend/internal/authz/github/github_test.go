package github

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
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
	userAccount *extsvc.ExternalAccount
	repos       map[authz.Repo]struct{}
	wantPerms   map[api.RepoName]map[authz.Perm]bool
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
				} else if !reflect.DeepEqual(gotPerms, c.wantPerms) {
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
	tests := []Provider_RepoPerms_Test{
		{
			description: "common_case",
			githubURL:   mustURL(t, "https://github.com"),
			cacheTTL:    3 * time.Hour,
			calls: []Provider_RepoPerms_call{
				{
					description: "t0_repos",
					userAccount: ua("u0", "t0"),
					repos: map[authz.Repo]struct{}{
						rp("r0", "u0/private", "https://github.com/"):  {},
						rp("r1", "u0/public", "https://github.com/"):   {},
						rp("r2", "u1/private", "https://github.com/"):  {},
						rp("r3", "u1/public", "https://github.com/"):   {},
						rp("r4", "u99/private", "https://github.com/"): {},
						rp("r5", "u99/public", "https://github.com/"):  {},
					},
					wantPerms: map[api.RepoName]map[authz.Perm]bool{
						"r0": readPerms,
						"r1": readPerms,
						"r2": noPerms,
						"r3": readPerms,
						"r4": noPerms,
						"r5": readPerms,
					},
				},
				{
					description: "t1_repos",
					userAccount: ua("u1", "t1"),
					repos: map[authz.Repo]struct{}{
						rp("r0", "u0/private", "https://github.com/"):  {},
						rp("r1", "u0/public", "https://github.com/"):   {},
						rp("r2", "u1/private", "https://github.com/"):  {},
						rp("r3", "u1/public", "https://github.com/"):   {},
						rp("r4", "u99/private", "https://github.com/"): {},
						rp("r5", "u99/public", "https://github.com/"):  {},
					},
					wantPerms: map[api.RepoName]map[authz.Perm]bool{
						"r0": noPerms,
						"r1": readPerms,
						"r2": readPerms,
						"r3": readPerms,
						"r4": noPerms,
						"r5": readPerms,
					},
				},
				{
					description: "repos_with_unknown_token_(only_public_repos)",
					userAccount: ua("unknown-user", "unknown-token"),
					repos: map[authz.Repo]struct{}{
						rp("r0", "u0/private", "https://github.com/"):  {},
						rp("r1", "u0/public", "https://github.com/"):   {},
						rp("r2", "u1/private", "https://github.com/"):  {},
						rp("r3", "u1/public", "https://github.com/"):   {},
						rp("r4", "u99/private", "https://github.com/"): {},
						rp("r5", "u99/public", "https://github.com/"):  {},
					},
					wantPerms: map[api.RepoName]map[authz.Perm]bool{
						"r0": noPerms,
						"r1": readPerms,
						"r2": noPerms,
						"r3": readPerms,
						"r4": noPerms,
						"r5": readPerms,
					},
				},
				{
					description: "public repos",
					userAccount: nil,
					repos: map[authz.Repo]struct{}{
						rp("r0", "u0/private", "https://github.com/"):  {},
						rp("r1", "u0/public", "https://github.com/"):   {},
						rp("r2", "u1/private", "https://github.com/"):  {},
						rp("r3", "u1/public", "https://github.com/"):   {},
						rp("r4", "u99/private", "https://github.com/"): {},
						rp("r5", "u99/public", "https://github.com/"):  {},
					},
					wantPerms: map[api.RepoName]map[authz.Perm]bool{
						"r1": readPerms,
						"r3": readPerms,
						"r5": readPerms,
					},
				},
				{
					description: "t0 select",
					userAccount: ua("u0", "t0"),
					repos: map[authz.Repo]struct{}{
						rp("r2", "u1/private", "https://github.com/"): {},
					},
					wantPerms: map[api.RepoName]map[authz.Perm]bool{
						"r2": noPerms,
					},
				},
				{
					description: "t0 missing",
					userAccount: ua("u0", "t0"),
					repos: map[authz.Repo]struct{}{
						rp("r00", "404", "https://github.com/"):             {},
						rp("r11", "u0/public", "https://other.github.com/"): {},
					},
					wantPerms: map[api.RepoName]map[authz.Perm]bool{
						"r00": noPerms,
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

var (
	readPerms = map[authz.Perm]bool{authz.Read: true}
	noPerms   = map[authz.Perm]bool{authz.Read: false}
)

func mustURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func ua(accountID, token string) *extsvc.ExternalAccount {
	var a extsvc.ExternalAccount
	a.AccountID = accountID
	github.SetExternalAccountData(&a.ExternalAccountData, nil, &oauth2.Token{
		AccessToken: token,
	})
	return &a
}
func rp(name, ghid, serviceID string) authz.Repo {
	return authz.Repo{
		RepoName: api.RepoName(name),
		ExternalRepoSpec: api.ExternalRepoSpec{
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
	repos := make(map[string]*github.Repository)
	for rid := range m.PublicRepos {
		repos[rid] = m.Repos[rid]
	}
	for rid := range m.TokenRepos[token] {
		repos[rid] = m.Repos[rid]
	}
	return repos, nil
}
