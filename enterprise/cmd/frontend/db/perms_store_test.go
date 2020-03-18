package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/gitchander/permutation"
	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"golang.org/x/sync/errgroup"
)

func cleanupPermsTables(t *testing.T, s *PermsStore) {
	if t.Failed() {
		return
	}

	q := `TRUNCATE TABLE user_permissions, repo_permissions, user_pending_permissions, repo_pending_permissions;`
	if err := s.execute(context.Background(), sqlf.Sprintf(q)); err != nil {
		t.Fatal(err)
	}
}

func bitmapToArray(bm *roaring.Bitmap) []int {
	if bm == nil {
		return nil
	}

	uint32s := bm.ToArray()
	ints := make([]int, len(uint32s))
	for i := range uint32s {
		ints[i] = int(uint32s[i])
	}
	return ints
}

func toBitmap(ids ...uint32) *roaring.Bitmap {
	bm := roaring.NewBitmap()
	bm.AddMany(ids)
	return bm
}

var now = time.Now().Truncate(time.Microsecond).UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now)).Truncate(time.Microsecond)
}

func testPermsStore_LoadUserPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching", func(t *testing.T) {
			s := NewPermsStore(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toBitmap(2),
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
			equal(t, "IDs", 0, len(bitmapToArray(up.IDs)))
		})

		t.Run("found matching", func(t *testing.T) {
			s := NewPermsStore(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toBitmap(2),
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
			equal(t, "IDs", []int{1}, bitmapToArray(up.IDs))
			equal(t, "UpdatedAt", now, up.UpdatedAt.UnixNano())
		})

		t.Run("add and change", func(t *testing.T) {
			s := NewPermsStore(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toBitmap(1, 2),
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toBitmap(2, 3),
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
			equal(t, "IDs", 0, len(bitmapToArray(up1.IDs)))

			up2 := &authz.UserPermissions{
				UserID: 2,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			if err := s.LoadUserPermissions(context.Background(), up2); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, bitmapToArray(up2.IDs))
			equal(t, "UpdatedAt", now, up2.UpdatedAt.UnixNano())

			up3 := &authz.UserPermissions{
				UserID: 3,
				Perm:   authz.Read,
				Type:   authz.PermRepos,
			}
			if err := s.LoadUserPermissions(context.Background(), up3); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", []int{1}, bitmapToArray(up3.IDs))
			equal(t, "UpdatedAt", now, up3.UpdatedAt.UnixNano())
		})
	}
}

func testPermsStore_LoadRepoPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching", func(t *testing.T) {
			s := NewPermsStore(db, time.Now)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toBitmap(2),
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &authz.RepoPermissions{
				RepoID: 2,
				Perm:   authz.Read,
			}
			err := s.LoadRepoPermissions(context.Background(), rp)
			if err != authz.ErrPermsNotFound {
				t.Fatalf("err: want %q but got %q", authz.ErrPermsNotFound, err)
			}
			equal(t, "rp.UserIDs", 0, len(bitmapToArray(rp.UserIDs)))
		})

		t.Run("found matching", func(t *testing.T) {
			s := NewPermsStore(db, time.Now)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			rp := &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toBitmap(2),
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
			equal(t, "rp.UserIDs", []int{2}, bitmapToArray(rp.UserIDs))
		})
	}
}

func checkRegularPermsTable(s *PermsStore, sql string, expects map[int32][]uint32) error {
	rows, err := s.db.QueryContext(context.Background(), sql)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id int32
		var ids []byte
		if err = rows.Scan(&id, &ids); err != nil {
			return err
		}

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return err
		}

		objIDs := bitmapToArray(bm)
		if expects[id] == nil {
			return fmt.Errorf("unexpected row in table: (id: %v) -> (ids: %v)", id, objIDs)
		}

		have := fmt.Sprintf("%v", objIDs)
		want := fmt.Sprintf("%v", expects[id])
		if have != want {
			return fmt.Errorf("key %v: want %q but got %q", id, want, have)
		}

		delete(expects, id)
	}

	if err = rows.Close(); err != nil {
		return err
	}

	if len(expects) > 0 {
		return fmt.Errorf("missing rows from table: %v", expects)
	}

	return nil
}

