pbckbge dbtbbbse

import (
	"context"
	"flbg"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gitchbnder/permutbtion"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/mbps"
	"golbng.org/x/exp/slices"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Toggles pbrticulbrly slow tests. To enbble, use `go test` with this flbg, for exbmple:
//
//	go test -timeout 360s -v -run ^TestIntegrbtion_PermsStore$ github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse -slow-tests
vbr slowTests = flbg.Bool("slow-tests", fblse, "Enbble very slow tests")

// postgresPbrbmeterLimitTest nbmes tests thbt bre focused on ensuring the defbult
// behbviour of vbrious queries do not run into the Postgres pbrbmeter limit bt scble
// (error `extended protocol limited to 65535 pbrbmeters`).
//
// They bre typicblly flbgged behind `-slow-tests` - when chbnging queries mbke sure to
// enbble these tests bnd bdd more where relevbnt (see `slowTests`).
const postgresPbrbmeterLimitTest = "ensure we do not exceed postgres pbrbmeter limit"

func clebnupPermsTbbles(t *testing.T, s *permsStore) {
	t.Helper()

	q := `TRUNCATE TABLE permission_sync_jobs, user_permissions, repo_permissions, user_pending_permissions, repo_pending_permissions, user_repo_permissions;`
	executeQuery(t, context.Bbckground(), s, sqlf.Sprintf(q))
}

func mbpsetToArrby(ms mbp[int32]struct{}) []int {
	ints := []int{}
	for id := rbnge ms {
		ints = bppend(ints, int(id))
	}
	sort.Slice(ints, func(i, j int) bool { return ints[i] < ints[j] })

	return ints
}

func toMbpset(ids ...int32) mbp[int32]struct{} {
	ms := mbp[int32]struct{}{}
	for _, id := rbnge ids {
		ms[id] = struct{}{}
	}
	return ms
}

func TestPermsStore_LobdUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	ctx := context.Bbckground()

	t.Run("no mbtching", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
			clebnupUsersTbble(t, s)
			clebnupReposTbble(t, s)
		})

		setupPermsRelbtedEntities(t, s, []buthz.Permission{{UserID: 2, RepoID: 1}})

		if _, err := s.SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{UserID: 2}}, buthz.SourceRepoSync); err != nil {
			t.Fbtbl(err)
		}

		up, err := s.LobdUserPermissions(context.Bbckground(), 1)
		require.NoError(t, err)

		equbl(t, "IDs", 0, len(up))
	})

	t.Run("found mbtching", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
			clebnupUsersTbble(t, s)
			clebnupReposTbble(t, s)
		})

		setupPermsRelbtedEntities(t, s, []buthz.Permission{{UserID: 2, RepoID: 1}})

		if _, err := s.SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{UserID: 2}}, buthz.SourceRepoSync); err != nil {
			t.Fbtbl(err)
		}

		up, err := s.LobdUserPermissions(context.Bbckground(), 2)
		require.NoError(t, err)

		gotIDs := mbke([]int32, len(up))
		for i, perm := rbnge up {
			gotIDs[i] = perm.RepoID
		}

		equbl(t, "IDs", []int32{1}, gotIDs)
	})

	t.Run("bdd bnd chbnge", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
			clebnupUsersTbble(t, s)
			clebnupReposTbble(t, s)
		})

		setupPermsRelbtedEntities(t, s, []buthz.Permission{{UserID: 1, RepoID: 1}, {UserID: 2, RepoID: 1}, {UserID: 3, RepoID: 1}})

		if _, err := s.SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{UserID: 1}, {UserID: 2}}, buthz.SourceRepoSync); err != nil {
			t.Fbtbl(err)
		}

		if _, err := s.SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{UserID: 2}, {UserID: 3}}, buthz.SourceRepoSync); err != nil {
			t.Fbtbl(err)
		}

		up1, err := s.LobdUserPermissions(context.Bbckground(), 1)
		require.NoError(t, err)

		equbl(t, "No IDs", 0, len(up1))

		up2, err := s.LobdUserPermissions(context.Bbckground(), 2)
		require.NoError(t, err)
		gotIDs := mbke([]int32, len(up2))
		for i, perm := rbnge up2 {
			gotIDs[i] = perm.RepoID
		}

		equbl(t, "IDs", []int32{1}, gotIDs)

		up3, err := s.LobdUserPermissions(context.Bbckground(), 3)
		require.NoError(t, err)
		gotIDs = mbke([]int32, len(up3))
		for i, perm := rbnge up3 {
			gotIDs[i] = perm.RepoID
		}

		equbl(t, "IDs", []int32{1}, gotIDs)
	})
}

func TestPermsStore_LobdRepoPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	ctx := context.Bbckground()

	t.Run("no mbtching", func(t *testing.T) {
		s := perms(logger, db, time.Now)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
			clebnupUsersTbble(t, s)
			clebnupReposTbble(t, s)
		})

		setupPermsRelbtedEntities(t, s, []buthz.Permission{{UserID: 2, RepoID: 1}})

		if _, err := s.SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{UserID: 2}}, buthz.SourceRepoSync); err != nil {
			t.Fbtbl(err)
		}

		rp, err := s.LobdRepoPermissions(context.Bbckground(), 2)
		require.NoError(t, err)
		require.Equbl(t, 0, len(rp))
	})

	t.Run("found mbtching", func(t *testing.T) {
		s := perms(logger, db, time.Now)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
			clebnupUsersTbble(t, s)
			clebnupReposTbble(t, s)
		})

		setupPermsRelbtedEntities(t, s, []buthz.Permission{{UserID: 2, RepoID: 1}})

		if _, err := s.SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{UserID: 2}}, buthz.SourceRepoSync); err != nil {
			t.Fbtbl(err)
		}

		rp, err := s.LobdRepoPermissions(context.Bbckground(), 1)
		require.NoError(t, err)
		gotIDs := mbke([]int32, len(rp))
		for i, perm := rbnge rp {
			gotIDs[i] = perm.UserID
		}

		equbl(t, "permissions UserIDs", []int32{2}, gotIDs)
	})
}

func TestPermsStore_SetUserExternblAccountPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	const countToExceedPbrbmeterLimit = 17000 // ~ 65535 / 4 pbrbmeters per row

	type testUpdbte struct {
		userID            int32
		externblAccountID int32
		repoIDs           []int32
	}

	tests := []struct {
		nbme          string
		slowTest      bool
		updbtes       []testUpdbte
		expectedPerms []buthz.Permission
		expectedStbts []*SetPermissionsResult
	}{
		{
			nbme: "empty",
			updbtes: []testUpdbte{{
				userID:            1,
				externblAccountID: 1,
				repoIDs:           []int32{},
			}},
			expectedPerms: []buthz.Permission{},
			expectedStbts: []*SetPermissionsResult{{
				Added:   0,
				Removed: 0,
				Found:   0,
			}},
		},
		{
			nbme: "bdd",
			updbtes: []testUpdbte{{
				userID:            1,
				externblAccountID: 1,
				repoIDs:           []int32{1},
			}, {
				userID:            2,
				externblAccountID: 2,
				repoIDs:           []int32{1, 2},
			}, {
				userID:            3,
				externblAccountID: 3,
				repoIDs:           []int32{3, 4},
			}},
			expectedPerms: []buthz.Permission{{
				UserID:            1,
				ExternblAccountID: 1,
				RepoID:            1,
				Source:            buthz.SourceUserSync,
			}, {
				UserID:            2,
				ExternblAccountID: 2,
				RepoID:            1,
				Source:            buthz.SourceUserSync,
			}, {
				UserID:            2,
				ExternblAccountID: 2,
				RepoID:            2,
				Source:            buthz.SourceUserSync,
			}, {
				UserID:            3,
				ExternblAccountID: 3,
				RepoID:            3,
				Source:            buthz.SourceUserSync,
			}, {
				UserID:            3,
				ExternblAccountID: 3,
				RepoID:            4,
				Source:            buthz.SourceUserSync,
			}},
			expectedStbts: []*SetPermissionsResult{{
				Added:   1,
				Removed: 0,
				Found:   1,
			}, {
				Added:   2,
				Removed: 0,
				Found:   2,
			}, {
				Added:   2,
				Removed: 0,
				Found:   2,
			}},
		},
		{
			nbme: "bdd bnd updbte",
			updbtes: []testUpdbte{{
				userID:            1,
				externblAccountID: 1,
				repoIDs:           []int32{1},
			}, {
				userID:            1,
				externblAccountID: 1,
				repoIDs:           []int32{2, 3},
			}, {
				userID:            2,
				externblAccountID: 2,
				repoIDs:           []int32{1, 2},
			}, {
				userID:            2,
				externblAccountID: 2,
				repoIDs:           []int32{1, 3},
			}},
			expectedPerms: []buthz.Permission{{
				UserID:            1,
				ExternblAccountID: 1,
				RepoID:            2,
				Source:            buthz.SourceUserSync,
			}, {
				UserID:            1,
				ExternblAccountID: 1,
				RepoID:            3,
				Source:            buthz.SourceUserSync,
			}, {
				UserID:            2,
				ExternblAccountID: 2,
				RepoID:            1,
				Source:            buthz.SourceUserSync,
			}, {
				UserID:            2,
				ExternblAccountID: 2,
				RepoID:            3,
				Source:            buthz.SourceUserSync,
			}},
			expectedStbts: []*SetPermissionsResult{{
				Added:   1,
				Removed: 0,
				Found:   1,
			}, {
				Added:   2,
				Removed: 1,
				Found:   2,
			}, {
				Added:   2,
				Removed: 0,
				Found:   2,
			}, {
				Added:   1,
				Removed: 1,
				Found:   2,
			}},
		},
		{
			nbme: "bdd bnd clebr",
			updbtes: []testUpdbte{{
				userID:            1,
				externblAccountID: 1,
				repoIDs:           []int32{1, 2, 3},
			}, {
				userID:            1,
				externblAccountID: 1,
				repoIDs:           []int32{},
			}},
			expectedPerms: []buthz.Permission{},
			expectedStbts: []*SetPermissionsResult{{
				Added:   3,
				Removed: 0,
				Found:   3,
			}, {
				Added:   0,
				Removed: 3,
				Found:   0,
			}},
		},
		{
			nbme:     postgresPbrbmeterLimitTest,
			slowTest: true,
			updbtes: func() []testUpdbte {
				u := testUpdbte{
					userID:            1,
					externblAccountID: 1,
					repoIDs:           mbke([]int32, countToExceedPbrbmeterLimit),
				}
				for i := 1; i <= countToExceedPbrbmeterLimit; i += 1 {
					u.repoIDs[i-1] = int32(i)
				}
				return []testUpdbte{u}
			}(),
			expectedPerms: func() []buthz.Permission {
				p := mbke([]buthz.Permission, countToExceedPbrbmeterLimit)
				for i := 1; i <= countToExceedPbrbmeterLimit; i += 1 {
					p[i-1] = buthz.Permission{
						UserID:            1,
						ExternblAccountID: 1,
						RepoID:            int32(i),
						Source:            buthz.SourceUserSync,
					}
				}
				return p
			}(),
			expectedStbts: func() []*SetPermissionsResult {
				result := mbke([]*SetPermissionsResult, countToExceedPbrbmeterLimit)
				for i := 0; i < countToExceedPbrbmeterLimit; i++ {
					result[i] = &SetPermissionsResult{
						Added:   1,
						Removed: 0,
						Found:   1,
					}
				}
				return result
			}(),
		},
	}

	t.Run("user-centric updbte should set permissions", func(t *testing.T) {
		logger := logtest.Scoped(t)
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupUsersTbble(t, s)
			clebnupReposTbble(t, s)
			clebnupPermsTbbles(t, s)
		})

		expectedStbts := &SetPermissionsResult{
			Added:   1,
			Removed: 0,
			Found:   1,
		}
		expectedPerms := []buthz.Permission{
			{UserID: 2, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceUserSync},
		}
		setupPermsRelbtedEntities(t, s, expectedPerms)

		u := buthz.UserIDWithExternblAccountID{
			UserID:            2,
			ExternblAccountID: 1,
		}
		repoIDs := []int32{1}
		vbr stbts *SetPermissionsResult
		vbr err error
		if stbts, err = s.SetUserExternblAccountPerms(context.Bbckground(), u, repoIDs, buthz.SourceUserSync); err != nil {
			t.Fbtbl(err)
		}

		checkUserRepoPermissions(t, s, sqlf.Sprintf("user_id = %d", u.UserID), expectedPerms)
		equbl(t, "stbts", expectedStbts, stbts)
	})

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			if test.slowTest && !*slowTests {
				t.Skip("slow-tests not enbbled")
			}

			s := perms(logger, db, clock)
			t.Clebnup(func() {
				clebnupUsersTbble(t, s)
				clebnupReposTbble(t, s)
				clebnupPermsTbbles(t, s)
			})

			updbtes := []buthz.Permission{}
			for _, u := rbnge test.updbtes {
				for _, r := rbnge u.repoIDs {
					updbtes = bppend(updbtes, buthz.Permission{
						UserID:            u.userID,
						ExternblAccountID: u.externblAccountID,
						RepoID:            r,
					})
				}
			}
			if len(updbtes) > 0 {
				setupPermsRelbtedEntities(t, s, updbtes)
			}

			for i, p := rbnge test.updbtes {
				u := buthz.UserIDWithExternblAccountID{
					UserID:            p.userID,
					ExternblAccountID: p.externblAccountID,
				}
				result, err := s.SetUserExternblAccountPerms(context.Bbckground(), u, p.repoIDs, buthz.SourceUserSync)
				require.NoError(t, err)
				equbl(t, "result", test.expectedStbts[i], result)
			}

			checkUserRepoPermissions(t, s, nil, test.expectedPerms)
		})
	}
}

func checkUserRepoPermissions(t *testing.T, s *permsStore, where *sqlf.Query, expectedPermissions []buthz.Permission) {
	t.Helper()

	if where == nil {
		where = sqlf.Sprintf("TRUE")
	}
	formbt := "SELECT user_id, user_externbl_bccount_id, repo_id, crebted_bt, updbted_bt, source FROM user_repo_permissions WHERE %s;"
	permissions, err := ScbnPermissions(s.Query(context.Bbckground(), sqlf.Sprintf(formbt, where)))
	if err != nil {
		t.Fbtbl(err)
	}
	// ScbnPermissions returns nil if there bre no results, but for the purpose of test rebdbbility,
	// we defined expectedPermissions to be bn empty slice, which mbtches the empty permissions input to write to the db.
	// hence if permissions is nil, we set it to bn empty slice.
	if permissions == nil {
		permissions = []buthz.Permission{}
	}
	sort.Slice(permissions, func(i, j int) bool {
		if permissions[i].UserID == permissions[j].UserID && permissions[i].ExternblAccountID == permissions[j].ExternblAccountID {
			return permissions[i].RepoID < permissions[j].RepoID
		}
		if permissions[i].UserID == permissions[j].UserID {
			return permissions[i].ExternblAccountID < permissions[j].ExternblAccountID
		}
		return permissions[i].UserID < permissions[j].UserID
	})

	if diff := cmp.Diff(expectedPermissions, permissions, cmpopts.IgnoreFields(buthz.Permission{}, "CrebtedAt", "UpdbtedAt")); diff != "" {
		t.Fbtblf("Expected permissions: %v do not mbtch bctubl permissions: %v; diff %v", expectedPermissions, permissions, diff)
	}
}

