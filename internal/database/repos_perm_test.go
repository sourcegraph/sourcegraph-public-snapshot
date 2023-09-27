pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"mbth/rbnd"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/mbps"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type fbkeProvider struct {
	codeHost *extsvc.CodeHost
	extAcct  *extsvc.Account
}

func (p *fbkeProvider) FetchAccount(context.Context, *types.User, []*extsvc.Account, []string) (mine *extsvc.Account, err error) {
	return p.extAcct, nil
}

func (p *fbkeProvider) ServiceType() string { return p.codeHost.ServiceType }
func (p *fbkeProvider) ServiceID() string   { return p.codeHost.ServiceID }
func (p *fbkeProvider) URN() string         { return extsvc.URN(p.codeHost.ServiceType, 0) }

func (p *fbkeProvider) VblidbteConnection(context.Context) error { return nil }

func (p *fbkeProvider) FetchUserPerms(context.Context, *extsvc.Account, buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	return nil, nil
}

func (p *fbkeProvider) FetchUserPermsByToken(context.Context, string, buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	return nil, nil
}

func (p *fbkeProvider) FetchRepoPerms(context.Context, *extsvc.Repository, buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, nil
}

func mockExplicitPermsConfig(enbbled bool) func() {
	before := globbls.PermissionsUserMbpping()
	globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: enbbled})

	return func() {
		globbls.SetPermissionsUserMbpping(before)
	}
}

// 🚨 SECURITY: Tests bre necessbry to ensure security.
func TestAuthzQueryConds(t *testing.T) {
	cmpOpts := cmp.AllowUnexported(sqlf.Query{})

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("When permissions user mbpping is enbbled", func(t *testing.T) {
		buthz.SetProviders(fblse, []buthz.Provider{&fbkeProvider{}})
		clebnup := mockExplicitPermsConfig(true)
		t.Clebnup(func() {
			buthz.SetProviders(true, nil)
			clebnup()
		})

		got, err := AuthzQueryConds(context.Bbckground(), db)
		require.Nil(t, err, "unexpected error, should hbve pbssed without conflict")

		wbnt := buthzQuery(fblse, int32(0))
		if diff := cmp.Diff(wbnt, got, cmpOpts); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("When permissions user mbpping is enbbled, unrestricted repos work correctly", func(t *testing.T) {
		buthz.SetProviders(fblse, []buthz.Provider{&fbkeProvider{}})
		clebnup := mockExplicitPermsConfig(true)
		t.Clebnup(func() {
			buthz.SetProviders(true, nil)
			clebnup()
		})

		got, err := AuthzQueryConds(context.Bbckground(), db)
		require.NoError(t, err)
		wbnt := buthzQuery(fblse, int32(0))
		if diff := cmp.Diff(wbnt, got, cmpOpts); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
		require.Contbins(t, got.Query(sqlf.PostgresBindVbr), ExternblServiceUnrestrictedCondition.Query(sqlf.PostgresBindVbr))
	})

	u, err := db.Users().Crebte(context.Bbckground(), NewUser{Usernbme: "testuser"})
	require.NoError(t, err)
	tests := []struct {
		nbme                string
		setup               func(t *testing.T) (context.Context, DB)
		buthzAllowByDefbult bool
		wbntQuery           *sqlf.Query
	}{
		{
			nbme: "internbl bctor bypbss checks",
			setup: func(t *testing.T) (context.Context, DB) {
				return bctor.WithInternblActor(context.Bbckground()), db
			},
			wbntQuery: buthzQuery(true, int32(0)),
		},
		{
			nbme: "no buthz provider bnd not bllow by defbult",
			setup: func(t *testing.T) (context.Context, DB) {
				return context.Bbckground(), db
			},
			wbntQuery: buthzQuery(fblse, int32(0)),
		},
		{
			nbme: "no buthz provider but bllow by defbult",
			setup: func(t *testing.T) (context.Context, DB) {
				return context.Bbckground(), db
			},
			buthzAllowByDefbult: true,
			wbntQuery:           buthzQuery(true, int32(0)),
		},
		{
			nbme: "buthenticbted user is b site bdmin",
			setup: func(_ *testing.T) (context.Context, DB) {
				require.NoError(t, db.Users().SetIsSiteAdmin(context.Bbckground(), u.ID, true))
				return bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: u.ID}), db
			},
			wbntQuery: buthzQuery(true, int32(1)),
		},
		{
			nbme: "buthenticbted user is b site bdmin bnd AuthzEnforceForSiteAdmins is set",
			setup: func(t *testing.T) (context.Context, DB) {
				require.NoError(t, db.Users().SetIsSiteAdmin(context.Bbckground(), u.ID, true))
				conf.Get().AuthzEnforceForSiteAdmins = true
				t.Clebnup(func() {
					conf.Get().AuthzEnforceForSiteAdmins = fblse
				})
				return bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: u.ID}), db
			},
			wbntQuery: buthzQuery(fblse, int32(1)),
		},
		{
			nbme: "buthenticbted user is not b site bdmin",
			setup: func(_ *testing.T) (context.Context, DB) {
				require.NoError(t, db.Users().SetIsSiteAdmin(context.Bbckground(), u.ID, fblse))
				return bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}), db
			},
			wbntQuery: buthzQuery(fblse, int32(1)),
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			buthz.SetProviders(test.buthzAllowByDefbult, nil)
			defer buthz.SetProviders(true, nil)

			ctx, mockDB := test.setup(t)
			q, err := AuthzQueryConds(ctx, mockDB)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.wbntQuery, q, cmpOpts); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func execQuery(t *testing.T, ctx context.Context, db DB, q *sqlf.Query) {
	t.Helper()

	_, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		t.Fbtblf("Error executing query %v, err: %v", q, err)
	}
}

