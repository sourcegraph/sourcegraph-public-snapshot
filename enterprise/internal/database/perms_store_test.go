package database

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gitchander/permutation"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func cleanupPermsTables(t *testing.T, s *permsStore) {
	t.Helper()

	q := `TRUNCATE TABLE permission_sync_jobs, user_permissions, repo_permissions, user_pending_permissions, repo_pending_permissions, user_repo_permissions;`
	execQuery(t, context.Background(), s, sqlf.Sprintf(q))
}

func mapsetToArray(ms map[int32]struct{}) []int {
	ints := []int{}
	for id := range ms {
		ints = append(ints, int(id))
	}
	sort.Slice(ints, func(i, j int) bool { return ints[i] < ints[j] })

	return ints
}

func toMapset(ids ...int32) map[int32]struct{} {
	ms := map[int32]struct{}{}
	for _, id := range ids {
		ms[id] = struct{}{}
	}
	return ms
}

var now = timeutil.Now().UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now))
}

func TestPermsStore_LoadUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	ctx := context.Background()

	runTests := func(t *testing.T) {
		t.Helper()

		t.Run("no matching", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
				cleanupUsersTable(t, s)
				cleanupReposTable(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}
			setupPermsRelatedEntities(t, s, []authz.Permission{{UserID: 2, RepoID: 1}})

			if err := s.SetRepoPerms(ctx, 1, []authz.UserIDWithExternalAccountID{{UserID: 2}}); err != nil {
				t.Fatal(err)
			} else if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up, err := s.LoadUserPermissions(context.Background(), 1)
			require.NoError(t, err)

			equal(t, "IDs", 0, len(up))
		})

		t.Run("found matching", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
				cleanupUsersTable(t, s)
				cleanupReposTable(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}
			setupPermsRelatedEntities(t, s, []authz.Permission{{UserID: 2, RepoID: 1}})

			if err := s.SetRepoPerms(ctx, 1, []authz.UserIDWithExternalAccountID{{UserID: 2}}); err != nil {
				t.Fatal(err)
			} else if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up, err := s.LoadUserPermissions(context.Background(), 2)
			require.NoError(t, err)

			gotIDs := make([]int32, len(up))
			for i, perm := range up {
				gotIDs[i] = perm.RepoID
			}

			equal(t, "IDs", []int32{1}, gotIDs)
		})

		t.Run("add and change", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
				cleanupUsersTable(t, s)
				cleanupReposTable(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(1, 2),
			}
			setupPermsRelatedEntities(t, s, []authz.Permission{{UserID: 1, RepoID: 1}, {UserID: 2, RepoID: 1}, {UserID: 3, RepoID: 1}})

			if err := s.SetRepoPerms(ctx, 1, []authz.UserIDWithExternalAccountID{{UserID: 1}, {UserID: 2}}); err != nil {
				t.Fatal(err)
			} else if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2, 3),
			}
			if err := s.SetRepoPerms(ctx, 1, []authz.UserIDWithExternalAccountID{{UserID: 2}, {UserID: 3}}); err != nil {
				t.Fatal(err)
			} else if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up1, err := s.LoadUserPermissions(context.Background(), 1)
			require.NoError(t, err)

			equal(t, "No IDs", 0, len(up1))

			up2, err := s.LoadUserPermissions(context.Background(), 2)
			require.NoError(t, err)
			gotIDs := make([]int32, len(up2))
			for i, perm := range up2 {
				gotIDs[i] = perm.RepoID
			}

			equal(t, "IDs", []int32{1}, gotIDs)

			up3, err := s.LoadUserPermissions(context.Background(), 3)
			require.NoError(t, err)
			gotIDs = make([]int32, len(up3))
			for i, perm := range up3 {
				gotIDs[i] = perm.RepoID
			}

			equal(t, "IDs", []int32{1}, gotIDs)
		})
	}

	t.Run("With legacy perms tables", func(t *testing.T) {
		mockUnifiedPermsConfig(false)

		t.Cleanup(func() {
			conf.Mock(nil)
		})

		runTests(t)
	})

	t.Run("With unified perms table", func(t *testing.T) {
		mockUnifiedPermsConfig(true)

		t.Cleanup(func() {
			conf.Mock(nil)
		})

		runTests(t)
	})
}

func TestPermsStore_LoadRepoPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	ctx := context.Background()

	runTests := func(t *testing.T) {
		t.Helper()
		t.Run("no matching", func(t *testing.T) {
			s := perms(logger, db, time.Now)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
				cleanupUsersTable(t, s)
				cleanupReposTable(t, s)
			})

			setupPermsRelatedEntities(t, s, []authz.Permission{{UserID: 2, RepoID: 1}})

			up := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			}
			if err := s.SetRepoPerms(ctx, 1, []authz.UserIDWithExternalAccountID{{UserID: 2}}); err != nil {
				t.Fatal(err)
			} else if _, err := s.SetUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}

			rp, err := s.LoadRepoPermissions(context.Background(), 2)
			require.NoError(t, err)
			require.Equal(t, 0, len(rp))
		})

		t.Run("found matching", func(t *testing.T) {
			s := perms(logger, db, time.Now)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
				cleanupUsersTable(t, s)
				cleanupReposTable(t, s)
			})

			setupPermsRelatedEntities(t, s, []authz.Permission{{UserID: 2, RepoID: 1}})

			up := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			}
			if err := s.SetRepoPerms(ctx, 1, []authz.UserIDWithExternalAccountID{{UserID: 2}}); err != nil {
				t.Fatal(err)
			} else if _, err := s.SetUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}

			rp, err := s.LoadRepoPermissions(context.Background(), 1)
			require.NoError(t, err)
			gotIDs := make([]int32, len(rp))
			for i, perm := range rp {
				gotIDs[i] = perm.UserID
			}

			equal(t, "permissions UserIDs", []int32{2}, gotIDs)
		})
	}

	t.Run("With legacy perms tables", func(t *testing.T) {
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		mockUnifiedPermsConfig(false)

		runTests(t)
	})

	t.Run("With unified perms tables", func(t *testing.T) {
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		mockUnifiedPermsConfig(true)

		runTests(t)
	})
}

func testPermsStore_FetchReposByUserAndExternalService(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		t.Run("found matching", func(t *testing.T) {
			logger := logtest.Scoped(t)
			ctx := context.Background()
			s := perms(logger, db, clock)
			if _, err := db.ExecContext(ctx, `INSERT into repo (name, external_service_type, external_service_id) values ('github.com/test/test', 'github', 'https://github.com/')`); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				cleanupReposTable(t, s)
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}

			if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			repos, err := s.FetchReposByUserAndExternalService(ctx, 2, "github", "https://github.com/")
			if err != nil {
				t.Fatal(err)
			}
			equal(t, "repos", []api.RepoID{1}, repos)
		})
		t.Run("skips non matching", func(t *testing.T) {
			ctx := context.Background()
			s := perms(logger, db, clock)
			if _, err := db.ExecContext(ctx, `INSERT into repo (name, external_service_type, external_service_id) values ('github.com/test/test', 'github', 'https://github.com/')`); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				cleanupReposTable(t, s)
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}
			if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			repos, err := s.FetchReposByUserAndExternalService(ctx, 2, "gitlab", "https://gitlab.com/")
			if err != nil {
				t.Fatal(err)
			}
			equal(t, "repos", 0, len(repos))
		})
	}
}

func checkRegularPermsTable(s *permsStore, sql string, expects map[int32][]uint32) error {
	rows, err := s.Handle().QueryContext(context.Background(), sql)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id int32
		var ids []int64
		if err = rows.Scan(&id, pq.Array(&ids)); err != nil {
			return err
		}

		intIDs := make([]uint32, 0, len(ids))
		for _, id := range ids {
			intIDs = append(intIDs, uint32(id))
		}

		if expects[id] == nil {
			return errors.Errorf("unexpected row in table: (id: %v) -> (ids: %v)", id, intIDs)
		}

		comparator := func(a, b uint32) bool {
			return a < b
		}

		if cmp.Diff(expects[id], intIDs, cmpopts.SortSlices(comparator)) != "" {
			return errors.Errorf("intIDs - key %v: want %d but got %d", id, expects[id], intIDs)
		}

		delete(expects, id)
	}

	if err = rows.Close(); err != nil {
		return err
	}

	if len(expects) > 0 {
		return errors.Errorf("missing rows from table: %v", expects)
	}

	return nil
}

func TestPermsStore_SetUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)

	const countToExceedParameterLimit = 17000 // ~ 65535 / 4 parameters per row

	tests := []struct {
		name            string
		slowTest        bool
		updates         []*authz.UserPermissions
		expectUserPerms map[int32][]uint32 // user_id -> object_ids
		expectRepoPerms map[int32][]uint32 // repo_id -> user_ids
		expectedResult  []*database.SetPermissionsResult

		upsertRepoPermissionsPageSize int
	}{
		{
			name: "empty",
			updates: []*authz.UserPermissions{
				{
					UserID: 1,
					Perm:   authz.Read,
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {},
			},
			expectedResult: []*database.SetPermissionsResult{{
				Added:   0,
				Removed: 0,
				Found:   0,
			}},
		},
		{
			name: "add",
			updates: []*authz.UserPermissions{
				{
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toMapset(1),
				}, {
					UserID: 2,
					Perm:   authz.Read,
					IDs:    toMapset(1, 2),
				}, {
					UserID: 3,
					Perm:   authz.Read,
					IDs:    toMapset(3, 4),
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1},
				2: {1, 2},
				3: {3, 4},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1, 2},
				2: {2},
				3: {3},
				4: {3},
			},
			expectedResult: []*database.SetPermissionsResult{
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
			name: "add and update",
			updates: []*authz.UserPermissions{
				{
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toMapset(1),
				}, {
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toMapset(2, 3),
				}, {
					UserID: 2,
					Perm:   authz.Read,
					IDs:    toMapset(1, 2),
				}, {
					UserID: 2,
					Perm:   authz.Read,
					IDs:    toMapset(1, 3),
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {2, 3},
				2: {1, 3},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {2},
				2: {1},
				3: {1, 2},
			},
			expectedResult: []*database.SetPermissionsResult{
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
					Added:   1,
					Removed: 1,
					Found:   2,
				},
			},
		},
		{
			name: "add and clear",
			updates: []*authz.UserPermissions{
				{
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toMapset(1, 2, 3),
				}, {
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toMapset(),
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {},
				2: {},
				3: {},
			},
			expectedResult: []*database.SetPermissionsResult{
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
		{
			name:                          "add and page",
			upsertRepoPermissionsPageSize: 2,
			updates: []*authz.UserPermissions{
				{
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toMapset(1, 2, 3),
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1, 2, 3},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1},
				2: {1},
				3: {1},
			},
			expectedResult: []*database.SetPermissionsResult{
				{
					Added:   3,
					Removed: 0,
					Found:   3,
				},
			},
		},
		{
			name:     postgresParameterLimitTest,
			slowTest: true,
			updates: func() []*authz.UserPermissions {
				user := &authz.UserPermissions{
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toMapset(),
				}
				for i := 1; i <= countToExceedParameterLimit; i += 1 {
					user.IDs[int32(i)] = struct{}{}
				}
				return []*authz.UserPermissions{user}
			}(),
			expectUserPerms: func() map[int32][]uint32 {
				repos := make([]uint32, countToExceedParameterLimit)
				for i := 1; i <= countToExceedParameterLimit; i += 1 {
					repos[i-1] = uint32(i)
				}
				return map[int32][]uint32{1: repos}
			}(),
			expectRepoPerms: func() map[int32][]uint32 {
				repos := make(map[int32][]uint32, countToExceedParameterLimit)
				for i := 1; i <= countToExceedParameterLimit; i += 1 {
					repos[int32(i)] = []uint32{1}
				}
				return repos
			}(),
		},
	}

	t.Run("user-centric update should set permissions", func(t *testing.T) {
		logger := logtest.Scoped(t)
		s := perms(logger, db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		up := &authz.UserPermissions{
			UserID: 2,
			Perm:   authz.Read,
			Type:   authz.PermRepos,
			IDs:    toMapset(1),
		}
		if _, err := s.SetUserPermissions(context.Background(), up); err != nil {
			t.Fatal(err)
		}

		p, err := s.LoadUserPermissions(context.Background(), up.UserID)
		require.NoError(t, err)
		gotIDs := make([]int32, len(p))
		for i, perm := range p {
			gotIDs[i] = perm.RepoID
		}

		equal(t, "up.IDs", []int32{1}, gotIDs)
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.slowTest && !*slowTests {
				t.Skip("slow-tests not enabled")
			}

			if test.upsertRepoPermissionsPageSize > 0 {
				upsertRepoPermissionsPageSize = test.upsertRepoPermissionsPageSize
			}

			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
				if test.upsertRepoPermissionsPageSize > 0 {
					upsertRepoPermissionsPageSize = defaultUpsertRepoPermissionsPageSize
				}
			})

			for index, p := range test.updates {
				tmp := &authz.UserPermissions{
					UserID:    p.UserID,
					Perm:      p.Perm,
					UpdatedAt: p.UpdatedAt,
				}
				if p.IDs != nil {
					tmp.IDs = p.IDs
				}
				result, err := s.SetUserPermissions(context.Background(), tmp)

				if diff := cmp.Diff(test.expectedResult[index], result); diff != "" {
					t.Fatal(diff)
				}

				if err != nil {
					t.Fatal(err)
				}
			}

			err := checkRegularPermsTable(s, `SELECT user_id, object_ids_ints FROM user_permissions`, test.expectUserPerms)
			if err != nil {
				t.Fatal("user_permissions:", err)
			}

			err = checkRegularPermsTable(s, `SELECT repo_id, user_ids_ints FROM repo_permissions`, test.expectRepoPerms)
			if err != nil {
				t.Fatal("repo_permissions:", err)
			}
		})
	}
}

