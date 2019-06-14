package bitbucketserver

import (
	"context"
	"database/sql"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtest"
)

func BenchmarkStore(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()

	db, cleanup := dbtest.NewDB(b, *dsn)
	defer cleanup()

	clock := func() time.Time { return time.Now().UTC() }
	ids := make([]uint32, 30000)
	for i := range ids {
		ids[i] = uint32(i)
	}

	update := func() ([]uint32, error) {
		time.Sleep(2 * time.Second) // Emulate slow code host
		return ids, nil
	}

	ctx := context.Background()

	b.Run("ttl=0", func(b *testing.B) {
		ttl := time.Duration(0)
		s := newStore(db, ttl, clock, newCache(ttl, clock))
		ps := &Permissions{
			UserID: 99,
			Perm:   authz.Read,
			Type:   "repos",
		}

		for i := 0; i < b.N; i++ {
			err := s.LoadPermissions(ctx, &ps, update)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ttl=60s/no-in-memory-cache", func(b *testing.B) {
		ttl := 60 * time.Second
		s := newStore(db, ttl, clock, nil)
		ps := &Permissions{
			UserID: 99,
			Perm:   authz.Read,
			Type:   "repos",
		}

		for i := 0; i < b.N; i++ {
			err := s.LoadPermissions(ctx, &ps, update)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ttl=60s/in-memory-cache", func(b *testing.B) {
		ttl := 60 * time.Second
		s := newStore(db, ttl, clock, newCache(ttl, clock))
		ps := &Permissions{
			UserID: 99,
			Perm:   authz.Read,
			Type:   "repos",
		}

		for i := 0; i < b.N; i++ {
			err := s.LoadPermissions(ctx, &ps, update)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func testStore(db *sql.DB) func(*testing.T) {
	equal := func(t testing.TB, name string, have, want interface{}) {
		t.Helper()
		if !reflect.DeepEqual(have, want) {
			t.Fatalf("%q: %s", name, cmp.Diff(have, want))
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

			type op struct {
				ps  *Permissions
				err error
			}

			ch := make(chan op, 25)
			for i := 1; i <= cap(ch); i++ {
				go func(i int) {
					s := newStore(db, ttl, clock, newCache(ttl, clock))
					ps, err := load(s)
					ch <- op{ps, err}
				}(i)
			}

			results := make([]op, 0, cap(ch))
			for i := 0; i < cap(ch); i++ {
				results = append(results, <-ch)
			}

			for _, r := range results {
				equal(t, "err", r.err, nil)
				equal(t, "ids", r.ps.IDs.ToArray(), ids)
			}

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
