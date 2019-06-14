package bitbucketserver

import (
	"context"
	"database/sql"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
)

func testStore(db *sql.DB) func(*testing.T) {
	equal := func(t testing.TB, name string, have, want interface{}) {
		t.Helper()
		if !reflect.DeepEqual(have, want) {
			t.Errorf("%q: %s", name, cmp.Diff(have, want))
		}
	}

	return func(t *testing.T) {
		now := time.Now().UTC()
		ttl := time.Second
		clock := func() time.Time { return now }

		s := newStore(db, ttl, clock, newCache(ttl, clock))

		ids := []uint32{1, 2, 3}
		e := error(nil)
		calls := uint64(0)
		update := func() ([]uint32, error) {
			atomic.AddUint64(&calls, 1)
			return ids, e
		}

		ctx := context.Background()
		load := func(s *store) (*Permissions, error) {
			ps := &Permissions{
				UserID: 42,
				Perm:   authz.Read,
				Type:   "repos",
			}

			err := s.LoadPermissions(ctx, &ps, update)
			return ps, err
		}

		{
			// not cached nor stored
			ps, err := load(s)
			equal(t, "err", err, nil)
			equal(t, "ids", ps.IDs.ToArray(), ids)
		}

		ids = append(ids, 4, 5, 6)

		{
			// Still cached, no update should have happened even
			// if the permissions would have changed.
			ps, err := load(s)
			equal(t, "err", err, nil)
			equal(t, "ids", ps.IDs.ToArray(), ids[:3])
		}

		{
			// Not cached but stored.  Clear in-memory cache, still
			// no update should have happened, but permissions get
			// loaded from the store.
			s.cache.cache = map[cacheKey]*Permissions{}
			ps, err := load(s)
			equal(t, "err", err, nil)
			equal(t, "ids", ps.IDs.ToArray(), ids[:3])
		}

		{
			// Cache expired, update called
			now = now.Add(ttl)
			ps, err := load(s)
			equal(t, "err", err, nil)
			equal(t, "ids", ps.IDs.ToArray(), ids)
		}

		ids = ids[:3]

		{
			// Cache expired, no concurrent updates
			now = now.Add(2 * ttl)

			want := atomic.LoadUint64(&calls) + 1

			var wg sync.WaitGroup
			for i := 1; i <= 25; i++ {
				wg.Add(1)
				go func(i int) {
					s := newStore(db, ttl, clock, newCache(ttl, clock))
					defer wg.Done()
					ps, err := load(s)
					equal(t, "err", err, nil)
					equal(t, "ids", ps.IDs.ToArray(), ids)
				}(i)
			}

			wg.Wait()

			equal(t, "updates", atomic.LoadUint64(&calls), want)
		}

		{
			// Cache expired, error updating
			now = now.Add(ttl)
			e = errors.New("boom")
			ps, err := load(s)
			equal(t, "err", err, e)
			equal(t, "ids", ps.IDs, (*roaring.Bitmap)(nil))
		}
	}
}
