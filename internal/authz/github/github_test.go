package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func mustURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestProvider_FetchUserPerms(t *testing.T) {
	t.Run("nil account", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchUserPerms(context.Background(), nil, nil)
		want := "no account provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the account", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gitlab",
					ServiceID:   "https://gitlab.com/",
				},
			},
			nil,
		)
		want := `not a code host of the account: want "https://gitlab.com/" but have "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("no token found in account data", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchUserPerms(context.Background(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				},
				AccountData: extsvc.AccountData{},
			},
			nil,
		)
		want := `no token found in the external account data`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	var (
		mockListAffiliatedRepositories = func(ctx context.Context, visibility github.Visibility, page int, affiliations ...github.Affiliation) ([]*github.Repository, bool, int, error) {
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
		}
		mockListOrgDetails = func(ctx context.Context, page int) (orgs []*github.OrgDetails, hasNextPage bool, rateLimitCost int, err error) {
			switch page {
			case 1:
				return []*github.OrgDetails{
					// does not have access to this org
					{Org: github.Org{Login: "not-sourcegraph"}, DefaultRepositoryPermission: "none"},
				}, true, 1, nil
			case 2:
				return []*github.OrgDetails{
					// has access to this org
					{Org: github.Org{Login: "sourcegraph"}, DefaultRepositoryPermission: "read"},
				}, false, 1, nil
			}
			return nil, false, 1, nil
		}
		mockListOrgRepositories = func(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
			switch org {
			case "sourcegraph":
				switch page {
				case 1:
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE="}, // existing repo
						{ID: "MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234="},
					}, true, 1, nil
				case 2:
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zaXRvcnkyNDI2NTE5678="},
					}, false, 1, nil
				}
			}
			t.Fatalf("unexpected call to ListOrgRepositories with org %q page %d", org, page)
			return nil, false, 1, nil
		}
	)

	t.Run("user has no orgs and teams", func(t *testing.T) {
		mockClient := &mockClient{
			MockListAffiliatedRepositories: mockListAffiliatedRepositories,
			MockGetAuthenticatedUserOrgsDetails: func(ctx context.Context, page int) (orgs []*github.OrgDetails, hasNextPage bool, rateLimitCost int, err error) {
				// No orgs
				return nil, false, 1, nil
			},
			MockGetAuthenticatedUserTeams: func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
				// No teams
				return nil, false, 1, nil
			},
		}
		// should call with token
		calledWithToken := false
		mockClient.MockWithToken = func(token string) client {
			calledWithToken = true
			return mockClient
		}

		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
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
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}

		if !calledWithToken {
			t.Fatal("!calledWithToken")
		}

		wantRepoIDs := []extsvc.RepoID{
			"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
			"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
			"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
		}
		if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
			t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("uncached: user in orgs", func(t *testing.T) {
		mockClient := &mockClient{
			MockListAffiliatedRepositories:      mockListAffiliatedRepositories,
			MockGetAuthenticatedUserOrgsDetails: mockListOrgDetails,
			MockGetAuthenticatedUserTeams: func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
				// No teams
				return nil, false, 1, nil
			},
			MockListOrgRepositories: mockListOrgRepositories,
		}

		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
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
			&authz.FetchPermOptions{InvalidateCaches: true},
		)
		if err != nil {
			t.Fatal(err)
		}

		wantRepoIDs := []extsvc.RepoID{
			"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
			"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
			"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
			"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234=",
			"MDEwOlJlcG9zaXRvcnkyNDI2NTE5678=",
		}
		if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
			t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("uncached: user in orgs and teams", func(t *testing.T) {
		mockClient := &mockClient{
			MockListAffiliatedRepositories:      mockListAffiliatedRepositories,
			MockGetAuthenticatedUserOrgsDetails: mockListOrgDetails,
			MockGetAuthenticatedUserTeams: func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
				switch page {
				case 1:
					return []*github.Team{
						// should not get repos from this team because parent org has default read permissions
						{Organization: github.Org{Login: "sourcegraph"}, Name: "ns team", Slug: "ns-team"},
						// should not get repos from this team since it has no repos
						{Organization: github.Org{Login: "not-sourcegraph"}, Name: "ns team", Slug: "ns-team", ReposCount: 0},
					}, true, 1, nil
				case 2:
					return []*github.Team{
						// should get repos from this team
						{Organization: github.Org{Login: "not-sourcegraph"}, Name: "ns team 2", Slug: "ns-team-2", ReposCount: 3},
					}, false, 1, nil
				}
				return nil, false, 1, nil
			},
			MockListOrgRepositories: mockListOrgRepositories,
			MockListTeamRepositories: func(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
				switch org {
				case "not-sourcegraph":
					switch team {
					case "ns-team-2":
						switch page {
						case 1:
							return []*github.Repository{
								{ID: "MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA="}, // existing repo
								{ID: "MDEwOlJlcG9zaXRvcnkyNDQ1nsteam1="},
							}, true, 1, nil
						case 2:
							return []*github.Repository{
								{ID: "MDEwOlJlcG9zaXRvcnkyNDI2nsteam2="},
							}, false, 1, nil
						}
					}
				}
				t.Fatalf("unexpected call to ListOrgRepositories with org %q team %q page %d", org, team, page)
				return nil, false, 1, nil
			},
		}

		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
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
			&authz.FetchPermOptions{InvalidateCaches: true},
		)
		if err != nil {
			t.Fatal(err)
		}

		wantRepoIDs := []extsvc.RepoID{
			"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
			"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
			"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
			"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234=",
			"MDEwOlJlcG9zaXRvcnkyNDI2NTE5678=",
			"MDEwOlJlcG9zaXRvcnkyNDQ1nsteam1=",
			"MDEwOlJlcG9zaXRvcnkyNDI2nsteam2=",
		}
		if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
			t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchRepoPerms(context.Background(), nil)
		want := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
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

	p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
	p.client = &mockClient{
		MockListRepositoryCollaborators: func(ctx context.Context, owner, repo string, page int) ([]*github.Collaborator, bool, error) {
			switch page {
			case 1:
				return []*github.Collaborator{
					{DatabaseID: 57463526},
					{DatabaseID: 67471},
				}, true, nil
			case 2:
				return []*github.Collaborator{
					{DatabaseID: 187831},
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

	wantAccountIDs := []extsvc.AccountID{
		"57463526",
		"67471",
		"187831",
	}
	if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
		t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
	}
}