func checkUserRepoPermissions(t *testing.T, s *permsStore, where *sqlf.Query, expectedPermissions []authz.Permission) {
	t.Helper()
	format := "SELECT user_id, user_external_account_id, repo_id, created_at, updated_at, source FROM user_repo_permissions WHERE %s;"
	permissions, err := ScanPermissions(s.Query(context.Background(), sqlf.Sprintf(format, where)))
	if err != nil {
		t.Fatal(err)
	}
	// ScanPermissions returns nil if there are no results, but for the purpose of test readability,
	// we defined expectedPermissions to be an empty slice, which matches the empty permissions input to write to the db.
	// hence if permissions is nil, we set it to an empty slice.
	if permissions == nil {
		permissions = []authz.Permission{}
	}
	sort.Slice(permissions, func(i, j int) bool {
		if permissions[i].UserID == permissions[j].UserID && permissions[i].ExternalAccountID == permissions[j].ExternalAccountID {
			return permissions[i].RepoID < permissions[j].RepoID
		}
		if permissions[i].UserID == permissions[j].UserID {
			return permissions[i].ExternalAccountID < permissions[j].ExternalAccountID
		}
		return permissions[i].UserID < permissions[j].UserID
	})

	if diff := cmp.Diff(expectedPermissions, permissions, cmpopts.IgnoreFields(authz.Permission{}, "CreatedAt", "UpdatedAt", "Source")); diff != "" {
		t.Fatalf("Expected permissions: %v do not match actual permissions: %v; diff %v", expectedPermissions, permissions, diff)
	}
}

func setupPermsRelatedEntities(t *testing.T, s *permsStore, permissions []authz.Permission) {
	t.Helper()
	if permissions == nil || len(permissions) == 0 {
		t.Fatal("no permissions to setup related entities for")
	}

	users := make(map[int32]*sqlf.Query, len(permissions))
	externalAccounts := make(map[int32]*sqlf.Query, len(permissions))
	repos := make(map[int32]*sqlf.Query, len(permissions))
	for _, p := range permissions {
		users[p.UserID] = sqlf.Sprintf("(%s::integer, %s::text)", p.UserID, fmt.Sprintf("user-%d", p.UserID))
		externalAccounts[p.ExternalAccountID] = sqlf.Sprintf("(%s::integer, %s::integer, %s::text, %s::text, %s::text, %s::text)", p.ExternalAccountID, p.UserID, "service_type", "service_id", fmt.Sprintf("account_id_%d", p.ExternalAccountID), "client_id")
		repos[p.RepoID] = sqlf.Sprintf("(%s::integer, %s::text)", p.RepoID, fmt.Sprintf("repo-%d", p.RepoID))
	}

	defaultErrMessage := "setup test related entities before actual test"
	usersQuery := sqlf.Sprintf(`INSERT INTO users(id, username) VALUES %s ON CONFLICT (id) DO NOTHING`, sqlf.Join(maps.Values(users), ","))
	if err := s.execute(context.Background(), usersQuery); err != nil {
		t.Fatal(defaultErrMessage, err)
	}
	externalAccountsQuery := sqlf.Sprintf(`INSERT INTO user_external_accounts(id, user_id, service_type, service_id, account_id, client_id) VALUES %s ON CONFLICT(id) DO NOTHING`, sqlf.Join(maps.Values(externalAccounts), ","))
	if err := s.execute(context.Background(), externalAccountsQuery); err != nil {
		t.Fatal(defaultErrMessage, err)
	}
	reposQuery := sqlf.Sprintf(`INSERT INTO repo(id, name) VALUES %s ON CONFLICT(id) DO NOTHING`, sqlf.Join(maps.Values(repos), ","))
	if err := s.execute(context.Background(), reposQuery); err != nil {
		t.Fatal(defaultErrMessage, err)
	}
}

func testPermsStore_SetUserRepoPermissions(db database.DB) func(*testing.T) {
	source := "test"

	tests := []struct {
		name                string
		origPermissions     []authz.Permission
		permissions         []authz.Permission
		expectedPermissions []authz.Permission
		entity              authz.PermissionEntity
	}{
		{
			name:                "empty",
			permissions:         []authz.Permission{},
			expectedPermissions: []authz.Permission{},
			entity:              authz.PermissionEntity{UserID: 1, ExternalAccountID: 1},
		},
		{
			name: "add",
			permissions: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 1, ExternalAccountID: 1, RepoID: 3},
			},
			expectedPermissions: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 1, ExternalAccountID: 1, RepoID: 3},
			},
			entity: authz.PermissionEntity{UserID: 1, ExternalAccountID: 1},
		},
		{
			name: "add, update and remove",
			origPermissions: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 1, ExternalAccountID: 1, RepoID: 3},
			},
			permissions: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 4},
			},
			expectedPermissions: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 4},
			},
			entity: authz.PermissionEntity{UserID: 1, ExternalAccountID: 1},
		},
		{
			name: "remove only",
			origPermissions: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 1, ExternalAccountID: 1, RepoID: 3},
			},
			permissions:         []authz.Permission{},
			expectedPermissions: []authz.Permission{},
			entity:              authz.PermissionEntity{UserID: 1, ExternalAccountID: 1},
		},
	}

	return func(t *testing.T) {
		ctx := actor.WithInternalActor(context.Background())
		logger := logtest.Scoped(t)

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := perms(logger, db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
					cleanupUsersTable(t, s)
					cleanupReposTable(t, s)
				})

				if len(test.origPermissions) > 0 {
					setupPermsRelatedEntities(t, s, test.origPermissions)
					err := s.setUserRepoPermissions(ctx, test.origPermissions, test.entity, source)
					if err != nil {
						t.Fatal("setup test permissions before actual test", err)
					}
				}

				if len(test.permissions) > 0 {
					setupPermsRelatedEntities(t, s, test.permissions)
				}
				if err := s.setUserRepoPermissions(ctx, test.permissions, test.entity, source); err != nil {
					t.Fatal("testing user repo permissions", err)
				}

				if test.entity.UserID > 0 {
					checkUserRepoPermissions(t, s, sqlf.Sprintf("user_id = %d", test.entity.UserID), test.expectedPermissions)
				} else if test.entity.RepoID > 0 {
					checkUserRepoPermissions(t, s, sqlf.Sprintf("repo_id = %d", test.entity.RepoID), test.expectedPermissions)
				}
			})
		}
	}
}

func testPermsStore_FetchReposByExternalAccount(db database.DB) func(*testing.T) {
	source := "test"

	tests := []struct {
		name              string
		origPermissions   []authz.Permission
		expected          []api.RepoID
		externalAccountID int32
	}{
		{
			name:              "empty",
			externalAccountID: 1,
			expected:          nil,
		},
		{
			name:              "one match",
			externalAccountID: 1,
			expected:          []api.RepoID{1},
			origPermissions: []authz.Permission{
				{
					UserID:            1,
					ExternalAccountID: 1,
					RepoID:            1,
				},
				{
					UserID:            1,
					ExternalAccountID: 2,
					RepoID:            2,
				},
				{
					UserID: 1,
					RepoID: 3,
				},
			},
		},
		{
			name:              "multiple matches",
			externalAccountID: 1,
			expected:          []api.RepoID{1, 2},
			origPermissions: []authz.Permission{
				{
					UserID:            1,
					ExternalAccountID: 1,
					RepoID:            1,
				},
				{
					UserID:            1,
					ExternalAccountID: 1,
					RepoID:            2,
				},
				{
					UserID: 1,
					RepoID: 3,
				},
			},
		},
	}

	return func(t *testing.T) {
		ctx := actor.WithInternalActor(context.Background())
		logger := logtest.Scoped(t)

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := perms(logger, db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
				})

				if test.origPermissions != nil && len(test.origPermissions) > 0 {
					setupPermsRelatedEntities(t, s, test.origPermissions)
					err := s.setUserRepoPermissions(ctx, test.origPermissions, authz.PermissionEntity{UserID: 42}, source)
					if err != nil {
						t.Fatal("setup test permissions before actual test", err)
					}
				}

				ids, err := s.FetchReposByExternalAccount(ctx, test.externalAccountID)
				if err != nil {
					t.Fatal("testing fetch repos by user and external account", err)
				}

				assert.Equal(t, test.expected, ids, "no match found for repo IDs")
			})
		}
	}
}

func TestPermsStore_SetRepoPermissionsUnrestricted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)

	ctx := context.Background()
	s := setupTestPerms(t, db, clock)

	legacyUnrestricted := func(t *testing.T, id int32, want bool) {
		t.Helper()

		p, err := s.LoadRepoPermissions(ctx, id)
		require.NoErrorf(t, err, "loading permissions for %d", id)

		unrestricted := (len(p) == 1 && p[0].UserID == 0)

		if unrestricted != want {
			t.Fatalf("Want %v, got %v for %d", want, unrestricted, id)
		}
	}

	assertUnrestricted := func(t *testing.T, id int32, want bool) {
		t.Helper()
		legacyUnrestricted(t, id, want)

		q := sqlf.Sprintf("SELECT repo_id FROM user_repo_permissions WHERE repo_id = %d AND user_id IS NULL", id)
		results, err := basestore.ScanInt32s(s.Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
		if err != nil {
			t.Fatalf("loading user repo permissions for %d: %v", id, err)
		}
		if want && len(results) == 0 {
			t.Fatalf("Want unrestricted, but found no results for %d", id)
		}
		if !want && len(results) > 0 {
			t.Fatalf("Want restricted, but found results for %d: %v", id, results)
		}
	}

	createRepo := func(t *testing.T, id int) {
		t.Helper()
		execQuery(t, ctx, s, sqlf.Sprintf(`
		INSERT INTO repo (id, name, private)
		VALUES (%d, %s, TRUE)`, id, fmt.Sprintf("repo-%d", id)))
	}

	// Add a couple of repos and a user
	execQuery(t, ctx, s, sqlf.Sprintf(`INSERT INTO users (username) VALUES ('alice')`))
	execQuery(t, ctx, s, sqlf.Sprintf(`INSERT INTO users (username) VALUES ('bob')`))
	for i := 0; i < 2; i++ {
		createRepo(t, i+1)
		rp := &authz.RepoPermissions{
			RepoID:  int32(i + 1),
			Perm:    authz.Read,
			UserIDs: toMapset(2),
		}
		if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
			t.Fatal(err)
		}
		if err := s.SetRepoPerms(context.Background(), int32(i+1), []authz.UserIDWithExternalAccountID{{UserID: 2}}); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Both repos are restricted by default", func(t *testing.T) {
		assertUnrestricted(t, 1, false)
		assertUnrestricted(t, 2, false)
	})

	t.Run("Set both repos to unrestricted", func(t *testing.T) {
		if err := s.SetRepoPermissionsUnrestricted(ctx, []int32{1, 2}, true); err != nil {
			t.Fatal(err)
		}
		assertUnrestricted(t, 1, true)
		assertUnrestricted(t, 2, true)
	})

	t.Run("Set unrestricted on a repo not in permissions table", func(t *testing.T) {
		createRepo(t, 3)
		if err := s.SetRepoPermissionsUnrestricted(ctx, []int32{3}, true); err != nil {
			t.Fatal(err)
		}

		assertUnrestricted(t, 1, true)
		assertUnrestricted(t, 2, true)
		assertUnrestricted(t, 3, true)
	})

	t.Run("Unset restricted on a repo in and not in permissions table", func(t *testing.T) {
		createRepo(t, 4)
		if err := s.SetRepoPermissionsUnrestricted(ctx, []int32{2, 3, 4}, false); err != nil {
			t.Fatal(err)
		}
		assertUnrestricted(t, 1, true)
		assertUnrestricted(t, 2, false)
		assertUnrestricted(t, 3, false)
		assertUnrestricted(t, 4, false)
	})

	t.Run("Set repos back to restricted again", func(t *testing.T) {
		// Also checking that more than 65535 IDs can be processed without an error
		var ids [66000]int32
		for i := range ids {
			ids[i] = int32(i + 1)
		}
		if err := s.SetRepoPermissionsUnrestricted(ctx, ids[:], false); err != nil {
			t.Fatal(err)
		}
		assertUnrestricted(t, 1, false)
		assertUnrestricted(t, 500, false)
		assertUnrestricted(t, 66000, false)
	})
}