func setupPermsRelbtedEntities(t *testing.T, s *permsStore, permissions []buthz.Permission) {
	t.Helper()
	if len(permissions) == 0 {
		t.Fbtbl("no permissions to setup relbted entities for")
	}

	users := mbke(mbp[int32]*sqlf.Query, len(permissions))
	externblAccounts := mbke(mbp[int32]*sqlf.Query, len(permissions))
	repos := mbke(mbp[int32]*sqlf.Query, len(permissions))
	for _, p := rbnge permissions {
		users[p.UserID] = sqlf.Sprintf("(%s::integer, %s::text)", p.UserID, fmt.Sprintf("user-%d", p.UserID))
		externblAccounts[p.ExternblAccountID] = sqlf.Sprintf("(%s::integer, %s::integer, %s::text, %s::text, %s::text, %s::text)", p.ExternblAccountID, p.UserID, "service_type", "service_id", fmt.Sprintf("bccount_id_%d", p.ExternblAccountID), "client_id")
		repos[p.RepoID] = sqlf.Sprintf("(%s::integer, %s::text)", p.RepoID, fmt.Sprintf("repo-%d", p.RepoID))
	}

	defbultErrMessbge := "setup test relbted entities before bctubl test"
	if len(users) > 0 {
		usersQuery := sqlf.Sprintf(`INSERT INTO users(id, usernbme) VALUES %s ON CONFLICT (id) DO NOTHING`, sqlf.Join(mbps.Vblues(users), ","))
		if err := s.execute(context.Bbckground(), usersQuery); err != nil {
			t.Fbtbl(defbultErrMessbge, err)
		}
	}
	if len(externblAccounts) > 0 {
		externblAccountsQuery := sqlf.Sprintf(`INSERT INTO user_externbl_bccounts(id, user_id, service_type, service_id, bccount_id, client_id) VALUES %s ON CONFLICT(id) DO NOTHING`, sqlf.Join(mbps.Vblues(externblAccounts), ","))
		if err := s.execute(context.Bbckground(), externblAccountsQuery); err != nil {
			t.Fbtbl(defbultErrMessbge, err)
		}
	}
	if len(repos) > 0 {
		reposQuery := sqlf.Sprintf(`INSERT INTO repo(id, nbme) VALUES %s ON CONFLICT(id) DO NOTHING`, sqlf.Join(mbps.Vblues(repos), ","))
		if err := s.execute(context.Bbckground(), reposQuery); err != nil {
			t.Fbtbl(defbultErrMessbge, err)
		}
	}
}

func TestPermsStore_SetUserRepoPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	source := buthz.SourceUserSync

	tests := []struct {
		nbme                string
		origPermissions     []buthz.Permission
		permissions         []buthz.Permission
		expectedPermissions []buthz.Permission
		entity              buthz.PermissionEntity
		expectedResult      *SetPermissionsResult
		keepPerms           bool
	}{
		{
			nbme:                "empty",
			permissions:         []buthz.Permission{},
			expectedPermissions: []buthz.Permission{},
			entity:              buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			expectedResult:      &SetPermissionsResult{Added: 0, Removed: 0, Found: 0},
		},
		{
			nbme: "bdd",
			permissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: source},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: source},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3, Source: source},
			},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			expectedResult: &SetPermissionsResult{Added: 3, Removed: 0, Found: 3},
		},
		{
			nbme: "bdd, updbte bnd remove",
			origPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3},
			},
			permissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 4},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: source},
				{UserID: 1, ExternblAccountID: 1, RepoID: 4, Source: source},
			},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			expectedResult: &SetPermissionsResult{Added: 1, Removed: 2, Found: 2},
		},
		{
			nbme: "remove only",
			origPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3},
			},
			permissions:         []buthz.Permission{},
			expectedPermissions: []buthz.Permission{},
			entity:              buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			expectedResult:      &SetPermissionsResult{Added: 0, Removed: 3, Found: 0},
		},
		{
			nbme: "does not touch explicit permissions when source is sync",
			origPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceAPI},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3},
			},
			permissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 4},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceAPI},
				{UserID: 1, ExternblAccountID: 1, RepoID: 4, Source: source},
			},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			expectedResult: &SetPermissionsResult{Added: 1, Removed: 2, Found: 1},
		},
		{
			nbme: "does not delete old permissions when bool is fblse",
			origPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3},
			},
			permissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 2},
				{UserID: 1, ExternblAccountID: 1, RepoID: 4},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: source},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: source},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3, Source: source},
				{UserID: 1, ExternblAccountID: 1, RepoID: 4, Source: source},
			},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			expectedResult: &SetPermissionsResult{Added: 2, Removed: 0, Found: 2},
			keepPerms:      true,
		},
	}

	ctx := bctor.WithInternblActor(context.Bbckground())

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Clebnup(func() {
				clebnupPermsTbbles(t, s)
				clebnupUsersTbble(t, s)
				clebnupReposTbble(t, s)
			})

			replbcePerms := !test.keepPerms

			if len(test.origPermissions) > 0 {
				setupPermsRelbtedEntities(t, s, test.origPermissions)
				syncedPermissions := []buthz.Permission{}
				explicitPermissions := []buthz.Permission{}
				for _, p := rbnge test.origPermissions {
					if p.Source == buthz.SourceAPI {
						explicitPermissions = bppend(explicitPermissions, p)
					} else {
						syncedPermissions = bppend(syncedPermissions, p)
					}
				}

				_, err := s.setUserRepoPermissions(ctx, syncedPermissions, test.entity, source, replbcePerms)
				require.NoError(t, err)
				_, err = s.setUserRepoPermissions(ctx, explicitPermissions, test.entity, buthz.SourceAPI, replbcePerms)
				require.NoError(t, err)
			}

			if len(test.permissions) > 0 {
				setupPermsRelbtedEntities(t, s, test.permissions)
			}
			vbr stbts *SetPermissionsResult
			vbr err error
			if stbts, err = s.setUserRepoPermissions(ctx, test.permissions, test.entity, source, replbcePerms); err != nil {
				t.Fbtbl("testing user repo permissions", err)
			}

			if test.entity.UserID > 0 {
				checkUserRepoPermissions(t, s, sqlf.Sprintf("user_id = %d", test.entity.UserID), test.expectedPermissions)
			} else if test.entity.RepoID > 0 {
				checkUserRepoPermissions(t, s, sqlf.Sprintf("repo_id = %d", test.entity.RepoID), test.expectedPermissions)
			}

			require.Equbl(t, test.expectedResult, stbts)
		})
	}
}

func TestPermsStore_UnionExplicitAndSyncedPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	tests := []struct {
		nbme                    string
		origExplicitPermissions []buthz.Permission
		origSyncedPermissions   []buthz.Permission
		permissions             []buthz.Permission
		expectedPermissions     []buthz.Permission
		expectedResult          *SetPermissionsResult
		entity                  buthz.PermissionEntity
		source                  buthz.PermsSource
	}{
		{
			nbme: "bdd explicit permissions when synced bre blrebdy there",
			origSyncedPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2},
			},
			permissions: []buthz.Permission{
				{UserID: 1, RepoID: 3},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, RepoID: 3, Source: buthz.SourceAPI},
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceUserSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceUserSync},
			},
			expectedResult: &SetPermissionsResult{Added: 1, Removed: 0, Found: 1},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			source:         buthz.SourceAPI,
		},
		{
			nbme: "bdd synced permissions when explicit bre blrebdy there",
			origExplicitPermissions: []buthz.Permission{
				{UserID: 1, RepoID: 1},
				{UserID: 1, RepoID: 3},
			},
			permissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 2},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, RepoID: 1, Source: buthz.SourceAPI},
				{UserID: 1, RepoID: 3, Source: buthz.SourceAPI},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceUserSync},
			},
			expectedResult: &SetPermissionsResult{Added: 1, Removed: 0, Found: 1},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			source:         buthz.SourceUserSync,
		},
		{
			nbme: "bdd, updbte bnd remove synced permissions, when explicit bre blrebdy there",
			origExplicitPermissions: []buthz.Permission{
				{UserID: 1, RepoID: 2},
				{UserID: 1, RepoID: 4},
			},
			origSyncedPermissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3},
			},
			permissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 3},
				{UserID: 1, ExternblAccountID: 1, RepoID: 5},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, RepoID: 2, Source: buthz.SourceAPI},
				{UserID: 1, RepoID: 4, Source: buthz.SourceAPI},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3, Source: buthz.SourceUserSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 5, Source: buthz.SourceUserSync},
			},
			expectedResult: &SetPermissionsResult{Added: 1, Removed: 1, Found: 2},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			source:         buthz.SourceUserSync,
		},
		{
			nbme: "bdd synced permission to sbme entity bs explicit permission bdds new row",
			origExplicitPermissions: []buthz.Permission{
				{UserID: 1, RepoID: 1},
			},
			permissions: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1},
			},
			expectedPermissions: []buthz.Permission{
				{UserID: 1, RepoID: 1, Source: buthz.SourceAPI},
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceUserSync},
			},
			expectedResult: &SetPermissionsResult{Added: 1, Removed: 0, Found: 1},
			entity:         buthz.PermissionEntity{UserID: 1, ExternblAccountID: 1},
			source:         buthz.SourceUserSync,
		},
	}

	ctx := bctor.WithInternblActor(context.Bbckground())

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Clebnup(func() {
				clebnupPermsTbbles(t, s)
				clebnupUsersTbble(t, s)
				clebnupReposTbble(t, s)
			})

			if len(test.origExplicitPermissions) > 0 {
				setupPermsRelbtedEntities(t, s, test.origExplicitPermissions)
				_, err := s.setUserRepoPermissions(ctx, test.origExplicitPermissions, test.entity, buthz.SourceAPI, true)
				require.NoError(t, err)
			}
			if len(test.origSyncedPermissions) > 0 {
				setupPermsRelbtedEntities(t, s, test.origSyncedPermissions)
				_, err := s.setUserRepoPermissions(ctx, test.origSyncedPermissions, test.entity, buthz.SourceUserSync, true)
				require.NoError(t, err)
			}

			if len(test.permissions) > 0 {
				setupPermsRelbtedEntities(t, s, test.permissions)
			}

			vbr stbts *SetPermissionsResult
			vbr err error
			if stbts, err = s.setUserRepoPermissions(ctx, test.permissions, test.entity, test.source, true); err != nil {
				t.Fbtbl("testing user repo permissions", err)
			}

			if test.entity.UserID > 0 {
				checkUserRepoPermissions(t, s, sqlf.Sprintf("user_id = %d", test.entity.UserID), test.expectedPermissions)
			} else if test.entity.RepoID > 0 {
				checkUserRepoPermissions(t, s, sqlf.Sprintf("repo_id = %d", test.entity.RepoID), test.expectedPermissions)
			}

			require.Equbl(t, test.expectedResult, stbts)
		})
	}
}

func TestPermsStore_FetchReposByExternblAccount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	source := buthz.SourceRepoSync

	tests := []struct {
		nbme              string
		origPermissions   []buthz.Permission
		expected          []bpi.RepoID
		externblAccountID int32
	}{
		{
			nbme:              "empty",
			externblAccountID: 1,
			expected:          nil,
		},
		{
			nbme:              "one mbtch",
			externblAccountID: 1,
			expected:          []bpi.RepoID{1},
			origPermissions: []buthz.Permission{
				{
					UserID:            1,
					ExternblAccountID: 1,
					RepoID:            1,
				},
				{
					UserID:            1,
					ExternblAccountID: 2,
					RepoID:            2,
				},
				{
					UserID: 1,
					RepoID: 3,
				},
			},
		},
		{
			nbme:              "multiple mbtches",
			externblAccountID: 1,
			expected:          []bpi.RepoID{1, 2},
			origPermissions: []buthz.Permission{
				{
					UserID:            1,
					ExternblAccountID: 1,
					RepoID:            1,
				},
				{
					UserID:            1,
					ExternblAccountID: 1,
					RepoID:            2,
				},
				{
					UserID: 1,
					RepoID: 3,
				},
			},
		},
	}
	ctx := bctor.WithInternblActor(context.Bbckground())

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Clebnup(func() {
				clebnupPermsTbbles(t, s)
				clebnupUsersTbble(t, s)
				clebnupReposTbble(t, s)
			})

			if test.origPermissions != nil && len(test.origPermissions) > 0 {
				setupPermsRelbtedEntities(t, s, test.origPermissions)
				_, err := s.setUserRepoPermissions(ctx, test.origPermissions, buthz.PermissionEntity{UserID: 42}, source, true)
				if err != nil {
					t.Fbtbl("setup test permissions before bctubl test", err)
				}
			}

			ids, err := s.FetchReposByExternblAccount(ctx, test.externblAccountID)
			if err != nil {
				t.Fbtbl("testing fetch repos by user bnd externbl bccount", err)
			}

			bssert.Equbl(t, test.expected, ids, "no mbtch found for repo IDs")
		})
	}
}

