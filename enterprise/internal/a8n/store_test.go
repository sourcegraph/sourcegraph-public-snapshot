package a8n

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// Ran in integration_test.go
func testStore(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		tx, done := dbtest.NewTx(t, db)
		defer done()

		now := time.Now().UTC().Truncate(time.Microsecond)
		s := NewStoreWithClock(tx, func() time.Time {
			return now.UTC().Truncate(time.Microsecond)
		})

		ctx := context.Background()

		t.Run("Campaigns", func(t *testing.T) {
			campaigns := make([]*a8n.Campaign, 0, 3)

			t.Run("Create", func(t *testing.T) {
				for i := 0; i < cap(campaigns); i++ {
					c := &a8n.Campaign{
						Name:           fmt.Sprintf("Upgrade ES-Lint %d", i),
						Description:    "All the Javascripts are belong to us",
						AuthorID:       23,
						ChangesetIDs:   []int64{int64(i) + 1},
						CampaignPlanID: 42,
						ClosedAt:       now,
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

				if have, want := count, int64(len(campaigns)); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}

				count, err = s.CountCampaigns(ctx, CountCampaignsOpts{ChangesetID: 1})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(1); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}
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

					now = now.Add(time.Second)
					want := c
					want.UpdatedAt = now

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

					if have, want := count, int64(len(campaigns)-(i+1)); have != want {
						t.Fatalf("have count: %d, want: %d", have, want)
					}
				}
			})

		})

		t.Run("Changesets", func(t *testing.T) {
			githubActor := github.Actor{
				AvatarURL: "https://avatars2.githubusercontent.com/u/1185253",
				Login:     "mrnugget",
				URL:       "https://github.com/mrnugget",
			}
			githubPR := &github.PullRequest{
				ID:           "FOOBARID",
				Title:        "Fix a bunch of bugs",
				Body:         "This fixes a bunch of bugs",
				URL:          "https://github.com/sourcegraph/sourcegraph/pull/12345",
				Number:       12345,
				Author:       githubActor,
				Participants: []github.Actor{githubActor},
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			changesets := make([]*a8n.Changeset, 0, 3)

			t.Run("Create", func(t *testing.T) {
				for i := 0; i < cap(changesets); i++ {
					th := &a8n.Changeset{
						RepoID:              42,
						CreatedAt:           now,
						UpdatedAt:           now,
						Metadata:            githubPR,
						CampaignIDs:         []int64{int64(i) + 1},
						ExternalID:          fmt.Sprintf("foobar-%d", i),
						ExternalServiceType: "github",
					}

					changesets = append(changesets, th)
				}

				err := s.CreateChangesets(ctx, changesets...)
				if err != nil {
					t.Fatal(err)
				}

				for _, have := range changesets {
					if have.ID == 0 {
						t.Fatal("id should not be zero")
					}

					want := have.Clone()

					want.ID = have.ID
					want.CreatedAt = now
					want.UpdatedAt = now

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			})

			t.Run("CreateAlreadyExistingChangesets", func(t *testing.T) {
				ids := make([]int64, len(changesets))
				for i, c := range changesets {
					ids[i] = c.ID
				}

				clones := make([]*a8n.Changeset, len(changesets))

				for i, c := range changesets {
					// Set only the fields on which we have a unique constraint
					clones[i] = &a8n.Changeset{
						RepoID:              c.RepoID,
						ExternalID:          c.ExternalID,
						ExternalServiceType: c.ExternalServiceType,
					}
				}

				// Advance clock so store can determine whether Changeset was
				// inserted or not
				now = now.Add(time.Second)

				err := s.CreateChangesets(ctx, clones...)
				ae, ok := err.(AlreadyExistError)
				if !ok {
					t.Fatal(err)
				}

				{
					sort.Slice(ae.ChangesetIDs, func(i, j int) bool { return ae.ChangesetIDs[i] < ae.ChangesetIDs[j] })
					sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

					have, want := ae.ChangesetIDs, ids
					if len(have) != len(want) {
						t.Fatalf("%d changesets already exist, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}

				{
					// Verify that we got the original changesets back
					have, want := clones, changesets
					if len(have) != len(want) {
						t.Fatalf("created %d changesets, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			})

			t.Run("Count", func(t *testing.T) {
				count, err := s.CountChangesets(ctx, CountChangesetsOpts{})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(len(changesets)); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}

				count, err = s.CountChangesets(ctx, CountChangesetsOpts{CampaignID: 1})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(1); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}
			})

			t.Run("List", func(t *testing.T) {
				for i := 1; i <= len(changesets); i++ {
					opts := ListChangesetsOpts{CampaignID: int64(i)}

					ts, next, err := s.ListChangesets(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if have, want := next, int64(0); have != want {
						t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
					}

					have, want := ts, changesets[i-1:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d changesets, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatalf("opts: %+v, diff: %s", opts, diff)
					}
				}

				for i := 1; i <= len(changesets); i++ {
					ts, next, err := s.ListChangesets(ctx, ListChangesetsOpts{Limit: i})
					if err != nil {
						t.Fatal(err)
					}

					{
						have, want := next, int64(0)
						if i < len(changesets) {
							want = changesets[i].ID
						}

						if have != want {
							t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
						}
					}

					{
						have, want := ts, changesets[:i]
						if len(have) != len(want) {
							t.Fatalf("listed %d changesets, want: %d", len(have), len(want))
						}

						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatal(diff)
						}
					}
				}

				{
					ids := make([]int64, len(changesets))
					for i := range changesets {
						ids[i] = changesets[i].ID
					}

					have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{IDs: ids})
					if err != nil {
						t.Fatal(err)
					}

					want := changesets
					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}

				{
					var cursor int64
					for i := 1; i <= len(changesets); i++ {
						opts := ListChangesetsOpts{Cursor: cursor, Limit: 1}
						have, next, err := s.ListChangesets(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						want := changesets[i-1 : i]
						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatalf("opts: %+v, diff: %s", opts, diff)
						}

						cursor = next
					}
				}
			})

			t.Run("Get", func(t *testing.T) {
				t.Run("ByID", func(t *testing.T) {
					want := changesets[0]
					opts := GetChangesetOpts{ID: want.ID}

					have, err := s.GetChangeset(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("ByExternalID", func(t *testing.T) {
					want := changesets[0]
					opts := GetChangesetOpts{
						ExternalID:          want.ExternalID,
						ExternalServiceType: want.ExternalServiceType,
					}

					have, err := s.GetChangeset(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("NoResults", func(t *testing.T) {
					opts := GetChangesetOpts{ID: 0xdeadbeef}

					_, have := s.GetChangeset(ctx, opts)
					want := ErrNoResults

					if have != want {
						t.Fatalf("have err %v, want %v", have, want)
					}
				})
			})

			t.Run("Update", func(t *testing.T) {
				want := make([]*a8n.Changeset, 0, len(changesets))
				have := make([]*a8n.Changeset, 0, len(changesets))

				now = now.Add(time.Second)
				for _, c := range changesets {
					c.Metadata = &bitbucketserver.PullRequest{ID: 1234}
					c.ExternalServiceType = bitbucketserver.ServiceType

					if c.RepoID != 0 {
						c.RepoID++
					}

					have = append(have, c.Clone())

					c.UpdatedAt = now
					want = append(want, c)
				}

				if err := s.UpdateChangesets(ctx, have...); err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}

				for i := range have {
					// Test that duplicates are not introduced.
					have[i].CampaignIDs = append(have[i].CampaignIDs, have[i].CampaignIDs...)
				}

				if err := s.UpdateChangesets(ctx, have...); err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}

				for i := range have {
					// Test we can add to the set.
					have[i].CampaignIDs = append(have[i].CampaignIDs, 42)
					want[i].CampaignIDs = append(want[i].CampaignIDs, 42)
				}

				if err := s.UpdateChangesets(ctx, have...); err != nil {
					t.Fatal(err)
				}

				for i := range have {
					sort.Slice(have[i].CampaignIDs, func(a, b int) bool {
						return have[i].CampaignIDs[a] < have[i].CampaignIDs[b]
					})

					if diff := cmp.Diff(have[i], want[i]); diff != "" {
						t.Fatal(diff)
					}
				}

				for i := range have {
					// Test we can remove from the set.
					have[i].CampaignIDs = have[i].CampaignIDs[:0]
					want[i].CampaignIDs = want[i].CampaignIDs[:0]
				}

				if err := s.UpdateChangesets(ctx, have...); err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			})
		})

		t.Run("ChangesetEvents", func(t *testing.T) {
			events := make([]*a8n.ChangesetEvent, 0, 3)

			issueComment := &github.IssueComment{
				DatabaseID: 443827703,
				Author: github.Actor{
					AvatarURL: "https://avatars0.githubusercontent.com/u/1976?v=4",
					Login:     "sqs",
					URL:       "https://github.com/sqs",
				},
				Editor:              nil,
				AuthorAssociation:   "MEMBER",
				Body:                "> Just to be sure: you mean the \"searchFilters\" \"Filters\" should be lowercase, not the \"Search Filters\" from the description, right?\r\n\r\nNo, the prose “Search Filters” should have the F lowercased to fit with our style guide preference for sentence case over title case. (Can’t find this comment on the GitHub mobile interface anymore so quoting the email.)",
				URL:                 "https://github.com/sourcegraph/sourcegraph/pull/999#issuecomment-443827703",
				CreatedAt:           now,
				UpdatedAt:           now,
				IncludesCreatedEdit: false,
			}

			t.Run("Upsert", func(t *testing.T) {
				for i := 1; i < cap(events); i++ {
					e := &a8n.ChangesetEvent{
						ChangesetID: int64(i),
						Kind:        a8n.ChangesetEventKindGitHubCommented,
						Key:         issueComment.Key(),
						CreatedAt:   now,
						Metadata:    issueComment,
					}

					events = append(events, e)
				}

				// Verify that no duplicates are introduced and no error is returned.
				for i := 0; i < 2; i++ {
					err := s.UpsertChangesetEvents(ctx, events...)
					if err != nil {
						t.Fatal(err)
					}
				}

				for _, have := range events {
					if have.ID == 0 {
						t.Fatal("id should not be zero")
					}

					want := have.Clone()

					want.ID = have.ID
					want.CreatedAt = now
					want.UpdatedAt = now

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			})

			t.Run("Count", func(t *testing.T) {
				count, err := s.CountChangesetEvents(ctx, CountChangesetEventsOpts{})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(len(events)); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}

				count, err = s.CountChangesetEvents(ctx, CountChangesetEventsOpts{ChangesetID: 1})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(1); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}
			})

			t.Run("Get", func(t *testing.T) {
				t.Run("ByID", func(t *testing.T) {
					want := events[0]
					opts := GetChangesetEventOpts{ID: want.ID}

					have, err := s.GetChangesetEvent(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("ByKey", func(t *testing.T) {
					want := events[0]
					opts := GetChangesetEventOpts{
						ChangesetID: want.ChangesetID,
						Kind:        want.Kind,
						Key:         want.Key,
					}

					have, err := s.GetChangesetEvent(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("NoResults", func(t *testing.T) {
					opts := GetChangesetEventOpts{ID: 0xdeadbeef}

					_, have := s.GetChangesetEvent(ctx, opts)
					want := ErrNoResults

					if have != want {
						t.Fatalf("have err %v, want %v", have, want)
					}
				})
			})

			t.Run("List", func(t *testing.T) {
				t.Run("ByChangesetIDs", func(t *testing.T) {
					for i := 1; i <= len(events); i++ {
						opts := ListChangesetEventsOpts{ChangesetIDs: []int64{int64(i)}}

						ts, next, err := s.ListChangesetEvents(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						if have, want := next, int64(0); have != want {
							t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
						}

						have, want := ts, events[i-1:i]
						if len(have) != len(want) {
							t.Fatalf("listed %d events, want: %d", len(have), len(want))
						}

						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatalf("opts: %+v, diff: %s", opts, diff)
						}
					}

					{
						opts := ListChangesetEventsOpts{ChangesetIDs: []int64{}}

						for i := 1; i <= len(events); i++ {
							opts.ChangesetIDs = append(opts.ChangesetIDs, int64(i))
						}

						ts, next, err := s.ListChangesetEvents(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						if have, want := next, int64(0); have != want {
							t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
						}

						have, want := ts, events
						if len(have) != len(want) {
							t.Fatalf("listed %d events, want: %d", len(have), len(want))
						}
					}
				})

				t.Run("WithLimit", func(t *testing.T) {
					for i := 1; i <= len(events); i++ {
						cs, next, err := s.ListChangesetEvents(ctx, ListChangesetEventsOpts{Limit: i})
						if err != nil {
							t.Fatal(err)
						}

						{
							have, want := next, int64(0)
							if i < len(events) {
								want = events[i].ID
							}

							if have != want {
								t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
							}
						}

						{
							have, want := cs, events[:i]
							if len(have) != len(want) {
								t.Fatalf("listed %d events, want: %d", len(have), len(want))
							}

							if diff := cmp.Diff(have, want); diff != "" {
								t.Fatal(diff)
							}
						}
					}
				})

				t.Run("WithCursor", func(t *testing.T) {
					var cursor int64
					for i := 1; i <= len(events); i++ {
						opts := ListChangesetEventsOpts{Cursor: cursor, Limit: 1}
						have, next, err := s.ListChangesetEvents(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						want := events[i-1 : i]
						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatalf("opts: %+v, diff: %s", opts, diff)
						}

						cursor = next
					}
				})

				t.Run("EmptyResultListingAll", func(t *testing.T) {
					opts := ListChangesetEventsOpts{ChangesetIDs: []int64{99999}, Limit: -1}

					ts, next, err := s.ListChangesetEvents(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if have, want := next, int64(0); have != want {
						t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
					}

					if len(ts) != 0 {
						t.Fatalf("listed %d events, want: %d", len(ts), 0)
					}
				})
			})
		})

		t.Run("CampaignPlans", func(t *testing.T) {
			campaignPlans := make([]*a8n.CampaignPlan, 0, 3)

			t.Run("Create", func(t *testing.T) {
				for i := 0; i < cap(campaignPlans); i++ {
					c := &a8n.CampaignPlan{
						CampaignType: "COMBY",
						Arguments:    `{"scopeQuery": "file:README.md"}`,
						CanceledAt:   now,
					}

					want := c.Clone()
					have := c

					err := s.CreateCampaignPlan(ctx, have)
					if err != nil {
						t.Fatal(err)
					}

					if have.ID == 0 {
						t.Fatal("ID should not be zero")
					}

					want.ID = have.ID
					want.CreatedAt = now
					want.UpdatedAt = now

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}

					campaignPlans = append(campaignPlans, c)
				}
			})

			t.Run("Count", func(t *testing.T) {
				count, err := s.CountCampaignPlans(ctx)
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(len(campaignPlans)); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}
			})

			t.Run("List", func(t *testing.T) {
				opts := ListCampaignPlansOpts{}

				ts, next, err := s.ListCampaignPlans(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if have, want := next, int64(0); have != want {
					t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
				}

				have, want := ts, campaignPlans
				if len(have) != len(want) {
					t.Fatalf("listed %d campaignPlans, want: %d", len(have), len(want))
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}

				for i := 1; i <= len(campaignPlans); i++ {
					cs, next, err := s.ListCampaignPlans(ctx, ListCampaignPlansOpts{Limit: i})
					if err != nil {
						t.Fatal(err)
					}

					{
						have, want := next, int64(0)
						if i < len(campaignPlans) {
							want = campaignPlans[i].ID
						}

						if have != want {
							t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
						}
					}

					{
						have, want := cs, campaignPlans[:i]
						if len(have) != len(want) {
							t.Fatalf("listed %d campaignPlans, want: %d", len(have), len(want))
						}

						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatal(diff)
						}
					}
				}

				{
					var cursor int64
					for i := 1; i <= len(campaignPlans); i++ {
						opts := ListCampaignPlansOpts{Cursor: cursor, Limit: 1}
						have, next, err := s.ListCampaignPlans(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						want := campaignPlans[i-1 : i]
						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatalf("opts: %+v, diff: %s", opts, diff)
						}

						cursor = next
					}
				}
			})

			t.Run("Update", func(t *testing.T) {
				for _, c := range campaignPlans {
					c.CampaignType += "-updated"
					c.Arguments = `{"updated": true}`
					c.CanceledAt = now.Add(5 * time.Second)

					now = now.Add(time.Second)
					want := c
					want.UpdatedAt = now

					have := c.Clone()
					if err := s.UpdateCampaignPlan(ctx, have); err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			})

			t.Run("Get", func(t *testing.T) {
				t.Run("ByID", func(t *testing.T) {
					if len(campaignPlans) == 0 {
						t.Fatalf("campaignPlans is empty")
					}
					want := campaignPlans[0]
					opts := GetCampaignPlanOpts{ID: want.ID}

					have, err := s.GetCampaignPlan(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("NoResults", func(t *testing.T) {
					opts := GetCampaignPlanOpts{ID: 0xdeadbeef}

					_, have := s.GetCampaignPlan(ctx, opts)
					want := ErrNoResults

					if have != want {
						t.Fatalf("have err %v, want %v", have, want)
					}
				})
			})

			t.Run("Delete", func(t *testing.T) {
				for i := range campaignPlans {
					err := s.DeleteCampaignPlan(ctx, campaignPlans[i].ID)
					if err != nil {
						t.Fatal(err)
					}

					count, err := s.CountCampaignPlans(ctx)
					if err != nil {
						t.Fatal(err)
					}

					if have, want := count, int64(len(campaignPlans)-(i+1)); have != want {
						t.Fatalf("have count: %d, want: %d", have, want)
					}
				}
			})
		})

		t.Run("CampaignJobs", func(t *testing.T) {
			campaignJobs := make([]*a8n.CampaignJob, 0, 3)

			t.Run("Create", func(t *testing.T) {
				for i := 0; i < cap(campaignJobs); i++ {
					c := &a8n.CampaignJob{
						CampaignPlanID: int64(i + 1),
						RepoID:         1,
						Rev:            api.CommitID("deadbeef"),
						BaseRef:        "master",
						Diff:           "+ foobar - barfoo",
						Description:    "- Removed 3 instances of foobar\n",
						Error:          "only set on error",
					}

					want := c.Clone()
					have := c

					err := s.CreateCampaignJob(ctx, have)
					if err != nil {
						t.Fatal(err)
					}

					if have.ID == 0 {
						t.Fatal("ID should not be zero")
					}

					want.ID = have.ID
					want.CreatedAt = now
					want.UpdatedAt = now

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}

					campaignJobs = append(campaignJobs, c)
				}
			})

			t.Run("Count", func(t *testing.T) {
				count, err := s.CountCampaignJobs(ctx, CountCampaignJobsOpts{})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(len(campaignJobs)); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}

				count, err = s.CountCampaignJobs(ctx, CountCampaignJobsOpts{CampaignPlanID: 1})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(1); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}
			})

			t.Run("List", func(t *testing.T) {
				t.Run("WithCampaignPlanID", func(t *testing.T) {
					for i := 1; i <= len(campaignJobs); i++ {
						opts := ListCampaignJobsOpts{CampaignPlanID: int64(i)}

						ts, next, err := s.ListCampaignJobs(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						if have, want := next, int64(0); have != want {
							t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
						}

						have, want := ts, campaignJobs[i-1:i]
						if len(have) != len(want) {
							t.Fatalf("listed %d campaignJobs, want: %d", len(have), len(want))
						}

						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatalf("opts: %+v, diff: %s", opts, diff)
						}
					}
				})

				t.Run("WithPositiveLimit", func(t *testing.T) {
					for i := 1; i <= len(campaignJobs); i++ {
						cs, next, err := s.ListCampaignJobs(ctx, ListCampaignJobsOpts{Limit: i})
						if err != nil {
							t.Fatal(err)
						}

						{
							have, want := next, int64(0)
							if i < len(campaignJobs) {
								want = campaignJobs[i].ID
							}

							if have != want {
								t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
							}
						}

						{
							have, want := cs, campaignJobs[:i]
							if len(have) != len(want) {
								t.Fatalf("listed %d campaignJobs, want: %d", len(have), len(want))
							}

							if diff := cmp.Diff(have, want); diff != "" {
								t.Fatal(diff)
							}
						}
					}
				})

				t.Run("WithNegativeLimitToListAll", func(t *testing.T) {
					cs, next, err := s.ListCampaignJobs(ctx, ListCampaignJobsOpts{Limit: -1})
					if err != nil {
						t.Fatal(err)
					}

					if have, want := next, int64(0); have != want {
						t.Fatalf("have next %v, want %v", have, want)
					}

					have, want := cs, campaignJobs
					if len(have) != len(want) {
						t.Fatalf("listed %d campaignJobs, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("EmptyResultListingAll", func(t *testing.T) {
					opts := ListCampaignJobsOpts{CampaignPlanID: 99999, Limit: -1}

					js, next, err := s.ListCampaignJobs(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if have, want := next, int64(0); have != want {
						t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
					}

					if len(js) != 0 {
						t.Fatalf("listed %d jobs, want: %d", len(js), 0)
					}
				})

				t.Run("WithCursor", func(t *testing.T) {
					{
						var cursor int64
						for i := 1; i <= len(campaignJobs); i++ {
							opts := ListCampaignJobsOpts{Cursor: cursor, Limit: 1}
							have, next, err := s.ListCampaignJobs(ctx, opts)
							if err != nil {
								t.Fatal(err)
							}

							want := campaignJobs[i-1 : i]
							if diff := cmp.Diff(have, want); diff != "" {
								t.Fatalf("opts: %+v, diff: %s", opts, diff)
							}

							cursor = next
						}
					}
				})
			})

			t.Run("Listing and Counting OnlyFinished", func(t *testing.T) {
				listOpts := ListCampaignJobsOpts{OnlyFinished: true}
				countOpts := CountCampaignJobsOpts{OnlyFinished: true}

				have, _, err := s.ListCampaignJobs(ctx, listOpts)
				if err != nil {
					t.Fatal(err)
				}

				if len(have) != 0 {
					t.Errorf("jobs returned: %d", len(have))
				}

				count, err := s.CountCampaignJobs(ctx, countOpts)
				if err != nil {
					t.Fatal(err)
				}

				if count != 0 {
					t.Errorf("jobs counted: %d", count)
				}

				for _, j := range campaignJobs {
					j.FinishedAt = now

					err := s.UpdateCampaignJob(ctx, j)
					if err != nil {
						t.Fatal(err)
					}
				}

				have, _, err = s.ListCampaignJobs(ctx, listOpts)
				if err != nil {
					t.Fatal(err)
				}

				have, want := have, campaignJobs
				if len(have) != len(want) {
					t.Fatalf("listed %d campaignJobs, want: %d", len(have), len(want))
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
				}

				count, err = s.CountCampaignJobs(ctx, countOpts)
				if err != nil {
					t.Fatal(err)
				}

				if int(count) != len(campaignJobs) {
					t.Errorf("jobs counted: %d", count)
				}
			})

			t.Run("Listing and Counting OnlyWithDiff", func(t *testing.T) {
				listOpts := ListCampaignJobsOpts{OnlyWithDiff: true}
				countOpts := CountCampaignJobsOpts{OnlyWithDiff: true}

				have, _, err := s.ListCampaignJobs(ctx, listOpts)
				if err != nil {
					t.Fatal(err)
				}

				have, want := have, campaignJobs
				if len(have) != len(want) {
					t.Fatalf("listed %d campaignJobs, want: %d", len(have), len(want))
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
				}

				count, err := s.CountCampaignJobs(ctx, countOpts)
				if err != nil {
					t.Fatal(err)
				}

				if int(count) != len(want) {
					t.Errorf("jobs counted: %d", count)
				}

				for _, j := range campaignJobs {
					j.Diff = ""

					err := s.UpdateCampaignJob(ctx, j)
					if err != nil {
						t.Fatal(err)
					}
				}

				have, _, err = s.ListCampaignJobs(ctx, listOpts)
				if err != nil {
					t.Fatal(err)
				}

				if len(have) != 0 {
					t.Errorf("jobs returned: %d", len(have))
				}

				count, err = s.CountCampaignJobs(ctx, countOpts)
				if err != nil {
					t.Fatal(err)
				}

				if count != 0 {
					t.Errorf("jobs counted: %d", count)
				}
			})

			t.Run("Update", func(t *testing.T) {
				for _, c := range campaignJobs {
					now = now.Add(time.Second)
					c.StartedAt = now
					c.FinishedAt = now
					c.Diff += "-updated"
					c.Description += "-updated"
					c.Error += "-updated"

					want := c
					want.UpdatedAt = now

					have := c.Clone()
					if err := s.UpdateCampaignJob(ctx, have); err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			})

			t.Run("Get", func(t *testing.T) {
				t.Run("ByID", func(t *testing.T) {
					if len(campaignJobs) == 0 {
						t.Fatal("campaignJobs is empty")
					}
					want := campaignJobs[0]
					opts := GetCampaignJobOpts{ID: want.ID}

					have, err := s.GetCampaignJob(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("NoResults", func(t *testing.T) {
					opts := GetCampaignJobOpts{ID: 0xdeadbeef}

					_, have := s.GetCampaignJob(ctx, opts)
					want := ErrNoResults

					if have != want {
						t.Fatalf("have err %v, want %v", have, want)
					}
				})
			})

			t.Run("Delete", func(t *testing.T) {
				for i := range campaignJobs {
					err := s.DeleteCampaignJob(ctx, campaignJobs[i].ID)
					if err != nil {
						t.Fatal(err)
					}

					count, err := s.CountCampaignJobs(ctx, CountCampaignJobsOpts{})
					if err != nil {
						t.Fatal(err)
					}

					if have, want := count, int64(len(campaignJobs)-(i+1)); have != want {
						t.Fatalf("have count: %d, want: %d", have, want)
					}
				}
			})
		})

		t.Run("CampaignPlan BackgroundProcessStatus", func(t *testing.T) {
			tests := []struct {
				planCanceledAt time.Time
				jobs           []*a8n.CampaignJob
				want           *a8n.BackgroundProcessStatus
			}{
				{
					jobs: []*a8n.CampaignJob{}, // no jobs
					want: &a8n.BackgroundProcessStatus{
						ProcessState:  a8n.BackgroundProcessStateCompleted,
						Total:         0,
						Completed:     0,
						Pending:       0,
						ProcessErrors: nil,
					},
				},
				{
					jobs: []*a8n.CampaignJob{
						// not started (pending)
						{},
						// started (pending)
						{StartedAt: now},
					},
					want: &a8n.BackgroundProcessStatus{
						ProcessState:  a8n.BackgroundProcessStateProcessing,
						Total:         2,
						Completed:     0,
						Pending:       2,
						ProcessErrors: nil,
					},
				},
				{
					jobs: []*a8n.CampaignJob{
						// completed, no errors, no diff
						{StartedAt: now, FinishedAt: now},
						// completed, no errors, diff
						{StartedAt: now, FinishedAt: now, Diff: "+foobar\n-barfoo"},
					},
					want: &a8n.BackgroundProcessStatus{
						ProcessState:  a8n.BackgroundProcessStateCompleted,
						Total:         2,
						Completed:     1,
						Pending:       0,
						ProcessErrors: nil,
					},
				},
				{
					jobs: []*a8n.CampaignJob{
						// completed, error
						{StartedAt: now, FinishedAt: now, Error: "error1"},
					},
					want: &a8n.BackgroundProcessStatus{
						ProcessState:  a8n.BackgroundProcessStateErrored,
						Total:         1,
						Completed:     1,
						Pending:       0,
						ProcessErrors: []string{"error1"},
					},
				},
				{
					jobs: []*a8n.CampaignJob{
						// not started (pending)
						{},
						// started (pending)
						{StartedAt: now},
						// completed, no errors, no diff
						{StartedAt: now, FinishedAt: now},
						// completed, no errors, diff
						{StartedAt: now, FinishedAt: now, Diff: "+foobar\n-barfoo"},
						// completed, error
						{StartedAt: now, FinishedAt: now, Error: "error1"},
						// completed, another error
						{StartedAt: now, FinishedAt: now, Error: "error2"},
					},
					want: &a8n.BackgroundProcessStatus{
						ProcessState:  a8n.BackgroundProcessStateProcessing,
						Total:         6,
						Completed:     3,
						Pending:       2,
						ProcessErrors: []string{"error1", "error2"},
					},
				},
				{
					planCanceledAt: now,
					jobs: []*a8n.CampaignJob{
						// not started (pending)
						{},
						// started (pending)
						{StartedAt: now},
					},
					want: &a8n.BackgroundProcessStatus{
						// Instead of "Processing" it's "Canceled"
						ProcessState:  a8n.BackgroundProcessStateCanceled,
						Canceled:      true,
						Total:         2,
						Completed:     0,
						Pending:       2,
						ProcessErrors: nil,
					},
				},
				{
					planCanceledAt: now,
					jobs: []*a8n.CampaignJob{
						// completed, error
						{StartedAt: now, FinishedAt: now, Error: "error1"},
					},
					want: &a8n.BackgroundProcessStatus{
						// Instead of "Errored" it's "Canceled"
						ProcessState:  a8n.BackgroundProcessStateCanceled,
						Canceled:      true,
						Total:         1,
						Completed:     1,
						Pending:       0,
						ProcessErrors: []string{"error1"},
					},
				},
				{
					planCanceledAt: now,
					jobs: []*a8n.CampaignJob{
						// completed, no errors
						{StartedAt: now, FinishedAt: now, Diff: "+foobar\n-foobar"},
					},
					want: &a8n.BackgroundProcessStatus{
						// Instead of "Completed" it's "Canceled"
						ProcessState:  a8n.BackgroundProcessStateCanceled,
						Canceled:      true,
						Total:         1,
						Completed:     1,
						Pending:       0,
						ProcessErrors: nil,
					},
				},
			}
			for _, tc := range tests {
				plan := &a8n.CampaignPlan{
					CampaignType: "comby",
					CanceledAt:   tc.planCanceledAt,
				}
				err := s.CreateCampaignPlan(ctx, plan)
				if err != nil {
					t.Fatal(err)
				}

				for i, j := range tc.jobs {
					j.CampaignPlanID = plan.ID
					j.RepoID = int32(i)
					j.Rev = api.CommitID(fmt.Sprintf("deadbeef-%d", i))
					j.BaseRef = "master"

					err := s.CreateCampaignJob(ctx, j)
					if err != nil {
						t.Fatal(err)
					}
				}

				status, err := s.GetCampaignPlanStatus(ctx, plan.ID)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(status, tc.want); diff != "" {
					t.Fatalf("wrong diff: %s", diff)
				}
			}
		})

		t.Run("CampaignPlan DeleteExpired", func(t *testing.T) {
			tests := []struct {
				hasCampaign bool
				jobs        []*a8n.CampaignJob
				wantDeleted bool
				want        *a8n.BackgroundProcessStatus
			}{
				{
					hasCampaign: false,
					jobs: []*a8n.CampaignJob{
						// completed more than 1 hour ago
						{FinishedAt: now.Add(-61 * time.Minute)},
					},
					wantDeleted: true,
				},
				{
					hasCampaign: false,
					jobs: []*a8n.CampaignJob{
						// completed 30 min ago
						{FinishedAt: now.Add(30 * time.Minute)},
					},
					wantDeleted: false,
				},
				{
					hasCampaign: false,
					jobs: []*a8n.CampaignJob{
						// completed more than 1 hour ago
						{FinishedAt: now.Add(-61 * time.Minute)},
						// completed 30 min ago
						{FinishedAt: now.Add(30 * time.Minute)},
					},
					wantDeleted: false,
				},
				{
					hasCampaign: false,
					jobs: []*a8n.CampaignJob{
						// completed more than 1 hour ago
						{FinishedAt: now.Add(-61 * time.Minute)},
						// not completed
						{},
					},
					wantDeleted: false,
				},
				{
					hasCampaign: false,
					jobs: []*a8n.CampaignJob{
						// completed more than 1 hour ago
						{FinishedAt: now.Add(-61 * time.Minute)},
						// completed more than 2 hours ago
						{FinishedAt: now.Add(-121 * time.Minute)},
					},
					wantDeleted: true,
				},
			}

			for _, tc := range tests {
				plan := &a8n.CampaignPlan{CampaignType: "comby", Arguments: `{"foobar":"barfoo"}`}

				err := s.CreateCampaignPlan(ctx, plan)
				if err != nil {
					t.Fatal(err)
				}
				// Clean up before test
				existingJobs, _, err := s.ListCampaignJobs(ctx, ListCampaignJobsOpts{CampaignPlanID: plan.ID})
				if err != nil {
					t.Fatal(err)
				}
				for _, j := range existingJobs {
					err := s.DeleteCampaignJob(ctx, j.ID)
					if err != nil {
						t.Fatal(err)
					}
				}

				// TODO(a8n): Create a Campaign with CampaignPlanID = plan.ID

				for i, j := range tc.jobs {
					j.StartedAt = now.Add(-2 * time.Hour)
					j.CampaignPlanID = plan.ID
					j.RepoID = int32(i)
					j.Rev = api.CommitID(fmt.Sprintf("deadbeef-%d", i))
					j.BaseRef = "master"

					err := s.CreateCampaignJob(ctx, j)
					if err != nil {
						t.Fatal(err)
					}
				}

				err = s.DeleteExpiredCampaignPlans(ctx)
				if err != nil {
					t.Fatal(err)
				}

				havePlan, err := s.GetCampaignPlan(ctx, GetCampaignPlanOpts{ID: plan.ID})
				if err != nil && err != ErrNoResults {
					t.Fatal(err)
				}

				if tc.wantDeleted && err == nil {
					t.Fatalf("want campaign to be deleted. got: %v", havePlan)
				}

				if !tc.wantDeleted && err == ErrNoResults {
					t.Fatalf("want campaign not to be deletedbut got deleted")
				}
			}
		})

		t.Run("ChangesetJobs", func(t *testing.T) {
			changesetJobs := make([]*a8n.ChangesetJob, 0, 3)

			t.Run("Create", func(t *testing.T) {
				for i := 0; i < cap(changesetJobs); i++ {
					c := &a8n.ChangesetJob{
						CampaignID:    int64(i + 1),
						CampaignJobID: int64(i + 1),
						ChangesetID:   int64(i + 1),
						Error:         "only set on error",
						StartedAt:     now,
						FinishedAt:    now,
					}

					want := c.Clone()
					have := c

					err := s.CreateChangesetJob(ctx, have)
					if err != nil {
						t.Fatal(err)
					}

					if have.ID == 0 {
						t.Fatal("ID should not be zero")
					}

					want.ID = have.ID
					want.CreatedAt = now
					want.UpdatedAt = now

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}

					changesetJobs = append(changesetJobs, c)
				}
			})

			t.Run("Count", func(t *testing.T) {
				count, err := s.CountChangesetJobs(ctx, CountChangesetJobsOpts{})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(len(changesetJobs)); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}

				count, err = s.CountChangesetJobs(ctx, CountChangesetJobsOpts{CampaignID: 1})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := count, int64(1); have != want {
					t.Fatalf("have count: %d, want: %d", have, want)
				}
			})

			t.Run("List", func(t *testing.T) {
				t.Run("WithCampaignID", func(t *testing.T) {
					for i := 1; i <= len(changesetJobs); i++ {
						opts := ListChangesetJobsOpts{CampaignID: int64(i)}

						ts, next, err := s.ListChangesetJobs(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						if have, want := next, int64(0); have != want {
							t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
						}

						have, want := ts, changesetJobs[i-1:i]
						if len(have) != len(want) {
							t.Fatalf("listed %d changesetJobs, want: %d", len(have), len(want))
						}

						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatalf("opts: %+v, diff: %s", opts, diff)
						}
					}
				})

				t.Run("WithPositiveLimit", func(t *testing.T) {
					for i := 1; i <= len(changesetJobs); i++ {
						cs, next, err := s.ListChangesetJobs(ctx, ListChangesetJobsOpts{Limit: i})
						if err != nil {
							t.Fatal(err)
						}

						{
							have, want := next, int64(0)
							if i < len(changesetJobs) {
								want = changesetJobs[i].ID
							}

							if have != want {
								t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
							}
						}

						{
							have, want := cs, changesetJobs[:i]
							if len(have) != len(want) {
								t.Fatalf("listed %d changesetJobs, want: %d", len(have), len(want))
							}

							if diff := cmp.Diff(have, want); diff != "" {
								t.Fatal(diff)
							}
						}
					}
				})

				t.Run("WithNegativeLimitToListAll", func(t *testing.T) {
					cs, next, err := s.ListChangesetJobs(ctx, ListChangesetJobsOpts{Limit: -1})
					if err != nil {
						t.Fatal(err)
					}

					if have, want := next, int64(0); have != want {
						t.Fatalf("have next %v, want %v", have, want)
					}

					have, want := cs, changesetJobs
					if len(have) != len(want) {
						t.Fatalf("listed %d campaignJobs, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("EmptyResultListingAll", func(t *testing.T) {
					opts := ListChangesetJobsOpts{CampaignID: 99999, Limit: -1}

					cs, next, err := s.ListChangesetJobs(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if have, want := next, int64(0); have != want {
						t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
					}

					if len(cs) != 0 {
						t.Fatalf("listed %d jobs, want: %d", len(cs), 0)
					}
				})

				t.Run("WithCursor", func(t *testing.T) {
					var cursor int64
					for i := 1; i <= len(changesetJobs); i++ {
						opts := ListChangesetJobsOpts{Cursor: cursor, Limit: 1}
						have, next, err := s.ListChangesetJobs(ctx, opts)
						if err != nil {
							t.Fatal(err)
						}

						want := changesetJobs[i-1 : i]
						if diff := cmp.Diff(have, want); diff != "" {
							t.Fatalf("opts: %+v, diff: %s", opts, diff)
						}

						cursor = next
					}
				})
			})

			t.Run("Update", func(t *testing.T) {
				for _, c := range changesetJobs {
					now = now.Add(time.Second)
					c.StartedAt = now.Add(1 * time.Second)
					c.FinishedAt = now.Add(1 * time.Second)
					c.Error = "updated-error"

					want := c
					want.UpdatedAt = now

					have := c.Clone()
					if err := s.UpdateChangesetJob(ctx, have); err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			})

			t.Run("Get", func(t *testing.T) {
				t.Run("ByID", func(t *testing.T) {
					if len(changesetJobs) == 0 {
						t.Fatal("changesetJobs is empty")
					}
					want := changesetJobs[0]
					opts := GetChangesetJobOpts{ID: want.ID}

					have, err := s.GetChangesetJob(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("ByCampaignJobID", func(t *testing.T) {
					if len(changesetJobs) == 0 {
						t.Fatal("changesetJobs is empty")
					}
					want := changesetJobs[0]
					opts := GetChangesetJobOpts{CampaignJobID: want.CampaignJobID}

					have, err := s.GetChangesetJob(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("ByChangesetID", func(t *testing.T) {
					if len(changesetJobs) == 0 {
						t.Fatal("changesetJobs is empty")
					}
					want := changesetJobs[0]
					opts := GetChangesetJobOpts{ChangesetID: want.ChangesetID}

					have, err := s.GetChangesetJob(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				})

				t.Run("NoResults", func(t *testing.T) {
					opts := GetChangesetJobOpts{ID: 0xdeadbeef}

					_, have := s.GetChangesetJob(ctx, opts)
					want := ErrNoResults

					if have != want {
						t.Fatalf("have err %v, want %v", have, want)
					}
				})
			})

			t.Run("Delete", func(t *testing.T) {
				for i := range changesetJobs {
					err := s.DeleteChangesetJob(ctx, changesetJobs[i].ID)
					if err != nil {
						t.Fatal(err)
					}

					count, err := s.CountChangesetJobs(ctx, CountChangesetJobsOpts{})
					if err != nil {
						t.Fatal(err)
					}

					if have, want := count, int64(len(changesetJobs)-(i+1)); have != want {
						t.Fatalf("have count: %d, want: %d", have, want)
					}
				}
			})

			t.Run("BackgroundProcessStatus", func(t *testing.T) {
				tests := []struct {
					jobs []*a8n.ChangesetJob
					want *a8n.BackgroundProcessStatus
				}{
					{
						jobs: []*a8n.ChangesetJob{}, // no jobs
						want: &a8n.BackgroundProcessStatus{
							ProcessState:  a8n.BackgroundProcessStateCompleted,
							Total:         0,
							Completed:     0,
							Pending:       0,
							ProcessErrors: nil,
						},
					},
					{
						jobs: []*a8n.ChangesetJob{
							// not started (pending)
							{},
							// started (pending)
							{StartedAt: now},
						},
						want: &a8n.BackgroundProcessStatus{
							ProcessState:  a8n.BackgroundProcessStateProcessing,
							Total:         2,
							Completed:     0,
							Pending:       2,
							ProcessErrors: nil,
						},
					},
					{
						jobs: []*a8n.ChangesetJob{
							// completed, no errors
							{StartedAt: now, FinishedAt: now, ChangesetID: 23},
						},
						want: &a8n.BackgroundProcessStatus{
							ProcessState:  a8n.BackgroundProcessStateCompleted,
							Total:         1,
							Completed:     1,
							Pending:       0,
							ProcessErrors: nil,
						},
					},
					{
						jobs: []*a8n.ChangesetJob{
							// completed, error
							{StartedAt: now, FinishedAt: now, Error: "error1"},
						},
						want: &a8n.BackgroundProcessStatus{
							ProcessState:  a8n.BackgroundProcessStateErrored,
							Total:         1,
							Completed:     1,
							Pending:       0,
							ProcessErrors: []string{"error1"},
						},
					},
					{
						jobs: []*a8n.ChangesetJob{
							// not started (pending)
							{},
							// started (pending)
							{StartedAt: now},
							// completed, no errors
							{StartedAt: now, FinishedAt: now, ChangesetID: 23},
							// completed, error
							{StartedAt: now, FinishedAt: now, Error: "error1"},
							// completed, another error
							{StartedAt: now, FinishedAt: now, Error: "error2"},
						},
						want: &a8n.BackgroundProcessStatus{
							ProcessState:  a8n.BackgroundProcessStateProcessing,
							Total:         5,
							Completed:     3,
							Pending:       2,
							ProcessErrors: []string{"error1", "error2"},
						},
					},
				}

				for campaignID, tc := range tests {
					for i, j := range tc.jobs {
						j.CampaignID = int64(campaignID)
						j.CampaignJobID = int64(i)

						err := s.CreateChangesetJob(ctx, j)
						if err != nil {
							t.Fatal(err)
						}
					}

					status, err := s.GetCampaignStatus(ctx, int64(campaignID))
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(status, tc.want); diff != "" {
						t.Fatalf("wrong diff: %s", diff)
					}
				}
			})

			t.Run("ResetFailedChangesetJobs", func(t *testing.T) {
				campaignID := 9999
				jobs := []*a8n.ChangesetJob{
					{StartedAt: now},
					// completed, no errors
					{StartedAt: now, FinishedAt: now, ChangesetID: 23},
					// not completed, error
					{StartedAt: now, Error: "error1"},
					// completed, error
					{StartedAt: now, FinishedAt: now, Error: "error1"},
					// completed, another error
					{StartedAt: now, FinishedAt: now, Error: "error2"},
				}

				mustReset := make(map[int64]bool)
				for i, j := range jobs {
					j.CampaignID = int64(campaignID)
					j.CampaignJobID = int64(i)

					err := s.CreateChangesetJob(ctx, j)
					if err != nil {
						t.Fatal(err)
					}

					if j.Error != "" && !j.StartedAt.IsZero() && !j.FinishedAt.IsZero() {
						mustReset[j.ID] = true
					}
				}

				err := s.ResetFailedChangesetJobs(ctx, int64(campaignID))
				if err != nil {
					t.Fatal(err)
				}

				have, _, err := s.ListChangesetJobs(ctx, ListChangesetJobsOpts{CampaignID: int64(campaignID)})
				if err != nil {
					t.Fatal(err)
				}

				if len(have) != len(jobs) {
					t.Fatalf("wrong number of jobs returned. have=%d, want=%d", len(have), len(jobs))
				}

				for _, job := range have {
					if _, ok := mustReset[job.ID]; ok {
						if job.Error != "" {
							t.Errorf("job should be reset but has error: %+v", job.Error)
						}
						if !job.FinishedAt.IsZero() {
							t.Errorf("job should be reset but has FinishedAt: %+v", job.FinishedAt)
						}
						if !job.StartedAt.IsZero() {
							t.Errorf("job should be reset but has StartedAt: %+v", job.StartedAt)
						}
					}
				}
			})
		})
	}
}
