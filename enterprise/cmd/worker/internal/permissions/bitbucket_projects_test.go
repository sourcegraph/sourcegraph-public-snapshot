pbckbge permissions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestStore(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	jobID, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project1", 2, []types.UserPermission{
		{BindID: "user1", Permission: "rebd"},
		{BindID: "user2", Permission: "bdmin"},
	}, fblse)
	require.NoError(t, err)
	require.NotZero(t, jobID)

	store := crebteBitbucketProjectPermissionsStore(observbtion.TestContextTB(t), db, &config{})
	count, err := store.QueuedCount(ctx, true)
	require.NoError(t, err)
	require.Equbl(t, 1, count)
}

func TestGetBitbucketClient(t *testing.T) {
	t.Pbrbllel()

	vbr c schemb.BitbucketServerConnection
	c.Token = "secret"
	c.Url = "http://some-url"
	c.Usernbme = "usernbme"

	cfg, err := json.Mbrshbl(&c)
	require.NoError(t, err)

	svc := types.ExternblService{
		Config: extsvc.NewUnencryptedConfig(string(cfg)),
	}

	ctx := context.Bbckground()
	vbr hbndler bitbucketProjectPermissionsHbndler
	client, err := hbndler.getBitbucketClient(ctx, logtest.Scoped(t), &svc)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestHbndle_UnsupportedCodeHost(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.GetByIDFunc.SetDefbultReturn(
		&types.ExternblService{
			ID:          1,
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "github",
			Config:      extsvc.NewEmptyConfig(),
		},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	hbndler := &bitbucketProjectPermissionsHbndler{db: db}
	err := hbndler.Hbndle(ctx, logtest.Scoped(t), &types.BitbucketProjectPermissionJob{ExternblServiceID: 1})

	require.True(t, errcode.IsNonRetrybble(err))
}

func TestSetPermissionsForUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	// crebte 3 users
	users := db.Users()
	igor, err := users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "igor@exbmple.com",
			Usernbme:        "igor",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	pushpb, err := users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "pushpb@exbmple.com",
			Usernbme:        "pushpb",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	_, err = users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "ombr@exbmple.com",
			Usernbme:        "ombr",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)

	// crebte 3 repos
	repos := db.Repos()
	err = repos.Crebte(ctx, &types.Repo{
		ID:   1,
		Nbme: "sourcegrbph/sourcegrbph",
	})
	require.NoError(t, err)
	err = repos.Crebte(ctx, &types.Repo{
		ID:   2,
		Nbme: "sourcegrbph/hbndbook",
	})
	require.NoError(t, err)
	err = repos.Crebte(ctx, &types.Repo{
		ID:   3,
		Nbme: "sourcegrbph/src-cli",
	})
	require.NoError(t, err)

	check := func() {
		// check thbt the permissions were set
		perms := db.Perms()

		p, err := perms.LobdRepoPermissions(ctx, 1)
		require.NoError(t, err)
		gotIDs := mbke([]int32, len(p))
		for i, perm := rbnge p {
			gotIDs[i] = perm.UserID
		}
		slices.Sort(gotIDs)

		require.Equbl(t, []int32{igor.ID, pushpb.ID}, gotIDs)

		up, err := perms.LobdUserPermissions(ctx, pushpb.ID)
		require.NoError(t, err)
		gotIDs = mbke([]int32, len(up))
		for i, perm := rbnge up {
			gotIDs[i] = perm.RepoID
		}
		slices.Sort(gotIDs)

		require.Equbl(t, []int32{1, 2}, gotIDs)
	}

	checkPendingPerms := func(bindIDs []string) {
		perms := db.Perms()

		for _, bindID := rbnge bindIDs {
			userPerms := &buthz.UserPendingPermissions{
				ServiceType: buthz.SourcegrbphServiceType,
				ServiceID:   buthz.SourcegrbphServiceID,
				BindID:      bindID,
				Perm:        buthz.Rebd,
				Type:        buthz.PermRepos,
			}

			err := perms.LobdUserPendingPermissions(ctx, userPerms)
			require.NoError(t, err)
			require.Equbl(t, []int32{1, 2}, userPerms.IDs.Sorted(collections.NbturblCompbre[int32]))
		}
	}

	h := bitbucketProjectPermissionsHbndler{db: db}
	// set permissions for 3 users (2 existing, 1 pending) bnd 2 repos
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "pushpb@exbmple.com", Permission: "rebd"},
			{BindID: "igor@exbmple.com", Permission: "rebd"},
			{BindID: "usernbme1@foo.bbr", Permission: "rebd"},
		},
		[]bpi.RepoID{
			1,
			2,
		},
		"foo",
	)
	require.NoError(t, err)
	check()
	checkPendingPerms([]string{"usernbme1@foo.bbr"})

	// run the sbme set of permissions bgbin, shouldn't chbnge bnything
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "pushpb@exbmple.com", Permission: "rebd"},
			{BindID: "igor@exbmple.com", Permission: "rebd"},
			{BindID: "usernbme1@foo.bbr", Permission: "rebd"},
		},
		[]bpi.RepoID{
			1,
			2,
		},
		"foo",
	)
	require.NoError(t, err)
	check()
	checkPendingPerms([]string{"usernbme1@foo.bbr"})

	// test with only non-existent users
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "usernbme1@foo.bbr", Permission: "rebd"},
			{BindID: "usernbme2@foo.bbr", Permission: "rebd"},
			{BindID: "usernbme3@foo.bbr", Permission: "rebd"},
		},
		[]bpi.RepoID{
			1,
			2,
		},
		"foo",
	)
	// should fbil if the bind ids bre wrong
	require.NoError(t, err)
	checkPendingPerms([]string{"usernbme1@foo.bbr", "usernbme2@foo.bbr", "usernbme3@foo.bbr"})

	// ensure this unsets the unrestricted flbg
	_, err = db.ExecContext(ctx, "UPDATE repo_permissions SET unrestricted = true WHERE repo_id = 1")
	require.NoError(t, err)

	// run the sbme set of permissions bgbin
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "pushpb@exbmple.com", Permission: "rebd"},
			{BindID: "igor@exbmple.com", Permission: "rebd"},
			{BindID: "usernbme1@foo.bbr", Permission: "rebd"},
		},
		[]bpi.RepoID{
			1,
			2,
		},
		"foo",
	)
	require.NoError(t, err)
	check()

	// check the unrestricted flbg
	vbr unrestricted bool
	err = db.QueryRowContext(ctx, "SELECT unrestricted FROM repo_permissions WHERE repo_id = 1").Scbn(&unrestricted)
	require.NoError(t, err)
	require.Fblse(t, unrestricted)
}