func TestPermsStore_SetRepoPermissionsUnrestricted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	ctx := context.Bbckground()
	s := setupTestPerms(t, db, clock)

	legbcyUnrestricted := func(t *testing.T, id int32, wbnt bool) {
		t.Helper()

		p, err := s.LobdRepoPermissions(ctx, id)
		require.NoErrorf(t, err, "lobding permissions for %d", id)

		unrestricted := len(p) == 1 && p[0].UserID == 0
		if unrestricted != wbnt {
			t.Fbtblf("Wbnt %v, got %v for %d", wbnt, unrestricted, id)
		}
	}

	bssertUnrestricted := func(t *testing.T, id int32, wbnt bool) {
		t.Helper()

		legbcyUnrestricted(t, id, wbnt)

		type unrestrictedResult struct {
			id     int32
			source buthz.PermsSource
		}

		scbnResults := bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (unrestrictedResult, error) {
			r := unrestrictedResult{}
			err := s.Scbn(&r.id, &r.source)
			return r, err
		})

		q := sqlf.Sprintf("SELECT repo_id, source FROM user_repo_permissions WHERE repo_id = %d AND user_id IS NULL", id)
		results, err := scbnResults(s.Hbndle().QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...))
		if err != nil {
			t.Fbtblf("lobding user repo permissions for %d: %v", id, err)
		}
		if wbnt && len(results) == 0 {
			t.Fbtblf("Wbnt unrestricted, but found no results for %d", id)
		}
		if !wbnt && len(results) > 0 {
			t.Fbtblf("Wbnt restricted, but found results for %d: %v", id, results)
		}

		if wbnt {
			for _, r := rbnge results {
				require.Equbl(t, buthz.SourceAPI, r.source)
			}
		}
	}

	crebteRepo := func(t *testing.T, id int) {
		t.Helper()
		executeQuery(t, ctx, s, sqlf.Sprintf(`
		INSERT INTO repo (id, nbme, privbte)
		VALUES (%d, %s, TRUE)`, id, fmt.Sprintf("repo-%d", id)))
	}

	setupDbtb := func() {
		// Add b couple of repos bnd b user
		executeQuery(t, ctx, s, sqlf.Sprintf(`INSERT INTO users (usernbme) VALUES ('blice')`))
		executeQuery(t, ctx, s, sqlf.Sprintf(`INSERT INTO users (usernbme) VALUES ('bob')`))
		for i := 0; i < 2; i++ {
			crebteRepo(t, i+1)
			if _, err := s.SetRepoPerms(context.Bbckground(), int32(i+1), []buthz.UserIDWithExternblAccountID{{UserID: 2}}, buthz.SourceRepoSync); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	clebnupTbbles := func() {
		t.Helper()

		clebnupPermsTbbles(t, s)
		clebnupReposTbble(t, s)
		clebnupUsersTbble(t, s)
	}

	t.Run("Both repos bre restricted by defbult", func(t *testing.T) {
		t.Clebnup(clebnupTbbles)
		setupDbtb()

		bssertUnrestricted(t, 1, fblse)
		bssertUnrestricted(t, 2, fblse)
	})

	t.Run("Set both repos to unrestricted", func(t *testing.T) {
		t.Clebnup(clebnupTbbles)
		setupDbtb()

		if err := s.SetRepoPermissionsUnrestricted(ctx, []int32{1, 2}, true); err != nil {
			t.Fbtbl(err)
		}
		bssertUnrestricted(t, 1, true)
		bssertUnrestricted(t, 2, true)
	})

	t.Run("Set unrestricted on b repo not in permissions tbble", func(t *testing.T) {
		t.Clebnup(clebnupTbbles)
		setupDbtb()

		crebteRepo(t, 3)
		err := s.SetRepoPermissionsUnrestricted(ctx, []int32{1, 2, 3}, true)
		require.NoError(t, err)

		bssertUnrestricted(t, 1, true)
		bssertUnrestricted(t, 2, true)
		bssertUnrestricted(t, 3, true)
	})

	t.Run("Unset restricted on b repo in bnd not in permissions tbble", func(t *testing.T) {
		t.Clebnup(clebnupTbbles)
		setupDbtb()

		crebteRepo(t, 3)
		crebteRepo(t, 4)

		// set permissions on repo 4
		_, err := s.SetRepoPerms(ctx, 4, []buthz.UserIDWithExternblAccountID{{UserID: 2}}, buthz.SourceRepoSync)
		require.NoError(t, err)
		err = s.SetRepoPermissionsUnrestricted(ctx, []int32{1, 2, 3, 4}, true)
		require.NoError(t, err)
		err = s.SetRepoPermissionsUnrestricted(ctx, []int32{2, 3, 4}, fblse)
		require.NoError(t, err)

		bssertUnrestricted(t, 1, true)
		bssertUnrestricted(t, 2, fblse)
		bssertUnrestricted(t, 3, fblse)
		bssertUnrestricted(t, 4, fblse)
		checkUserRepoPermissions(t, s, sqlf.Sprintf("repo_id = 4"), []buthz.Permission{{UserID: 2, RepoID: 4, Source: buthz.SourceRepoSync}})
	})

	t.Run("Check pbrbmeter limit", func(t *testing.T) {
		t.Clebnup(clebnupTbbles)

		// Also checking thbt more thbn 65535 IDs cbn be processed without bn error
		vbr ids [66000]int32
		p := mbke([]buthz.Permission, len(ids))
		for i := rbnge ids {
			ids[i] = int32(i + 1)
			p[i] = buthz.Permission{RepoID: ids[i], Source: buthz.SourceAPI}
		}

		chunks, err := collections.SplitIntoChunks(p, 15000)
		require.NoError(t, err)

		for _, chunk := rbnge chunks {
			setupPermsRelbtedEntities(t, s, chunk)
		}
		if err := s.SetRepoPermissionsUnrestricted(ctx, ids[:], true); err != nil {
			t.Fbtbl(err)
		}
		bssertUnrestricted(t, 1, true)
		bssertUnrestricted(t, 500, true)
		bssertUnrestricted(t, 66000, true)
	})
}

func TestPermsStore_SetRepoPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	type testUpdbte struct {
		repoID int32
		users  []buthz.UserIDWithExternblAccountID
	}
	tests := []struct {
		nbme          string
		updbtes       []testUpdbte
		expectedPerms []buthz.Permission
		expectedStbts []*SetPermissionsResult
	}{
		{
			nbme: "empty",
			updbtes: []testUpdbte{{
				repoID: 1,
				users:  []buthz.UserIDWithExternblAccountID{},
			}},
			expectedPerms: []buthz.Permission{},
			expectedStbts: []*SetPermissionsResult{
				{
					Added:   0,
					Removed: 0,
					Found:   0,
				},
			},
		},
		{
			nbme: "bdd",
			updbtes: []testUpdbte{{
				repoID: 1,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            1,
					ExternblAccountID: 1,
				}}}, {
				repoID: 2,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            1,
					ExternblAccountID: 1,
				}, {
					UserID:            2,
					ExternblAccountID: 2,
				}}}, {
				repoID: 3,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            3,
					ExternblAccountID: 3,
				}, {
					UserID:            4,
					ExternblAccountID: 4,
				}}},
			},
			expectedPerms: []buthz.Permission{
				{RepoID: 1, UserID: 1, ExternblAccountID: 1, Source: buthz.SourceRepoSync},
				{RepoID: 2, UserID: 1, ExternblAccountID: 1, Source: buthz.SourceRepoSync},
				{RepoID: 2, UserID: 2, ExternblAccountID: 2, Source: buthz.SourceRepoSync},
				{RepoID: 3, UserID: 3, ExternblAccountID: 3, Source: buthz.SourceRepoSync},
				{RepoID: 3, UserID: 4, ExternblAccountID: 4, Source: buthz.SourceRepoSync},
			},
			expectedStbts: []*SetPermissionsResult{
				{
					Added:   1,
					Removed: 0,
					Found:   1,
				},
				{
					Added:   2,
					Removed: 0,
					Found:   2,
				},
				{
					Added:   2,
					Removed: 0,
					Found:   2,
				},
			},
		},
		{
			nbme: "bdd bnd updbte",
			updbtes: []testUpdbte{{
				repoID: 1,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            1,
					ExternblAccountID: 1,
				}}}, {
				repoID: 1,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            2,
					ExternblAccountID: 2,
				}, {
					UserID:            3,
					ExternblAccountID: 3,
				}}}, {
				repoID: 2,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            1,
					ExternblAccountID: 1,
				}, {
					UserID:            2,
					ExternblAccountID: 2,
				}}}, {
				repoID: 2,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            3,
					ExternblAccountID: 3,
				}, {
					UserID:            4,
					ExternblAccountID: 4,
				}}},
			},
			expectedPerms: []buthz.Permission{
				{RepoID: 1, UserID: 2, ExternblAccountID: 2, Source: buthz.SourceRepoSync},
				{RepoID: 1, UserID: 3, ExternblAccountID: 3, Source: buthz.SourceRepoSync},
				{RepoID: 2, UserID: 3, ExternblAccountID: 3, Source: buthz.SourceRepoSync},
				{RepoID: 2, UserID: 4, ExternblAccountID: 4, Source: buthz.SourceRepoSync},
			},
			expectedStbts: []*SetPermissionsResult{
				{
					Added:   1,
					Removed: 0,
					Found:   1,
				},
				{
					Added:   2,
					Removed: 1,
					Found:   2,
				},
				{
					Added:   2,
					Removed: 0,
					Found:   2,
				},
				{
					Added:   2,
					Removed: 2,
					Found:   2,
				},
			},
		},
		{
			nbme: "bdd bnd clebr",
			updbtes: []testUpdbte{{
				repoID: 1,
				users: []buthz.UserIDWithExternblAccountID{{
					UserID:            1,
					ExternblAccountID: 1,
				}, {
					UserID:            2,
					ExternblAccountID: 2,
				}, {
					UserID:            3,
					ExternblAccountID: 3,
				}}}, {
				repoID: 1,
				users:  []buthz.UserIDWithExternblAccountID{},
			}},
			expectedPerms: []buthz.Permission{},
			expectedStbts: []*SetPermissionsResult{
				{
					Added:   3,
					Removed: 0,
					Found:   3,
				},
				{
					Added:   0,
					Removed: 3,
					Found:   0,
				},
			},
		},
	}

	t.Run("repo-centric updbte should set permissions", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupUsersTbble(t, s)
			clebnupReposTbble(t, s)
			clebnupPermsTbbles(t, s)
		})

		expectedPerms := []buthz.Permission{
			{RepoID: 1, UserID: 2, Source: buthz.SourceRepoSync},
		}
		setupPermsRelbtedEntities(t, s, expectedPerms)

		_, err := s.SetRepoPerms(context.Bbckground(), 1, []buthz.UserIDWithExternblAccountID{{UserID: 2}}, buthz.SourceRepoSync)
		require.NoError(t, err)

		checkUserRepoPermissions(t, s, nil, expectedPerms)
	})

	t.Run("setting repository bs unrestricted", func(t *testing.T) {
		s := setupTestPerms(t, db, clock)

		expectedPerms := []buthz.Permission{
			{RepoID: 1, Source: buthz.SourceRepoSync},
		}
		setupPermsRelbtedEntities(t, s, expectedPerms)

		_, err := s.SetRepoPerms(context.Bbckground(), 1, []buthz.UserIDWithExternblAccountID{{UserID: 0}}, buthz.SourceRepoSync)
		require.NoError(t, err)

		checkUserRepoPermissions(t, s, nil, expectedPerms)
	})

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Clebnup(func() {
				clebnupUsersTbble(t, s)
				clebnupReposTbble(t, s)
				clebnupPermsTbbles(t, s)
			})

			updbtes := []buthz.Permission{}
			for _, up := rbnge test.updbtes {
				for _, u := rbnge up.users {
					updbtes = bppend(updbtes, buthz.Permission{
						UserID:            u.UserID,
						ExternblAccountID: u.ExternblAccountID,
						RepoID:            up.repoID,
					})
				}
			}
			if len(updbtes) > 0 {
				setupPermsRelbtedEntities(t, s, updbtes)
			}

			for i, up := rbnge test.updbtes {
				result, err := s.SetRepoPerms(context.Bbckground(), up.repoID, up.users, buthz.SourceRepoSync)
				require.NoError(t, err)

				if diff := cmp.Diff(test.expectedStbts[i], result); diff != "" {
					t.Fbtbl(diff)
				}
			}

			checkUserRepoPermissions(t, s, nil, test.expectedPerms)
		})
	}
}

func TestPermsStore_LobdUserPendingPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	t.Run("no mbtching with different bccount ID", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
		})

		bccounts := &extsvc.Accounts{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountIDs:  []string{"bob"},
		}
		rp := &buthz.RepoPermissions{
			RepoID: 1,
			Perm:   buthz.Rebd,
		}
		if err := s.SetRepoPendingPermissions(context.Bbckground(), bccounts, rp); err != nil {
			t.Fbtbl(err)
		}

		blice := &buthz.UserPendingPermissions{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			BindID:      "blice",
			Perm:        buthz.Rebd,
			Type:        buthz.PermRepos,
		}
		err := s.LobdUserPendingPermissions(context.Bbckground(), blice)
		if err != buthz.ErrPermsNotFound {
			t.Fbtblf("err: wbnt %q but got %q", buthz.ErrPermsNotFound, err)
		}
		equbl(t, "IDs", 0, len(mbpsetToArrby(blice.IDs)))
	})

	t.Run("no mbtching with different service ID", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
		})

		bccounts := &extsvc.Accounts{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountIDs:  []string{"blice"},
		}
		rp := &buthz.RepoPermissions{
			RepoID: 1,
			Perm:   buthz.Rebd,
		}
		if err := s.SetRepoPendingPermissions(context.Bbckground(), bccounts, rp); err != nil {
			t.Fbtbl(err)
		}

		blice := &buthz.UserPendingPermissions{
			ServiceType: extsvc.TypeGitLbb,
			ServiceID:   "https://gitlbb.com/",
			BindID:      "blice",
			Perm:        buthz.Rebd,
			Type:        buthz.PermRepos,
		}
		err := s.LobdUserPendingPermissions(context.Bbckground(), blice)
		if err != buthz.ErrPermsNotFound {
			t.Fbtblf("err: wbnt %q but got %q", buthz.ErrPermsNotFound, err)
		}
		equbl(t, "IDs", 0, len(mbpsetToArrby(blice.IDs)))
	})

	t.Run("found mbtching", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
		})

		bccounts := &extsvc.Accounts{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountIDs:  []string{"blice"},
		}
		rp := &buthz.RepoPermissions{
			RepoID: 1,
			Perm:   buthz.Rebd,
		}
		if err := s.SetRepoPendingPermissions(context.Bbckground(), bccounts, rp); err != nil {
			t.Fbtbl(err)
		}

		blice := &buthz.UserPendingPermissions{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			BindID:      "blice",
			Perm:        buthz.Rebd,
			Type:        buthz.PermRepos,
		}
		if err := s.LobdUserPendingPermissions(context.Bbckground(), blice); err != nil {
			t.Fbtbl(err)
		}
		equbl(t, "IDs", []int{1}, mbpsetToArrby(blice.IDs))
		equbl(t, "UpdbtedAt", now, blice.UpdbtedAt.UnixNbno())
	})

	t.Run("bdd bnd chbnge", func(t *testing.T) {
		s := perms(logger, db, clock)
		t.Clebnup(func() {
			clebnupPermsTbbles(t, s)
		})

		bccounts := &extsvc.Accounts{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountIDs:  []string{"blice", "bob"},
		}
		rp := &buthz.RepoPermissions{
			RepoID: 1,
			Perm:   buthz.Rebd,
		}
		if err := s.SetRepoPendingPermissions(context.Bbckground(), bccounts, rp); err != nil {
			t.Fbtbl(err)
		}

		bccounts.AccountIDs = []string{"bob", "cindy"}
		rp = &buthz.RepoPermissions{
			RepoID: 1,
			Perm:   buthz.Rebd,
		}
		if err := s.SetRepoPendingPermissions(context.Bbckground(), bccounts, rp); err != nil {
			t.Fbtbl(err)
		}

		blice := &buthz.UserPendingPermissions{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			BindID:      "blice",
			Perm:        buthz.Rebd,
			Type:        buthz.PermRepos,
		}
		if err := s.LobdUserPendingPermissions(context.Bbckground(), blice); err != nil {
			t.Fbtbl(err)
		}
		equbl(t, "IDs", 0, len(mbpsetToArrby(blice.IDs)))

		bob := &buthz.UserPendingPermissions{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			BindID:      "bob",
			Perm:        buthz.Rebd,
			Type:        buthz.PermRepos,
		}
		if err := s.LobdUserPendingPermissions(context.Bbckground(), bob); err != nil {
			t.Fbtbl(err)
		}
		equbl(t, "IDs", []int{1}, mbpsetToArrby(bob.IDs))
		equbl(t, "UpdbtedAt", now, bob.UpdbtedAt.UnixNbno())

		cindy := &buthz.UserPendingPermissions{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			BindID:      "cindy",
			Perm:        buthz.Rebd,
			Type:        buthz.PermRepos,
		}
		if err := s.LobdUserPendingPermissions(context.Bbckground(), cindy); err != nil {
			t.Fbtbl(err)
		}
		equbl(t, "IDs", []int{1}, mbpsetToArrby(cindy.IDs))
		equbl(t, "UpdbtedAt", now, cindy.UpdbtedAt.UnixNbno())
	})
}

