package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func newTestClient(t *testing.T, cli httpcli.Doer) *V3Client {
	return newTestClientWithAuthenticator(t, nil, cli)
}

func newTestClientWithAuthenticator(t *testing.T, auth auth.Authenticator, cli httpcli.Doer) *V3Client {
	rcache.SetupForTest(t)

	apiURL := &url.URL{Scheme: "https", Host: "example.com", Path: "/"}
	return NewV3Client(logtest.Scoped(t), "Test", apiURL, auth, cli)
}

func TestListAffiliatedRepositories(t *testing.T) {
	tests := []struct {
		name         string
		visibility   Visibility
		affiliations []RepositoryAffiliation
		wantRepos    []*Repository
	}{
		{
			name:       "list all repositories",
			visibility: VisibilityAll,
			wantRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzQxNTE=",
					DatabaseID:       263034151,
					NameWithOwner:    "sourcegraph-vcr-repos/private-org-repo-1",
					URL:              "https://github.com/sourcegraph-vcr-repos/private-org-repo-1",
					IsPrivate:        true,
					ViewerPermission: "ADMIN",
				}, {
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzQwNzM=",
					DatabaseID:       263034073,
					NameWithOwner:    "sourcegraph-vcr/private-user-repo-1",
					URL:              "https://github.com/sourcegraph-vcr/private-user-repo-1",
					IsPrivate:        true,
					ViewerPermission: "ADMIN",
				}, {
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzM5NDk=",
					DatabaseID:       263033949,
					NameWithOwner:    "sourcegraph-vcr/public-user-repo-1",
					URL:              "https://github.com/sourcegraph-vcr/public-user-repo-1",
					ViewerPermission: "ADMIN",
				}, {
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzM3NjE=",
					DatabaseID:       263033761,
					NameWithOwner:    "sourcegraph-vcr-repos/public-org-repo-1",
					URL:              "https://github.com/sourcegraph-vcr-repos/public-org-repo-1",
					ViewerPermission: "ADMIN",
				},
			},
		},
		{
			name:       "list public repositories",
			visibility: VisibilityPublic,
			wantRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzM5NDk=",
					DatabaseID:       263033949,
					NameWithOwner:    "sourcegraph-vcr/public-user-repo-1",
					URL:              "https://github.com/sourcegraph-vcr/public-user-repo-1",
					ViewerPermission: "ADMIN",
				}, {
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzM3NjE=",
					DatabaseID:       263033761,
					NameWithOwner:    "sourcegraph-vcr-repos/public-org-repo-1",
					URL:              "https://github.com/sourcegraph-vcr-repos/public-org-repo-1",
					ViewerPermission: "ADMIN",
				},
			},
		},
		{
			name:       "list private repositories",
			visibility: VisibilityPrivate,
			wantRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzQxNTE=",
					DatabaseID:       263034151,
					NameWithOwner:    "sourcegraph-vcr-repos/private-org-repo-1",
					URL:              "https://github.com/sourcegraph-vcr-repos/private-org-repo-1",
					IsPrivate:        true,
					ViewerPermission: "ADMIN",
				}, {
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzQwNzM=",
					DatabaseID:       263034073,
					NameWithOwner:    "sourcegraph-vcr/private-user-repo-1",
					URL:              "https://github.com/sourcegraph-vcr/private-user-repo-1",
					IsPrivate:        true,
					ViewerPermission: "ADMIN",
				},
			},
		},
		{
			name:         "list collaborator and owner affiliated repositories",
			affiliations: []RepositoryAffiliation{AffiliationCollaborator, AffiliationOwner},
			wantRepos: []*Repository{
				{
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzQwNzM=",
					DatabaseID:       263034073,
					NameWithOwner:    "sourcegraph-vcr/private-user-repo-1",
					URL:              "https://github.com/sourcegraph-vcr/private-user-repo-1",
					IsPrivate:        true,
					ViewerPermission: "ADMIN",
				}, {
					ID:               "MDEwOlJlcG9zaXRvcnkyNjMwMzM5NDk=",
					DatabaseID:       263033949,
					NameWithOwner:    "sourcegraph-vcr/public-user-repo-1",
					URL:              "https://github.com/sourcegraph-vcr/public-user-repo-1",
					ViewerPermission: "ADMIN",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, save := newV3TestClient(t, "ListAffiliatedRepositories_"+test.name)
			defer save()

			repos, _, _, err := client.ListAffiliatedRepositories(context.Background(), test.visibility, 1, test.affiliations...)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantRepos, repos); diff != "" {
				t.Fatalf("Repos mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_GetAuthenticatedOAuthScopes(t *testing.T) {
	client, save := newV3TestClient(t, "GetAuthenticatedOAuthScopes")
	defer save()

	scopes, err := client.GetAuthenticatedOAuthScopes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"admin:enterprise", "admin:gpg_key", "admin:org", "admin:org_hook", "admin:public_key", "admin:repo_hook", "delete:packages", "delete_repo", "gist", "notifications", "repo", "user", "workflow", "write:discussion", "write:packages"}
	sort.Strings(scopes)
	if diff := cmp.Diff(want, scopes); diff != "" {
		t.Fatalf("Scopes mismatch (-want +got):\n%s", diff)
	}
}

// NOTE: To update VCR for this test, please use the token of "sourcegraph-vcr"
// for GITHUB_TOKEN, which can be found in 1Password.
func TestListRepositoryCollaborators(t *testing.T) {
	tests := []struct {
		name            string
		owner           string
		repo            string
		affiliation     CollaboratorAffiliation
		wantUsers       []*Collaborator
		wantHasNextPage bool
	}{
		{
			name:  "public repo",
			owner: "sourcegraph-vcr-repos",
			repo:  "public-org-repo-1",
			wantUsers: []*Collaborator{
				{
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegraph-vcr as owner
					DatabaseID: 63290851,
				},
			},
			wantHasNextPage: false,
		},
		{
			name:  "private repo",
			owner: "sourcegraph-vcr-repos",
			repo:  "private-org-repo-1",
			wantUsers: []*Collaborator{
				{
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegraph-vcr as owner
					DatabaseID: 63290851,
				}, {
					ID:         "MDQ6VXNlcjY2NDY0Nzcz", // sourcegraph-vcr-amy as team member
					DatabaseID: 66464773,
				}, {
					ID:         "MDQ6VXNlcjY2NDY0OTI2", // sourcegraph-vcr-bob as outside collaborator
					DatabaseID: 66464926,
				}, {
					ID:         "MDQ6VXNlcjg5NDk0ODg0", // sourcegraph-vcr-dave as team member
					DatabaseID: 89494884,
				},
			},
			wantHasNextPage: false,
		},
		{
			name:        "direct collaborator outside collaborator",
			owner:       "sourcegraph-vcr-repos",
			repo:        "private-org-repo-1",
			affiliation: AffiliationDirect,
			wantUsers: []*Collaborator{
				{
					ID:         "MDQ6VXNlcjY2NDY0OTI2", // sourcegraph-vcr-bob as outside collaborator
					DatabaseID: 66464926,
				},
			},
			wantHasNextPage: false,
		},
		{
			name:        "direct collaborator repo owner",
			owner:       "sourcegraph-vcr",
			repo:        "public-user-repo-1",
			affiliation: AffiliationDirect,
			wantUsers: []*Collaborator{
				{
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegraph-vcr as owner
					DatabaseID: 63290851,
				},
			},
			wantHasNextPage: false,
		},
		{
			name:            "has next page is true",
			owner:           "sourcegraph-vcr",
			repo:            "private-repo-1",
			affiliation:     AffiliationDirect,
			wantUsers:       nil,
			wantHasNextPage: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, save := newV3TestClient(t, "ListRepositoryCollaborators_"+test.name)
			defer save()

			users, hasNextPage, err := client.ListRepositoryCollaborators(context.Background(), test.owner, test.repo, 1, test.affiliation)
			if err != nil {
				t.Fatal(err)
			}

			if test.wantUsers != nil {
				if diff := cmp.Diff(test.wantUsers, users); diff != "" {
					t.Fatalf("Users mismatch (-want +got):\n%s", diff)
				}
			}

			if diff := cmp.Diff(test.wantHasNextPage, hasNextPage); diff != "" {
				t.Fatalf("HasNextPage mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetAuthenticatedUserOrgs(t *testing.T) {
	cli, save := newV3TestClient(t, "GetAuthenticatedUserOrgs")
	defer save()

	ctx := context.Background()
	orgs, _, _, err := cli.GetAuthenticatedUserOrgsForPage(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t,
		"testdata/golden/GetAuthenticatedUserOrgs",
		update("GetAuthenticatedUserOrgs"),
		orgs,
	)
}

func TestGetAuthenticatedUserOrgDetailsAndMembership(t *testing.T) {
	cli, save := newV3TestClient(t, "GetAuthenticatedUserOrgDetailsAndMembership")
	defer save()

	ctx := context.Background()
	var err error
	orgs := make([]OrgDetailsAndMembership, 0)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var pageOrgs []OrgDetailsAndMembership
		pageOrgs, hasNextPage, _, err = cli.GetAuthenticatedUserOrgsDetailsAndMembership(ctx, page)
		if err != nil {
			t.Fatal(err)
		}
		orgs = append(orgs, pageOrgs...)
	}

	for _, org := range orgs {
		if org.OrgDetails == nil {
			t.Fatal("expected org details, got nil")
		}
		if org.OrgDetails.DefaultRepositoryPermission == "" {
			t.Fatal("expected default repo permissions data")
		}
		if org.OrgMembership == nil {
			t.Fatal("expected org membership, got nil")
		}
		if org.OrgMembership.Role == "" {
			t.Fatal("expected org membership data")
		}
	}

	testutil.AssertGolden(t,
		"testdata/golden/GetAuthenticatedUserOrgDetailsAndMembership",
		update("GetAuthenticatedUserOrgDetailsAndMembership"),
		orgs,
	)
}

func TestListOrgRepositories(t *testing.T) {
	cli, save := newV3TestClient(t, "ListOrgRepositories")
	defer save()

	ctx := context.Background()
	var err error
	repos := make([]*Repository, 0)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var pageRepos []*Repository
		pageRepos, hasNextPage, _, err = cli.ListOrgRepositories(ctx, "sourcegraph-vcr-repos", page, "")
		if err != nil {
			t.Fatal(err)
		}
		repos = append(repos, pageRepos...)
	}

	testutil.AssertGolden(t,
		"testdata/golden/ListOrgRepositories",
		update("ListOrgRepositories"),
		repos,
	)
}

func TestListTeamRepositories(t *testing.T) {
	cli, save := newV3TestClient(t, "ListTeamRepositories")
	defer save()

	ctx := context.Background()
	var err error
	repos := make([]*Repository, 0)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var pageRepos []*Repository
		pageRepos, hasNextPage, _, err = cli.ListTeamRepositories(ctx, "sourcegraph-vcr-repos", "private-access", page)
		if err != nil {
			t.Fatal(err)
		}
		repos = append(repos, pageRepos...)
	}

	testutil.AssertGolden(t,
		"testdata/golden/ListTeamRepositories",
		update("ListTeamRepositories"),
		repos,
	)
}

func TestGetAuthenticatedUserTeams(t *testing.T) {
	cli, save := newV3TestClient(t, "GetAuthenticatedUserTeams")
	defer save()

	ctx := context.Background()
	var err error
	teams := make([]*Team, 0)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var pageTeams []*Team
		pageTeams, hasNextPage, _, err = cli.GetAuthenticatedUserTeams(ctx, page)
		if err != nil {
			t.Fatal(err)
		}
		teams = append(teams, pageTeams...)
	}

	testutil.AssertGolden(t,
		"testdata/golden/GetAuthenticatedUserTeams",
		update("GetAuthenticatedUserTeams"),
		teams,
	)
}

func TestListRepositoryTeams(t *testing.T) {
	cli, save := newV3TestClient(t, "ListRepositoryTeams")
	defer save()

	ctx := context.Background()
	var err error
	teams := make([]*Team, 0)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var pageTeams []*Team
		pageTeams, hasNextPage, err = cli.ListRepositoryTeams(ctx, "sourcegraph-vcr-repos", "private-org-repo-1", page)
		if err != nil {
			t.Fatal(err)
		}
		teams = append(teams, pageTeams...)
	}

	testutil.AssertGolden(t,
		"testdata/golden/ListRepositoryTeams",
		update("ListRepositoryTeams"),
		teams,
	)
}

func TestGetOrganization(t *testing.T) {
	cli, save := newV3TestClient(t, "GetOrganization")
	defer save()

	t.Run("real org", func(t *testing.T) {
		ctx := context.Background()
		org, err := cli.GetOrganization(ctx, "sourcegraph")
		if err != nil {
			t.Fatal(err)
		}
		if org == nil {
			t.Fatal("expected org, got nil")
		}
		if org.Login != "sourcegraph" {
			t.Fatalf("expected org 'sourcegraph', got %+v", org)
		}
	})

	t.Run("actually an user", func(t *testing.T) {
		ctx := context.Background()
		_, err := cli.GetOrganization(ctx, "sourcegraph-vcr")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Fatalf("expected not found, got %q", err.Error())
		}
	})
}

