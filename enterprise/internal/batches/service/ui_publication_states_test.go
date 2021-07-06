package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestUiPublicationStates_Add(t *testing.T) {
	var ps UiPublicationStates

	// Add a single publication state, ensuring that ps.rand is initialised.
	if err := ps.Add("foo", batches.PublishedValue{Val: true}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ps.rand) != 1 {
		t.Errorf("unexpected number of elements: %d", len(ps.rand))
	}

	// Add another publication state.
	if err := ps.Add("bar", batches.PublishedValue{Val: true}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ps.rand) != 2 {
		t.Errorf("unexpected number of elements: %d", len(ps.rand))
	}

	// Try to add a duplicate publication state.
	if err := ps.Add("bar", batches.PublishedValue{Val: true}); err == nil {
		t.Error("unexpected nil error")
	}
	if len(ps.rand) != 2 {
		t.Errorf("unexpected number of elements: %d", len(ps.rand))
	}
}

func TestUiPublicationStates_get(t *testing.T) {
	var ps UiPublicationStates

	// Verify that an uninitialised UiPublicationStates can have get() called
	// without panicking.
	ps.get(0)

	ps.id = map[int64]*btypes.ChangesetUiPublicationState{
		1: &btypes.ChangesetUiPublicationStateDraft,
		2: &btypes.ChangesetUiPublicationStateUnpublished,
		3: nil,
	}

	for id, want := range map[int64]*btypes.ChangesetUiPublicationState{
		1: &btypes.ChangesetUiPublicationStateDraft,
		2: &btypes.ChangesetUiPublicationStateUnpublished,
		3: nil,
		4: nil,
	} {
		t.Run(strconv.FormatInt(id, 10), func(t *testing.T) {
			if have := ps.get(id); have != want {
				t.Errorf("unexpected result: have=%v want=%v", have, want)
			}
		})
	}
}

func TestUiPublicationStates_prepare(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtest.NewDB(t, "")

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, nil, clock)

	repos, _ := bt.CreateGitHubSSHTestRepos(t, ctx, db, 1)
	repo := repos[0]
	userID := bt.CreateTestUser(t, db, true).ID

	batchSpecA := bt.CreateBatchSpec(t, ctx, bstore, "a", userID)
	batchSpecB := bt.CreateBatchSpec(t, ctx, bstore, "b", userID)

	changesetSpecA := bt.CreateChangesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		User:      userID,
		Repo:      repo.ID,
		BatchSpec: batchSpecA.ID,
		HeadRef:   "main",
	})
	changesetSpecB := bt.CreateChangesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		User:      userID,
		Repo:      repo.ID,
		BatchSpec: batchSpecB.ID,
		HeadRef:   "main",
		Published: true,
	})

	t.Run("errors", func(t *testing.T) {
		for name, tc := range map[string]struct {
			batchSpecID int64
			setup       func(ps *UiPublicationStates)
		}{
			"incorrect batch spec": {
				batchSpecID: batchSpecB.ID,
				setup: func(ps *UiPublicationStates) {
					ps.Add(changesetSpecA.RandID, batches.PublishedValue{Val: true})
				},
			},
			"spec with published field": {
				batchSpecID: batchSpecB.ID,
				setup: func(ps *UiPublicationStates) {
					ps.Add(changesetSpecB.RandID, batches.PublishedValue{Val: true})
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var ps UiPublicationStates
				tc.setup(&ps)

				if err := ps.prepareAndValidate(ctx, bstore, tc.batchSpecID); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("database error", func(t *testing.T) {
		var ps UiPublicationStates

		ps.Add(changesetSpecA.RandID, batches.PublishedValue{Val: true})
		if err := ps.prepareAndValidate(ctx, &brokenListChangesetSpeccer{}, batchSpecA.ID); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		var ps UiPublicationStates

		ps.Add(changesetSpecA.RandID, batches.PublishedValue{Val: true})
		if err := ps.prepareAndValidate(ctx, bstore, batchSpecA.ID); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(ps.rand) != 0 {
			t.Errorf("unexpected elements remaining in ps.rand: %+v", ps.rand)
		}

		want := map[int64]*btypes.ChangesetUiPublicationState{
			changesetSpecA.ID: &btypes.ChangesetUiPublicationStatePublished,
		}
		if diff := cmp.Diff(want, ps.id); diff != "" {
			t.Errorf("unexpected ps.id (-want +have):\n%s", diff)
		}
	})
}

func TestUiPublicationStates_prepareEmpty(t *testing.T) {
	ctx := context.Background()

	for name, ps := range map[string]UiPublicationStates{
		"nil":   {},
		"empty": {rand: map[string]batches.PublishedValue{}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ps.prepareAndValidate(ctx, nil, 0); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

type brokenListChangesetSpeccer struct{}

var _ ListChangesetSpeccer = &brokenListChangesetSpeccer{}

func (*brokenListChangesetSpeccer) ListChangesetSpecs(context.Context, store.ListChangesetSpecsOpts) (btypes.ChangesetSpecs, int64, error) {
	return nil, 0, errors.New("database error")
}
