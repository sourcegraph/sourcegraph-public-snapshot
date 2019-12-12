package authz

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
)

func cleanup(t *testing.T, s *Store) {
	if t.Failed() {
		return
	}

	str := `DELETE FROM user_permissions;
DELETE FROM repo_permissions;
DELETE FROM user_pending_permissions;
DELETE FROM repo_pending_permissions;
`
	if err := s.execute(context.Background(), sqlf.Sprintf(str)); err != nil {
		t.Fatal(err)
	}
}

func equal(t testing.TB, name string, have, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("%q: %s", name, cmp.Diff(have, want))
	}
}

func array(bm *roaring.Bitmap) []uint32 {
	if bm == nil {
		return nil
	}
	return bm.ToArray()
}

func bitmap(ids ...uint32) *roaring.Bitmap {
	bm := roaring.NewBitmap()
	bm.AddMany(ids)
	return bm
}

var now = time.Now().Truncate(time.Microsecond).UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now)).Truncate(time.Microsecond)
}

func testStoreLoadUserPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching", func(t *testing.T) {
			s := NewStore(db, clock)
			defer cleanup(t, s)

			rp := &RepoPermissions{
				RepoID:   1,
				Perm:     authz.Read,
				UserIDs:  bitmap(2),
				Provider: ProviderSourcegraph,
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up := &UserPermissions{
				UserID:   1,
				Perm:     authz.Read,
				Type:     PermRepos,
				Provider: ProviderSourcegraph,
			}
			err := s.LoadUserPermissions(context.Background(), up)
			equal(t, "err", err, ErrNotFound)
			equal(t, "IDs", len(array(up.IDs)), 0)
		})

		t.Run("found matching", func(t *testing.T) {
			s := NewStore(db, clock)
			defer cleanup(t, s)

			rp := &RepoPermissions{
				RepoID:   1,
				Perm:     authz.Read,
				UserIDs:  bitmap(2),
				Provider: ProviderSourcegraph,
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up := &UserPermissions{
				UserID:   2,
				Perm:     authz.Read,
				Type:     PermRepos,
				Provider: ProviderSourcegraph,
			}
			if err := s.LoadUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", array(up.IDs), []uint32{1})
			equal(t, "UpdatedAt", up.UpdatedAt.UnixNano(), now)
		})

		t.Run("add and change", func(t *testing.T) {
			s := NewStore(db, clock)
			defer cleanup(t, s)

			rp := &RepoPermissions{
				RepoID:   1,
				Perm:     authz.Read,
				UserIDs:  bitmap(1, 2),
				Provider: ProviderSourcegraph,
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &RepoPermissions{
				RepoID:   1,
				Perm:     authz.Read,
				UserIDs:  bitmap(2, 3),
				Provider: ProviderSourcegraph,
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			up1 := &UserPermissions{
				UserID:   1,
				Perm:     authz.Read,
				Type:     PermRepos,
				Provider: ProviderSourcegraph,
			}
			if err := s.LoadUserPermissions(context.Background(), up1); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", len(array(up1.IDs)), 0)

			up2 := &UserPermissions{
				UserID:   2,
				Perm:     authz.Read,
				Type:     PermRepos,
				Provider: ProviderSourcegraph,
			}
			if err := s.LoadUserPermissions(context.Background(), up2); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", array(up2.IDs), []uint32{1})
			equal(t, "UpdatedAt", up2.UpdatedAt.UnixNano(), now)

			up3 := &UserPermissions{
				UserID:   3,
				Perm:     authz.Read,
				Type:     PermRepos,
				Provider: ProviderSourcegraph,
			}
			if err := s.LoadUserPermissions(context.Background(), up3); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", array(up3.IDs), []uint32{1})
			equal(t, "UpdatedAt", up3.UpdatedAt.UnixNano(), now)
		})
	}
}

func testStoreLoadRepoPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching", func(t *testing.T) {
			s := NewStore(db, time.Now)
			defer cleanup(t, s)

			rp := &RepoPermissions{
				RepoID:   1,
				Perm:     authz.Read,
				UserIDs:  bitmap(2),
				Provider: ProviderSourcegraph,
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &RepoPermissions{
				RepoID:   2,
				Perm:     authz.Read,
				Provider: ProviderSourcegraph,
			}
			err := s.LoadRepoPermissions(context.Background(), rp)
			equal(t, "err", err, ErrNotFound)
			equal(t, "rp.UserIDs", len(array(rp.UserIDs)), 0)
		})

		t.Run("found matching", func(t *testing.T) {
			s := NewStore(db, time.Now)
			defer cleanup(t, s)

			rp := &RepoPermissions{
				RepoID:   1,
				Perm:     authz.Read,
				UserIDs:  bitmap(2),
				Provider: ProviderSourcegraph,
			}
			if err := s.SetRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}

			rp = &RepoPermissions{
				RepoID:   1,
				Perm:     authz.Read,
				Provider: ProviderSourcegraph,
			}
			if err := s.LoadRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}
			equal(t, "rp.UserIDs", array(rp.UserIDs), []uint32{2})
		})
	}
}