func TestGetRepository(t *testing.T) {
	rcache.SetupForTest(t)

	cli, save := newV3TestClient(t, "GetRepository")
	defer save()

	t.Run("cached-response", func(t *testing.T) {
		var remaining int

		t.Run("first run", func(t *testing.T) {
			repo, err := cli.GetRepository(context.Background(), "sourcegraph", "sourcegraph")
			if err != nil {
				t.Fatal(err)
			}

			if repo == nil {
				t.Fatal("expected repo, but got nil")
			}

			want := "sourcegraph/sourcegraph"
			if repo.NameWithOwner != want {
				t.Fatalf("expected NameWithOwner %s, but got %s", want, repo.NameWithOwner)
			}

			remaining, _, _, _ = cli.RateLimitMonitor().Get()
		})

		t.Run("second run", func(t *testing.T) {
			repo, err := cli.GetRepository(context.Background(), "sourcegraph", "sourcegraph")
			if err != nil {
				t.Fatal(err)
			}

			if repo == nil {
				t.Fatal("expected repo, but got nil")
			}

			want := "sourcegraph/sourcegraph"
			if repo.NameWithOwner != want {
				t.Fatalf("expected NameWithOwner %s, but got %s", want, repo.NameWithOwner)
			}

			remaining2, _, _, _ := cli.RateLimitMonitor().Get()
			if remaining2 < remaining {
				t.Fatalf("expected cached repsonse, but API quota used")
			}
		})
	})

	t.Run("repo not found", func(t *testing.T) {
		repo, err := cli.GetRepository(context.Background(), "owner", "repo")
		if !IsNotFound(err) {
			t.Errorf("got err == %v, want IsNotFound(err) == true", err)
		}
		if err != ErrRepoNotFound {
			t.Errorf("got err == %q, want ErrNotFound", err)
		}
		if repo != nil {
			t.Error("repo != nil")
		}
	})
}

