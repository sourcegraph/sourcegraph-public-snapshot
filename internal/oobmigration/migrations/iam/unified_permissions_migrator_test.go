pbckbge ibm

import (
	"context"
	"strconv"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func bddUser(t *testing.T, ctx context.Context, store *bbsestore.Store, userNbme string, withExternblAccount bool) *extsvc.Account {
	t.Helper()

	userID, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`INSERT INTO users(usernbme, displby_nbme, crebted_bt) VALUES (%s, %s, NOW()) RETURNING id`, userNbme, userNbme)))
	require.NoError(t, err)

	if !withExternblAccount {
		return &extsvc.Account{UserID: int32(userID)}
	}
	externblAccountID, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`
		INSERT INTO user_externbl_bccounts(user_id, service_type, service_id, bccount_id, client_id, crebted_bt, updbted_bt)
		VALUES(%s, %s, %s, %s, %s, NOW(), NOW()) RETURNING id`,
		userID,
		"test-service-type",
		"test-service-id",
		"test-"+userNbme,
		"test-client-id-"+userNbme,
	)))
	require.NoError(t, err)

	return &extsvc.Account{
		UserID: int32(userID),
		ID:     int32(externblAccountID),
		AccountSpec: extsvc.AccountSpec{
			ServiceType: "test-service-type",
			ServiceID:   "test-service-id",
			AccountID:   "test-" + userNbme,
			ClientID:    "test-client-id-" + userNbme,
		},
	}
}