func testPermsStore_SetRepoPermissions(db database.DB) func(*testing.T) {
	tests := []struct {
		name            string
		updates         []*authz.RepoPermissions
		expectUserPerms map[int32][]uint32 // user_id -> object_ids
		expectRepoPerms map[int32][]uint32 // repo_id -> user_ids
		expectedResult  []*database.SetPermissionsResult
	}{
		{
			name: "empty",
			updates: []*authz.RepoPermissions{
				{
					RepoID: 1,
					Perm:   authz.Read,
				},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {},
			},
			expectedResult: []*database.SetPermissionsResult{
				{
					Added:   0,
					Removed: 0,
					Found:   0,
				},
			},
		},
		{
			name: "add",
			updates: []*authz.RepoPermissions{
				{
					RepoID:  1,
					Perm:    authz.Read,
					UserIDs: toMapset(1),
				}, {
					RepoID:  2,
					Perm:    authz.Read,
					UserIDs: toMapset(1, 2),
				}, {
					RepoID:  3,
					Perm:    authz.Read,
					UserIDs: toMapset(3, 4),
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1, 2},
				2: {2},
				3: {3},
				4: {3},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1},
				2: {1, 2},
				3: {3, 4},
			},
			expectedResult: []*database.SetPermissionsResult{
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
			name: "add and update",
			updates: []*authz.RepoPermissions{
				{
					RepoID:  1,
					Perm:    authz.Read,
					UserIDs: toMapset(1),
				}, {
					RepoID:  1,
					Perm:    authz.Read,
					UserIDs: toMapset(2, 3),
				}, {
					RepoID:  2,
					Perm:    authz.Read,
					UserIDs: toMapset(1, 2),
				}, {
					RepoID:  2,
					Perm:    authz.Read,
					UserIDs: toMapset(3, 4),
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {},
				2: {1},
				3: {1, 2},
				4: {2},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {2, 3},
				2: {3, 4},
			},
			expectedResult: []*database.SetPermissionsResult{
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
			name: "add and clear",
			updates: []*authz.RepoPermissions{
				{
					RepoID:  1,
					Perm:    authz.Read,
					UserIDs: toMapset(1, 2, 3),
				}, {
					RepoID:  1,
					Perm:    authz.Read,
					UserIDs: toMapset(),
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {},
				2: {},
				3: {},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {},
			},
			expectedResult: []*database.SetPermissionsResult{
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

	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		t.Run("repo-centric update should set synced_at", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}
			if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			perms, err := s.LoadRepoPermissions(context.Background(), 1)
			require.NoError(t, err)
			gotIDs := make([]int32, len(perms))
			for i, perm := range perms {
				gotIDs[i] = perm.RepoID
			}

			equal(t, "rp.UserIDs", []int{2}, gotIDs)
		})

		t.Run("unrestricted columns should be set", func(t *testing.T) {
			// TOOD: Use this in other tests
			s := setupTestPerms(t, db, clock)

			rp := &authz.RepoPermissions{
				RepoID:       1,
				Perm:         authz.Read,
				UserIDs:      toMapset(2),
				Unrestricted: true,
			}
			if _, err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			perms, err := s.LoadRepoPermissions(context.Background(), 1)
			require.NoError(t, err)

			if len(perms) != 1 || perms[0].UserID != 0 {
				t.Fatal("Want unrestricted, got %v", perms)
			}
		})

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := perms(logger, db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
				})

				for index, p := range test.updates {
					tmp := &authz.RepoPermissions{
						RepoID:    p.RepoID,
						Perm:      p.Perm,
						UpdatedAt: p.UpdatedAt,
					}
					if p.UserIDs != nil {
						tmp.UserIDs = p.UserIDs
					}
					result, err := s.SetRepoPermissions(context.Background(), tmp)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(test.expectedResult[index], result); diff != "" {
						t.Fatal(diff)
					}
				}

				err := checkRegularPermsTable(s, `SELECT user_id, object_ids_ints FROM user_permissions`, test.expectUserPerms)
				if err != nil {
					t.Fatal("user_permissions:", err)
				}

				err = checkRegularPermsTable(s, `SELECT repo_id, user_ids_ints FROM repo_permissions`, test.expectRepoPerms)
				if err != nil {
					t.Fatal("repo_permissions:", err)
				}
			})
		}
	}
}

func testPermsStore_LoadUserPendingPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		t.Run("no matching with different account ID", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.Accounts{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				AccountIDs:  []string{"bob"},
			}
			rp := &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), accounts, rp); err != nil {
				t.Fatal(err)
			}

			alice := &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      "alice",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}
			err := s.LoadUserPendingPermissions(context.Background(), alice)
			if err != authz.ErrPermsNotFound {
				t.Fatalf("err: want %q but got %q", authz.ErrPermsNotFound, err)
			}
			equal(t, "IDs", 0, len(mapsetToArray(alice.IDs)))
		})

		t.Run("no matching with different service ID", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.Accounts{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				AccountIDs:  []string{"alice"},
			}
			rp := &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), accounts, rp); err != nil {
				t.Fatal(err)
			}

			alice := &authz.UserPendingPermissions{
				ServiceType: extsvc.TypeGitLab,
				ServiceID:   "https://gitlab.com/",
				BindID:      "alice",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}
			err := s.LoadUserPendingPermissions(context.Background(), alice)
			if err != authz.ErrPermsNotFound {
				t.Fatalf("err: want %q but got %q", authz.ErrPermsNotFound, err)
			}
			equal(t, "IDs", 0, len(mapsetToArray(alice.IDs)))
		})

		t.Run("found matching", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.Accounts{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				AccountIDs:  []string{"alice"},
			}
			rp := &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), accounts, rp); err != nil {
				t.Fatal(err)
			}

			alice := &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      "alice",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), alice); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, mapsetToArray(alice.IDs))
			equal(t, "UpdatedAt", now, alice.UpdatedAt.UnixNano())
		})

		t.Run("add and change", func(t *testing.T) {
			s := perms(logger, db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.Accounts{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				AccountIDs:  []string{"alice", "bob"},
			}
			rp := &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), accounts, rp); err != nil {
				t.Fatal(err)
			}

			accounts.AccountIDs = []string{"bob", "cindy"}
			rp = &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), accounts, rp); err != nil {
				t.Fatal(err)
			}

			alice := &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      "alice",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), alice); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", 0, len(mapsetToArray(alice.IDs)))

			bob := &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      "bob",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), bob); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, mapsetToArray(bob.IDs))
			equal(t, "UpdatedAt", now, bob.UpdatedAt.UnixNano())

			cindy := &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      "cindy",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), cindy); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, mapsetToArray(cindy.IDs))
			equal(t, "UpdatedAt", now, cindy.UpdatedAt.UnixNano())
		})
	}
}

func checkUserPendingPermsTable(
	ctx context.Context,
	s *permsStore,
	expects map[extsvc.AccountSpec][]uint32,
) (
	idToSpecs map[int32]extsvc.AccountSpec,
	err error,
) {
	q := `SELECT id, service_type, service_id, bind_id, object_ids_ints FROM user_pending_permissions`
	rows, err := s.Handle().QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	// Collect id -> account mappings for later used by checkRepoPendingPermsTable.
	idToSpecs = make(map[int32]extsvc.AccountSpec)
	for rows.Next() {
		var id int32
		var spec extsvc.AccountSpec
		var ids []int64
		if err := rows.Scan(&id, &spec.ServiceType, &spec.ServiceID, &spec.AccountID, pq.Array(&ids)); err != nil {
			return nil, err
		}
		idToSpecs[id] = spec

		intIDs := make([]uint32, 0, len(ids))
		for _, id := range ids {
			intIDs = append(intIDs, uint32(id))
		}

		if expects[spec] == nil {
			return nil, errors.Errorf("unexpected row in table: (spec: %v) -> (ids: %v)", spec, intIDs)
		}
		want := fmt.Sprintf("%v", expects[spec])

		have := fmt.Sprintf("%v", intIDs)
		if have != want {
			return nil, errors.Errorf("intIDs - spec %q: want %q but got %q", spec, want, have)
		}
		delete(expects, spec)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if len(expects) > 0 {
		return nil, errors.Errorf("missing rows from table: %v", expects)
	}

	return idToSpecs, nil
}

func checkRepoPendingPermsTable(
	ctx context.Context,
	s *permsStore,
	idToSpecs map[int32]extsvc.AccountSpec,
	expects map[int32][]extsvc.AccountSpec,
) error {
	rows, err := s.Handle().QueryContext(ctx, `SELECT repo_id, user_ids_ints FROM repo_pending_permissions`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id int32
		var ids []int64
		if err := rows.Scan(&id, pq.Array(&ids)); err != nil {
			return err
		}

		intIDs := make([]int, 0, len(ids))
		for _, id := range ids {
			intIDs = append(intIDs, int(id))
		}

		if expects[id] == nil {
			return errors.Errorf("unexpected row in table: (id: %v) -> (ids: %v)", id, intIDs)
		}

		haveSpecs := make([]extsvc.AccountSpec, 0, len(intIDs))
		for _, userID := range intIDs {
			spec, ok := idToSpecs[int32(userID)]
			if !ok {
				continue
			}

			haveSpecs = append(haveSpecs, spec)
		}
		wantSpecs := expects[id]

		// Verify Specs are the same, the ordering might not be the same but the elements/length are.
		if len(wantSpecs) != len(haveSpecs) {
			return errors.Errorf("initIDs - id %d: want %q but got %q", id, wantSpecs, haveSpecs)
		}
		wantSpecsSet := map[extsvc.AccountSpec]struct{}{}
		for _, spec := range wantSpecs {
			wantSpecsSet[spec] = struct{}{}
		}

		for _, spec := range haveSpecs {
			if _, ok := wantSpecsSet[spec]; !ok {
				return errors.Errorf("initIDs - id %d: want %q but got %q", id, wantSpecs, haveSpecs)
			}
		}

		delete(expects, id)
	}

	if err = rows.Close(); err != nil {
		return err
	}

	if len(expects) > 0 {
		return errors.Errorf("missing rows from table: %v", expects)
	}

	return nil
}

func testPermsStore_SetRepoPendingPermissions(db database.DB) func(*testing.T) {
	alice := extsvc.AccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "alice",
	}
	bob := extsvc.AccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "bob",
	}
	cindy := extsvc.AccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "cindy",
	}
	cindyGitHub := extsvc.AccountSpec{
		ServiceType: "github",
		ServiceID:   "https://github.com/",
		AccountID:   "cindy",
	}
	const countToExceedParameterLimit = 11000 // ~ 65535 / 6 parameters per row

	type update struct {
		accounts *extsvc.Accounts
		perm     *authz.RepoPermissions
	}
	tests := []struct {
		name                   string
		slowTest               bool
		updates                []update
		expectUserPendingPerms map[extsvc.AccountSpec][]uint32 // account -> object_ids
		expectRepoPendingPerms map[int32][]extsvc.AccountSpec  // repo_id -> accounts
	}{
		{
			name: "empty",
			updates: []update{
				{
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  nil,
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
			},
		},
		{
			name: "add",
			updates: []update{
				{
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice", "bob"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 2,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.Accounts{
						ServiceType: "github",
						ServiceID:   "https://github.com/",
						AccountIDs:  []string{"cindy"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 3,
						Perm:   authz.Read,
					},
				},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{
				alice:       {1, 2},
				bob:         {2},
				cindyGitHub: {3},
			},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {alice},
				2: {alice, bob},
				3: {cindyGitHub},
			},
		},
		{
			name: "add and update",
			updates: []update{
				{
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice", "bob"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"bob", "cindy"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.Accounts{
						ServiceType: "github",
						ServiceID:   "https://github.com/",
						AccountIDs:  []string{"cindy"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 2,
						Perm:   authz.Read,
					},
				},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{
				alice:       {},
				bob:         {1},
				cindy:       {1},
				cindyGitHub: {2},
			},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {bob, cindy},
				2: {cindyGitHub},
			},
		},
		{
			name: "add and clear",
			updates: []update{
				{
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice", "bob", "cindy"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{
				alice: {},
				bob:   {},
				cindy: {},
			},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {},
			},
		},
		{
			name:     postgresParameterLimitTest,
			slowTest: true,
			updates: func() []update {
				u := update{
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  make([]string, countToExceedParameterLimit),
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}
				for i := 1; i <= countToExceedParameterLimit; i++ {
					u.accounts.AccountIDs[i-1] = fmt.Sprintf("%d", i)
				}
				return []update{u}
			}(),
			expectUserPendingPerms: func() map[extsvc.AccountSpec][]uint32 {
				perms := make(map[extsvc.AccountSpec][]uint32, countToExceedParameterLimit)
				for i := 1; i <= countToExceedParameterLimit; i++ {
					perms[extsvc.AccountSpec{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountID:   fmt.Sprintf("%d", i),
					}] = []uint32{1}
				}
				return perms
			}(),
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: func() []extsvc.AccountSpec {
					accounts := make([]extsvc.AccountSpec, countToExceedParameterLimit)
					for i := 1; i <= countToExceedParameterLimit; i++ {
						accounts[i-1] = extsvc.AccountSpec{
							ServiceType: authz.SourcegraphServiceType,
							ServiceID:   authz.SourcegraphServiceID,
							AccountID:   fmt.Sprintf("%d", i),
						}
					}
					return accounts
				}(),
			},
		},
	}

	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.slowTest && !*slowTests {
					t.Skip("slow-tests not enabled")
				}

				s := perms(logger, db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
				})

				ctx := context.Background()

				for _, update := range test.updates {
					const numOps = 30
					g, ctx := errgroup.WithContext(ctx)
					for i := 0; i < numOps; i++ {
						// Make local copy to prevent race conditions
						accounts := *update.accounts
						perm := &authz.RepoPermissions{
							RepoID:    update.perm.RepoID,
							Perm:      update.perm.Perm,
							UpdatedAt: update.perm.UpdatedAt,
						}
						if update.perm.UserIDs != nil {
							perm.UserIDs = update.perm.UserIDs
						}
						g.Go(func() error {
							return s.SetRepoPendingPermissions(ctx, &accounts, perm)
						})
					}
					if err := g.Wait(); err != nil {
						t.Fatal(err)
					}
				}

				// Query and check rows in "user_pending_permissions" table.
				idToSpecs, err := checkUserPendingPermsTable(ctx, s, test.expectUserPendingPerms)
				if err != nil {
					t.Fatal("user_pending_permissions:", err)
				}

				// Query and check rows in "repo_pending_permissions" table.
				err = checkRepoPendingPermsTable(ctx, s, idToSpecs, test.expectRepoPendingPerms)
				if err != nil {
					t.Fatal("repo_pending_permissions:", err)
				}
			})
		}
	}
}