// ListOrganizations is primarily used for GitHub Enterprise clients. As a result we test against
// ghe.sgdev.org.  To update this test, access the GitHub Enterprise Admin Account (ghe.sgdev.org)
// with username milton in 1password. The token used for this test is named sourcegraph-vcr-token
// and is also saved in 1Password under this account.
func TestListOrganizations(t *testing.T) {
	// Note: Testing against enterprise does not return the x-rate-remaining header at the moment,
	// as a result it is not possible to assert the remaining API calls after each APi request the
	// way we do in TestGetRepository.
	t.Run("enterprise-integration-cached-response", func(t *testing.T) {
		rcache.SetupForTest(t)

		cli, save := newV3TestEnterpriseClient(t, "ListOrganizations")
		defer save()

		t.Run("first run", func(t *testing.T) {
			orgs, nextSince, err := cli.ListOrganizations(context.Background(), 0)
			if err != nil {
				t.Fatal(err)
			}

			if orgs == nil {
				t.Fatal("expected orgs but got nil")
			}

			if len(orgs) != 100 {
				t.Fatalf("expected 100 orgs but got %d", len(orgs))
			}

			if nextSince < 1 {
				t.Fatalf("expected nextSince to be a positive int but got %v", nextSince)
			}
		})

		t.Run("second run", func(t *testing.T) {
			// Make the same API call again. This should hit the cache.
			orgs, nextSince, err := cli.ListOrganizations(context.Background(), 0)
			if err != nil {
				t.Fatal(err)
			}

			if orgs == nil {
				t.Fatal("expected orgs but got nil")
			}

			if len(orgs) != 100 {
				t.Fatalf("expected 100 orgs but got %d", len(orgs))
			}

			if nextSince < 1 {
				t.Fatalf("expected nextSince to be a positive int but got %v", nextSince)
			}
		})
	})

	t.Run("enterprise-pagination", func(t *testing.T) {
		rcache.SetupForTest(t)

		mockOrgs := make([]*Org, 200)

		for i := 0; i < 200; i++ {
			mockOrgs[i] = &Org{
				ID:    i + 1,
				Login: fmt.Sprint("foo-", i+1),
			}
		}

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val, ok := r.URL.Query()["since"]
			if !ok {
				t.Fatal(`unexpected test scenario, no query parameter "since"`)
			}

			writeJson := func(orgs []*Org) {
				data, err := json.Marshal(orgs)
				if err != nil {
					t.Fatalf("failed to marshal orgs into json: %v", err)
				}

				_, err = w.Write(data)
				if err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
			}

			switch val[0] {
			case "0":
				writeJson(mockOrgs[0:100])
			case "100":
				writeJson(mockOrgs[100:])
			case "200":
				writeJson([]*Org{})
			}
		}))

		uri, _ := url.Parse(testServer.URL)
		testCli := NewV3Client(logtest.Scoped(t), "Test", uri, gheToken, testServer.Client())

		runTest := func(since int, expectedNextSince int, expectedOrgs []*Org) {
			orgs, nextSince, err := testCli.ListOrganizations(context.Background(), since)
			if err != nil {
				t.Fatal(err)
			}
			if nextSince != expectedNextSince {
				t.Fatalf("expected nextSince: %d but got %d", nextSince, expectedNextSince)
			}

			if diff := cmp.Diff(expectedOrgs, orgs); diff != "" {
				t.Fatalf("mismatch in expected orgs and orgs received in response: (-want +got):\n%s", diff)
			}
		}

		t.Run("orgs 0 to 100", func(t *testing.T) {
			runTest(0, 100, mockOrgs[:100])
		})

		t.Run("orgs 100 to 200", func(t *testing.T) {
			runTest(100, 200, mockOrgs[100:])
		})

		t.Run("orgs out of bounds", func(t *testing.T) {
			runTest(200, -1, []*Org{})
		})
	})
}

