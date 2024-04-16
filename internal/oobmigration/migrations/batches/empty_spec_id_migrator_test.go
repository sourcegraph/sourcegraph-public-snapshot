package batches

import (
	"context"
	"strings"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	bstore "github.com/sourcegraph/sourcegraph/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestEmptySpecIDMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	s := bstore.New(db, observation.TestContextTB(t), nil)

	migrator := NewEmptySpecIDMigrator(s.Store)
	progress, err := migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress with no DB entries, want=%f have=%f", want, have)
	}

	user := bt.CreateTestUser(t, db, true)

	testData := []struct {
		bcName string
		// We use IDs in the 1000s to avoid collisions with the auto-incrementing ID of
		// the spec inserted with the store method.
		initialEmptyIDs []int64
		nonEmptyIDs     []int64
		wantEmptyID     int64
	}{
		// A batch change that only has one spec, which is an empty one. The ID of the
		// spec should not change.
		{bcName: "test-batch-change-0",
			initialEmptyIDs: []int64{1001},
			nonEmptyIDs:     []int64{},
			wantEmptyID:     1001},
		// A batch change that has one non-empty spec that suceeds the empty one. Since
		// the empty spec is already ordered first, its ID should not change.
		{bcName: "test-batch-change-1",
			initialEmptyIDs: []int64{1011},
			nonEmptyIDs:     []int64{1012},
			wantEmptyID:     1011},
		// A batch change that has one non-empty spec that precedes the empty one. Since
		// the empty spec is out-of-order, it should be assigned the first available ID
		// lower than 1012, which in this case is 1011.
		{bcName: "test-batch-change-2",
			initialEmptyIDs: []int64{1022},
			nonEmptyIDs:     []int64{1021},
			wantEmptyID:     1020},
		// A batch change that has multiple non-empty specs that suceed the empty one.
		// Since the empty spec is already ordered first, its ID should not change.
		{bcName: "test-batch-change-3",
			initialEmptyIDs: []int64{1031},
			nonEmptyIDs:     []int64{1032, 1033, 1034},
			wantEmptyID:     1031},
		// Two batch changes that have multiple, interweaving, non-empty and empty specs.
		{bcName: "test-batch-change-4",
			initialEmptyIDs: []int64{1045, 1051},
			nonEmptyIDs:     []int64{1043, 1048, 1050},
			// Since neither empty spec was in order, once they have been de-duped, we
			// expect the remaining empty spec to be assigned the first available ID lower
			// than 1043, which is 1041.
			wantEmptyID: 1041},
		{bcName: "test-batch-change-5",
			initialEmptyIDs: []int64{1040, 1099},
			nonEmptyIDs:     []int64{1042, 1044, 1047},
			// Since one of the empty specs was in order, we expect it to not change, but
			// the other spec to be de-duped.
			wantEmptyID: 1040},
	}

	for _, tc := range testData {
		for _, id := range tc.initialEmptyIDs {
			emptySpec := bt.CreateEmptyBatchSpec(t, ctx, s, tc.bcName, user.ID, 0)
			if id != emptySpec.ID {
				err = s.Exec(ctx, sqlf.Sprintf("UPDATE batch_specs SET id = %d WHERE id = %d", id, emptySpec.ID))
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		batchChange := &btypes.BatchChange{
			CreatorID:       user.ID,
			NamespaceUserID: user.ID,
			BatchSpecID:     tc.initialEmptyIDs[0],
			Name:            tc.bcName,
		}
		if err := s.CreateBatchChange(ctx, batchChange); err != nil {
			t.Fatal(err)
		}
		for _, id := range tc.nonEmptyIDs {
			spec := bt.CreateBatchSpec(t, ctx, s, tc.bcName, user.ID, 0)
			if id != spec.ID {
				err = s.Exec(ctx, sqlf.Sprintf("UPDATE batch_specs SET id = %d WHERE id = %d", id, spec.ID))
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT count(*) FROM batch_specs")))
	if err != nil {
		t.Fatal(err)
	}
	if count != 19 {
		t.Fatalf("got %d batch specs, want %d", count, 19)
	}

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	// We expect to start with progress at 50% because 4 of the 8 empty batch specs are
	// already in the correct order.
	if have, want := progress, 0.5; have != want {
		t.Fatalf("got invalid progress with unmigrated entries, want=%f have=%f", want, have)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress after up migration, want=%f have=%f", want, have)
	}

	for _, tc := range testData {
		// Check that we can find the empty spec with its new ID.
		emptySpec, err := s.GetBatchSpec(ctx, bstore.GetBatchSpecOpts{ID: tc.wantEmptyID})
		if err != nil {
			t.Fatalf("could not locate empty spec with ID %d after migration", tc.wantEmptyID)
		}
		wantRaw := "name: " + tc.bcName
		gotRaw := strings.Trim(emptySpec.RawSpec, "\n")
		if gotRaw != wantRaw {
			t.Fatalf("empty spec has wrong raw spec. got %q, want %q", gotRaw, wantRaw)
		}

		// If we updated the ID, check that we _can't_ find the empty spec with its old ID.
		if tc.initialEmptyIDs[0] != tc.wantEmptyID {
			for _, id := range tc.initialEmptyIDs {
				_, err = s.GetBatchSpec(ctx, bstore.GetBatchSpecOpts{ID: id})
				if err == nil {
					t.Fatalf("empty spec still found with original ID %d after migration", id)
				}
			}
		}

		// Check that batch change has the new batch spec ID assigned.
		batchChange, err := s.GetBatchChange(ctx, bstore.GetBatchChangeOpts{Name: tc.bcName})
		if err != nil {
			t.Fatal(err)
		}
		if batchChange.BatchSpecID != tc.wantEmptyID {
			t.Fatalf("got batch change with batch spec ID %d, want %d", batchChange.BatchSpecID, tc.wantEmptyID)
		}

	}
}
