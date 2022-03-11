package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	authzGitHub "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/github"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	extsvcGitHub "github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

// NOTE: To update VCR for these tests, please use the token of "sourcegraph-vcr"
// for GITHUB_TOKEN, which can be found in 1Password.
//
// We also recommend setting up a new token for "sourcegraph-vcr" using the auth scope
// guidelines https://docs.sourcegraph.com/admin/external_service/github#github-api-token-and-access
// to ensure everything works, in case of new scopes being required.
func TestIntegration_GitHubPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	token := os.Getenv("GITHUB_TOKEN")

	spec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   "https://github.com/",
		AccountID:   "66464926",
	}
	svc := types.ExternalService{
		Kind:      extsvc.KindGitHub,
		CreatedAt: timeutil.Now(),
		Config:    `{"url": "https://github.com", "authorization": {}}`,
	}
	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	// This integration tests performs a repository-centric permissions syncing against
	// https://github.com, then check if permissions are correctly granted for the test
	// user "sourcegraph-vcr-bob", who is a outside collaborator of the repository
	// "sourcegraph-vcr-repos/private-org-repo-1".
	t.Run("repo-centric", func(t *testing.T) {
		newUser := database.NewUser{
			Email:           "sourcegraph-vcr-bob@sourcegraph.com",
			Username:        "sourcegraph-vcr-bob",
			EmailIsVerified: true,
		}
		t.Run("no-groups", func(t *testing.T) {
			name := t.Name()
			cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
			defer save()

			doer, err := cf.Doer()
			if err != nil {
				t.Fatal(err)
			}
			cli := extsvcGitHub.NewV3Client(svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := dbtest.NewDB(t)
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(testDB, sql.TxOptions{})

			err = reposStore.ExternalServiceStore.Upsert(ctx, &svc)
			if err != nil {
				t.Fatal(err)
			}

			provider := authzGitHub.NewProvider(svc.URN(), authzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BaseToken:      token,
				GroupsCacheTTL: -1,
			})

			authz.SetProviders(false, []authz.Provider{provider})
			defer authz.SetProviders(true, nil)

			repo := types.Repo{
				Name:    "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				Private: true,
				URI:     "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "MDEwOlJlcG9zaXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: map[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore.Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			userID, err := database.ExternalAccounts(testDB).CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{})
			if err != nil {
				t.Fatal(err)
			}

			db := database.NewDB(testDB)
			permsStore := edb.Perms(testDB, timeutil.Now)
			syncer := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

			err = syncer.syncRepoPerms(ctx, repo.ID, false, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}

			p := &authz.UserPermissions{
				UserID: userID,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			err = permsStore.LoadUserPermissions(ctx, p)
			if err != nil {
				t.Fatal(err)
			}

			wantIDs := []int32{1}
			if diff := cmp.Diff(wantIDs, p.GenerateSortedIDsSlice()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run("groups-enabled", func(t *testing.T) {
			name := t.Name()
			cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
			defer save()

			doer, err := cf.Doer()
			if err != nil {
				t.Fatal(err)
			}
			cli := extsvcGitHub.NewV3Client(svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := dbtest.NewDB(t)
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(testDB, sql.TxOptions{})

			err = reposStore.ExternalServiceStore.Upsert(ctx, &svc)
			if err != nil {
				t.Fatal(err)
			}

			provider := authzGitHub.NewProvider(svc.URN(), authzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BaseToken:      token,
				GroupsCacheTTL: 72,
			})

			authz.SetProviders(false, []authz.Provider{provider})
			defer authz.SetProviders(true, nil)

			repo := types.Repo{
				Name:    "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				Private: true,
				URI:     "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "MDEwOlJlcG9zaXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: map[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore.Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			userID, err := database.ExternalAccounts(testDB).CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{})
			if err != nil {
				t.Fatal(err)
			}

			db := database.NewDB(testDB)
			permsStore := edb.Perms(testDB, timeutil.Now)
			syncer := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

			err = syncer.syncRepoPerms(ctx, repo.ID, false, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}

			p := &authz.UserPermissions{
				UserID: userID,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			err = permsStore.LoadUserPermissions(ctx, p)
			if err != nil {
				t.Fatal(err)
			}

			wantIDs := []int32{1}
			if diff := cmp.Diff(wantIDs, p.GenerateSortedIDsSlice()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}

			// sync again and check
			err = syncer.syncRepoPerms(ctx, repo.ID, false, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}
			err = permsStore.LoadUserPermissions(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(wantIDs, p.GenerateSortedIDsSlice()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}
		})
	})

	// This integration tests performs a repository-centric permissions syncing against
	// https://github.com, then check if permissions are correctly granted for the test
	// user "sourcegraph-vcr", who is a collaborator of "sourcegraph-vcr-repos/private-org-repo-1".
	t.Run("user-centric", func(t *testing.T) {
		newUser := database.NewUser{
			Email:           "sourcegraph-vcr@sourcegraph.com",
			Username:        "sourcegraph-vcr",
			EmailIsVerified: true,
		}
		t.Run("no-groups", func(t *testing.T) {
			name := t.Name()

			cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
			defer save()
			doer, err := cf.Doer()
			if err != nil {
				t.Fatal(err)
			}
			cli := extsvcGitHub.NewV3Client(svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := dbtest.NewDB(t)
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(testDB, sql.TxOptions{})

			err = reposStore.ExternalServiceStore.Upsert(ctx, &svc)
			if err != nil {
				t.Fatal(err)
			}

			provider := authzGitHub.NewProvider(svc.URN(), authzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BaseToken:      token,
				GroupsCacheTTL: -1,
			})

			authz.SetProviders(false, []authz.Provider{provider})
			defer authz.SetProviders(true, nil)

			repo := types.Repo{
				Name:    "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				Private: true,
				URI:     "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "MDEwOlJlcG9zaXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: map[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore.Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			authData := json.RawMessage(fmt.Sprintf(`{"access_token": "%s"}`, token))
			userID, err := database.ExternalAccounts(testDB).CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{
				AuthData: &authData,
			})
			if err != nil {
				t.Fatal(err)
			}

			db := database.NewDB(testDB)
			permsStore := edb.Perms(testDB, timeutil.Now)
			syncer := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

			err = syncer.syncUserPerms(ctx, userID, false, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}

			p := &authz.UserPermissions{
				UserID: userID,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			err = permsStore.LoadUserPermissions(ctx, p)
			if err != nil {
				t.Fatal(err)
			}

			wantIDs := []int32{1}
			if diff := cmp.Diff(wantIDs, p.GenerateSortedIDsSlice()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run("groups-enabled", func(t *testing.T) {
			name := t.Name()

			cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
			defer save()
			doer, err := cf.Doer()
			if err != nil {
				t.Fatal(err)
			}
			cli := extsvcGitHub.NewV3Client(svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := dbtest.NewDB(t)
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(testDB, sql.TxOptions{})

			err = reposStore.ExternalServiceStore.Upsert(ctx, &svc)
			if err != nil {
				t.Fatal(err)
			}

			provider := authzGitHub.NewProvider(svc.URN(), authzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BaseToken:      token,
				GroupsCacheTTL: 72,
			})

			authz.SetProviders(false, []authz.Provider{provider})
			defer authz.SetProviders(true, nil)

			repo := types.Repo{
				Name:    "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				Private: true,
				URI:     "github.com/sourcegraph-vcr-repos/private-org-repo-1",
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "MDEwOlJlcG9zaXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: map[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore.Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			authData := json.RawMessage(fmt.Sprintf(`{"access_token": "%s"}`, token))
			userID, err := database.ExternalAccounts(testDB).CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{
				AuthData: &authData,
			})
			if err != nil {
				t.Fatal(err)
			}

			db := database.NewDB(testDB)
			permsStore := edb.Perms(testDB, timeutil.Now)
			syncer := NewPermsSyncer(db, reposStore, permsStore, timeutil.Now, nil)

			err = syncer.syncUserPerms(ctx, userID, false, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}

			p := &authz.UserPermissions{
				UserID: userID,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			err = permsStore.LoadUserPermissions(ctx, p)
			if err != nil {
				t.Fatal(err)
			}

			wantIDs := []int32{1}
			if diff := cmp.Diff(wantIDs, p.GenerateSortedIDsSlice()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}

			// sync again and check
			err = syncer.syncUserPerms(ctx, userID, false, authz.FetchPermsOptions{})
			if err != nil {
				t.Fatal(err)
			}
			err = permsStore.LoadUserPermissions(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(wantIDs, p.GenerateSortedIDsSlice()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}
		})
	})
}

func TestIntegration_GitHubEnterprisePermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Token with org:write is required, so we can use a token from milton's account in
	// ghe.sgdev.org.
	token := os.Getenv("GHE_TOKEN")

	spec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   "https://ghe.sgdev.org/api/v3/",
		AccountID:   "2383", // TODO: For now indradhanush.
	}

	svc := types.ExternalService{
		Kind:      extsvc.KindGitHub,
		CreatedAt: timeutil.Now(),
		Config:    `{"url": "https://ghe.sgdev.org/api/v3", "authorization": {}}`,
	}

	url, err := url.Parse("https://ghe.sgdev.org/api/v3")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("repo-centric", func(t *testing.T) {
		rcache.SetupForTest(t)

		// The user on ghe.sgdev.org with username "integration-test" does not belong to any
		// orgs. Yet, it should be able to access internal repos belonging to any org in that GHE
		// instance. This test case uses the following resources:
		//
		// 1. GHE Instance: https://ghe.sgdev.org
		// 2. GHE Org: https://ghe.sgdev.org/sgtest
		// 3. GHE Internal repo: https://ghe.sgdev.org/sgtest/internal
		// 4. GHE user: https://ghe.sgdev.org/integration-test
		//
		// And this test case verifies that the user is able to access the internal repo on
		// Sourcegraph even though they are not a part of the GitHub organization on the GHE
		// instance.

		newUser := database.NewUser{
			Email:           "integration-test@sourcegraph.com",
			Username:        "integration-test",
			EmailIsVerified: true,
		}

		// Test case when groups cache is enabled.
		t.Run("groups-enabled", func(t *testing.T) {
			// Internal repo support is guarded behind a feature flag and will be removed once it is
			// declared stable.
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						EnableGithubInternalRepoVisibility: true,
					},
				},
			})

			// Setup test stubs and prerequisites.
			name := t.Name()
			cf, save := httptestutil.NewGitHubRecorderFactory(t, update(name), name)
			defer save()

			doer, err := cf.Doer()
			if err != nil {
				t.Fatal(err)
			}
			cli := extsvcGitHub.NewV3Client(url, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := dbtest.NewDB(t)
			ctx := actor.WithInternalActor(context.Background())

			repoStore := repos.NewStore(testDB, sql.TxOptions{})
			if err := repoStore.ExternalServiceStore.Upsert(ctx, &svc); err != nil {
				t.Fatal(err)
			}

			// Provider with groupsCacheTTL enabled and configured.
			provider := authzGitHub.NewProvider(svc.URN(), authzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      url,
				BaseToken:      token,
				GroupsCacheTTL: 72,
			})
			provider.EnableGithubInternalRepoVisibility()

			authz.SetProviders(false, []authz.Provider{provider})

			repo := types.Repo{
				Name:    "ghe.sgdev.org/sgtest/internal",
				Private: true,
				URI:     "ghe.sgdev.org/sgtest/internal",
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "MDEwOlJlcG9zaXRvcnk0NDIyODI=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://ghe.sgdev.org/api/v3/",
				},
				Sources: map[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}

			if err := repoStore.RepoStore.Create(ctx, &repo); err != nil {
				t.Fatal(err)
			}

			userID, err := database.ExternalAccounts(testDB).CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{})
			if err != nil {
				t.Fatal(err)
			}

			db := database.NewDB(testDB)
			permsStore := edb.Perms(testDB, timeutil.Now)
			syncer := NewPermsSyncer(db, repoStore, permsStore, timeutil.Now, nil)

			if err := syncer.syncRepoPerms(ctx, repo.ID, false, authz.FetchPermsOptions{}); err != nil {
				t.Fatal(err)
			}

			p := &authz.UserPermissions{
				UserID: userID,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}

			if err := permsStore.LoadUserPermissions(ctx, p); err != nil {
				t.Fatal(err)
			}

			wantIDs := []uint32{1}
			if diff := cmp.Diff(wantIDs, p.IDs.ToArray()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}

			// sync again and check
			if err := syncer.syncRepoPerms(ctx, repo.ID, false, authz.FetchPermsOptions{}); err != nil {
				t.Fatal(err)
			}

			if err := permsStore.LoadUserPermissions(ctx, p); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantIDs, p.IDs.ToArray()); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}
		})
	})
}
