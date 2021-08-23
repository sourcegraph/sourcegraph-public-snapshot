package store

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

func testStoreChangesetSpecExecutions(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
	testBatchSpec := `theSpec: yeah`

	execs := make([]*btypes.BatchSpecExecution, 0, 2)
	for i := 0; i < cap(execs); i++ {
		c := &btypes.BatchSpecExecution{
			State:           btypes.BatchSpecExecutionStateQueued,
			BatchSpec:       testBatchSpec,
			UserID:          int32(i + 123),
			NamespaceUserID: int32(i + 345),
			BatchSpecID:     int64(i + 567),
		}

		execs = append(execs, c)
	}

	t.Run("Create", func(t *testing.T) {
		for _, exec := range execs {
			if err := s.CreateBatchSpecExecution(ctx, exec); err != nil {
				t.Fatal(err)
			}

			have := exec
			want := &btypes.BatchSpecExecution{
				ID:              have.ID,
				RandID:          have.RandID,
				CreatedAt:       clock.Now(),
				UpdatedAt:       clock.Now(),
				State:           btypes.BatchSpecExecutionStateQueued,
				BatchSpec:       have.BatchSpec,
				UserID:          have.UserID,
				NamespaceUserID: have.NamespaceUserID,
				BatchSpecID:     have.BatchSpecID,
			}

			if have.ID == 0 {
				t.Fatal("ID should not be zero")
			}

			if have.RandID == "" {
				t.Fatal("RandID should not be empty")
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("GetByID", func(t *testing.T) {
			for i, exec := range execs {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					have, err := s.GetBatchSpecExecution(ctx, GetBatchSpecExecutionOpts{ID: exec.ID})
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, exec); diff != "" {
						t.Fatal(diff)
					}
				})
			}
		})

		t.Run("GetByRandID", func(t *testing.T) {
			for i, exec := range execs {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					have, err := s.GetBatchSpecExecution(ctx, GetBatchSpecExecutionOpts{RandID: exec.RandID})
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, exec); diff != "" {
						t.Fatal(diff)
					}
				})
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBatchSpecExecutionOpts{ID: 0xdeadbeef}

			_, have := s.GetBatchSpecExecution(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		execs[0].WorkerHostname = "asdf-host"
		execs[0].Cancel = true
		execs[0].State = btypes.BatchSpecExecutionStateProcessing
		if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_executions SET worker_hostname = %s, cancel = %s, state = %s WHERE id = %s", execs[0].WorkerHostname, execs[0].Cancel, execs[0].State, execs[0].ID)); err != nil {
			t.Fatal(err)
		}
		execs[1].WorkerHostname = "nvm-host"
		if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_executions SET worker_hostname = %s WHERE id = %s", execs[1].WorkerHostname, execs[1].ID)); err != nil {
			t.Fatal(err)
		}

		// The batch spec execution store returns the executions in reversed order.
		reversedBatchSpecExecutions := make([]*btypes.BatchSpecExecution, len(execs))
		for i, c := range execs {
			reversedBatchSpecExecutions[len(execs)-i-1] = c
		}

		t.Run("All", func(t *testing.T) {
			have, _, err := s.ListBatchSpecExecutions(ctx, ListBatchSpecExecutionsOpts{})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(have, reversedBatchSpecExecutions); diff != "" {
				t.Fatalf("invalid executions returned: %s", diff)
			}
		})

		t.Run("WorkerHostname", func(t *testing.T) {
			for _, exec := range reversedBatchSpecExecutions {
				have, _, err := s.ListBatchSpecExecutions(ctx, ListBatchSpecExecutionsOpts{
					WorkerHostname: exec.WorkerHostname,
				})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, []*btypes.BatchSpecExecution{exec}); diff != "" {
					t.Fatalf("invalid executions returned: %s", diff)
				}
			}
		})

		t.Run("State", func(t *testing.T) {
			for _, exec := range reversedBatchSpecExecutions {
				have, _, err := s.ListBatchSpecExecutions(ctx, ListBatchSpecExecutionsOpts{
					State: exec.State,
				})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, []*btypes.BatchSpecExecution{exec}); diff != "" {
					t.Fatalf("invalid executions returned: %s", diff)
				}
			}
		})

		t.Run("Cancel", func(t *testing.T) {
			for _, exec := range reversedBatchSpecExecutions {
				have, _, err := s.ListBatchSpecExecutions(ctx, ListBatchSpecExecutionsOpts{
					Cancel: &exec.Cancel,
				})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, []*btypes.BatchSpecExecution{exec}); diff != "" {
					t.Fatalf("invalid executions returned: %s", diff)
				}
			}
		})

		t.Run("With Limit", func(t *testing.T) {
			for i := 1; i <= len(reversedBatchSpecExecutions); i++ {
				cs, next, err := s.ListBatchSpecExecutions(ctx, ListBatchSpecExecutionsOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(reversedBatchSpecExecutions) {
						want = reversedBatchSpecExecutions[i].ID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := cs, reversedBatchSpecExecutions[:i]
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
			for i := 1; i <= len(reversedBatchSpecExecutions); i++ {
				opts := ListBatchSpecExecutionsOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				have, next, err := s.ListBatchSpecExecutions(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				want := reversedBatchSpecExecutions[i-1 : i]
				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

	})

	t.Run("CancelBatchSpecExecution", func(t *testing.T) {
		t.Run("Queued", func(t *testing.T) {
			record, err := s.CancelBatchSpecExecution(ctx, execs[1].RandID)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := record.State, btypes.BatchSpecExecutionStateFailed; have != want {
				t.Errorf("invalid state: have=%q want=%q", have, want)
			}
			if have, want := record.Cancel, true; have != want {
				t.Errorf("invalid cancel value: have=%t want=%t", have, want)
			}
			if record.FinishedAt == nil {
				t.Error("finished_at not set")
			} else if have, want := *record.FinishedAt, s.now(); !have.Equal(want) {
				t.Errorf("invalid finished_at: have=%s want=%s", have, want)
			}
			if have, want := record.UpdatedAt, s.now(); !have.Equal(want) {
				t.Errorf("invalid updated_at: have=%s want=%s", have, want)
			}
		})
		t.Run("Processing", func(t *testing.T) {
			record, err := s.CancelBatchSpecExecution(ctx, execs[0].RandID)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := record.State, btypes.BatchSpecExecutionStateProcessing; have != want {
				t.Errorf("invalid state: have=%q want=%q", have, want)
			}
			if have, want := record.Cancel, true; have != want {
				t.Errorf("invalid cancel value: have=%t want=%t", have, want)
			}
			if record.FinishedAt != nil {
				t.Error("finished_at set")
			}
			if have, want := record.UpdatedAt, s.now(); !have.Equal(want) {
				t.Errorf("invalid updated_at: have=%s want=%s", have, want)
			}
		})
		t.Run("Invalid current state", func(t *testing.T) {
			if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_executions SET state = 'completed' WHERE id = %s", execs[0].ID)); err != nil {
				t.Fatal(err)
			}
			_, err := s.CancelBatchSpecExecution(ctx, execs[0].RandID)
			if err == nil {
				t.Fatal("got unexpected nil error")
			}
			if err != ErrNoResults {
				t.Fatal(err)
			}
		})
	})
}