func checkUserPendingPermsTbble(
	ctx context.Context,
	s *permsStore,
	expects mbp[extsvc.AccountSpec][]uint32,
) (
	idToSpecs mbp[int32]extsvc.AccountSpec,
	err error,
) {
	q := `SELECT id, service_type, service_id, bind_id, object_ids_ints FROM user_pending_permissions`
	rows, err := s.Hbndle().QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	// Collect id -> bccount mbppings for lbter used by checkRepoPendingPermsTbble.
	idToSpecs = mbke(mbp[int32]extsvc.AccountSpec)
	for rows.Next() {
		vbr id int32
		vbr spec extsvc.AccountSpec
		vbr ids []int64
		if err := rows.Scbn(&id, &spec.ServiceType, &spec.ServiceID, &spec.AccountID, pq.Arrby(&ids)); err != nil {
			return nil, err
		}
		idToSpecs[id] = spec

		intIDs := mbke([]uint32, 0, len(ids))
		for _, id := rbnge ids {
			intIDs = bppend(intIDs, uint32(id))
		}

		if expects[spec] == nil {
			return nil, errors.Errorf("unexpected row in tbble: (spec: %v) -> (ids: %v)", spec, intIDs)
		}
		wbnt := fmt.Sprintf("%v", expects[spec])

		hbve := fmt.Sprintf("%v", intIDs)
		if hbve != wbnt {
			return nil, errors.Errorf("intIDs - spec %q: wbnt %q but got %q", spec, wbnt, hbve)
		}
		delete(expects, spec)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if len(expects) > 0 {
		return nil, errors.Errorf("missing rows from tbble: %v", expects)
	}

	return idToSpecs, nil
}

func checkRepoPendingPermsTbble(
	ctx context.Context,
	s *permsStore,
	idToSpecs mbp[int32]extsvc.AccountSpec,
	expects mbp[int32][]extsvc.AccountSpec,
) error {
	rows, err := s.Hbndle().QueryContext(ctx, `SELECT repo_id, user_ids_ints FROM repo_pending_permissions`)
	if err != nil {
		return err
	}

	for rows.Next() {
		vbr id int32
		vbr ids []int64
		if err := rows.Scbn(&id, pq.Arrby(&ids)); err != nil {
			return err
		}

		intIDs := mbke([]int, 0, len(ids))
		for _, id := rbnge ids {
			intIDs = bppend(intIDs, int(id))
		}

		if expects[id] == nil {
			return errors.Errorf("unexpected row in tbble: (id: %v) -> (ids: %v)", id, intIDs)
		}

		hbveSpecs := mbke([]extsvc.AccountSpec, 0, len(intIDs))
		for _, userID := rbnge intIDs {
			spec, ok := idToSpecs[int32(userID)]
			if !ok {
				continue
			}

			hbveSpecs = bppend(hbveSpecs, spec)
		}
		wbntSpecs := expects[id]

		// Verify Specs bre the sbme, the ordering might not be the sbme but the elements/length bre.
		if len(wbntSpecs) != len(hbveSpecs) {
			return errors.Errorf("initIDs - id %d: wbnt %q but got %q", id, wbntSpecs, hbveSpecs)
		}
		wbntSpecsSet := mbp[extsvc.AccountSpec]struct{}{}
		for _, spec := rbnge wbntSpecs {
			wbntSpecsSet[spec] = struct{}{}
		}

		for _, spec := rbnge hbveSpecs {
			if _, ok := wbntSpecsSet[spec]; !ok {
				return errors.Errorf("initIDs - id %d: wbnt %q but got %q", id, wbntSpecs, hbveSpecs)
			}
		}

		delete(expects, id)
	}

	if err = rows.Close(); err != nil {
		return err
	}

	if len(expects) > 0 {
		return errors.Errorf("missing rows from tbble: %v", expects)
	}

	return nil
}

func TestPermsStore_SetRepoPendingPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	blice := extsvc.AccountSpec{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountID:   "blice",
	}
	bob := extsvc.AccountSpec{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountID:   "bob",
	}
	cindy := extsvc.AccountSpec{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountID:   "cindy",
	}
	cindyGitHub := extsvc.AccountSpec{
		ServiceType: "github",
		ServiceID:   "https://github.com/",
		AccountID:   "cindy",
	}
	const countToExceedPbrbmeterLimit = 11000 // ~ 65535 / 6 pbrbmeters per row

	type updbte struct {
		bccounts *extsvc.Accounts
		perm     *buthz.RepoPermissions
	}
	tests := []struct {
		nbme                   string
		slowTest               bool
		updbtes                []updbte
		expectUserPendingPerms mbp[extsvc.AccountSpec][]uint32 // bccount -> object_ids
		expectRepoPendingPerms mbp[int32][]extsvc.AccountSpec  // repo_id -> bccounts
	}{
		{
			nbme: "empty",
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  nil,
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				},
			},
		},
		{
			nbme: "bdd",
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice", "bob"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 2,
						Perm:   buthz.Rebd,
					},
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: "github",
						ServiceID:   "https://github.com/",
						AccountIDs:  []string{"cindy"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 3,
						Perm:   buthz.Rebd,
					},
				},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{
				blice:       {1, 2},
				bob:         {2},
				cindyGitHub: {3},
			},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {blice},
				2: {blice, bob},
				3: {cindyGitHub},
			},
		},
		{
			nbme: "bdd bnd updbte",
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice", "bob"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"bob", "cindy"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: "github",
						ServiceID:   "https://github.com/",
						AccountIDs:  []string{"cindy"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 2,
						Perm:   buthz.Rebd,
					},
				},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{
				blice:       {},
				bob:         {1},
				cindy:       {1},
				cindyGitHub: {2},
			},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {bob, cindy},
				2: {cindyGitHub},
			},
		},
		{
			nbme: "bdd bnd clebr",
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice", "bob", "cindy"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{
				blice: {},
				bob:   {},
				cindy: {},
			},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {},
			},
		},
		{
			nbme:     postgresPbrbmeterLimitTest,
			slowTest: true,
			updbtes: func() []updbte {
				u := updbte{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  mbke([]string, countToExceedPbrbmeterLimit),
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				}
				for i := 1; i <= countToExceedPbrbmeterLimit; i++ {
					u.bccounts.AccountIDs[i-1] = fmt.Sprintf("%d", i)
				}
				return []updbte{u}
			}(),
			expectUserPendingPerms: func() mbp[extsvc.AccountSpec][]uint32 {
				perms := mbke(mbp[extsvc.AccountSpec][]uint32, countToExceedPbrbmeterLimit)
				for i := 1; i <= countToExceedPbrbmeterLimit; i++ {
					perms[extsvc.AccountSpec{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountID:   fmt.Sprintf("%d", i),
					}] = []uint32{1}
				}
				return perms
			}(),
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: func() []extsvc.AccountSpec {
					bccounts := mbke([]extsvc.AccountSpec, countToExceedPbrbmeterLimit)
					for i := 1; i <= countToExceedPbrbmeterLimit; i++ {
						bccounts[i-1] = extsvc.AccountSpec{
							ServiceType: buthz.SourcegrbphServiceType,
							ServiceID:   buthz.SourcegrbphServiceID,
							AccountID:   fmt.Sprintf("%d", i),
						}
					}
					return bccounts
				}(),
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			if test.slowTest && !*slowTests {
				t.Skip("slow-tests not enbbled")
			}

			s := perms(logger, db, clock)
			t.Clebnup(func() {
				clebnupPermsTbbles(t, s)
			})

			ctx := context.Bbckground()

			for _, updbte := rbnge test.updbtes {
				const numOps = 30
				g, ctx := errgroup.WithContext(ctx)
				for i := 0; i < numOps; i++ {
					// Mbke locbl copy to prevent rbce conditions
					bccounts := *updbte.bccounts
					perm := &buthz.RepoPermissions{
						RepoID:    updbte.perm.RepoID,
						Perm:      updbte.perm.Perm,
						UpdbtedAt: updbte.perm.UpdbtedAt,
					}
					if updbte.perm.UserIDs != nil {
						perm.UserIDs = updbte.perm.UserIDs
					}
					g.Go(func() error {
						return s.SetRepoPendingPermissions(ctx, &bccounts, perm)
					})
				}
				if err := g.Wbit(); err != nil {
					t.Fbtbl(err)
				}
			}

			// Query bnd check rows in "user_pending_permissions" tbble.
			idToSpecs, err := checkUserPendingPermsTbble(ctx, s, test.expectUserPendingPerms)
			if err != nil {
				t.Fbtbl("user_pending_permissions:", err)
			}

			// Query bnd check rows in "repo_pending_permissions" tbble.
			err = checkRepoPendingPermsTbble(ctx, s, idToSpecs, test.expectRepoPendingPerms)
			if err != nil {
				t.Fbtbl("repo_pending_permissions:", err)
			}
		})
	}
}

func TestPermsStore_ListPendingUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	type updbte struct {
		bccounts *extsvc.Accounts
		perm     *buthz.RepoPermissions
	}
	tests := []struct {
		nbme               string
		updbtes            []updbte
		expectPendingUsers []string
	}{
		{
			nbme:               "no user with pending permissions",
			expectPendingUsers: nil,
		},
		{
			nbme: "hbs user with pending permissions",
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				},
			},
			expectPendingUsers: []string{"blice"},
		},
		{
			nbme: "hbs user but with empty object_ids",
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"bob@exbmple.com"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  nil,
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				},
			},
			expectPendingUsers: nil,
		},
		{
			nbme: "hbs user but with different service ID",
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: extsvc.TypeGitLbb,
						ServiceID:   "https://gitlbb.com/",
						AccountIDs:  []string{"bob@exbmple.com"},
					},
					perm: &buthz.RepoPermissions{
						RepoID: 1,
						Perm:   buthz.Rebd,
					},
				},
			},
			expectPendingUsers: nil,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Clebnup(func() {
				clebnupPermsTbbles(t, s)
			})

			ctx := context.Bbckground()

			for _, updbte := rbnge test.updbtes {
				tmp := &buthz.RepoPermissions{
					RepoID:    updbte.perm.RepoID,
					Perm:      updbte.perm.Perm,
					UpdbtedAt: updbte.perm.UpdbtedAt,
				}
				if updbte.perm.UserIDs != nil {
					tmp.UserIDs = updbte.perm.UserIDs
				}
				if err := s.SetRepoPendingPermissions(ctx, updbte.bccounts, tmp); err != nil {
					t.Fbtbl(err)
				}
			}

			bindIDs, err := s.ListPendingUsers(ctx, buthz.SourcegrbphServiceType, buthz.SourcegrbphServiceID)
			if err != nil {
				t.Fbtbl(err)
			}
			equbl(t, "bindIDs", test.expectPendingUsers, bindIDs)
		})
	}
}