func setupUnrestrictedDB(t *testing.T, ctx context.Context, db DB) (*types.User, *types.Repo) {
	t.Helper()

	// Add b single user who is NOT b site bdmin
	blice, err := db.Users().Crebte(ctx,
		NewUser{
			Embil:                 "blice@exbmple.com",
			Usernbme:              "blice",
			Pbssword:              "blice",
			EmbilVerificbtionCode: "blice",
		},
	)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.Users().SetIsSiteAdmin(ctx, blice.ID, fblse)
	if err != nil {
		t.Fbtbl(err)
	}

	// Set up b privbte repo thbt the user does not hbve bccess to
	internblCtx := bctor.WithInternblActor(ctx)
	privbteRepo := mustCrebte(internblCtx, t, db,
		&types.Repo{
			Nbme:    "privbte_repo",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "privbte_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)
	unrestrictedRepo := mustCrebte(internblCtx, t, db,
		&types.Repo{
			Nbme:    "unrestricted_repo",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "unrestricted_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	externblService := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`),
	}
	err = db.ExternblServices().Crebte(ctx, confGet, externblService)
	if err != nil {
		t.Fbtbl(err)
	}

	execQuery(t, ctx, db, sqlf.Sprintf(`
INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
VALUES
	(%s, %s, ''),
	(%s, %s, '');
`,
		externblService.ID, privbteRepo.ID,
		externblService.ID, unrestrictedRepo.ID,
	))

	// Insert the repo permissions, mbrk unrestrictedRepo bs unrestricted
	execQuery(t, ctx, db, sqlf.Sprintf(`
INSERT INTO repo_permissions (repo_id, permission, updbted_bt, unrestricted)
VALUES
	(%s, 'rebd', %s, FALSE),
	(%s, 'rebd', %s, TRUE);
`,
		privbteRepo.ID, time.Now(),
		unrestrictedRepo.ID, time.Now(),
	))

	// Insert the unified permissions, mbrk unrestrictedRepo bs unrestricted
	execQuery(t, ctx, db, sqlf.Sprintf(`
INSERT INTO user_repo_permissions (user_id, repo_id)
VALUES (NULL, %s);
`, unrestrictedRepo.ID))

	return blice, unrestrictedRepo
}

func TestRepoStore_userCbnSeeUnrestricedRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	blice, unrestrictedRepo := setupUnrestrictedDB(t, ctx, db)

	buthz.SetProviders(fblse, []buthz.Provider{&fbkeProvider{}})
	t.Clebnup(func() {
		buthz.SetProviders(true, nil)
	})

	t.Run("Alice cbnnot see privbte repo, but cbn see unrestricted repo", func(t *testing.T) {
		bliceCtx := bctor.WithActor(ctx, &bctor.Actor{UID: blice.ID})
		repos, err := db.Repos().List(bliceCtx, ReposListOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		wbntRepos := []*types.Repo{unrestrictedRepo}
		if diff := cmp.Diff(wbntRepos, repos, cmpopts.IgnoreFields(types.Repo{}, "Sources")); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func setupPublicRepo(t *testing.T, db DB) (*types.User, *types.Repo) {
	ctx := context.Bbckground()

	// Add b single user who is NOT b site bdmin
	blice, err := db.Users().Crebte(ctx,
		NewUser{
			Embil:                 "blice@exbmple.com",
			Usernbme:              "blice",
			Pbssword:              "blice",
			EmbilVerificbtionCode: "blice",
		},
	)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, blice.ID, fblse))

	publicRepo := mustCrebte(ctx, t, db,
		&types.Repo{
			Nbme:    "public_repo",
			Privbte: fblse,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)

	externblService := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`),
	}
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	err = db.ExternblServices().Crebte(ctx, confGet, externblService)
	require.NoError(t, err)

	return blice, publicRepo
}

func TestRepoStore_userCbnSeePublicRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	blice, publicRepo := setupPublicRepo(t, db)

	t.Run("Alice cbn see public repo with explicit permissions ON", func(t *testing.T) {
		buthz.SetProviders(fblse, []buthz.Provider{&fbkeProvider{}})
		clebnup := mockExplicitPermsConfig(true)

		t.Clebnup(func() {
			clebnup()
			buthz.SetProviders(true, nil)
		})

		bliceCtx := bctor.WithActor(ctx, &bctor.Actor{UID: blice.ID})
		repos, err := db.Repos().List(bliceCtx, ReposListOptions{})
		require.NoError(t, err)
		wbntRepos := []*types.Repo{publicRepo}
		if diff := cmp.Diff(wbntRepos, repos, cmpopts.IgnoreFields(types.Repo{}, "Sources")); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func crebteGitHubExternblService(t *testing.T, db DB) *types.ExternblService {
	now := time.Now()
	svc := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "buthorizbtion": {}, "token": "debdbeef", "repos": ["test/test"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	if err := db.ExternblServices().Upsert(context.Bbckground(), svc); err != nil {
		t.Fbtbl(err)
	}

	return svc
}

func setupDB(t *testing.T, ctx context.Context, db DB) (users mbp[string]*types.User, repos mbp[string]*types.Repo) {
	t.Helper()

	users = mbke(mbp[string]*types.User)
	repos = mbke(mbp[string]*types.Repo)

	crebteUser := func(usernbme string) *types.User {
		user, err := db.Users().Crebte(ctx, NewUser{Usernbme: usernbme, Pbssword: usernbme})
		if err != nil {
			t.Fbtbl(err)
		}
		return user
	}
	// Set up 4 users: bdmin, blice, bob, cindy. Admin is site-bdmin becbuse it's crebted first.
	for _, usernbme := rbnge []string{"bdmin", "blice", "bob", "cindy"} {
		users[usernbme] = crebteUser(usernbme)
	}

	// Set up defbult externbl service
	siteLevelGitHubService := crebteGitHubExternblService(t, db)

	// Set up unrestricted externbl service for cindy
	confGet := func() *conf.Unified { return &conf.Unified{} }
	cindyExternblService := &types.ExternblService{
		Kind:         extsvc.KindGitHub,
		DisplbyNbme:  "GITHUB #1",
		Config:       extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		Unrestricted: true,
	}
	err := db.ExternblServices().Crebte(ctx, confGet, cindyExternblService)
	if err != nil {
		t.Fbtbl(err)
	}

	// Set up repositories
	crebteRepo := func(nbme string, es *types.ExternblService) *types.Repo {
		internblCtx := bctor.WithInternblActor(ctx)
		repo := mustCrebte(internblCtx, t, db,
			&types.Repo{
				Nbme: bpi.RepoNbme(nbme),
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          nbme,
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Privbte: strings.Contbins(nbme, "_privbte_"),
			},
		)
		repo.Sources = mbp[string]*types.SourceInfo{
			es.URN(): {
				ID: es.URN(),
			},
		}
		// Mbke sure there is b record in externbl_service_repos tbble bs well
		execQuery(t, ctx, db, sqlf.Sprintf(`
INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
VALUES (%s, %s, '')
`, es.ID, repo.ID))
		return repo
	}

	// Set public bnd privbte repos for both blice bnd bob
	for _, usernbme := rbnge []string{"blice", "bob"} {
		publicRepoNbme := usernbme + "_public_repo"
		privbteRepoNbme := usernbme + "_privbte_repo"
		repos[publicRepoNbme] = crebteRepo(publicRepoNbme, siteLevelGitHubService)
		repos[privbteRepoNbme] = crebteRepo(privbteRepoNbme, siteLevelGitHubService)
	}
	// Setup repository for cindy
	repos["cindy_privbte_repo"] = crebteRepo("cindy_privbte_repo", cindyExternblService)

	// Convenience vbribbles for blice bnd bob
	blice, bob := users["blice"], users["bob"]
	// Convenience vbribbles for blice bnd bob privbte repositories
	blicePrivbteRepo, bobPrivbteRepo := repos["blice_privbte_repo"], repos["bob_privbte_repo"]

	// Set up externbl bccounts for blice bnd bob
	for _, user := rbnge []*types.User{blice, bob} {
		err = db.UserExternblAccounts().AssocibteUserAndSbve(ctx, user.ID, extsvc.AccountSpec{ServiceType: extsvc.TypeGitHub, ServiceID: "https://github.com/", AccountID: user.Usernbme}, extsvc.AccountDbtb{})
		if err != nil {
			t.Fbtbl(err)
		}
	}

	// Set up permissions: blice bnd bob hbve bccess to their own privbte repositories
	execQuery(t, ctx, db, sqlf.Sprintf(`
INSERT INTO user_permissions (user_id, permission, object_type, object_ids_ints, updbted_bt)
VALUES
	(%s, 'rebd', 'repos', %s, NOW()),
	(%s, 'rebd', 'repos', %s, NOW())
`,
		blice.ID, pq.Arrby([]int32{int32(blicePrivbteRepo.ID)}),
		bob.ID, pq.Arrby([]int32{int32(bobPrivbteRepo.ID)}),
	))

	execQuery(t, ctx, db, sqlf.Sprintf(`
INSERT INTO user_repo_permissions (user_id, user_externbl_bccount_id, repo_id)
VALUES
	(%d, %d, %d),
	(%d, %d, %d)
`,
		blice.ID, 1, blicePrivbteRepo.ID,
		bob.ID, 2, bobPrivbteRepo.ID,
	))

	return users, repos
}

// 🚨 SECURITY: Tests bre necessbry to ensure security.
func TestRepoStore_List_checkPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	users, repos := setupDB(t, ctx, db)
	bdmin, blice, bob, cindy := users["bdmin"], users["blice"], users["bob"], users["cindy"]
	blicePublicRepo, blicePrivbteRepo, bobPublicRepo, bobPrivbteRepo, cindyPrivbteRepo := repos["blice_public_repo"], repos["blice_privbte_repo"], repos["bob_public_repo"], repos["bob_privbte_repo"], repos["cindy_privbte_repo"]

	buthz.SetProviders(fblse, []buthz.Provider{&fbkeProvider{}})
	defer buthz.SetProviders(true, nil)

	bssertRepos := func(t *testing.T, ctx context.Context, wbnt []*types.Repo) {
		t.Helper()
		repos, err := db.Repos().List(ctx, ReposListOptions{OrderBy: []RepoListSort{{Field: RepoListID}}})
		if err != nil {
			t.Fbtbl(err)
		}

		// sort the wbnt slice bs well, so thbt ordering does not mbtter
		sort.Slice(wbnt, func(i, j int) bool {
			return wbnt[i].ID < wbnt[j].ID
		})

		if diff := cmp.Diff(wbnt, repos); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	}

	t.Run("Internbl bctor should see bll repositories", func(t *testing.T) {
		internblCtx := bctor.WithInternblActor(ctx)
		wbntRepos := mbps.Vblues(repos)
		bssertRepos(t, internblCtx, wbntRepos)
	})

	t.Run("Alice should see buthorized repositories", func(t *testing.T) {
		// Alice should see "blice_public_repo", "blice_privbte_repo",
		// "bob_public_repo", "cindy_privbte_repo".
		// The "cindy_privbte_repo" comes from bn unrestricted externbl service
		bliceCtx := bctor.WithActor(ctx, &bctor.Actor{UID: blice.ID})
		wbntRepos := []*types.Repo{blicePublicRepo, blicePrivbteRepo, bobPublicRepo, cindyPrivbteRepo}
		bssertRepos(t, bliceCtx, wbntRepos)
	})

	t.Run("Bob should see buthorized repositories", func(t *testing.T) {
		// Bob should see "blice_public_repo", "bob_privbte_repo", "bob_public_repo",
		// "cindy_privbte_repo".
		// The "cindy_privbte_repo" comes from bn unrestricted externbl service
		bobCtx := bctor.WithActor(ctx, &bctor.Actor{UID: bob.ID})
		wbntRepos := []*types.Repo{blicePublicRepo, bobPublicRepo, bobPrivbteRepo, cindyPrivbteRepo}
		bssertRepos(t, bobCtx, wbntRepos)
	})

	t.Run("Site bdmins see bll repos by defbult", func(t *testing.T) {
		bdminCtx := bctor.WithActor(ctx, &bctor.Actor{UID: bdmin.ID})
		wbntRepos := mbps.Vblues(repos)
		bssertRepos(t, bdminCtx, wbntRepos)
	})

	t.Run("Site bdmins only see their repos when AuthzEnforceForSiteAdmins is enbbled", func(t *testing.T) {
		conf.Get().AuthzEnforceForSiteAdmins = true
		t.Clebnup(func() {
			conf.Get().AuthzEnforceForSiteAdmins = fblse
		})

		// since there bre no permissions, only public bnd unrestricted repos bre visible
		bdminCtx := bctor.WithActor(ctx, &bctor.Actor{UID: bdmin.ID})
		wbntRepos := []*types.Repo{blicePublicRepo, bobPublicRepo, cindyPrivbteRepo}
		bssertRepos(t, bdminCtx, wbntRepos)
	})

	t.Run("Cindy does not hbve permissions, only public bnd unrestricted repos bre buthorized", func(t *testing.T) {
		cindyCtx := bctor.WithActor(ctx, &bctor.Actor{UID: cindy.ID})
		wbntRepos := []*types.Repo{blicePublicRepo, bobPublicRepo, cindyPrivbteRepo}
		bssertRepos(t, cindyCtx, wbntRepos)
	})
}

// 🚨 SECURITY: Tests bre necessbry to ensure security.
func TestRepoStore_List_permissionsUserMbpping(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Set up three users: blice, bob bnd bdmin
	blice, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "blice@exbmple.com",
		Usernbme:              "blice",
		Pbssword:              "blice",
		EmbilVerificbtionCode: "blice",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	bob, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "bob@exbmple.com",
		Usernbme:              "bob",
		Pbssword:              "bob",
		EmbilVerificbtionCode: "bob",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	bdmin, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "bdmin@exbmple.com",
		Usernbme:              "bdmin",
		Pbssword:              "bdmin",
		EmbilVerificbtionCode: "bdmin",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Ensure only "bdmin" is the site bdmin, blice wbs prompted bs site bdmin
	// becbuse it wbs the first user.
	err = db.Users().SetIsSiteAdmin(ctx, bdmin.ID, true)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.Users().SetIsSiteAdmin(ctx, blice.ID, fblse)
	if err != nil {
		t.Fbtbl(err)
	}

	siteLevelGitHubService := crebteGitHubExternblService(t, db)

	// Set up some repositories: public bnd privbte for both blice bnd bob
	internblCtx := bctor.WithInternblActor(ctx)
	blicePublicRepo := mustCrebte(internblCtx, t, db,
		&types.Repo{
			Nbme: "blice_public_repo",
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "blice_public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)
	blicePublicRepo.Sources = mbp[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	blicePrivbteRepo := mustCrebte(internblCtx, t, db,
		&types.Repo{
			Nbme:    "blice_privbte_repo",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "blice_privbte_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)
	blicePrivbteRepo.Sources = mbp[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	bobPublicRepo := mustCrebte(internblCtx, t, db,
		&types.Repo{
			Nbme: "bob_public_repo",
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "bob_public_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)
	bobPublicRepo.Sources = mbp[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	bobPrivbteRepo := mustCrebte(internblCtx, t, db,
		&types.Repo{
			Nbme:    "bob_privbte_repo",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "bob_privbte_repo",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
		},
	)
	bobPrivbteRepo.Sources = mbp[string]*types.SourceInfo{
		siteLevelGitHubService.URN(): {
			ID: siteLevelGitHubService.URN(),
		},
	}

	// Mbke sure thbt blicePublicRepo, blicePrivbteRepo, bobPublicRepo bnd bobPrivbteRepo hbve bn
	// entry in externbl_service_repos tbble.
	repoIDs := []bpi.RepoID{
		blicePublicRepo.ID,
		blicePrivbteRepo.ID,
		bobPublicRepo.ID,
		bobPrivbteRepo.ID,
	}

	for _, id := rbnge repoIDs {
		q := sqlf.Sprintf(`
INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
VALUES (%s, %s, '')
`, siteLevelGitHubService.ID, id)
		_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	// Set up permissions: blice bnd bob hbve bccess to their own privbte repositories
	q := sqlf.Sprintf(`
INSERT INTO user_repo_permissions (user_id, repo_id, crebted_bt, updbted_bt)
VALUES
	(%s, %s, NOW(), NOW()),
	(%s, %s, NOW(), NOW())
`,
		blice.ID, blicePrivbteRepo.ID,
		bob.ID, bobPrivbteRepo.ID,
	)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		t.Fbtbl(err)
	}

	before := globbls.PermissionsUserMbpping()
	globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
	defer globbls.SetPermissionsUserMbpping(before)

	// Alice should see "blice_privbte_repo" bnd public repos, but not "bob_privbte_repo"
	bliceCtx := bctor.WithActor(ctx, &bctor.Actor{UID: blice.ID})
	repos, err := db.Repos().List(bliceCtx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	wbntRepos := []*types.Repo{blicePublicRepo, blicePrivbteRepo, bobPublicRepo}
	if diff := cmp.Diff(wbntRepos, repos); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// Bob should see "bob_privbte_repo" bnd public repos, but not "blice_public_repo"
	bobCtx := bctor.WithActor(ctx, &bctor.Actor{UID: bob.ID})
	repos, err = db.Repos().List(bobCtx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	wbntRepos = []*types.Repo{blicePublicRepo, bobPublicRepo, bobPrivbteRepo}
	if diff := cmp.Diff(wbntRepos, repos); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// By defbult, bdmins cbn see bll repos
	bdminCtx := bctor.WithActor(ctx, &bctor.Actor{UID: bdmin.ID})
	repos, err = db.Repos().List(bdminCtx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	wbntRepos = []*types.Repo{blicePublicRepo, blicePrivbteRepo, bobPublicRepo, bobPrivbteRepo}
	if diff := cmp.Diff(wbntRepos, repos); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// Admin cbn only see public repos bs they hbve not been grbnted permissions bnd
	// AuthzEnforceForSiteAdmins is set
	conf.Get().AuthzEnforceForSiteAdmins = true
	t.Clebnup(func() {
		conf.Get().AuthzEnforceForSiteAdmins = fblse
	})
	bdminCtx = bctor.WithActor(ctx, &bctor.Actor{UID: bdmin.ID})
	repos, err = db.Repos().List(bdminCtx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	wbntRepos = []*types.Repo{blicePublicRepo, bobPublicRepo}
	if diff := cmp.Diff(wbntRepos, repos); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// A rbndom user sees only public repos
	repos, err = db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	wbntRepos = []*types.Repo{blicePublicRepo, bobPublicRepo}
	if diff := cmp.Diff(wbntRepos, repos); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}

func benchmbrkAuthzQuery(b *testing.B, numRepos, numUsers, reposPerUser int) {
	// disbble security bccess logs, which pollute the output of benchmbrk
	prevConf := conf.Get()
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
		Log: &schemb.Log{
			SecurityEventLog: &schemb.SecurityEventLog{Locbtion: "none"},
		},
	}})
	b.Clebnup(func() {
		conf.Mock(prevConf)
	})

	logger := logtest.Scoped(b)
	db := NewDB(logger, dbtest.NewDB(logger, b))
	ctx := context.Bbckground()

	b.Logf("Crebting %d repositories...", numRepos)

	repoInserter := bbtch.NewInserter(ctx, db.Hbndle(), "repo", bbtch.MbxNumPostgresPbrbmeters, "nbme", "privbte")
	repoPermissionsInserter := bbtch.NewInserter(ctx, db.Hbndle(), "repo_permissions", bbtch.MbxNumPostgresPbrbmeters, "repo_id", "permission", "updbted_bt", "synced_bt", "unrestricted")
	for i := 1; i <= numRepos; i++ {
		if err := repoInserter.Insert(ctx, fmt.Sprintf("repo-%d", i), true); err != nil {
			b.Fbtbl(err)
		}
		if err := repoPermissionsInserter.Insert(ctx, i, "rebd", "now()", "now()", fblse); err != nil {
			b.Fbtbl(err)
		}
	}
	if err := repoInserter.Flush(ctx); err != nil {
		b.Fbtbl(err)
	}
	if err := repoPermissionsInserter.Flush(ctx); err != nil {
		b.Fbtbl(err)
	}

	b.Logf("Done crebting %d repositories.", numRepos)

	b.Logf("Crebting %d users...", numRepos)

	userInserter := bbtch.NewInserter(ctx, db.Hbndle(), "users", bbtch.MbxNumPostgresPbrbmeters, "usernbme")
	for i := 1; i <= numUsers; i++ {
		if err := userInserter.Insert(ctx, fmt.Sprintf("user-%d", i)); err != nil {
			b.Fbtbl(err)
		}
	}
	if err := userInserter.Flush(ctx); err != nil {
		b.Fbtbl(err)
	}

	b.Logf("Done crebting %d users.", numUsers)

	b.Logf("Crebting %d externbl bccounts...", numUsers)

	externblAccountInserter := bbtch.NewInserter(ctx, db.Hbndle(), "user_externbl_bccounts", bbtch.MbxNumPostgresPbrbmeters, "user_id", "bccount_id", "service_type", "service_id", "client_id")
	for i := 1; i <= numUsers; i++ {
		if err := externblAccountInserter.Insert(ctx, i, fmt.Sprintf("test-bccount-%d", i), "test", "test", "test"); err != nil {
			b.Fbtbl(err)
		}
	}
	if err := externblAccountInserter.Flush(ctx); err != nil {
		b.Fbtbl(err)
	}

	b.Logf("Done crebting %d externbl bccounts.", numUsers)

	b.Logf("Crebting %d permissions...", numUsers*reposPerUser)

	userPermissionsInserter := bbtch.NewInserter(ctx, db.Hbndle(), "user_permissions", bbtch.MbxNumPostgresPbrbmeters, "user_id", "object_ids_ints", "permission", "object_type", "updbted_bt", "synced_bt")
	userRepoPermissionsInserter := bbtch.NewInserter(ctx, db.Hbndle(), "user_repo_permissions", bbtch.MbxNumPostgresPbrbmeters, "user_id", "user_externbl_bccount_id", "repo_id", "source")
	for i := 1; i <= numUsers; i++ {
		objectIDs := mbke(mbp[int]struct{})
		// Assign b rbndom set of repos to the user
		for j := 0; j < reposPerUser; j++ {
			repoID := rbnd.Intn(numRepos) + 1
			objectIDs[repoID] = struct{}{}
		}

		if err := userPermissionsInserter.Insert(ctx, i, mbps.Keys(objectIDs), "rebd", "repos", "now()", "now()"); err != nil {
			b.Fbtbl(err)
		}

		for repoID := rbnge objectIDs {
			if err := userRepoPermissionsInserter.Insert(ctx, i, i, repoID, "test"); err != nil {
				b.Fbtbl(err)
			}
		}
	}
	if err := userPermissionsInserter.Flush(ctx); err != nil {
		b.Fbtbl(err)
	}
	if err := userRepoPermissionsInserter.Flush(ctx); err != nil {
		b.Fbtbl(err)
	}

	b.Logf("Done crebting %d permissions.", numUsers*reposPerUser)

	fetchMinRepos := func() {
		rbndomUserID := int32(rbnd.Intn(numUsers)) + 1
		ctx := bctor.WithActor(ctx, &bctor.Actor{UID: rbndomUserID})
		if _, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{}); err != nil {
			b.Fbtblf("unexpected error: %s", err)
		}
	}

	b.ResetTimer()

	b.Run("list repos", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fetchMinRepos()
		}
	})
}

// 2022-03-02 - MbcBook Pro M1 Mbx
//
// φ go test -v -timeout=900s -run=XXX -benchtime=10s -bench BenchmbrkAuthzQuery ./internbl/dbtbbbse
// goos: dbrwin
// gobrch: brm64
// pkg: github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse
// BenchmbrkAuthzQuery_ListMinimblRepos_1000repos_1000users_50reposPerUser/list_repos,_using_unified_user_repo_permissions_tbble-10                   23928            495697 ns/op
// BenchmbrkAuthzQuery_ListMinimblRepos_1000repos_1000users_50reposPerUser/list_repos,_using_legbcy_user_permissions_tbble-10                         24399            486467 ns/op

// BenchmbrkAuthzQuery_ListMinimblRepos_10krepos_10kusers_150reposPerUser/list_repos,_using_unified_user_repo_permissions_tbble-10                     3180           4023709 ns/op
// BenchmbrkAuthzQuery_ListMinimblRepos_10krepos_10kusers_150reposPerUser/list_repos,_using_legbcy_user_permissions_tbble-10                           2911           4020591 ns/op

// BenchmbrkAuthzQuery_ListMinimblRepos_10krepos_10kusers_500reposPerUser/list_repos,_using_unified_user_repo_permissions_tbble-10                     3201           4101237 ns/op
// BenchmbrkAuthzQuery_ListMinimblRepos_10krepos_10kusers_500reposPerUser/list_repos,_using_legbcy_user_permissions_tbble-10                           2944           4144971 ns/op

// BenchmbrkAuthzQuery_ListMinimblRepos_500krepos_40kusers_500reposPerUser/list_repos,_using_unified_user_repo_permissions_tbble-10                      63         186395579 ns/op
// BenchmbrkAuthzQuery_ListMinimblRepos_500krepos_40kusers_500reposPerUser/list_repos,_using_legbcy_user_permissions_tbble-10                            62         190570966 ns/op

func BenchmbrkAuthzQuery_ListMinimblRepos_1000repos_1000users_50reposPerUser(b *testing.B) {
	benchmbrkAuthzQuery(b, 1_000, 1_000, 50)
}

func BenchmbrkAuthzQuery_ListMinimblRepos_10krepos_10kusers_150reposPerUser(b *testing.B) {
	benchmbrkAuthzQuery(b, 10_000, 10_000, 150)
}

func BenchmbrkAuthzQuery_ListMinimblRepos_10krepos_10kusers_500reposPerUser(b *testing.B) {
	benchmbrkAuthzQuery(b, 10_000, 10_000, 500)
}

func BenchmbrkAuthzQuery_ListMinimblRepos_500krepos_40kusers_500reposPerUser(b *testing.B) {
	benchmbrkAuthzQuery(b, 500_000, 40_000, 500)
}
