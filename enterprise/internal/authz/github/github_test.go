package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gregjones/httpcache"

	gh "github.com/google/go-github/v48/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//nolint:unparam // unparam complains that `u` always has same value across call-sites, but that's OK
func mustURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func memGroupsCache() *cachedGroups {
	return &cachedGroups{cache: httpcache.NewMemoryCache()}
}

func mockClientFunc(mockClient client) func() (client, error) {
	return func() (client, error) {
		return mockClient, nil
	}
}

// newMockClientWithTokenMock is used to keep the behaviour of WithToken function mocking
// which is lost during moving the client interface to mockgen usage
func newMockClientWithTokenMock() *MockClient {
	mockClient := NewMockClient()
	mockClient.WithAuthenticatorFunc.SetDefaultReturn(mockClient)
	return mockClient
}

func TestProvider_FetchUserPerms(t *testing.T) {
	t.Run("nil account", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchUserPerms(context.Background(), nil, authz.FetchPermsOptions{})
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
			authz.FetchPermsOptions{},
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
			authz.FetchPermsOptions{},
		)
		want := `no token found in the external account data`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	var (
		authToken   = "my_access_token"
		authData    = json.RawMessage(fmt.Sprintf(`{"access_token": "%s"}`, authToken))
		mockAccount = &extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				AccountID:   "4567",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
			AccountData: extsvc.AccountData{
				AuthData: extsvc.NewUnencryptedData(authData),
			},
		}

		//nolint:unparam // Returning constant value for 'int' result is OK
		mockListAffiliatedRepositories = mock.WithRequestMatchPages(
			mock.GetUserRepos,
			[]*gh.Repository{
				{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=")},
				{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=")},
			},
			[]*gh.Repository{
				{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=")},
			},
		)

		mockOrgNoRead      = &github.OrgDetails{Org: github.Org{Login: "not-sourcegraph"}, DefaultRepositoryPermission: "none"}
		mockOrgRead        = &github.OrgDetails{Org: github.Org{Login: "sourcegraph"}, DefaultRepositoryPermission: "read"}
		newmockOrgNoRead   = &gh.Organization{Login: gh.String("not-sourcegraph"), DefaultRepoPermission: gh.String("none")}
		newmockOrgNoRead2  = &gh.Organization{Login: gh.String("not-sourcegraph-2"), DefaultRepoPermission: gh.String("none")}
		newmockOrgRead     = &gh.Organization{Login: gh.String("sourcegraph"), DefaultRepoPermission: gh.String("read")}
		mockListOrgDetails = mock.WithRequestMatchPages(
			mock.GetUserOrgs,
			[]*gh.Organization{
				newmockOrgNoRead,
				newmockOrgNoRead2,
			},
			[]*gh.Organization{newmockOrgRead})
		mockListOrgMembership = mock.WithRequestMatchHandler(
			mock.GetUserMembershipsOrgsByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), *newmockOrgNoRead2.Login) {
					w.Write(mock.MustMarshal(&gh.Membership{
						State: gh.String("active"),
						Role:  gh.String("admin"),
					}))
				}
			}),
		)

		//nolint:unparam // Returning constant value for 'int' result is OK
		mockListOrgRepositories = func(_ context.Context, org string, page int, _ string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
			switch org {
			case *newmockOrgRead.Login:
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
			case *newmockOrgNoRead2.Login:
				return []*github.Repository{{ID: "MDEwOlJlcG9zaXRvcnkyNDI2NTadmin="}}, false, 1, nil
			}
			t.Fatalf("unexpected call to ListOrgRepositories with org %q page %d", org, page)
			return nil, false, 1, nil
		}
		newMockListOrgRepositories = mock.WithRequestMatchHandler(
			mock.GetOrgsReposByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "/"+*(newmockOrgRead.Login)+"/") {
					switch r.URL.Query().Get("page") {
					case "":
						fallthrough
					case "1":
						addLinkHeader(t, w, 2)
						w.Write(mock.MustMarshal([]*gh.Repository{
							{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=")},
							{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234=")},
						}))
					case "2":
						w.Write(mock.MustMarshal([]*gh.Repository{
							{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDI2NTE5678=")},
						}))
					}
					return
				}
				if strings.Contains(r.URL.String(), "/"+*(newmockOrgNoRead2.Login)+"/") {
					w.Write(mock.MustMarshal([]*gh.Repository{
						{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDI2NTadmin=")},
					}))
					return
				}
				t.Fatalf("unexpected call to list org repositories with URL %q", r.URL.String())
			}),
		)
	)

	t.Run("cache disabled", func(t *testing.T) {
		mockHTTPClient := mock.NewMockedHTTPClient(
			mockListAffiliatedRepositories,
		)
		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCacheTTL: time.Duration(-1),
		})
		p.baseHTTPClient = mockHTTPClient
		if p.groupsCache != nil {
			t.Fatal("expected nil groupsCache")
		}

		repoIDs, err := p.FetchUserPerms(context.Background(),
			mockAccount,
			authz.FetchPermsOptions{},
		)
		if err != nil {
			t.Fatal(err)
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

	t.Run("cache enabled", func(t *testing.T) {
		t.Run("user has no orgs and teams", func(t *testing.T) {
			mockHTTPClient := mock.NewMockedHTTPClient(
				mockListAffiliatedRepositories,
				mock.WithRequestMatch(
					mock.GetUserOrgs,
					[]*gh.Organization{},
				),
				mock.WithRequestMatch(
					mock.GetUserTeams,
					[]*gh.Team{},
				),
			)
			assertClientCalledWithAuth(t, mockHTTPClient, authToken)

			p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
			p.baseHTTPClient = mockHTTPClient
			if p.groupsCache == nil {
				t.Fatal("expected groupsCache")
			}
			p.groupsCache = memGroupsCache()

			repoIDs, err := p.FetchUserPerms(context.Background(), mockAccount, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
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

		t.Run("user in orgs", func(t *testing.T) {
			p := setupProvider(t, nil)
			mockHTTPClient := mock.NewMockedHTTPClient(
				mockListAffiliatedRepositories,
				mockListOrgDetails,
				mockListOrgMembership,
				newMockListOrgRepositories,
				mock.WithRequestMatch(
					mock.GetUserTeams,
					[]*gh.Team{},
				),
			)
			p.baseHTTPClient = mockHTTPClient

			repoIDs, err := p.FetchUserPerms(context.Background(),
				mockAccount,
				authz.FetchPermsOptions{},
			)
			if err != nil {
				t.Fatal(err)
			}

			wantRepoIDs := []extsvc.RepoID{
				"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
				"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTadmin=",
				"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTE5678=",
			}
			if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
				t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run("user in orgs and teams", func(t *testing.T) {
			p := setupProvider(t, nil)
			mockHTTPClient := mock.NewMockedHTTPClient(
				mockListAffiliatedRepositories,
				mockListOrgDetails,
				mockListOrgMembership,
				newMockListOrgRepositories,
				mock.WithRequestMatchPages(
					mock.GetUserTeams,
					[]*gh.Team{
						{Organization: newmockOrgRead, Name: gh.String("ns team"), Slug: gh.String("ns-team")},
						{Organization: newmockOrgNoRead, Name: gh.String("ns team"), Slug: gh.String("ns-team"), ReposCount: gh.Int(0)},
					},
					[]*gh.Team{
						{Organization: newmockOrgNoRead, Name: gh.String("ns team 2"), Slug: gh.String("ns-team-2"), ReposCount: gh.Int(3)},
					},
				),
				mock.WithRequestMatchHandler(
					mock.GetOrgsTeamsReposByOrgByTeamSlug,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if strings.Contains(r.URL.String(), fmt.Sprintf("/%s/teams/%s/", *(newmockOrgNoRead.Login), "ns-team-2")) {
							switch r.URL.Query().Get("page") {
							case "":
								fallthrough
							case "1":
								addLinkHeader(t, w, 2)
								w.Write(mock.MustMarshal([]*gh.Repository{
									{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=")},
									{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDQ1nsteam1=")},
								}))
							case "2":
								w.Write(mock.MustMarshal([]*gh.Repository{
									{NodeID: gh.String("MDEwOlJlcG9zaXRvcnkyNDI2nsteam2=")},
								}))
							}
							return
						}
						t.Fatalf("unexpected call to list team repositories with url %q", r.URL.String())
					}),
				),
			)
			p.baseHTTPClient = mockHTTPClient

			repoIDs, err := p.FetchUserPerms(context.Background(),
				mockAccount,
				authz.FetchPermsOptions{InvalidateCaches: true},
			)
			if err != nil {
				t.Fatal(err)
			}

			wantRepoIDs := []extsvc.RepoID{
				"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
				"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTadmin=",
				"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTE5678=",
				"MDEwOlJlcG9zaXRvcnkyNDQ1nsteam1=",
				"MDEwOlJlcG9zaXRvcnkyNDI2nsteam2=",
			}
			if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
				t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
			}
		})

		makeStatusCodeTest := func(code int) func(t *testing.T) {
			return func(t *testing.T) {
				mockClient := newMockClientWithTokenMock()
				mockClient.GetAuthenticatedUserTeamsFunc.SetDefaultHook(
					func(_ context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
						switch page {
						case 1:
							return []*github.Team{
								// should not get repos from this team because parent org has default read permissions
								{Organization: &mockOrgRead.Org, Name: "ns team", Slug: "ns-team"},
								// should not get repos from this team since it has no repos
								{Organization: &mockOrgNoRead.Org, Name: "ns team", Slug: "ns-team", ReposCount: 0},
							}, true, 1, nil
						case 2:
							return []*github.Team{
								// should get repos from this team
								{Organization: &mockOrgNoRead.Org, Name: "ns team 2", Slug: "ns-team-2", ReposCount: 3},
							}, false, 1, nil
						}
						return nil, false, 1, nil
					})
				mockClient.ListOrgRepositoriesFunc.SetDefaultHook(mockListOrgRepositories)
				mockClient.ListTeamRepositoriesFunc.SetDefaultHook(
					func(_ context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
						return nil, false, 1, &github.APIError{Code: code}
					})

				p := setupProvider(t, mockClient)
				mockHTTPClient := mock.NewMockedHTTPClient(
					mockListAffiliatedRepositories,
					mockListOrgDetails,
				)
				p.baseHTTPClient = mockHTTPClient

				repoIDs, err := p.FetchUserPerms(context.Background(),
					mockAccount,
					authz.FetchPermsOptions{InvalidateCaches: true},
				)
				if err != nil {
					t.Fatal(err)
				}

				wantRepoIDs := []extsvc.RepoID{
					"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=", // from ListAffiliatedRepos
					"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=", // from ListAffiliatedRepos
					"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=", // from ListAffiliatedRepos
					"MDEwOlJlcG9zaXRvcnkyNDI2NTadmin=", // from ListOrgRepositories
					"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234=", // from ListOrgRepositories
					"MDEwOlJlcG9zaXRvcnkyNDI2NTE5678=", // from ListOrgRepositories
				}
				if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
					t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
				}
				_, found := p.groupsCache.getGroup("not-sourcegraph", "ns-team-2")
				if !found {
					t.Error("expected to find group in cache")
				}
			}
		}

		t.Run("special case: ListTeamRepositories returns 404", makeStatusCodeTest(404))
		t.Run("special case: ListTeamRepositories returns 403", makeStatusCodeTest(403))

		t.Run("cache and invalidate: user in orgs and teams", func(t *testing.T) {
			callsToListOrgRepos := 0
			callsToListTeamRepos := 0
			mockClient := newMockClientWithTokenMock()
			mockClient.GetAuthenticatedUserTeamsFunc.SetDefaultHook(
				func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
					return []*github.Team{
						{Organization: &mockOrgNoRead.Org, Name: "ns team 2", Slug: "ns-team-2", ReposCount: 3},
					}, false, 1, nil
				})
			mockClient.ListOrgRepositoriesFunc.SetDefaultHook(
				func(ctx context.Context, org string, page int, repoType string) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
					callsToListOrgRepos++
					return mockListOrgRepositories(ctx, org, page, repoType)
				})
			mockClient.ListTeamRepositoriesFunc.SetDefaultHook(
				func(_ context.Context, _, _ string, _ int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
					callsToListTeamRepos++
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zaXRvcnkyNDI2nsteam1="},
					}, false, 1, nil
				})

			p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
			p.client = mockClientFunc(mockClient)
			mockHTTPClient := mock.NewMockedHTTPClient(
				mockListAffiliatedRepositories,
				mockListOrgDetails,
			)
			p.baseHTTPClient = mockHTTPClient
			memCache := memGroupsCache()
			p.groupsCache = memCache

			wantRepoIDs := []extsvc.RepoID{
				"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE=",
				"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1MzY=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTEwMDA=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTadmin=",
				"MDEwOlJlcG9zaXRvcnkyNDQ1MTc1234=",
				"MDEwOlJlcG9zaXRvcnkyNDI2NTE5678=",
				"MDEwOlJlcG9zaXRvcnkyNDI2nsteam1=",
			}

			// first call
			t.Run("first call", func(t *testing.T) {
				repoIDs, err := p.FetchUserPerms(context.Background(),
					mockAccount,
					authz.FetchPermsOptions{},
				)
				if err != nil {
					t.Fatal(err)
				}
				if callsToListOrgRepos == 0 || callsToListTeamRepos == 0 {
					t.Fatalf("expected repos to be listed: callsToListOrgRepos=%d, callsToListTeamRepos=%d",
						callsToListOrgRepos, callsToListTeamRepos)
				}
				if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
					t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
				}
			})

			// second call should use cache
			t.Run("second call", func(t *testing.T) {
				callsToListOrgRepos = 0
				callsToListTeamRepos = 0
				repoIDs, err := p.FetchUserPerms(context.Background(),
					mockAccount,
					authz.FetchPermsOptions{InvalidateCaches: false},
				)
				if err != nil {
					t.Fatal(err)
				}
				if callsToListOrgRepos > 0 || callsToListTeamRepos > 0 {
					t.Fatalf("expected repos not to be listed: callsToListOrgRepos=%d, callsToListTeamRepos=%d",
						callsToListOrgRepos, callsToListTeamRepos)
				}
				if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
					t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
				}
			})

			// third call should make a fresh query when invalidating cache
			t.Run("third call", func(t *testing.T) {
				callsToListOrgRepos = 0
				callsToListTeamRepos = 0
				repoIDs, err := p.FetchUserPerms(context.Background(),
					mockAccount,
					authz.FetchPermsOptions{InvalidateCaches: true},
				)
				if err != nil {
					t.Fatal(err)
				}
				if callsToListOrgRepos == 0 || callsToListTeamRepos == 0 {
					t.Fatalf("expected repos to be listed: callsToListOrgRepos=%d, callsToListTeamRepos=%d",
						callsToListOrgRepos, callsToListTeamRepos)
				}
				if diff := cmp.Diff(wantRepoIDs, repoIDs.Exacts); diff != "" {
					t.Fatalf("RepoIDs mismatch (-want +got):\n%s", diff)
				}
			})
		})

		t.Run("cache partial update", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.GetAuthenticatedUserTeamsFunc.SetDefaultHook(
				func(ctx context.Context, page int) (teams []*github.Team, hasNextPage bool, rateLimitCost int, err error) {
					return []*github.Team{
						{Organization: &mockOrgNoRead.Org, Name: "ns team 2", Slug: "ns-team-2", ReposCount: 3},
					}, false, 1, nil
				})
			mockClient.ListOrgRepositoriesFunc.SetDefaultHook(mockListOrgRepositories)
			mockClient.ListTeamRepositoriesFunc.SetDefaultHook(
				func(ctx context.Context, org, team string, page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zaXRvcnkyNDI2nsteam1="},
					}, false, 1, nil
				})

			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockHTTPClient := mock.NewMockedHTTPClient(
				mockListAffiliatedRepositories,
				mockListOrgDetails,
			)
			p.baseHTTPClient = mockHTTPClient
			p.client = mockClientFunc(mockClient)
			memCache := memGroupsCache()
			p.groupsCache = memCache

			// cache populated from repo-centric sync (should add self)
			p.groupsCache.setGroup(cachedGroup{
				Org:          mockOrgRead.Login,
				Users:        []extsvc.AccountID{"1234"},
				Repositories: []extsvc.RepoID{},
			},
			)
			// cache populated from user-centric sync (should not add self)
			p.groupsCache.setGroup(cachedGroup{
				Org:          mockOrgNoRead.Login,
				Team:         "ns-team-2",
				Users:        []extsvc.AccountID{},
				Repositories: []extsvc.RepoID{"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE="},
			},
			)

			// run a sync
			_, err := p.FetchUserPerms(context.Background(),
				mockAccount,
				authz.FetchPermsOptions{InvalidateCaches: false},
			)
			if err != nil {
				t.Fatal(err)
			}

			// mock user should have added self to complete cache
			group, found := p.groupsCache.getGroup(mockOrgRead.Login, "")
			if !found {
				t.Fatal("expected group")
			}
			if len(group.Users) != 2 {
				t.Fatal("expected an additional user in partial cache group")
			}

			// mock user should not have added self to incomplete cache
			group, found = p.groupsCache.getGroup(mockOrgNoRead.Login, "ns-team-2")
			if !found {
				t.Fatal("expected group")
			}
			if len(group.Users) != 0 {
				t.Fatal("expected users not to be updated")
			}
		})
	})
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchRepoPerms(context.Background(), nil, authz.FetchPermsOptions{})
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
			authz.FetchPermsOptions{},
		)
		want := `not a code host of the repository: want "https://gitlab.com/" but have "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != want {
			t.Fatalf("err: want %q but got %q", want, got)
		}
	})

	var (
		mockUserRepo = extsvc.Repository{
			URI: "github.com/user/user-repo",
			ExternalRepoSpec: api.ExternalRepoSpec{
				ID:          "github_project_id",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
		}

		mockOrgRepo = extsvc.Repository{
			URI: "github.com/org/org-repo",
			ExternalRepoSpec: api.ExternalRepoSpec{
				ID:          "github_project_id",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
		}

		//nolint:unparam // Allow returning nil error on all code paths
		mockListCollaborators = mock.WithRequestMatchPages(
			mock.GetReposCollaboratorsByOwnerByRepo,
			[]*gh.User{
				{ID: gh.Int64(57463526)},
				{ID: gh.Int64(67471)},
			},
			[]*gh.User{{ID: gh.Int64(187831)}},
		)
	)

	t.Run("cache disabled", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCacheTTL: -1,
		})
		mockHTTPClient := mock.NewMockedHTTPClient(
			mockListCollaborators,
		)
		p.baseHTTPClient = mockHTTPClient

		accountIDs, err := p.FetchRepoPerms(context.Background(), &mockUserRepo,
			authz.FetchPermsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		wantAccountIDs := []extsvc.AccountID{
			// mockListCollaborators members
			"57463526",
			"67471",
			"187831",
		}
		if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
			t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("cache enabled", func(t *testing.T) {
		t.Run("repo not in org", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetOrgsByOrg,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if strings.Contains(r.URL.String(), "user") {
							w.WriteHeader(404)
							return
						}
						t.Fatalf("unexpected call to get organization url %q", r.URL.String())
					}),
				),
				mockListCollaborators,
			)
			p.baseHTTPClient = mockHTTPClient
			if p.groupsCache == nil {
				t.Fatal("expected groupsCache")
			}
			memCache := memGroupsCache()
			p.groupsCache = memCache

			accountIDs, err := p.FetchRepoPerms(context.Background(), &mockUserRepo,
				authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}

			wantAccountIDs := []extsvc.AccountID{
				// mockListCollaborators members
				"57463526",
				"67471",
				"187831",
			}
			if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
				t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run("repo in read org", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchPages(
					mock.GetOrgsByOrg,
					gh.Organization{DefaultRepoPermission: gh.String("read")},
				),
				mock.WithRequestMatchHandler(
					mock.GetOrgsMembersByOrg,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Query().Get("role") == "admin" {
							t.Fatal("unexpected admin only")
							return
						}
						switch r.URL.Query().Get("page") {
						case "":
							fallthrough
						case "1":
							addLinkHeader(t, w, 2)
							w.Write(mock.MustMarshal([]*gh.User{
								{ID: gh.Int64(1234)},
								{ID: gh.Int64(67471)}, // duplicate from collaborators
							}))
						case "2":
							w.Write(mock.MustMarshal([]*gh.User{{ID: gh.Int64(5678)}}))
						default:
							w.Write(mock.MustMarshal([]*gh.User{}))
						}
					})),
				mockListCollaborators,
			)
			p.baseHTTPClient = mockHTTPClient
			memCache := memGroupsCache()
			p.groupsCache = memCache

			accountIDs, err := p.FetchRepoPerms(context.Background(), &mockOrgRepo,
				authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}

			wantAccountIDs := []extsvc.AccountID{
				// mockListCollaborators members
				"57463526",
				"67471",
				"187831",
				// dedpulicated MockListOrganizationMembers users
				"1234",
				"5678",
			}
			if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
				t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run("internal repo in org", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchPages(
					mock.GetOrgsByOrg,
					gh.Organization{DefaultRepoPermission: gh.String("none")},
				),
				mock.WithRequestMatchHandler(
					mock.GetOrgsMembersByOrg,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Query().Get("role") == "admin" {
							w.Write(mock.MustMarshal([]*gh.User{{ID: gh.Int64(9999)}}))
							return
						}
						switch r.URL.Query().Get("page") {
						case "":
							fallthrough
						case "1":
							addLinkHeader(t, w, 2)
							w.Write(mock.MustMarshal([]*gh.User{
								{ID: gh.Int64(1234)},
								{ID: gh.Int64(67471)}, // duplicate from collaborators
							}))
						case "2":
							w.Write(mock.MustMarshal([]*gh.User{{ID: gh.Int64(5678)}}))
						default:
							w.Write(mock.MustMarshal([]*gh.User{}))
						}
					})),

				mock.WithRequestMatch(
					mock.GetReposTeamsByOwnerByRepo,
					[]gh.Team{},
				),
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					&gh.Repository{
						NodeID:     gh.String("github_repo_id"),
						Private:    gh.Bool(true),
						Visibility: gh.String("internal"),
					}),
				mockListCollaborators,
			)
			p.baseHTTPClient = mockHTTPClient

			// Ideally don't want a feature flag for this and want this internal repos to sync for
			// all users inside an org. Since we're introducing a new feature this is guarded behind
			// a feature flag, thus we also test against it. Once we're reasonably sure this works
			// as intended, we will remove the feature flag and enable the behaviour by default.
			t.Run("feature flag disabled", func(t *testing.T) {
				p.enableGithubInternalRepoVisibility = false

				memCache := memGroupsCache()
				p.groupsCache = memCache

				accountIDs, err := p.FetchRepoPerms(
					context.Background(), &mockOrgRepo, authz.FetchPermsOptions{},
				)
				if err != nil {
					t.Fatal(err)
				}

				// These account IDs will have access to the internal repo.
				wantAccountIDs := []extsvc.AccountID{
					// expect mockListCollaborators members only - we do not want to include org members
					// if internal repository support is not enabled.
					"57463526",
					"67471",
					"187831",
					// The admin is expected to be in this list.
					"9999",
				}
				if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
					t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
				}
			})

			t.Run("feature flag enabled", func(t *testing.T) {
				p.enableGithubInternalRepoVisibility = true
				memCache := memGroupsCache()
				p.groupsCache = memCache

				accountIDs, err := p.FetchRepoPerms(
					context.Background(), &mockOrgRepo, authz.FetchPermsOptions{},
				)
				if err != nil {
					t.Fatal(err)
				}

				// These account IDs will have access to the internal repo.
				wantAccountIDs := []extsvc.AccountID{
					// mockListCollaborators members.
					"57463526",
					"67471",
					"187831",
					// expect dedpulicated MockListOrganizationMembers users as well since we want to grant access
					// to org members as well if the target repo has visibility "internal"
					"1234",
					"5678",
				}
				if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
					t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
				}
			})
		})

		t.Run("repo in non-read org but in teams", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchPages(
					mock.GetOrgsByOrg,
					gh.Organization{
						DefaultRepoPermission: gh.String("none"),
					},
				),
				mock.WithRequestMatch(
					mock.GetOrgsMembersByOrg,
					[]gh.User{{ID: gh.Int64(3456)}},
				),
				mock.WithRequestMatchPages(
					mock.GetReposTeamsByOwnerByRepo,
					[]gh.Team{{Slug: gh.String("team1")}},
					[]gh.Team{{Slug: gh.String("team2")}},
				),
				mock.WithRequestMatchHandler(
					mock.GetOrgsTeamsMembersByOrgByTeamSlug,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if strings.Contains(r.URL.String(), "team1") {
							w.Write(mock.MustMarshal([]gh.User{
								{ID: gh.Int64(1234)},
								{ID: gh.Int64(5678)},
							}))
							return
						}
						if strings.Contains(r.URL.String(), "team2") {
							w.Write(mock.MustMarshal([]gh.User{
								{ID: gh.Int64(1234)},
								{ID: gh.Int64(6789)},
							}))
							return
						}
						mock.WriteError(w, http.StatusNotFound, "team not found")
					}),
				),
				mockListCollaborators,
			)
			p.baseHTTPClient = mockHTTPClient
			memCache := memGroupsCache()
			p.groupsCache = memCache

			accountIDs, err := p.FetchRepoPerms(context.Background(), &mockOrgRepo,
				authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}

			wantAccountIDs := []extsvc.AccountID{
				// mockListCollaborators members
				"57463526",
				"67471",
				"187831",
				// MockListOrganizationMembers users
				"3456",
				// deduplicated MockListTeamMembers users
				"1234",
				"5678",
				"6789",
			}
			if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
				t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run("cache and invalidate", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			callsToListOrgMembers := 0
			mockHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatchPages(
					mock.GetOrgsByOrg,
					gh.Organization{DefaultRepoPermission: gh.String("read")},
				),
				mock.WithRequestMatchHandler(
					mock.GetOrgsMembersByOrg,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						callsToListOrgMembers++
						switch r.URL.Query().Get("page") {
						case "":
							fallthrough
						case "1":
							addLinkHeader(t, w, 2)
							w.Write(mock.MustMarshal([]gh.User{
								{ID: gh.Int64(1234)},
							}))
						case "2":
							w.Write(mock.MustMarshal([]gh.User{{ID: gh.Int64(5678)}}))
						}
					}),
				),
				mockListCollaborators,
			)
			p.baseHTTPClient = mockHTTPClient
			memCache := memGroupsCache()
			p.groupsCache = memCache

			wantAccountIDs := []extsvc.AccountID{
				// mockListCollaborators members
				"57463526",
				"67471",
				"187831",
				// MockListOrganizationMembers users
				"1234",
				"5678",
			}

			// first call
			t.Run("first call", func(t *testing.T) {
				accountIDs, err := p.FetchRepoPerms(context.Background(), &mockOrgRepo,
					authz.FetchPermsOptions{})
				if err != nil {
					t.Fatal(err)
				}
				if callsToListOrgMembers == 0 {
					t.Fatalf("expected members to be listed: callsToListOrgMembers=%d",
						callsToListOrgMembers)
				}
				if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
					t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
				}
			})

			// second call should use cache
			t.Run("second call", func(t *testing.T) {
				callsToListOrgMembers = 0
				accountIDs, err := p.FetchRepoPerms(context.Background(), &mockOrgRepo,
					authz.FetchPermsOptions{})
				if err != nil {
					t.Fatal(err)
				}
				if callsToListOrgMembers > 0 {
					t.Fatalf("expected members not to be listed: callsToListOrgMembers=%d",
						callsToListOrgMembers)
				}
				if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
					t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
				}
			})

			// third call should make a fresh query when invalidating cache
			t.Run("third call", func(t *testing.T) {
				callsToListOrgMembers = 0
				accountIDs, err := p.FetchRepoPerms(context.Background(), &mockOrgRepo,
					authz.FetchPermsOptions{InvalidateCaches: true})
				if err != nil {
					t.Fatal(err)
				}
				if callsToListOrgMembers == 0 {
					t.Fatalf("expected members to be listed: callsToListOrgMembers=%d",
						callsToListOrgMembers)
				}
				if diff := cmp.Diff(wantAccountIDs, accountIDs); diff != "" {
					t.Fatalf("AccountIDs mismatch (-want +got):\n%s", diff)
				}
			})
		})

		t.Run("cache partial update", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockHTTPClient := mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetOrgsByOrg,
					gh.Organization{
						DefaultRepoPermission: gh.String("none"),
					},
				),
				mock.WithRequestMatch(
					mock.GetReposTeamsByOwnerByRepo,
					[]gh.Team{
						{Slug: gh.String("team1")},
						{Slug: gh.String("team2")},
					},
				),
				mock.WithRequestMatchHandler(
					mock.GetOrgsTeamsMembersByOrgByTeamSlug,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if strings.Contains(r.URL.String(), "team1") {
							w.Write(mock.MustMarshal([]gh.User{
								{ID: gh.Int64(5678)},
							}))
							return
						}
						if strings.Contains(r.URL.String(), "team2") {
							w.Write(mock.MustMarshal([]gh.User{
								{ID: gh.Int64(6789)},
							}))
							return
						}
						mock.WriteError(w, http.StatusNotFound, "team not found")
					}),
				),
				mockListCollaborators,
				mock.WithRequestMatch(
					mock.GetOrgsMembersByOrg,
					[]*gh.User{},
				),
			)
			p.baseHTTPClient = mockHTTPClient
			memCache := memGroupsCache()
			p.groupsCache = memCache

			// cache populated from user-centric sync (should add self)
			p.groupsCache.setGroup(cachedGroup{
				Org:          "org",
				Team:         "team1",
				Users:        []extsvc.AccountID{},
				Repositories: []extsvc.RepoID{"MDEwOlJlcG9zaXRvcnkyNTI0MjU2NzE="},
			},
			)
			// cache populated from repo-centric sync (should not add self)
			p.groupsCache.setGroup(cachedGroup{
				Org:          "org",
				Team:         "team2",
				Users:        []extsvc.AccountID{"1234"},
				Repositories: []extsvc.RepoID{},
			},
			)

			// run a sync
			_, err := p.FetchRepoPerms(context.Background(),
				&mockOrgRepo,
				authz.FetchPermsOptions{InvalidateCaches: false},
			)
			if err != nil {
				t.Fatal(err)
			}

			// mock user should have added self to complete cache
			group, found := p.groupsCache.getGroup("org", "team1")
			if !found {
				t.Fatal("expected group")
			}
			if len(group.Repositories) != 2 {
				t.Fatal("expected an additional repo in partial cache group")
			}

			// mock user should not have added self to incomplete cache
			group, found = p.groupsCache.getGroup("org", "team2")
			if !found {
				t.Fatal("expected group")
			}
			if len(group.Repositories) != 0 {
				t.Fatal("expected repos not to be updated")
			}
		})
	})
}

func TestProvider_ValidateConnection(t *testing.T) {
	t.Run("cache disabled: scopes ok", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCacheTTL: -1,
		})
		problems := p.ValidateConnection(context.Background())
		if len(problems) > 0 {
			t.Fatal("expected validate to pass")
		}
	})

	t.Run("cache enabled", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCacheTTL: 72,
		})

		t.Run("error getting scopes", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.GetAuthenticatedOAuthScopesFunc.SetDefaultHook(
				func(ctx context.Context) ([]string, error) {
					return nil, errors.New("scopes error")
				})
			p.client = mockClientFunc(mockClient)
			problems := p.ValidateConnection(context.Background())
			if len(problems) != 1 {
				t.Fatal("expected 1 problem")
			}
			if !strings.Contains(problems[0], "scopes error") {
				t.Fatalf("unexpected problem: %q", problems[0])
			}
		})

		t.Run("missing org scope", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.GetAuthenticatedOAuthScopesFunc.SetDefaultHook(
				func(ctx context.Context) ([]string, error) {
					return []string{}, nil
				})
			p.client = mockClientFunc(mockClient)
			problems := p.ValidateConnection(context.Background())
			if len(problems) != 1 {
				t.Fatal("expected 1 problem")
			}
			if !strings.Contains(problems[0], "read:org") {
				t.Fatalf("unexpected problem: %q", problems[0])
			}
		})

		t.Run("scopes ok org scope", func(t *testing.T) {
			for _, testCase := range [][]string{
				{"read:org"},
				{"write:org"},
				{"admin:org"},
			} {
				mockClient := newMockClientWithTokenMock()
				mockClient.GetAuthenticatedOAuthScopesFunc.SetDefaultHook(
					func(ctx context.Context) ([]string, error) {
						return testCase, nil
					})
				p.client = mockClientFunc(mockClient)
				problems := p.ValidateConnection(context.Background())
				if len(problems) != 0 {
					t.Fatalf("expected validate to pass for scopes=%+v", testCase)
				}
			}
		})
	})
}

func setupProvider(t *testing.T, mc *MockClient) *Provider {
	p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
	p.client = mockClientFunc(mc)
	p.groupsCache = memGroupsCache()
	return p
}

func addLinkHeader(t *testing.T, w http.ResponseWriter, nextPage int) {
	t.Helper()
	w.Header().Add("Link", fmt.Sprintf(`<https://api.github.com/orgs/org/members?page=%d>; rel="next"`, nextPage))
}

type confirmAuthTransport struct {
	t     *testing.T
	base  http.RoundTripper
	token string
}

func (tr *confirmAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr.t.Helper()
	want := fmt.Sprintf("Bearer %s", tr.token)
	got := r.Header.Get("Authorization")
	if got != want {
		tr.t.Fatalf("Incorrect token got %s expected %s", got, want)
	}
	return tr.base.RoundTrip(r)
}

func assertClientCalledWithAuth(t *testing.T, client *http.Client, token string) {
	t.Helper()
	confirmAuth := &confirmAuthTransport{
		t:     t,
		base:  client.Transport,
		token: token,
	}
	client.Transport = confirmAuth
}