func testPermsStore_ListPendingUsers(db database.DB) func(*testing.T) {
	type update struct {
		accounts *extsvc.Accounts
		perm     *authz.RepoPermissions
	}
	tests := []struct {
		name               string
		updates            []update
		expectPendingUsers []string
	}{
		{
			name:               "no user with pending permissions",
			expectPendingUsers: nil,
		},
		{
			name: "has user with pending permissions",
			updates: []update{
				{
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
			},
			expectPendingUsers: []string{"alice"},
		},
		{
			name: "has user but with empty object_ids",
			updates: []update{
				{
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"bob@example.com"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.Accounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  nil,
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
			},
			expectPendingUsers: nil,
		},
		{
			name: "has user but with different service ID",
			updates: []update{
				{
					accounts: &extsvc.Accounts{
						ServiceType: extsvc.TypeGitLab,
						ServiceID:   "https://gitlab.com/",
						AccountIDs:  []string{"bob@example.com"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
			},
			expectPendingUsers: nil,
		},
	}
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := perms(logger, db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
				})

				ctx := context.Background()

				for _, update := range test.updates {
					tmp := &authz.RepoPermissions{
						RepoID:    update.perm.RepoID,
						Perm:      update.perm.Perm,
						UpdatedAt: update.perm.UpdatedAt,
					}
					if update.perm.UserIDs != nil {
						tmp.UserIDs = update.perm.UserIDs
					}
					if err := s.SetRepoPendingPermissions(ctx, update.accounts, tmp); err != nil {
						t.Fatal(err)
					}
				}

				bindIDs, err := s.ListPendingUsers(ctx, authz.SourcegraphServiceType, authz.SourcegraphServiceID)
				if err != nil {
					t.Fatal(err)
				}
				equal(t, "bindIDs", test.expectPendingUsers, bindIDs)
			})
		}
	}
}