func testPermsStore_SetUserPermissions(db *sql.DB) func(*testing.T) {
	tests := []struct {
		name            string
		updates         []*authz.UserPermissions
		expectUserPerms map[int32][]uint32 // user_id -> object_ids
		expectRepoPerms map[int32][]uint32 // repo_id -> user_ids
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
					IDs:    toBitmap(1),
				}, {
					UserID: 2,
					Perm:   authz.Read,
					IDs:    toBitmap(1, 2),
				}, {
					UserID: 3,
					Perm:   authz.Read,
					IDs:    toBitmap(3, 4),
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
					IDs:    toBitmap(1),
				}, {
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toBitmap(2, 3),
				}, {
					UserID: 2,
					Perm:   authz.Read,
					IDs:    toBitmap(1, 2),
				}, {
					UserID: 2,
					Perm:   authz.Read,
					IDs:    toBitmap(1, 3),
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
					IDs:    toBitmap(1, 2, 3),
				}, {
					UserID: 1,
					Perm:   authz.Read,
					IDs:    toBitmap(),
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
	}

	return func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := NewPermsStore(db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
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
								tmp.IDs = p.IDs.Clone()
							}
							return s.SetUserPermissions(ctx, tmp)
						})
					}
					if err := g.Wait(); err != nil {
						t.Fatal(err)
					}
				}

				err := checkRegularPermsTable(s, `SELECT user_id, object_ids FROM user_permissions`, test.expectUserPerms)
				if err != nil {
					t.Fatal("user_permissions:", err)
				}

				err = checkRegularPermsTable(s, `SELECT repo_id, user_ids FROM repo_permissions`, test.expectRepoPerms)
				if err != nil {
					t.Fatal("repo_permissions:", err)
				}
			})
		}
	}
}

func testPermsStore_SetRepoPermissions(db *sql.DB) func(*testing.T) {
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
					UserIDs: toBitmap(1),
				}, {
					RepoID:  2,
					Perm:    authz.Read,
					UserIDs: toBitmap(1, 2),
				}, {
					RepoID:  3,
					Perm:    authz.Read,
					UserIDs: toBitmap(3, 4),
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
					UserIDs: toBitmap(1),
				}, {
					RepoID:  1,
					Perm:    authz.Read,
					UserIDs: toBitmap(2, 3),
				}, {
					RepoID:  2,
					Perm:    authz.Read,
					UserIDs: toBitmap(1, 2),
				}, {
					RepoID:  2,
					Perm:    authz.Read,
					UserIDs: toBitmap(3, 4),
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
					UserIDs: toBitmap(1, 2, 3),
				}, {
					RepoID:  1,
					Perm:    authz.Read,
					UserIDs: toBitmap(),
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
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := NewPermsStore(db, clock)
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
								tmp.UserIDs = p.UserIDs.Clone()
							}
							return s.SetRepoPermissions(ctx, tmp)
						})
					}
					if err := g.Wait(); err != nil {
						t.Fatal(err)
					}
				}

				err := checkRegularPermsTable(s, `SELECT user_id, object_ids FROM user_permissions`, test.expectUserPerms)
				if err != nil {
					t.Fatal("user_permissions:", err)
				}

				err = checkRegularPermsTable(s, `SELECT repo_id, user_ids FROM repo_permissions`, test.expectRepoPerms)
				if err != nil {
					t.Fatal("repo_permissions:", err)
				}
			})
		}
	}
}

