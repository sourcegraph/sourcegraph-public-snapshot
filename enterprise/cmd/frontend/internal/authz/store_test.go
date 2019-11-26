package authz

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
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

var now = time.Now().UTC().UnixNano()

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
			if err := s.LoadUserPermissions(context.Background(), up); err != nil {
				t.Fatal(err)
			}
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
			if err := s.LoadRepoPermissions(context.Background(), rp); err != nil {
				t.Fatal(err)
			}
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

func testStoreSetRepoPermissions(db *sql.DB) func(*testing.T) {
	tests := []struct {
		name            string
		updates         []*RepoPermissions
		expectUserPerms map[int32][]uint32
		expectRepoPerms map[int32][]uint32
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

	checkTable := func(s *Store, sql string, expects map[int32][]uint32) error {
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

			if expects[id] == nil {
				return fmt.Errorf("unexpected row in table: %v, %v", id, array(bm))
			}

			have := fmt.Sprintf("%v", array(bm))
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

	return func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := NewStore(db, clock)
				defer cleanup(t, s)

				for _, p := range test.updates {
					if err := s.SetRepoPermissions(context.Background(), p); err != nil {
						t.Fatal(err)
					}
				}

				err := checkTable(s, `SELECT user_id, object_ids FROM user_permissions`, test.expectUserPerms)
				if err != nil {
					t.Fatal("user_permissions", err)
				}

				err = checkTable(s, `SELECT repo_id, user_ids FROM repo_permissions`, test.expectRepoPerms)
				if err != nil {
					t.Fatal("repo_permissions", err)
				}
			})
		}
	}
}
