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
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func cleanupPermsTables(t *testing.T, s *permsStore) {
	if t.Failed() {
		return
	}

	q := `TRUNCATE TABLE user_permissions, repo_permissions, user_pending_permissions, repo_pending_permissions;`
	if err := s.execute(context.Background(), sqlf.Sprintf(q)); err != nil {
		t.Fatal(err)
	}
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

func testPermsStore_LoadUserPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching", func(t *testing.T) {
			s := perms(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up := &authz.UserPermissions{
				UserID: 1,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			err := s.LoadUserPermissions(context.Background(), up)
			if err != authz.ErrPermsNotFound {
				t.Fatalf("err: want %q but got %v", authz.ErrPermsNotFound, err)
			}
			equal(t, "IDs", 0, len(mapsetToArray(up.IDs)))
		})

		t.Run("found matching", func(t *testing.T) {
			s := perms(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			if err := s.LoadUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, mapsetToArray(up.IDs))
			equal(t, "UpdatedAt", now, up.UpdatedAt.UnixNano())

			if !up.SyncedAt.IsZero() {
				t.Fatal("SyncedAt was updated but not supposed to")
			}
		})

		t.Run("add and change", func(t *testing.T) {
			s := perms(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(1, 2),
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2, 3),
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up1 := &authz.UserPermissions{
				UserID: 1,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			if err := s.LoadUserPermissions(context.Background(), up1); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{}, mapsetToArray(up1.IDs))
			equal(t, "UpdatedAt", now, up1.UpdatedAt.UnixNano())

			if !up1.SyncedAt.IsZero() {
				t.Fatal("SyncedAt was updated but not supposed to")
			}

			up2 := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			if err := s.LoadUserPermissions(context.Background(), up2); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, mapsetToArray(up2.IDs))
			equal(t, "UpdatedAt", now, up2.UpdatedAt.UnixNano())

			if !up2.SyncedAt.IsZero() {
				t.Fatal("SyncedAt was updated but not supposed to")
			}

			up3 := &authz.UserPermissions{
				UserID: 3,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			if err := s.LoadUserPermissions(context.Background(), up3); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, mapsetToArray(up3.IDs))
			equal(t, "UpdatedAt", now, up3.UpdatedAt.UnixNano())

			if !up3.SyncedAt.IsZero() {
				t.Fatal("SyncedAt was updated but not supposed to")
			}
		})
	}
}

func testPermsStore_LoadRepoPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching", func(t *testing.T) {
			s := perms(db, time.Now)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			up := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			}
			if err := s.SetUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}

			rp := &authz.RepoPermissions{
				RepoID: 2,
				Perm:   authz.Read,
			}
			err := s.LoadRepoPermissions(context.Background(), rp)
			if err != authz.ErrPermsNotFound {
				t.Fatalf("err: want %q but got %q", authz.ErrPermsNotFound, err)
			}
			equal(t, "rp.UserIDs", []int{}, mapsetToArray(rp.UserIDs))
		})

		t.Run("found matching", func(t *testing.T) {
			s := perms(db, time.Now)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			up := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			}
			if err := s.SetUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}

			rp := &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.LoadRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}
			equal(t, "rp.UserIDs", []int{2}, mapsetToArray(rp.UserIDs))

			if !rp.SyncedAt.IsZero() {
				t.Fatal("SyncedAt was updated but not supposed to")
			}
		})
	}
}

