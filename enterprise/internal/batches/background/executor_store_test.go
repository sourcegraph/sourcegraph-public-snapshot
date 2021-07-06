package background

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

func TestLoadAndExtractBatchSpecRandID(t *testing.T) {
	db := dbtest.NewDB(t, "")
	user := ct.CreateTestUser(t, db, true)

	s := store.New(db, nil)
	workStore := dbworkerstore.NewWithMetrics(s.Handle(), executorWorkerStoreOptions, &observation.TestContext)

	t.Run("success", func(t *testing.T) {
		specExec := &btypes.BatchSpecExecution{
			State:         btypes.BatchSpecExecutionStateProcessing,
			BatchSpec:     `name: testing`,
			UserID:        user.ID,
			ExecutionLogs: []workerutil.ExecutionLogEntry{},
		}

		if err := s.CreateBatchSpecExecution(context.Background(), specExec); err != nil {
			t.Fatal(err)
		}

		logEntry := workerutil.ExecutionLogEntry{
			Key:       "step.src.0",
			Command:   []string{"src", "batch", "preview", "-f", "spec.yml", "-text-only"},
			StartTime: time.Now().Add(-5 * time.Second),
			ExitCode:  0,
			Out: `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"SUCCESS"}
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.528Z","status":"STARTED"}
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.535Z","status":"SUCCESS","message":"http://USERNAME_REMOVED:PASSWORD_REMOVED@localhost:3080/users/mrnugget/batch-changes/apply/QmF0Y2hTcGVjOiJBZFBMTDU5SXJmWCI="}
`,
			DurationMs: 200,
		}

		err := workStore.AddExecutionLogEntry(context.Background(), int(specExec.ID), logEntry)
		if err != nil {
			t.Fatal(err)
		}

		want := "AdPLL59IrfX"
		have, err := loadAndExtractBatchSpecRandID(context.Background(), s, specExec.ID)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if have != want {
			t.Fatalf("wrong rand id extracted. want=%s, have=%s", want, have)
		}
	})

	t.Run("without log entry", func(t *testing.T) {
		specExec := &btypes.BatchSpecExecution{
			State:         btypes.BatchSpecExecutionStateProcessing,
			BatchSpec:     `name: testing`,
			UserID:        user.ID,
			ExecutionLogs: []workerutil.ExecutionLogEntry{},
		}

		if err := s.CreateBatchSpecExecution(context.Background(), specExec); err != nil {
			t.Fatal(err)
		}

		_, err := loadAndExtractBatchSpecRandID(context.Background(), s, specExec.ID)
		if err == nil {
			t.Fatalf("expected error but got none")
		}

		if err.Error() != "no execution logs" {
			t.Fatalf("wrong error: %q", err.Error())
		}
	})
}

func TestExtractBatchSpecRandID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Reduced log output because we don't care about _all_ lines
		executionLogOut := `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"SUCCESS"}
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.528Z","status":"STARTED"}
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.535Z","status":"SUCCESS","message":"http://USERNAME_REMOVED:PASSWORD_REMOVED@localhost:3080/users/mrnugget/batch-changes/apply/QmF0Y2hTcGVjOiJBZFBMTDU5SXJmWCI="}
`
		// Run `echo "QmF0Y2hTcGVjOiJBZFBMTDU5SXJmWCI=" |base64 -d` to get this
		want := "AdPLL59IrfX"

		have, err := extractBatchSpecRandID(executionLogOut)
		if err != nil {
			t.Fatal(err)
		}

		if have != want {
			t.Fatalf("wrong batch spec rand id extracted. want=%q, have=%q", want, have)
		}
	})

	t.Run("no url in the output", func(t *testing.T) {
		executionLogOut := `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"SUCCESS"}
`
		_, err := extractBatchSpecRandID(executionLogOut)
		if err != ErrNoBatchSpecRandID {
			t.Fatalf("wrong error: %s", err)
		}
	})

	t.Run("invalid url in the output", func(t *testing.T) {
		executionLogOut := `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"SUCCESS"}
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.528Z","status":"STARTED"}
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.535Z","status":"SUCCESS","message":"http://horse.txt"}
`
		_, err := extractBatchSpecRandID(executionLogOut)
		if err != ErrNoBatchSpecRandID {
			t.Fatalf("wrong error: %s", err)
		}
	})

	t.Run("additional text in log output", func(t *testing.T) {
		executionLogOut := `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stderr: HORSE
stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"SUCCESS"}
stderr: HORSE
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.528Z","status":"STARTED"}
stderr: HORSE
stdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.535Z","status":"SUCCESS","message":"http://USERNAME_REMOVED:PASSWORD_REMOVED@localhost:3080/users/mrnugget/batch-changes/apply/QmF0Y2hTcGVjOiJBZFBMTDU5SXJmWCI="}
`
		want := "AdPLL59IrfX"
		have, err := extractBatchSpecRandID(executionLogOut)
		if err != nil {
			t.Fatal(err)
		}

		if have != want {
			t.Fatalf("wrong batch spec rand id extracted. want=%q, have=%q", want, have)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		executionLogOut := `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {HOOOORSE}
stdout: {HORSE}
stdout: {HORSE}
`
		_, err := extractBatchSpecRandID(executionLogOut)
		if err != ErrNoBatchSpecRandID {
			t.Fatalf("wrong error: %s", err)
		}
	})
}