func checkRegularTable(s *Store, sql string, expects map[int32][]uint32) error {
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

		objIDs := array(bm)
		if expects[id] == nil {
			return fmt.Errorf("unexpected row in table: (id: %v) -> (ids: %v)", id, objIDs)
		}

		have := fmt.Sprintf("%v", objIDs)
		want := fmt.Sprintf("%v", expects[id])
		if have != want {
			return fmt.Errorf("key %v want %v but got %v", id, want, have)
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

func testStoreSetRepoPermissions(db *sql.DB) func(*testing.T) {
	tests := []struct {
		name            string
		updates         []*RepoPermissions
		expectUserPerms map[int32][]uint32 // user_id -> object_ids
		expectRepoPerms map[int32][]uint32 // repo_id -> user_ids
	}{
		{
			name: "empty",
			updates: []*RepoPermissions{
				{
					RepoID:   1,
					Perm:     authz.Read,
					Provider: ProviderSourcegraph,
				},
			},
		},
		{
			name: "add",
			updates: []*RepoPermissions{
				{
					RepoID:   1,
					Perm:     authz.Read,
					UserIDs:  bitmap(1),
					Provider: ProviderSourcegraph,
				},
				{
					RepoID:   2,
					Perm:     authz.Read,
					UserIDs:  bitmap(1, 2),
					Provider: ProviderSourcegraph,
				},
				{
					RepoID:   3,
					Perm:     authz.Read,
					UserIDs:  bitmap(3, 4),
					Provider: ProviderSourcegraph,
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
			updates: []*RepoPermissions{
				{
					RepoID:   1,
					Perm:     authz.Read,
					UserIDs:  bitmap(1),
					Provider: ProviderSourcegraph,
				},
				{
					RepoID:   1,
					Perm:     authz.Read,
					UserIDs:  bitmap(2, 3),
					Provider: ProviderSourcegraph,
				},
				{
					RepoID:   2,
					Perm:     authz.Read,
					UserIDs:  bitmap(1, 2),
					Provider: ProviderSourcegraph,
				},
				{
					RepoID:   2,
					Perm:     authz.Read,
					UserIDs:  bitmap(3, 4),
					Provider: ProviderSourcegraph,
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
			updates: []*RepoPermissions{
				{
					RepoID:   1,
					Perm:     authz.Read,
					UserIDs:  bitmap(1, 2, 3),
					Provider: ProviderSourcegraph,
				},
				{
					RepoID:   1,
					Perm:     authz.Read,
					UserIDs:  bitmap(),
					Provider: ProviderSourcegraph,
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
				s := NewStore(db, clock)
				defer cleanup(t, s)

				for _, p := range test.updates {
					const numOps = 30
					var wg sync.WaitGroup
					wg.Add(numOps)
					for i := 0; i < numOps; i++ {
						go func() {
							defer wg.Done()
							tmp := &RepoPermissions{
								RepoID:    p.RepoID,
								Perm:      p.Perm,
								Provider:  p.Provider,
								UpdatedAt: p.UpdatedAt,
							}
							if p.UserIDs != nil {
								tmp.UserIDs = p.UserIDs.Clone()
							}
							if err := s.SetRepoPermissions(context.Background(), tmp); err != nil {
								t.Fatal(err)
							}
						}()
					}
					wg.Wait()
				}

				err := checkRegularTable(s, `SELECT user_id, object_ids FROM user_permissions`, test.expectUserPerms)
				if err != nil {
					t.Fatal("user_permissions:", err)
				}

				err = checkRegularTable(s, `SELECT repo_id, user_ids FROM repo_permissions`, test.expectRepoPerms)
				if err != nil {
					t.Fatal("repo_permissions:", err)
				}
			})
		}
	}
}

func testStoreLoadUserPendingPermissions(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("no matching", func(t *testing.T) {
			s := NewStore(db, clock)
			defer cleanup(t, s)

			bindIDs := []string{"bob"}
			rp := &RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), bindIDs, rp); err != nil {
				t.Fatal(err)
			}

			up := &UserPendingPermissions{
				BindID: "alice",
				Perm:   authz.Read,
				Type:   PermRepos,
			}
			err := s.LoadUserPendingPermissions(context.Background(), up)
			equal(t, "err", err, ErrNotFound)
			equal(t, "IDs", len(array(up.IDs)), 0)
		})

		t.Run("found matching", func(t *testing.T) {
			s := NewStore(db, clock)
			defer cleanup(t, s)

			bindIDs := []string{"alice"}
			rp := &RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), bindIDs, rp); err != nil {
				t.Fatal(err)
			}

			up := &UserPendingPermissions{
				BindID: "alice",
				Perm:   authz.Read,
				Type:   PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", array(up.IDs), []uint32{1})
			equal(t, "UpdatedAt", up.UpdatedAt.UnixNano(), now)
		})

		t.Run("add and change", func(t *testing.T) {
			s := NewStore(db, clock)
			defer cleanup(t, s)

			bindIDs := []string{"alice", "bob"}
			rp := &RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), bindIDs, rp); err != nil {
				t.Fatal(err)
			}

			bindIDs = []string{"bob", "cindy"}
			rp = &RepoPermissions{
				RepoID: 1,
				Perm:   authz.Read,
			}
			if err := s.SetRepoPendingPermissions(context.Background(), bindIDs, rp); err != nil {
				t.Fatal(err)
			}

			up1 := &UserPendingPermissions{
				BindID: "alice",
				Perm:   authz.Read,
				Type:   PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), up1); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", len(array(up1.IDs)), 0)

			up2 := &UserPendingPermissions{
				BindID: "bob",
				Perm:   authz.Read,
				Type:   PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), up2); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", array(up2.IDs), []uint32{1})
			equal(t, "UpdatedAt", up2.UpdatedAt.UnixNano(), now)

			up3 := &UserPendingPermissions{
				BindID: "cindy",
				Perm:   authz.Read,
				Type:   PermRepos,
			}
			if err := s.LoadUserPendingPermissions(context.Background(), up3); err != nil {
				t.Fatal(err)
			}
			equal(t, "IDs", array(up3.IDs), []uint32{1})
			equal(t, "UpdatedAt", up3.UpdatedAt.UnixNano(), now)
		})
	}
}