func TestPermsStore_GrantPendingPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Background()

	alice := extsvc.AccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "alice",
	}
	bob := extsvc.AccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "bob",
	}

	type ExternalAccount struct {
		ID     int32
		UserID int32
		extsvc.AccountSpec
	}

	setupExternalAccounts := func(accounts []ExternalAccount) {
		users := make(map[int32]*sqlf.Query)
		values := make([]*sqlf.Query, 0, len(accounts))
		for _, a := range accounts {
			if _, ok := users[a.UserID]; !ok {
				users[a.UserID] = sqlf.Sprintf("(%s::integer, %s::text)", a.UserID, fmt.Sprintf("user-%d", a.UserID))
			}
			values = append(values, sqlf.Sprintf("(%s::integer, %s::integer, %s::text, %s::text, %s::text, %s::text)",
				a.ID, a.UserID, a.ServiceType, a.ServiceID, a.AccountID, a.ClientID))
		}
		userQuery := sqlf.Sprintf("INSERT INTO users(id, username) VALUES %s", sqlf.Join(maps.Values(users), ","))
		execQuery(t, ctx, s, userQuery)

		accountQuery := sqlf.Sprintf("INSERT INTO user_external_accounts(id, user_id, service_type, service_id, account_id, client_id) VALUES %s", sqlf.Join(values, ","))
		execQuery(t, ctx, s, accountQuery)
	}

	// this limit will also exceed param limit for user_repo_permissions,
	// as we are sending 6 parameter per row
	const countToExceedParameterLimit = 17000 // ~ 65535 / 4 parameters per row

	type pending struct {
		accounts *extsvc.Accounts
		perm     *authz.RepoPermissions
	}
	type update struct {
		regulars []*authz.RepoPermissions
		pendings []pending
	}
	tests := []struct {
		name                   string
		slowTest               bool
		updates                []update
		grants                 []*authz.UserGrantPermissions
		expectUserRepoPerms    []authz.Permission
		expectUserPerms        map[int32][]uint32              // user_id -> object_ids
		expectRepoPerms        map[int32][]uint32              // repo_id -> user_ids
		expectUserPendingPerms map[extsvc.AccountSpec][]uint32 // account -> object_ids
		expectRepoPendingPerms map[int32][]extsvc.AccountSpec  // repo_id -> accounts

		upsertRepoPermissionsPageSize int
	}{
		{
			name: "empty",
			grants: []*authz.UserGrantPermissions{
				{
					UserID:                1,
					UserExternalAccountID: 1,
					ServiceType:           authz.SourcegraphServiceType,
					ServiceID:             authz.SourcegraphServiceID,
					AccountID:             "alice",
				},
			},
			expectUserRepoPerms: []authz.Permission{},
		},
		{
			name: "no matching pending permissions",
			updates: []update{
				{
					regulars: []*authz.RepoPermissions{
						{
							RepoID:  1,
							Perm:    authz.Read,
							UserIDs: toMapset(1),
						}, {
							RepoID:  2,
							Perm:    authz.Read,
							UserIDs: toMapset(1, 2),
						},
					},
					pendings: []pending{
						{
							accounts: &extsvc.Accounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"alice"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 1,
								Perm:   authz.Read,
							},
						}, {
							accounts: &extsvc.Accounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"bob"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 2,
								Perm:   authz.Read,
							},
						},
					},
				},
			},
			grants: []*authz.UserGrantPermissions{
				{
					UserID:                1,
					UserExternalAccountID: 3,
					ServiceType:           authz.SourcegraphServiceType,
					ServiceID:             authz.SourcegraphServiceID,
					AccountID:             "cindy",
				},
			},
			expectUserRepoPerms: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 2, ExternalAccountID: 2, RepoID: 2},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1, 2},
				2: {2},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1},
				2: {1, 2},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{
				alice: {1},
				bob:   {2},
			},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {alice},
				2: {bob},
			},
		},
		{
			name: "grant pending permission",
			updates: []update{
				{
					regulars: []*authz.RepoPermissions{},
					pendings: []pending{{
						accounts: &extsvc.Accounts{
							ServiceType: authz.SourcegraphServiceType,
							ServiceID:   authz.SourcegraphServiceID,
							AccountIDs:  []string{"alice"},
						},
						perm: &authz.RepoPermissions{
							RepoID: 1,
							Perm:   authz.Read,
						},
					}},
				},
			},
			grants: []*authz.UserGrantPermissions{{
				UserID:                1,
				UserExternalAccountID: 1,
				ServiceType:           authz.SourcegraphServiceType,
				ServiceID:             authz.SourcegraphServiceID,
				AccountID:             "alice",
			}},
			expectUserRepoPerms: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {},
			},
		},
		{
			name: "union matching pending permissions for same account ID but different service IDs",
			updates: []update{
				{
					regulars: []*authz.RepoPermissions{
						{
							RepoID:  1,
							Perm:    authz.Read,
							UserIDs: toMapset(1),
						}, {
							RepoID:  2,
							Perm:    authz.Read,
							UserIDs: toMapset(1, 2),
						},
					},
					pendings: []pending{
						{
							accounts: &extsvc.Accounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"alice"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 1,
								Perm:   authz.Read,
							},
						},
						{
							accounts: &extsvc.Accounts{
								ServiceType: extsvc.TypeGitLab,
								ServiceID:   "https://gitlab.com/",
								AccountIDs:  []string{"alice"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 2,
								Perm:   authz.Read,
							},
						}, {
							accounts: &extsvc.Accounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"bob"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 3,
								Perm:   authz.Read,
							},
						},
					},
				},
			},
			grants: []*authz.UserGrantPermissions{
				{
					UserID:                3,
					UserExternalAccountID: 3,
					ServiceType:           authz.SourcegraphServiceType,
					ServiceID:             authz.SourcegraphServiceID,
					AccountID:             "alice",
				}, {
					UserID:                3,
					UserExternalAccountID: 4,
					ServiceType:           extsvc.TypeGitLab,
					ServiceID:             "https://gitlab.com/",
					AccountID:             "alice",
				},
			},
			expectUserRepoPerms: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 2, ExternalAccountID: 2, RepoID: 2},
				{UserID: 3, ExternalAccountID: 3, RepoID: 1},
				{UserID: 3, ExternalAccountID: 4, RepoID: 2},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1, 2},
				2: {2},
				3: {1, 2},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1, 3},
				2: {1, 2, 3},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{
				bob: {3},
			},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {},
				2: {},
				3: {bob},
			},
		},
		{
			name: "union matching pending permissions for same service ID but different account IDs",
			updates: []update{
				{
					regulars: []*authz.RepoPermissions{
						{
							RepoID:  1,
							Perm:    authz.Read,
							UserIDs: toMapset(1),
						}, {
							RepoID:  2,
							Perm:    authz.Read,
							UserIDs: toMapset(1, 2),
						},
					},
					pendings: []pending{
						{
							accounts: &extsvc.Accounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"alice@example.com"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 1,
								Perm:   authz.Read,
							},
						}, {
							accounts: &extsvc.Accounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"alice2@example.com"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 2,
								Perm:   authz.Read,
							},
						},
					},
				},
			},
			grants: []*authz.UserGrantPermissions{
				{
					UserID:                3,
					UserExternalAccountID: 3,
					ServiceType:           authz.SourcegraphServiceType,
					ServiceID:             authz.SourcegraphServiceID,
					AccountID:             "alice@example.com",
				}, {
					UserID:                3,
					UserExternalAccountID: 4,
					ServiceType:           authz.SourcegraphServiceType,
					ServiceID:             authz.SourcegraphServiceID,
					AccountID:             "alice2@example.com",
				},
			},
			expectUserRepoPerms: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 2, ExternalAccountID: 2, RepoID: 2},
				{UserID: 3, ExternalAccountID: 3, RepoID: 1},
				{UserID: 3, ExternalAccountID: 4, RepoID: 2},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1, 2},
				2: {2},
				3: {1, 2},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1, 3},
				2: {1, 2, 3},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {},
				2: {},
			},
		},
		{
			name:                          "grant pending permission with pagination",
			upsertRepoPermissionsPageSize: 2,
			updates: []update{
				{
					regulars: []*authz.RepoPermissions{},
					pendings: []pending{{
						accounts: &extsvc.Accounts{
							ServiceType: authz.SourcegraphServiceType,
							ServiceID:   authz.SourcegraphServiceID,
							AccountIDs:  []string{"alice"},
						},
						perm: &authz.RepoPermissions{
							RepoID: 1,
							Perm:   authz.Read,
						},
					}, {
						accounts: &extsvc.Accounts{
							ServiceType: authz.SourcegraphServiceType,
							ServiceID:   authz.SourcegraphServiceID,
							AccountIDs:  []string{"alice"},
						},
						perm: &authz.RepoPermissions{
							RepoID: 2,
							Perm:   authz.Read,
						},
					}, {
						accounts: &extsvc.Accounts{
							ServiceType: authz.SourcegraphServiceType,
							ServiceID:   authz.SourcegraphServiceID,
							AccountIDs:  []string{"alice"},
						},
						perm: &authz.RepoPermissions{
							RepoID: 3,
							Perm:   authz.Read,
						},
					}},
				},
			},
			grants: []*authz.UserGrantPermissions{{
				UserID:                1,
				UserExternalAccountID: 1,
				ServiceType:           authz.SourcegraphServiceType,
				ServiceID:             authz.SourcegraphServiceID,
				AccountID:             "alice",
			}},
			expectUserRepoPerms: []authz.Permission{
				{UserID: 1, ExternalAccountID: 1, RepoID: 1},
				{UserID: 1, ExternalAccountID: 1, RepoID: 2},
				{UserID: 1, ExternalAccountID: 1, RepoID: 3},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1, 2, 3},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1},
				2: {1},
				3: {1},
			},
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: map[int32][]extsvc.AccountSpec{
				1: {},
				2: {},
				3: {},
			},
		},
		{
			name:     postgresParameterLimitTest,
			slowTest: true,
			updates: []update{
				{
					regulars: []*authz.RepoPermissions{},
					pendings: func() []pending {
						accounts := &extsvc.Accounts{
							ServiceType: authz.SourcegraphServiceType,
							ServiceID:   authz.SourcegraphServiceID,
							AccountIDs:  []string{"alice"},
						}
						pendings := make([]pending, countToExceedParameterLimit)
						for i := 1; i <= countToExceedParameterLimit; i += 1 {
							pendings[i-1] = pending{
								accounts: accounts,
								perm: &authz.RepoPermissions{
									RepoID: int32(i),
									Perm:   authz.Read,
								},
							}
						}
						return pendings
					}(),
				},
			},
			grants: []*authz.UserGrantPermissions{
				{
					UserID:                1,
					UserExternalAccountID: 1,
					ServiceType:           authz.SourcegraphServiceType,
					ServiceID:             authz.SourcegraphServiceID,
					AccountID:             "alice",
				},
			},
			expectUserRepoPerms: func() []authz.Permission {
				perms := make([]authz.Permission, 0, countToExceedParameterLimit)
				for i := 1; i <= countToExceedParameterLimit; i += 1 {
					perms = append(perms, authz.Permission{
						UserID:            1,
						ExternalAccountID: 1,
						RepoID:            int32(i),
					})
				}
				return perms
			}(),
			expectUserPerms: func() map[int32][]uint32 {
				repos := make([]uint32, countToExceedParameterLimit)
				for i := 1; i <= countToExceedParameterLimit; i += 1 {
					repos[i-1] = uint32(i)
				}
				return map[int32][]uint32{1: repos}
			}(),
			expectRepoPerms: func() map[int32][]uint32 {
				repos := make(map[int32][]uint32, countToExceedParameterLimit)
				for i := 1; i <= countToExceedParameterLimit; i += 1 {
					repos[int32(i)] = []uint32{1}
				}
				return repos
			}(),
			expectUserPendingPerms: map[extsvc.AccountSpec][]uint32{},
			expectRepoPendingPerms: func() map[int32][]extsvc.AccountSpec {
				repos := make(map[int32][]extsvc.AccountSpec, countToExceedParameterLimit)
				for i := 1; i <= countToExceedParameterLimit; i += 1 {
					repos[int32(i)] = []extsvc.AccountSpec{}
				}
				return repos
			}(),
		},
	}
	for _, test := range tests {
		if t.Failed() {
			break
		}

		t.Run(test.name, func(t *testing.T) {
			if test.slowTest && !*slowTests {
				t.Skip("slow-tests not enabled")
			}

			if test.upsertRepoPermissionsPageSize > 0 {
				upsertRepoPermissionsPageSize = test.upsertRepoPermissionsPageSize
			}

			t.Cleanup(func() {
				cleanupPermsTables(t, s)
				cleanupUsersTable(t, s)
				cleanupReposTable(t, s)

				if test.upsertRepoPermissionsPageSize > 0 {
					upsertRepoPermissionsPageSize = defaultUpsertRepoPermissionsPageSize
				}
			})

			accounts := make([]ExternalAccount, 0)
			for _, grant := range test.grants {
				accounts = append(accounts, ExternalAccount{
					ID:     grant.UserExternalAccountID,
					UserID: grant.UserID,
					AccountSpec: extsvc.AccountSpec{
						ServiceType: grant.ServiceType,
						ServiceID:   grant.ServiceID,
						AccountID:   grant.AccountID,
						ClientID:    "client_id",
					},
				})
			}

			// create related entities
			if len(accounts) > 0 {
				setupExternalAccounts(accounts)
			}

			if len(test.expectUserRepoPerms) > 0 {
				setupPermsRelatedEntities(t, s, test.expectUserRepoPerms)
			}

			for _, update := range test.updates {
				for _, p := range update.regulars {
					repoID := p.RepoID
					users := make([]authz.UserIDWithExternalAccountID, 0, len(p.UserIDs))
					for userID := range p.UserIDs {
						users = append(users, authz.UserIDWithExternalAccountID{
							UserID:            userID,
							ExternalAccountID: userID,
						})
					}

					if err := s.SetRepoPerms(ctx, repoID, users); err != nil {
						t.Fatal(err)
					}

					if _, err := s.SetRepoPermissions(ctx, p); err != nil {
						t.Fatal(err)
					}
				}
				for _, p := range update.pendings {
					if err := s.SetRepoPendingPermissions(ctx, p.accounts, p.perm); err != nil {
						t.Fatal(err)
					}
				}
			}

			for _, grant := range test.grants {
				err := s.GrantPendingPermissions(ctx, grant)
				if err != nil {
					t.Fatal(err)
				}
			}

			checkUserRepoPermissions(t, s, sqlf.Sprintf("TRUE"), test.expectUserRepoPerms)

			err := checkRegularPermsTable(s, `SELECT user_id, object_ids_ints FROM user_permissions`, test.expectUserPerms)
			if err != nil {
				t.Fatal("user_permissions:", err)
			}

			err = checkRegularPermsTable(s, `SELECT repo_id, user_ids_ints FROM repo_permissions`, test.expectRepoPerms)
			if err != nil {
				t.Fatal("repo_permissions:", err)
			}

			// Query and check rows in "user_pending_permissions" table.
			idToSpecs, err := checkUserPendingPermsTable(ctx, s, test.expectUserPendingPerms)
			if err != nil {
				t.Fatal("user_pending_permissions:", err)
			}

			// Query and check rows in "repo_pending_permissions" table.
			err = checkRepoPendingPermsTable(ctx, s, idToSpecs, test.expectRepoPendingPerms)
			if err != nil {
				t.Fatal("repo_pending_permissions:", err)
			}
		})
	}
}