func TestListMembers(t *testing.T) {
	tests := []struct {
		name        string
		fn          func(*V3Client) ([]*Collaborator, error)
		wantMembers []*Collaborator
	}{{
		name: "org members",
		fn: func(cli *V3Client) ([]*Collaborator, error) {
			members, _, err := cli.ListOrganizationMembers(context.Background(), "sourcegraph-vcr-repos", 1, false)
			return members, err
		},
		wantMembers: []*Collaborator{
			{ID: "MDQ6VXNlcjYzMjkwODUx", DatabaseID: 63290851}, // sourcegraph-vcr as owner
			{ID: "MDQ6VXNlcjY2NDY0Nzcz", DatabaseID: 66464773}, // sourcegraph-vcr-amy
			{ID: "MDQ6VXNlcjg5NDk0ODg0", DatabaseID: 89494884}, // sourcegraph-vcr-dave
		},
	}, {
		name: "org admins",
		fn: func(cli *V3Client) ([]*Collaborator, error) {
			members, _, err := cli.ListOrganizationMembers(context.Background(), "sourcegraph-vcr-repos", 1, true)
			return members, err
		},
		wantMembers: []*Collaborator{
			{ID: "MDQ6VXNlcjYzMjkwODUx", DatabaseID: 63290851}, // sourcegraph-vcr as owner
		},
	}, {
		name: "team members",
		fn: func(cli *V3Client) ([]*Collaborator, error) {
			members, _, err := cli.ListTeamMembers(context.Background(), "sourcegraph-vcr-repos", "private-access", 1)
			return members, err
		},
		wantMembers: []*Collaborator{
			{ID: "MDQ6VXNlcjYzMjkwODUx", DatabaseID: 63290851}, // sourcegraph-vcr
			{ID: "MDQ6VXNlcjY2NDY0Nzcz", DatabaseID: 66464773}, // sourcegraph-vcr-amy
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli, save := newV3TestClient(t, t.Name())
			defer save()

			members, err := test.fn(cli)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantMembers, members); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestV3Client_WithAuthenticator(t *testing.T) {
	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	oldClient := &V3Client{
		log:    logtest.Scoped(t),
		apiURL: uri,
		auth:   &auth.OAuthBearerToken{Token: "old_token"},
	}

	newToken := &auth.OAuthBearerToken{Token: "new_token"}
	newClient := oldClient.WithAuthenticator(newToken)
	if oldClient == newClient {
		t.Fatal("both clients have the same address")
	}

	if newClient.auth != newToken {
		t.Fatalf("token: want %p but got %p", newToken, newClient.auth)
	}
}

func TestV3Client_Fork(t *testing.T) {
	ctx := context.Background()
	testName := func(t *testing.T) string {
		return strings.ReplaceAll(t.Name(), "/", "_")
	}

	t.Run("success", func(t *testing.T) {
		// For this test, we only need a repository that can be forked into the
		// user's namespace and sourcegraph-testing: it doesn't matter whether it
		// already has been or not because of the way the GitHub API operates.
		// We'll use github.com/sourcegraph/automation-testing as our guinea pig.
		for name, org := range map[string]*string{
			"user":                nil,
			"sourcegraph-testing": strPtr("sourcegraph-testing"),
		} {
			t.Run(name, func(t *testing.T) {
				testName := testName(t)
				client, save := newV3TestClient(t, testName)
				defer save()

				fork, err := client.Fork(ctx, "sourcegraph", "automation-testing", org, "sourcegraph-automation-testing")
				assert.Nil(t, err)
				assert.NotNil(t, fork)
				if org != nil {
					owner, err := fork.Owner()
					assert.Nil(t, err)
					assert.Equal(t, *org, owner)
				}

				testutil.AssertGolden(t, filepath.Join("testdata", "golden", testName), update(testName), fork)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		// For this test, we need a repository that cannot be forked. Conveniently,
		// we have one at github.com/sourcegraph-testing/unforkable.
		testName := testName(t)
		client, save := newV3TestClient(t, testName)
		defer save()

		fork, err := client.Fork(ctx, "sourcegraph-testing", "unforkable", nil, "sourcegraph-testing-unforkable")
		assert.NotNil(t, err)
		assert.Nil(t, fork)

		testutil.AssertGolden(t, filepath.Join("testdata", "golden", testName), update(testName), fork)
	})
}

func newV3TestClient(t testing.TB, name string) (*V3Client, func()) {
	t.Helper()

	cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	doer, err := cf.Doer()
	if err != nil {
		t.Fatal(err)
	}

	return NewV3Client(logtest.Scoped(t), "Test", uri, vcrToken, doer), save
}

func newV3TestEnterpriseClient(t testing.TB, name string) (*V3Client, func()) {
	t.Helper()

	cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
	uri, err := url.Parse("https://ghe.sgdev.org/api/v3")
	if err != nil {
		t.Fatal(err)
	}

	doer, err := cf.Doer()
	if err != nil {
		t.Fatal(err)
	}

	return NewV3Client(logtest.Scoped(t), "Test", uri, gheToken, doer), save
}

func strPtr(s string) *string { return &s }

func TestClient_ListRepositoriesForSearch(t *testing.T) {
	cli, save := newV3TestClient(t, "ListRepositoriesForSearch")
	defer save()

	rcache.SetupForTest(t)
	reposPage, err := cli.ListRepositoriesForSearch(context.Background(), "org:sourcegraph-vcr-repos", 1)
	if err != nil {
		t.Fatal(err)
	}

	if reposPage.Repos == nil {
		t.Fatal("expected repos but got nil")
	}

	testutil.AssertGolden(t,
		"testdata/golden/ListRepositoriesForSearch",
		update("ListRepositoriesForSearch"),
		reposPage.Repos,
	)
}

func TestClient_ListRepositoriesForSearch_incomplete(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
  "total_count": 2,
  "incomplete_results": true,
  "items": [
    {
      "node_id": "i",
      "full_name": "o/r",
      "description": "d",
      "html_url": "https://github.example.com/o/r",
      "fork": true
    },
    {
      "node_id": "j",
      "full_name": "a/b",
      "description": "c",
      "html_url": "https://github.example.com/a/b",
      "fork": false
    }
  ]
}
`,
	}
	c := newTestClient(t, &mock)

	// If we have incomplete results we want to fail. Our syncer requires all
	// repositories to be returned, otherwise it will delete the missing
	// repositories.
	_, err := c.ListRepositoriesForSearch(context.Background(), "org:sourcegraph", 1)

	if have, want := err, ErrIncompleteResults; want != have {
		t.Errorf("\nhave: %s\nwant: %s", have, want)
	}
}

type testCase struct {
	repoName    string
	expectedUrl string
}

var testCases = map[string]testCase{
	"github.com": {
		repoName:    "github.com/sd9/sourcegraph",
		expectedUrl: "https://api.github.com/repos/sd9/sourcegraph/hooks",
	},
	"enterprise": {
		repoName:    "ghe.sgdev.org/milton/test",
		expectedUrl: "https://ghe.sgdev.org/api/v3/repos/milton/test/hooks",
	},
}

func TestSyncWebhook_CreateListFindDelete(t *testing.T) {
	ctx := context.Background()

	client, save := newV3TestClient(t, "CreateListFindDeleteWebhooks")
	defer save()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			token := os.Getenv(fmt.Sprintf("%s_ACCESS_TOKEN", name))
			client = client.WithAuthenticator(&auth.OAuthBearerToken{Token: token})

			id, err := client.CreateSyncWebhook(ctx, tc.repoName, "https://target-url.com", "secret")
			if err != nil {
				t.Fatal(err)
			}

			if _, err := client.FindSyncWebhook(ctx, tc.repoName); err != nil {
				t.Error(`Could not find webhook with "/github-webhooks" endpoint`)
			}

			deleted, err := client.DeleteSyncWebhook(ctx, tc.repoName, id)
			if err != nil {
				t.Error(err)
			}

			if !deleted {
				t.Fatal("Could not delete created repo")
			}
		})
	}
}

func TestSyncWebhook_webhookURLBuilderPlain(t *testing.T) {
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			want := tc.expectedUrl
			have, err := webhookURLBuilder(tc.repoName)
			if err != nil {
				t.Fatal(err)
			}
			if have != want {
				t.Fatalf("expected: %s, got: %s", want, have)
			}
		})
	}
}

func TestSyncWebhook_webhookURLBuilderWithID(t *testing.T) {
	type testCaseWithID struct {
		repoName    string
		id          int
		expectedUrl string
	}

	testCases := map[string]testCaseWithID{
		"github.com": {
			repoName:    "github.com/sd9/sourcegraph",
			id:          42,
			expectedUrl: "https://api.github.com/repos/sd9/sourcegraph/hooks/42",
		},
		"enterprise": {
			repoName:    "ghe.sgdev.org/milton/test",
			id:          69,
			expectedUrl: "https://ghe.sgdev.org/api/v3/repos/milton/test/hooks/69",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			want := tc.expectedUrl
			have, err := webhookURLBuilderWithID(tc.repoName, tc.id)
			if err != nil {
				t.Fatal(err)
			}
			if have != want {
				t.Fatalf("expected: %s, got: %s", want, have)
			}
		})
	}
}

func TestResponseHasNextPage(t *testing.T) {
	t.Run("has next page", func(t *testing.T) {
		headers := http.Header{}
		headers.Add("Link", `<https://api.github.com/sourcegraph-vcr/private-repo-1/collaborators?page=2&per_page=100&affiliation=direct>; rel="next", <https://api.github.com/sourcegraph-vcr/private-repo-1/collaborators?page=8&per_page=100&affiliation=direct>; rel="last"`)
		responseState := &httpResponseState{
			statusCode: 200,
			headers:    headers,
		}

		if responseState.hasNextPage() != true {
			t.Fatal("expected true, got false")
		}
	})

	t.Run("does not have next page", func(t *testing.T) {
		headers := http.Header{}
		headers.Add("Link", `<https://api.github.com/sourcegraph-vcr/private-repo-1/collaborators?page=2&per_page=100&affiliation=direct>; rel="prev", <https://api.github.com/sourcegraph-vcr/private-repo-1/collaborators?page=1&per_page=100&affiliation=direct>; rel="first"`)
		responseState := &httpResponseState{
			statusCode: 200,
			headers:    headers,
		}

		if responseState.hasNextPage() != false {
			t.Fatal("expected false, got true")
		}
	})

	t.Run("no header returns false", func(t *testing.T) {
		headers := http.Header{}
		responseState := &httpResponseState{
			statusCode: 200,
			headers:    headers,
		}

		if responseState.hasNextPage() != false {
			t.Fatal("expected false, got true")
		}
	})
}

func TestListPublicRepositories(t *testing.T) {
	t.Run("should skip null REST repositories", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(`[{"node_id": "1"}, null, {}, {"node_id": "2"}]`))
			if err != nil {
				t.Fatalf("failed to write response: %v", err)
			}
		}))

		uri, _ := url.Parse(testServer.URL)
		testCli := NewV3Client(logtest.Scoped(t), "Test", uri, gheToken, testServer.Client())

		repositories, hasNextPage, err := testCli.ListPublicRepositories(context.Background(), 0)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, repositories, 2)
		assert.False(t, hasNextPage)
		assert.Equal(t, "1", repositories[0].ID)
		assert.Equal(t, "2", repositories[1].ID)
	})
}
