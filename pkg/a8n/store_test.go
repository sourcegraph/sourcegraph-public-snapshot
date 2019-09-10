package a8n

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtest"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	d, cleanup := dbtest.NewDB(t, *dsn)
	defer cleanup()

	tx, done := dbtest.NewTx(t, d)
	defer done()

	s := NewStore(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)
	s.now = func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	ctx := context.Background()

	t.Run("Campaigns", func(t *testing.T) {
		campaigns := make([]*Campaign, 0, 3)

		t.Run("Create", func(t *testing.T) {
			for i := 0; i < cap(campaigns); i++ {
				c := &Campaign{
					Name:        fmt.Sprintf("Upgrade ES-Lint %d", i),
					Description: "All the Javascripts are belong to us",
					AuthorID:    23,
					ThreadIDs:   []int64{int64(i) + 1},
				}

				if i%2 == 0 {
					c.NamespaceOrgID = 23
				} else {
					c.NamespaceUserID = 42
				}

				want := c.Clone()
				have := c

				err := s.CreateCampaign(ctx, have)
				if err != nil {
					t.Fatal(err)
				}

				if have.ID == 0 {
					t.Fatal("ID should not be zero")
				}

				want.ID = have.ID
				want.CreatedAt = now
				want.UpdatedAt = now

				if !reflect.DeepEqual(have, want) {
					t.Fatal(cmp.Diff(have, want))
				}

				campaigns = append(campaigns, c)
			}
		})

		t.Run("Count", func(t *testing.T) {
			count, err := s.CountCampaigns(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(len(campaigns)); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})

		t.Run("List", func(t *testing.T) {
			for i := 1; i <= len(campaigns); i++ {
				cs, next, err := s.ListCampaigns(ctx, ListCampaignsOpts{Limit: i})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(campaigns) {
						want = campaigns[i].ID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := cs, campaigns[:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d campaigns, want: %d", len(have), len(want))
					}

					if !reflect.DeepEqual(have, want) {
						t.Fatal(cmp.Diff(have, want))
					}
				}
			}
		})

		t.Run("Update", func(t *testing.T) {
			for _, c := range campaigns {
				c.Name += "-updated"
				c.Description += "-updated"
				c.AuthorID++

				if c.NamespaceUserID != 0 {
					c.NamespaceUserID++
				}

				if c.NamespaceOrgID != 0 {
					c.NamespaceOrgID++
				}

				now = now.Add(time.Second)
				want := c
				want.UpdatedAt = now

				have := c.Clone()
				if err := s.UpdateCampaign(ctx, have); err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(have, want) {
					t.Fatal(cmp.Diff(have, want))
				}

				// Test that duplicates are not introduced.
				have.ThreadIDs = append(have.ThreadIDs, have.ThreadIDs...)
				if err := s.UpdateCampaign(ctx, have); err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(have, want) {
					t.Fatal(cmp.Diff(have, want))
				}

				// Test we can add to the set.
				have.ThreadIDs = append(have.ThreadIDs, 42)
				want.ThreadIDs = append(want.ThreadIDs, 42)

				if err := s.UpdateCampaign(ctx, have); err != nil {
					t.Fatal(err)
				}

				sort.Slice(have.ThreadIDs, func(a, b int) bool {
					return have.ThreadIDs[a] < have.ThreadIDs[b]
				})

				if !reflect.DeepEqual(have, want) {
					t.Fatal(cmp.Diff(have, want))
				}

				// Test we can remove from the set.
				have.ThreadIDs = have.ThreadIDs[:0]
				want.ThreadIDs = want.ThreadIDs[:0]

				if err := s.UpdateCampaign(ctx, have); err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(have, want) {
					t.Fatal(cmp.Diff(have, want))
				}
			}
		})
	})

	t.Run("Threads", func(t *testing.T) {
		threads := make([]*Thread, 0, 3)
		t.Run("Create", func(t *testing.T) {
			for i := 0; i < cap(threads); i++ {
				th := &Thread{
					RepoID:      42,
					CreatedAt:   now,
					UpdatedAt:   now,
					Metadata:    []byte("{}"),
					CampaignIDs: []int64{int64(i) + 1},
				}

				want := th.Clone()
				have := th

				err := s.CreateThread(ctx, have)
				if err != nil {
					t.Fatal(err)
				}

				if have.ID == 0 {
					t.Fatal("id should not be zero")
				}

				want.ID = have.ID
				want.CreatedAt = now
				want.UpdatedAt = now

				if !reflect.DeepEqual(have, want) {
					t.Fatal(cmp.Diff(have, want))
				}

				threads = append(threads, th)
			}
		})

		t.Run("Count", func(t *testing.T) {
			count, err := s.CountThreads(ctx, CountThreadsOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(len(threads)); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}

			count, err = s.CountThreads(ctx, CountThreadsOpts{CampaignID: 1})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(1); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})

		t.Run("List", func(t *testing.T) {
			for i := 1; i <= len(threads); i++ {
				opts := ListThreadsOpts{CampaignID: int64(i)}

				ts, next, err := s.ListThreads(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if have, want := next, int64(0); have != want {
					t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
				}

				have, want := ts, threads[i-1:i]
				if len(have) != len(want) {
					t.Fatalf("listed %d threads, want: %d", len(have), len(want))
				}

				if !reflect.DeepEqual(have, want) {
					t.Fatalf("opts: %+v, diff: %s", opts, cmp.Diff(have, want))
				}
			}

			for i := 1; i <= len(threads); i++ {
				ts, next, err := s.ListThreads(ctx, ListThreadsOpts{Limit: i})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(threads) {
						want = threads[i].ID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := ts, threads[:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d threads, want: %d", len(have), len(want))
					}

					if !reflect.DeepEqual(have, want) {
						t.Fatal(cmp.Diff(have, want))
					}
				}
			}
		})
	})
}