func testPermsStore_LoadUserPendingPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching with different account ID", func(t *testing.T) {
			s := NewPermsStore(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.ExternalAccounts{
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
			equal(t, "IDs", 0, len(bitmapToArray(alice.IDs)))
		})

		t.Run("no matching with different service ID", func(t *testing.T) {
			s := NewPermsStore(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.ExternalAccounts{
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
				ServiceType: "gitlab",
				ServiceID:   "https://gitlab.com/",
				BindID:      "alice",
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}
			err := s.LoadUserPendingPermissions(context.Background(), alice)
			if err != authz.ErrPermsNotFound {
				t.Fatalf("err: want %q but got %q", authz.ErrPermsNotFound, err)
			}
			equal(t, "IDs", 0, len(bitmapToArray(alice.IDs)))
		})

		t.Run("found matching", func(t *testing.T) {
			s := NewPermsStore(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.ExternalAccounts{
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
			equal(t, "IDs", []int{1}, bitmapToArray(alice.IDs))
			equal(t, "UpdatedAt", now, alice.UpdatedAt.UnixNano())
		})

		t.Run("add and change", func(t *testing.T) {
			s := NewPermsStore(db, clock)
			t.Cleanup(func() {
				cleanupPermsTables(t, s)
			})

			accounts := &extsvc.ExternalAccounts{
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
			equal(t, "IDs", 0, len(bitmapToArray(alice.IDs)))

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
			equal(t, "IDs", []int{1}, bitmapToArray(bob.IDs))
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
			equal(t, "IDs", []int{1}, bitmapToArray(cindy.IDs))
			equal(t, "UpdatedAt", now, cindy.UpdatedAt.UnixNano())
		})
	}
}

func checkUserPendingPermsTable(
	ctx context.Context,
	s *PermsStore,
	expects map[extsvc.ExternalAccountSpec][]uint32,
) (
	idToSpecs map[int32]extsvc.ExternalAccountSpec,
	err error,
) {
	q := `SELECT id, service_type, service_id, bind_id, object_ids FROM user_pending_permissions`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	// Collect id -> account mappings for later used by checkRepoPendingPermsTable.
	idToSpecs = make(map[int32]extsvc.ExternalAccountSpec)
	for rows.Next() {
		var id int32
		var spec extsvc.ExternalAccountSpec
		var ids []byte
		if err := rows.Scan(&id, &spec.ServiceType, &spec.ServiceID, &spec.AccountID, &ids); err != nil {
			return nil, err
		}
		idToSpecs[id] = spec

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return nil, err
		}

		repoIDs := bitmapToArray(bm)
		if expects[spec] == nil {
			return nil, fmt.Errorf("unexpected row in table: (spec: %v) -> (ids: %v)", spec, repoIDs)
		}

		have := fmt.Sprintf("%v", repoIDs)
		want := fmt.Sprintf("%v", expects[spec])
		if have != want {
			return nil, fmt.Errorf("spec %q: want %q but got %q", spec, want, have)
		}
		delete(expects, spec)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if len(expects) > 0 {
		return nil, fmt.Errorf("missing rows from table: %v", expects)
	}

	return idToSpecs, nil
}

func checkRepoPendingPermsTable(
	ctx context.Context,
	s *PermsStore,
	idToSpecs map[int32]extsvc.ExternalAccountSpec,
	expects map[int32][]extsvc.ExternalAccountSpec,
) error {
	rows, err := s.db.QueryContext(ctx, `SELECT repo_id, user_ids FROM repo_pending_permissions`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id int32
		var ids []byte
		if err := rows.Scan(&id, &ids); err != nil {
			return err
		}

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return err
		}

		userIDs := bitmapToArray(bm)
		if expects[id] == nil {
			return fmt.Errorf("unexpected row in table: (id: %v) -> (ids: %v)", id, userIDs)
		}
		sort.Ints(userIDs)

		haveSpecs := make([]extsvc.ExternalAccountSpec, 0, len(userIDs))
		for _, userID := range userIDs {
			spec, ok := idToSpecs[int32(userID)]
			if !ok {
				continue
			}

			haveSpecs = append(haveSpecs, spec)
		}

		have := fmt.Sprintf("%v", haveSpecs)
		want := fmt.Sprintf("%v", expects[id])
		if have != want {
			return fmt.Errorf("id %d: want %q but got %q", id, want, have)
		}
		delete(expects, id)
	}

	if err = rows.Close(); err != nil {
		return err
	}

	if len(expects) > 0 {
		return fmt.Errorf("missing rows from table: %v", expects)
	}

	return nil
}

func testPermsStore_SetRepoPendingPermissions(db *sql.DB) func(*testing.T) {
	alice := extsvc.ExternalAccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "alice",
	}
	bob := extsvc.ExternalAccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "bob",
	}
	cindy := extsvc.ExternalAccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "cindy",
	}
	cindyGitHub := extsvc.ExternalAccountSpec{
		ServiceType: "github",
		ServiceID:   "https://github.com/",
		AccountID:   "cindy",
	}

	type update struct {
		accounts *extsvc.ExternalAccounts
		perm     *authz.RepoPermissions
	}
	tests := []struct {
		name                   string
		updates                []update
		expectUserPendingPerms map[extsvc.ExternalAccountSpec][]uint32 // account -> object_ids
		expectRepoPendingPerms map[int32][]extsvc.ExternalAccountSpec  // repo_id -> accounts
	}{
		{
			name: "empty",
			updates: []update{
				{
					accounts: &extsvc.ExternalAccounts{
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
					accounts: &extsvc.ExternalAccounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.ExternalAccounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice", "bob"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 2,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.ExternalAccounts{
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
			expectUserPendingPerms: map[extsvc.ExternalAccountSpec][]uint32{
				alice:       {1, 2},
				bob:         {2},
				cindyGitHub: {3},
			},
			expectRepoPendingPerms: map[int32][]extsvc.ExternalAccountSpec{
				1: {alice},
				2: {alice, bob},
				3: {cindyGitHub},
			},
		},
		{
			name: "add and update",
			updates: []update{
				{
					accounts: &extsvc.ExternalAccounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice", "bob"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.ExternalAccounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"bob", "cindy"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.ExternalAccounts{
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
			expectUserPendingPerms: map[extsvc.ExternalAccountSpec][]uint32{
				alice:       {},
				bob:         {1},
				cindy:       {1},
				cindyGitHub: {2},
			},
			expectRepoPendingPerms: map[int32][]extsvc.ExternalAccountSpec{
				1: {bob, cindy},
				2: {cindyGitHub},
			},
		},
		{
			name: "add and clear",
			updates: []update{
				{
					accounts: &extsvc.ExternalAccounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"alice", "bob", "cindy"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.ExternalAccounts{
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
			expectUserPendingPerms: map[extsvc.ExternalAccountSpec][]uint32{
				alice: {},
				bob:   {},
				cindy: {},
			},
			expectRepoPendingPerms: map[int32][]extsvc.ExternalAccountSpec{
				1: {},
			},
		},
	}

	return func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := NewPermsStore(db, clock)
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
							perm.UserIDs = update.perm.UserIDs.Clone()
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

func testPermsStore_ListPendingUsers(db *sql.DB) func(*testing.T) {
	type update struct {
		accounts *extsvc.ExternalAccounts
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
					accounts: &extsvc.ExternalAccounts{
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
					accounts: &extsvc.ExternalAccounts{
						ServiceType: authz.SourcegraphServiceType,
						ServiceID:   authz.SourcegraphServiceID,
						AccountIDs:  []string{"bob@example.com"},
					},
					perm: &authz.RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				}, {
					accounts: &extsvc.ExternalAccounts{
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
					accounts: &extsvc.ExternalAccounts{
						ServiceType: "gitlab",
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
				s := NewPermsStore(db, clock)
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
						tmp.UserIDs = update.perm.UserIDs.Clone()
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

func testPermsStore_GrantPendingPermissions(db *sql.DB) func(*testing.T) {
	alice := extsvc.ExternalAccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "alice",
	}
	bob := extsvc.ExternalAccountSpec{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountID:   "bob",
	}

	type pending struct {
		accounts *extsvc.ExternalAccounts
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
		updates                []update
		grants                 []grant
		expectUserPerms        map[int32][]uint32                      // user_id -> object_ids
		expectRepoPerms        map[int32][]uint32                      // repo_id -> user_ids
		expectUserPendingPerms map[extsvc.ExternalAccountSpec][]uint32 // account -> object_ids
		expectRepoPendingPerms map[int32][]extsvc.ExternalAccountSpec  // repo_id -> accounts
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
							UserIDs: toBitmap(1),
						}, {
							RepoID:  2,
							Perm:    authz.Read,
							UserIDs: toBitmap(1, 2),
						},
					},
					pendings: []pending{
						{
							accounts: &extsvc.ExternalAccounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"alice"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 1,
								Perm:   authz.Read,
							},
						}, {
							accounts: &extsvc.ExternalAccounts{
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
			expectUserPendingPerms: map[extsvc.ExternalAccountSpec][]uint32{
				alice: {1},
				bob:   {2},
			},
			expectRepoPendingPerms: map[int32][]extsvc.ExternalAccountSpec{
				1: {alice},
				2: {bob},
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
							UserIDs: toBitmap(1),
						}, {
							RepoID:  2,
							Perm:    authz.Read,
							UserIDs: toBitmap(1, 2),
						},
					},
					pendings: []pending{
						{
							accounts: &extsvc.ExternalAccounts{
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
							accounts: &extsvc.ExternalAccounts{
								ServiceType: "gitlab",
								ServiceID:   "https://gitlab.com/",
								AccountIDs:  []string{"alice"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 2,
								Perm:   authz.Read,
							},
						}, {
							accounts: &extsvc.ExternalAccounts{
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
						ServiceType: "gitlab",
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
			expectUserPendingPerms: map[extsvc.ExternalAccountSpec][]uint32{
				bob: {3},
			},
			expectRepoPendingPerms: map[int32][]extsvc.ExternalAccountSpec{
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
							UserIDs: toBitmap(1),
						}, {
							RepoID:  2,
							Perm:    authz.Read,
							UserIDs: toBitmap(1, 2),
						},
					},
					pendings: []pending{
						{
							accounts: &extsvc.ExternalAccounts{
								ServiceType: authz.SourcegraphServiceType,
								ServiceID:   authz.SourcegraphServiceID,
								AccountIDs:  []string{"alice@example.com"},
							},
							perm: &authz.RepoPermissions{
								RepoID: 1,
								Perm:   authz.Read,
							},
						}, {
							accounts: &extsvc.ExternalAccounts{
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
			expectUserPendingPerms: map[extsvc.ExternalAccountSpec][]uint32{},
			expectRepoPendingPerms: map[int32][]extsvc.ExternalAccountSpec{
				1: {},
				2: {},
			},
		},
	}
	return func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := NewPermsStore(db, clock)
				t.Cleanup(func() {
					cleanupPermsTables(t, s)
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

				err := checkRegularPermsTable(s, `SELECT user_id, object_ids FROM user_permissions`, test.expectUserPerms)
				if err != nil {
					t.Fatal("user_permissions:", err)
				}

				err = checkRegularPermsTable(s, `SELECT repo_id, user_ids FROM repo_permissions`, test.expectRepoPerms)
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

func testPermsStore_DeleteAllUserPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		// Set permissions for user 1 and 2
		if err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toBitmap(1, 2),
		}); err != nil {
			t.Fatal(err)
		}
		if err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  2,
			Perm:    authz.Read,
			UserIDs: toBitmap(1, 2),
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
		equal(t, "p.IDs", []int{1, 2}, bitmapToArray(p.IDs))
	}
}

func testPermsStore_DeleteAllUserPendingPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		accounts := &extsvc.ExternalAccounts{
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
		equal(t, "p.IDs", []int{1}, bitmapToArray(p.IDs))
	}
}

func testPermsStore_DatabaseDeadlocks(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, time.Now)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		setUserPermissions := func(ctx context.Context, t *testing.T) {
			if err := s.SetUserPermissions(ctx, &authz.UserPermissions{
				UserID: 1,
				Perm:   authz.Read,
				IDs:    toBitmap(1),
			}); err != nil {
				t.Fatal(err)
			}
		}
		setRepoPermissions := func(ctx context.Context, t *testing.T) {
			if err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
				RepoID:  1,
				Perm:    authz.Read,
				UserIDs: toBitmap(1),
			}); err != nil {
				t.Fatal(err)
			}
		}
		setRepoPendingPermissions := func(ctx context.Context, t *testing.T) {
			accounts := &extsvc.ExternalAccounts{
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

func cleanupUsersTable(t *testing.T, s *PermsStore) {
	if t.Failed() {
		return
	}

	q := `TRUNCATE TABLE users RESTART IDENTITY CASCADE;`
	if err := s.execute(context.Background(), sqlf.Sprintf(q)); err != nil {
		t.Fatal(err)
	}
}

func testPermsStore_ListExternalAccounts(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, time.Now)
		t.Cleanup(func() {
			cleanupUsersTable(t, s)
		})

		ctx := context.Background()

		// Set up test users and external accounts
		extSQL := `
INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id, client_id, created_at, updated_at)
	VALUES(%s, %s, %s, %s, %s, %s, %s)
`
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`), // ID=1
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),   // ID=2

			sqlf.Sprintf(extSQL, 1, "gitlab", "https://gitlab.com/", "alice_gitlab", "alice_gitlab_client_id", clock(), clock()), // ID=1
			sqlf.Sprintf(extSQL, 1, "github", "https://github.com/", "alice_github", "alice_github_client_id", clock(), clock()), // ID=2
			sqlf.Sprintf(extSQL, 2, "gitlab", "https://gitlab.com/", "bob_gitlab", "bob_gitlab_client_id", clock(), clock()),     // ID=3
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		{
			// Check external accounts for "alice"
			accounts, err := s.ListExternalAccounts(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}

			expAccounts := []*extsvc.ExternalAccount{
				{
					ID:     1,
					UserID: 1,
					ExternalAccountSpec: extsvc.ExternalAccountSpec{
						ServiceType: "gitlab",
						ServiceID:   "https://gitlab.com/",
						AccountID:   "alice_gitlab",
						ClientID:    "alice_gitlab_client_id",
					},
					CreatedAt: clock(),
					UpdatedAt: clock(),
				},
				{
					ID:     2,
					UserID: 1,
					ExternalAccountSpec: extsvc.ExternalAccountSpec{
						ServiceType: "github",
						ServiceID:   "https://github.com/",
						AccountID:   "alice_github",
						ClientID:    "alice_github_client_id",
					},
					CreatedAt: clock(),
					UpdatedAt: clock(),
				},
			}
			if diff := cmp.Diff(expAccounts, accounts); diff != "" {
				t.Fatalf(diff)
			}
		}

		{
			// Check external accounts for "bob"
			accounts, err := s.ListExternalAccounts(ctx, 2)
			if err != nil {
				t.Fatal(err)
			}

			expAccounts := []*extsvc.ExternalAccount{
				{
					ID:     3,
					UserID: 2,
					ExternalAccountSpec: extsvc.ExternalAccountSpec{
						ServiceType: "gitlab",
						ServiceID:   "https://gitlab.com/",
						AccountID:   "bob_gitlab",
						ClientID:    "bob_gitlab_client_id",
					},
					CreatedAt: clock(),
					UpdatedAt: clock(),
				},
			}
			if diff := cmp.Diff(expAccounts, accounts); diff != "" {
				t.Fatalf(diff)
			}
		}
	}
}

func testPermsStore_GetUserIDsByExternalAccounts(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, time.Now)
		t.Cleanup(func() {
			cleanupUsersTable(t, s)
		})

		ctx := context.Background()

		// Set up test users and external accounts
		extSQL := `
INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id, client_id, created_at, updated_at)
	VALUES(%s, %s, %s, %s, %s, %s, %s)
`
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`), // ID=1
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),   // ID=2
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('cindy')`), // ID=3

			sqlf.Sprintf(extSQL, 1, "gitlab", "https://gitlab.com/", "alice_gitlab", "alice_gitlab_client_id", clock(), clock()), // ID=1
			sqlf.Sprintf(extSQL, 1, "github", "https://github.com/", "alice_github", "alice_github_client_id", clock(), clock()), // ID=2
			sqlf.Sprintf(extSQL, 2, "gitlab", "https://gitlab.com/", "bob_gitlab", "bob_gitlab_client_id", clock(), clock()),     // ID=3
			sqlf.Sprintf(extSQL, 3, "gitlab", "https://gitlab.com/", "cindy_gitlab", "cindy_gitlab_client_id", clock(), clock()), // ID=4
		}
		for _, q := range qs {
			if err := s.execute(ctx, q); err != nil {
				t.Fatal(err)
			}
		}

		accounts := &extsvc.ExternalAccounts{
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
	}
}

func testPermsStore_UserIDsWithNoPerms(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, time.Now)
		t.Cleanup(func() {
			cleanupUsersTable(t, s)
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		// Create test users "alice" and "bob"
		qs := []*sqlf.Query{
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('alice')`), // ID=1
			sqlf.Sprintf(`INSERT INTO users(username) VALUES('bob')`),   // ID=2
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
			UserIDs: toBitmap(1),
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

func cleanupReposTable(t *testing.T, s *PermsStore) {
	if t.Failed() {
		return
	}

	q := `TRUNCATE TABLE repo RESTART IDENTITY CASCADE;`
	if err := s.execute(context.Background(), sqlf.Sprintf(q)); err != nil {
		t.Fatal(err)
	}
}

func testPermsStore_RepoIDsWithNoPerms(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, time.Now)
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
			UserIDs: toBitmap(1),
		})
		if err != nil {
			t.Fatal(err)
		}
		err = s.SetRepoPendingPermissions(ctx,
			&extsvc.ExternalAccounts{
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

func testPermsStore_UserIDsWithOldestPerms(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		// Set up some permissions
		err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toBitmap(1, 2),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Mock user user 2's permissions to be updated in the future
		q := sqlf.Sprintf(`
UPDATE user_permissions
SET updated_at = %s
WHERE user_id = 2`, clock().AddDate(1, 0, 0))
		if err := s.execute(ctx, q); err != nil {
			t.Fatal(err)
		}

		// Should only get user 1 back
		results, err := s.UserIDsWithOldestPerms(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}

		expResults := map[int32]time.Time{1: clock()}
		if diff := cmp.Diff(expResults, results); diff != "" {
			t.Fatal(diff)
		}

		// Should get both users back
		results, err = s.UserIDsWithOldestPerms(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}

		expResults = map[int32]time.Time{
			1: clock(),
			2: clock().AddDate(1, 0, 0),
		}
		if diff := cmp.Diff(expResults, results); diff != "" {
			t.Fatal(diff)
		}
	}
}

func testPermsStore_ReposIDsWithOldestPerms(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		s := NewPermsStore(db, clock)
		t.Cleanup(func() {
			cleanupPermsTables(t, s)
		})

		ctx := context.Background()

		// Set up some permissions
		err := s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  1,
			Perm:    authz.Read,
			UserIDs: toBitmap(1),
		})
		if err != nil {
			t.Fatal(err)
		}
		err = s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  2,
			Perm:    authz.Read,
			UserIDs: toBitmap(1),
		})
		if err != nil {
			t.Fatal(err)
		}
		err = s.SetRepoPermissions(ctx, &authz.RepoPermissions{
			RepoID:  3,
			Perm:    authz.Read,
			UserIDs: toBitmap(1),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Mock user repo 2's permissions to be updated in the future
		q := sqlf.Sprintf(`
UPDATE repo_permissions
SET updated_at = %s
WHERE repo_id = 2`, clock().AddDate(1, 0, 0))
		if err := s.execute(ctx, q); err != nil {
			t.Fatal(err)
		}

		// Mock repo 3 to be soft-deleted
		q = sqlf.Sprintf(`
INSERT INTO repo(id, name, private, deleted_at)
	VALUES(3, 'private_repo_3', TRUE, NOW())`)
		if err := s.execute(ctx, q); err != nil {
			t.Fatal(err)
		}

		// Should only get repo 1 back
		results, err := s.ReposIDsWithOldestPerms(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}

		expResults := map[api.RepoID]time.Time{1: clock()}
		if diff := cmp.Diff(expResults, results); diff != "" {
			t.Fatal(diff)
		}

		// Should get both repos back
		results, err = s.ReposIDsWithOldestPerms(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}

		expResults = map[api.RepoID]time.Time{
			1: clock(),
			2: clock().AddDate(1, 0, 0),
		}
		if diff := cmp.Diff(expResults, results); diff != "" {
			t.Fatal(diff)
		}
	}
}