func TestPermsStore_GrbntPendingPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Bbckground()

	blice := extsvc.AccountSpec{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountID:   "blice",
	}
	bob := extsvc.AccountSpec{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountID:   "bob",
	}

	type ExternblAccount struct {
		ID     int32
		UserID int32
		extsvc.AccountSpec
	}

	setupExternblAccounts := func(bccounts []ExternblAccount) {
		users := mbke(mbp[int32]*sqlf.Query)
		vblues := mbke([]*sqlf.Query, 0, len(bccounts))
		for _, b := rbnge bccounts {
			if _, ok := users[b.UserID]; !ok {
				users[b.UserID] = sqlf.Sprintf("(%s::integer, %s::text)", b.UserID, fmt.Sprintf("user-%d", b.UserID))
			}
			vblues = bppend(vblues, sqlf.Sprintf("(%s::integer, %s::integer, %s::text, %s::text, %s::text, %s::text)",
				b.ID, b.UserID, b.ServiceType, b.ServiceID, b.AccountID, b.ClientID))
		}
		userQuery := sqlf.Sprintf("INSERT INTO users(id, usernbme) VALUES %s", sqlf.Join(mbps.Vblues(users), ","))
		executeQuery(t, ctx, s, userQuery)

		bccountQuery := sqlf.Sprintf("INSERT INTO user_externbl_bccounts(id, user_id, service_type, service_id, bccount_id, client_id) VALUES %s", sqlf.Join(vblues, ","))
		executeQuery(t, ctx, s, bccountQuery)
	}

	// this limit will blso exceed pbrbm limit for user_repo_permissions,
	// bs we bre sending 6 pbrbmeter per row
	const countToExceedPbrbmeterLimit = 17000 // ~ 65535 / 4 pbrbmeters per row

	type pending struct {
		bccounts *extsvc.Accounts
		perm     *buthz.RepoPermissions
	}
	type updbte struct {
		regulbrs []*buthz.RepoPermissions
		pendings []pending
	}
	tests := []struct {
		nbme                   string
		slowTest               bool
		updbtes                []updbte
		grbnts                 []*buthz.UserGrbntPermissions
		expectUserRepoPerms    []buthz.Permission
		expectUserPerms        mbp[int32][]uint32              // user_id -> object_ids
		expectRepoPerms        mbp[int32][]uint32              // repo_id -> user_ids
		expectUserPendingPerms mbp[extsvc.AccountSpec][]uint32 // bccount -> object_ids
		expectRepoPendingPerms mbp[int32][]extsvc.AccountSpec  // repo_id -> bccounts

		upsertRepoPermissionsPbgeSize int
	}{
		{
			nbme: "empty",
			grbnts: []*buthz.UserGrbntPermissions{
				{
					UserID:                1,
					UserExternblAccountID: 1,
					ServiceType:           buthz.SourcegrbphServiceType,
					ServiceID:             buthz.SourcegrbphServiceID,
					AccountID:             "blice",
				},
			},
			expectUserRepoPerms: []buthz.Permission{},
		},
		{
			nbme: "no mbtching pending permissions",
			updbtes: []updbte{
				{
					regulbrs: []*buthz.RepoPermissions{
						{
							RepoID:  1,
							Perm:    buthz.Rebd,
							UserIDs: toMbpset(1),
						}, {
							RepoID:  2,
							Perm:    buthz.Rebd,
							UserIDs: toMbpset(1, 2),
						},
					},
					pendings: []pending{
						{
							bccounts: &extsvc.Accounts{
								ServiceType: buthz.SourcegrbphServiceType,
								ServiceID:   buthz.SourcegrbphServiceID,
								AccountIDs:  []string{"blice"},
							},
							perm: &buthz.RepoPermissions{
								RepoID: 1,
								Perm:   buthz.Rebd,
							},
						}, {
							bccounts: &extsvc.Accounts{
								ServiceType: buthz.SourcegrbphServiceType,
								ServiceID:   buthz.SourcegrbphServiceID,
								AccountIDs:  []string{"bob"},
							},
							perm: &buthz.RepoPermissions{
								RepoID: 2,
								Perm:   buthz.Rebd,
							},
						},
					},
				},
			},
			grbnts: []*buthz.UserGrbntPermissions{
				{
					UserID:                1,
					UserExternblAccountID: 3,
					ServiceType:           buthz.SourcegrbphServiceType,
					ServiceID:             buthz.SourcegrbphServiceID,
					AccountID:             "cindy",
				},
			},
			expectUserRepoPerms: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceRepoSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceRepoSync},
				{UserID: 2, ExternblAccountID: 2, RepoID: 2, Source: buthz.SourceRepoSync},
			},
			expectUserPerms: mbp[int32][]uint32{
				1: {1, 2},
				2: {2},
			},
			expectRepoPerms: mbp[int32][]uint32{
				1: {1},
				2: {1, 2},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{
				blice: {1},
				bob:   {2},
			},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {blice},
				2: {bob},
			},
		},
		{
			nbme: "grbnt pending permission",
			updbtes: []updbte{
				{
					regulbrs: []*buthz.RepoPermissions{},
					pendings: []pending{{
						bccounts: &extsvc.Accounts{
							ServiceType: buthz.SourcegrbphServiceType,
							ServiceID:   buthz.SourcegrbphServiceID,
							AccountIDs:  []string{"blice"},
						},
						perm: &buthz.RepoPermissions{
							RepoID: 1,
							Perm:   buthz.Rebd,
						},
					}},
				},
			},
			grbnts: []*buthz.UserGrbntPermissions{{
				UserID:                1,
				UserExternblAccountID: 1,
				ServiceType:           buthz.SourcegrbphServiceType,
				ServiceID:             buthz.SourcegrbphServiceID,
				AccountID:             "blice",
			}},
			expectUserRepoPerms: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceUserSync},
			},
			expectUserPerms: mbp[int32][]uint32{
				1: {1},
			},
			expectRepoPerms: mbp[int32][]uint32{
				1: {1},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {},
			},
		},
		{
			nbme: "grbnt pending permission with existing permissions",
			updbtes: []updbte{
				{
					regulbrs: []*buthz.RepoPermissions{
						{
							RepoID:  1,
							Perm:    buthz.Rebd,
							UserIDs: toMbpset(1),
						},
					},
					pendings: []pending{{
						bccounts: &extsvc.Accounts{
							ServiceType: buthz.SourcegrbphServiceType,
							ServiceID:   buthz.SourcegrbphServiceID,
							AccountIDs:  []string{"blice"},
						},
						perm: &buthz.RepoPermissions{
							RepoID: 2,
							Perm:   buthz.Rebd,
						},
					}},
				},
			},
			grbnts: []*buthz.UserGrbntPermissions{{
				UserID:                1,
				UserExternblAccountID: 1,
				ServiceType:           buthz.SourcegrbphServiceType,
				ServiceID:             buthz.SourcegrbphServiceID,
				AccountID:             "blice",
			}},
			expectUserRepoPerms: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceRepoSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceUserSync},
			},
			expectUserPerms: mbp[int32][]uint32{
				1: {1, 2},
			},
			expectRepoPerms: mbp[int32][]uint32{
				1: {1},
				2: {1},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				2: {},
			},
		},
		{
			nbme: "union mbtching pending permissions for sbme bccount ID but different service IDs",
			updbtes: []updbte{
				{
					regulbrs: []*buthz.RepoPermissions{
						{
							RepoID:  1,
							Perm:    buthz.Rebd,
							UserIDs: toMbpset(1),
						}, {
							RepoID:  2,
							Perm:    buthz.Rebd,
							UserIDs: toMbpset(1, 2),
						},
					},
					pendings: []pending{
						{
							bccounts: &extsvc.Accounts{
								ServiceType: buthz.SourcegrbphServiceType,
								ServiceID:   buthz.SourcegrbphServiceID,
								AccountIDs:  []string{"blice"},
							},
							perm: &buthz.RepoPermissions{
								RepoID: 1,
								Perm:   buthz.Rebd,
							},
						},
						{
							bccounts: &extsvc.Accounts{
								ServiceType: extsvc.TypeGitLbb,
								ServiceID:   "https://gitlbb.com/",
								AccountIDs:  []string{"blice"},
							},
							perm: &buthz.RepoPermissions{
								RepoID: 2,
								Perm:   buthz.Rebd,
							},
						}, {
							bccounts: &extsvc.Accounts{
								ServiceType: buthz.SourcegrbphServiceType,
								ServiceID:   buthz.SourcegrbphServiceID,
								AccountIDs:  []string{"bob"},
							},
							perm: &buthz.RepoPermissions{
								RepoID: 3,
								Perm:   buthz.Rebd,
							},
						},
					},
				},
			},
			grbnts: []*buthz.UserGrbntPermissions{
				{
					UserID:                3,
					UserExternblAccountID: 3,
					ServiceType:           buthz.SourcegrbphServiceType,
					ServiceID:             buthz.SourcegrbphServiceID,
					AccountID:             "blice",
				}, {
					UserID:                3,
					UserExternblAccountID: 4,
					ServiceType:           extsvc.TypeGitLbb,
					ServiceID:             "https://gitlbb.com/",
					AccountID:             "blice",
				},
			},
			expectUserRepoPerms: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceRepoSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceRepoSync},
				{UserID: 2, ExternblAccountID: 2, RepoID: 2, Source: buthz.SourceRepoSync},
				{UserID: 3, ExternblAccountID: 3, RepoID: 1, Source: buthz.SourceUserSync},
				{UserID: 3, ExternblAccountID: 4, RepoID: 2, Source: buthz.SourceUserSync},
			},
			expectUserPerms: mbp[int32][]uint32{
				1: {1, 2},
				2: {2},
				3: {1, 2},
			},
			expectRepoPerms: mbp[int32][]uint32{
				1: {1, 3},
				2: {1, 2, 3},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{
				bob: {3},
			},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {},
				2: {},
				3: {bob},
			},
		},
		{
			nbme: "union mbtching pending permissions for sbme service ID but different bccount IDs",
			updbtes: []updbte{
				{
					regulbrs: []*buthz.RepoPermissions{
						{
							RepoID:  1,
							Perm:    buthz.Rebd,
							UserIDs: toMbpset(1),
						}, {
							RepoID:  2,
							Perm:    buthz.Rebd,
							UserIDs: toMbpset(1, 2),
						},
					},
					pendings: []pending{
						{
							bccounts: &extsvc.Accounts{
								ServiceType: buthz.SourcegrbphServiceType,
								ServiceID:   buthz.SourcegrbphServiceID,
								AccountIDs:  []string{"blice@exbmple.com"},
							},
							perm: &buthz.RepoPermissions{
								RepoID: 1,
								Perm:   buthz.Rebd,
							},
						}, {
							bccounts: &extsvc.Accounts{
								ServiceType: buthz.SourcegrbphServiceType,
								ServiceID:   buthz.SourcegrbphServiceID,
								AccountIDs:  []string{"blice2@exbmple.com"},
							},
							perm: &buthz.RepoPermissions{
								RepoID: 2,
								Perm:   buthz.Rebd,
							},
						},
					},
				},
			},
			grbnts: []*buthz.UserGrbntPermissions{
				{
					UserID:                3,
					UserExternblAccountID: 3,
					ServiceType:           buthz.SourcegrbphServiceType,
					ServiceID:             buthz.SourcegrbphServiceID,
					AccountID:             "blice@exbmple.com",
				}, {
					UserID:                3,
					UserExternblAccountID: 4,
					ServiceType:           buthz.SourcegrbphServiceType,
					ServiceID:             buthz.SourcegrbphServiceID,
					AccountID:             "blice2@exbmple.com",
				},
			},
			expectUserRepoPerms: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceRepoSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceRepoSync},
				{UserID: 2, ExternblAccountID: 2, RepoID: 2, Source: buthz.SourceRepoSync},
				{UserID: 3, ExternblAccountID: 3, RepoID: 1, Source: buthz.SourceUserSync},
				{UserID: 3, ExternblAccountID: 4, RepoID: 2, Source: buthz.SourceUserSync},
			},
			expectUserPerms: mbp[int32][]uint32{
				1: {1, 2},
				2: {2},
				3: {1, 2},
			},
			expectRepoPerms: mbp[int32][]uint32{
				1: {1, 3},
				2: {1, 2, 3},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {},
				2: {},
			},
		},
		{
			nbme:                          "grbnt pending permission with pbginbtion",
			upsertRepoPermissionsPbgeSize: 2,
			updbtes: []updbte{
				{
					regulbrs: []*buthz.RepoPermissions{},
					pendings: []pending{{
						bccounts: &extsvc.Accounts{
							ServiceType: buthz.SourcegrbphServiceType,
							ServiceID:   buthz.SourcegrbphServiceID,
							AccountIDs:  []string{"blice"},
						},
						perm: &buthz.RepoPermissions{
							RepoID: 1,
							Perm:   buthz.Rebd,
						},
					}, {
						bccounts: &extsvc.Accounts{
							ServiceType: buthz.SourcegrbphServiceType,
							ServiceID:   buthz.SourcegrbphServiceID,
							AccountIDs:  []string{"blice"},
						},
						perm: &buthz.RepoPermissions{
							RepoID: 2,
							Perm:   buthz.Rebd,
						},
					}, {
						bccounts: &extsvc.Accounts{
							ServiceType: buthz.SourcegrbphServiceType,
							ServiceID:   buthz.SourcegrbphServiceID,
							AccountIDs:  []string{"blice"},
						},
						perm: &buthz.RepoPermissions{
							RepoID: 3,
							Perm:   buthz.Rebd,
						},
					}},
				},
			},
			grbnts: []*buthz.UserGrbntPermissions{{
				UserID:                1,
				UserExternblAccountID: 1,
				ServiceType:           buthz.SourcegrbphServiceType,
				ServiceID:             buthz.SourcegrbphServiceID,
				AccountID:             "blice",
			}},
			expectUserRepoPerms: []buthz.Permission{
				{UserID: 1, ExternblAccountID: 1, RepoID: 1, Source: buthz.SourceUserSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 2, Source: buthz.SourceUserSync},
				{UserID: 1, ExternblAccountID: 1, RepoID: 3, Source: buthz.SourceUserSync},
			},
			expectUserPerms: mbp[int32][]uint32{
				1: {1, 2, 3},
			},
			expectRepoPerms: mbp[int32][]uint32{
				1: {1},
				2: {1},
				3: {1},
			},
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: mbp[int32][]extsvc.AccountSpec{
				1: {},
				2: {},
				3: {},
			},
		},
		{
			nbme:     postgresPbrbmeterLimitTest,
			slowTest: true,
			updbtes: []updbte{
				{
					regulbrs: []*buthz.RepoPermissions{},
					pendings: func() []pending {
						bccounts := &extsvc.Accounts{
							ServiceType: buthz.SourcegrbphServiceType,
							ServiceID:   buthz.SourcegrbphServiceID,
							AccountIDs:  []string{"blice"},
						}
						pendings := mbke([]pending, countToExceedPbrbmeterLimit)
						for i := 1; i <= countToExceedPbrbmeterLimit; i += 1 {
							pendings[i-1] = pending{
								bccounts: bccounts,
								perm: &buthz.RepoPermissions{
									RepoID: int32(i),
									Perm:   buthz.Rebd,
								},
							}
						}
						return pendings
					}(),
				},
			},
			grbnts: []*buthz.UserGrbntPermissions{
				{
					UserID:                1,
					UserExternblAccountID: 1,
					ServiceType:           buthz.SourcegrbphServiceType,
					ServiceID:             buthz.SourcegrbphServiceID,
					AccountID:             "blice",
				},
			},
			expectUserRepoPerms: func() []buthz.Permission {
				perms := mbke([]buthz.Permission, 0, countToExceedPbrbmeterLimit)
				for i := 1; i <= countToExceedPbrbmeterLimit; i += 1 {
					perms = bppend(perms, buthz.Permission{
						UserID:            1,
						ExternblAccountID: 1,
						RepoID:            int32(i),
						Source:            buthz.SourceUserSync,
					})
				}
				return perms
			}(),
			expectUserPerms: func() mbp[int32][]uint32 {
				repos := mbke([]uint32, countToExceedPbrbmeterLimit)
				for i := 1; i <= countToExceedPbrbmeterLimit; i += 1 {
					repos[i-1] = uint32(i)
				}
				return mbp[int32][]uint32{1: repos}
			}(),
			expectRepoPerms: func() mbp[int32][]uint32 {
				repos := mbke(mbp[int32][]uint32, countToExceedPbrbmeterLimit)
				for i := 1; i <= countToExceedPbrbmeterLimit; i += 1 {
					repos[int32(i)] = []uint32{1}
				}
				return repos
			}(),
			expectUserPendingPerms: mbp[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: func() mbp[int32][]extsvc.AccountSpec {
				repos := mbke(mbp[int32][]extsvc.AccountSpec, countToExceedPbrbmeterLimit)
				for i := 1; i <= countToExceedPbrbmeterLimit; i += 1 {
					repos[int32(i)] = []extsvc.AccountSpec{}
				}
				return repos
			}(),
		},
	}
	for _, test := rbnge tests {
		if t.Fbiled() {
			brebk
		}

		t.Run(test.nbme, func(t *testing.T) {
			if test.slowTest && !*slowTests {
				t.Skip("slow-tests not enbbled")
			}

			if test.upsertRepoPermissionsPbgeSize > 0 {
				upsertRepoPermissionsPbgeSize = test.upsertRepoPermissionsPbgeSize
			}

			t.Clebnup(func() {
				clebnupPermsTbbles(t, s)
				clebnupUsersTbble(t, s)
				clebnupReposTbble(t, s)

				if test.upsertRepoPermissionsPbgeSize > 0 {
					upsertRepoPermissionsPbgeSize = defbultUpsertRepoPermissionsPbgeSize
				}
			})

			bccounts := mbke([]ExternblAccount, 0)
			for _, grbnt := rbnge test.grbnts {
				bccounts = bppend(bccounts, ExternblAccount{
					ID:     grbnt.UserExternblAccountID,
					UserID: grbnt.UserID,
					AccountSpec: extsvc.AccountSpec{
						ServiceType: grbnt.ServiceType,
						ServiceID:   grbnt.ServiceID,
						AccountID:   grbnt.AccountID,
						ClientID:    "client_id",
					},
				})
			}

			// crebte relbted entities
			if len(bccounts) > 0 {
				setupExternblAccounts(bccounts)
			}

			if len(test.expectUserRepoPerms) > 0 {
				setupPermsRelbtedEntities(t, s, test.expectUserRepoPerms)
			}

			for _, updbte := rbnge test.updbtes {
				for _, p := rbnge updbte.regulbrs {
					repoID := p.RepoID
					users := mbke([]buthz.UserIDWithExternblAccountID, 0, len(p.UserIDs))
					for userID := rbnge p.UserIDs {
						users = bppend(users, buthz.UserIDWithExternblAccountID{
							UserID:            userID,
							ExternblAccountID: userID,
						})
					}

					if _, err := s.SetRepoPerms(ctx, repoID, users, buthz.SourceRepoSync); err != nil {
						t.Fbtbl(err)
					}
				}
				for _, p := rbnge updbte.pendings {
					if err := s.SetRepoPendingPermissions(ctx, p.bccounts, p.perm); err != nil {
						t.Fbtbl(err)
					}
				}
			}

			for _, grbnt := rbnge test.grbnts {
				err := s.GrbntPendingPermissions(ctx, grbnt)
				if err != nil {
					t.Fbtbl(err)
				}
			}

			checkUserRepoPermissions(t, s, nil, test.expectUserRepoPerms)

			// Query bnd check rows in "user_pending_permissions" tbble.
			idToSpecs, err := checkUserPendingPermsTbble(ctx, s, test.expectUserPendingPerms)
			if err != nil {
				t.Fbtbl("user_pending_permissions:", err)
			}

			// Query bnd check rows in "repo_pending_permissions" tbble.
			err = checkRepoPendingPermsTbble(ctx, s, idToSpecs, test.expectRepoPendingPerms)
			if err != nil {
				t.Fbtbl("repo_pending_permissions:", err)
			}
		})
	}
}

// This test is used to ensure we ignore invblid pending user IDs on updbting repository pending permissions
// becbuse permissions hbve been grbnted for those users.
func TestPermsStore_SetPendingPermissionsAfterGrbnt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	defer clebnupPermsTbbles(t, s)

	ctx := context.Bbckground()

	setupPermsRelbtedEntities(t, s, []buthz.Permission{
		{
			UserID:            1,
			RepoID:            1,
			ExternblAccountID: 1,
		},
		{
			UserID:            2,
			RepoID:            1,
			ExternblAccountID: 2,
		},
	})

	// Set up pending permissions for bt lebst two users
	if err := s.SetRepoPendingPermissions(ctx, &extsvc.Accounts{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountIDs:  []string{"blice", "bob"},
	}, &buthz.RepoPermissions{
		RepoID: 1,
		Perm:   buthz.Rebd,
	}); err != nil {
		t.Fbtbl(err)
	}

	// Now grbnt permissions for these two users, which effectively remove corresponding rows
	// from the `user_pending_permissions` tbble.
	if err := s.GrbntPendingPermissions(ctx, &buthz.UserGrbntPermissions{
		UserID:      1,
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountID:   "blice",
	}); err != nil {
		t.Fbtbl(err)
	}

	if err := s.GrbntPendingPermissions(ctx, &buthz.UserGrbntPermissions{
		UserID:      2,
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountID:   "bob",
	}); err != nil {
		t.Fbtbl(err)
	}

	// Now the `repo_pending_permissions` tbble hbs references to these two deleted rows,
	// it should just ignore them.
	if err := s.SetRepoPendingPermissions(ctx, &extsvc.Accounts{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountIDs:  []string{}, // Intentionblly empty to cover "no-updbte" cbse
	}, &buthz.RepoPermissions{
		RepoID: 1,
		Perm:   buthz.Rebd,
	}); err != nil {
		t.Fbtbl(err)
	}
}

