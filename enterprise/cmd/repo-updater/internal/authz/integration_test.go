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
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	extsvcGitHub "github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
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
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := database.NewDB(dbtest.NewDB(t))
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(logtest.Scoped(t), database.NewDB(testDB), sql.TxOptions{})

			err = reposStore.ExternalServiceStore().Upsert(ctx, &svc)
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
			err = reposStore.RepoStore().Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			userID, err := testDB.UserExternalAccounts().CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{})
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
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := database.NewDB(dbtest.NewDB(t))
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(logtest.Scoped(t), database.NewDB(testDB), sql.TxOptions{})

			err = reposStore.ExternalServiceStore().Upsert(ctx, &svc)
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
			err = reposStore.RepoStore().Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			userID, err := testDB.UserExternalAccounts().CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{})
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
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := database.NewDB(dbtest.NewDB(t))
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(logtest.Scoped(t), database.NewDB(testDB), sql.TxOptions{})

			err = reposStore.ExternalServiceStore().Upsert(ctx, &svc)
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
			err = reposStore.RepoStore().Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			authData := json.RawMessage(fmt.Sprintf(`{"access_token": "%s"}`, token))
			userID, err := testDB.UserExternalAccounts().CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{
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
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &auth.OAuthBearerToken{Token: token}, doer)

			testDB := database.NewDB(dbtest.NewDB(t))
			ctx := actor.WithInternalActor(context.Background())

			reposStore := repos.NewStore(logtest.Scoped(t), database.NewDB(testDB), sql.TxOptions{})

			err = reposStore.ExternalServiceStore().Upsert(ctx, &svc)
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
			err = reposStore.RepoStore().Create(ctx, &repo)
			if err != nil {
				t.Fatal(err)
			}

			authData := json.RawMessage(fmt.Sprintf(`{"access_token": "%s"}`, token))
			userID, err := testDB.UserExternalAccounts().CreateUserAndSave(ctx, newUser, spec, extsvc.AccountData{
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