func checkRegularPermsTable(s *permsStore, sql string, expects map[int32][]uint32) error {
	rows, err := s.Handle().DBUtilDB().QueryContext(context.Background(), sql)
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

func testPermsStore_SetUserPermissions(db database.DB) func(*testing.T) {
	const countToExceedParameterLimit = 17000 // ~ 65535 / 4 parameters per row

	tests := []struct {
		name            string
		slowTest        bool
		updates         []*authz.UserPermissions
		expectUserPerms map[int32][]uint32 // user_id -> object_ids
		expectRepoPerms map[int32][]uint32 // repo_id -> user_ids

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

	return func(t *testing.T) {
		t.Run("user-centric update should set synced_at", func(t *testing.T) {
			s := perms(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			up := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			}
			if err := s.SetUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}

			up = &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			if err := s.LoadUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}
			equal(t, "up.IDs", []int{1}, mapsetToArray(up.IDs))

			if up.SyncedAt.IsZero() {
				t.Fatal("SyncedAt was not updated but supposed to")
			}
		})

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.slowTest && !*slowTests {
					t.Skip("slow-tests not enabled")
				}

				if test.upsertRepoPermissionsPageSize > 0 {
					upsertRepoPermissionsPageSize = test.upsertRepoPermissionsPageSize
				}

				s := perms(db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
					if test.upsertRepoPermissionsPageSize > 0 {
						upsertRepoPermissionsPageSize = defaultUpsertRepoPermissionsPageSize
					}
				})

				for _, p := range test.updates {
					const numOps = 30
					g, ctx := errgroup.WithContext(context.Background())
					for i := 0; i < numOps; i++ {
						g.Go(func() error {
							tmp := &authz.UserPermissions{
								UserID:    p.UserID,
								Perm:      p.Perm,
								UpdatedAt: p.UpdatedAt,
							}
							if p.IDs != nil {
								tmp.IDs = p.IDs
							}
							return s.SetUserPermissions(ctx, tmp)
						})
					}
					if err := g.Wait(); err != nil {
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
}

func testPermsStore_SetRepoPermissionsUnrestricted(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		assertUnrestricted := func(ctx context.Context, t *testing.T, s *permsStore, id int32, want bool) {
			t.Helper()
			p := &authz.RepoPermissions{
				RepoID: id,
				Perm:   authz.Read,
			}
			if err := s.LoadRepoPermissions(ctx, p); err != nil {
				t.Fatalf("loading permissions for %d: %v", id, err)
			}
			if p.Unrestricted != want {
				t.Fatalf("Want %v, got %v for %d", want, p.Unrestricted, id)
			}
		}

		t.Run("test simple set", func(t *testing.T) {
			ctx := context.Background()
			s := setupTestPerms(t, db, clock)

			// Add a couple of repos
			for i := 0; i < 2; i++ {
				rp := &authz.RepoPermissions{
					RepoID:  int32(i + 1),
					Perm:    authz.Read,
					UserIDs: toMapset(2),
				}
				if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
					t.Fatal(err)
				}
			}
			assertUnrestricted(ctx, t, s, 1, false)
			assertUnrestricted(ctx, t, s, 2, false)

			// Set them both to unrestricted
			if err := s.SetRepoPermissionsUnrestricted(ctx, []int32{1, 2}, true); err != nil {
				t.Fatal(err)
			}
			assertUnrestricted(ctx, t, s, 1, true)
			assertUnrestricted(ctx, t, s, 2, true)

			// Set them back to false again, also checking that more than 65535 IDs
			// can be processed without an error
			var ids [66000]int32
			ids[0] = 1
			ids[65900] = 2
			if err := s.SetRepoPermissionsUnrestricted(ctx, ids[:], false); err != nil {
				t.Fatal(err)
			}
			assertUnrestricted(ctx, t, s, 1, false)
			assertUnrestricted(ctx, t, s, 2, false)
		})
	}
}