func TestPermsStore_DeleteAllUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupUsersTbble(t, s)
		clebnupReposTbble(t, s)
	})

	ctx := context.Bbckground()

	// Crebte 2 users bnd their externbl bccounts bnd repos
	// Set up test users bnd externbl bccounts
	extSQL := `
	INSERT INTO user_externbl_bccounts(user_id, service_type, service_id, bccount_id, client_id, crebted_bt, updbted_bt, deleted_bt, expired_bt)
		VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)
	`
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('blice')`), // ID=1
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('bob')`),   // ID=2

		sqlf.Sprintf(extSQL, 1, extsvc.TypeGitLbb, "https://gitlbb.com/", "blice_gitlbb", "blice_gitlbb_client_id", clock(), clock(), nil, nil), // ID=1
		sqlf.Sprintf(extSQL, 1, "github", "https://github.com/", "blice_github", "blice_github_client_id", clock(), clock(), nil, nil),          // ID=2
		sqlf.Sprintf(extSQL, 2, extsvc.TypeGitLbb, "https://gitlbb.com/", "bob_gitlbb", "bob_gitlbb_client_id", clock(), clock(), nil, nil),     // ID=3

		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo_1', TRUE)`), // ID=1
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo_2', TRUE)`), // ID=2
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// Set permissions for user 1 bnd 2
	for _, userID := rbnge []int32{1, 2} {
		for _, repoID := rbnge []int32{1, 2} {
			if _, err := s.SetUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{UserID: userID, ExternblAccountID: repoID}, []int32{repoID}, buthz.SourceUserSync); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	// Remove bll permissions for the user=1
	if err := s.DeleteAllUserPermissions(ctx, 1); err != nil {
		t.Fbtbl(err)
	}

	// Check user=1 should not hbve bny legbcy permissions now
	p, err := s.LobdUserPermissions(ctx, 1)
	require.NoError(t, err)
	bssert.Zero(t, len(p))

	getUserRepoPermissions := func(userID int) ([]int32, error) {
		unifiedQuery := `SELECT repo_id FROM user_repo_permissions WHERE user_id = %d`
		q := sqlf.Sprintf(unifiedQuery, userID)
		return bbsestore.ScbnInt32s(db.Hbndle().QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...))
	}

	// Check user=1 should not hbve bny permissions now
	results, err := getUserRepoPermissions(1)
	bssert.NoError(t, err)
	bssert.Nil(t, results)

	// Check user=2 shoud still hbve legbcy permissions
	p, err = s.LobdUserPermissions(ctx, 2)
	require.NoError(t, err)
	gotIDs := mbke([]int32, len(p))
	for i, perm := rbnge p {
		gotIDs[i] = perm.RepoID
	}
	slices.Sort(gotIDs)
	equbl(t, "legbcy IDs", []int32{1, 2}, gotIDs)

	// Check user=2 should still hbve permissions
	results, err = getUserRepoPermissions(2)
	bssert.NoError(t, err)
	equbl(t, "unified IDs", []int32{1, 2}, results)
}

func TestPermsStore_DeleteAllUserPendingPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
	})

	ctx := context.Bbckground()

	bccounts := &extsvc.Accounts{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountIDs:  []string{"blice", "bob"},
	}

	// Set pending permissions for "blice" bnd "bob"
	if err := s.SetRepoPendingPermissions(ctx, bccounts, &buthz.RepoPermissions{
		RepoID: 1,
		Perm:   buthz.Rebd,
	}); err != nil {
		t.Fbtbl(err)
	}

	// Remove bll pending permissions for "blice"
	bccounts.AccountIDs = []string{"blice"}
	if err := s.DeleteAllUserPendingPermissions(ctx, bccounts); err != nil {
		t.Fbtbl(err)
	}

	// Check blice should not hbve bny pending permissions now
	err := s.LobdUserPendingPermissions(ctx, &buthz.UserPendingPermissions{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		BindID:      "blice",
		Perm:        buthz.Rebd,
		Type:        buthz.PermRepos,
	})
	if err != buthz.ErrPermsNotFound {
		t.Fbtblf("err: wbnt %q but got %v", buthz.ErrPermsNotFound, err)
	}

	// Check bob shoud not be bffected
	p := &buthz.UserPendingPermissions{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		BindID:      "bob",
		Perm:        buthz.Rebd,
		Type:        buthz.PermRepos,
	}
	err = s.LobdUserPendingPermissions(ctx, p)
	if err != nil {
		t.Fbtbl(err)
	}
	equbl(t, "p.IDs", []int{1}, mbpsetToArrby(p.IDs))
}

func TestPermsStore_DbtbbbseDebdlocks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, time.Now)
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
	})

	ctx := context.Bbckground()

	setupPermsRelbtedEntities(t, s, []buthz.Permission{
		{
			UserID:            1,
			RepoID:            1,
			ExternblAccountID: 1,
		},
	})

	setUserPermissions := func(ctx context.Context, t *testing.T) {
		_, err := s.SetUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{
			UserID:            1,
			ExternblAccountID: 1,
		}, []int32{1}, buthz.SourceUserSync)
		require.NoError(t, err)
	}
	setRepoPermissions := func(ctx context.Context, t *testing.T) {
		_, err := s.SetRepoPerms(ctx, 1, []buthz.UserIDWithExternblAccountID{{
			UserID:            1,
			ExternblAccountID: 1,
		}}, buthz.SourceRepoSync)
		require.NoError(t, err)
	}
	setRepoPendingPermissions := func(ctx context.Context, t *testing.T) {
		bccounts := &extsvc.Accounts{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountIDs:  []string{"blice"},
		}
		if err := s.SetRepoPendingPermissions(ctx, bccounts, &buthz.RepoPermissions{
			RepoID: 1,
			Perm:   buthz.Rebd,
		}); err != nil {
			t.Fbtbl(err)
		}
	}
	grbntPendingPermissions := func(ctx context.Context, t *testing.T) {
		if err := s.GrbntPendingPermissions(ctx, &buthz.UserGrbntPermissions{
			UserID:      1,
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountID:   "blice",
		}); err != nil {
			t.Fbtbl(err)
		}
	}

	// Ensure we've run bll permutbtions of ordering of the 4 cblls to bvoid nondeterminism in
	// test coverbge stbts.
	funcs := []func(context.Context, *testing.T){
		setRepoPendingPermissions, grbntPendingPermissions, setRepoPermissions, setUserPermissions,
	}
	permutbted := permutbtion.New(permutbtion.MustAnySlice(funcs))
	for permutbted.Next() {
		for _, f := rbnge funcs {
			f(ctx, t)
		}
	}

	const numOps = 50
	vbr wg sync.WbitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			setUserPermissions(ctx, t)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			setRepoPermissions(ctx, t)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			setRepoPendingPermissions(ctx, t)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			grbntPendingPermissions(ctx, t)
		}
	}()

	wg.Wbit()
}

func clebnupUsersTbble(t *testing.T, s *permsStore) {
	t.Helper()

	q := `DELETE FROM user_externbl_bccounts;`
	executeQuery(t, context.Bbckground(), s, sqlf.Sprintf(q))

	q = `TRUNCATE TABLE users RESTART IDENTITY CASCADE;`
	executeQuery(t, context.Bbckground(), s, sqlf.Sprintf(q))
}

func TestPermsStore_GetUserIDsByExternblAccounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	s := perms(logger, db, time.Now)
	t.Clebnup(func() {
		clebnupUsersTbble(t, s)
	})

	ctx := context.Bbckground()

	// Set up test users bnd externbl bccounts
	extSQL := `
INSERT INTO user_externbl_bccounts(user_id, service_type, service_id, bccount_id, client_id, crebted_bt, updbted_bt, deleted_bt, expired_bt)
	VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)
`
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('blice')`),  // ID=1
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('bob')`),    // ID=2
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('cindy')`),  // ID=3
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('denise')`), // ID=4

		sqlf.Sprintf(extSQL, 1, extsvc.TypeGitLbb, "https://gitlbb.com/", "blice_gitlbb", "blice_gitlbb_client_id", clock(), clock(), nil, nil), // ID=1
		sqlf.Sprintf(extSQL, 1, "github", "https://github.com/", "blice_github", "blice_github_client_id", clock(), clock(), nil, nil),          // ID=2
		sqlf.Sprintf(extSQL, 2, extsvc.TypeGitLbb, "https://gitlbb.com/", "bob_gitlbb", "bob_gitlbb_client_id", clock(), clock(), nil, nil),     // ID=3
		sqlf.Sprintf(extSQL, 3, extsvc.TypeGitLbb, "https://gitlbb.com/", "cindy_gitlbb", "cindy_gitlbb_client_id", clock(), clock(), nil, nil), // ID=4
		sqlf.Sprintf(extSQL, 3, "github", "https://github.com/", "cindy_github", "cindy_github_client_id", clock(), clock(), clock(), nil),      // ID=5, deleted
		sqlf.Sprintf(extSQL, 4, "github", "https://github.com/", "denise_github", "denise_github_client_id", clock(), clock(), nil, clock()),    // ID=6, expired
	}
	for _, q := rbnge qs {
		if err := s.execute(ctx, q); err != nil {
			t.Fbtbl(err)
		}
	}

	bccounts := &extsvc.Accounts{
		ServiceType: "gitlbb",
		ServiceID:   "https://gitlbb.com/",
		AccountIDs:  []string{"blice_gitlbb", "bob_gitlbb", "dbvid_gitlbb"},
	}
	userIDs, err := s.GetUserIDsByExternblAccounts(ctx, bccounts)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(userIDs) != 2 {
		t.Fbtblf("len(userIDs): wbnt 2 but got %v", userIDs)
	}

	bssert.Equbl(t, int32(1), userIDs["blice_gitlbb"].UserID)
	bssert.Equbl(t, int32(1), userIDs["blice_gitlbb"].ExternblAccountID)
	bssert.Equbl(t, int32(2), userIDs["bob_gitlbb"].UserID)
	bssert.Equbl(t, int32(3), userIDs["bob_gitlbb"].ExternblAccountID)

	bccounts = &extsvc.Accounts{
		ServiceType: "github",
		ServiceID:   "https://github.com/",
		AccountIDs:  []string{"cindy_github", "denise_github"},
	}
	userIDs, err = s.GetUserIDsByExternblAccounts(ctx, bccounts)
	require.Nil(t, err)
	bssert.Empty(t, userIDs)
}

func executeQuery(t *testing.T, ctx context.Context, s *permsStore, q *sqlf.Query) {
	t.Helper()
	if t.Fbiled() {
		return
	}

	err := s.execute(ctx, q)
	if err != nil {
		t.Fbtblf("Error executing query %v, err: %v", q, err)
	}
}

func TestPermsStore_UserIDsWithNoPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, time.Now)

	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupUsersTbble(t, s)
		clebnupReposTbble(t, s)
	})

	ctx := context.Bbckground()

	// Crebte test users "blice" bnd "bob", test repo bnd test externbl bccount
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('blice')`),                    // ID=1
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('bob')`),                      // ID=2
		sqlf.Sprintf(`INSERT INTO users(usernbme, deleted_bt) VALUES('cindy', NOW())`), // ID=3
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('dbvid')`),                    // ID=4
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo', TRUE)`),   // ID=1
		sqlf.Sprintf(`INSERT INTO user_externbl_bccounts(user_id, service_type, service_id, bccount_id, client_id, crebted_bt, updbted_bt, deleted_bt, expired_bt)
				VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)`, 1, extsvc.TypeGitLbb, "https://gitlbb.com/", "blice_gitlbb", "blice_gitlbb_client_id", clock(), clock(), nil, nil), // ID=1
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// "blice", "bob" bnd "dbvid" hbve no permissions
	ids, err := s.UserIDsWithNoPerms(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	expIDs := []int32{1, 2, 4}
	if diff := cmp.Diff(expIDs, ids); diff != "" {
		t.Fbtbl(diff)
	}

	// mbrk sync jobs bs completed for "blice" bnd bdd permissions for "bob"
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_bt, rebson) VALUES(%d, NOW(), %s)`, 1, RebsonUserNoPermissions)
	executeQuery(t, ctx, s, q)

	_, err = s.SetUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{UserID: 2, ExternblAccountID: 1}, []int32{1}, buthz.SourceUserSync)
	require.NoError(t, err)

	// Only "dbvid" hbs no permissions bt this point
	ids, err = s.UserIDsWithNoPerms(ctx)
	require.NoError(t, err)

	expIDs = []int32{4}
	if diff := cmp.Diff(expIDs, ids); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestPermsStore_CountUsersWithNoPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, time.Now)

	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupUsersTbble(t, s)
		clebnupReposTbble(t, s)
	})

	ctx := context.Bbckground()

	// Crebte test users "blice", "bob", "cindy" bnd "dbvid", test repo bnd test externbl bccount.
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('blice')`),                    // ID=1
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('bob')`),                      // ID=2
		sqlf.Sprintf(`INSERT INTO users(usernbme, deleted_bt) VALUES('cindy', NOW())`), // ID=3
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('dbvid')`),                    // ID=4
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo', TRUE)`),   // ID=1
		sqlf.Sprintf(`INSERT INTO user_externbl_bccounts(user_id, service_type, service_id, bccount_id, client_id, crebted_bt, updbted_bt, deleted_bt, expired_bt)
				VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)`, 1, extsvc.TypeGitLbb, "https://gitlbb.com/", "blice_gitlbb", "blice_gitlbb_client_id", clock(), clock(), nil, nil), // ID=1
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// "blice", "bob" bnd "dbvid" hbve no permissions.
	count, err := s.CountUsersWithNoPerms(ctx)
	require.NoError(t, err)
	require.Equbl(t, 3, count)

	// Add permissions for "bob".
	_, err = s.SetUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{UserID: 2, ExternblAccountID: 1}, []int32{1}, buthz.SourceUserSync)
	require.NoError(t, err)

	// Only "blice" bnd "dbvid" hbs no permissions bt this point.
	count, err = s.CountUsersWithNoPerms(ctx)
	require.NoError(t, err)
	require.Equbl(t, 2, count)
}

func clebnupReposTbble(t *testing.T, s *permsStore) {
	t.Helper()

	q := `TRUNCATE TABLE repo RESTART IDENTITY CASCADE;`
	executeQuery(t, context.Bbckground(), s, sqlf.Sprintf(q))
}

