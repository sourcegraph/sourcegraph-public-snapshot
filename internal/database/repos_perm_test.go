package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type fakeProvider struct {
	codeHost *extsvc.CodeHost
	extAcct  *extsvc.Account
}

func (p *fakeProvider) FetchAccount(context.Context, *types.User, []*extsvc.Account, []string) (mine *extsvc.Account, err error) {
	return p.extAcct, nil
}

func (p *fakeProvider) ServiceType() string { return p.codeHost.ServiceType }
func (p *fakeProvider) ServiceID() string   { return p.codeHost.ServiceID }
func (p *fakeProvider) URN() string         { return extsvc.URN(p.codeHost.ServiceType, 0) }

func (p *fakeProvider) ValidateConnection(context.Context) (problems []string) { return nil }

func (p *fakeProvider) FetchUserPerms(context.Context, *extsvc.Account, authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, nil
}

func (p *fakeProvider) FetchUserPermsByToken(context.Context, string, authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, nil
}

func (p *fakeProvider) FetchRepoPerms(context.Context, *extsvc.Repository, authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, nil
}

// 🚨 SECURITY: Tests are necessary to ensure security.
func TestAuthzQueryConds(t *testing.T) {
	cmpOpts := cmp.AllowUnexported(sqlf.Query{})
	db := NewDB(dbtest.NewDB(t))

	t.Run("Conflict with permissions user mapping", func(t *testing.T) {
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
		defer authz.SetProviders(true, nil)

		_, err := AuthzQueryConds(context.Background(), db)
		gotErr := fmt.Sprintf("%v", err)
		if diff := cmp.Diff(errPermissionsUserMappingConflict.Error(), gotErr); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("When permissions user mapping is enabled", func(t *testing.T) {
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		got, err := AuthzQueryConds(context.Background(), db)
		if err != nil {
			t.Fatal(err)
		}
		want := authzQuery(false, true, int32(0), authz.Read)
		if diff := cmp.Diff(want, got, cmpOpts); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	tests := []struct {
		name                string
		setup               func(t *testing.T) (context.Context, DB)
		authzAllowByDefault bool
		wantQuery           *sqlf.Query
	}{
		{
			name: "internal actor bypass checks",
			setup: func(t *testing.T) (context.Context, DB) {
				return actor.WithInternalActor(context.Background()), db
			},
			wantQuery: authzQuery(true, false, int32(0), authz.Read),
		},
		{
			name: "no authz provider and not allow by default",
			setup: func(t *testing.T) (context.Context, DB) {
				return context.Background(), db
			},
			wantQuery: authzQuery(false, false, int32(0), authz.Read),
		},
		{
			name: "no authz provider but allow by default",
			setup: func(t *testing.T) (context.Context, DB) {
				return context.Background(), db
			},
			authzAllowByDefault: true,
			wantQuery:           authzQuery(true, false, int32(0), authz.Read),
		},
		{
			name: "authenticated user is a site admin",
			setup: func(_ *testing.T) (context.Context, DB) {
				users := NewMockUserStoreFrom(db.Users())
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
				mockDB := NewMockDBFrom(db)
				mockDB.UsersFunc.SetDefaultReturn(users)
				return actor.WithActor(context.Background(), &actor.Actor{UID: 1}), mockDB
			},
			wantQuery: authzQuery(true, false, int32(1), authz.Read),
		},
		{
			name: "authenticated user is a site admin and AuthzEnforceForSiteAdmins is set",
			setup: func(t *testing.T) (context.Context, DB) {
				users := NewMockUserStoreFrom(db.Users())
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
				mockDB := NewMockDBFrom(db)
				mockDB.UsersFunc.SetDefaultReturn(users)
				conf.Get().AuthzEnforceForSiteAdmins = true
				t.Cleanup(func() {
					conf.Get().AuthzEnforceForSiteAdmins = false
				})
				return actor.WithActor(context.Background(), &actor.Actor{UID: 1}), mockDB
			},
			wantQuery: authzQuery(false, false, int32(1), authz.Read),
		},
		{
			name: "authenticated user is not a site admin",
			setup: func(_ *testing.T) (context.Context, DB) {
				users := NewMockUserStoreFrom(db.Users())
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
				mockDB := NewMockDBFrom(db)
				mockDB.UsersFunc.SetDefaultReturn(users)
				return actor.WithActor(context.Background(), &actor.Actor{UID: 1}), mockDB
			},
			wantQuery: authzQuery(false, false, int32(1), authz.Read),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authz.SetProviders(test.authzAllowByDefault, nil)
			defer authz.SetProviders(true, nil)

			ctx, mockDB := test.setup(t)
			q, err := AuthzQueryConds(ctx, mockDB)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantQuery, q, cmpOpts); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRepoStore_nonSiteAdminCanViewOwnPrivateCode(t *testing.T) {
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Add a single user who is NOT a site admin
	alice, err := db.Users().Create(ctx, NewUser{
		Email:                 "alice@example.com",
		Username:              "alice",
		Password:              "alice",
		EmailVerificationCode: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Users().SetIsSiteAdmin(ctx, alice.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	// Add both a public and private repo
	internalCtx := actor.WithInternalActor(ctx)
	alicePublicRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name: "alice_public_repo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "alice_public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	alicePrivateRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "alice_private_repo",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "alice_private_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	aliceExternalService := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
	}
	err = db.ExternalServices().Create(ctx, confGet, aliceExternalService)
	if err != nil {
		t.Fatal(err)
	}

	// Set it up so that Alice added the repo via an external service
	q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, user_id, clone_url)
VALUES (%s, %s, NULLIF(%s, 0), '')
`, aliceExternalService.ID, alicePrivateRepo.ID, aliceExternalService.NamespaceUserID)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	q = sqlf.Sprintf(`
INSERT INTO user_permissions (user_id, permission, object_type, object_ids_ints, updated_at)
VALUES
	(%s, 'read', 'repos', %s, NOW())
`,
		alice.ID, pq.Array([]int32{int32(alicePrivateRepo.ID)}),
	)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
	defer authz.SetProviders(true, nil)

	// Alice should be able to see both her public and private repos
	aliceCtx := actor.WithActor(ctx, &actor.Actor{UID: alice.ID})
	repos, err := db.Repos().List(aliceCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos := []*types.Repo{alicePublicRepo, alicePrivateRepo}
	if diff := cmp.Diff(wantRepos, repos, cmpopts.IgnoreFields(types.Repo{}, "Sources")); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestRepoStore_userCanSeeUnrestricedRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Add a single user who is NOT a site admin
	alice, err := db.Users().Create(ctx,
		NewUser{
			Email:                 "alice@example.com",
			Username:              "alice",
			Password:              "alice",
			EmailVerificationCode: "alice",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Users().SetIsSiteAdmin(ctx, alice.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	// Set up a private repo that the user does not have access to
	internalCtx := actor.WithInternalActor(ctx)
	privateRepo1 := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "private_repo_1",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "private_repo_1",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	extsvc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
	}
	err = db.ExternalServices().Create(ctx, confGet, extsvc)
	if err != nil {
		t.Fatal(err)
	}

	q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
VALUES (%s, %s, '')
`, extsvc.ID, privateRepo1.ID)

	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	q = sqlf.Sprintf(`
INSERT INTO repo_permissions (repo_id, permission, updated_at)
VALUES (%s, 'read', %s)
`, privateRepo1.ID, time.Now())

	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
	defer authz.SetProviders(true, nil)

	// Alice should NOT be able to see the private repo
	aliceCtx := actor.WithActor(ctx, &actor.Actor{UID: alice.ID})
	repos, err := db.Repos().List(aliceCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var wantRepos []*types.Repo
	if diff := cmp.Diff(wantRepos, repos, cmpopts.IgnoreFields(types.Repo{}, "Sources")); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// Mark is unrestricted
	q = sqlf.Sprintf(`
UPDATE repo_permissions SET unrestricted = true
`)

	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	// Alice should now be able to see the repo
	aliceCtx = actor.WithActor(ctx, &actor.Actor{UID: alice.ID})
	repos, err = db.Repos().List(aliceCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = []*types.Repo{privateRepo1}
	if diff := cmp.Diff(wantRepos, repos, cmpopts.IgnoreFields(types.Repo{}, "Sources")); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestRepoStore_nonSiteAdminCanViewOrgPrivateCode(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Add a single user who is NOT a site admin
	alice, err := db.Users().Create(ctx,
		NewUser{
			Email:                 "alice@example.com",
			Username:              "alice",
			Password:              "alice",
			EmailVerificationCode: "alice",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Users().SetIsSiteAdmin(ctx, alice.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	// Set up two private repositories the user has access to both on the code host:
	//  1. One is not added to the organization code host connection
	//  2. One is added to the organization code host connection
	internalCtx := actor.WithInternalActor(ctx)
	privateRepo1 := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "private_repo_1",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "private_repo_1",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	privateRepo2 := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "private_repo_2",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "private_repo_2",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]

	// Create an organization and add alice as a member
	org, err := db.Orgs().Create(ctx, "org", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.OrgMembers().Create(ctx, org.ID, alice.ID)
	if err != nil {
		t.Fatal(err)
	}

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	extsvc := &types.ExternalService{
		Kind:           extsvc.KindGitHub,
		DisplayName:    "GITHUB #1",
		Config:         `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
		NamespaceOrgID: org.ID,
	}
	err = db.ExternalServices().Create(ctx, confGet, extsvc)
	if err != nil {
		t.Fatal(err)
	}

	q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, org_id, clone_url)
VALUES (%s, %s, NULLIF(%s, 0), '')
`, extsvc.ID, privateRepo2.ID, extsvc.NamespaceOrgID)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	q = sqlf.Sprintf(`
INSERT INTO user_permissions (user_id, permission, object_type, object_ids_ints, updated_at)
VALUES
	(%s, 'read', 'repos', %s, NOW())
`,
		alice.ID, pq.Array([]int32{int32(privateRepo1.ID), int32(privateRepo2.ID)}),
	)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
	defer authz.SetProviders(true, nil)

	// Alice should be able to see both her public and private repos
	aliceCtx := actor.WithActor(ctx, &actor.Actor{UID: alice.ID})
	repos, err := db.Repos().List(aliceCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos := []*types.Repo{privateRepo2}
	if diff := cmp.Diff(wantRepos, repos, cmpopts.IgnoreFields(types.Repo{}, "Sources")); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func createGitHubExternalService(t *testing.T, db DB, userID int32) *types.ExternalService {
	now := time.Now()
	svc := &types.ExternalService{
		Kind:            extsvc.KindGitHub,
		DisplayName:     "Github - Test",
		Config:          `{"url": "https://github.com", "authorization": {}}`,
		NamespaceUserID: userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := db.ExternalServices().Upsert(context.Background(), svc); err != nil {
		t.Fatal(err)
	}

	return svc
}

// 🚨 SECURITY: Tests are necessary to ensure security.
func TestRepoStore_List_checkPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Set up three users: alice, bob and admin
	alice, err := db.Users().Create(ctx, NewUser{
		Email:                 "alice@example.com",
		Username:              "alice",
		Password:              "alice",
		EmailVerificationCode: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	bob, err := db.Users().Create(ctx, NewUser{
		Email:                 "bob@example.com",
		Username:              "bob",
		Password:              "bob",
		EmailVerificationCode: "bob",
	})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := db.Users().Create(ctx, NewUser{
		Email:                 "admin@example.com",
		Username:              "admin",
		Password:              "admin",
		EmailVerificationCode: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Ensure only "admin" is the site admin, alice was prompted as site admin
	// because it was the first user.
	err = db.Users().SetIsSiteAdmin(ctx, admin.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Users().SetIsSiteAdmin(ctx, alice.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	siteLevelGitHubService := createGitHubExternalService(t, db, 0)

	// Set up some repositories: public and private for both alice and bob
	internalCtx := actor.WithInternalActor(ctx)
	alicePublicRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name: "alice_public_repo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "alice_public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	alicePublicRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	alicePrivateRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "alice_private_repo",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "alice_private_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	alicePrivateRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	bobPublicRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name: "bob_public_repo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "bob_public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	bobPublicRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	bobPrivateRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "bob_private_repo",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "bob_private_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	bobPrivateRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	// Make sure that alicePublicRepo, alicePrivateRepo, bobPublicRepo and bobPrivateRepo have an
	// entry in external_service_repos table.
	repoIDs := []api.RepoID{
		alicePublicRepo.ID,
		alicePrivateRepo.ID,
		bobPublicRepo.ID,
		bobPrivateRepo.ID,
	}

	for _, id := range repoIDs {
		q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
VALUES (%s, %s, '')
`, siteLevelGitHubService.ID, id)
		_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Set up another unrestricted private repo from cindy
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	cindyExternalService := &types.ExternalService{
		Kind:         extsvc.KindGitHub,
		DisplayName:  "GITHUB #1",
		Config:       `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
		Unrestricted: true,
	}
	err = db.ExternalServices().Create(ctx, confGet, cindyExternalService)
	if err != nil {
		t.Fatal(err)
	}

	cindyPrivateRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "cindy_private_repo",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "cindy_private_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	cindyPrivateRepo.Sources = map[string]*types.SourceInfo{
		cindyExternalService.URN(): {ID: cindyExternalService.URN()},
	}

	q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, user_id, clone_url)
VALUES (%s, %s, NULLIF(%s, 0), '')
`, cindyExternalService.ID, cindyPrivateRepo.ID, cindyExternalService.NamespaceUserID)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	// Set up permissions: alice and bob have access to their own private repositories
	q = sqlf.Sprintf(`
INSERT INTO user_permissions (user_id, permission, object_type, object_ids_ints, updated_at)
VALUES
	(%s, 'read', 'repos', %s, NOW()),
	(%s, 'read', 'repos', %s, NOW())
`,
		alice.ID, pq.Array([]int32{int32(alicePrivateRepo.ID)}),
		bob.ID, pq.Array([]int32{int32(bobPrivateRepo.ID)}),
	)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
	defer authz.SetProviders(true, nil)

	// Alice should see "alice_public_repo", "alice_private_repo",
	// "bob_public_repo", "cindy_private_repo". The "cindy_private_repo" comes from
	// an unrestricted external service
	aliceCtx := actor.WithActor(ctx, &actor.Actor{UID: alice.ID})
	repos, err := db.Repos().List(aliceCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos := []*types.Repo{alicePublicRepo, alicePrivateRepo, bobPublicRepo, cindyPrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// Bob should see "alice_public_repo", "bob_private_repo", "bob_public_repo",
	// "cindy_private_repo". The "cindy_private_repo" comes from an unrestricted
	// external service
	bobCtx := actor.WithActor(ctx, &actor.Actor{UID: bob.ID})
	repos, err = db.Repos().List(bobCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = []*types.Repo{alicePublicRepo, bobPublicRepo, bobPrivateRepo, cindyPrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// By default, site admins can see all repos
	adminCtx := actor.WithActor(ctx, &actor.Actor{UID: admin.ID})
	repos, err = db.Repos().List(adminCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = []*types.Repo{alicePublicRepo, alicePrivateRepo, bobPublicRepo, bobPrivateRepo, cindyPrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// When AuthzEnforceForSiteAdmins is set, site admins can only see repos they have access
	// to based on our authz model
	conf.Get().AuthzEnforceForSiteAdmins = true
	t.Cleanup(func() {
		conf.Get().AuthzEnforceForSiteAdmins = false
	})
	adminCtx = actor.WithActor(ctx, &actor.Actor{UID: admin.ID})
	repos, err = db.Repos().List(adminCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = []*types.Repo{alicePublicRepo, bobPublicRepo, cindyPrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// A random user should only see "alice_public_repo", "bob_public_repo", "cindy_private_repo"
	// "cindy_private_repos" comes from an unrestricted external service
	repos, err = db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = []*types.Repo{alicePublicRepo, bobPublicRepo, cindyPrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

// 🚨 SECURITY: Tests are necessary to ensure security.
func TestRepoStore_List_permissionsUserMapping(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Set up three users: alice, bob and admin
	alice, err := db.Users().Create(ctx, NewUser{
		Email:                 "alice@example.com",
		Username:              "alice",
		Password:              "alice",
		EmailVerificationCode: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	bob, err := db.Users().Create(ctx, NewUser{
		Email:                 "bob@example.com",
		Username:              "bob",
		Password:              "bob",
		EmailVerificationCode: "bob",
	})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := db.Users().Create(ctx, NewUser{
		Email:                 "admin@example.com",
		Username:              "admin",
		Password:              "admin",
		EmailVerificationCode: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Ensure only "admin" is the site admin, alice was prompted as site admin
	// because it was the first user.
	err = db.Users().SetIsSiteAdmin(ctx, admin.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Users().SetIsSiteAdmin(ctx, alice.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	siteLevelGitHubService := createGitHubExternalService(t, db, 0)

	// Set up some repositories: public and private for both alice and bob
	internalCtx := actor.WithInternalActor(ctx)
	alicePublicRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name: "alice_public_repo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "alice_public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	alicePublicRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	alicePrivateRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "alice_private_repo",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "alice_private_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	alicePrivateRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	bobPublicRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name: "bob_public_repo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "bob_public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	bobPublicRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	bobPrivateRepo := mustCreate(internalCtx, t, db,
		&types.Repo{
			Name:    "bob_private_repo",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "bob_private_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)[0]
	bobPrivateRepo.Sources = map[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	// Make sure that alicePublicRepo, alicePrivateRepo, bobPublicRepo and bobPrivateRepo have an
	// entry in external_service_repos table.
	repoIDs := []api.RepoID{
		alicePublicRepo.ID,
		alicePrivateRepo.ID,
		bobPublicRepo.ID,
		bobPrivateRepo.ID,
	}

	for _, id := range repoIDs {
		q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
VALUES (%s, %s, '')
`, siteLevelGitHubService.ID, id)
		_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Set up permissions: alice and bob have access to their own private repositories
	q := sqlf.Sprintf(`
INSERT INTO user_permissions (user_id, permission, object_type, object_ids_ints, updated_at)
VALUES
	(%s, 'read', 'repos', %s, NOW()),
	(%s, 'read', 'repos', %s, NOW())
`,
		alice.ID, pq.Array([]int32{int32(alicePrivateRepo.ID)}),
		bob.ID, pq.Array([]int32{int32(bobPrivateRepo.ID)}),
	)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	before := globals.PermissionsUserMapping()
	globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
	defer globals.SetPermissionsUserMapping(before)

	// Alice should see "alice_private_repo" but not "alice_public_repo", "bob_public_repo", "bob_private_repo"
	aliceCtx := actor.WithActor(ctx, &actor.Actor{UID: alice.ID})
	repos, err := db.Repos().List(aliceCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos := []*types.Repo{alicePrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// Bob should see "bob_private_repo" but not "alice_public_repo" or "bob_public_repo"
	bobCtx := actor.WithActor(ctx, &actor.Actor{UID: bob.ID})
	repos, err = db.Repos().List(bobCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = []*types.Repo{bobPrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// By default, admins can see all repos
	adminCtx := actor.WithActor(ctx, &actor.Actor{UID: admin.ID})
	repos, err = db.Repos().List(adminCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = []*types.Repo{alicePublicRepo, alicePrivateRepo, bobPublicRepo, bobPrivateRepo}
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// Admin should not see anything as they have not been granted permissions and
	// AuthzEnforceForSiteAdmins is set
	conf.Get().AuthzEnforceForSiteAdmins = true
	t.Cleanup(func() {
		conf.Get().AuthzEnforceForSiteAdmins = false
	})
	adminCtx = actor.WithActor(ctx, &actor.Actor{UID: admin.ID})
	repos, err = db.Repos().List(adminCtx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = ([]*types.Repo)(nil)
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// A random user sees nothing
	repos, err = db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantRepos = ([]*types.Repo)(nil)
	if diff := cmp.Diff(wantRepos, repos); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}