func checkUserPendingTable(ctx context.Context, s *Store, expects map[string][]uint32) (map[int32]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, bind_id, object_ids FROM user_pending_permissions`)
	if err != nil {
		return nil, err
	}

	// Collect id -> bind_id mappings for later use.
	bindIDs := make(map[int32]string)
	for rows.Next() {
		var id int32
		var bindID string
		var ids []byte
		if err := rows.Scan(&id, &bindID, &ids); err != nil {
			return nil, err
		}
		bindIDs[id] = bindID

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return nil, err
		}

		repoIDs := array(bm)
		if expects[bindID] == nil {
			return nil, fmt.Errorf("unexpected row in table: (bind_id: %v) -> (ids: %v)", bindID, repoIDs)
		}

		have := fmt.Sprintf("%v", repoIDs)
		want := fmt.Sprintf("%v", expects[bindID])
		if have != want {
			return nil, fmt.Errorf("bindID %q want %v but got %v", bindID, want, have)
		}
		delete(expects, bindID)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if len(expects) > 0 {
		return nil, fmt.Errorf("missing rows from table: %v", expects)
	}

	return bindIDs, nil
}

func checkRepoPendingTable(ctx context.Context, s *Store, bindIDs map[int32]string, expects map[int32][]string) error {
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

		userIDs := array(bm)
		if expects[id] == nil {
			return fmt.Errorf("unexpected row in table: (id: %v) -> (ids: %v)", id, userIDs)
		}

		haveBindIDs := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			bindID, ok := bindIDs[int32(userID)]
			if !ok {
				continue
			}

			haveBindIDs = append(haveBindIDs, bindID)
		}
		sort.Strings(haveBindIDs)

		have := fmt.Sprintf("%v", haveBindIDs)
		want := fmt.Sprintf("%v", expects[id])
		if have != want {
			return fmt.Errorf("id %d want %v but got %v", id, want, have)
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

func testStoreSetRepoPendingPermissions(db *sql.DB) func(*testing.T) {
	type update struct {
		bindIDs []string
		perm    *RepoPermissions
	}
	tests := []struct {
		name                   string
		updates                []update
		expectUserPendingPerms map[string][]uint32 // bind_id -> object_ids
		expectRepoPendingPerms map[int32][]string  // repo_id -> bind_ids
	}{
		{
			name: "empty",
			updates: []update{
				{
					bindIDs: nil,
					perm: &RepoPermissions{
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
					bindIDs: []string{"alice"},
					perm: &RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
				{
					bindIDs: []string{"alice", "bob"},
					perm: &RepoPermissions{
						RepoID: 2,
						Perm:   authz.Read,
					},
				},
				{
					bindIDs: []string{"cindy", "david"},
					perm: &RepoPermissions{
						RepoID: 3,
						Perm:   authz.Read,
					},
				},
			},
			expectUserPendingPerms: map[string][]uint32{
				"alice": {1, 2},
				"bob":   {2},
				"cindy": {3},
				"david": {3},
			},
			expectRepoPendingPerms: map[int32][]string{
				1: {"alice"},
				2: {"alice", "bob"},
				3: {"cindy", "david"},
			},
		},
		{
			name: "add and update",
			updates: []update{
				{
					bindIDs: []string{"alice", "bob"},
					perm: &RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
				{
					bindIDs: []string{"bob", "cindy"},
					perm: &RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
				{
					bindIDs: []string{"alice", "bob"},
					perm: &RepoPermissions{
						RepoID: 2,
						Perm:   authz.Read,
					},
				},
				{
					bindIDs: []string{"cindy", "david"},
					perm: &RepoPermissions{
						RepoID: 2,
						Perm:   authz.Read,
					},
				},
			},
			expectUserPendingPerms: map[string][]uint32{
				"alice": {},
				"bob":   {1},
				"cindy": {1, 2},
				"david": {2},
			},
			expectRepoPendingPerms: map[int32][]string{
				1: {"bob", "cindy"},
				2: {"cindy", "david"},
			},
		},
		{
			name: "add and clear",
			updates: []update{
				{
					bindIDs: []string{"alice", "bob", "cindy"},
					perm: &RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
				{
					bindIDs: []string{},
					perm: &RepoPermissions{
						RepoID: 1,
						Perm:   authz.Read,
					},
				},
			},
			expectUserPendingPerms: map[string][]uint32{
				"alice": {},
				"bob":   {},
				"cindy": {},
			},
			expectRepoPendingPerms: map[int32][]string{
				1: {},
			},
		},
	}

	return func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := NewStore(db, clock)
				defer cleanup(t, s)

				ctx := context.Background()

				for _, update := range test.updates {
					const numOps = 30
					var wg sync.WaitGroup
					wg.Add(numOps)
					for i := 0; i < numOps; i++ {
						go func() {
							defer wg.Done()
							tmp := &RepoPermissions{
								RepoID:    update.perm.RepoID,
								Perm:      update.perm.Perm,
								Provider:  update.perm.Provider,
								UpdatedAt: update.perm.UpdatedAt,
							}
							if update.perm.UserIDs != nil {
								tmp.UserIDs = update.perm.UserIDs.Clone()
							}
							if err := s.SetRepoPendingPermissions(ctx, update.bindIDs, tmp); err != nil {
								t.Fatal(err)
							}
						}()
					}
					wg.Wait()
				}

				// Query and check rows in "user_pending_permissions" table.
				bindIDs, err := checkUserPendingTable(ctx, s, test.expectUserPendingPerms)
				if err != nil {
					t.Fatal(err)
				}

				// Query and check rows in "repo_pending_permissions" table.
				err = checkRepoPendingTable(ctx, s, bindIDs, test.expectRepoPendingPerms)
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func testStoreGrantPendingPermissions(db *sql.DB) func(t *testing.T) {
	type pending struct {
		bindIDs []string
		perm    *RepoPermissions
	}
	type update struct {
		regulars []*RepoPermissions
		pendings []pending
	}
	type grant struct {
		userID int32
		perm   *UserPendingPermissions
	}
	tests := []struct {
		name                   string
		updates                []update
		grant                  grant
		expectUserPerms        map[int32][]uint32  // user_id -> object_ids
		expectRepoPerms        map[int32][]uint32  // repo_id -> user_ids
		expectUserPendingPerms map[string][]uint32 // bind_id -> object_ids
		expectRepoPendingPerms map[int32][]string  // repo_id -> bind_ids
	}{
		{
			name: "empty",
			grant: grant{
				userID: 1,
				perm: &UserPendingPermissions{
					BindID: "alice",
					Perm:   authz.Read,
					Type:   PermRepos,
				},
			},
		},
		{
			name: "no matching pending permissions",
			updates: []update{
				{
					regulars: []*RepoPermissions{
						{
							RepoID:   1,
							Perm:     authz.Read,
							UserIDs:  bitmap(1),
							Provider: ProviderSourcegraph,
						},
						{
							RepoID:   2,
							Perm:     authz.Read,
							UserIDs:  bitmap(1, 2),
							Provider: ProviderSourcegraph,
						},
					},
					pendings: []pending{
						{
							bindIDs: []string{"alice"},
							perm: &RepoPermissions{
								RepoID: 1,
								Perm:   authz.Read,
							},
						},
						{
							bindIDs: []string{"bob"},
							perm: &RepoPermissions{
								RepoID: 2,
								Perm:   authz.Read,
							},
						},
					},
				},
			},
			grant: grant{
				userID: 1,
				perm: &UserPendingPermissions{
					BindID: "cindy",
					Perm:   authz.Read,
					Type:   PermRepos,
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
			expectUserPendingPerms: map[string][]uint32{
				"alice": {1},
				"bob":   {2},
			},
			expectRepoPendingPerms: map[int32][]string{
				1: {"alice"},
				2: {"bob"},
			},
		},
		{
			name: "found matching pending permissions",
			updates: []update{
				{
					regulars: []*RepoPermissions{
						{
							RepoID:   1,
							Perm:     authz.Read,
							UserIDs:  bitmap(1),
							Provider: ProviderSourcegraph,
						},
						{
							RepoID:   2,
							Perm:     authz.Read,
							UserIDs:  bitmap(1, 2),
							Provider: ProviderSourcegraph,
						},
					},
					pendings: []pending{
						{
							bindIDs: []string{"alice"},
							perm: &RepoPermissions{
								RepoID: 1,
								Perm:   authz.Read,
							},
						},
						{
							bindIDs: []string{"bob"},
							perm: &RepoPermissions{
								RepoID: 2,
								Perm:   authz.Read,
							},
						},
					},
				},
			},
			grant: grant{
				userID: 3,
				perm: &UserPendingPermissions{
					BindID: "alice",
					Perm:   authz.Read,
					Type:   PermRepos,
				},
			},
			expectUserPerms: map[int32][]uint32{
				1: {1, 2},
				2: {2},
				3: {1},
			},
			expectRepoPerms: map[int32][]uint32{
				1: {1, 3},
				2: {1, 2},
			},
			expectUserPendingPerms: map[string][]uint32{
				"bob": {2},
			},
			expectRepoPendingPerms: map[int32][]string{
				1: {},
				2: {"bob"},
			},
		},
	}
	return func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := NewStore(db, clock)
				defer cleanup(t, s)

				ctx := context.Background()

				for _, update := range test.updates {
					for _, p := range update.regulars {
						if err := s.SetRepoPermissions(ctx, p); err != nil {
							t.Fatal(err)
						}
					}
					for _, p := range update.pendings {
						if err := s.SetRepoPendingPermissions(ctx, p.bindIDs, p.perm); err != nil {
							t.Fatal(err)
						}
					}
				}

				err := s.GrantPendingPermissions(ctx, test.grant.userID, test.grant.perm)
				if err != nil {
					t.Fatal(err)
				}

				err = checkRegularTable(s, `SELECT user_id, object_ids FROM user_permissions`, test.expectUserPerms)
				if err != nil {
					t.Fatal("user_permissions:", err)
				}

				err = checkRegularTable(s, `SELECT repo_id, user_ids FROM repo_permissions`, test.expectRepoPerms)
				if err != nil {
					t.Fatal("repo_permissions:", err)
				}

				// Query and check rows in "user_pending_permissions" table.
				bindIDs, err := checkUserPendingTable(ctx, s, test.expectUserPendingPerms)
				if err != nil {
					t.Fatal(err)
				}

				// Query and check rows in "repo_pending_permissions" table.
				err = checkRepoPendingTable(ctx, s, bindIDs, test.expectRepoPendingPerms)
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func testStoreDatabaseDeadlocks(db *sql.DB) func(t *testing.T) {
	return func(t *testing.T) {
		s := NewStore(db, time.Now)
		defer cleanup(t, s)

		ctx := context.Background()

		const numOps = 50
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			for i := 0; i < numOps; i++ {
				if err := s.SetRepoPermissions(ctx, &RepoPermissions{
					RepoID:   1,
					Perm:     authz.Read,
					UserIDs:  bitmap(1),
					Provider: ProviderSourcegraph,
				}); err != nil {
					t.Fatal(err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < numOps; i++ {
				if err := s.SetRepoPendingPermissions(ctx, []string{"alice"}, &RepoPermissions{
					RepoID:   1,
					Perm:     authz.Read,
					Provider: ProviderSourcegraph,
				}); err != nil {
					t.Fatal(err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < numOps; i++ {
				if err := s.GrantPendingPermissions(ctx, 1, &UserPendingPermissions{
					BindID: "alice",
					Perm:   authz.Read,
					Type:   PermRepos,
				}); err != nil &&
					!strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "user_permissions_perm_object_provider_unique"`) {
					t.Fatal(err)
				}
			}
		}()

		wg.Wait()
	}
}
