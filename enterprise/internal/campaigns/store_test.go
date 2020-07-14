package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

type clock interface {
	now() time.Time
	add(time.Duration) time.Time
}

type testClock struct {
	t time.Time
}

func (c *testClock) now() time.Time                { return c.t }
func (c *testClock) add(d time.Duration) time.Time { c.t = c.t.Add(d); return c.t }

type storeTestFunc func(*testing.T, context.Context, *Store, repos.Store, clock)

// storeTest converts a storeTestFunc into a func(*testing.T) in which all
// dependencies are set up and injected into the storeTestFunc.
func storeTest(db *sql.DB, f storeTestFunc) func(*testing.T) {
	return func(t *testing.T) {
		c := &testClock{t: time.Now().UTC().Truncate(time.Microsecond)}

		// Store tests all run in a transaction that's rolled back at the end
		// of the tests, so that foreign key constraints can be deferred and we
		// don't need to insert a lot of dependencies into the DB (users,
		// repos, ...) to setup the tests.
		tx := dbtest.NewTx(t, db)
		s := NewStoreWithClock(tx, c.now)

		rs := repos.NewDBStore(db, sql.TxOptions{})

		f(t, context.Background(), s, rs, c)
	}
}

// The following tests are executed in integration_test.go.

func testStoreCampaigns(t *testing.T, ctx context.Context, s *Store, _ repos.Store, clock clock) {
	campaigns := make([]*cmpgn.Campaign, 0, 3)

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(campaigns); i++ {
			c := &cmpgn.Campaign{
				Name:         fmt.Sprintf("Upgrade ES-Lint %d", i),
				Description:  "All the Javascripts are belong to us",
				Branch:       "upgrade-es-lint",
				AuthorID:     int32(i) + 50,
				ChangesetIDs: []int64{int64(i) + 1},
				PatchSetID:   42 + int64(i),
				ClosedAt:     clock.now(),
			}
			if i == 0 {
				// don't have a patch set for the first one
				c.PatchSetID = 0
				// Don't close the first one
				c.ClosedAt = time.Time{}
			}

			if i%2 == 0 {
				c.NamespaceOrgID = 23
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

		hasPatchSet := false
		count, err = s.CountCampaigns(ctx, CountCampaignsOpts{HasPatchSet: &hasPatchSet})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, int64(1); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		hasPatchSet = true
		count, err = s.CountCampaigns(ctx, CountCampaignsOpts{HasPatchSet: &hasPatchSet})
		if err != nil {
			t.Fatal(err)
		}
		if have, want := count, int64(2); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		t.Run("OnlyForAuthor set", func(t *testing.T) {
			for _, c := range campaigns {
				count, err = s.CountCampaigns(ctx, CountCampaignsOpts{OnlyForAuthor: c.AuthorID})
				if err != nil {
					t.Fatal(err)
				}
				if have, want := count, int64(1); have != want {
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

		t.Run("ListCampaigns HasPatchSet true", func(t *testing.T) {
			hasPatchSet := true
			have, _, err := s.ListCampaigns(ctx, ListCampaignsOpts{HasPatchSet: &hasPatchSet})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(have, campaigns[1:]); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ListCampaigns HasPatchSet false", func(t *testing.T) {
			hasPatchSet := false
			have, _, err := s.ListCampaigns(ctx, ListCampaignsOpts{HasPatchSet: &hasPatchSet})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(have, campaigns[0:1]); diff != "" {
				t.Fatal(diff)
			}
		})

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

		t.Run("ByPatchSetID", func(t *testing.T) {
			want := campaigns[0]
			opts := GetCampaignOpts{PatchSetID: want.PatchSetID}

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
}

func testStoreChangesets(t *testing.T, ctx context.Context, s *Store, reposStore repos.Store, clock clock) {
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
		CreatedAt:    clock.now(),
		UpdatedAt:    clock.now(),
		HeadRefName:  "campaigns/test",
	}

	repo := testRepo(1, extsvc.TypeGitHub)
	deletedRepo := testRepo(2, extsvc.TypeGitHub).With(repos.Opt.RepoDeletedAt(clock.now()))

	if err := reposStore.UpsertRepos(ctx, deletedRepo, repo); err != nil {
		t.Fatal(err)
	}

	changesets := make(cmpgn.Changesets, 0, 3)

	deletedRepoChangeset := &cmpgn.Changeset{
		RepoID:              deletedRepo.ID,
		ExternalID:          fmt.Sprintf("foobar-%d", cap(changesets)),
		ExternalServiceType: extsvc.TypeGitHub,
	}

	var (
		added   int32 = 77
		deleted int32 = 88
		changed int32 = 99
	)

	t.Run("Create", func(t *testing.T) {
		var i int
		for i = 0; i < cap(changesets); i++ {
			th := &cmpgn.Changeset{
				RepoID:              repo.ID,
				CreatedAt:           clock.now(),
				UpdatedAt:           clock.now(),
				Metadata:            githubPR,
				CampaignIDs:         []int64{int64(i) + 1},
				ExternalID:          fmt.Sprintf("foobar-%d", i),
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalBranch:      "campaigns/test",
				ExternalUpdatedAt:   clock.now(),
				ExternalState:       cmpgn.ChangesetStateOpen,
				ExternalReviewState: cmpgn.ChangesetReviewStateApproved,
				ExternalCheckState:  cmpgn.ChangesetCheckStatePassed,
			}

			// Only set the diff stats on a subset to make sure that
			// we handle nil pointers correctly
			if i != cap(changesets)-1 {
				th.DiffStatAdded = &added
				th.DiffStatChanged = &changed
				th.DiffStatDeleted = &deleted
			}

			changesets = append(changesets, th)
		}

		err := s.CreateChangesets(ctx, changesets...)
		if err != nil {
			t.Fatal(err)
		}

		err = s.CreateChangesets(ctx, deletedRepoChangeset)
		if err != nil {
			t.Fatal(err)
		}

		for _, have := range changesets {
			if have.ID == 0 {
				t.Fatal("id should not be zero")
			}

			if have.IsDeleted() {
				t.Fatal("changeset is deleted")
			}

			want := have.Clone()

			want.ID = have.ID
			want.CreatedAt = clock.now()
			want.UpdatedAt = clock.now()

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("GetChangesetExternalIDs", func(t *testing.T) {
		have, err := s.GetChangesetExternalIDs(ctx, repo.ExternalRepo, []string{githubPR.HeadRefName})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"foobar-0", "foobar-1", "foobar-2"}
		if diff := cmp.Diff(want, have); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("GetChangesetExternalIDs no branch", func(t *testing.T) {
		spec := api.ExternalRepoSpec{
			ID:          "external-id",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		}
		have, err := s.GetChangesetExternalIDs(ctx, spec, []string{"foo"})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{}
		if diff := cmp.Diff(want, have); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("GetChangesetExternalIDs invalid external-id", func(t *testing.T) {
		spec := api.ExternalRepoSpec{
			ID:          "invalid",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		}
		have, err := s.GetChangesetExternalIDs(ctx, spec, []string{"campaigns/test"})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{}
		if diff := cmp.Diff(want, have); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("GetChangesetExternalIDs invalid external service id", func(t *testing.T) {
		spec := api.ExternalRepoSpec{
			ID:          "external-id",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "invalid",
		}
		have, err := s.GetChangesetExternalIDs(ctx, spec, []string{"campaigns/test"})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{}
		if diff := cmp.Diff(want, have); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("CreateAlreadyExistingChangesets", func(t *testing.T) {
		ids := changesets.IDs()
		clones := make(cmpgn.Changesets, len(changesets))

		for i, c := range changesets {
			// Set only the fields on which we have a unique constraint
			clones[i] = &cmpgn.Changeset{
				RepoID:              c.RepoID,
				ExternalID:          c.ExternalID,
				ExternalServiceType: c.ExternalServiceType,
			}
		}

		// Advance clock so store can determine whether Changeset was
		// inserted or not
		clock.add(1 * time.Second)

		err := s.CreateChangesets(ctx, clones...)
		ae, ok := err.(AlreadyExistError)
		if !ok {
			t.Fatalf("error is not AlreadyExistsError: %+v", err)
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
			have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{IDs: changesets.IDs()})
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

		{
			have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{WithoutDeleted: true})
			if err != nil {
				t.Fatal(err)
			}

			if len(have) != len(changesets) {
				t.Fatalf("have 0 changesets. want %d", len(changesets))
			}

			for _, c := range changesets {
				c.SetDeleted()
				c.UpdatedAt = clock.now()
			}

			if err := s.UpdateChangesets(ctx, changesets...); err != nil {
				t.Fatal(err)
			}

			have, _, err = s.ListChangesets(ctx, ListChangesetsOpts{WithoutDeleted: true})
			if err != nil {
				t.Fatal(err)
			}

			if len(have) != 0 {
				t.Fatalf("have %d changesets. want 0", len(changesets))
			}
		}

		{
			have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{OnlyWithoutDiffStats: true})
			if err != nil {
				t.Fatal(err)
			}

			want := 1
			if len(have) != want {
				t.Fatalf("have %d changesets; want %d", len(have), want)
			}

			if have[0].ID != changesets[cap(changesets)-1].ID {
				t.Fatalf("unexpected changeset: have %+v; want %+v", have[0], changesets[cap(changesets)-1])
			}
		}

		// Limit of -1 should return all ChangeSets
		{
			have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{Limit: -1})
			if err != nil {
				t.Fatal(err)
			}

			if len(have) != 3 {
				t.Fatalf("have %d changesets. want 3", len(have))
			}
		}

		stateOpen := cmpgn.ChangesetStateOpen
		stateClosed := cmpgn.ChangesetStateClosed
		stateApproved := cmpgn.ChangesetReviewStateApproved
		stateChangesRequested := cmpgn.ChangesetReviewStateChangesRequested
		statePassed := cmpgn.ChangesetCheckStatePassed
		stateFailed := cmpgn.ChangesetCheckStateFailed

		filterCases := []struct {
			opts      ListChangesetsOpts
			wantCount int
		}{
			{
				opts: ListChangesetsOpts{
					ExternalState: &stateOpen,
				},
				wantCount: 3,
			},
			{
				opts: ListChangesetsOpts{
					ExternalState: &stateClosed,
				},
				wantCount: 0,
			},
			{
				opts: ListChangesetsOpts{
					ExternalReviewState: &stateApproved,
				},
				wantCount: 3,
			},
			{
				opts: ListChangesetsOpts{
					ExternalReviewState: &stateChangesRequested,
				},
				wantCount: 0,
			},
			{
				opts: ListChangesetsOpts{
					ExternalCheckState: &statePassed,
				},
				wantCount: 3,
			},
			{
				opts: ListChangesetsOpts{
					ExternalCheckState: &stateFailed,
				},
				wantCount: 0,
			},
			{
				opts: ListChangesetsOpts{
					ExternalState:      &stateOpen,
					ExternalCheckState: &stateFailed,
				},
				wantCount: 0,
			},
			{
				opts: ListChangesetsOpts{
					ExternalState:       &stateOpen,
					ExternalReviewState: &stateChangesRequested,
				},
				wantCount: 0,
			},
		}

		for _, tc := range filterCases {
			t.Run("", func(t *testing.T) {
				have, _, err := s.ListChangesets(ctx, tc.opts)
				if err != nil {
					t.Fatal(err)
				}
				if len(have) != tc.wantCount {
					t.Fatalf("have %d changesets. want %d", len(have), tc.wantCount)
				}
			})
		}
	})

	t.Run("Null changeset state", func(t *testing.T) {
		cs := &cmpgn.Changeset{
			RepoID:              repo.ID,
			Metadata:            githubPR,
			CampaignIDs:         []int64{1},
			ExternalID:          fmt.Sprintf("foobar-%d", 42),
			ExternalServiceType: extsvc.TypeGitHub,
			ExternalBranch:      "campaigns/test",
			ExternalUpdatedAt:   clock.now(),
			ExternalState:       "",
			ExternalReviewState: "",
			ExternalCheckState:  "",
		}

		err := s.CreateChangesets(ctx, cs)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := s.DeleteChangeset(ctx, cs.ID)
			if err != nil {
				t.Fatal(err)
			}
		}()

		fromDB, err := s.GetChangeset(ctx, GetChangesetOpts{
			ID: cs.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(cs.ExternalState, fromDB.ExternalState); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(cs.ExternalReviewState, fromDB.ExternalReviewState); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(cs.ExternalCheckState, fromDB.ExternalCheckState); diff != "" {
			t.Error(diff)
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

		t.Run("ByRepoID", func(t *testing.T) {
			want := changesets[0]
			opts := GetChangesetOpts{
				RepoID: want.RepoID,
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

		t.Run("RepoDeleted", func(t *testing.T) {
			opts := GetChangesetOpts{ID: deletedRepoChangeset.ID}

			_, have := s.GetChangeset(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		want := make([]*cmpgn.Changeset, 0, len(changesets))
		have := make([]*cmpgn.Changeset, 0, len(changesets))

		clock.add(1 * time.Second)
		for _, c := range changesets {
			c.Metadata = &bitbucketserver.PullRequest{ID: 1234}
			c.ExternalServiceType = extsvc.TypeBitbucketServer

			have = append(have, c.Clone())

			c.UpdatedAt = clock.now()
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

		for _, c := range changesets {
			c.Metadata = &gitlab.MergeRequest{ID: 1234, IID: 123}
			c.ExternalServiceType = extsvc.TypeGitLab

			have = append(have, c.Clone())

			c.UpdatedAt = clock.now()
			want = append(want, c)
		}

		if err := s.UpdateChangesets(ctx, have...); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatal(diff)
		}

	})
}

func testStoreChangesetEvents(t *testing.T, ctx context.Context, s *Store, _ repos.Store, clock clock) {
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
		CreatedAt:           clock.now(),
		UpdatedAt:           clock.now(),
		IncludesCreatedEdit: false,
	}

	events := make([]*cmpgn.ChangesetEvent, 0, 3)

	t.Run("Upsert", func(t *testing.T) {
		for i := 1; i < cap(events); i++ {
			e := &cmpgn.ChangesetEvent{
				ChangesetID: int64(i),
				Kind:        cmpgn.ChangesetEventKindGitHubCommented,
				Key:         issueComment.Key(),
				CreatedAt:   clock.now(),
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
			want.CreatedAt = clock.now()
			want.UpdatedAt = clock.now()

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
}

func testStoreListChangesetSyncData(t *testing.T, ctx context.Context, s *Store, reposStore repos.Store, clock clock) {
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
		CreatedAt:    clock.now(),
		UpdatedAt:    clock.now(),
		HeadRefName:  "campaigns/test",
	}
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
		CreatedAt:           clock.now(),
		UpdatedAt:           clock.now(),
		IncludesCreatedEdit: false,
	}

	var extSvcID int64 = 1
	repo := testRepo(int(extSvcID), extsvc.TypeGitHub)
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	changesets := make([]*cmpgn.Changeset, 0, 3)
	events := make([]*cmpgn.ChangesetEvent, 0)

	for i := 0; i < cap(changesets); i++ {
		changesets = append(changesets, &cmpgn.Changeset{
			RepoID:              repo.ID,
			CreatedAt:           clock.now(),
			UpdatedAt:           clock.now(),
			Metadata:            githubPR,
			CampaignIDs:         []int64{int64(i) + 1},
			ExternalID:          fmt.Sprintf("foobar-%d", i),
			ExternalServiceType: extsvc.TypeGitHub,
			ExternalBranch:      "campaigns/test",
			ExternalUpdatedAt:   clock.now(),
			ExternalState:       cmpgn.ChangesetStateOpen,
			ExternalReviewState: cmpgn.ChangesetReviewStateApproved,
			ExternalCheckState:  cmpgn.ChangesetCheckStatePassed,
		})
	}

	err := s.CreateChangesets(ctx, changesets...)
	if err != nil {
		t.Fatal(err)
	}

	// We need campaigns attached to each changeset
	for _, cs := range changesets {
		c := &cmpgn.Campaign{
			Name:           "ListChangesetSyncData test",
			ChangesetIDs:   []int64{cs.ID},
			NamespaceOrgID: 23,
		}
		err := s.CreateCampaign(ctx, c)
		if err != nil {
			t.Fatal(err)
		}
		cs.CampaignIDs = []int64{c.ID}

	}

	if err := s.UpdateChangesets(ctx, changesets...); err != nil {
		t.Fatal(err)
	}

	// The changesets, except one, get changeset events
	for _, cs := range changesets[:len(changesets)-1] {
		e := &cmpgn.ChangesetEvent{
			ChangesetID: cs.ID,
			Kind:        cmpgn.ChangesetEventKindGitHubCommented,
			Key:         issueComment.Key(),
			CreatedAt:   clock.now(),
			Metadata:    issueComment,
		}

		events = append(events, e)
	}
	err = s.UpsertChangesetEvents(ctx, events...)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		hs, err := s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		want := []cmpgn.ChangesetSyncData{
			{
				ChangesetID:           changesets[0].ID,
				UpdatedAt:             clock.now(),
				LatestEvent:           clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
			{
				ChangesetID:           changesets[1].ID,
				UpdatedAt:             clock.now(),
				LatestEvent:           clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
			{
				// No events
				ChangesetID:           changesets[2].ID,
				UpdatedAt:             clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
		}
		if diff := cmp.Diff(want, hs); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("ignore closed campaign", func(t *testing.T) {
		closedCampaignID := changesets[0].CampaignIDs[0]
		c, err := s.GetCampaign(ctx, GetCampaignOpts{ID: closedCampaignID})
		if err != nil {
			t.Fatal(err)
		}
		c.ClosedAt = clock.now()
		err = s.UpdateCampaign(ctx, c)
		if err != nil {
			t.Fatal(err)
		}

		hs, err := s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		want := []cmpgn.ChangesetSyncData{
			{
				ChangesetID:           changesets[1].ID,
				UpdatedAt:             clock.now(),
				LatestEvent:           clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
			{
				// No events
				ChangesetID:           changesets[2].ID,
				UpdatedAt:             clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
		}
		if diff := cmp.Diff(want, hs); diff != "" {
			t.Fatal(diff)
		}

		// If a changeset has ANY open campaigns we should list it
		// Attach cs1 to both an open and closed campaign
		openCampaignID := changesets[1].CampaignIDs[0]
		changesets[0].CampaignIDs = []int64{closedCampaignID, openCampaignID}
		err = s.UpdateChangesets(ctx, changesets[0])
		if err != nil {
			t.Fatal(err)
		}

		c1, err := s.GetCampaign(ctx, GetCampaignOpts{ID: openCampaignID})
		if err != nil {
			t.Fatal(err)
		}
		c1.ChangesetIDs = []int64{changesets[0].ID, changesets[1].ID}
		err = s.UpdateCampaign(ctx, c1)
		if err != nil {
			t.Fatal(err)
		}

		hs, err = s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		want = []cmpgn.ChangesetSyncData{
			{
				ChangesetID:           changesets[0].ID,
				UpdatedAt:             clock.now(),
				LatestEvent:           clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
			{
				ChangesetID:           changesets[1].ID,
				UpdatedAt:             clock.now(),
				LatestEvent:           clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
			{
				// No events
				ChangesetID:           changesets[2].ID,
				UpdatedAt:             clock.now(),
				ExternalUpdatedAt:     clock.now(),
				RepoExternalServiceID: "https://example.com/",
			},
		}
		if diff := cmp.Diff(want, hs); diff != "" {
			t.Fatal(diff)
		}
	})
}

func testStorePatchSets(t *testing.T, ctx context.Context, s *Store, _ repos.Store, clock clock) {
	patchSets := make([]*cmpgn.PatchSet, 0, 3)

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(patchSets); i++ {
			c := &cmpgn.PatchSet{UserID: 999}

			want := c.Clone()
			have := c

			err := s.CreatePatchSet(ctx, have)
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

			patchSets = append(patchSets, c)
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountPatchSets(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, int64(len(patchSets)); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}
	})

	t.Run("List", func(t *testing.T) {
		opts := ListPatchSetsOpts{}

		ts, next, err := s.ListPatchSets(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := next, int64(0); have != want {
			t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
		}

		have, want := ts, patchSets
		if len(have) != len(want) {
			t.Fatalf("listed %d patchSets, want: %d", len(have), len(want))
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("opts: %+v, diff: %s", opts, diff)
		}

		for i := 1; i <= len(patchSets); i++ {
			cs, next, err := s.ListPatchSets(ctx, ListPatchSetsOpts{Limit: i})
			if err != nil {
				t.Fatal(err)
			}

			{
				have, want := next, int64(0)
				if i < len(patchSets) {
					want = patchSets[i].ID
				}

				if have != want {
					t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
				}
			}

			{
				have, want := cs, patchSets[:i]
				if len(have) != len(want) {
					t.Fatalf("listed %d patchSets, want: %d", len(have), len(want))
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		}

		{
			var cursor int64
			for i := 1; i <= len(patchSets); i++ {
				opts := ListPatchSetsOpts{Cursor: cursor, Limit: 1}
				have, next, err := s.ListPatchSets(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				want := patchSets[i-1 : i]
				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		}
	})

	t.Run("Update", func(t *testing.T) {
		for _, c := range patchSets {
			c.UserID += 1234

			clock.add(1 * time.Second)

			want := c
			want.UpdatedAt = clock.now()

			have := c.Clone()
			if err := s.UpdatePatchSet(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			if len(patchSets) == 0 {
				t.Fatalf("patchSets is empty")
			}
			want := patchSets[0]
			opts := GetPatchSetOpts{ID: want.ID}

			have, err := s.GetPatchSet(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetPatchSetOpts{ID: 0xdeadbeef}

			_, have := s.GetPatchSet(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		for i := range patchSets {
			err := s.DeletePatchSet(ctx, patchSets[i].ID)
			if err != nil {
				t.Fatal(err)
			}

			count, err := s.CountPatchSets(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(len(patchSets)-(i+1)); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		}
	})
}

func testStorePatches(t *testing.T, ctx context.Context, s *Store, reposStore repos.Store, clock clock) {
	patches := make([]*cmpgn.Patch, 0, 3)

	repo := testRepo(1, extsvc.TypeGitHub)
	deletedRepo := testRepo(2, extsvc.TypeGitHub).With(repos.Opt.RepoDeletedAt(clock.now()))
	if err := reposStore.UpsertRepos(ctx, deletedRepo, repo); err != nil {
		t.Fatal(err)
	}

	var (
		added   int32 = 77
		deleted int32 = 88
		changed int32 = 99
	)

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(patches); i++ {
			p := &cmpgn.Patch{
				PatchSetID: int64(i + 1),
				RepoID:     repo.ID,
				Rev:        api.CommitID("deadbeef"),
				BaseRef:    "master",
				Diff:       "+ foobar - barfoo",
			}

			// Only set the diff stats on a subset to make sure that
			// we handle nil pointers correctly
			if i != cap(patches)-1 {
				p.DiffStatAdded = &added
				p.DiffStatChanged = &changed
				p.DiffStatDeleted = &deleted
			}

			want := p.Clone()
			have := p

			err := s.CreatePatch(ctx, have)
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

			patches = append(patches, p)
		}
	})

	// Create patch to deleted repo.
	deletedRepoPatch := &cmpgn.Patch{
		PatchSetID: 1000,
		RepoID:     deletedRepo.ID,
		Rev:        api.CommitID("deadbeef"),
		BaseRef:    "master",
		Diff:       "+ foobar - barfoo",
	}
	err := s.CreatePatch(ctx, deletedRepoPatch)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountPatches(ctx, CountPatchesOpts{})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, int64(len(patches)); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}

		count, err = s.CountPatches(ctx, CountPatchesOpts{PatchSetID: 1})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, int64(1); have != want {
			t.Fatalf("have count: %d, want: %d", have, want)
		}
	})

	t.Run("List", func(t *testing.T) {
		t.Run("WithPatchSetID", func(t *testing.T) {
			for i := 1; i <= len(patches); i++ {
				opts := ListPatchesOpts{PatchSetID: int64(i)}

				ts, next, err := s.ListPatches(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				if have, want := next, int64(0); have != want {
					t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
				}

				have, want := ts, patches[i-1:i]
				if len(have) != len(want) {
					t.Fatalf("listed %d patches, want: %d", len(have), len(want))
				}

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}
			}
		})

		t.Run("WithPositiveLimit", func(t *testing.T) {
			for i := 1; i <= len(patches); i++ {
				cs, next, err := s.ListPatches(ctx, ListPatchesOpts{Limit: i})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(patches) {
						want = patches[i].ID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := cs, patches[:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d patches, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			}
		})

		t.Run("WithNegativeLimitToListAll", func(t *testing.T) {
			cs, next, err := s.ListPatches(ctx, ListPatchesOpts{Limit: -1})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := next, int64(0); have != want {
				t.Fatalf("have next %v, want %v", have, want)
			}

			have, want := cs, patches
			if len(have) != len(want) {
				t.Fatalf("listed %d patches, want: %d", len(have), len(want))
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("EmptyResultListingAll", func(t *testing.T) {
			opts := ListPatchesOpts{PatchSetID: 99999, Limit: -1}

			js, next, err := s.ListPatches(ctx, opts)
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
				for i := 1; i <= len(patches); i++ {
					opts := ListPatchesOpts{Cursor: cursor, Limit: 1}
					have, next, err := s.ListPatches(ctx, opts)
					if err != nil {
						t.Fatal(err)
					}

					want := patches[i-1 : i]
					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatalf("opts: %+v, diff: %s", opts, diff)
					}

					cursor = next
				}
			}
		})

		t.Run("NoDiff", func(t *testing.T) {
			ps, _, err := s.ListPatches(ctx, ListPatchesOpts{NoDiff: true})
			if err != nil {
				fmt.Printf("err=%T", err)
				t.Fatal(err)
			}

			have, want := ps, patches
			if len(have) != len(want) {
				t.Fatalf("listed %d patches, want: %d", len(have), len(want))
			}

			for _, p := range ps {
				if p.Diff != "" {
					t.Fatalf("patch has non-blank diff: %+v", p)
				}
			}
		})
	})

	t.Run("Listing OnlyWithoutChangesetJob", func(t *testing.T) {
		// Define a fake campaign.
		campaignID := int64(1220)

		// Set up two changeset jobs within the campaign: one successful, one
		// failed.
		jobSuccess := &cmpgn.ChangesetJob{
			PatchID:     patches[0].ID,
			CampaignID:  campaignID,
			ChangesetID: 1220,
			StartedAt:   clock.now(),
			FinishedAt:  clock.now(),
		}
		if err := s.CreateChangesetJob(ctx, jobSuccess); err != nil {
			t.Fatal(err)
		}

		jobFailed := &cmpgn.ChangesetJob{
			PatchID:    patches[1].ID,
			CampaignID: campaignID,
			StartedAt:  clock.now(),
			FinishedAt: clock.now(),
			Error:      "Octocat knocked pull request off desk",
		}
		if err := s.CreateChangesetJob(ctx, jobFailed); err != nil {
			t.Fatal(err)
		}

		// List the patches and see what we get back.
		listOpts := ListPatchesOpts{OnlyWithoutChangesetJob: campaignID}
		have, _, err := s.ListPatches(ctx, listOpts)
		if err != nil {
			t.Fatal(err)
		}

		// Since patches[0] and patches[1] exist in the changeset jobs created
		// above, we only expect to see patches[2] in the results.
		want := patches[2:]
		if len(have) != len(want) {
			t.Fatalf("listed %d patches, want: %d", len(have), len(want))
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
		}

		countOpts := CountPatchesOpts{OnlyWithoutChangesetJob: campaignID}
		count, err := s.CountPatches(ctx, countOpts)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, int64(len(patches[2:])); have != want {
			t.Fatalf("Invalid count retrieved: want=%d have=%d", want, have)
		}

		// Update the changeset jobs to change the campaign IDs and try again.
		// This time, we should get all three elements of patches back.
		for _, job := range []*cmpgn.ChangesetJob{jobSuccess, jobFailed} {
			job.CampaignID = 0
			if err = s.UpdateChangesetJob(ctx, job); err != nil {
				t.Fatal(err)
			}
		}

		have, _, err = s.ListPatches(ctx, listOpts)
		if err != nil {
			t.Fatal(err)
		}

		want = patches
		if len(have) != len(want) {
			t.Fatalf("listed %d patches, want: %d", len(have), len(want))
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
		}

		count, err = s.CountPatches(ctx, countOpts)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := count, int64(len(patches)); have != want {
			t.Fatalf("Invalid count retrieved: want=%d have=%d", want, have)
		}
	})

	t.Run("Listing OnlyWithoutDiffStats", func(t *testing.T) {
		listOpts := ListPatchesOpts{OnlyWithoutDiffStats: true}
		have, _, err := s.ListPatches(ctx, listOpts)
		if err != nil {
			t.Fatal(err)
		}

		want := []*cmpgn.Patch{}
		for _, p := range patches {
			_, ok := p.DiffStat()
			if !ok {
				want = append(want, p)
			}
		}

		if len(want) == 0 {
			t.Fatalf("test needs patches without diff stats")
		}
		if len(have) != len(want) {
			t.Fatalf("listed %d patches, want: %d", len(have), len(want))
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
		}
	})

	t.Run("Listing and Counting OnlyWithDiff", func(t *testing.T) {
		listOpts := ListPatchesOpts{OnlyWithDiff: true}
		countOpts := CountPatchesOpts{OnlyWithDiff: true}

		have, _, err := s.ListPatches(ctx, listOpts)
		if err != nil {
			t.Fatal(err)
		}

		have, want := have, patches
		if len(have) != len(want) {
			t.Fatalf("listed %d patches, want: %d", len(have), len(want))
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
		}

		count, err := s.CountPatches(ctx, countOpts)
		if err != nil {
			t.Fatal(err)
		}

		if int(count) != len(want) {
			t.Errorf("jobs counted: %d", count)
		}

		for _, p := range patches {
			p.Diff = ""

			err := s.UpdatePatch(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
		}

		have, _, err = s.ListPatches(ctx, listOpts)
		if err != nil {
			t.Fatal(err)
		}

		if len(have) != 0 {
			t.Errorf("jobs returned: %d", len(have))
		}

		count, err = s.CountPatches(ctx, countOpts)
		if err != nil {
			t.Fatal(err)
		}

		if count != 0 {
			t.Errorf("jobs counted: %d", count)
		}
	})

	t.Run("Listing and Counting OnlyUnpublishedInCampaign", func(t *testing.T) {
		campaignID := int64(999)
		changesetJob := &cmpgn.ChangesetJob{
			PatchID:     patches[0].ID,
			CampaignID:  campaignID,
			ChangesetID: 789,
			StartedAt:   clock.now(),
			FinishedAt:  clock.now(),
		}
		err := s.CreateChangesetJob(ctx, changesetJob)
		if err != nil {
			t.Fatal(err)
		}

		listOpts := ListPatchesOpts{OnlyUnpublishedInCampaign: campaignID}
		countOpts := CountPatchesOpts{OnlyUnpublishedInCampaign: campaignID}

		have, _, err := s.ListPatches(ctx, listOpts)
		if err != nil {
			t.Fatal(err)
		}

		have, want := have, patches[1:] // Except patches[0]
		if len(have) != len(want) {
			t.Fatalf("listed %d patches, want: %d", len(have), len(want))
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
		}

		count, err := s.CountPatches(ctx, countOpts)
		if err != nil {
			t.Fatal(err)
		}

		if int(count) != len(want) {
			t.Errorf("jobs counted: %d", count)
		}

		// Update ChangesetJob so condition does not apply
		changesetJob.ChangesetID = 0
		err = s.UpdateChangesetJob(ctx, changesetJob)
		if err != nil {
			t.Fatal(err)
		}

		have, _, err = s.ListPatches(ctx, listOpts)
		if err != nil {
			t.Fatal(err)
		}

		want = patches // All Patches
		if len(have) != len(want) {
			t.Fatalf("listed %d patches, want: %d", len(have), len(want))
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("opts: %+v, diff: %s", listOpts, diff)
		}

		count, err = s.CountPatches(ctx, countOpts)
		if err != nil {
			t.Fatal(err)
		}

		if int(count) != len(want) {
			t.Errorf("jobs counted: %d", count)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var (
			newAdded   int32 = 333
			newDeleted int32 = 444
			newChanged int32 = 555
		)

		for _, p := range patches {
			clock.add(1 * time.Second)
			p.Diff += "-updated"

			p.DiffStatAdded = &newAdded
			p.DiffStatDeleted = &newDeleted
			p.DiffStatChanged = &newChanged

			want := p
			want.UpdatedAt = clock.now()

			have := p.Clone()
			if err := s.UpdatePatch(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			if len(patches) == 0 {
				t.Fatal("patches is empty")
			}
			want := patches[0]
			opts := GetPatchOpts{ID: want.ID}

			have, err := s.GetPatch(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetPatchOpts{ID: 0xdeadbeef}

			_, have := s.GetPatch(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})

		t.Run("RepoDeleted", func(t *testing.T) {
			opts := GetPatchOpts{ID: deletedRepoPatch.ID}

			_, have := s.GetPatch(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		for i := range patches {
			err := s.DeletePatch(ctx, patches[i].ID)
			if err != nil {
				t.Fatal(err)
			}

			count, err := s.CountPatches(ctx, CountPatchesOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, int64(len(patches)-(i+1)); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		}
	})
}

func testStorePatchSetsDeleteExpired(t *testing.T, ctx context.Context, s *Store, _ repos.Store, clock clock) {
	tests := []struct {
		createdAt                      time.Time
		hasCampaign                    bool
		patchesAttachedToOtherCampaign bool
		patches                        []*cmpgn.Patch
		wantDeleted                    bool
		want                           *cmpgn.BackgroundProcessStatus
	}{
		{
			hasCampaign: false,
			createdAt:   clock.now(),
			wantDeleted: false,
		},
		{
			hasCampaign: false,
			createdAt:   clock.now().Add(-8 * 24 * time.Hour),
			wantDeleted: true,
		},
		{
			hasCampaign: true,
			createdAt:   clock.now(),
			wantDeleted: false,
		},
		{
			hasCampaign: true,
			createdAt:   clock.now().Add(-8 * 24 * time.Hour),
			wantDeleted: false,
		},
		{
			hasCampaign: false,
			createdAt:   clock.now().Add(-8 * 24 * time.Hour),

			patchesAttachedToOtherCampaign: true,
			patches: []*cmpgn.Patch{
				{Diff: "foobar", Rev: "f00b4r", BaseRef: "refs/heads/master"},
				{Diff: "barfoo", Rev: "b4rf00", BaseRef: "refs/heads/master"},
			},

			wantDeleted: false,
		},
	}

	for _, tc := range tests {
		patchSet := &cmpgn.PatchSet{CreatedAt: tc.createdAt}

		err := s.CreatePatchSet(ctx, patchSet)
		if err != nil {
			t.Fatal(err)
		}

		for _, p := range tc.patches {
			p.PatchSetID = patchSet.ID
			err := s.CreatePatch(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
		}

		if tc.hasCampaign {
			err = s.CreateCampaign(ctx, &cmpgn.Campaign{
				PatchSetID:      patchSet.ID,
				Name:            "Test",
				AuthorID:        4567,
				NamespaceUserID: 4567,
			})
			if err != nil {
				t.Fatal(err)
			}
		}

		if tc.patchesAttachedToOtherCampaign {
			otherPatchSet := &cmpgn.PatchSet{}
			err = s.CreatePatchSet(ctx, otherPatchSet)
			if err != nil {
				t.Fatal(err)
			}
			otherCampaign := &cmpgn.Campaign{
				PatchSetID:      otherPatchSet.ID,
				AuthorID:        4567,
				Name:            "Other campaign",
				NamespaceUserID: 4567,
			}

			err = s.CreateCampaign(ctx, otherCampaign)
			if err != nil {
				t.Fatal(err)
			}

			for i, p := range tc.patches {
				changeset := &cmpgn.Changeset{
					RepoID:              api.RepoID(99 + i),
					CreatedAt:           clock.now(),
					UpdatedAt:           clock.now(),
					Metadata:            &github.PullRequest{},
					CampaignIDs:         []int64{otherCampaign.ID},
					ExternalID:          fmt.Sprintf("foobar-%d", i),
					ExternalServiceType: extsvc.TypeGitHub,
					ExternalBranch:      "campaigns/test",
					ExternalUpdatedAt:   clock.now(),
					ExternalState:       cmpgn.ChangesetStateOpen,
					ExternalReviewState: cmpgn.ChangesetReviewStateApproved,
					ExternalCheckState:  cmpgn.ChangesetCheckStatePassed,
				}
				err = s.CreateChangesets(ctx, changeset)
				if err != nil {
					t.Fatal(err)
				}
				job := &cmpgn.ChangesetJob{
					CampaignID:  otherCampaign.ID,
					PatchID:     p.ID,
					ChangesetID: changeset.ID,
				}
				err = s.CreateChangesetJob(ctx, job)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					err = s.DeleteChangesetJob(ctx, job.ID)
					if err != nil {
						t.Fatal(err)
					}
				}()
			}
		}

		err = s.DeleteExpiredPatchSets(ctx)
		if err != nil {
			t.Fatal(err)
		}

		havePatchSet, err := s.GetPatchSet(ctx, GetPatchSetOpts{ID: patchSet.ID})
		if err != nil && err != ErrNoResults {
			t.Fatal(err)
		}

		if tc.wantDeleted && err == nil {
			t.Fatalf("tc=%+v\n\t want patch set to be deleted. got: %v", tc, havePatchSet)
		}

		if !tc.wantDeleted && err == ErrNoResults {
			t.Fatalf("want patch set not to be deleted, but got deleted")
		}
	}
}

func testStoreChangesetJobs(t *testing.T, ctx context.Context, s *Store, _ repos.Store, clock clock) {
	changesetJobs := make([]*cmpgn.ChangesetJob, 0, 3)

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(changesetJobs); i++ {
			c := &cmpgn.ChangesetJob{
				CampaignID:  int64(i + 1),
				PatchID:     int64(i + 1),
				ChangesetID: int64(i + 1),
				Branch:      "test-branch",
				Error:       "only set on error",
				StartedAt:   clock.now(),
				FinishedAt:  clock.now(),
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
			want.CreatedAt = clock.now()
			want.UpdatedAt = clock.now()

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
				t.Fatalf("listed %d patches, want: %d", len(have), len(want))
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

		t.Run("WithPatchSetID", func(t *testing.T) {
			for i := 1; i <= len(changesetJobs); i++ {
				c := &cmpgn.Campaign{
					Name:            fmt.Sprintf("Upgrade ES-Lint %d", i),
					Description:     "All the Javascripts are belong to us",
					AuthorID:        4567,
					NamespaceUserID: 4567,
					PatchSetID:      1234 + int64(i),
				}

				err := s.CreateCampaign(ctx, c)
				if err != nil {
					t.Fatal(err)
				}
				job := changesetJobs[i-1]

				job.CampaignID = c.ID
				err = s.UpdateChangesetJob(ctx, job)
				if err != nil {
					t.Fatal(err)
				}

				opts := ListChangesetJobsOpts{PatchSetID: c.PatchSetID}
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
	})

	t.Run("Update", func(t *testing.T) {
		for _, c := range changesetJobs {
			clock.add(1 * time.Second)
			c.StartedAt = clock.now().Add(1 * time.Second)
			c.FinishedAt = clock.now().Add(1 * time.Second)
			c.Branch = "upgrade-es-lint"
			c.Error = "updated-error"

			want := c
			want.UpdatedAt = clock.now()

			have := c.Clone()
			if err := s.UpdateChangesetJob(ctx, have); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("UpdateWithLargeError", func(t *testing.T) {
		// We were seeing errors creating / updating changeset jobs with errors
		// larger than 2704 bytes
		// https://github.com/sourcegraph/sourcegraph/issues/10798
		c := changesetJobs[0]
		clock.add(1 * time.Second)
		c.StartedAt = clock.now().Add(1 * time.Second)
		c.FinishedAt = clock.now().Add(1 * time.Second)
		c.Branch = "upgrade-es-lint"
		// Strangely this value needs to be a lot higher than the threshold before
		// an error is actually raised. It must be due to the way PostgreSQL constructs
		// the index.
		c.Error = strings.Repeat("X", 1000000)

		want := c
		want.UpdatedAt = clock.now()

		have := c.Clone()
		if err := s.UpdateChangesetJob(ctx, have); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatal(diff)
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

		t.Run("ByPatchID", func(t *testing.T) {
			if len(changesetJobs) == 0 {
				t.Fatal("changesetJobs is empty")
			}
			want := changesetJobs[0]
			opts := GetChangesetJobOpts{PatchID: want.PatchID}

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

		t.Run("ByCampaignID", func(t *testing.T) {
			if len(changesetJobs) == 0 {
				t.Fatal("changesetJobs is empty")
			}
			// Use the last changesetJob, which we don't get by
			// accident when selecting all with LIMIT 1
			want := changesetJobs[2]
			opts := GetChangesetJobOpts{CampaignID: want.CampaignID}

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
			jobs []*cmpgn.ChangesetJob
			want *cmpgn.BackgroundProcessStatus
			opts GetCampaignStatusOpts
		}{
			{
				jobs: []*cmpgn.ChangesetJob{}, // no jobs
				want: &cmpgn.BackgroundProcessStatus{
					ProcessState:  cmpgn.BackgroundProcessStateCompleted,
					Total:         0,
					Completed:     0,
					Pending:       0,
					ProcessErrors: nil,
				},
			},
			{
				jobs: []*cmpgn.ChangesetJob{
					// not started (pending)
					{},
					// started (pending)
					{StartedAt: clock.now()},
				},
				want: &cmpgn.BackgroundProcessStatus{
					ProcessState:  cmpgn.BackgroundProcessStateProcessing,
					Total:         2,
					Completed:     0,
					Pending:       2,
					ProcessErrors: nil,
				},
			},
			{
				jobs: []*cmpgn.ChangesetJob{
					// completed, no errors
					{StartedAt: clock.now(), FinishedAt: clock.now(), ChangesetID: 23},
				},
				want: &cmpgn.BackgroundProcessStatus{
					ProcessState:  cmpgn.BackgroundProcessStateCompleted,
					Total:         1,
					Completed:     1,
					Pending:       0,
					ProcessErrors: nil,
				},
			},
			{
				jobs: []*cmpgn.ChangesetJob{
					// completed, error
					{StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error1"},
				},
				want: &cmpgn.BackgroundProcessStatus{
					ProcessState:  cmpgn.BackgroundProcessStateErrored,
					Total:         1,
					Completed:     1,
					Failed:        1,
					Pending:       0,
					ProcessErrors: []string{"error1"},
				},
			},
			{
				jobs: []*cmpgn.ChangesetJob{
					// not started (pending)
					{},
					// started (pending)
					{StartedAt: clock.now()},
					// completed, no errors
					{StartedAt: clock.now(), FinishedAt: clock.now(), ChangesetID: 23},
					// completed, error
					{StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error1"},
					// completed, another error
					{StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error2"},
				},
				want: &cmpgn.BackgroundProcessStatus{
					ProcessState:  cmpgn.BackgroundProcessStateProcessing,
					Total:         5,
					Completed:     3,
					Failed:        2,
					Pending:       2,
					ProcessErrors: []string{"error1", "error2"},
				},
			},

			{
				jobs: []*cmpgn.ChangesetJob{
					// completed, error
					{StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error1"},
				},
				// but we want to exclude errors
				opts: GetCampaignStatusOpts{ExcludeErrors: true},
				want: &cmpgn.BackgroundProcessStatus{
					ProcessState:  cmpgn.BackgroundProcessStateErrored,
					Total:         1,
					Completed:     1,
					Failed:        1,
					Pending:       0,
					ProcessErrors: nil,
				},
			},
		}

		for campaignID, tc := range tests {
			for i, j := range tc.jobs {
				p := &cmpgn.Patch{
					RepoID:     api.RepoID(i),
					PatchSetID: int64(campaignID),
					BaseRef:    "deadbeef",
					Diff:       "foobar",
				}

				if err := s.CreatePatch(ctx, p); err != nil {
					t.Fatal(err)
				}

				j.CampaignID = int64(campaignID)
				j.PatchID = p.ID

				err := s.CreateChangesetJob(ctx, j)
				if err != nil {
					t.Fatal(err)
				}
			}

			opts := tc.opts
			opts.ID = int64(campaignID)

			status, err := s.GetCampaignStatus(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(status, tc.want); diff != "" {
				t.Fatalf("wrong diff: %s", diff)
			}
		}
	})
	t.Run("BackgroundProcessStatus_ErrorsOnlyInRepos", func(t *testing.T) {
		var campaignID int64 = 123456

		patches := []*cmpgn.Patch{
			{RepoID: 444, PatchSetID: 888, BaseRef: "deadbeef", Diff: "foobar"},
			{RepoID: 555, PatchSetID: 888, BaseRef: "deadbeef", Diff: "foobar"},
			{RepoID: 666, PatchSetID: 888, BaseRef: "deadbeef", Diff: "foobar"},
		}
		for _, p := range patches {
			if err := s.CreatePatch(ctx, p); err != nil {
				t.Fatal(err)
			}
		}

		jobs := []*cmpgn.ChangesetJob{
			// completed, no errors
			{PatchID: patches[0].ID, StartedAt: clock.now(), FinishedAt: clock.now(), ChangesetID: 23},
			// completed, error
			{PatchID: patches[1].ID, StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error1"},
			// completed, another error
			{PatchID: patches[2].ID, StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error2"},
		}

		for _, j := range jobs {
			j.CampaignID = campaignID

			err := s.CreateChangesetJob(ctx, j)
			if err != nil {
				t.Fatal(err)
			}
		}

		opts := GetCampaignStatusOpts{
			ID:                   campaignID,
			ExcludeErrorsInRepos: []api.RepoID{patches[2].RepoID},
		}

		want := &cmpgn.BackgroundProcessStatus{
			ProcessState:  cmpgn.BackgroundProcessStateErrored,
			Total:         3,
			Completed:     3,
			Failed:        2,
			Pending:       0,
			ProcessErrors: []string{"error1"},
			// error2 should be excluded
		}

		status, err := s.GetCampaignStatus(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(status, want); diff != "" {
			t.Fatalf("wrong diff: %s", diff)
		}

		// Now we filter out all errors, but still want the ProcessState to be
		// correct
		opts = GetCampaignStatusOpts{
			ID: campaignID,
			ExcludeErrorsInRepos: []api.RepoID{
				patches[0].RepoID,
				patches[1].RepoID,
				patches[2].RepoID,
			},
		}

		want = &cmpgn.BackgroundProcessStatus{
			// This should stay "Errored", even though no errors are returned.
			ProcessState:  cmpgn.BackgroundProcessStateErrored,
			Total:         3,
			Completed:     3,
			Failed:        2,
			Pending:       0,
			ProcessErrors: nil,
			// error1 and error2 should be excluded
		}

		status, err = s.GetCampaignStatus(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(status, want); diff != "" {
			t.Fatalf("wrong diff: %s", diff)
		}
	})

	t.Run("ResetChangesetJobs", func(t *testing.T) {
		tests := []struct {
			jobs              []*cmpgn.ChangesetJob
			opts              ResetChangesetJobsOpts
			wantResetPatchIDs []int64
		}{
			{
				jobs: []*cmpgn.ChangesetJob{
					// completed, no errors
					{PatchID: 1, StartedAt: clock.now(), FinishedAt: clock.now(), ChangesetID: 23},
					// completed, error
					{PatchID: 2, StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error1"},
					// completed, another error
					{PatchID: 3, StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error2"},
				},
				opts:              ResetChangesetJobsOpts{OnlyFailed: true},
				wantResetPatchIDs: []int64{2, 3},
			},
			{
				jobs: []*cmpgn.ChangesetJob{
					// completed, no errors
					{PatchID: 1, StartedAt: clock.now(), FinishedAt: clock.now(), ChangesetID: 23},
					// completed, error
					{PatchID: 2, StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error1"},
				},
				opts:              ResetChangesetJobsOpts{},
				wantResetPatchIDs: []int64{1, 2},
			},
			{
				jobs: []*cmpgn.ChangesetJob{
					// completed, no errors
					{PatchID: 1, StartedAt: clock.now(), FinishedAt: clock.now(), ChangesetID: 23},
					// completed, error
					{PatchID: 2, StartedAt: clock.now(), FinishedAt: clock.now(), Error: "error1"},
				},
				opts:              ResetChangesetJobsOpts{PatchIDs: []int64{2}},
				wantResetPatchIDs: []int64{2},
			},
		}

		for i, tc := range tests {
			var campaignID = int64(9999 + i)

			for _, j := range tc.jobs {
				j.CampaignID = campaignID

				if err := s.CreateChangesetJob(ctx, j); err != nil {
					t.Fatal(err)
				}

			}

			tc.opts.CampaignID = campaignID
			if err := s.ResetChangesetJobs(ctx, tc.opts); err != nil {
				t.Fatal(err)
			}

			have, _, err := s.ListChangesetJobs(ctx, ListChangesetJobsOpts{CampaignID: campaignID})
			if err != nil {
				t.Fatal(err)
			}

			if len(have) != len(tc.jobs) {
				t.Fatalf("wrong number of jobs returned. have=%d, want=%d", len(have), len(tc.jobs))
			}

			mustReset := map[int64]bool{}
			for _, patchID := range tc.wantResetPatchIDs {
				mustReset[patchID] = true
			}

			for _, job := range have {
				if _, ok := mustReset[job.PatchID]; ok {
					if job.UnsuccessfullyCompleted() {
						t.Errorf("job should be reset but is not: %+v", job)
					}
				} else {
					if job.StartedAt.IsZero() {
						t.Errorf("job should not be reset but StartedAt is zero: %+v", job.StartedAt)
					}
					if job.FinishedAt.IsZero() {
						t.Errorf("job should not be reset but FinishedAt is zero: %+v", job.FinishedAt)
					}
				}
			}
		}
	})

	t.Run("GetRepoIDsForFailedChangesetJobs", func(t *testing.T) {
		var campaignID int64 = 654321

		patches := []*cmpgn.Patch{
			{RepoID: 111, PatchSetID: 888, BaseRef: "deadbeef", Diff: "foobar"},
			{RepoID: 222, PatchSetID: 888, BaseRef: "deadbeef", Diff: "foobar"},
			{RepoID: 333, PatchSetID: 888, BaseRef: "deadbeef", Diff: "foobar"},
		}
		for _, p := range patches {
			if err := s.CreatePatch(ctx, p); err != nil {
				t.Fatal(err)
			}
		}

		jobs := []*cmpgn.ChangesetJob{
			// completed, no errors
			{
				PatchID:     patches[0].ID,
				StartedAt:   clock.now(),
				FinishedAt:  clock.now(),
				ChangesetID: 23,
			},
			// completed, error
			{
				PatchID:    patches[1].ID,
				StartedAt:  clock.now(),
				FinishedAt: clock.now(),
				Error:      "error1",
			},
			// completed, another error
			{
				PatchID:    patches[2].ID,
				StartedAt:  clock.now(),
				FinishedAt: clock.now(),
				Error:      "error2",
			},
		}

		for _, j := range jobs {
			j.CampaignID = campaignID

			err := s.CreateChangesetJob(ctx, j)
			if err != nil {
				t.Fatal(err)
			}
		}

		want := []api.RepoID{
			patches[1].RepoID,
			patches[2].RepoID,
		}

		have, err := s.GetRepoIDsForFailedChangesetJobs(ctx, campaignID)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("wrong diff: %s", diff)
		}
	})
}

func testProcessChangesetJob(db *sql.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		clock := func() time.Time { return now.UTC().Truncate(time.Microsecond) }
		ctx := context.Background()

		// Create a test repo
		reposStore := repos.NewDBStore(db, sql.TxOptions{})
		repo := &repos.Repo{
			Name: "github.com/sourcegraph/changeset-job-test",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "external-id",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com/",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:github:4": {
					ID:       "extsvc:github:4",
					CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
				},
			},
		}
		if err := reposStore.UpsertRepos(context.Background(), repo); err != nil {
			t.Fatal(err)
		}

		s := NewStoreWithClock(db, clock)
		patchSet := &cmpgn.PatchSet{UserID: userID}
		err := s.CreatePatchSet(context.Background(), patchSet)
		if err != nil {
			t.Fatal(err)
		}

		patch := &cmpgn.Patch{
			PatchSetID: patchSet.ID,
			RepoID:     repo.ID,
			BaseRef:    "abc",
		}
		err = s.CreatePatch(context.Background(), patch)
		if err != nil {
			t.Fatal(err)
		}

		patchSet2 := &cmpgn.PatchSet{UserID: userID}
		err = s.CreatePatchSet(context.Background(), patchSet2)
		if err != nil {
			t.Fatal(err)
		}
		patch2 := &cmpgn.Patch{
			PatchSetID: patchSet2.ID,
			RepoID:     repo.ID,
			BaseRef:    "abc",
		}
		err = s.CreatePatch(context.Background(), patch2)
		if err != nil {
			t.Fatal(err)
		}

		campaign := &cmpgn.Campaign{
			PatchSetID:      patchSet.ID,
			Name:            "testcampaign",
			Description:     "testcampaign",
			AuthorID:        userID,
			NamespaceUserID: userID,
		}
		err = s.CreateCampaign(context.Background(), campaign)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("GetPendingChangesetJobsWhenNoneAvailable", func(t *testing.T) {
			tx := dbtest.NewTx(t, db)
			s := NewStoreWithClock(tx, clock)

			process := func(ctx context.Context, s *Store, job cmpgn.ChangesetJob) error {
				return errors.New("rollback")
			}

			ran, err := s.ProcessPendingChangesetJobs(ctx, process)
			if err != nil {
				t.Fatal(err)
			}
			if ran {
				// We shouldn't have any pending jobs yet
				t.Fatalf("process function should not have run")
			}
		})

		t.Run("GetPendingChangesetJobWhenAvailable", func(t *testing.T) {
			tx := dbtest.NewTx(t, db)
			s := NewStoreWithClock(tx, clock)

			process := func(ctx context.Context, s *Store, job cmpgn.ChangesetJob) error {
				return errors.New("rollback")
			}

			job := &cmpgn.ChangesetJob{
				CampaignID: campaign.ID,
				PatchID:    patch.ID,
			}
			err := s.CreateChangesetJob(ctx, job)
			if err != nil {
				t.Fatal(err)
			}

			ran, err := s.ProcessPendingChangesetJobs(ctx, process)
			if err != nil && err.Error() != "rollback" {
				t.Fatal(err)
			}
			if !ran {
				// We shouldn't have any pending jobs yet
				t.Fatalf("process function should have run")
			}
		})

		t.Run("GetPendingChangesetJobOrder", func(t *testing.T) {
			// Test that we get the oldest job first
			tx := dbtest.NewTx(t, db)
			// NOTE: We don't use a clock here as we need ordering by updated_at to work
			s := NewStore(tx)

			var idRun int64
			process := func(ctx context.Context, s *Store, job cmpgn.ChangesetJob) error {
				idRun = job.ID
				return errors.New("rollback")
			}

			job1 := &cmpgn.ChangesetJob{
				CampaignID: campaign.ID,
				PatchID:    patch.ID,
			}
			err := s.CreateChangesetJob(ctx, job1)
			if err != nil {
				t.Fatal(err)
			}

			job2 := &cmpgn.ChangesetJob{
				CampaignID: campaign.ID,
				PatchID:    patch2.ID,
			}
			err = s.CreateChangesetJob(ctx, job2)
			if err != nil {
				t.Fatal(err)
			}

			// Move the old job to the back of the queue
			err = s.UpdateChangesetJob(ctx, job1)
			if err != nil {
				t.Fatal(err)
			}

			ran, err := s.ProcessPendingChangesetJobs(ctx, process)
			if err != nil && err.Error() != "rollback" {
				t.Fatal(err)
			}
			if !ran {
				// We shouldn't have any pending jobs yet
				t.Fatalf("process function should have run")
			}
			if idRun != job2.ID {
				t.Fatalf("Job with oldest update_at should have run")
			}
		})

		t.Run("GetPendingChangesetJobsWhenAvailableLocking", func(t *testing.T) {
			s := NewStoreWithClock(db, clock)

			process := func(ctx context.Context, s *Store, job cmpgn.ChangesetJob) error {
				time.Sleep(100 * time.Millisecond)
				return errors.New("rollback")
			}

			job := &cmpgn.ChangesetJob{
				CampaignID: campaign.ID,
				PatchID:    patch.ID,
			}

			err := s.CreateChangesetJob(ctx, job)
			if err != nil {
				t.Fatal(err)
			}
			var runCount int64
			errChan := make(chan error, 2)

			for i := 0; i < 2; i++ {
				go func() {
					ran, err := s.ProcessPendingChangesetJobs(ctx, process)
					if ran {
						atomic.AddInt64(&runCount, 1)
					}
					errChan <- err
				}()
			}
			for i := 0; i < 2; i++ {
				err := <-errChan
				if err != nil && err.Error() != "rollback" {
					t.Fatal(err)
				}
			}

			rc := atomic.LoadInt64(&runCount)
			if rc != 1 {
				t.Errorf("Want %d, got %d", 1, rc)
			}
		})
	}
}

func testStoreLocking(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := NewStoreWithClock(db, func() time.Time {
			return now.UTC().Truncate(time.Microsecond)
		})

		testKey := "test-acquire"
		s1, err := s.Transact(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer s1.Done(nil)

		s2, err := s.Transact(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer s2.Done(nil)

		// Get lock
		ok, err := s1.TryAcquireAdvisoryLock(context.Background(), testKey)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("Could not acquire lock")
		}

		// Try and get acquired lock
		ok, err = s2.TryAcquireAdvisoryLock(context.Background(), testKey)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatal("Should not have acquired lock")
		}

		// Release lock
		s1.Done(nil)

		// Try and get released lock
		ok, err = s2.TryAcquireAdvisoryLock(context.Background(), testKey)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("Could not acquire lock")
		}
	}
}
