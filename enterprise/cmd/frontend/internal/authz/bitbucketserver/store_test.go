package bitbucketserver

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
)

func testStore(db *sql.DB) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("nothing stored nor cached", transact(db, func(tx *sql.Tx) {
			now := time.Now().UTC()
			ttl := time.Second
			clock := func() time.Time { return now }

			s := newStore(tx, ttl, clock, newCache(ttl, clock))

			calls := 0
			ids := []uint32{1, 2, 3}
			update := func() ([]uint32, error) {
				calls++
				return ids, nil
			}

			ps := &Permissions{
				UserID: 0,
				Perm:   authz.Read,
				Type:   "repos",
			}

			ctx := context.Background()
			err := s.LoadPermissions(ctx, &ps, update)
			if err != nil {
				t.Fatal(err)
			}

			for _, id := range ids {
				if !ps.IDs.Contains(id) {
					t.Errorf("id %d missing from Permissions.IDs", id)
				}
			}
		}))
	}
}
