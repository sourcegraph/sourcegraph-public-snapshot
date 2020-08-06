package campaigns

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
)

func testStoreCampaigns(t *testing.T, ctx context.Context, s *Store, _ repos.Store, clock clock) {
	campaigns := make([]*cmpgn.Campaign, 0, 3)

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(campaigns); i++ {
			c := &cmpgn.Campaign{
				Name:           fmt.Sprintf("test-campaign-%d", i),
				Description:    "All the Javascripts are belong to us",
				Branch:         "upgrade-es-lint",
				AuthorID:       int32(i) + 50,
				ChangesetIDs:   []int64{int64(i) + 1},
				CampaignSpecID: 1742 + int64(i),
				ClosedAt:       clock.now(),
			}
			if i == 0 {
				// don't have associations for the first one
				c.CampaignSpecID = 0
				// Don't close the first one
				c.ClosedAt = time.Time{}
			}

			if i%2 == 0 {
				c.NamespaceOrgID = int32(i) + 23
			} else {
				c.NamespaceUserID = c.AuthorID
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
			want.CreatedAt = clock.now()
			want.UpdatedAt = clock.now()

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}

			campaigns = append(campaigns, c)
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountCampaigns(ctx, CountCampaignsOpts{})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, len(campaigns); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		count, err = s.CountCampaigns(ctx, CountCampaignsOpts{ChangesetID: 1})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, 1; have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		t.Run("OnlyForAuthor set", func(t *testing.T) {
			for _, c := range campaigns {
				count, err = s.CountCampaigns(ctx, CountCampaignsOpts{OnlyForAuthor: c.AuthorID})
				if err != nil {
					t.Fatal(err)
				}
				if have, want := count, 1; have != want {
					t.Fatalf("Incorrect number of campaigns counted, want=%d have=%d", want, have)
				}
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		for i := 1; i <= len(campaigns); i++ {
			opts := ListCampaignsOpts{ChangesetID: int64(i)}

			ts, next, err := s.ListCampaigns(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := next, int64(0); have != want {
				t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
			}

			have, want := ts, campaigns[i-1:i]
			if len(have) != len(want) {
				t.Fatalf("listed %d campaigns, want: %d", len(have), len(want))
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatalf("opts: %+v, diff: %s", opts, diff)
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

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		}

		{
			var cursor int64
			for i := 1; i <= len(campaigns); i++ {
				opts := ListCampaignsOpts{Cursor: cursor, Limit: 1}
				have, next, err := s.ListCampaigns(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				want := campaigns[i-1 : i]
				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		}

		filterTests := []struct {
			name  string
			state cmpgn.CampaignState
			want  []*cmpgn.Campaign
		}{
			{
				name:  "Any",
				state: cmpgn.CampaignStateAny,
				want:  campaigns,
			},
			{
				name:  "Closed",
				state: cmpgn.CampaignStateClosed,
				want:  campaigns[1:],
			},
			{
				name:  "Open",
				state: cmpgn.CampaignStateOpen,
				want:  campaigns[0:1],
			},
		}

		for _, tc := range filterTests {
			t.Run("ListCampaigns State "+tc.name, func(t *testing.T) {
				have, _, err := s.ListCampaigns(ctx, ListCampaignsOpts{State: tc.state})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Fatal(diff)
				}
			})
		}

		t.Run("ListCampaigns OnlyForAuthor set", func(t *testing.T) {
			for _, c := range campaigns {
				have, next, err := s.ListCampaigns(ctx, ListCampaignsOpts{OnlyForAuthor: c.AuthorID})
				if err != nil {
					t.Fatal(err)
				}
				if next != 0 {
					t.Fatal("Next value was true, but false expected")
				}
				if have, want := len(have), 1; have != want {
					t.Fatalf("Incorrect number of campaigns returned, want=%d have=%d", want, have)
				}
				if diff := cmp.Diff(have[0], c); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		for _, c := range campaigns {
			c.Name += "-updated"
			c.Description += "-updated"
			c.AuthorID++
			c.ClosedAt = c.ClosedAt.Add(5 * time.Second)

			if c.NamespaceUserID != 0 {
				c.NamespaceUserID++
			}

			if c.NamespaceOrgID != 0 {
				c.NamespaceOrgID++
			}

			clock.add(1 * time.Second)

			want := c
			want.UpdatedAt = clock.now()

			have := c.Clone()
			if err := s.UpdateCampaign(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}

			// Test that duplicates are not introduced.
			have.ChangesetIDs = append(have.ChangesetIDs, have.ChangesetIDs...)
			if err := s.UpdateCampaign(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}

			// Test we can add to the set.
			have.ChangesetIDs = append(have.ChangesetIDs, 42)
			want.ChangesetIDs = append(want.ChangesetIDs, 42)

			if err := s.UpdateCampaign(ctx, have); err != nil {
				t.Fatal(err)
			}

			sort.Slice(have.ChangesetIDs, func(a, b int) bool {
				return have.ChangesetIDs[a] < have.ChangesetIDs[b]
			})

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}

			// Test we can remove from the set.
			have.ChangesetIDs = have.ChangesetIDs[:0]
			want.ChangesetIDs = want.ChangesetIDs[:0]

			if err := s.UpdateCampaign(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			want := campaigns[0]
			opts := GetCampaignOpts{ID: want.ID}

			have, err := s.GetCampaign(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByCampaignSpecID", func(t *testing.T) {
			want := campaigns[0]
			opts := GetCampaignOpts{CampaignSpecID: want.CampaignSpecID}

			have, err := s.GetCampaign(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByName", func(t *testing.T) {
			want := campaigns[0]

			have, err := s.GetCampaign(ctx, GetCampaignOpts{Name: want.Name})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByNamespaceUserID", func(t *testing.T) {
			for _, c := range campaigns {
				if c.NamespaceUserID == 0 {
					continue
				}

				want := c
				opts := GetCampaignOpts{NamespaceUserID: c.NamespaceUserID}

				have, err := s.GetCampaign(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		})

		t.Run("ByNamespaceOrgID", func(t *testing.T) {
			for _, c := range campaigns {
				if c.NamespaceOrgID == 0 {
					continue
				}

				want := c
				opts := GetCampaignOpts{NamespaceOrgID: c.NamespaceOrgID}

				have, err := s.GetCampaign(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetCampaignOpts{ID: 0xdeadbeef}

			_, have := s.GetCampaign(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		for i := range campaigns {
			err := s.DeleteCampaign(ctx, campaigns[i].ID)
			if err != nil {
				t.Fatal(err)
			}

			count, err := s.CountCampaigns(ctx, CountCampaignsOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, len(campaigns)-(i+1); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		}
	})
}
