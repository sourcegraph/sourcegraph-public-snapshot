package worker

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

func TestDequeueAndProcessNoUpload(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueFunc.SetDefaultReturn(db.Upload{}, nil, false, nil)

	worker := &Worker{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewWorkerMetrics(metrics.TestRegisterer),
	}

	dequeued, err := worker.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing upload: %s", err)
	}
	if dequeued {
		t.Errorf("unexpected upload dequeued")
	}
}

func TestDequeueAndProcessSuccess(t *testing.T) {
	upload := db.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueFunc.SetDefaultReturn(upload, mockDB, true, nil)

	worker := &Worker{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewWorkerMetrics(metrics.TestRegisterer),
	}

	dequeued, err := worker.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing upload: %s", err)
	}
	if !dequeued {
		t.Errorf("expected upload dequeue")
	}
	if len(mockDB.MarkErroredFunc.History()) != 0 {
		t.Errorf("unexpected call to MarkErrored")
	}
	if len(mockDB.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockDB.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error to Done: %s", doneErr)
	}
}

func TestDequeueAndProcessProcessFailure(t *testing.T) {
	upload := db.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueFunc.SetDefaultReturn(upload, mockDB, true, nil)
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("process failure"))

	worker := &Worker{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewWorkerMetrics(metrics.TestRegisterer),
	}

	dequeued, err := worker.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing upload: %s", err)
	}
	if !dequeued {
		t.Errorf("expected upload dequeue")
	}
	if len(mockDB.MarkErroredFunc.History()) != 1 {
		t.Errorf("expected call to MarkErrored")
	} else if errText := mockDB.MarkErroredFunc.History()[0].Arg2; errText != "process failure" {
		t.Errorf("unexpected failure text. want=%q have=%q", "process failure", errText)
	}
	if len(mockDB.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockDB.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error to Done: %s", doneErr)
	}
}

func TestDequeueAndProcessMarkErrorFailure(t *testing.T) {
	upload := db.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockDB := dbmocks.NewMockDB()
	mockDB.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockProcessor := NewMockProcessor()
	mockDB.DequeueFunc.SetDefaultReturn(upload, mockDB, true, nil)
	mockDB.MarkErroredFunc.SetDefaultReturn(fmt.Errorf("db failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("failed"))

	worker := &Worker{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewWorkerMetrics(metrics.TestRegisterer),
	}

	_, err := worker.dequeueAndProcess(context.Background())
	if err == nil || !strings.Contains(err.Error(), "db failure") {
		t.Errorf("unexpected error to Done. want=%q have=%q", "db failure", err)
	}
	if len(mockDB.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockDB.DoneFunc.History()[0].Arg0; doneErr != nil && !strings.Contains(doneErr.Error(), "db failure") {
		t.Errorf("unexpected error to Done. want=%q have=%q", "db failure", doneErr)
	}
}

func TestDequeueAndProcessDoneFailure(t *testing.T) {
	upload := db.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueFunc.SetDefaultReturn(upload, mockDB, true, nil)
	mockDB.DoneFunc.SetDefaultReturn(fmt.Errorf("db failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("failed"))

	worker := &Worker{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewWorkerMetrics(metrics.TestRegisterer),
	}

	_, err := worker.dequeueAndProcess(context.Background())
	if err == nil || !strings.Contains(err.Error(), "db failure") {
		t.Errorf("unexpected error to Done. want=%q have=%q", "db failure", err)
	}
}