// This test is used to ensure we ignore invalid pending user IDs on updating repository pending permissions
// because permissions have been granted for those users.
func TestPermsStore_SetPendingPermissionsAfterGrant(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, clock)
	defer cleanupPermsTables(t, s)

	ctx := context.Background()

	setupPermsRelatedEntities(t, s, []authz.Permission{
		{
			UserID:            1,
			RepoID:            1,
			ExternalAccountID: 1,
		},
		{
			UserID:            2,
			RepoID:            1,
			ExternalAccountID: 2,
		},
	})

	// Set up pending permissions for at least two users
	if err := s.SetRepoPendingPermissions(ctx, &extsvc.Accounts{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountIDs:  []string{"alice", "bob"},
	}, &authz.RepoPermissions{
		RepoID: 1,
		Perm:   authz.Read,
	}); err != nil {
		t.Fatal(err)
	}

	// Now grant permissions for these two users, which effectively remove corresponding rows
	// from the `user_pending_permissions` table.
	if err := s.GrantPendingPermissions(ctx, &authz.UserGrantPermissions{
		UserID:      1,
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "alice",
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.GrantPendingPermissions(ctx, &authz.UserGrantPermissions{
		UserID:      2,
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "bob",
	}); err != nil {
		t.Fatal(err)
	}

	// Now the `repo_pending_permissions` table has references to these two deleted rows,
	// it should just ignore them.
	if err := s.SetRepoPendingPermissions(ctx, &extsvc.Accounts{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountIDs:  []string{}, // Intentionally empty to cover "no-update" case
	}, &authz.RepoPermissions{
		RepoID: 1,
		Perm:   authz.Read,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPermsStore_DeleteAllUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, clock)
	t.Cleanup(func() {
		cleanupPermsTables(t, s)
		cleanupUsersTable(t, s)
		cleanupReposTable(t, s)
	})

	ctx := context.Background()

	// Create 2 users and their external accounts and repos
	// Set up test users and external accounts
	extSQL := `
	INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id, client_id, created_at, updated_at, deleted_at, expired_at)
		VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)
	`
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`), // ID=1
		sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),   // ID=2

		sqlf.Sprintf(extSQL, 1, extsvc.TypeGitLab, "https://gitlab.com/", "alice_gitlab", "alice_gitlab_client_id", clock(), clock(), nil, nil), // ID=1
		sqlf.Sprintf(extSQL, 1, "github", "https://github.com/", "alice_github", "alice_github_client_id", clock(), clock(), nil, nil),          // ID=2
		sqlf.Sprintf(extSQL, 2, extsvc.TypeGitLab, "https://gitlab.com/", "bob_gitlab", "bob_gitlab_client_id", clock(), clock(), nil, nil),     // ID=3

		sqlf.Sprintf(`INSERT INTO repo(name, private) VALUES('private_repo_1', TRUE)`), // ID=1
		sqlf.Sprintf(`INSERT INTO repo(name, private) VALUES('private_repo_2', TRUE)`), // ID=2
	}
	for _, q := range qs {
		execQuery(t, ctx, s, q)
	}

	// Set permissions for user 1 and 2
	for _, repoID := range []int32{1, 2} {
		if _, err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  repoID,
			Perm:    authz.Read,
			UserIDs: toMapset(1, 2),
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Set unified permissions for user 1 and 2
	for _, userID := range []int32{1, 2} {
		for _, repoID := range []int32{1, 2} {
			if err := s.SetUserExternalAccountPerms(ctx, authz.UserIDWithExternalAccountID{UserID: userID, ExternalAccountID: repoID}, []int32{repoID}); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Remove all permissions for the user=1
	if err := s.DeleteAllUserPermissions(ctx, 1); err != nil {
		t.Fatal(err)
	}

	// Check user=1 should not have any legacy permissions now
	p, err := s.LoadUserPermissions(ctx, 1)
	require.NoError(t, err)
	assert.Zero(t, len(p))

	getUserRepoPermissions := func(userID int) ([]int32, error) {
		unifiedQuery := `SELECT repo_id FROM user_repo_permissions WHERE user_id = %d`
		q := sqlf.Sprintf(unifiedQuery, userID)
		return basestore.ScanInt32s(db.Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
	}

	// Check user=1 should not have any permissions now
	results, err := getUserRepoPermissions(1)
	assert.NoError(t, err)
	assert.Nil(t, results)

	// Check user=2 shoud still have legacy permissions
	p, err = s.LoadUserPermissions(ctx, 2)
	require.NoError(t, err)
	gotIDs := make([]int32, len(p))
	for i, perm := range p {
		gotIDs[i] = perm.RepoID
	}
	slices.Sort(gotIDs)
	equal(t, "legacy IDs", []int32{1, 2}, gotIDs)

	// Check user=2 should still have permissions
	results, err = getUserRepoPermissions(2)
	assert.NoError(t, err)
	equal(t, "unified IDs", []int32{1, 2}, results)
}

func testPermsStore_DeleteAllUserPendingPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		s := perms(logger, db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		accounts := &extsvc.Accounts{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountIDs:  []string{"alice", "bob"},
		}

		// Set pending permissions for "alice" and "bob"
		if err := s.SetRepoPendingPermissions(ctx, accounts, &authz.RepoPermissions{
			RepoID: 1,
			Perm:   authz.Read,
		}); err != nil {
			t.Fatal(err)
		}

		// Remove all pending permissions for "alice"
		accounts.AccountIDs = []string{"alice"}
		if err := s.DeleteAllUserPendingPermissions(ctx, accounts); err != nil {
			t.Fatal(err)
		}

		// Check alice should not have any pending permissions now
		err := s.LoadUserPendingPermissions(ctx, &authz.UserPendingPermissions{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			BindID:      "alice",
			Perm:        authz.Read,
			Type:        authz.PermRepos,
		})
		if err != authz.ErrPermsNotFound {
			t.Fatalf("err: want %q but got %v", authz.ErrPermsNotFound, err)
		}

		// Check bob shoud not be affected
		p := &authz.UserPendingPermissions{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			BindID:      "bob",
			Perm:        authz.Read,
			Type:        authz.PermRepos,
		}
		err = s.LoadUserPendingPermissions(ctx, p)
		if err != nil {
			t.Fatal(err)
		}
		equal(t, "p.IDs", []int{1}, mapsetToArray(p.IDs))
	}
}

func TestPermsStore_DatabaseDeadlocks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, time.Now)
	t.Cleanup(func() {
		cleanupPermsTables(t, s)
	})

	ctx := context.Background()

	setupPermsRelatedEntities(t, s, []authz.Permission{
		{
			UserID:            1,
			RepoID:            1,
			ExternalAccountID: 1,
		},
	})

	setUserPermissions := func(ctx context.Context, t *testing.T) {
		if _, err := s.SetUserPermissions(ctx, &authz.UserPermissions{
			UserID: 1,
			Perm:   authz.Read,
			IDs:    toMapset(1),
		}); err != nil {
			t.Fatal(err)
		}
	}
	setRepoPermissions := func(ctx context.Context, t *testing.T) {
		if _, err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toMapset(1),
		}); err != nil {
			t.Fatal(err)
		}
	}
	setRepoPendingPermissions := func(ctx context.Context, t *testing.T) {
		accounts := &extsvc.Accounts{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountIDs:  []string{"alice"},
		}
		if err := s.SetRepoPendingPermissions(ctx, accounts, &authz.RepoPermissions{
			RepoID: 1,
			Perm:   authz.Read,
		}); err != nil {
			t.Fatal(err)
		}
	}
	grantPendingPermissions := func(ctx context.Context, t *testing.T) {
		if err := s.GrantPendingPermissions(ctx, &authz.UserGrantPermissions{
			UserID:      1,
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountID:   "alice",
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Ensure we've run all permutations of ordering of the 4 calls to avoid nondeterminism in
	// test coverage stats.
	funcs := []func(context.Context, *testing.T){
		setRepoPendingPermissions, grantPendingPermissions, setRepoPermissions, setUserPermissions,
	}
	permutated := permutation.New(permutation.MustAnySlice(funcs))
	for permutated.Next() {
		for _, f := range funcs {
			f(ctx, t)
		}
	}

	const numOps = 50
	var wg sync.WaitGroup
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
			grantPendingPermissions(ctx, t)
		}
	}()

	wg.Wait()
}

func cleanupUsersTable(t *testing.T, s *permsStore) {
	t.Helper()

	q := `DELETE FROM user_external_accounts;`
	execQuery(t, context.Background(), s, sqlf.Sprintf(q))

	q = `TRUNCATE TABLE users RESTART IDENTITY CASCADE;`
	execQuery(t, context.Background(), s, sqlf.Sprintf(q))
}

func testPermsStore_GetUserIDsByExternalAccounts(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		s := perms(logger, db, time.Now)
		t.Cleanup(func() {
			cleanupUsersTable(t, s)
		})

		ctx := context.Background()

		// Set up test users and external accounts
		extSQL := `
INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id, client_id, created_at, updated_at, deleted_at, expired_at)
	VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)
`
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`),  // ID=1
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),    // ID=2
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('cindy')`),  // ID=3
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('denise')`), // ID=4

			sqlf.Sprintf(extSQL, 1, extsvc.TypeGitLab, "https://gitlab.com/", "alice_gitlab", "alice_gitlab_client_id", clock(), clock(), nil, nil), // ID=1
			sqlf.Sprintf(extSQL, 1, "github", "https://github.com/", "alice_github", "alice_github_client_id", clock(), clock(), nil, nil),          // ID=2
			sqlf.Sprintf(extSQL, 2, extsvc.TypeGitLab, "https://gitlab.com/", "bob_gitlab", "bob_gitlab_client_id", clock(), clock(), nil, nil),     // ID=3
			sqlf.Sprintf(extSQL, 3, extsvc.TypeGitLab, "https://gitlab.com/", "cindy_gitlab", "cindy_gitlab_client_id", clock(), clock(), nil, nil), // ID=4
			sqlf.Sprintf(extSQL, 3, "github", "https://github.com/", "cindy_github", "cindy_github_client_id", clock(), clock(), clock(), nil),      // ID=5, deleted
			sqlf.Sprintf(extSQL, 4, "github", "https://github.com/", "denise_github", "denise_github_client_id", clock(), clock(), nil, clock()),    // ID=6, expired
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		accounts := &extsvc.Accounts{
			ServiceType: "gitlab",
			ServiceID:   "https://gitlab.com/",
			AccountIDs:  []string{"alice_gitlab", "bob_gitlab", "david_gitlab"},
		}
		userIDs, err := s.GetUserIDsByExternalAccounts(ctx, accounts)
		if err != nil {
			t.Fatal(err)
		}

		if len(userIDs) != 2 {
			t.Fatalf("len(userIDs): want 2 but got %v", userIDs)
		}

		assert.Equal(t, int32(1), userIDs["alice_gitlab"].UserID)
		assert.Equal(t, int32(1), userIDs["alice_gitlab"].ExternalAccountID)
		assert.Equal(t, int32(2), userIDs["bob_gitlab"].UserID)
		assert.Equal(t, int32(3), userIDs["bob_gitlab"].ExternalAccountID)

		accounts = &extsvc.Accounts{
			ServiceType: "github",
			ServiceID:   "https://github.com/",
			AccountIDs:  []string{"cindy_github", "denise_github"},
		}
		userIDs, err = s.GetUserIDsByExternalAccounts(ctx, accounts)
		require.Nil(t, err)
		assert.Empty(t, userIDs)
	}
}

func mockUnifiedPermsConfig(val bool) {
	cfg := &conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			UnifiedPermissions: val,
		},
	}}
	conf.Mock(cfg)
}

func execQuery(t *testing.T, ctx context.Context, s *permsStore, q *sqlf.Query) {
	t.Helper()
	if t.Failed() {
		return
	}

	err := s.execute(ctx, q)
	if err != nil {
		t.Fatalf("Error executing query %v, err: %v", q, err)
	}
}

func TestPermsStore_UserIDsWithNoPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, time.Now)

	t.Cleanup(func() {
		cleanupPermsTables(t, s)
		cleanupUsersTable(t, s)
		cleanupReposTable(t, s)
	})

	ctx := context.Background()

	// Create test users "alice" and "bob", test repo and test external account
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`),                    // ID=1
		sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),                      // ID=2
		sqlf.Sprintf(`INSERT INTO users(username, deleted_at) VALUES('cindy', NOW())`), // ID=3
		sqlf.Sprintf(`INSERT INTO users(username) VALUES('david')`),                    // ID=4
		sqlf.Sprintf(`INSERT INTO repo(name, private) VALUES('private_repo', TRUE)`),   // ID=1
		sqlf.Sprintf(`INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id, client_id, created_at, updated_at, deleted_at, expired_at)
				VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)`, 1, extsvc.TypeGitLab, "https://gitlab.com/", "alice_gitlab", "alice_gitlab_client_id", clock(), clock(), nil, nil), // ID=1
	}
	for _, q := range qs {
		execQuery(t, ctx, s, q)
	}

	// "alice", "bob" and "david" have no permissions
	ids, err := s.UserIDsWithNoPerms(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	expIDs := []int32{1, 2, 4}
	if diff := cmp.Diff(expIDs, ids); diff != "" {
		t.Fatal(diff)
	}

	t.Run("legacy user_permissions table", func(t *testing.T) {
		mockUnifiedPermsConfig(false)

		s.SetUserPermissions(ctx, &authz.UserPermissions{UserID: 1, IDs: map[int32]struct{}{1: {}}})
		s.SetUserPermissions(ctx, &authz.UserPermissions{UserID: 2, IDs: make(map[int32]struct{})})

		// Only "david" has no permissions at this point
		ids, err = s.UserIDsWithNoPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}

		expIDs = []int32{4}
		if diff := cmp.Diff(expIDs, ids); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("unified user_repo_permissions table", func(t *testing.T) {
		mockUnifiedPermsConfig(true)

		// mark sync jobs as completed for "alice" and add permissions for "bob"
		q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_at, reason) VALUES(%d, NOW(), %s)`, 1, database.ReasonUserNoPermissions)
		execQuery(t, ctx, s, q)

		s.SetUserExternalAccountPerms(ctx, authz.UserIDWithExternalAccountID{UserID: 2, ExternalAccountID: 1}, []int32{1})

		// Only "david" has no permissions at this point
		ids, err = s.UserIDsWithNoPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}

		expIDs = []int32{4}
		if diff := cmp.Diff(expIDs, ids); diff != "" {
			t.Fatal(diff)
		}
	})

}

func cleanupReposTable(t *testing.T, s *permsStore) {
	t.Helper()

	q := `TRUNCATE TABLE repo RESTART IDENTITY CASCADE;`
	execQuery(t, context.Background(), s, sqlf.Sprintf(q))
}