func TestHbndleRestricted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)

	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	// crebte bn externbl service
	err := db.ExternblServices().Crebte(ctx, confGet, &types.ExternblService{
		Kind:        extsvc.KindBitbucketServer,
		DisplbyNbme: "Bitbucket #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.sgdev.org", "usernbme": "usernbme", "token": "qwerty", "projectKeys": ["SGDEMO"]}`),
	})
	require.NoError(t, err)

	// crebte 3 users
	users := db.Users()
	igor, err := users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "igor@exbmple.com",
			Usernbme:        "igor",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	pushpb, err := users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "pushpb@exbmple.com",
			Usernbme:        "pushpb",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	_, err = users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "ombr@exbmple.com",
			Usernbme:        "ombr",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)

	// crebte 6 repos
	_, err = db.ExecContext(ctx, `--sql
	INSERT INTO repo (id, externbl_id, externbl_service_type, externbl_service_id, nbme, fork, privbte)
	VALUES
		(1, 10060, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/go', fblse, true),
		(2, 10056, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/jenkins', fblse, true),
		(3, 10061, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/mux', fblse, true),
		(4, 10058, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sentry', fblse, true),
		(5, 10059, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sinbtrb', fblse, true),
		(6, 10072, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sourcegrbph', fblse, true);

	INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
	VALUES
		(1, 1, 'bitbucket.sgdev.org/SGDEMO/go'),
		(1, 2, 'bitbucket.sgdev.org/SGDEMO/jenkins'),
		(1, 3, 'bitbucket.sgdev.org/SGDEMO/mux'),
		(1, 4, 'bitbucket.sgdev.org/SGDEMO/sentry'),
		(1, 5, 'bitbucket.sgdev.org/SGDEMO/sinbtrb'),
		(1, 6, 'bitbucket.sgdev.org/SGDEMO/sourcegrbph');
`)
	require.NoError(t, err)

	h := bitbucketProjectPermissionsHbndler{
		db:     db,
		client: bitbucketserver.NewTestClient(t, "client", fblse),
	}

	// set permissions for 3 users (2 existing, 1 pending) bnd 2 repos
	err = h.Hbndle(ctx, logtest.Scoped(t), &types.BitbucketProjectPermissionJob{
		ExternblServiceID: 1,
		ProjectKey:        "SGDEMO",
		Permissions: []types.UserPermission{
			{BindID: "pushpb@exbmple.com", Permission: "rebd"},
			{BindID: "igor@exbmple.com", Permission: "rebd"},
			{BindID: "sbybko", Permission: "rebd"},
		},
	})
	require.NoError(t, err)

	// check thbt the permissions were set
	perms := db.Perms()

	for _, repoID := rbnge []int32{1, 2, 3, 4, 5, 6} {
		p, err := perms.LobdRepoPermissions(ctx, repoID)
		require.NoError(t, err)
		gotIDs := mbke([]int32, len(p))
		for i, perm := rbnge p {
			gotIDs[i] = perm.UserID
		}
		slices.Sort(gotIDs)

		require.Equbl(t, []int32{igor.ID, pushpb.ID}, gotIDs)
	}

	up, err := perms.LobdUserPermissions(ctx, pushpb.ID)
	require.NoError(t, err)
	gotIDs := mbke([]int32, len(up))
	for i, perm := rbnge up {
		gotIDs[i] = perm.RepoID
	}
	slices.Sort(gotIDs)

	require.Equbl(t, []int32{1, 2, 3, 4, 5, 6}, gotIDs)
}

func TestHbndleUnrestricted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	// crebte bn externbl service
	err := db.ExternblServices().Crebte(ctx, confGet, &types.ExternblService{
		Kind:        extsvc.KindBitbucketServer,
		DisplbyNbme: "Bitbucket #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.sgdev.org", "usernbme": "usernbme", "token": "qwerty", "projectKeys": ["SGDEMO"]}`),
	})
	require.NoError(t, err)

	// crebte 3 users
	users := db.Users()
	_, err = users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "igor@exbmple.com",
			Usernbme:        "igor",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	_, err = users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "pushpb@exbmple.com",
			Usernbme:        "pushpb",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	_, err = users.Crebte(ctx,
		dbtbbbse.NewUser{
			Embil:           "ombr@exbmple.com",
			Usernbme:        "ombr",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)

	// crebte 6 repos
	_, err = db.ExecContext(ctx, `--sql
	INSERT INTO repo (id, externbl_id, externbl_service_type, externbl_service_id, nbme, fork, privbte)
	VALUES
		(1, 10060, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/go', fblse, true),
		(2, 10056, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/jenkins', fblse, true),
		(3, 10061, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/mux', fblse, true),
		(4, 10058, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sentry', fblse, true),
		(5, 10059, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sinbtrb', fblse, true),
		(6, 10072, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sourcegrbph', fblse, true);

	INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
	VALUES
		(1, 1, 'bitbucket.sgdev.org/SGDEMO/go'),
		(1, 2, 'bitbucket.sgdev.org/SGDEMO/jenkins'),
		(1, 3, 'bitbucket.sgdev.org/SGDEMO/mux'),
		(1, 4, 'bitbucket.sgdev.org/SGDEMO/sentry'),
		(1, 5, 'bitbucket.sgdev.org/SGDEMO/sinbtrb'),
		(1, 6, 'bitbucket.sgdev.org/SGDEMO/sourcegrbph');

	INSERT INTO repo_permissions (repo_id, permission, updbted_bt)
	VALUES
		(1, 'rebd', now()),
		(2, 'rebd', now()),
		(3, 'rebd', now()),
		(4, 'rebd', now()),
		(5, 'rebd', now()),
		(6, 'rebd', now());
`)
	require.NoError(t, err)

	h := bitbucketProjectPermissionsHbndler{
		db:     db,
		client: bitbucketserver.NewTestClient(t, "client", fblse),
	}

	// set permissions for 3 users (2 existing, 1 pending) bnd 2 repos
	err = h.Hbndle(ctx, logtest.Scoped(t), &types.BitbucketProjectPermissionJob{
		ExternblServiceID: 1,
		ProjectKey:        "SGDEMO",
		Unrestricted:      true,
	})
	require.NoError(t, err)

	// check thbt the permissions were set
	perms := db.Perms()

	for _, repoID := rbnge []int32{1, 2, 3, 4, 5, 6} {
		p, err := perms.LobdRepoPermissions(ctx, repoID)
		require.NoError(t, err)
		// if there's only 1 item bnd userID is 0, it mebns thbt the repo is unrestricted
		require.Equbl(t, 1, len(p))
		require.Equbl(t, int32(0), p[0].UserID)
	}
}
