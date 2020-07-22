package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
				Name:           fmt.Sprintf("Upgrade ES-Lint %d", i),
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

		t.Run("ByCampaignSpecName", func(t *testing.T) {
			want := campaigns[0]

			campaignSpec := &cmpgn.CampaignSpec{
				Spec:           cmpgn.CampaignSpecFields{Name: "the-name"},
				NamespaceOrgID: want.NamespaceOrgID,
			}
			if err := s.CreateCampaignSpec(ctx, campaignSpec); err != nil {
				t.Fatal(err)
			}

			want.CampaignSpecID = campaignSpec.ID
			if err := s.UpdateCampaign(ctx, want); err != nil {
				t.Fatal(err)
			}

			opts := GetCampaignOpts{CampaignSpecName: campaignSpec.Spec.Name}
			have, err := s.GetCampaign(ctx, opts)
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
				ExternalState:       cmpgn.ChangesetExternalStateOpen,
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

		stateOpen := cmpgn.ChangesetExternalStateOpen
		stateClosed := cmpgn.ChangesetExternalStateClosed
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
			ExternalState:       cmpgn.ChangesetExternalStateOpen,
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
