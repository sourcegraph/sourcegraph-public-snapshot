package bitbucketserver

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
)

func BenchmarkStore(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()

	db, cleanup := dbtest.NewDB(b, *dsn)
	defer cleanup()

	ids := make([]uint32, 30000)
	for i := range ids {
		ids[i] = uint32(i)
	}

	c := extsvc.CodeHost{
		ServiceID:   "https://bitbucketserver.example.com",
		ServiceType: bitbucketserver.ServiceType,
	}

	update := func(context.Context) ([]uint32, *extsvc.CodeHost, error) {
		return ids, &c, nil
	}

	ctx := context.Background()

	b.Run("ttl=0", func(b *testing.B) {
		s := newStore(db, 0, DefaultHardTTL, clock)
		s.block = true

		ps := &iauthz.UserPermissions{
			UserID: 99,
			Perm:   authz.Read,
			Type:   "repos",
		}

		for i := 0; i < b.N; i++ {
			err := s.LoadPermissions(ctx, ps, update)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ttl=60s", func(b *testing.B) {
		s := newStore(db, 60*time.Second, DefaultHardTTL, clock)
		s.block = true

		ps := &iauthz.UserPermissions{
			UserID: 100,
			Perm:   authz.Read,
			Type:   "repos",
		}

		for i := 0; i < b.N; i++ {
			err := s.LoadPermissions(ctx, ps, update)
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
		now := time.Now().UTC().UnixNano()
		ttl := time.Second
		hardTTL := 10 * time.Second

		clock := func() time.Time {
			return time.Unix(0, atomic.LoadInt64(&now)).Truncate(time.Microsecond)
		}

		codeHost := extsvc.CodeHost{
			ServiceID:   "https://bitbucketserver.example.com",
			ServiceType: bitbucketserver.ServiceType,
		}

		rs := make([]*repos.Repo, 0, 7)
		for i := 1; i <= 7; i++ {
			rs = append(rs, &repos.Repo{
				Name: fmt.Sprintf("bitbucketserver.example.com/PROJ/%d", i),
				ExternalRepo: api.ExternalRepoSpec{
					ID:          strconv.Itoa(i),
					ServiceType: codeHost.ServiceType,
					ServiceID:   codeHost.ServiceID,
				},
				Sources: map[string]*repos.SourceInfo{},
			})
		}

		ctx := context.Background()
		err := repos.NewDBStore(db, sql.TxOptions{}).UpsertRepos(ctx, rs...)
		if err != nil {
			t.Fatal(err)
		}

		s := newStore(db, ttl, hardTTL, clock)
		s.updates = make(chan *iauthz.UserPermissions)

		ids := []uint32{1, 2, 3}
		e := error(nil)
		update := func(context.Context) ([]uint32, *extsvc.CodeHost, error) {
			return ids, &codeHost, e
		}

		ps := &iauthz.UserPermissions{UserID: 42, Perm: authz.Read, Type: "repos"}
		load := func(s *store) (*iauthz.UserPermissions, error) {
			ps := *ps
			return &ps, s.LoadPermissions(ctx, &ps, update)
		}

		array := func(ids *roaring.Bitmap) []uint32 {
			if ids == nil {
				return nil
			}
			return ids.ToArray()
		}

		{
			// No permissions cached.
			ps, err := load(s)
			equal(t, "err", err, &StalePermissionsError{UserPermissions: ps})
			equal(t, "ids", array(ps.IDs), []uint32(nil))
		}

		<-s.updates

		{
			// Hard TTL elapsed
			atomic.AddInt64(&now, int64(hardTTL))

			ps, err := load(s)
			equal(t, "err", err, &StalePermissionsError{UserPermissions: ps})
			equal(t, "ids", array(ps.IDs), ids)
		}

		<-s.updates

		ids = append(ids, 4, 5, 6)

		{
			// Source of truth changed (i.e. ids variable), but
			// cached permissions are not expired, so previous permissions
			// version is returned and no background update is started.
			ps, err := load(s)
			equal(t, "err", err, nil)
			equal(t, "ids", array(ps.IDs), ids[:3])
		}

		{
			// Cache expired, update called in the background, but stale
			// permissions are returned immediatelly.
			atomic.AddInt64(&now, int64(ttl))
			ps, err := load(s)
			equal(t, "err", err, nil)
			equal(t, "ids", array(ps.IDs), ids[:3])
		}

		// Wait for background update.
		<-s.updates

		{
			// Update is done, so we now have fresh permissions returned.
			ps, err := load(s)
			equal(t, "err", err, nil)
			equal(t, "ids", array(ps.IDs), ids)
		}

		ids = append(ids, 7)

		{
			// Cache expired, and source of truth changed. Here we test
			// that no concurrent updates are performed.
			atomic.AddInt64(&now, int64(2*ttl))

			delay := make(chan struct{})
			update = func(context.Context) ([]uint32, *extsvc.CodeHost, error) {
				<-delay
				return ids, &codeHost, e
			}

			type op struct {
				id  int
				ps  *iauthz.UserPermissions
				err error
			}

			ch := make(chan op, 30)
			updates := make(chan *iauthz.UserPermissions)

			for i := 0; i < cap(ch); i++ {
				go func(i int) {
					s := newStore(db, ttl, hardTTL, clock)
					s.updates = updates
					ps, err := load(s)
					ch <- op{i, ps, err}
				}(i)
			}

			results := make([]op, 0, cap(ch))
			for i := 0; i < cap(ch); i++ {
				results = append(results, <-ch)
			}

			for _, r := range results {
				equal(t, fmt.Sprintf("%d.err", r.id), r.err, nil)
				equal(t, fmt.Sprintf("%d.ids", r.id), array(r.ps.IDs), ids[:6])
			}

			close(delay)
			calls := 0
			timeout := time.After(500 * time.Millisecond)

		wait:
			for {
				select {
				case p := <-updates:
					calls++
					equal(t, "updated.ids", array(p.IDs), ids)
				case <-timeout:
					break wait
				}
			}

			equal(t, "updates", calls, 1)
		}
	}
}