func testPermsStore_SetRepoPermissions(db database.DB) func(*testing.T) {
	tests := []struct {
		name            string
		updates         []*authz.RepoPermissions
		expectUserPerms map[int32][]uint32 // user_id -> object_ids
		expectRepoPerms map[int32][]uint32 // repo_id -> user_ids
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
		},
	}

	return func(t *testing.T) {
		t.Run("repo-centric update should set synced_at", func(t *testing.T) {
			s := perms(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.LoadRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}
			equal(t, "rp.UserIDs", []int{2}, mapsetToArray(rp.UserIDs))

			if rp.SyncedAt.IsZero() {
				t.Fatal("SyncedAt was not updated but supposed to")
			}
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
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &authz.RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.LoadRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}
			if rp.Unrestricted != true {
				t.Fatal("Want true")
			}
		})

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := perms(db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
				})

				for _, p := range test.updates {
					const numOps = 30
					g, ctx := errgroup.WithContext(context.Background())
					for i := 0; i < numOps; i++ {
						g.Go(func() error {
							tmp := &authz.RepoPermissions{
								RepoID:    p.RepoID,
								Perm:      p.Perm,
								UpdatedAt: p.UpdatedAt,
							}
							if p.UserIDs != nil {
								tmp.UserIDs = p.UserIDs
							}
							return s.SetRepoPermissions(ctx, tmp)
						})
					}
					if err := g.Wait(); err != nil {
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
}

func testPermsStore_TouchRepoPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		now := timeutil.Now().Unix()
		s := perms(db, func() time.Time {
			return time.Unix(atomic.LoadInt64(&now), 0)
		})
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		// Touch is an upsert
		if err := s.TouchRepoPermissions(context.Background(), 1); err != nil {
			t.Fatal(err)
		}

		// Set up some permissions
		rp := &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toMapset(2),
		}
		if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
			t.Fatal(err)
		}

		// Touch the permissions in an hour late
		now += 3600
		if err := s.TouchRepoPermissions(context.Background(), 1); err != nil {
			t.Fatal(err)
		}

		// Permissions bits shouldn't be affected
		rp = &authz.RepoPermissions{
			RepoID: 1,
			Perm:   authz.Read,
		}
		if err := s.LoadRepoPermissions(context.Background(), rp); err != nil {
			t.Fatal(err)
		}
		equal(t, "rp.UserIDs", []int{2}, mapsetToArray(rp.UserIDs))

		// Both times should be updated to "now"
		if rp.UpdatedAt.Unix() != now || rp.SyncedAt.Unix() != now {
			t.Fatal("UpdatedAt or SyncedAt was not updated but supposed to")
		}
	}
}

func testPermsStore_LoadUserPendingPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching with different account ID", func(t *testing.T) {
			s := perms(db, clock)
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
			s := perms(db, clock)
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
			s := perms(db, clock)
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
			s := perms(db, clock)
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
	rows, err := s.Handle().DBUtilDB().QueryContext(ctx, q)
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
	rows, err := s.Handle().DBUtilDB().QueryContext(ctx, `SELECT repo_id, user_ids_ints FROM repo_pending_permissions`)
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
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.slowTest && !*slowTests {
					t.Skip("slow-tests not enabled")
				}

				s := perms(db, clock)
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
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := perms(db, clock)
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