func TestPermsStore_RepoIDsWithNoPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, time.Now)

	t.Cleanup(func() {
		cleanupPermsTables(t, s)
		cleanupReposTable(t, s)
		cleanupUsersTable(t, s)
	})

	ctx := context.Background()

	// Create three test repositories
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO repo(name, private) VALUES('private_repo', TRUE)`),                      // ID=1
		sqlf.Sprintf(`INSERT INTO repo(name) VALUES('public_repo')`),                                      // ID=2
		sqlf.Sprintf(`INSERT INTO repo(name, private) VALUES('private_repo_2', TRUE)`),                    // ID=3
		sqlf.Sprintf(`INSERT INTO repo(name, private, deleted_at) VALUES('private_repo_3', TRUE, NOW())`), // ID=4
		sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`),                                       // ID=1
		sqlf.Sprintf(`INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id, client_id, created_at, updated_at, deleted_at, expired_at)
				VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s)`, 1, extsvc.TypeGitLab, "https://gitlab.com/", "alice_gitlab", "alice_gitlab_client_id", clock(), clock(), nil, nil), // ID=1
	}
	for _, q := range qs {
		execQuery(t, ctx, s, q)
	}

	// Should get back two private repos that are not deleted
	ids, err := s.RepoIDsWithNoPerms(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	expIDs := []api.RepoID{1, 3}
	if diff := cmp.Diff(expIDs, ids); diff != "" {
		t.Fatal(diff)
	}

	t.Run("legacy user_permissions table", func(t *testing.T) {
		mockUnifiedPermsConfig(false)

		s.SetRepoPermissions(ctx, &authz.RepoPermissions{RepoID: 1, UserIDs: map[int32]struct{}{1: {}}})
		s.SetRepoPermissions(ctx, &authz.RepoPermissions{RepoID: 3, UserIDs: make(map[int32]struct{})})

		// No private repositories have any permissions at this point
		ids, err = s.RepoIDsWithNoPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, ids)
	})

	t.Run("unified user_repo_permissions table", func(t *testing.T) {
		mockUnifiedPermsConfig(true)

		// mark sync jobs as completed for "private_repo" and add permissions for "private_repo_2"
		q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_at, reason) VALUES(%d, NOW(), %s)`, 1, database.ReasonRepoNoPermissions)
		execQuery(t, ctx, s, q)

		err := s.SetRepoPerms(ctx, 3, []authz.UserIDWithExternalAccountID{{UserID: 1, ExternalAccountID: 1}})
		if err != nil {
			t.Fatal(err)
		}

		// No private repositories have any permissions at this point
		ids, err = s.RepoIDsWithNoPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, ids)
	})
}

func TestPermsStore_UserIDsWithOldestPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Background()

	t.Cleanup(func() {
		cleanupPermsTables(t, s)
		cleanupReposTable(t, s)
		cleanupUsersTable(t, s)
	})

	// Set up some users and permissions
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO users(id, username) VALUES(1, 'alice')`),
		sqlf.Sprintf(`INSERT INTO users(id, username) VALUES(2, 'bob')`),
		sqlf.Sprintf(`INSERT INTO users(id, username, deleted_at) VALUES(3, 'cindy', NOW())`),
	}
	for _, q := range qs {
		execQuery(t, ctx, s, q)
	}

	// mark sync jobs as completed for users 1 and 2
	user1UpdatedAt := clock().Add(-15 * time.Minute)
	user2UpdatedAt := clock().Add(-5 * time.Minute)
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_at, reason) VALUES(%d, %s, %s)`, 1, user1UpdatedAt, database.ReasonUserOutdatedPermissions)
	execQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(user_id, finished_at, reason) VALUES(%d, %s, %s)`, 2, user2UpdatedAt, database.ReasonUserOutdatedPermissions)
	execQuery(t, ctx, s, q)

	t.Run("One result when limit is 1", func(t *testing.T) {
		// Should only get user 1 back, because limit is 1
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[int32]time.Time{1: user1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("One result when limit is 10 and age is 10 minutes", func(t *testing.T) {
		// Should only get user 1 back, because age is 10 minutes
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 10*time.Minute)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[int32]time.Time{1: user1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Both users are returned when limit is 10 and age is 1 minute", func(t *testing.T) {
		// Should get both users, since the limit is 10 and age is 1 minute only
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 1*time.Minute)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[int32]time.Time{1: user1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Both users are returned when limit is 10 and age is 0", func(t *testing.T) {
		// Should get both users, since the limit is 10 and age is 0
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[int32]time.Time{1: user1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Ignore users that have synced recently", func(t *testing.T) {
		// Should get no results, since the and age is 1 hour
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 1*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := make(map[int32]time.Time)
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestPermsStore_ReposIDsWithOldestPerms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	testDb := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, testDb)
	s := perms(logger, db, clock)
	ctx := context.Background()
	t.Cleanup(func() {
		cleanupPermsTables(t, s)
		cleanupReposTable(t, s)
	})

	// Set up some repositories and permissions
	qs := []*sqlf.Query{
		sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(1, 'private_repo_1', TRUE)`),                    // id=1
		sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(2, 'private_repo_2', TRUE)`),                    // id=2
		sqlf.Sprintf(`INSERT INTO repo(id, name, private, deleted_at) VALUES(3, 'private_repo_3', TRUE, NOW())`), // id=3
	}
	for _, q := range qs {
		execQuery(t, ctx, s, q)
	}

	// mark sync jobs as completed for private_repo_1 and private_repo_2
	repo1UpdatedAt := clock().Add(-15 * time.Minute)
	repo2UpdatedAt := clock().Add(-5 * time.Minute)
	q := sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_at, reason) VALUES(%d, %s, %s)`, 1, repo1UpdatedAt, database.ReasonRepoOutdatedPermissions)
	execQuery(t, ctx, s, q)
	q = sqlf.Sprintf(`INSERT INTO permission_sync_jobs(repository_id, finished_at, reason) VALUES(%d, %s, %s)`, 2, repo2UpdatedAt, database.ReasonRepoOutdatedPermissions)
	execQuery(t, ctx, s, q)

	t.Run("One result when limit is 1", func(t *testing.T) {
		// Should only get private_repo_1 back, because limit is 1
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[api.RepoID]time.Time{1: repo1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("One result when limit is 10 and age is 10 minutes", func(t *testing.T) {
		// Should only get private_repo_1 back, because age is 10 minutes
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 10*time.Minute)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[api.RepoID]time.Time{1: repo1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Both users are returned when limit is 10 and age is 1 minute", func(t *testing.T) {
		// Should get both private_repo_1 and private_repo_2, since the limit is 10 and age is 1 minute only
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 1*time.Minute)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[api.RepoID]time.Time{1: repo1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Both users are returned when limit is 10 and age is 0", func(t *testing.T) {
		// Should get both private_repo_1 and private_repo_2, since the limit is 10 and age is 0
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[api.RepoID]time.Time{1: repo1UpdatedAt}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Ignore repos that have synced recently", func(t *testing.T) {
		// Should get no results, since the and age is 1 hour
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 1*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := make(map[api.RepoID]time.Time)
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}
	})
}

func testPermsStore_MapUsers(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		s := perms(logger, db, clock)
		ctx := context.Background()
		t.Cleanup(func() {
			if t.Failed() {
				return
			}

			q := `TRUNCATE TABLE external_services, orgs, users CASCADE`
			if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
				t.Fatal(err)
			}
		})

		// Set up 3 users
		users := db.Users()

		igor, err := users.Create(ctx,
			database.NewUser{
				Email:           "igor@example.com",
				Username:        "igor",
				EmailIsVerified: true,
			},
		)
		require.NoError(t, err)
		shreah, err := users.Create(ctx,
			database.NewUser{
				Email:           "shreah@example.com",
				Username:        "shreah",
				EmailIsVerified: true,
			},
		)
		require.NoError(t, err)
		omar, err := users.Create(ctx,
			database.NewUser{
				Email:           "omar@example.com",
				Username:        "omar",
				EmailIsVerified: true,
			},
		)
		require.NoError(t, err)

		// emails: map with a mixed load of existing, space only and non existing users
		has, err := s.MapUsers(ctx, []string{"igor@example.com", "", "omar@example.com", "  	", "sayako@example.com"}, &schema.PermissionsUserMapping{BindID: "email"})
		assert.NoError(t, err)
		assert.Equal(t, map[string]int32{
			"igor@example.com": igor.ID,
			"omar@example.com": omar.ID,
		}, has)

		// usernames: map with a mixed load of existing, space only and non existing users
		has, err = s.MapUsers(ctx, []string{"igor", "", "shreah", "  	", "carlos"}, &schema.PermissionsUserMapping{BindID: "username"})
		assert.NoError(t, err)
		assert.Equal(t, map[string]int32{
			"igor":   igor.ID,
			"shreah": shreah.ID,
		}, has)

		// use a non-existing mapping
		_, err = s.MapUsers(ctx, []string{"igor", "", "shreah", "  	", "carlos"}, &schema.PermissionsUserMapping{BindID: "shoeSize"})
		assert.Error(t, err)
	}
}

func testPermsStore_Metrics(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		s := perms(logger, db, clock)

		ctx := context.Background()
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
			cleanupUsersTable(t, s)
			if t.Failed() {
				return
			}

			if err := s.execute(ctx, sqlf.Sprintf(`DELETE FROM repo`)); err != nil {
				t.Fatal(err)
			}
		})

		// Set up repositories in various states (public/private, deleted/not, etc.)
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(1, 'private_repo_1', TRUE)`),
			sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(2, 'private_repo_2', TRUE)`),
			sqlf.Sprintf(`INSERT INTO repo(id, name, private, deleted_at) VALUES(3, 'private_repo_3', FALSE, NOW())`),
			sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(4, 'private_repo_4', TRUE)`),
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		// Set up users in various states (deleted/not, etc.)
		qs = []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(id, username) VALUES(1, 'user1')`),
			sqlf.Sprintf(`INSERT INTO users(id, username) VALUES(2, 'user2')`),
			sqlf.Sprintf(`INSERT INTO users(id, username, deleted_at) VALUES(3, 'user3', NOW())`),
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		// Set up permissions for the various repos.
		for i := 0; i < 4; i++ {
			_, err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
				RepoID:  int32(i),
				Perm:    authz.Read,
				UserIDs: toMapset(1, 2, 3, 4),
			})
			if err != nil {
				t.Fatal(err)
			}
		}

		// Mock rows for testing
		qs = []*sqlf.Query{
			sqlf.Sprintf(`UPDATE user_permissions SET updated_at = %s WHERE user_id = 1`, clock()),
			sqlf.Sprintf(`UPDATE user_permissions SET updated_at = %s WHERE user_id = 2`, clock().Add(-1*time.Minute)),
			sqlf.Sprintf(`UPDATE user_permissions SET updated_at = %s WHERE user_id = 3`, clock().Add(-2*time.Minute)), // Meant to be excluded because it has been deleted
			sqlf.Sprintf(`UPDATE repo_permissions SET updated_at = %s WHERE repo_id = 1`, clock()),
			sqlf.Sprintf(`UPDATE repo_permissions SET updated_at = %s WHERE repo_id = 2`, clock().Add(-2*time.Minute)),
			sqlf.Sprintf(`UPDATE repo_permissions SET updated_at = %s WHERE repo_id = 3`, clock().Add(-3*time.Minute)), // Meant to be excluded because it has been deleted
			sqlf.Sprintf(`UPDATE repo_permissions SET updated_at = %s WHERE repo_id = 4`, clock().Add(-3*time.Minute)), // Meant to be excluded because it is public
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		m, err := s.Metrics(ctx, time.Minute)
		if err != nil {
			t.Fatal(err)
		}

		expMetrics := &PermsMetrics{
			UsersWithStalePerms:  1,
			UsersPermsGapSeconds: 60,
			ReposWithStalePerms:  1,
			ReposPermsGapSeconds: 120,
		}
		if diff := cmp.Diff(expMetrics, m); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	}
}

func setupTestPerms(t *testing.T, db database.DB, clock func() time.Time) *permsStore {
	t.Helper()
	logger := logtest.Scoped(t)
	s := perms(logger, db, clock)
	t.Cleanup(func() {
		cleanupPermsTables(t, s)
	})
	return s
}

func TestPermsStore_ListUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	tests := map[string]struct {
		unifiedPermissionsEnabled bool
	}{
		"unified permissions disabled": {
			unifiedPermissionsEnabled: false,
		},
		"unified permissions enabled": {
			unifiedPermissionsEnabled: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockUnifiedPermsConfig(tt.unifiedPermissionsEnabled)
			testDb := dbtest.NewDB(logger, t)
			db := database.NewDB(logger, testDb)
			s := perms(logger, db, clock)
			ctx := context.Background()

			t.Cleanup(func() {
				cleanupPermsTables(t, s)

				if t.Failed() {
					return
				}
				q := `TRUNCATE TABLE external_services, repo, users CASCADE`
				if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
					t.Fatal(err)
				}
			})
			// Set fake authz providers otherwise authz is bypassed
			authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
			defer authz.SetProviders(true, nil)
			// Set up some repositories and permissions
			qs := []*sqlf.Query{
				sqlf.Sprintf(`INSERT INTO users(id, username, site_admin) VALUES(555, 'user555', FALSE)`),
				sqlf.Sprintf(`INSERT INTO users(id, username, site_admin) VALUES(777, 'user777', TRUE)`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(1, 'private_repo_1', TRUE)`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(2, 'private_repo_2', TRUE)`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private, deleted_at) VALUES(3, 'private_repo_3_deleted', TRUE, NOW())`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(4, 'public_repo_4', FALSE)`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(5, 'public_repo_5', TRUE)`),
				sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config) VALUES(1, 'GitHub #1', 'GITHUB', '{}')`),
				sqlf.Sprintf(`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url)
                                 VALUES(1, 1, ''), (2, 1, ''), (3, 1, ''), (4, 1, '')`),
				sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, unrestricted) VALUES(2, 'GitHub #2 Unrestricted', 'GITHUB', '{}', TRUE)`),
				sqlf.Sprintf(`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url)
                                 VALUES(5, 2, '')`),
			}
			for _, q := range qs {
				if err := s.execute(ctx, q); err != nil {
					t.Fatal(err)
				}
			}
			if tt.unifiedPermissionsEnabled {
				q := sqlf.Sprintf(`INSERT INTO user_repo_permissions(user_id, repo_id) VALUES(555, 1), (777, 2), (555, 3), (777, 3);`)
				if err := s.execute(ctx, q); err != nil {
					t.Fatal(err)
				}
			} else {
				perms := []*authz.RepoPermissions{
					{
						RepoID:  1,
						Perm:    authz.Read,
						UserIDs: toMapset(555),
					}, {
						RepoID:  2,
						Perm:    authz.Read,
						UserIDs: toMapset(777),
					}, {
						RepoID:  3,
						Perm:    authz.Read,
						UserIDs: toMapset(555, 777),
					},
				}
				for _, perm := range perms {
					_, err := s.SetRepoPermissions(ctx, perm)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			tests := []listUserPermissionsTest{
				{
					Name:   "TestNonSiteAdminUser",
					UserID: 555,
					WantResults: []*listUserPermissionsResult{
						{
							// private repo but have access via user_permissions
							RepoId: 1,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
						{
							// public repo
							RepoId: 4,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							// private repo but unrestricted
							RepoId: 5,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
					},
				},
				{
					Name:   "TestPagination",
					UserID: 555,
					Args: &ListUserPermissionsArgs{
						PaginationArgs: &database.PaginationArgs{First: toIntPtr(2), After: toStringPtr("'public_repo_5'"), OrderBy: database.OrderBy{{Field: "repo.name"}}},
					},
					WantResults: []*listUserPermissionsResult{
						{
							RepoId: 4,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							RepoId: 1,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
					},
				},
				{
					Name:   "TestSearchQuery",
					UserID: 555,
					Args: &ListUserPermissionsArgs{
						Query: "repo_5",
					},
					WantResults: []*listUserPermissionsResult{
						{
							RepoId: 5,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
					},
				},
				{
					Name:   "TestSiteAdminUser",
					UserID: 777,
					WantResults: []*listUserPermissionsResult{
						{
							// do not have direct access but user is site admin
							RepoId: 1,
							Reason: UserRepoPermissionReasonSiteAdmin,
						},
						{
							// private repo but have access via user_permissions
							RepoId: 2,
							Reason: UserRepoPermissionReasonSiteAdmin,
						},
						{
							// public repo
							RepoId: 4,
							Reason: UserRepoPermissionReasonSiteAdmin,
						},
						{
							// private repo but unrestricted
							RepoId: 5,
							Reason: UserRepoPermissionReasonSiteAdmin,
						},
					},
				},
			}
			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) {
					results, err := s.ListUserPermissions(ctx, int32(test.UserID), test.Args)
					if err != nil {
						t.Fatal(err)
					}
					if len(test.WantResults) != len(results) {
						t.Fatalf("Results mismatch. Want: %d Got: %d", len(test.WantResults), len(results))
					}
					for index, result := range results {
						if diff := cmp.Diff(test.WantResults[index], &listUserPermissionsResult{RepoId: int32(result.Repo.ID), Reason: result.Reason}); diff != "" {
							t.Fatalf("Results (%d) mismatch (-want +got):\n%s", index, diff)
						}
					}
				})
			}
		})
	}
}

