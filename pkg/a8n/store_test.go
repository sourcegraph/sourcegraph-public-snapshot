package a8n

import (
	"context"
	"flag"
	"fmt"
	"reflect"
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
	ctx := context.Background()

	t.Run("Campaigns", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		campaigns := make([]*Campaign, 3)
		for i := range campaigns {
			campaigns[i] = &Campaign{
				Name:        fmt.Sprintf("Upgrade ES-Lint %d", i),
				Description: "All the Javascripts are belong to us",
				AuthorID:    23,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if i%2 == 0 {
				campaigns[i].NamespaceOrgID = 23
			} else {
				campaigns[i].NamespaceUserID = 42
			}

			err := s.CreateCampaign(ctx, campaigns[i])
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			count, err := s.CountCampaigns(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(len(campaigns)); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		}

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

	t.Run("Threads", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		threads := make([]*Thread, 3)
		for i := range threads {
			threads[i] = &Thread{
				CampaignID: int64(i + 1),
				RepoID:     42,
				CreatedAt:  now,
				UpdatedAt:  now,
				Metadata:   "{}",
			}

			err := s.CreateThread(ctx, threads[i])
			if err != nil {
				t.Fatal(err)
			}
		}

		t.Run("Count-All", func(t *testing.T) {
			count, err := s.CountThreads(ctx, CountThreadsOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(len(threads)); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})

		t.Run("Count-Campaign", func(t *testing.T) {
			count, err := s.CountThreads(ctx, CountThreadsOpts{CampaignID: 1})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(1); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})

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
}