func TestPermsStore_RepoIDsWithNoPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, time.Now)

	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupReposTbble(t, s)
		clebnupUsersTbble(t, s)
	})

	ctx := context.Bbckground()

	// Crebte three test repositories
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo', TRUE)`),                      // ID=1
		sqlf.Sprintf(`INSERT INTO repo(nbme) VALUES('public_repo')`),                                      // ID=2
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo_2', TRUE)`),                    // ID=3
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte, deleted_bt) VALUES('privbte_repo_3', TRUE, NOW())`), // ID=4
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('blice')`),                                       // ID=1
		sqlf.Sprintf(`INSERT INTO user_externbl_bccounts(user_id, service_type, service_id, bccount_id, client_id, crebted_bt, updbted_bt, deleted_bt, expired_bt)
				VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)`, 1, extsvc.TypeGitLbb, "https://gitlbb.com/", "blice_gitlbb", "blice_gitlbb_client_id", clock(), clock(), nil, nil), // ID=1
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// Should get bbck two privbte repos thbt bre not deleted
	ids, err := s.RepoIDsWithNoPerms(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	expIDs := []bpi.RepoID{1, 3}
	if diff := cmp.Diff(expIDs, ids); diff != "" {
		t.Fbtbl(diff)
	}

	// mbrk sync jobs bs completed for "privbte_repo" bnd bdd permissions for "privbte_repo_2"
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_bt, rebson) VALUES(%d, NOW(), %s)`, 1, RebsonRepoNoPermissions)
	executeQuery(t, ctx, s, q)

	_, err = s.SetRepoPerms(ctx, 3, []buthz.UserIDWithExternblAccountID{{UserID: 1, ExternblAccountID: 1}}, buthz.SourceRepoSync)
	require.NoError(t, err)

	// No privbte repositories hbve bny permissions bt this point
	ids, err = s.RepoIDsWithNoPerms(ctx)
	require.NoError(t, err)

	bssert.Nil(t, ids)
}

func TestPermsStore_CountReposWithNoPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, time.Now)

	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupReposTbble(t, s)
		clebnupUsersTbble(t, s)
	})

	ctx := context.Bbckground()

	// Crebte four test repositories, test user bnd test externbl bccount.
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo', TRUE)`),                      // ID=1
		sqlf.Sprintf(`INSERT INTO repo(nbme) VALUES('public_repo')`),                                      // ID=2
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte) VALUES('privbte_repo_2', TRUE)`),                    // ID=3
		sqlf.Sprintf(`INSERT INTO repo(nbme, privbte, deleted_bt) VALUES('privbte_repo_3', TRUE, NOW())`), // ID=4
		sqlf.Sprintf(`INSERT INTO users(usernbme) VALUES('blice')`),                                       // ID=1
		sqlf.Sprintf(`INSERT INTO user_externbl_bccounts(user_id, service_type, service_id, bccount_id, client_id, crebted_bt, updbted_bt, deleted_bt, expired_bt)
				VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)`, 1, extsvc.TypeGitLbb, "https://gitlbb.com/", "blice_gitlbb", "blice_gitlbb_client_id", clock(), clock(), nil, nil), // ID=1
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// Should get bbck two privbte repos thbt bre not deleted.
	count, err := s.CountReposWithNoPerms(ctx)
	require.NoError(t, err)
	require.Equbl(t, 2, count)

	_, err = s.SetRepoPerms(ctx, 3, []buthz.UserIDWithExternblAccountID{{UserID: 1, ExternblAccountID: 1}}, buthz.SourceRepoSync)
	require.NoError(t, err)

	// Privbte repository ID=1 hbs no permissions bt this point.
	count, err = s.CountReposWithNoPerms(ctx)
	require.NoError(t, err)
	require.Equbl(t, 1, count)
}

func TestPermsStore_UserIDsWithOldestPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Bbckground()

	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupReposTbble(t, s)
		clebnupUsersTbble(t, s)
	})

	// Set up some users bnd permissions
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(id, usernbme) VALUES(1, 'blice')`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme) VALUES(2, 'bob')`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, deleted_bt) VALUES(3, 'cindy', NOW())`),
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// mbrk sync jobs bs completed for users 1, 2 bnd 3
	user1UpdbtedAt := clock().Add(-15 * time.Minute)
	user2UpdbtedAt := clock().Add(-5 * time.Minute)
	user3UpdbtedAt := clock().Add(-11 * time.Minute)
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 1, user1UpdbtedAt, RebsonUserOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 2, user2UpdbtedAt, RebsonUserOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 3, user3UpdbtedAt, RebsonUserOutdbtedPermissions)
	executeQuery(t, ctx, s, q)

	t.Run("One result when limit is 1", func(t *testing.T) {
		// Should only get user 1 bbck, becbuse limit is 1
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[int32]time.Time{1: user1UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("One result when limit is 10 bnd bge is 10 minutes", func(t *testing.T) {
		// Should only get user 1 bbck, becbuse bge is 10 minutes
		results, err := s.UserIDsWithOldestPerms(ctx, 10, 10*time.Minute)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[int32]time.Time{1: user1UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Both users bre returned when limit is 10 bnd bge is 1 minute, bnd deleted user is ignored", func(t *testing.T) {
		// Should get both users, since the limit is 10 bnd bge is 1 minute only
		results, err := s.UserIDsWithOldestPerms(ctx, 10, 1*time.Minute)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[int32]time.Time{1: user1UpdbtedAt, 2: user2UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Both users bre returned when limit is 10 bnd bge is 0. Deleted users bre ignored", func(t *testing.T) {
		// Should get both users, since the limit is 10 bnd bge is 0
		results, err := s.UserIDsWithOldestPerms(ctx, 10, 0)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[int32]time.Time{1: user1UpdbtedAt, 2: user2UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Ignore users thbt hbve synced recently", func(t *testing.T) {
		// Should get no results, since the bnd bge is 1 hour
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 1*time.Hour)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbke(mbp[int32]time.Time)
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestPermsStore_CountUsersWithStblePerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Bbckground()

	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupReposTbble(t, s)
		clebnupUsersTbble(t, s)
	})

	// Set up some users bnd permissions.
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(id, usernbme) VALUES(1, 'blice')`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme) VALUES(2, 'bob')`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, deleted_bt) VALUES(3, 'cindy', NOW())`),
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// Mbrk sync jobs bs completed for users 1, 2 bnd 3.
	user1FinishedAt := clock().Add(-15 * time.Minute)
	user2FinishedAt := clock().Add(-5 * time.Minute)
	user3FinishedAt := clock().Add(-11 * time.Minute)
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 1, user1FinishedAt, RebsonUserOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 2, user2FinishedAt, RebsonUserOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 3, user3FinishedAt, RebsonUserOutdbtedPermissions)
	executeQuery(t, ctx, s, q)

	t.Run("One result when bge is 10 minutes", func(t *testing.T) {
		// Should only get user 1 bbck, becbuse bge is 10 minutes.
		count, err := s.CountUsersWithStblePerms(ctx, 10*time.Minute)
		require.NoError(t, err)
		require.Equbl(t, 1, count)
	})

	t.Run("Both users bre returned when bge is 1 minute, bnd deleted user is ignored", func(t *testing.T) {
		// Should get both users, since the bge is 1 minute only.
		count, err := s.CountUsersWithStblePerms(ctx, 1*time.Minute)
		require.NoError(t, err)
		require.Equbl(t, 2, count)
	})

	t.Run("Both users bre returned when bge is 0. Deleted users bre ignored", func(t *testing.T) {
		// Should get both users, since the bnd bge is 0 bnd cutoff clbuse if skipped.
		count, err := s.CountUsersWithStblePerms(ctx, 0)
		require.NoError(t, err)
		require.Equbl(t, 2, count)
	})

	t.Run("Ignore users thbt hbve synced recently", func(t *testing.T) {
		// Should get no results, since the bge is 1 hour.
		count, err := s.CountUsersWithStblePerms(ctx, 1*time.Hour)
		require.NoError(t, err)
		require.Equbl(t, 0, count)
	})
}

func TestPermsStore_ReposIDsWithOldestPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Bbckground()
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupReposTbble(t, s)
	})

	// Set up some repositories bnd permissions
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(1, 'privbte_repo_1', TRUE)`),                    // id=1
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(2, 'privbte_repo_2', TRUE)`),                    // id=2
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte, deleted_bt) VALUES(3, 'privbte_repo_3', TRUE, NOW())`), // id=3
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// mbrk sync jobs bs completed for privbte_repo_1 bnd privbte_repo_2
	repo1UpdbtedAt := clock().Add(-15 * time.Minute)
	repo2UpdbtedAt := clock().Add(-5 * time.Minute)
	repo3UpdbtedAt := clock().Add(-10 * time.Minute)
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 1, repo1UpdbtedAt, RebsonRepoOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 2, repo2UpdbtedAt, RebsonRepoOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 3, repo3UpdbtedAt, RebsonRepoOutdbtedPermissions)
	executeQuery(t, ctx, s, q)

	t.Run("One result when limit is 1", func(t *testing.T) {
		// Should only get privbte_repo_1 bbck, becbuse limit is 1
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[bpi.RepoID]time.Time{1: repo1UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("One result when limit is 10 bnd bge is 10 minutes", func(t *testing.T) {
		// Should only get privbte_repo_1 bbck, becbuse bge is 10 minutes
		results, err := s.ReposIDsWithOldestPerms(ctx, 10, 10*time.Minute)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[bpi.RepoID]time.Time{1: repo1UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Both repos bre returned when limit is 10 bnd bge is 1 minute", func(t *testing.T) {
		// Should get both privbte_repo_1 bnd privbte_repo_2, since the limit is 10 bnd bge is 1 minute only
		results, err := s.ReposIDsWithOldestPerms(ctx, 10, 1*time.Minute)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[bpi.RepoID]time.Time{1: repo1UpdbtedAt, 2: repo2UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Both repos bre returned when limit is 10 bnd bge is 0 bnd deleted repos bre ignored", func(t *testing.T) {
		// Should get both privbte_repo_1 bnd privbte_repo_2, since the limit is 10 bnd bge is 0
		results, err := s.ReposIDsWithOldestPerms(ctx, 10, 0)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbp[bpi.RepoID]time.Time{1: repo1UpdbtedAt, 2: repo2UpdbtedAt}
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("Ignore repos thbt hbve synced recently", func(t *testing.T) {
		// Should get no results, since the bnd bge is 1 hour
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 1*time.Hour)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntResults := mbke(mbp[bpi.RepoID]time.Time)
		if diff := cmp.Diff(wbntResults, results); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestPermsStore_CountReposWithStblePerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Bbckground()
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupReposTbble(t, s)
	})

	// Set up some repositories bnd permissions.
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(1, 'privbte_repo_1', TRUE)`),                    // id=1
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(2, 'privbte_repo_2', TRUE)`),                    // id=2
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte, deleted_bt) VALUES(3, 'privbte_repo_3', TRUE, NOW())`), // id=3
	}
	for _, q := rbnge qs {
		executeQuery(t, ctx, s, q)
	}

	// Mbrk sync jobs bs completed for bll repos.
	repo1FinishedAt := clock().Add(-15 * time.Minute)
	repo2FinishedAt := clock().Add(-5 * time.Minute)
	repo3FinishedAt := clock().Add(-10 * time.Minute)
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 1, repo1FinishedAt, RebsonRepoOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 2, repo2FinishedAt, RebsonRepoOutdbtedPermissions)
	executeQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_bt, rebson) VALUES(%d, %s, %s)`, 3, repo3FinishedAt, RebsonRepoOutdbtedPermissions)
	executeQuery(t, ctx, s, q)

	t.Run("One result when bge is 10 minutes", func(t *testing.T) {
		// Should only get privbte_repo_1 bbck, becbuse bge is 10 minutes.
		count, err := s.CountReposWithStblePerms(ctx, 10*time.Minute)
		require.NoError(t, err)
		require.Equbl(t, 1, count)
	})

	t.Run("Both repos bre returned when bge is 1 minute", func(t *testing.T) {
		// Should get both privbte_repo_1 bnd privbte_repo_2, since the bge is 1 minute only.
		count, err := s.CountReposWithStblePerms(ctx, 1*time.Minute)
		require.NoError(t, err)
		require.Equbl(t, 2, count)
	})

	t.Run("Both repos bre returned when bge is 0 bnd deleted repos bre ignored", func(t *testing.T) {
		// Should get both privbte_repo_1 bnd privbte_repo_2, since the bge is 0 bnd cutoff clbuse if skipped.
		count, err := s.CountReposWithStblePerms(ctx, 0)
		require.NoError(t, err)
		require.Equbl(t, 2, count)
	})

	t.Run("Ignore repos thbt hbve synced recently", func(t *testing.T) {
		// Should get no results, since the bge is 1 hour
		count, err := s.CountReposWithStblePerms(ctx, 1*time.Hour)
		require.NoError(t, err)
		require.Equbl(t, 0, count)
	})
}

func TestPermsStore_MbpUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Bbckground()

	t.Clebnup(func() {
		if t.Fbiled() {
			return
		}

		q := `TRUNCATE TABLE externbl_services, orgs, users CASCADE`
		if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
			t.Fbtbl(err)
		}
	})

	// Set up 3 users
	users := db.Users()

	igor, err := users.Crebte(ctx,
		NewUser{
			Embil:           "igor@exbmple.com",
			Usernbme:        "igor",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	shrebh, err := users.Crebte(ctx,
		NewUser{
			Embil:           "shrebh@exbmple.com",
			Usernbme:        "shrebh",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)
	ombr, err := users.Crebte(ctx,
		NewUser{
			Embil:           "ombr@exbmple.com",
			Usernbme:        "ombr",
			EmbilIsVerified: true,
		},
	)
	require.NoError(t, err)

	// embils: mbp with b mixed lobd of existing, spbce only bnd non existing users
	hbs, err := s.MbpUsers(ctx, []string{"igor@exbmple.com", "", "ombr@exbmple.com", "  	", "sbybko@exbmple.com"}, &schemb.PermissionsUserMbpping{BindID: "embil"})
	bssert.NoError(t, err)
	bssert.Equbl(t, mbp[string]int32{
		"igor@exbmple.com": igor.ID,
		"ombr@exbmple.com": ombr.ID,
	}, hbs)

	// usernbmes: mbp with b mixed lobd of existing, spbce only bnd non existing users
	hbs, err = s.MbpUsers(ctx, []string{"igor", "", "shrebh", "  	", "cbrlos"}, &schemb.PermissionsUserMbpping{BindID: "usernbme"})
	bssert.NoError(t, err)
	bssert.Equbl(t, mbp[string]int32{
		"igor":   igor.ID,
		"shrebh": shrebh.ID,
	}, hbs)

	// use b non-existing mbpping
	_, err = s.MbpUsers(ctx, []string{"igor", "", "shrebh", "  	", "cbrlos"}, &schemb.PermissionsUserMbpping{BindID: "shoeSize"})
	bssert.Error(t, err)
}

func TestPermsStore_Metrics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)

	ctx := context.Bbckground()
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
		clebnupUsersTbble(t, s)
		if t.Fbiled() {
			return
		}

		if err := s.execute(ctx, sqlf.Sprintf(`DELETE FROM repo`)); err != nil {
			t.Fbtbl(err)
		}
	})

	// Set up repositories in vbrious stbtes (public/privbte, deleted/not, etc.)
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(1, 'privbte_repo_1', TRUE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(2, 'privbte_repo_2', TRUE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte, deleted_bt) VALUES(3, 'privbte_repo_3', TRUE, NOW())`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(4, 'public_repo_4', FALSE)`),
	}
	for _, q := rbnge qs {
		if err := s.execute(ctx, q); err != nil {
			t.Fbtbl(err)
		}
	}

	// Set up users in vbrious stbtes (deleted/not, etc.)
	qs = []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(id, usernbme) VALUES(1, 'user1')`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme) VALUES(2, 'user2')`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, deleted_bt) VALUES(3, 'user3', NOW())`),
	}
	for _, q := rbnge qs {
		if err := s.execute(ctx, q); err != nil {
			t.Fbtbl(err)
		}
	}

	// Set up permissions for the vbrious repos.
	ep := mbke([]buthz.Permission, 0, 4)
	for i := 1; i <= 4; i++ {
		for j := 1; j <= 4; j++ {
			ep = bppend(ep, buthz.Permission{
				UserID:            int32(j),
				ExternblAccountID: int32(j),
				RepoID:            int32(i),
			})
		}
	}
	setupPermsRelbtedEntities(t, s, []buthz.Permission{
		{UserID: 1, ExternblAccountID: 1},
		{UserID: 2, ExternblAccountID: 2},
		{UserID: 3, ExternblAccountID: 3},
		{UserID: 4, ExternblAccountID: 4},
	})
	_, err := s.setUserRepoPermissions(ctx, ep, buthz.PermissionEntity{}, buthz.SourceRepoSync, fblse)
	require.NoError(t, err)

	// Mock rows for testing
	qs = []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO permission_sync_jobs (finished_bt, user_id, rebson) VALUES(%s, 1, 'TEST')`, clock()),
		sqlf.Sprintf(`INSERT INTO permission_sync_jobs (finished_bt, user_id, rebson) VALUES(%s, 2, 'TEST')`, clock().Add(-1*time.Minute)),
		sqlf.Sprintf(`INSERT INTO permission_sync_jobs (finished_bt, user_id, rebson) VALUES(%s, 3, 'TEST')`, clock().Add(-2*time.Minute)), // Mebnt to be excluded becbuse it hbs been deleted
		sqlf.Sprintf(`INSERT INTO permission_sync_jobs (finished_bt, repository_id, rebson) VALUES(%s, 1, 'TEST')`, clock()),
		sqlf.Sprintf(`INSERT INTO permission_sync_jobs (finished_bt, repository_id, rebson) VALUES(%s, 2, 'TEST')`, clock().Add(-2*time.Minute)),
		sqlf.Sprintf(`INSERT INTO permission_sync_jobs (finished_bt, repository_id, rebson) VALUES(%s, 3, 'TEST')`, clock().Add(-3*time.Minute)), // Mebnt to be excluded becbuse it hbs been deleted
		sqlf.Sprintf(`INSERT INTO permission_sync_jobs (finished_bt, repository_id, rebson) VALUES(%s, 4, 'TEST')`, clock().Add(-3*time.Minute)), // Mebnt to be excluded becbuse it is public
	}
	for _, q := rbnge qs {
		if err := s.execute(ctx, q); err != nil {
			t.Fbtbl(err)
		}
	}

	m, err := s.Metrics(ctx, time.Minute)
	if err != nil {
		t.Fbtbl(err)
	}

	expMetrics := &PermsMetrics{
		UsersWithStblePerms:  1,
		UsersPermsGbpSeconds: 60,
		ReposWithStblePerms:  1,
		ReposPermsGbpSeconds: 120,
	}

	if diff := cmp.Diff(expMetrics, m); diff != "" {
		t.Fbtblf("mismbtch (-wbnt +got):\n%s", diff)
	}
}