func testPermsStore_GrantPendingPermissions(db database.DB) func(*testing.T) {
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

	const countToExceedParameterLimit = 17000 // ~ 65535 / 4 parameters per row

	type pending struct {
		accounts *extsvc.Accounts
		perm     *authz.RepoPermissions
	}
	type update struct {
		regulars []*authz.RepoPermissions
		pendings []pending
	}
	type grant struct {
		userID int32
		perm   *authz.UserPendingPermissions
	}
	tests := []struct {
		name                   string
		slowTest               bool
		updates                []update
		grants                 []grant
		expectUserPerms        map[int32][]uint32              // user_id -> object_ids
		expectRepoPerms        map[int32][]uint32              // repo_id -> user_ids
		expectUserPendingPerms map[extsvc.AccountSpec][]uint32 // account -> object_ids
		expectRepoPendingPerms map[int32][]extsvc.AccountSpec  // repo_id -> accounts

		upsertRepoPermissionsPageSize int
	}{
		{
			name: "empty",
			grants: []grant{
				{
					userID: 1,
					perm: &authz.UserPendingPermissions{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						BindID:      "alice",
						Perm:        authz.Read,
						Type:        authz.PermRepos,
					},
				},
			},
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
			grants: []grant{
				{
					userID: 1,
					perm: &authz.UserPendingPermissions{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						BindID:      "cindy",
						Perm:        authz.Read,
						Type:        authz.PermRepos,
					},
				},
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
			grants: []grant{{
				userID: 1,
				perm: &authz.UserPendingPermissions{
					ServiceType: authz.SourcegraphServiceType,
					ServiceID:   authz.SourcegraphServiceID,
					BindID:      "alice",
					Perm:        authz.Read,
					Type:        authz.PermRepos,
				},
			}},
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
			grants: []grant{
				{
					userID: 3,
					perm: &authz.UserPendingPermissions{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						BindID:      "alice",
						Perm:        authz.Read,
						Type:        authz.PermRepos,
					},
				}, {
					userID: 3,
					perm: &authz.UserPendingPermissions{
						ServiceType: extsvc.TypeGitLab,
						ServiceID:   "https://gitlab.com/",
						BindID:      "alice",
						Perm:        authz.Read,
						Type:        authz.PermRepos,
					},
				},
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
			grants: []grant{
				{
					userID: 3,
					perm: &authz.UserPendingPermissions{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						BindID:      "alice@example.com",
						Perm:        authz.Read,
						Type:        authz.PermRepos,
					},
				}, {
					userID: 3,
					perm: &authz.UserPendingPermissions{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						BindID:      "alice2@example.com",
						Perm:        authz.Read,
						Type:        authz.PermRepos,
					},
				},
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
			name:                          "grant pending permission and page",
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
			grants: []grant{{
				userID: 1,
				perm: &authz.UserPendingPermissions{
					ServiceType: authz.SourcegraphServiceType,
					ServiceID:   authz.SourcegraphServiceID,
					BindID:      "alice",
					Perm:        authz.Read,
					Type:        authz.PermRepos,
				},
			}},
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
			grants: []grant{
				{
					userID: 1,
					perm: &authz.UserPendingPermissions{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						BindID:      "alice",
						Perm:        authz.Read,
						Type:        authz.PermRepos,
					},
				},
			},
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
	return func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.slowTest && !*slowTests {
					t.Skip("slow-tests not enabled")
				}

				if test.upsertRepoPermissionsPageSize > 0 {
					upsertRepoPermissionsPageSize = test.upsertRepoPermissionsPageSize
				}

				s := perms(db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
					if test.upsertRepoPermissionsPageSize > 0 {
						upsertRepoPermissionsPageSize = defaultUpsertRepoPermissionsPageSize
					}
				})

				ctx := context.Background()

				for _, update := range test.updates {
					for _, p := range update.regulars {
						if err := s.SetRepoPermissions(ctx, p); err != nil {
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
					err := s.GrantPendingPermissions(ctx, grant.userID, grant.perm)
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

// This test is used to ensure we ignore invalid pending user IDs on updating repository pending permissions
// because permissions have been granted for those users.
func testPermsStore_SetPendingPermissionsAfterGrant(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, clock)
		defer cleanupPermsTables(t, s)

		ctx := context.Background()

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
		if err := s.GrantPendingPermissions(ctx, 1, &authz.UserPendingPermissions{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			BindID:      "alice",
			Perm:        authz.Read,
			Type:        authz.PermRepos,
		}); err != nil {
			t.Fatal(err)
		}

		if err := s.GrantPendingPermissions(ctx, 1, &authz.UserPendingPermissions{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			BindID:      "bob",
			Perm:        authz.Read,
			Type:        authz.PermRepos,
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
}

func testPermsStore_DeleteAllUserPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		// Set permissions for user 1 and 2
		if err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toMapset(1, 2),
		}); err != nil {
			t.Fatal(err)
		}
		if err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  2,
			Perm:    authz.Read,
			UserIDs: toMapset(1, 2),
		}); err != nil {
			t.Fatal(err)
		}

		// Remove all permissions for the user=1
		if err := s.DeleteAllUserPermissions(ctx, 1); err != nil {
			t.Fatal(err)
		}

		// Check user=1 should not have any permissions now
		err := s.LoadUserPermissions(ctx, &authz.UserPermissions{
			UserID: 1,
			Perm:   authz.Read,
			Type:   authz.PermRepos,
		})
		if err != authz.ErrPermsNotFound {
			t.Fatalf("err: want %q but got %v", authz.ErrPermsNotFound, err)
		}

		// Check user=2 shoud not be affected
		p := &authz.UserPermissions{
			UserID: 2,
			Perm:   authz.Read,
			Type:   authz.PermRepos,
		}
		err = s.LoadUserPermissions(ctx, p)
		if err != nil {
			t.Fatal(err)
		}
		equal(t, "p.IDs", []int{1, 2}, mapsetToArray(p.IDs))
	}
}

func testPermsStore_DeleteAllUserPendingPermissions(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		accounts := &extsvc.Accounts{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountIDs:  []string{"alice", "bob"},
		}

		// Set pending permissions for alice and bob
		if err := s.SetRepoPendingPermissions(ctx, accounts, &authz.RepoPermissions{
			RepoID: 1,
			Perm:   authz.Read,
		}); err != nil {
			t.Fatal(err)
		}

		// Remove all pending permissions for alice
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

func testPermsStore_DatabaseDeadlocks(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, time.Now)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		setUserPermissions := func(ctx context.Context, t *testing.T) {
			if err := s.SetUserPermissions(ctx, &authz.UserPermissions{
				UserID: 1,
				Perm:   authz.Read,
				IDs:    toMapset(1),
			}); err != nil {
				t.Fatal(err)
			}
		}
		setRepoPermissions := func(ctx context.Context, t *testing.T) {
			if err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
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
			if err := s.GrantPendingPermissions(ctx, 1, &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      "alice",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
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
}

func cleanupUsersTable(t *testing.T, s *permsStore) {
	if t.Failed() {
		return
	}

	q := `TRUNCATE TABLE users RESTART IDENTITY CASCADE;`
	if err := s.execute(context.Background(), sqlf.Sprintf(q)); err != nil {
		t.Fatal(err)
	}
}

func testPermsStore_GetUserIDsByExternalAccounts(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, time.Now)
		t.Cleanup(func() {
			cleanupUsersTable(t, s)
		})

		ctx := context.Background()

		// Set up test users and external accounts
		extSQL := `
INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id, client_id, created_at, updated_at, deleted_at)
	VALUES(%s, %s, %s, %s, %s, %s, %s, %s)
`
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`), // ID=1
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),   // ID=2
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('cindy')`), // ID=3

			sqlf.Sprintf(extSQL, 1, extsvc.TypeGitLab, "https://gitlab.com/", "alice_gitlab", "alice_gitlab_client_id", clock(), clock(), nil), // ID=1
			sqlf.Sprintf(extSQL, 1, "github", "https://github.com/", "alice_github", "alice_github_client_id", clock(), clock(), nil),          // ID=2
			sqlf.Sprintf(extSQL, 2, extsvc.TypeGitLab, "https://gitlab.com/", "bob_gitlab", "bob_gitlab_client_id", clock(), clock(), nil),     // ID=3
			sqlf.Sprintf(extSQL, 3, extsvc.TypeGitLab, "https://gitlab.com/", "cindy_gitlab", "cindy_gitlab_client_id", clock(), clock(), nil), // ID=4
			sqlf.Sprintf(extSQL, 3, "github", "https://github.com/", "cindy_github", "cindy_github_client_id", clock(), clock(), clock()),      // ID=5, deleted
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

		if userIDs["alice_gitlab"] != 1 {
			t.Fatalf(`userIDs["alice_gitlab"]: want 1 but got %d`, userIDs["alice_gitlab"])
		} else if userIDs["bob_gitlab"] != 2 {
			t.Fatalf(`userIDs["bob_gitlab"]: want 2 but got %d`, userIDs["bob_gitlab"])
		}

		accounts = &extsvc.Accounts{
			ServiceType: "github",
			ServiceID:   "https://github.com/",
			AccountIDs:  []string{"cindy_github"},
		}
		userIDs, err = s.GetUserIDsByExternalAccounts(ctx, accounts)
		require.Nil(t, err)
		assert.Empty(t, userIDs)
	}
}

func testPermsStore_UserIDsWithOutdatedPerms(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, time.Now)
		ctx := context.Background()
		t.Cleanup(func() {
			cleanupUsersTable(t, s)
			cleanupPermsTables(t, s)

			q := `TRUNCATE TABLE external_services, orgs CASCADE`
			if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
				t.Fatal(err)
			}
		})

		// Create test users to include:
		//  1. A user with newer code host connection sync
		//  2. A user never had user-centric syncing
		//  3. A user with up-to-date permissions data
		//  4. A user with newer code host connection sync from the organization the user is a member of
		qs := []*sqlf.Query{
			// ID=1, with newer code host connection sync
			sqlf.Sprintf(`INSERT INTO users(username) VALUES ('alice')`),
			sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, namespace_user_id, last_sync_at) VALUES(1, 'GitHub #1', 'GITHUB', '{}', 1, NOW() + INTERVAL '10min')`),
			// ID=2, never had user-centric syncing
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),
			sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, namespace_user_id) VALUES(2, 'GitHub #2', 'GITHUB', '{}', 2)`),
			// ID=3, with up-to-date permissions data
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('cindy')`),
			sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, namespace_user_id) VALUES(3, 'GitHub #3', 'GITHUB', '{}', 3)`),
			// ID=4, with newer code host connection sync from the organization the user is a member of
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('david')`),
			sqlf.Sprintf(`INSERT INTO orgs(id, name) VALUES(1, 'david-org')`),
			sqlf.Sprintf(`INSERT INTO org_members(org_id, user_id) VALUES(1, 4)`),
			sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, namespace_org_id, last_sync_at) VALUES(4, 'GitHub #4', 'GITHUB', '{}', 1, NOW() + INTERVAL '10min')`),
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		// Give "alice" some permissions
		err := s.SetUserPermissions(ctx,
			&authz.UserPermissions{
				UserID: 1,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		// "bob" never had a user-centric syncing
		err = s.SetRepoPermissions(ctx,
			&authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(2),
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		// Give "cindy" some permissions
		err = s.SetUserPermissions(ctx,
			&authz.UserPermissions{
				UserID: 3,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		// Give "david" some permissions
		err = s.SetUserPermissions(ctx,
			&authz.UserPermissions{
				UserID: 4,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
				IDs:    toMapset(1),
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		// Both "alice" and "bob" have outdated permissions
		results, err := s.UserIDsWithOutdatedPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ids := make([]int32, 0, len(results))
		for id := range results {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

		expIDs := []int32{1, 2, 4}
		if diff := cmp.Diff(expIDs, ids); diff != "" {
			t.Fatal(diff)
		}
	}
}

func testPermsStore_UserIDsWithNoPerms(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, time.Now)
		t.Cleanup(func() {
			cleanupUsersTable(t, s)
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		// Create test users "alice" and "bob"
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`),                    // ID=1
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),                      // ID=2
			sqlf.Sprintf(`INSERT INTO users(username, deleted_at) VALUES('cindy', NOW())`), // ID=3
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		// Both "alice" and "bob" have no permissions
		ids, err := s.UserIDsWithNoPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

		expIDs := []int32{1, 2}
		if diff := cmp.Diff(expIDs, ids); diff != "" {
			t.Fatal(diff)
		}

		// Give "alice" some permissions
		err = s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toMapset(1),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Only "bob" has no permissions at this point
		ids, err = s.UserIDsWithNoPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}

		expIDs = []int32{2}
		if diff := cmp.Diff(expIDs, ids); diff != "" {
			t.Fatal(diff)
		}
	}
}

func cleanupReposTable(t *testing.T, s *permsStore) {
	if t.Failed() {
		return
	}

	q := `TRUNCATE TABLE repo RESTART IDENTITY CASCADE;`
	if err := s.execute(context.Background(), sqlf.Sprintf(q)); err != nil {
		t.Fatal(err)
	}
}

func testPermsStore_RepoIDsWithNoPerms(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, time.Now)
		t.Cleanup(func() {
			cleanupReposTable(t, s)
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		// Create three test repositories
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO repo(name, private) VALUES('private_repo', TRUE)`),                      // ID=1
			sqlf.Sprintf(`INSERT INTO repo(name) VALUES('public_repo')`),                                      // ID=2
			sqlf.Sprintf(`INSERT INTO repo(name, private) VALUES('private_repo_2', TRUE)`),                    // ID=3
			sqlf.Sprintf(`INSERT INTO repo(name, private, deleted_at) VALUES('private_repo_3', TRUE, NOW())`), // ID=4
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
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

		// Give "private_repo" regular permissions and "private_repo_2" pending permissions
		err = s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toMapset(1),
		})
		if err != nil {
			t.Fatal(err)
		}
		err = s.SetRepoPendingPermissions(ctx,
			&extsvc.Accounts{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				AccountIDs:  []string{"alice"},
			},
			&authz.RepoPermissions{
				RepoID: 3,
				Perm:   authz.Read,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		// No private repositories have any permissions at this point
		ids, err = s.RepoIDsWithNoPerms(ctx)
		if err != nil {
			t.Fatal(err)
		}

		expIDs = []api.RepoID{}
		if diff := cmp.Diff(expIDs, ids); diff != "" {
			t.Fatal(diff)
		}
	}
}

func testPermsStore_UserIDsWithOldestPerms(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, clock)
		ctx := context.Background()
		t.Cleanup(func() {
			cleanupPermsTables(t, s)

			if t.Failed() {
				return
			}

			if err := s.execute(ctx, sqlf.Sprintf(`DELETE FROM users`)); err != nil {
				t.Fatal(err)
			}
		})

		// Set up some users and permissions
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(id, username) VALUES(1, 'alice')`),
			sqlf.Sprintf(`INSERT INTO users(id, username) VALUES(2, 'bob')`),
			sqlf.Sprintf(`INSERT INTO users(id, username, deleted_at) VALUES(3, 'cindy', NOW())`),
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		// Set up some permissions
		err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toMapset(1, 2),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Mock user user 2's permissions to be synced in the future
		q := sqlf.Sprintf(`
UPDATE user_permissions
SET synced_at = %s
WHERE user_id = 2`, clock().AddDate(1, 0, 0))
		if err := s.execute(ctx, q); err != nil {
			t.Fatal(err)
		}

		// Should only get user 1 back (NULL FIRST)
		results, err := s.UserIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[int32]time.Time{1: {}}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}

		// Should get both users back
		results, err = s.UserIDsWithOldestPerms(ctx, 2, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults = map[int32]time.Time{
			1: {},
			2: clock().AddDate(1, 0, 0),
		}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}

		// Ignore users that have synced recently (or in the future)
		results, err = s.UserIDsWithOldestPerms(ctx, 5, 1*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		wantResults = map[int32]time.Time{
			1: {},
			// User 2 should be filtered out since it synced in the future
		}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}

		// Hard-delete user 2
		if err := s.execute(ctx, sqlf.Sprintf(`DELETE FROM users WHERE id = 2`)); err != nil {
			t.Fatal(err)
		}

		// Should only get user 1 back with limit=2
		results, err = s.UserIDsWithOldestPerms(ctx, 2, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults = map[int32]time.Time{1: {}}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatalf("Results mismatch (-want +got):\n%s", diff)
		}
	}
}

func testPermsStore_ReposIDsWithOldestPerms(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, clock)
		ctx := context.Background()
		t.Cleanup(func() {
			cleanupPermsTables(t, s)

			if t.Failed() {
				return
			}

			q := `TRUNCATE TABLE external_services, repo CASCADE`
			if err := s.execute(ctx, sqlf.Sprintf(q)); err != nil {
				t.Fatal(err)
			}
		})

		// Set up some repositories and permissions
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(1, 'private_repo_1', TRUE)`),
			sqlf.Sprintf(`INSERT INTO repo(id, name, private) VALUES(2, 'private_repo_2', TRUE)`),
			sqlf.Sprintf(`INSERT INTO repo(id, name, private, deleted_at) VALUES(3, 'private_repo_3', TRUE, NOW())`),
			sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config) VALUES(1, 'GitHub #1', 'GITHUB', '{}')`),
			sqlf.Sprintf(`INSERT INTO external_service_repos(repo_id, external_service_id, clone_url)
                                 VALUES(1, 1, ''), (2, 1, ''), (3, 1, '')`),
		}

		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		perms := []*authz.RepoPermissions{
			{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toMapset(1),
			}, {
				RepoID:  2,
				Perm:    authz.Read,
				UserIDs: toMapset(1),
			}, {
				RepoID:  3,
				Perm:    authz.Read,
				UserIDs: toMapset(1),
			},
		}
		for _, perm := range perms {
			err := s.SetRepoPermissions(ctx, perm)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Mock repo 2's permissions to be synced in the past
		q := sqlf.Sprintf(`
UPDATE repo_permissions
SET synced_at = %s
WHERE repo_id = 2`, clock().AddDate(-1, 0, 0))
		if err := s.execute(ctx, q); err != nil {
			t.Fatal(err)
		}

		// Should only get repo 1 back
		results, err := s.ReposIDsWithOldestPerms(ctx, 1, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults := map[api.RepoID]time.Time{2: clock().AddDate(-1, 0, 0)}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatalf("Results mismatch (-want +got):\n%s", diff)
		}

		// Should get two repos back
		results, err = s.ReposIDsWithOldestPerms(ctx, 2, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults = map[api.RepoID]time.Time{
			1: clock(),
			2: clock().AddDate(-1, 0, 0),
		}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatalf("Results mismatch (-want +got):\n%s", diff)
		}

		// Ignore repos that have synced recently (or in the future)
		results, err = s.ReposIDsWithOldestPerms(ctx, 2, 1*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		wantResults = map[api.RepoID]time.Time{
			// Only repo 2 should appear since it was synced a long time in the past
			2: clock().AddDate(-1, 0, 0),
		}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatal(diff)
		}

		// Hard-delete repo 2
		if err := s.execute(ctx, sqlf.Sprintf(`DELETE FROM repo WHERE id = 2`)); err != nil {
			t.Fatal(err)
		}

		// Should only get repo 1 back with limit=2
		results, err = s.ReposIDsWithOldestPerms(ctx, 2, 0)
		if err != nil {
			t.Fatal(err)
		}

		wantResults = map[api.RepoID]time.Time{1: clock()}
		if diff := cmp.Diff(wantResults, results); diff != "" {
			t.Fatalf("Results mismatch (-want +got):\n%s", diff)
		}
	}
}

func testPermsStore_UserIsMemberOfOrgHasCodeHostConnection(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, clock)
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

		// Set up users with:
		//  1. Is not a member of any organization
		//  2. Is a member of an organization without a code host connection
		//  3. Is a member of an organization with a code host connection
		users := db.Users()
		alice, err := users.Create(ctx,
			database.NewUser{
				Email:           "alice@example.com",
				Username:        "alice",
				EmailIsVerified: true,
			},
		)
		require.NoError(t, err)
		bob, err := users.Create(ctx,
			database.NewUser{
				Email:           "bob@example.com",
				Username:        "bob",
				EmailIsVerified: true,
			},
		)
		require.NoError(t, err)
		cindy, err := users.Create(ctx,
			database.NewUser{
				Email:           "cindy@example.com",
				Username:        "cindy",
				EmailIsVerified: true,
			},
		)
		require.NoError(t, err)

		orgs := db.Orgs()
		bobOrg, err := orgs.Create(ctx, "bob-org", nil)
		require.NoError(t, err)
		cindyOrg, err := orgs.Create(ctx, "cindy-org", nil)
		require.NoError(t, err)

		orgMembers := db.OrgMembers()
		_, err = orgMembers.Create(ctx, bobOrg.ID, bob.ID)
		require.NoError(t, err)
		_, err = orgMembers.Create(ctx, cindyOrg.ID, cindy.ID)
		require.NoError(t, err)

		err = db.ExternalServices().Create(ctx,
			func() *conf.Unified { return &conf.Unified{} },
			&types.ExternalService{
				Kind:           extsvc.KindGitHub,
				DisplayName:    "GitHub (cindy-org)",
				Config:         `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
				NamespaceOrgID: cindyOrg.ID,
			},
		)
		require.NoError(t, err)

		has, err := s.UserIsMemberOfOrgHasCodeHostConnection(ctx, alice.ID)
		assert.NoError(t, err)
		assert.False(t, has)

		has, err = s.UserIsMemberOfOrgHasCodeHostConnection(ctx, bob.ID)
		assert.NoError(t, err)
		assert.False(t, has)

		has, err = s.UserIsMemberOfOrgHasCodeHostConnection(ctx, cindy.ID)
		assert.NoError(t, err)
		assert.True(t, has)
	}
}

func testPermsStore_MapUsers(db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := perms(db, clock)
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
		s := perms(db, clock)

		ctx := context.Background()
		t.Cleanup(func() {
			cleanupPermsTables(t, s)

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
			err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
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
	s := perms(db, clock)
	t.Cleanup(func() {
		cleanupPermsTables(t, s)
	})
	return s
}