type listUserPermissionsTest struct {
	Name        string
	UserID      int
	Args        *ListUserPermissionsArgs
	WantResults []*listUserPermissionsResult
}

type listUserPermissionsResult struct {
	RepoId int32
	Reason UserRepoPermissionReason
}

func TestPermsStore_ListRepoPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	tests := map[string]struct {
		unifiedPermsEnabled bool
	}{
		"unified permissions disabled": {
			unifiedPermsEnabled: false,
		},
		"unified permissions enabled": {
			unifiedPermsEnabled: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			testDb := dbtest.NewDB(logger, t)
			db := database.NewDB(logger, testDb)

			s := perms(logtest.Scoped(t), db, clock)
			ctx := context.Background()
			t.Cleanup(func() {
				cleanupPermsTables(t, s)

				if t.Failed() {
					return
				}

				q := `TRUNCATE TABLE external_services, repo, users CASCADE`
				if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
					t.Fatal(err)
				}
			})

			// Set up some repositories and permissions
			qs := []*sqlf.Query{
				sqlf.Sprintf(`INSERT INTO users(id, username, site_admin) VALUES(555, 'user555', FALSE)`),
				sqlf.Sprintf(`INSERT INTO users(id, username, site_admin) VALUES(666, 'user666', FALSE)`),
				sqlf.Sprintf(`INSERT INTO users(id, username, site_admin) VALUES(777, 'user777', TRUE)`),
				sqlf.Sprintf(`INSERT INTO users(id, username, site_admin, deleted_at) VALUES(888, 'user888', TRUE, NOW())`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(1, 'private_repo_1', TRUE)`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(2, 'public_repo_2', FALSE)`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(3, 'unrestricted_repo_3', TRUE)`),
				sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(4, 'unrestricted_repo_4', TRUE)`),
				sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config) VALUES(1, 'GitHub #1', 'GITHUB', '{}')`),
				sqlf.Sprintf(`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url)
                                 VALUES(1, 1, ''), (2, 1, ''), (3, 1, '')`),
				sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, unrestricted) VALUES(2, 'GitHub #2 Unrestricted', 'GITHUB', '{}', TRUE)`),
				sqlf.Sprintf(`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url)
                                 VALUES(4, 2, '')`),
			}

			for _, q := range qs {
				if err := s.execute(ctx, q); err != nil {
					t.Fatal(err)
				}
			}

			if tt.unifiedPermsEnabled {
				q := sqlf.Sprintf(`INSERT INTO user_repo_permissions(user_id, repo_id) VALUES(555, 1), (666, 1), (NULL, 3), (666, 4)`)
				if err := s.execute(ctx, q); err != nil {
					t.Fatal(err)
				}
			} else {
				perms := []*authz.RepoPermissions{
					{
						// private repo
						RepoID: 1,
						Perm:   authz.Read,
						// non site-admin users
						UserIDs: toMapset(555, 666),
					}, {
						// private repo but unrestricted via perms_table
						RepoID:       3,
						Perm:         authz.Read,
						UserIDs:      toMapset(),
						Unrestricted: true,
					}, {
						// private repo but unrestricted via external service
						RepoID:  4,
						Perm:    authz.Read,
						UserIDs: toMapset(666),
					},
				}

				for _, perm := range perms {
					_, err := s.SetRepoPermissions(ctx, perm)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			tests := []listRepoPermissionsTest{
				{
					Name:   "TestPrivateRepo",
					RepoID: 1,
					Args:   nil,
					WantResults: []*listRepoPermissionsResult{
						{
							// do not have access but site-admin
							UserID: 777,
							Reason: UserRepoPermissionReasonSiteAdmin,
						},
						{
							// have access
							UserID: 666,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
						{
							// have access
							UserID: 555,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
					},
				},
				{
					Name:             "TestPrivateRepoWithNoAuthzProviders",
					RepoID:           1,
					Args:             nil,
					NoAuthzProviders: true,
					// all users have access
					WantResults: []*listRepoPermissionsResult{
						{
							UserID: 777,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 666,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 555,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
					},
				},
				{
					Name:   "TestPaginationWithPrivateRepo",
					RepoID: 1,
					Args: &ListRepoPermissionsArgs{
						PaginationArgs: &database.PaginationArgs{First: toIntPtr(1), After: toStringPtr("555"), OrderBy: database.OrderBy{{Field: "users.id"}}, Ascending: true},
					},
					WantResults: []*listRepoPermissionsResult{
						{
							UserID: 666,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
					},
				},
				{
					Name:   "TestSearchQueryWithPrivateRepo",
					RepoID: 1,
					Args: &ListRepoPermissionsArgs{
						Query: "6",
					},
					WantResults: []*listRepoPermissionsResult{
						{
							UserID: 666,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
					},
				},
				{
					Name:   "TestPublicRepo",
					RepoID: 2,
					Args:   nil,
					// all users have access
					WantResults: []*listRepoPermissionsResult{
						{
							UserID: 777,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 666,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 555,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
					},
				},
				{
					Name:   "TestUnrestrictedViaPermsTableRepo",
					RepoID: 3,
					Args:   nil,
					// all users have access
					WantResults: []*listRepoPermissionsResult{
						{
							UserID: 777,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 666,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 555,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
					},
				},
				{
					Name:   "TestUnrestrictedViaExternalServiceRepo",
					RepoID: 4,
					Args:   nil,
					// all users have access
					WantResults: []*listRepoPermissionsResult{
						{
							UserID: 777,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 666,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
						{
							UserID: 555,
							Reason: UserRepoPermissionReasonUnrestricted,
						},
					},
				},
				{
					Name:                      "TestUnrestrictedViaExternalServiceRepoWithoutPermsMapping",
					RepoID:                    4,
					Args:                      nil,
					NoAuthzProviders:          true,
					UsePermissionsUserMapping: true,
					// restricted access
					WantResults: []*listRepoPermissionsResult{
						{
							// do not have access but site-admin
							UserID: 777,
							Reason: UserRepoPermissionReasonSiteAdmin,
						},
						{
							// have access
							UserID: 666,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
					},
				},
				{
					Name:                      "TestPrivateRepoWithAuthzEnforceForSiteAdminsEnabled",
					RepoID:                    1,
					Args:                      nil,
					AuthzEnforceForSiteAdmins: true,
					WantResults: []*listRepoPermissionsResult{
						{
							// have access
							UserID: 666,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
						{
							// have access
							UserID: 555,
							Reason: UserRepoPermissionReasonPermissionsSync,
						},
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) {
					if !test.NoAuthzProviders {
						// Set fake authz providers otherwise authz is bypassed
						authz.SetProviders(false, []authz.Provider{&fakeProvider{}})
						defer authz.SetProviders(true, nil)
					}

					before := globals.PermissionsUserMapping()
					globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: test.UsePermissionsUserMapping})
					conf.Mock(
						&conf.Unified{
							SiteConfiguration: schema.SiteConfiguration{
								AuthzEnforceForSiteAdmins: test.AuthzEnforceForSiteAdmins,
								ExperimentalFeatures: &schema.ExperimentalFeatures{
									UnifiedPermissions: tt.unifiedPermsEnabled,
								},
							},
						},
					)
					t.Cleanup(func() {
						globals.SetPermissionsUserMapping(before)
						conf.Mock(nil)
					})

					results, err := s.ListRepoPermissions(ctx, api.RepoID(test.RepoID), test.Args)
					if err != nil {
						t.Fatal(err)
					}

					if len(test.WantResults) != len(results) {
						t.Fatalf("Results mismatch. Want: %d Got: %d", len(test.WantResults), len(results))
					}

					actualResults := make([]*listRepoPermissionsResult, 0, len(results))
					for _, result := range results {
						actualResults = append(actualResults, &listRepoPermissionsResult{UserID: result.User.ID, Reason: result.Reason})
					}

					if diff := cmp.Diff(test.WantResults, actualResults); diff != "" {
						t.Fatalf("Results mismatch (-want +got):\n%s", diff)
					}
				})
			}
		})
	}
}

type listRepoPermissionsTest struct {
	Name                      string
	RepoID                    int
	Args                      *ListRepoPermissionsArgs
	WantResults               []*listRepoPermissionsResult
	NoAuthzProviders          bool
	UsePermissionsUserMapping bool
	AuthzEnforceForSiteAdmins bool
}

type listRepoPermissionsResult struct {
	UserID int32
	Reason UserRepoPermissionReason
}

func toIntPtr(num int) *int {
	return &num
}

func toStringPtr(str string) *string {
	return &str
}

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

func (p *fakeProvider) ValidateConnection(context.Context) error { return nil }

func (p *fakeProvider) FetchUserPerms(context.Context, *extsvc.Account, authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, nil
}

func (p *fakeProvider) FetchUserPermsByToken(context.Context, string, authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, nil
}

func (p *fakeProvider) FetchRepoPerms(context.Context, *extsvc.Repository, authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, nil
}
