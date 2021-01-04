package store

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func testStoreChangesets(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
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
		CreatedAt:    clock.Now(),
		UpdatedAt:    clock.Now(),
		HeadRefName:  "campaigns/test",
	}

	rs := db.NewRepoStoreWith(s)
	es := db.NewExternalServicesStoreWith(s)

	repo := ct.TestRepo(t, es, extsvc.KindGitHub)
	otherRepo := ct.TestRepo(t, es, extsvc.KindGitHub)
	gitlabRepo := ct.TestRepo(t, es, extsvc.KindGitLab)

	if err := rs.Create(ctx, repo, otherRepo, gitlabRepo); err != nil {
		t.Fatal(err)
	}
	deletedRepo := otherRepo.With(types.Opt.RepoDeletedAt(clock.Now()))
	if err := rs.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fatal(err)
	}

	changesets := make(campaigns.Changesets, 0, 3)

	deletedRepoChangeset := &campaigns.Changeset{
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
			failureMessage := fmt.Sprintf("failure-%d", i)
			th := &campaigns.Changeset{
				RepoID:              repo.ID,
				CreatedAt:           clock.Now(),
				UpdatedAt:           clock.Now(),
				Metadata:            githubPR,
				CampaignIDs:         []int64{int64(i) + 1},
				ExternalID:          fmt.Sprintf("foobar-%d", i),
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalBranch:      fmt.Sprintf("refs/heads/campaigns/test/%d", i),
				ExternalUpdatedAt:   clock.Now(),
				ExternalState:       campaigns.ChangesetExternalStateOpen,
				ExternalReviewState: campaigns.ChangesetReviewStateApproved,
				ExternalCheckState:  campaigns.ChangesetCheckStatePassed,

				CurrentSpecID:     int64(i) + 1,
				PreviousSpecID:    int64(i) + 1,
				OwnedByCampaignID: int64(i) + 1,
				PublicationState:  campaigns.ChangesetPublicationStatePublished,

				ReconcilerState: campaigns.ReconcilerStateCompleted,
				FailureMessage:  &failureMessage,
				NumResets:       18,
				NumFailures:     25,

				Closing: true,
			}

			if i != 0 {
				th.PublicationState = campaigns.ChangesetPublicationStateUnpublished
			}

			// Only set these fields on a subset to make sure that
			// we handle nil pointers correctly
			if i != cap(changesets)-1 {
				th.DiffStatAdded = &added
				th.DiffStatChanged = &changed
				th.DiffStatDeleted = &deleted

				th.StartedAt = clock.Now()
				th.FinishedAt = clock.Now()
				th.ProcessAfter = clock.Now()
			}

			if err := s.CreateChangeset(ctx, th); err != nil {
				t.Fatal(err)
			}

			changesets = append(changesets, th)
		}

		if err := s.CreateChangeset(ctx, deletedRepoChangeset); err != nil {
			t.Fatal(err)
		}

		for _, have := range changesets {
			if have.ID == 0 {
				t.Fatal("id should not be zero")
			}

			if have.IsDeleted() {
				t.Fatal("changeset is deleted")
			}

			if !have.ReconcilerState.Valid() {
				t.Fatalf("reconciler state is invalid: %s", have.ReconcilerState)
			}

			want := have.Clone()

			want.ID = have.ID
			want.CreatedAt = clock.Now()
			want.UpdatedAt = clock.Now()

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		changeset := &campaigns.Changeset{
			RepoID:              repo.ID,
			CreatedAt:           clock.Now(),
			UpdatedAt:           clock.Now(),
			Metadata:            githubPR,
			CampaignIDs:         []int64{1},
			ExternalID:          "foobar-123",
			ExternalServiceType: extsvc.TypeGitHub,
			ExternalBranch:      "refs/heads/campaigns/test",
			ExternalUpdatedAt:   clock.Now(),
			ExternalState:       campaigns.ChangesetExternalStateOpen,
			ExternalReviewState: campaigns.ChangesetReviewStateApproved,
			ExternalCheckState:  campaigns.ChangesetCheckStatePassed,
			PreviousSpecID:      1,
			OwnedByCampaignID:   1,
			PublicationState:    campaigns.ChangesetPublicationStatePublished,
			ReconcilerState:     campaigns.ReconcilerStateCompleted,
			StartedAt:           clock.Now(),
			FinishedAt:          clock.Now(),
			ProcessAfter:        clock.Now(),
		}

		if err := s.UpsertChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		if changeset.ID == 0 {
			t.Fatal("id should not be zero")
		}

		prev := changeset.Clone()

		if err := s.UpsertChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(changeset, prev); diff != "" {
			t.Fatal(diff)
		}

		if err := s.DeleteChangeset(ctx, changeset.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ReconcilerState database representation", func(t *testing.T) {
		// campaigns.ReconcilerStates are defined as "enum" string constants.
		// The string values are uppercase, because that way they can easily be
		// serialized/deserialized in the GraphQL resolvers, since GraphQL
		// expects the `ChangesetReconcilerState` values to be uppercase.
		//
		// But workerutil.Worker expects those values to be lowercase.
		//
		// So, what we do is to lowercase the Changeset.ReconcilerState value
		// before it enters the database and uppercase it when it leaves the
		// DB.
		//
		// If workerutil.Worker supports custom mappings for the state-machine
		// states, we can remove this.

		// This test ensures that the database representation is lowercase.

		queryRawReconcilerState := func(ch *campaigns.Changeset) (string, error) {
			q := sqlf.Sprintf("SELECT reconciler_state FROM changesets WHERE id = %s", ch.ID)
			rawState, ok, err := basestore.ScanFirstString(s.Query(ctx, q))
			if err != nil || !ok {
				return rawState, err
			}
			return rawState, nil
		}

		for _, ch := range changesets {
			have, err := queryRawReconcilerState(ch)
			if err != nil {
				t.Fatal(err)
			}

			want := strings.ToLower(string(ch.ReconcilerState))
			if have != want {
				t.Fatalf("wrong database representation. want=%q, have=%q", want, have)
			}
		}
	})

	t.Run("GetChangesetExternalIDs", func(t *testing.T) {
		refs := make([]string, len(changesets))
		for i, c := range changesets {
			refs[i] = c.ExternalBranch
		}
		have, err := s.GetChangesetExternalIDs(ctx, repo.ExternalRepo, refs)
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
		var want []string
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
		var want []string
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
		var want []string
		if diff := cmp.Diff(want, have); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("No options", func(t *testing.T) {
			count, err := s.CountChangesets(ctx, CountChangesetsOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, len(changesets); have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}

		})

		t.Run("CampaignID", func(t *testing.T) {
			count, err := s.CountChangesets(ctx, CountChangesetsOpts{CampaignID: 1})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, 1; have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})

		t.Run("ReconcilerState", func(t *testing.T) {
			completed := campaigns.ReconcilerStateCompleted
			countCompleted, err := s.CountChangesets(ctx, CountChangesetsOpts{ReconcilerStates: []campaigns.ReconcilerState{completed}})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := countCompleted, len(changesets); have != want {
				t.Fatalf("have countCompleted: %d, want: %d", have, want)
			}

			processing := campaigns.ReconcilerStateProcessing
			countProcessing, err := s.CountChangesets(ctx, CountChangesetsOpts{ReconcilerStates: []campaigns.ReconcilerState{processing}})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := countProcessing, 0; have != want {
				t.Fatalf("have countProcessing: %d, want: %d", have, want)
			}
		})

		t.Run("OwnedByCampaignID", func(t *testing.T) {
			count, err := s.CountChangesets(ctx, CountChangesetsOpts{OwnedByCampaignID: int64(1)})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := count, 1; have != want {
				t.Fatalf("have count: %d, want: %d", have, want)
			}
		})
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
			ts, next, err := s.ListChangesets(ctx, ListChangesetsOpts{LimitOpts: LimitOpts{Limit: i}})
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
				opts := ListChangesetsOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
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
				c.UpdatedAt = clock.Now()

				if err := s.UpdateChangeset(ctx, c); err != nil {
					t.Fatal(err)
				}
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
			gitlabMR := &gitlab.MergeRequest{
				ID:        gitlab.ID(1),
				Title:     "Fix a bunch of bugs",
				CreatedAt: gitlab.Time{Time: clock.Now()},
				UpdatedAt: gitlab.Time{Time: clock.Now()},
			}
			gitlabChangeset := &campaigns.Changeset{
				Metadata:            gitlabMR,
				RepoID:              gitlabRepo.ID,
				ExternalServiceType: extsvc.TypeGitLab,
			}
			if err := s.CreateChangeset(ctx, gitlabChangeset); err != nil {
				t.Fatal(err)
			}
			have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{ExternalServiceID: "https://gitlab.com/"})
			if err != nil {
				t.Fatal(err)
			}

			want := 1
			if len(have) != want {
				t.Fatalf("have %d changesets; want %d", len(have), want)
			}

			if have[0].ID != gitlabChangeset.ID {
				t.Fatalf("unexpected changeset: have %+v; want %+v", have[0], gitlabChangeset)
			}
			if err := s.DeleteChangeset(ctx, gitlabChangeset.ID); err != nil {
				t.Fatal(err)
			}
		}

		// No Limit should return all Changesets
		{
			have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{})
			if err != nil {
				t.Fatal(err)
			}

			if len(have) != 3 {
				t.Fatalf("have %d changesets. want 3", len(have))
			}
		}

		statePublished := campaigns.ChangesetPublicationStatePublished
		stateUnpublished := campaigns.ChangesetPublicationStateUnpublished
		stateQueued := campaigns.ReconcilerStateQueued
		stateCompleted := campaigns.ReconcilerStateCompleted
		stateOpen := campaigns.ChangesetExternalStateOpen
		stateClosed := campaigns.ChangesetExternalStateClosed
		stateApproved := campaigns.ChangesetReviewStateApproved
		stateChangesRequested := campaigns.ChangesetReviewStateChangesRequested
		statePassed := campaigns.ChangesetCheckStatePassed
		stateFailed := campaigns.ChangesetCheckStateFailed

		filterCases := []struct {
			opts      ListChangesetsOpts
			wantCount int
		}{
			{
				opts: ListChangesetsOpts{
					PublicationState: &statePublished,
				},
				wantCount: 1,
			},
			{
				opts: ListChangesetsOpts{
					PublicationState: &stateUnpublished,
				},
				wantCount: 2,
			},
			{
				opts: ListChangesetsOpts{
					ReconcilerStates: []campaigns.ReconcilerState{stateQueued},
				},
				wantCount: 0,
			},
			{
				opts: ListChangesetsOpts{
					ReconcilerStates: []campaigns.ReconcilerState{stateCompleted},
				},
				wantCount: 3,
			},
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
			{
				opts: ListChangesetsOpts{
					OwnedByCampaignID: int64(1),
				},
				wantCount: 1,
			},
		}

		for i, tc := range filterCases {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				have, _, err := s.ListChangesets(ctx, tc.opts)
				if err != nil {
					t.Fatal(err)
				}
				if len(have) != tc.wantCount {
					t.Fatalf("opts: %+v. have %d changesets. want %d", tc.opts, len(have), tc.wantCount)
				}
			})
		}
	})

	t.Run("Null changeset external state", func(t *testing.T) {
		cs := &campaigns.Changeset{
			RepoID:              repo.ID,
			Metadata:            githubPR,
			CampaignIDs:         []int64{1},
			ExternalID:          fmt.Sprintf("foobar-%d", 42),
			ExternalServiceType: extsvc.TypeGitHub,
			ExternalBranch:      "refs/heads/campaigns/test",
			ExternalUpdatedAt:   clock.Now(),
			ExternalState:       "",
			ExternalReviewState: "",
			ExternalCheckState:  "",
		}

		err := s.CreateChangeset(ctx, cs)
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

		t.Run("ExternalBranch", func(t *testing.T) {
			for _, c := range changesets {
				opts := GetChangesetOpts{ExternalBranch: c.ExternalBranch}

				have, err := s.GetChangeset(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				want := c

				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		want := make([]*campaigns.Changeset, 0, len(changesets))
		have := make([]*campaigns.Changeset, 0, len(changesets))

		clock.Add(1 * time.Second)
		for _, c := range changesets {
			c.Metadata = &bitbucketserver.PullRequest{ID: 1234}
			c.ExternalServiceType = extsvc.TypeBitbucketServer

			c.CurrentSpecID = c.CurrentSpecID + 1
			c.PreviousSpecID = c.PreviousSpecID + 1
			c.OwnedByCampaignID = c.OwnedByCampaignID + 1

			c.PublicationState = campaigns.ChangesetPublicationStatePublished
			c.ReconcilerState = campaigns.ReconcilerStateErrored
			c.FailureMessage = nil
			c.StartedAt = clock.Now()
			c.FinishedAt = clock.Now()
			c.ProcessAfter = clock.Now()
			c.NumResets = 987
			c.NumFailures = 789

			clone := c.Clone()
			have = append(have, clone)

			c.UpdatedAt = clock.Now()
			want = append(want, c)

			if err := s.UpdateChangeset(ctx, clone); err != nil {
				t.Fatal(err)
			}
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatal(diff)
		}

		for i := range have {
			// Test that duplicates are not introduced.
			have[i].CampaignIDs = append(have[i].CampaignIDs, have[i].CampaignIDs...)

			if err := s.UpdateChangeset(ctx, have[i]); err != nil {
				t.Fatal(err)
			}

		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatal(diff)
		}

		for i := range have {
			// Test we can add to the set.
			have[i].CampaignIDs = append(have[i].CampaignIDs, 42)
			want[i].CampaignIDs = append(want[i].CampaignIDs, 42)

			if err := s.UpdateChangeset(ctx, have[i]); err != nil {
				t.Fatal(err)
			}

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

			if err := s.UpdateChangeset(ctx, have[i]); err != nil {
				t.Fatal(err)
			}
		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatal(diff)
		}

		clock.Add(1 * time.Second)
		want = want[0:0]
		have = have[0:0]
		for _, c := range changesets {
			c.Metadata = &gitlab.MergeRequest{ID: 1234, IID: 123}
			c.ExternalServiceType = extsvc.TypeGitLab

			clone := c.Clone()
			have = append(have, clone)

			c.UpdatedAt = clock.Now()
			want = append(want, c)

			if err := s.UpdateChangeset(ctx, clone); err != nil {
				t.Fatal(err)
			}

		}

		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("CancelQueuedCampaignChangesets", func(t *testing.T) {
		var campaignID int64 = 99999

		c1 := ct.CreateChangeset(t, ctx, s, ct.TestChangesetOpts{
			Repo:            repo.ID,
			Campaign:        campaignID,
			OwnedByCampaign: campaignID,
			ReconcilerState: campaigns.ReconcilerStateQueued,
		})

		c2 := ct.CreateChangeset(t, ctx, s, ct.TestChangesetOpts{
			Repo:            repo.ID,
			Campaign:        campaignID,
			OwnedByCampaign: campaignID,
			ReconcilerState: campaigns.ReconcilerStateErrored,
			NumFailures:     1,
		})

		c3 := ct.CreateChangeset(t, ctx, s, ct.TestChangesetOpts{
			Repo:            repo.ID,
			Campaign:        campaignID,
			OwnedByCampaign: campaignID,
			ReconcilerState: campaigns.ReconcilerStateCompleted,
		})

		c4 := ct.CreateChangeset(t, ctx, s, ct.TestChangesetOpts{
			Repo:             repo.ID,
			Campaign:         campaignID,
			OwnedByCampaign:  0,
			PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			ReconcilerState:  campaigns.ReconcilerStateQueued,
		})

		c5 := ct.CreateChangeset(t, ctx, s, ct.TestChangesetOpts{
			Repo:            repo.ID,
			Campaign:        campaignID,
			OwnedByCampaign: campaignID,
			ReconcilerState: campaigns.ReconcilerStateProcessing,
		})

		if err := s.CancelQueuedCampaignChangesets(ctx, campaignID); err != nil {
			t.Fatal(err)
		}

		ct.ReloadAndAssertChangeset(t, ctx, s, c1, ct.ChangesetAssertions{
			Repo:            repo.ID,
			ReconcilerState: campaigns.ReconcilerStateFailed,
			OwnedByCampaign: campaignID,
			FailureMessage:  &CanceledChangesetFailureMessage,
		})

		ct.ReloadAndAssertChangeset(t, ctx, s, c2, ct.ChangesetAssertions{
			Repo:            repo.ID,
			ReconcilerState: campaigns.ReconcilerStateFailed,
			OwnedByCampaign: campaignID,
			FailureMessage:  &CanceledChangesetFailureMessage,
			NumFailures:     1,
		})

		ct.ReloadAndAssertChangeset(t, ctx, s, c3, ct.ChangesetAssertions{
			Repo:            repo.ID,
			ReconcilerState: campaigns.ReconcilerStateCompleted,
			OwnedByCampaign: campaignID,
		})

		ct.ReloadAndAssertChangeset(t, ctx, s, c4, ct.ChangesetAssertions{
			Repo:             repo.ID,
			ReconcilerState:  campaigns.ReconcilerStateQueued,
			PublicationState: campaigns.ChangesetPublicationStateUnpublished,
		})

		ct.ReloadAndAssertChangeset(t, ctx, s, c5, ct.ChangesetAssertions{
			Repo:            repo.ID,
			ReconcilerState: campaigns.ReconcilerStateFailed,
			FailureMessage:  &CanceledChangesetFailureMessage,
			OwnedByCampaign: campaignID,
		})
	})

	t.Run("EnqueueChangesetsToClose", func(t *testing.T) {
		var campaignID int64 = 99999

		wantEnqueued := ct.ChangesetAssertions{
			Repo:             repo.ID,
			OwnedByCampaign:  campaignID,
			ReconcilerState:  campaigns.ReconcilerStateQueued,
			PublicationState: campaigns.ChangesetPublicationStatePublished,
			NumFailures:      0,
			FailureMessage:   nil,
			Closing:          true,
		}

		tests := []struct {
			have ct.TestChangesetOpts
			want ct.ChangesetAssertions
		}{
			{
				have: ct.TestChangesetOpts{
					ReconcilerState:  campaigns.ReconcilerStateQueued,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
				},
				want: wantEnqueued,
			},
			{
				have: ct.TestChangesetOpts{
					ReconcilerState:  campaigns.ReconcilerStateProcessing,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
				},
				want: wantEnqueued,
			},
			{
				have: ct.TestChangesetOpts{
					ReconcilerState:  campaigns.ReconcilerStateErrored,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
					FailureMessage:   "failed",
					NumFailures:      1,
				},
				want: wantEnqueued,
			},
			{
				have: ct.TestChangesetOpts{
					ExternalState:    campaigns.ChangesetExternalStateOpen,
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
				},
				want: ct.ChangesetAssertions{
					ReconcilerState:  campaigns.ReconcilerStateQueued,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
					Closing:          true,
					ExternalState:    campaigns.ChangesetExternalStateOpen,
				},
			},
			{
				have: ct.TestChangesetOpts{
					ExternalState:    campaigns.ChangesetExternalStateClosed,
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
				},
				want: ct.ChangesetAssertions{
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					ExternalState:    campaigns.ChangesetExternalStateClosed,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
				},
			},
			{
				have: ct.TestChangesetOpts{
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				},
				want: ct.ChangesetAssertions{
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				},
			},
		}

		changesets := make(map[*campaigns.Changeset]ct.ChangesetAssertions)
		for _, tc := range tests {
			opts := tc.have
			opts.Repo = repo.ID
			opts.Campaign = campaignID
			opts.OwnedByCampaign = campaignID

			c := ct.CreateChangeset(t, ctx, s, opts)
			changesets[c] = tc.want
		}

		if err := s.EnqueueChangesetsToClose(ctx, campaignID); err != nil {
			t.Fatal(err)
		}

		for changeset, want := range changesets {
			want.Repo = repo.ID
			want.OwnedByCampaign = campaignID
			ct.ReloadAndAssertChangeset(t, ctx, s, changeset, want)
		}
	})

	t.Run("GetChangesetsStats", func(t *testing.T) {
		currentStats, err := s.GetChangesetsStats(ctx, GetChangesetsStatsOpts{})
		if err != nil {
			t.Fatal(err)
		}
		var campaignID int64 = 191918
		currentCampaignStats, err := s.GetChangesetsStats(ctx, GetChangesetsStatsOpts{CampaignID: campaignID})
		if err != nil {
			t.Fatal(err)
		}

		baseOpts := ct.TestChangesetOpts{Repo: repo.ID}

		opts1 := baseOpts
		opts1.Campaign = campaignID
		opts1.ExternalState = campaigns.ChangesetExternalStateClosed
		opts1.PublicationState = campaigns.ChangesetPublicationStatePublished
		ct.CreateChangeset(t, ctx, s, opts1)

		opts2 := baseOpts
		opts2.Campaign = campaignID
		opts2.ExternalState = campaigns.ChangesetExternalStateDeleted
		opts2.PublicationState = campaigns.ChangesetPublicationStatePublished
		ct.CreateChangeset(t, ctx, s, opts2)

		opts3 := baseOpts
		opts3.Campaign = campaignID
		opts3.OwnedByCampaign = campaignID
		opts3.ExternalState = campaigns.ChangesetExternalStateOpen
		opts3.PublicationState = campaigns.ChangesetPublicationStatePublished
		ct.CreateChangeset(t, ctx, s, opts3)

		opts4 := baseOpts
		// In a deleted repository.
		opts4.Repo = deletedRepo.ID
		opts4.Campaign = campaignID
		opts4.ExternalState = campaigns.ChangesetExternalStateOpen
		opts4.PublicationState = campaigns.ChangesetPublicationStatePublished
		ct.CreateChangeset(t, ctx, s, opts4)

		opts5 := baseOpts
		// In a different campaign.
		opts5.Campaign = campaignID + 999
		opts5.ExternalState = campaigns.ChangesetExternalStateOpen
		opts5.PublicationState = campaigns.ChangesetPublicationStatePublished
		ct.CreateChangeset(t, ctx, s, opts5)

		t.Run("global", func(t *testing.T) {
			haveStats, err := s.GetChangesetsStats(ctx, GetChangesetsStatsOpts{})
			if err != nil {
				t.Fatal(err)
			}

			wantStats := currentStats
			wantStats.Open += 2
			wantStats.Closed += 1
			wantStats.Deleted += 1
			wantStats.Total += 4

			if diff := cmp.Diff(wantStats, haveStats); diff != "" {
				t.Fatalf("wrong stats returned. diff=%s", diff)
			}
		})
		t.Run("single campaign", func(t *testing.T) {
			haveStats, err := s.GetChangesetsStats(ctx, GetChangesetsStatsOpts{CampaignID: campaignID})
			if err != nil {
				t.Fatal(err)
			}

			wantStats := currentCampaignStats
			wantStats.Open += 1
			wantStats.Closed += 1
			wantStats.Deleted += 1
			wantStats.Total += 3

			if diff := cmp.Diff(wantStats, haveStats); diff != "" {
				t.Fatalf("wrong stats returned. diff=%s", diff)
			}
		})
	})
}

func testStoreListChangesetSyncData(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
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
		CreatedAt:    clock.Now(),
		UpdatedAt:    clock.Now(),
		HeadRefName:  "campaigns/test",
	}
	gitlabMR := &gitlab.MergeRequest{
		ID:        gitlab.ID(1),
		Title:     "Fix a bunch of bugs",
		CreatedAt: gitlab.Time{Time: clock.Now()},
		UpdatedAt: gitlab.Time{Time: clock.Now()},
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
		CreatedAt:           clock.Now(),
		UpdatedAt:           clock.Now(),
		IncludesCreatedEdit: false,
	}

	rs := db.NewRepoStoreWith(s)
	es := db.NewExternalServicesStoreWith(s)

	githubRepo := ct.TestRepo(t, es, extsvc.KindGitHub)
	gitlabRepo := ct.TestRepo(t, es, extsvc.KindGitLab)

	if err := rs.Create(ctx, githubRepo, gitlabRepo); err != nil {
		t.Fatal(err)
	}

	changesets := make(campaigns.Changesets, 0, 3)
	events := make([]*campaigns.ChangesetEvent, 0)

	for i := 0; i < cap(changesets); i++ {
		ch := &campaigns.Changeset{
			RepoID:              githubRepo.ID,
			CreatedAt:           clock.Now(),
			UpdatedAt:           clock.Now(),
			Metadata:            githubPR,
			CampaignIDs:         []int64{int64(i) + 1},
			ExternalID:          fmt.Sprintf("foobar-%d", i),
			ExternalServiceType: extsvc.TypeGitHub,
			ExternalBranch:      "refs/heads/campaigns/test",
			ExternalUpdatedAt:   clock.Now(),
			ExternalState:       campaigns.ChangesetExternalStateOpen,
			ExternalReviewState: campaigns.ChangesetReviewStateApproved,
			ExternalCheckState:  campaigns.ChangesetCheckStatePassed,
			PublicationState:    campaigns.ChangesetPublicationStatePublished,
			ReconcilerState:     campaigns.ReconcilerStateCompleted,
		}

		if i == cap(changesets)-1 {
			ch.Metadata = gitlabMR
			ch.ExternalServiceType = extsvc.TypeGitLab
			ch.RepoID = gitlabRepo.ID
		}

		if err := s.CreateChangeset(ctx, ch); err != nil {
			t.Fatal(err)
		}

		changesets = append(changesets, ch)
	}

	// We need campaigns attached to each changeset
	for _, cs := range changesets {
		c := &campaigns.Campaign{
			Name:           "ListChangesetSyncData test",
			NamespaceOrgID: 23,
			LastApplierID:  1,
			LastAppliedAt:  time.Now(),
			CampaignSpecID: 42,
		}
		err := s.CreateCampaign(ctx, c)
		if err != nil {
			t.Fatal(err)
		}
		cs.CampaignIDs = []int64{c.ID}

		if err := s.UpdateChangeset(ctx, cs); err != nil {
			t.Fatal(err)
		}
	}

	// The changesets, except one, get changeset events
	for _, cs := range changesets[:len(changesets)-1] {
		e := &campaigns.ChangesetEvent{
			ChangesetID: cs.ID,
			Kind:        campaigns.ChangesetEventKindGitHubCommented,
			Key:         issueComment.Key(),
			CreatedAt:   clock.Now(),
			Metadata:    issueComment,
		}

		events = append(events, e)
	}
	if err := s.UpsertChangesetEvents(ctx, events...); err != nil {
		t.Fatal(err)
	}

	checkChangesetIDs := func(t *testing.T, hs []*campaigns.ChangesetSyncData, want []int64) {
		t.Helper()

		haveIDs := []int64{}
		for _, sd := range hs {
			haveIDs = append(haveIDs, sd.ChangesetID)
		}
		if diff := cmp.Diff(want, haveIDs); diff != "" {
			t.Fatalf("wrong changesetIDs in changeset sync data (-want +got):\n%s", diff)
		}
	}

	t.Run("success", func(t *testing.T) {
		hs, err := s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		want := []*campaigns.ChangesetSyncData{
			{
				ChangesetID:           changesets[0].ID,
				UpdatedAt:             clock.Now(),
				LatestEvent:           clock.Now(),
				ExternalUpdatedAt:     clock.Now(),
				RepoExternalServiceID: "https://github.com/",
			},
			{
				ChangesetID:           changesets[1].ID,
				UpdatedAt:             clock.Now(),
				LatestEvent:           clock.Now(),
				ExternalUpdatedAt:     clock.Now(),
				RepoExternalServiceID: "https://github.com/",
			},
			{
				// No events
				ChangesetID:           changesets[2].ID,
				UpdatedAt:             clock.Now(),
				ExternalUpdatedAt:     clock.Now(),
				RepoExternalServiceID: "https://gitlab.com/",
			},
		}
		if diff := cmp.Diff(want, hs); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("only for specific external service", func(t *testing.T) {
		hs, err := s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{ExternalServiceID: "https://gitlab.com/"})
		if err != nil {
			t.Fatal(err)
		}
		want := []*campaigns.ChangesetSyncData{
			{
				ChangesetID:           changesets[2].ID,
				UpdatedAt:             clock.Now(),
				ExternalUpdatedAt:     clock.Now(),
				RepoExternalServiceID: "https://gitlab.com/",
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
		c.ClosedAt = clock.Now()
		err = s.UpdateCampaign(ctx, c)
		if err != nil {
			t.Fatal(err)
		}

		hs, err := s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		checkChangesetIDs(t, hs, changesets[1:].IDs())

		// If a changeset has ANY open campaigns we should list it
		// Attach cs1 to both an open and closed campaign
		openCampaignID := changesets[1].CampaignIDs[0]
		changesets[0].CampaignIDs = []int64{closedCampaignID, openCampaignID}
		err = s.UpdateChangeset(ctx, changesets[0])
		if err != nil {
			t.Fatal(err)
		}

		hs, err = s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		checkChangesetIDs(t, hs, changesets.IDs())
	})

	t.Run("ignore processing changesets", func(t *testing.T) {
		ch := changesets[0]
		ch.PublicationState = campaigns.ChangesetPublicationStatePublished
		ch.ReconcilerState = campaigns.ReconcilerStateProcessing
		if err := s.UpdateChangeset(ctx, ch); err != nil {
			t.Fatal(err)
		}

		hs, err := s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		checkChangesetIDs(t, hs, changesets[1:].IDs())
	})

	t.Run("ignore unpublished changesets", func(t *testing.T) {
		ch := changesets[0]
		ch.PublicationState = campaigns.ChangesetPublicationStateUnpublished
		ch.ReconcilerState = campaigns.ReconcilerStateCompleted
		if err := s.UpdateChangeset(ctx, ch); err != nil {
			t.Fatal(err)
		}

		hs, err := s.ListChangesetSyncData(ctx, ListChangesetSyncDataOpts{})
		if err != nil {
			t.Fatal(err)
		}
		checkChangesetIDs(t, hs, changesets[1:].IDs())
	})
}

func testStoreListChangesetsTextSearch(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
	// This is similar to the setup in testStoreChangesets(), but we need a more
	// fine grained set of changesets to handle the different scenarios. Namely,
	// we need to cover:
	//
	// 1. Metadata from each code host type to test title search.
	// 2. Unpublished changesets that don't have metadata to test the title
	//    search fallback to the spec title.
	// 3. Repo name search.
	// 4. Negation of all of the above.

	// Let's define some helpers.
	createChangesetSpec := func(title string) *campaigns.ChangesetSpec {
		spec := &campaigns.ChangesetSpec{
			Spec: &campaigns.ChangesetSpecDescription{
				Title: title,
			},
		}
		if err := s.CreateChangesetSpec(ctx, spec); err != nil {
			t.Fatalf("creating changeset spec: %v", err)
		}
		return spec
	}

	createChangeset := func(
		esType string,
		repo *types.Repo,
		externalID string,
		metadata interface{},
		spec *campaigns.ChangesetSpec,
	) *campaigns.Changeset {
		var specID int64
		if spec != nil {
			specID = spec.ID
		}

		cs := &campaigns.Changeset{
			RepoID:              repo.ID,
			CreatedAt:           clock.Now(),
			UpdatedAt:           clock.Now(),
			Metadata:            metadata,
			ExternalID:          externalID,
			ExternalServiceType: esType,
			ExternalBranch:      "refs/heads/campaigns/test",
			ExternalUpdatedAt:   clock.Now(),
			ExternalState:       campaigns.ChangesetExternalStateOpen,
			ExternalReviewState: campaigns.ChangesetReviewStateApproved,
			ExternalCheckState:  campaigns.ChangesetCheckStatePassed,

			CurrentSpecID:    specID,
			PublicationState: campaigns.ChangesetPublicationStatePublished,
		}

		if err := s.CreateChangeset(ctx, cs); err != nil {
			t.Fatalf("creating changeset:\nerr: %+v\nchangeset: %+v", err, cs)
		}
		return cs
	}

	rs := db.NewRepoStoreWith(s)
	es := db.NewExternalServicesStoreWith(s)

	// Set up repositories for each code host type we want to test.
	var (
		githubRepo = ct.TestRepo(t, es, extsvc.KindGitHub)
		bbsRepo    = ct.TestRepo(t, es, extsvc.KindBitbucketServer)
		gitlabRepo = ct.TestRepo(t, es, extsvc.KindGitLab)
	)
	if err := rs.Create(ctx, githubRepo, bbsRepo, gitlabRepo); err != nil {
		t.Fatal(err)
	}

	// Now let's create ourselves some changesets to test against.
	githubActor := github.Actor{
		AvatarURL: "https://avatars2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}

	githubChangeset := createChangeset(
		extsvc.TypeGitHub,
		githubRepo,
		"12345",
		&github.PullRequest{
			ID:           "FOOBARID",
			Title:        "Fix a bunch of bugs on GitHub",
			Body:         "This fixes a bunch of bugs",
			URL:          "https://github.com/sourcegraph/sourcegraph/pull/12345",
			Number:       12345,
			Author:       githubActor,
			Participants: []github.Actor{githubActor},
			CreatedAt:    clock.Now(),
			UpdatedAt:    clock.Now(),
			HeadRefName:  "campaigns/test",
		},
		createChangesetSpec("Fix a bunch of bugs"),
	)

	gitlabChangeset := createChangeset(
		extsvc.TypeGitLab,
		gitlabRepo,
		"12345",
		&gitlab.MergeRequest{
			ID:           12345,
			IID:          12345,
			ProjectID:    123,
			Title:        "Fix a bunch of bugs on GitLab",
			Description:  "This fixes a bunch of bugs",
			State:        gitlab.MergeRequestStateOpened,
			WebURL:       "https://gitlab.org/sourcegraph/sourcegraph/pull/12345",
			SourceBranch: "campaigns/test",
		},
		createChangesetSpec("Fix a bunch of bugs"),
	)

	bbsChangeset := createChangeset(
		extsvc.TypeBitbucketServer,
		bbsRepo,
		"12345",
		&bitbucketserver.PullRequest{
			ID:          12345,
			Version:     1,
			Title:       "Fix a bunch of bugs on Bitbucket Server",
			Description: "This fixes a bunch of bugs",
			State:       "open",
			Open:        true,
			Closed:      false,
			FromRef:     bitbucketserver.Ref{ID: "campaigns/test"},
		},
		createChangesetSpec("Fix a bunch of bugs"),
	)

	unpublishedChangeset := createChangeset(
		extsvc.TypeGitHub,
		githubRepo,
		"",
		map[string]interface{}{},
		createChangesetSpec("Eventually fix some bugs, but not a bunch"),
	)

	importedChangeset := createChangeset(
		extsvc.TypeGitHub,
		githubRepo,
		"123456",
		&github.PullRequest{
			ID:           "XYZ",
			Title:        "Do some stuff",
			Body:         "This does some stuff",
			URL:          "https://github.com/sourcegraph/sourcegraph/pull/123456",
			Number:       123456,
			Author:       githubActor,
			Participants: []github.Actor{githubActor},
			CreatedAt:    clock.Now(),
			UpdatedAt:    clock.Now(),
			HeadRefName:  "campaigns/stuff",
		},
		nil,
	)

	// All right, let's run some searches!
	for name, tc := range map[string]struct {
		textSearch []ListChangesetsTextSearchExpr
		want       campaigns.Changesets
	}{
		"single changeset based on GitHub metadata title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "on GitHub"},
			},
			want: campaigns.Changesets{githubChangeset},
		},
		"single changeset based on GitLab metadata title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "on GitLab"},
			},
			want: campaigns.Changesets{gitlabChangeset},
		},
		"single changeset based on Bitbucket Server metadata title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "on Bitbucket Server"},
			},
			want: campaigns.Changesets{bbsChangeset},
		},
		"all published changesets based on metadata title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "Fix a bunch of bugs"},
			},
			want: campaigns.Changesets{
				githubChangeset,
				gitlabChangeset,
				bbsChangeset,
			},
		},
		"imported changeset based on metadata title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "Do some stuff"},
			},
			want: campaigns.Changesets{importedChangeset},
		},
		"unpublished changeset based on spec title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "Eventually"},
			},
			want: campaigns.Changesets{unpublishedChangeset},
		},
		"negated metadata title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "bunch of bugs", Not: true},
			},
			want: campaigns.Changesets{
				unpublishedChangeset,
				importedChangeset,
			},
		},
		"negated spec title": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "Eventually", Not: true},
			},
			want: campaigns.Changesets{
				githubChangeset,
				gitlabChangeset,
				bbsChangeset,
				importedChangeset,
			},
		},
		"repo name": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: string(githubRepo.Name)},
			},
			want: campaigns.Changesets{
				githubChangeset,
				unpublishedChangeset,
				importedChangeset,
			},
		},
		"title and repo name together": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: string(githubRepo.Name)},
				{Term: "Eventually"},
			},
			want: campaigns.Changesets{
				unpublishedChangeset,
			},
		},
		"multiple title matches together": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "Eventually"},
				{Term: "fix"},
			},
			want: campaigns.Changesets{
				unpublishedChangeset,
			},
		},
		"negated repo name": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: string(githubRepo.Name), Not: true},
			},
			want: campaigns.Changesets{
				gitlabChangeset,
				bbsChangeset,
			},
		},
		"combined negated repo names": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: string(githubRepo.Name), Not: true},
				{Term: string(gitlabRepo.Name), Not: true},
			},
			want: campaigns.Changesets{bbsChangeset},
		},
		"no results due to conflicting requirements": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: string(githubRepo.Name)},
				{Term: string(gitlabRepo.Name)},
			},
			want: campaigns.Changesets{},
		},
		"no results due to a subset of a word": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "unch"},
			},
			want: campaigns.Changesets{},
		},
		"no results due to text that doesn't exist in the search scope": {
			textSearch: []ListChangesetsTextSearchExpr{
				{Term: "she dreamt she was a bulldozer, she dreamt she was in an empty field"},
			},
			want: campaigns.Changesets{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			have, _, err := s.ListChangesets(ctx, ListChangesetsOpts{
				TextSearch: tc.textSearch,
			})
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			if diff := cmp.Diff(tc.want, have); diff != "" {
				t.Errorf("unexpected result (-want +have):\n%s", diff)
			}
		})
	}
}