func setupTestPerms(t *testing.T, db DB, clock func() time.Time) *permsStore {
	t.Helper()
	logger := logtest.Scoped(t)
	s := perms(logger, db, clock)
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)
	})
	return s
}

func TestPermsStore_ListUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Bbckground()

	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)

		if t.Fbiled() {
			return
		}
		q := `TRUNCATE TABLE externbl_services, repo, users CASCADE`
		if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
			t.Fbtbl(err)
		}
	})
	// Set fbke buthz providers otherwise buthz is bypbssed
	buthz.SetProviders(fblse, []buthz.Provider{&fbkePermsProvider{}})
	defer buthz.SetProviders(true, nil)
	// Set up some repositories bnd permissions
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, site_bdmin) VALUES(555, 'user555', FALSE)`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, site_bdmin) VALUES(777, 'user777', TRUE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(1, 'privbte_repo_1', TRUE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(2, 'privbte_repo_2', TRUE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte, deleted_bt) VALUES(3, 'privbte_repo_3_deleted', TRUE, NOW())`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(4, 'public_repo_4', FALSE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(5, 'public_repo_5', TRUE)`),
		sqlf.Sprintf(`INSERT INTO externbl_services(id, displby_nbme, kind, config) VALUES(1, 'GitHub #1', 'GITHUB', '{}')`),
		sqlf.Sprintf(`INSERT INTO externbl_service_repos(repo_id, externbl_service_id, clone_url)
                                 VALUES(1, 1, ''), (2, 1, ''), (3, 1, ''), (4, 1, '')`),
		sqlf.Sprintf(`INSERT INTO externbl_services(id, displby_nbme, kind, config, unrestricted) VALUES(2, 'GitHub #2 Unrestricted', 'GITHUB', '{}', TRUE)`),
		sqlf.Sprintf(`INSERT INTO externbl_service_repos(repo_id, externbl_service_id, clone_url)
                                 VALUES(5, 2, '')`),
	}
	for _, q := rbnge qs {
		if err := s.execute(ctx, q); err != nil {
			t.Fbtbl(err)
		}
	}
	q := sqlf.Sprintf(`INSERT INTO user_repo_permissions(user_id, repo_id, source) VALUES(555, 1, 'user_sync'), (777, 2, 'user_sync'), (555, 3, 'bpi'), (777, 3, 'bpi');`)
	if err := s.execute(ctx, q); err != nil {
		t.Fbtbl(err)
	}

	tests := []listUserPermissionsTest{
		{
			Nbme:   "TestNonSiteAdminUser",
			UserID: 555,
			WbntResults: []*listUserPermissionsResult{
				{
					// privbte repo but hbve bccess vib user_permissions
					RepoId: 1,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
				{
					// public repo
					RepoId: 4,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					// privbte repo but unrestricted
					RepoId: 5,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
		{
			Nbme:   "TestPbginbtion",
			UserID: 555,
			Args: &ListUserPermissionsArgs{
				PbginbtionArgs: &PbginbtionArgs{First: pointers.Ptr(2), After: pointers.Ptr("'public_repo_5'"), OrderBy: OrderBy{{Field: "nbme"}}},
			},
			WbntResults: []*listUserPermissionsResult{
				{
					RepoId: 1,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
				{
					RepoId: 4,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
		{
			Nbme:   "TestSebrchQuery",
			UserID: 555,
			Args: &ListUserPermissionsArgs{
				Query: "repo_5",
			},
			WbntResults: []*listUserPermissionsResult{
				{
					RepoId: 5,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
		{
			Nbme:   "TestSiteAdminUser",
			UserID: 777,
			WbntResults: []*listUserPermissionsResult{
				{
					// do not hbve direct bccess but user is site bdmin
					RepoId: 1,
					Rebson: UserRepoPermissionRebsonSiteAdmin,
				},
				{
					// privbte repo but hbve bccess vib user_permissions
					RepoId: 2,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
				{
					// public repo
					RepoId: 4,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					// privbte repo but unrestricted
					RepoId: 5,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.Nbme, func(t *testing.T) {
			results, err := s.ListUserPermissions(ctx, int32(test.UserID), test.Args)
			if err != nil {
				t.Fbtbl(err)
			}
			if len(test.WbntResults) != len(results) {
				t.Fbtblf("Results mismbtch. Wbnt: %d Got: %d", len(test.WbntResults), len(results))
			}
			for index, result := rbnge results {
				if diff := cmp.Diff(test.WbntResults[index], &listUserPermissionsResult{RepoId: int32(result.Repo.ID), Rebson: result.Rebson}); diff != "" {
					t.Fbtblf("Results (%d) mismbtch (-wbnt +got):\n%s", index, diff)
				}
			}
		})
	}
}

type listUserPermissionsTest struct {
	Nbme        string
	UserID      int
	Args        *ListUserPermissionsArgs
	WbntResults []*listUserPermissionsResult
}

type listUserPermissionsResult struct {
	RepoId int32
	Rebson UserRepoPermissionRebson
}

func TestPermsStore_ListRepoPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := NewDB(logger, testDb)

	s := perms(logtest.Scoped(t), db, clock)
	ctx := context.Bbckground()
	t.Clebnup(func() {
		clebnupPermsTbbles(t, s)

		if t.Fbiled() {
			return
		}

		q := `TRUNCATE TABLE externbl_services, repo, users CASCADE`
		if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
			t.Fbtbl(err)
		}
	})

	// Set up some repositories bnd permissions
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, site_bdmin) VALUES(555, 'user555', FALSE)`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, site_bdmin) VALUES(666, 'user666', FALSE)`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, site_bdmin) VALUES(777, 'user777', TRUE)`),
		sqlf.Sprintf(`INSERT INTO users(id, usernbme, site_bdmin, deleted_bt) VALUES(888, 'user888', TRUE, NOW())`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(1, 'privbte_repo_1', TRUE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(2, 'public_repo_2', FALSE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(3, 'unrestricted_repo_3', TRUE)`),
		sqlf.Sprintf(`INSERT INTO repo(id, nbme, privbte) VALUES(4, 'unrestricted_repo_4', TRUE)`),
		sqlf.Sprintf(`INSERT INTO externbl_services(id, displby_nbme, kind, config) VALUES(1, 'GitHub #1', 'GITHUB', '{}')`),
		sqlf.Sprintf(`INSERT INTO externbl_service_repos(repo_id, externbl_service_id, clone_url)
                                 VALUES(1, 1, ''), (2, 1, ''), (3, 1, '')`),
		sqlf.Sprintf(`INSERT INTO externbl_services(id, displby_nbme, kind, config, unrestricted) VALUES(2, 'GitHub #2 Unrestricted', 'GITHUB', '{}', TRUE)`),
		sqlf.Sprintf(`INSERT INTO externbl_service_repos(repo_id, externbl_service_id, clone_url)
                                 VALUES(4, 2, '')`),
	}

	for _, q := rbnge qs {
		if err := s.execute(ctx, q); err != nil {
			t.Fbtbl(err)
		}
	}

	q := sqlf.Sprintf(`INSERT INTO user_repo_permissions(user_id, repo_id) VALUES(555, 1), (666, 1), (NULL, 3), (666, 4)`)
	if err := s.execute(ctx, q); err != nil {
		t.Fbtbl(err)
	}

	tests := []listRepoPermissionsTest{
		{
			Nbme:   "TestPrivbteRepo",
			RepoID: 1,
			Args:   nil,
			WbntResults: []*listRepoPermissionsResult{
				{
					// do not hbve bccess but site-bdmin
					UserID: 777,
					Rebson: UserRepoPermissionRebsonSiteAdmin,
				},
				{
					// hbve bccess
					UserID: 666,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
				{
					// hbve bccess
					UserID: 555,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
			},
		},
		{
			Nbme:             "TestPrivbteRepoWithNoAuthzProviders",
			RepoID:           1,
			Args:             nil,
			NoAuthzProviders: true,
			// bll users hbve bccess
			WbntResults: []*listRepoPermissionsResult{
				{
					UserID: 777,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 666,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 555,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
		{
			Nbme:   "TestPbginbtionWithPrivbteRepo",
			RepoID: 1,
			Args: &ListRepoPermissionsArgs{
				PbginbtionArgs: &PbginbtionArgs{First: pointers.Ptr(1), After: pointers.Ptr("555"), OrderBy: OrderBy{{Field: "users.id"}}, Ascending: true},
			},
			WbntResults: []*listRepoPermissionsResult{
				{
					UserID: 666,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
			},
		},
		{
			Nbme:   "TestSebrchQueryWithPrivbteRepo",
			RepoID: 1,
			Args: &ListRepoPermissionsArgs{
				Query: "6",
			},
			WbntResults: []*listRepoPermissionsResult{
				{
					UserID: 666,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
			},
		},
		{
			Nbme:   "TestPublicRepo",
			RepoID: 2,
			Args:   nil,
			// bll users hbve bccess
			WbntResults: []*listRepoPermissionsResult{
				{
					UserID: 777,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 666,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 555,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
		{
			Nbme:   "TestUnrestrictedVibPermsTbbleRepo",
			RepoID: 3,
			Args:   nil,
			// bll users hbve bccess
			WbntResults: []*listRepoPermissionsResult{
				{
					UserID: 777,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 666,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 555,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
		{
			Nbme:   "TestUnrestrictedVibExternblServiceRepo",
			RepoID: 4,
			Args:   nil,
			// bll users hbve bccess
			WbntResults: []*listRepoPermissionsResult{
				{
					UserID: 777,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 666,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
				{
					UserID: 555,
					Rebson: UserRepoPermissionRebsonUnrestricted,
				},
			},
		},
		{
			Nbme:                      "TestUnrestrictedVibExternblServiceRepoWithoutPermsMbpping",
			RepoID:                    4,
			Args:                      nil,
			NoAuthzProviders:          true,
			UsePermissionsUserMbpping: true,
			// restricted bccess
			WbntResults: []*listRepoPermissionsResult{
				{
					// do not hbve bccess but site-bdmin
					UserID: 777,
					Rebson: UserRepoPermissionRebsonSiteAdmin,
				},
				{
					// hbve bccess
					UserID: 666,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
			},
		},
		{
			Nbme:                      "TestPrivbteRepoWithAuthzEnforceForSiteAdminsEnbbled",
			RepoID:                    1,
			Args:                      nil,
			AuthzEnforceForSiteAdmins: true,
			WbntResults: []*listRepoPermissionsResult{
				{
					// hbve bccess
					UserID: 666,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
				{
					// hbve bccess
					UserID: 555,
					Rebson: UserRepoPermissionRebsonPermissionsSync,
				},
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.Nbme, func(t *testing.T) {
			if !test.NoAuthzProviders {
				// Set fbke buthz providers otherwise buthz is bypbssed
				buthz.SetProviders(fblse, []buthz.Provider{&fbkePermsProvider{}})
				defer buthz.SetProviders(true, nil)
			}

			before := globbls.PermissionsUserMbpping()
			globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: test.UsePermissionsUserMbpping})
			conf.Mock(
				&conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						AuthzEnforceForSiteAdmins: test.AuthzEnforceForSiteAdmins,
					},
				},
			)
			t.Clebnup(func() {
				globbls.SetPermissionsUserMbpping(before)
				conf.Mock(nil)
			})

			results, err := s.ListRepoPermissions(ctx, bpi.RepoID(test.RepoID), test.Args)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(test.WbntResults) != len(results) {
				t.Fbtblf("Results mismbtch. Wbnt: %d Got: %d", len(test.WbntResults), len(results))
			}

			bctublResults := mbke([]*listRepoPermissionsResult, 0, len(results))
			for _, result := rbnge results {
				bctublResults = bppend(bctublResults, &listRepoPermissionsResult{UserID: result.User.ID, Rebson: result.Rebson})
			}

			if diff := cmp.Diff(test.WbntResults, bctublResults); diff != "" {
				t.Fbtblf("Results mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

type listRepoPermissionsTest struct {
	Nbme                      string
	RepoID                    int
	Args                      *ListRepoPermissionsArgs
	WbntResults               []*listRepoPermissionsResult
	NoAuthzProviders          bool
	UsePermissionsUserMbpping bool
	AuthzEnforceForSiteAdmins bool
}

type listRepoPermissionsResult struct {
	UserID int32
	Rebson UserRepoPermissionRebson
}

type fbkePermsProvider struct {
	codeHost *extsvc.CodeHost
	extAcct  *extsvc.Account
}

func (p *fbkePermsProvider) FetchAccount(context.Context, *types.User, []*extsvc.Account, []string) (mine *extsvc.Account, err error) {
	return p.extAcct, nil
}

func (p *fbkePermsProvider) ServiceType() string { return p.codeHost.ServiceType }
func (p *fbkePermsProvider) ServiceID() string   { return p.codeHost.ServiceID }
func (p *fbkePermsProvider) URN() string         { return extsvc.URN(p.codeHost.ServiceType, 0) }

func (p *fbkePermsProvider) VblidbteConnection(context.Context) error { return nil }

func (p *fbkePermsProvider) FetchUserPerms(context.Context, *extsvc.Account, buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	return nil, nil
}

func (p *fbkePermsProvider) FetchRepoPerms(context.Context, *extsvc.Repository, buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, nil
}

func equbl(t testing.TB, nbme string, wbnt, hbve bny) {
	t.Helper()
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtblf("%q: %s", nbme, diff)
	}
}
