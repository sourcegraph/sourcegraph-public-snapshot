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

	t.Parallel()

	d, cleanup := dbtest.NewDB(t, *dsn)
	defer cleanup()

	tx, done := dbtest.NewTx(t, d)
	defer done()

	s := NewStore(tx)
	ctx := context.Background()

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
}