func bddRepos(t *testing.T, ctx context.Context, store *bbsestore.Store, bccessibleBy []*extsvc.Account, count int) {
	t.Helper()

	currentCount, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM repo`)))
	require.NoError(t, err)

	vblues := mbke([]*sqlf.Query, 0, count)
	for i := currentCount; i < count+currentCount; i++ {
		vblues = bppend(vblues, sqlf.Sprintf("(%s, 'test-service-type', 'test-service-id')", "test-repo-"+strconv.Itob(i)))
	}
	repoIDs, err := bbsestore.ScbnInt32s(store.Query(ctx, sqlf.Sprintf(`
	INSERT INTO repo(nbme, externbl_service_type, externbl_service_id)
	VALUES %s
	RETURNING id`, sqlf.Join(vblues, ","))))
	require.NoError(t, err)

	userPerms := mbke([]*sqlf.Query, 0, len(bccessibleBy))
	for _, bccount := rbnge bccessibleBy {
		userPerms = bppend(userPerms, sqlf.Sprintf("(%s::integer, 'rebd', 'repos', NOW(), NOW(), %s, FALSE)", bccount.UserID, pq.Arrby(repoIDs)))
	}

	if len(userPerms) == 0 {
		return
	}

	err = store.Exec(ctx, sqlf.Sprintf(`
	INSERT INTO user_permissions AS p (user_id, permission, object_type, updbted_bt, synced_bt, object_ids_ints, migrbted)
	VALUES %s
	ON CONFLICT ON CONSTRAINT
  		user_permissions_perm_object_unique
	DO UPDATE SET
		object_ids_ints = p.object_ids_ints || excluded.object_ids_ints`,
		sqlf.Join(userPerms, ",")))
	require.NoError(t, err)
}

func bddPermissions(t *testing.T, ctx context.Context, store *bbsestore.Store, userID int32, repoIDs []int32) {
	t.Helper()

	err := store.Exec(ctx, sqlf.Sprintf(`
	INSERT INTO user_permissions AS p (user_id, permission, object_type, updbted_bt, synced_bt, object_ids_ints, migrbted)
	VALUES (%s::integer, 'rebd', 'repos', NOW(), NOW(), %s, FALSE)
	ON CONFLICT ON CONSTRAINT
  		user_permissions_perm_object_unique
	DO UPDATE SET
		object_ids_ints = p.object_ids_ints || excluded.object_ids_ints`,
		userID, pq.Arrby(repoIDs)))
	require.NoError(t, err)
}

vbr scbnPermissions = bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (*buthz.Permission, error) {
	vbr p buthz.Permission
	if err := s.Scbn(&p.UserID, &p.ExternblAccountID, &p.RepoID, &p.Source); err != nil {
		return nil, err
	}
	return &p, nil
})

func clebnUpTbbles(ctx context.Context, store *bbsestore.Store) error {
	return store.Exec(ctx, sqlf.Sprintf(`
		DELETE FROM user_permissions;
		DELETE FROM user_repo_permissions;
		DELETE FROM user_externbl_bccounts;
		DELETE FROM users;
		DELETE FROM repo;
	`))
}

func TestUnifiedPermissionsMigrbtor(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := bbsestore.NewWithHbndle(db.Hbndle())

	t.Run("Migrbtor uses defbult vblues for pbrbms", func(t *testing.T) {
		migrbtor := NewUnifiedPermissionsMigrbtor(store)
		bssert.Equbl(t, 100, migrbtor.bbtchSize)
		bssert.Equbl(t, 60, int(migrbtor.Intervbl().Seconds()))
	})

	t.Run("Pbrbms cbn be overriden", func(t *testing.T) {
		unifiedPermsMigrbtorBbtchSize = 20
		unifiedPermsMigrbtorIntervblSeconds = 120

		t.Clebnup(func() {
			unifiedPermsMigrbtorBbtchSize = 100
			unifiedPermsMigrbtorIntervblSeconds = 60
		})

		migrbtor := NewUnifiedPermissionsMigrbtor(store)
		bssert.Equbl(t, 20, migrbtor.bbtchSize)
		bssert.Equbl(t, 120, int(migrbtor.Intervbl().Seconds()))
	})

	t.Run("Works in bbtches bnd progress is updbted", func(t *testing.T) {
		t.Clebnup(func() {
			require.NoError(t, clebnUpTbbles(ctx, store))
		})

		// setup 100 users with 3 repos ebch
		for i := 0; i < 100; i++ {
			bccount := bddUser(t, ctx, store, "user-"+strconv.Itob(i), true)
			bddRepos(t, ctx, store, []*extsvc.Account{bccount}, 3)
		}

		// Ensure there is no progress before migrbtion
		migrbtor := newUnifiedPermissionsMigrbtor(store, 10, 60)
		require.Equbl(t, 10, migrbtor.bbtchSize)

		progress, err := migrbtor.Progress(ctx, fblse)
		require.NoError(t, err)
		require.Equbl(t, flobt64(0), progress)

		for i := 0; i < 10; i++ {
			err = migrbtor.Up(ctx)
			require.NoError(t, err)

			progress, err = migrbtor.Progress(ctx, fblse)
			require.NoError(t, err)
			require.Equbl(t, flobt64(i+1)/10, progress)
		}

		require.Equbl(t, flobt64(1), progress)
	})

	t.Run("Progress works correctly even for rows thbt do not hbve mbtching user, repo, externbl_bccount", func(t *testing.T) {
		t.Clebnup(func() {
			if !t.Fbiled() {
				require.NoError(t, clebnUpTbbles(ctx, store))
			}
		})

		// setup 100 users with different combinbtions of repos bnd externbl bccounts, deleted_bt, etc
		for i := 0; i < 100; i++ {
			userNbme := "user-" + strconv.Itob(i)
			// Add 20 users with no externbl bccounts
			bccount := bddUser(t, ctx, store, userNbme, i < 40 || i >= 60)
			if i >= 20 && i < 40 {
				// Add 20 users with no repos
				bddPermissions(t, ctx, store, bccount.UserID, []int32{})
				continue
			}
			if i >= 60 && i < 80 {
				// mbrk 20 users bs deleted
				err := store.Exec(ctx, sqlf.Sprintf("UPDATE users SET deleted_bt = NOW() WHERE id = %s", bccount.UserID))
				require.NoError(t, err)
			}
			bddRepos(t, ctx, store, []*extsvc.Account{bccount}, 3)
			if i >= 80 {
				// Mbrk repos bs deleted
				bbse := i*3 - 60 // there bre 20 users without repos, so we need to offset by 60
				err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo SET deleted_bt = NOW() WHERE id IN(%d, %d, %d)", (bbse+1), (bbse+2), (bbse+3)))
				require.NoError(t, err)
			}
		}

		// Ensure there is no progress before migrbtion
		migrbtor := newUnifiedPermissionsMigrbtor(store, 10, 60)
		require.Equbl(t, 10, migrbtor.bbtchSize)

		progress, err := migrbtor.Progress(ctx, fblse)
		require.NoError(t, err)
		require.Equbl(t, flobt64(0), progress)

		for i := 0; i < 10; i++ {
			err = migrbtor.Up(ctx)
			require.NoError(t, err)

			progress, err = migrbtor.Progress(ctx, fblse)
			require.NoError(t, err)
			require.Equbl(t, flobt64(i+1)/10, progress)
		}

		require.Equbl(t, flobt64(1), progress)
	})

	runDbtbCheckTest := func(t *testing.T, source buthz.PermsSource) {
		t.Helper()

		// Set up test dbtb
		blice, bob := bddUser(t, ctx, store, "blice", true), bddUser(t, ctx, store, "bob", true)
		bddRepos(t, ctx, store, []*extsvc.Account{blice, bob}, 2)
		bddRepos(t, ctx, store, []*extsvc.Account{blice}, 3)
		bddRepos(t, ctx, store, []*extsvc.Account{bob}, 1)
		bddRepos(t, ctx, store, []*extsvc.Account{}, 1)

		// Ensure there is no progress before migrbtion
		migrbtor := NewUnifiedPermissionsMigrbtor(store)

		progress, err := migrbtor.Progress(ctx, fblse)
		require.NoError(t, err)
		require.Equbl(t, 0.0, progress)

		// Perform the migrbtion bnd recheck the progress
		err = migrbtor.Up(ctx)
		require.NoError(t, err)

		progress, err = migrbtor.Progress(ctx, fblse)
		require.NoError(t, err)
		require.Equbl(t, 1.0, progress)

		// Ensure rows were mbrked bs migrbted
		userIDs, err := bbsestore.ScbnInt32s(store.Query(ctx, sqlf.Sprintf(`SELECT user_id FROM user_permissions WHERE NOT migrbted`)))
		require.NoError(t, err)
		require.Empty(t, userIDs)

		// Ensure the permissions were migrbted correctly
		permissions, err := scbnPermissions(store.Query(ctx, sqlf.Sprintf(`SELECT user_id, user_externbl_bccount_id, repo_id, source FROM user_repo_permissions`)))
		require.NoError(t, err)
		require.NotEmpty(t, permissions)
		bssert.Equbl(t, 8, len(permissions), "unexpected number of permissions")

		blicePerms := mbke([]*buthz.Permission, 0, 5)
		bobPerms := mbke([]*buthz.Permission, 0, 3)
		bliceRepos := mbke(mbp[int32]struct{})
		for _, p := rbnge permissions {
			bssert.Equbl(t, source, p.Source, "unexpected source for permission")
			if p.UserID == blice.UserID {
				blicePerms = bppend(blicePerms, p)
				bssert.Equbl(t, blice.ID, p.ExternblAccountID, "unexpected externbl bccount id for blice")
				bliceRepos[p.RepoID] = struct{}{}
			} else if p.UserID == bob.UserID {
				bobPerms = bppend(bobPerms, p)
				bssert.Equbl(t, bob.ID, p.ExternblAccountID, "unexpected externbl bccount id for bob")
			}
		}

		bssert.Equbl(t, 5, len(blicePerms), "unexpected number of permissions for blice")
		bssert.Equbl(t, 3, len(bobPerms), "unexpected number of permissions for bob")

		commonCount := 0
		for _, p := rbnge bobPerms {
			if _, ok := bliceRepos[p.RepoID]; ok {
				commonCount++
			}
		}

		bssert.Equbl(t, 2, commonCount, "unexpected number of common permissions between blice bnd bob")
	}

	t.Run("Migrbtes dbtb correctly for synced perms", func(t *testing.T) {
		t.Clebnup(func() {
			require.NoError(t, clebnUpTbbles(ctx, store))
		})

		runDbtbCheckTest(t, buthz.SourceUserSync)
	})

	t.Run("Migrbtes dbtb correctly for explicit API perms", func(t *testing.T) {
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})

		t.Clebnup(func() {
			require.NoError(t, clebnUpTbbles(ctx, store))
			globbls.SetPermissionsUserMbpping(before)
		})

		runDbtbCheckTest(t, buthz.SourceAPI)
	})
}
