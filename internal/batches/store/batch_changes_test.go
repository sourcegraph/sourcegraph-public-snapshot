package store

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	godiff "github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func testStoreBatchChanges(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	bcs := make([]*btypes.BatchChange, 0, 5)

	logger := logtest.Scoped(t)

	// Set up users and organisations for later tests.
	var (
		adminUser  = bt.CreateTestUser(t, s.DatabaseDB(), true)
		orgUser    = bt.CreateTestUser(t, s.DatabaseDB(), false)
		nonOrgUser = bt.CreateTestUser(t, s.DatabaseDB(), false)
		org        = bt.CreateTestOrg(t, s.DatabaseDB(), "org", orgUser.ID)
	)

	t.Run("Create", func(t *testing.T) {
		// We're going to create five batch changes, both to unit test the
		// Create method here and for use in sub-tests further down.
		//
		// 0: draft, owned by org
		// 1: open, owned by orgUser
		// 2: open, owned by org
		// 3: open, owned by adminUser
		// 4: closed, owned by org
		for i, tc := range []struct {
			draft           bool
			closed          bool
			nonNilTimes     bool
			creatorID       int32
			namespaceUserID int32
			namespaceOrgID  int32
		}{
			{namespaceOrgID: org.ID, creatorID: orgUser.ID, draft: true, nonNilTimes: true},
			{namespaceUserID: orgUser.ID, creatorID: orgUser.ID},
			{namespaceOrgID: org.ID, creatorID: orgUser.ID},
			{namespaceUserID: adminUser.ID, creatorID: adminUser.ID},
			{namespaceOrgID: org.ID, creatorID: orgUser.ID, closed: true},
		} {
			c := &btypes.BatchChange{
				Name:        fmt.Sprintf("test-batch-change-%d", i),
				Description: "All the Javascripts are belong to us",

				BatchSpecID:     1742 + int64(i),
				NamespaceUserID: tc.namespaceUserID,
				NamespaceOrgID:  tc.namespaceOrgID,
			}

			// Check for nullability of fields by setting them to a non-nil,
			// zero value.
			if tc.nonNilTimes {
				c.ClosedAt = time.Time{}
				c.LastAppliedAt = time.Time{}
			}

			if !tc.draft {
				c.CreatorID = tc.creatorID
				c.LastAppliedAt = clock.Now()
				c.LastApplierID = tc.creatorID
			}

			if tc.closed {
				c.ClosedAt = clock.Now()
			}

			want := c.Clone()
			have := c

			err := s.CreateBatchChange(ctx, have)
			assert.NoError(t, err)
			assert.NotZero(t, have.ID)

			want.ID = have.ID
			want.CreatedAt = clock.Now()
			want.UpdatedAt = clock.Now()
			assert.Equal(t, want, have)

			bcs = append(bcs, c)
		}

		t.Run("invalid name", func(t *testing.T) {
			c := &btypes.BatchChange{
				Name:        "Invalid name",
				Description: "All the Javascripts are belong to us",

				NamespaceUserID: adminUser.ID,
			}
			tx, err := s.Transact(ctx)
			assert.NoError(t, err)
			defer tx.Done(errors.New("always rollback"))
			err = tx.CreateBatchChange(ctx, c)
			if err != ErrInvalidBatchChangeName {
				t.Fatal("invalid error returned", err)
			}
		})
	})

	t.Run("Upsert", func(t *testing.T) {
		c := &btypes.BatchChange{
			Name:        fmt.Sprintf("test-batch-change-upsert"),
			Description: "All the Javascripts are belong to us",

			NamespaceUserID: adminUser.ID,
		}

		c.BatchSpecID = 1742
		c.CreatorID = adminUser.ID
		c.LastAppliedAt = clock.Now()
		c.LastApplierID = adminUser.ID

		want := c.Clone()
		have := c

		err := s.UpsertBatchChange(ctx, have)
		assert.NoError(t, err)
		assert.NotZero(t, have.ID)

		t.Cleanup(func() {
			// Cleanup.
			assert.NoError(t, s.DeleteBatchChange(ctx, c.ID))
		})

		want.ID = have.ID
		want.CreatedAt = clock.Now()
		want.UpdatedAt = clock.Now()
		assert.Equal(t, want, have)

		c.ClosedAt = clock.Now()
		want = c.Clone()
		err = s.UpsertBatchChange(ctx, have)
		assert.NoError(t, err)
		assert.NotZero(t, have.ID)

		want.ID = have.ID
		want.CreatedAt = clock.Now()
		want.UpdatedAt = clock.Now()
		assert.Equal(t, want, have)

		// Invalid name:
		t.Run("Invalid name", func(t *testing.T) {
			tx, err := s.Transact(ctx)
			assert.NoError(t, err)
			defer tx.Done(errors.New("always rollback"))
			c.Name = "invalid name"
			err = tx.UpsertBatchChange(ctx, have)
			if err != ErrInvalidBatchChangeName {
				t.Fatal("Invalid error returned for invalid name")
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountBatchChanges(ctx, CountBatchChangesOpts{})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, len(bcs); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		t.Run("Global", func(t *testing.T) {
			count, err = s.CountBatchChanges(ctx, CountBatchChangesOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, len(bcs); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})

		t.Run("ChangesetID", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[0].ID}},
			})

			count, err = s.CountBatchChanges(ctx, CountBatchChangesOpts{ChangesetID: changeset.ID})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, 1; have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})

		t.Run("RepoID", func(t *testing.T) {
			repoStore := database.ReposWith(logger, s)
			esStore := database.ExternalServicesWith(logger, s)

			repo1 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo2 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo3 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			if err := repoStore.Create(ctx, repo1, repo2, repo3); err != nil {
				t.Fatal(err)
			}

			// 1 batch change + changeset is associated with the first repo
			bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:         repo1.ID,
				BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[0].ID}},
			})

			// 2 batch changes, each with 1 changeset, are associated with the second repo
			bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:         repo2.ID,
				BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[0].ID}},
			})
			bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:         repo2.ID,
				BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[1].ID}},
			})

			// no batch changes are associated with the third repo

			{
				tcs := []struct {
					repoID api.RepoID
					count  int
				}{
					{
						repoID: repo1.ID,
						count:  1,
					},
					{
						repoID: repo2.ID,
						count:  2,
					},
					{
						repoID: repo3.ID,
						count:  0,
					},
				}

				for i, tc := range tcs {
					t.Run(strconv.Itoa(i), func(t *testing.T) {
						opts := CountBatchChangesOpts{RepoID: tc.repoID}

						count, err := s.CountBatchChanges(actor.WithInternalActor(ctx), opts)
						if err != nil {
							t.Fatal(err)
						}

						if count != tc.count {
							t.Fatalf("listed the wrong number of batch changes: have %d, want %d", count, tc.count)
						}
					})
				}
			}
		})

		t.Run("OnlyAdministeredByUserID set", func(t *testing.T) {
			for name, tc := range map[string]struct {
				userID int32
				want   int
			}{
				// No adminUser test case because the store layer doesn't
				// actually know that site admins have access to everything.

				// orgUser has access to batch changes 0, 1, 2, and 4.
				"orgUser": {userID: orgUser.ID, want: 4},

				// nonOrgUser has access to no batch changes.
				"nonOrgUser": {userID: nonOrgUser.ID, want: 0},
			} {
				t.Run(name, func(t *testing.T) {
					count, err := s.CountBatchChanges(
						ctx,
						CountBatchChangesOpts{OnlyAdministeredByUserID: tc.userID},
					)
					assert.NoError(t, err)
					assert.EqualValues(t, tc.want, count)
				})
			}
		})

		t.Run("NamespaceUserID", func(t *testing.T) {
			wantCounts := map[int32]int{}
			for _, c := range bcs {
				if c.NamespaceUserID == 0 {
					continue
				}
				wantCounts[c.NamespaceUserID] += 1
			}
			if len(wantCounts) == 0 {
				t.Fatalf("No batch changes with NamespaceUserID")
			}

			for userID, want := range wantCounts {
				have, err := s.CountBatchChanges(ctx, CountBatchChangesOpts{NamespaceUserID: userID})
				if err != nil {
					t.Fatal(err)
				}

				if have != want {
					t.Fatalf("batch changes count for NamespaceUserID=%d wrong. want=%d, have=%d", userID, want, have)
				}
			}
		})

		t.Run("NamespaceOrgID", func(t *testing.T) {
			wantCounts := map[int32]int{}
			for _, c := range bcs {
				if c.NamespaceOrgID == 0 {
					continue
				}
				wantCounts[c.NamespaceOrgID] += 1
			}
			if len(wantCounts) == 0 {
				t.Fatalf("No batch changes with NamespaceOrgID")
			}

			for orgID, want := range wantCounts {
				have, err := s.CountBatchChanges(ctx, CountBatchChangesOpts{NamespaceOrgID: orgID})
				if err != nil {
					t.Fatal(err)
				}

				if have != want {
					t.Fatalf("batch changes count for NamespaceOrgID=%d wrong. want=%d, have=%d", orgID, want, have)
				}
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("By ChangesetID", func(t *testing.T) {
			for i := 1; i <= len(bcs); i++ {
				changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[i-1].ID}},
				})
				opts := ListBatchChangesOpts{ChangesetID: changeset.ID}

				ts, next, err := s.ListBatchChanges(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if have, want := next, int64(0); have != want {
					t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
				}

				have, want := ts, bcs[i-1:i]
				if len(have) != len(want) {
					t.Fatalf("listed %d batch changes, want: %d", len(have), len(want))
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}
			}
		})

		t.Run("By RepoID", func(t *testing.T) {
			repoStore := database.ReposWith(logger, s)
			esStore := database.ExternalServicesWith(logger, s)

			repo1 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo2 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo3 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			if err := repoStore.Create(ctx, repo1, repo2, repo3); err != nil {
				t.Fatal(err)
			}

			// 1 batch change + changeset is associated with the first repo
			bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:         repo1.ID,
				BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[0].ID}},
			})

			// 2 batch changes, each with 1 changeset, are associated with the second repo
			bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:         repo2.ID,
				BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[0].ID}},
			})
			bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:         repo2.ID,
				BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: bcs[1].ID}},
			})

			// no batch changes are associated with the third repo

			{
				tcs := []struct {
					repoID        api.RepoID
					listLen       int
					batchChangeID *int64
				}{
					{
						repoID:        repo1.ID,
						listLen:       1,
						batchChangeID: &bcs[0].ID,
					},
					{
						repoID:        repo2.ID,
						listLen:       2,
						batchChangeID: &bcs[1].ID,
					},
					{
						repoID:        repo3.ID,
						listLen:       0,
						batchChangeID: nil,
					},
				}

				for i, tc := range tcs {
					t.Run(strconv.Itoa(i), func(t *testing.T) {
						opts := ListBatchChangesOpts{RepoID: tc.repoID}

						ts, next, err := s.ListBatchChanges(actor.WithInternalActor(ctx), opts)
						if err != nil {
							t.Fatal(err)
						}

						if have, want := next, int64(0); have != want {
							t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
						}

						if len(ts) != tc.listLen {
							t.Fatalf("listed the wrong number of batch changes: have %v, want %v", len(ts), tc.listLen)
						}

						if len(ts) > 0 {
							have, want := ts[0].ID, *tc.batchChangeID
							if have != want {
								t.Fatalf("listed batch change with id %d, wanted %d", have, want)
							}
						}
					})
				}
			}
		})

		// The batch changes store returns the batch changes in reversed order.
		reversedBatchChanges := make([]*btypes.BatchChange, len(bcs))
		for i, c := range bcs {
			reversedBatchChanges[len(bcs)-i-1] = c
		}

		t.Run("With Limit", func(t *testing.T) {
			for i := 1; i <= len(reversedBatchChanges); i++ {
				cs, next, err := s.ListBatchChanges(ctx, ListBatchChangesOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(reversedBatchChanges) {
						want = reversedBatchChanges[i].ID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := cs, reversedBatchChanges[:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d batch changes, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			}
		})

		t.Run("With Cursor", func(t *testing.T) {
			var cursor int64
			for i := 1; i <= len(reversedBatchChanges); i++ {
				opts := ListBatchChangesOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				have, next, err := s.ListBatchChanges(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				want := reversedBatchChanges[i-1 : i]
				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

		filterTests := []struct {
			name  string
			state btypes.BatchChangeState
			want  []*btypes.BatchChange
		}{
			{
				name:  "Any",
				state: "",
				want:  reversedBatchChanges,
			},
			{
				name:  "Closed",
				state: btypes.BatchChangeStateClosed,
				want:  []*btypes.BatchChange{bcs[4]},
			},
			{
				name:  "Open",
				state: btypes.BatchChangeStateOpen,
				want:  []*btypes.BatchChange{bcs[3], bcs[2], bcs[1]},
			},
			{
				name:  "Draft",
				state: btypes.BatchChangeStateDraft,
				want:  []*btypes.BatchChange{bcs[0]},
			},
		}

		for _, tc := range filterTests {
			t.Run("ListBatchChanges Single State "+tc.name, func(t *testing.T) {
				have, _, err := s.ListBatchChanges(ctx, ListBatchChangesOpts{States: []btypes.BatchChangeState{tc.state}})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Fatal(diff)
				}
			})
		}

		multiFilterTests := []struct {
			name   string
			states []btypes.BatchChangeState
			want   []*btypes.BatchChange
		}{
			{
				name:   "Any",
				states: []btypes.BatchChangeState{},
				want:   reversedBatchChanges,
			},
			{
				name:   "All",
				states: []btypes.BatchChangeState{btypes.BatchChangeStateOpen, btypes.BatchChangeStateClosed, btypes.BatchChangeStateDraft},
				want:   reversedBatchChanges,
			},
			{
				name:   "Open + Draft",
				states: []btypes.BatchChangeState{btypes.BatchChangeStateOpen, btypes.BatchChangeStateDraft},
				want:   []*btypes.BatchChange{bcs[3], bcs[2], bcs[1], bcs[0]},
			},
			{
				name:   "Open + Closed",
				states: []btypes.BatchChangeState{btypes.BatchChangeStateOpen, btypes.BatchChangeStateClosed},
				want:   []*btypes.BatchChange{bcs[4], bcs[3], bcs[2], bcs[1]},
			},
			{
				name:   "Draft + Closed",
				states: []btypes.BatchChangeState{btypes.BatchChangeStateDraft, btypes.BatchChangeStateClosed},
				want:   []*btypes.BatchChange{bcs[4], bcs[0]},
			},
			// Multiple of the same state should behave as if it were only one
			{
				name:   "Draft, multiple times",
				states: []btypes.BatchChangeState{btypes.BatchChangeStateDraft, btypes.BatchChangeStateDraft, btypes.BatchChangeStateDraft},
				want:   []*btypes.BatchChange{bcs[0]},
			},
		}

		for _, tc := range multiFilterTests {
			t.Run("ListBatchChanges Multiple States "+tc.name, func(t *testing.T) {

				have, _, err := s.ListBatchChanges(ctx, ListBatchChangesOpts{States: tc.states})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Fatal(diff)
				}
			})
		}

		t.Run("ListBatchChanges OnlyAdministeredByUserID set", func(t *testing.T) {
			for name, tc := range map[string]struct {
				userID int32
				want   []*btypes.BatchChange
			}{
				// No adminUser test case because the store layer doesn't
				// actually know that site admins have access to everything.

				// orgUser has access to batch changes 0, 1, 2, and 4.
				"orgUser": {
					userID: orgUser.ID,
					want:   []*btypes.BatchChange{bcs[4], bcs[2], bcs[1], bcs[0]},
				},

				// nonOrgUser has access to no batch changes.
				"nonOrgUser": {
					userID: nonOrgUser.ID,
					want:   []*btypes.BatchChange{},
				},
			} {
				t.Run(name, func(t *testing.T) {
					have, _, err := s.ListBatchChanges(
						ctx,
						ListBatchChangesOpts{OnlyAdministeredByUserID: tc.userID},
					)
					assert.NoError(t, err)
					assert.Equal(t, tc.want, have)
				})
			}
		})

		t.Run("ListBatchChanges by NamespaceUserID", func(t *testing.T) {
			for _, c := range bcs {
				if c.NamespaceUserID == 0 {
					continue
				}
				opts := ListBatchChangesOpts{NamespaceUserID: c.NamespaceUserID}
				have, _, err := s.ListBatchChanges(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				for _, haveBatchChange := range have {
					if have, want := haveBatchChange.NamespaceUserID, opts.NamespaceUserID; have != want {
						t.Fatalf("batch change has wrong NamespaceUserID. want=%d, have=%d", want, have)
					}
				}
			}
		})

		t.Run("ListBatchChanges by NamespaceOrgID", func(t *testing.T) {
			want := []*btypes.BatchChange{bcs[4], bcs[2], bcs[0]}
			have, _, err := s.ListBatchChanges(ctx, ListBatchChangesOpts{
				NamespaceOrgID: org.ID,
			})
			assert.NoError(t, err)
			assert.Equal(t, want, have)
		})
	})

	t.Run("Update", func(t *testing.T) {
		for _, c := range bcs {
			c.Name += "-updated"
			c.Description += "-updated"
			c.CreatorID++
			c.ClosedAt = c.ClosedAt.Add(5 * time.Second)

			if c.NamespaceUserID != 0 {
				c.NamespaceUserID++
			}

			if c.NamespaceOrgID != 0 {
				c.NamespaceOrgID++
			}

			clock.Add(1 * time.Second)

			want := c
			want.UpdatedAt = clock.Now()

			have := c.Clone()
			if err := s.UpdateBatchChange(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}

		t.Run("invalid name", func(t *testing.T) {
			c := bcs[1].Clone()

			c.Name = "Invalid name"
			tx, err := s.Transact(ctx)
			assert.NoError(t, err)
			defer tx.Done(errors.New("always rollback"))
			err = tx.UpdateBatchChange(ctx, c)
			if err != ErrInvalidBatchChangeName {
				t.Fatal("invalid error returned", err)
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			want := bcs[0]
			opts := GetBatchChangeOpts{ID: want.ID}

			have, err := s.GetBatchChange(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByBatchSpecID", func(t *testing.T) {
			want := bcs[1]
			opts := GetBatchChangeOpts{BatchSpecID: want.BatchSpecID}

			have, err := s.GetBatchChange(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByName", func(t *testing.T) {
			want := bcs[0]

			have, err := s.GetBatchChange(ctx, GetBatchChangeOpts{Name: want.Name})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByNamespaceUserID", func(t *testing.T) {
			for _, c := range bcs {
				if c.NamespaceUserID == 0 {
					continue
				}

				want := c
				opts := GetBatchChangeOpts{NamespaceUserID: c.NamespaceUserID}

				have, err := s.GetBatchChange(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		})

		t.Run("ByNamespaceOrgID", func(t *testing.T) {
			have, err := s.GetBatchChange(ctx, GetBatchChangeOpts{
				// The organisation ID was changed by the Update test above.
				NamespaceOrgID: org.ID + 1,
			})
			assert.NoError(t, err)
			assert.Equal(t, bcs[4], have)
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBatchChangeOpts{ID: 0xdeadbeef}

			_, have := s.GetBatchChange(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("GetBatchChangeDiffStat", func(t *testing.T) {
		userID := bt.CreateTestUser(t, s.DatabaseDB(), false).ID
		otherUserID := bt.CreateTestUser(t, s.DatabaseDB(), false).ID
		userCtx := actor.WithActor(ctx, actor.FromUser(userID))
		repoStore := database.ReposWith(logger, s)
		esStore := database.ExternalServicesWith(logger, s)
		repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		repo.Private = true
		if err := repoStore.Create(ctx, repo); err != nil {
			t.Fatal(err)
		}

		batchChangeID := bcs[0].ID
		var testDiffStatCount int32 = 10
		bt.MockRepoPermissions(t, s.DatabaseDB(), userID, repo.ID)
		bt.MockRepoPermissions(t, s.DatabaseDB(), otherUserID, repo.ID)
		bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
			Repo:            repo.ID,
			BatchChanges:    []btypes.BatchChangeAssoc{{BatchChangeID: batchChangeID}},
			DiffStatAdded:   testDiffStatCount,
			DiffStatDeleted: testDiffStatCount,
		})

		{
			want := &godiff.Stat{
				Added:   testDiffStatCount,
				Deleted: testDiffStatCount,
			}
			opts := GetBatchChangeDiffStatOpts{BatchChangeID: batchChangeID}
			have, err := s.GetBatchChangeDiffStat(userCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}

		// Now give repo access only to otherUserID, and check that
		// userID cannot see it in the diff stat anymore.
		bt.MockRepoPermissions(t, s.DatabaseDB(), userID)
		{
			want := &godiff.Stat{
				Added:   0,
				Changed: 0,
				Deleted: 0,
			}
			opts := GetBatchChangeDiffStatOpts{BatchChangeID: batchChangeID}
			have, err := s.GetBatchChangeDiffStat(userCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("GetRepoDiffStat", func(t *testing.T) {
		userID := bt.CreateTestUser(t, s.DatabaseDB(), false).ID
		otherUserID := bt.CreateTestUser(t, s.DatabaseDB(), false).ID
		userCtx := actor.WithActor(ctx, actor.FromUser(userID))
		repoStore := database.ReposWith(logger, s)
		esStore := database.ExternalServicesWith(logger, s)
		repo1 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		repo2 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		repo3 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		if err := repoStore.Create(ctx, repo1, repo2, repo3); err != nil {
			t.Fatal(err)
		}
		bt.MockRepoPermissions(t, s.DatabaseDB(), userID, repo1.ID, repo2.ID, repo3.ID)

		batchChangeID := bcs[0].ID
		var testDiffStatCount1 int32 = 10
		var testDiffStatCount2 int32 = 20

		// two changesets on the first repo
		bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
			Repo:            repo1.ID,
			BatchChange:     batchChangeID,
			DiffStatAdded:   testDiffStatCount1,
			DiffStatDeleted: testDiffStatCount1,
		})
		bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
			Repo:            repo1.ID,
			BatchChange:     batchChangeID,
			DiffStatAdded:   testDiffStatCount2,
			DiffStatDeleted: testDiffStatCount2,
		})

		// one changeset on the second repo
		bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
			Repo:            repo2.ID,
			BatchChange:     batchChangeID,
			DiffStatAdded:   testDiffStatCount2,
			DiffStatDeleted: testDiffStatCount2,
		})

		// no changesets on the third repo

		{
			tcs := []struct {
				repoID api.RepoID
				want   *godiff.Stat
			}{
				{
					repoID: repo1.ID,
					want: &godiff.Stat{
						Added:   testDiffStatCount1 + testDiffStatCount2,
						Deleted: testDiffStatCount1 + testDiffStatCount2,
					},
				},
				{
					repoID: repo2.ID,
					want: &godiff.Stat{
						Added:   testDiffStatCount2,
						Deleted: testDiffStatCount2,
					},
				},
				{
					repoID: repo3.ID,
					want: &godiff.Stat{
						Added:   0,
						Deleted: 0,
					},
				},
			}

			for i, tc := range tcs {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					have, err := s.GetRepoDiffStat(userCtx, tc.repoID)
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, tc.want); diff != "" {
						t.Errorf("wrong diff returned. have=%+v want=%+v", have, tc.want)
					}
				})
			}

		}

		// Now give repo access only to otherUserID, and check that
		// userID cannot see it in the diff stat anymore.
		bt.MockRepoPermissions(t, s.DatabaseDB(), userID)
		bt.MockRepoPermissions(t, s.DatabaseDB(), otherUserID, repo1.ID)
		{
			want := &godiff.Stat{
				Added:   0,
				Changed: 0,
				Deleted: 0,
			}
			have, err := s.GetRepoDiffStat(userCtx, repo1.ID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		for i := range bcs {
			err := s.DeleteBatchChange(ctx, bcs[i].ID)
			if err != nil {
				t.Fatal(err)
			}

			count, err := s.CountBatchChanges(ctx, CountBatchChangesOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, len(bcs)-(i+1); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		}
	})
}

func testUserDeleteCascades(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	orgID := bt.CreateTestOrg(t, s.DatabaseDB(), "user-delete-cascades").ID
	user := bt.CreateTestUser(t, s.DatabaseDB(), false)

	logger := logtest.Scoped(t)

	t.Run("User delete", func(t *testing.T) {
		// Set up two batch changes and specs: one in the user's namespace (which
		// should be deleted when the user is hard deleted), and one that is
		// merely created by the user (which should remain).
		ownedSpec := &btypes.BatchSpec{
			NamespaceUserID: user.ID,
			UserID:          user.ID,
		}
		if err := s.CreateBatchSpec(ctx, ownedSpec); err != nil {
			t.Fatal(err)
		}

		unownedSpec := &btypes.BatchSpec{
			NamespaceOrgID: orgID,
			UserID:         user.ID,
		}
		if err := s.CreateBatchSpec(ctx, unownedSpec); err != nil {
			t.Fatal(err)
		}

		ownedBatchChange := &btypes.BatchChange{
			Name:            "owned",
			NamespaceUserID: user.ID,
			CreatorID:       user.ID,
			LastApplierID:   user.ID,
			LastAppliedAt:   clock.Now(),
			BatchSpecID:     ownedSpec.ID,
		}
		if err := s.CreateBatchChange(ctx, ownedBatchChange); err != nil {
			t.Fatal(err)
		}

		unownedBatchChange := &btypes.BatchChange{
			Name:           "unowned",
			NamespaceOrgID: orgID,
			CreatorID:      user.ID,
			LastApplierID:  user.ID,
			LastAppliedAt:  clock.Now(),
			BatchSpecID:    ownedSpec.ID,
		}
		if err := s.CreateBatchChange(ctx, unownedBatchChange); err != nil {
			t.Fatal(err)
		}

		// Now we soft-delete the user.
		if err := database.UsersWith(logger, s).Delete(ctx, user.ID); err != nil {
			t.Fatal(err)
		}

		var testBatchChangeIsGone = func(expectedErr error) {
			// We should now have the unowned batch change still be valid, but the
			// owned batch change should have gone away.
			cs, _, err := s.ListBatchChanges(ctx, ListBatchChangesOpts{})
			if err != nil {
				t.Fatal(err)
			}
			if len(cs) != 1 {
				t.Errorf("unexpected number of batch changes: have %d; want %d", len(cs), 1)
			}
			if cs[0].ID != unownedBatchChange.ID {
				t.Errorf("unexpected batch change: %+v", cs[0])
			}

			// The count of batch changes should also respect it.
			count, err := s.CountBatchChanges(ctx, CountBatchChangesOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, len(cs); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}

			// And getting the batch change by its ID also shouldn't work.
			if _, err := s.GetBatchChange(ctx, GetBatchChangeOpts{ID: ownedBatchChange.ID}); err == nil || err != expectedErr {
				t.Fatalf("got invalid error, want=%+v have=%+v", expectedErr, err)
			}

			// Both batch specs should still be in place, at least until we add
			// a foreign key constraint to batch_specs.namespace_user_id.
			specs, _, err := s.ListBatchSpecs(ctx, ListBatchSpecsOpts{
				IncludeLocallyExecutedSpecs: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(specs) != 2 {
				t.Errorf("unexpected number of batch specs: have %d; want %d", len(specs), 2)
			}
		}

		testBatchChangeIsGone(ErrDeletedNamespace)

		// Now we hard-delete the user.
		if err := database.UsersWith(logger, s).HardDelete(ctx, user.ID); err != nil {
			t.Fatal(err)
		}

		testBatchChangeIsGone(ErrNoResults)
	})
}

func testBatchChangesDeletedNamespace(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)

	t.Run("User Deleted", func(t *testing.T) {
		user := bt.CreateTestUser(t, s.DatabaseDB(), false)

		bc := &btypes.BatchChange{
			Name:            "my-batch-change",
			NamespaceUserID: user.ID,
			CreatorID:       user.ID,
			LastApplierID:   user.ID,
			LastAppliedAt:   clock.Now(),
		}
		err := s.CreateBatchChange(ctx, bc)
		require.NoError(t, err)

		t.Cleanup(func() {
			database.UsersWith(logger, s).HardDelete(ctx, user.ID)
			s.DeleteBatchChange(ctx, bc.ID)
		})

		err = database.UsersWith(logger, s).Delete(ctx, user.ID)
		require.NoError(t, err)

		actual, err := s.GetBatchChange(ctx, GetBatchChangeOpts{ID: bc.ID})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDeletedNamespace)
		assert.Nil(t, actual)
	})

	t.Run("Org Deleted", func(t *testing.T) {
		orgID := bt.CreateTestOrg(t, s.DatabaseDB(), "my-org").ID

		bc := &btypes.BatchChange{
			Name:           "my-batch-change",
			NamespaceOrgID: orgID,
			LastAppliedAt:  clock.Now(),
		}
		err := s.CreateBatchChange(ctx, bc)
		require.NoError(t, err)

		t.Cleanup(func() {
			database.OrgsWith(s).HardDelete(ctx, orgID)
			s.DeleteBatchChange(ctx, bc.ID)
		})

		err = database.OrgsWith(s).Delete(ctx, orgID)
		require.NoError(t, err)

		actual, err := s.GetBatchChange(ctx, GetBatchChangeOpts{ID: bc.ID})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDeletedNamespace)
		assert.Nil(t, actual)
	})
}
