package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"golang.org/x/oauth2"
)

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

	type call struct {
		description string
		userAccount *extsvc.Account
		repos       []*types.Repo
		mockRepos   map[string]*github.Repository
		wantPerms   []authz.RepoPerms
		wantErr     error
	}

	tests := []struct {
		description string
		githubURL   *url.URL
		cacheTTL    time.Duration
		calls       []call
	}{
		{
			description: "common_case",
			githubURL:   mustURL(t, "https://github.com"),
			cacheTTL:    3 * time.Hour,
			calls: []call{
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
					mockRepos: map[string]*github.Repository{
						"u0/private": {ID: "u0/private", IsPrivate: true},
						"u0/public":  {ID: "u0/public"},
						"u1/public":  {ID: "u1/public"},
						"u99/public": {ID: "u99/public"},
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
					mockRepos: map[string]*github.Repository{
						"u0/public":  {ID: "u0/public"},
						"u1/private": {ID: "u1/private", IsPrivate: true},
						"u1/public":  {ID: "u1/public"},
						"u99/public": {ID: "u99/public"},
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
					mockRepos: map[string]*github.Repository{
						"u0/public":  {ID: "u0/public"},
						"u1/public":  {ID: "u1/public"},
						"u99/public": {ID: "u99/public"},
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
					mockRepos: map[string]*github.Repository{
						"u0/private":  {ID: "u0/private", IsPrivate: true},
						"u0/public":   {ID: "u0/public"},
						"u1/private":  {ID: "u1/private", IsPrivate: true},
						"u1/public":   {ID: "u1/public"},
						"u99/private": {ID: "u99/private", IsPrivate: true},
						"u99/public":  {ID: "u99/public"},
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
		t.Run(test.description, func(t *testing.T) {
			mockClient := &mockClient{}
			mockClient.MockWithToken = func(token string) client {
				return mockClient
			}

			p := NewProvider(test.githubURL, "base-token", test.cacheTTL, make(authz.MockCache))
			p.client = mockClient

			for j := 0; j < 2; j++ { // run twice for cache coherency
				for _, c := range test.calls {
					t.Run(fmt.Sprintf("%s: run %d", c.description, j), func(t *testing.T) {
						c := c

						called := 0
						mockClient.MockGetRepositoriesByNodeIDFromAPI = func(ctx context.Context, nodeIDs []string) (map[string]*github.Repository, error) {
							called++
							return c.mockRepos, nil
						}
						mockClient.MockGetRepositoryByNodeID = func(ctx context.Context, id string) (*github.Repository, error) {
							called++
							r, ok := c.mockRepos[id]
							if !ok {
								return nil, github.ErrNotFound
							}
							return r, nil
						}

						perms, err := p.RepoPerms(context.Background(), c.userAccount, c.repos)
						if diff := cmp.Diff(c.wantErr, err); diff != "" {
							t.Fatal(diff)
						}

						for _, ps := range [][]authz.RepoPerms{perms, c.wantPerms} {
							sort.Slice(ps, func(i, j int) bool {
								return ps[i].Repo.Name <= ps[j].Repo.Name
							})
						}
						if diff := cmp.Diff(c.wantPerms, perms); diff != "" {
							t.Fatal(diff)
						}

						if j == 1 && called > 0 {
							t.Fatal("expected entries to be fully cached")
						}
					})
				}
			}
		})
	}
}

func TestFetchUserRepos(t *testing.T) {
	mockClient := &mockClient{
		MockGetRepositoriesByNodeIDFromAPI: func(ctx context.Context, nodeIDs []string) (map[string]*github.Repository, error) {
			return map[string]*github.Repository{
				"u0/private": {ID: "u0/private", IsPrivate: true},
				"u0/public":  {ID: "u0/public"},
				"u1/public":  {ID: "u1/public"},
				"u99/public": {ID: "u99/public"},
			}, nil
		},
	}
	mockClient.MockWithToken = func(token string) client {
		return mockClient
	}

	p := NewProvider(mustURL(t, "https://github.com"), "base-token", 0, make(authz.MockCache))
	p.client = mockClient

	canAccess, isPublic, err := p.fetchUserRepos(context.Background(), ua("u0", "t0"), []string{
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

	if diff := cmp.Diff(wantCanAccess, canAccess); diff != "" {
		t.Fatal(diff)
	}
	if diff := cmp.Diff(wantIsPublic, isPublic); diff != "" {
		t.Fatal(diff)
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
		MockListAffiliatedRepositories: func(ctx context.Context, visibility github.Visibility, page int) ([]*github.Repository, bool, int, error) {
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
