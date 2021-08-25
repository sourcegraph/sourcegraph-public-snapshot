package github

import (
	"context"
	"net/url"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestNewRepoCache(t *testing.T) {
	cmpOpts := cmp.AllowUnexported(rcache.Cache{})
	t.Run("GitHub.com", func(t *testing.T) {
		url, _ := url.Parse("https://www.github.com")
		token := &auth.OAuthBearerToken{Token: "asdf"}

		// github.com caches should:
		// (1) use githubProxyURL for the prefix hash rather than the given url
		// (2) have a TTL of 10 minutes
		prefix := "gh_repo:" + token.Hash()
		got := newRepoCache(url, token)
		want := rcache.NewWithTTL(prefix, 600)
		if diff := cmp.Diff(want, got, cmpOpts); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("GitHub Enterprise", func(t *testing.T) {
		url, _ := url.Parse("https://www.sourcegraph.com")
		token := &auth.OAuthBearerToken{Token: "asdf"}

		// GitHub Enterprise caches should:
		// (1) use the given URL for the prefix hash
		// (2) have a TTL of 30 seconds
		prefix := "gh_repo:" + token.Hash()
		got := newRepoCache(url, token)
		want := rcache.NewWithTTL(prefix, 30)
		if diff := cmp.Diff(want, got, cmpOpts); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestListAffiliatedRepositories(t *testing.T) {
	tests := []struct {
		name         string
		visibility   Visibility
		affiliations []Affiliation
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
			affiliations: []Affiliation{AffiliationCollaborator, AffiliationOwner},
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

func Test_GetAuthenticatedUserOAuthScopes(t *testing.T) {
	client, save := newV3TestClient(t, "GetAuthenticatedUserOAuthScopes")
	defer save()

	scopes, err := client.GetAuthenticatedUserOAuthScopes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"admin:enterprise", "admin:gpg_key", "admin:org", "admin:org_hook", "admin:public_key", "admin:repo_hook", "delete:packages", "delete_repo", "gist", "notifications", "repo", "user", "workflow", "write:discussion", "write:packages"}
	sort.Strings(scopes)
	if diff := cmp.Diff(want, scopes); diff != "" {
		t.Fatalf("Scopes mismatch (-want +got):\n%s", diff)
	}
}

func TestGetAuthenticatedUserOrgs(t *testing.T) {
	cli, save := newV3TestClient(t, "GetAuthenticatedUserOrgs")
	defer save()

	ctx := context.Background()
	orgs, err := cli.GetAuthenticatedUserOrgs(ctx)
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

func TestV3Client_WithAuthenticator(t *testing.T) {
	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	old := &V3Client{
		apiURL: uri,
		auth:   &auth.OAuthBearerToken{Token: "old_token"},
	}

	newToken := &auth.OAuthBearerToken{Token: "new_token"}
	new := old.WithAuthenticator(newToken)
	if old == new {
		t.Fatal("both clients have the same address")
	}

	if new.auth != newToken {
		t.Fatalf("token: want %q but got %q", newToken, new.auth)
	}
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

	return NewV3Client(uri, vcrToken, doer), save
}
