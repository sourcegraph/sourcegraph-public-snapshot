package indexer

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

func TestDequeueAndProcessNoIndex(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueIndexFunc.SetDefaultReturn(db.Index{}, nil, false, nil)

	indexer := &Indexer{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewIndexerMetrics(metrics.TestRegisterer),
	}

	dequeued, err := indexer.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing index: %s", err)
	}
	if dequeued {
		t.Errorf("unexpected index dequeued")
	}
}

func TestDequeueAndProcessSuccess(t *testing.T) {
	index := db.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueIndexFunc.SetDefaultReturn(index, mockDB, true, nil)

	indexer := &Indexer{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewIndexerMetrics(metrics.TestRegisterer),
	}

	dequeued, err := indexer.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing index: %s", err)
	}
	if !dequeued {
		t.Errorf("expected index dequeue")
	}
	if len(mockDB.MarkIndexCompleteFunc.History()) != 1 {
		t.Errorf("expected call to MarkIndexComplete")
	}
	if len(mockDB.MarkIndexErroredFunc.History()) != 0 {
		t.Errorf("unexpected call to MarkIndexErrored")
	}
	if len(mockDB.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockDB.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error to Done: %s", doneErr)
	}
}

func TestDequeueAndProcessProcessFailure(t *testing.T) {
	index := db.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueIndexFunc.SetDefaultReturn(index, mockDB, true, nil)
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("process failure"))

	indexer := &Indexer{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewIndexerMetrics(metrics.TestRegisterer),
	}

	dequeued, err := indexer.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing index: %s", err)
	}
	if !dequeued {
		t.Errorf("expected index dequeue")
	}
	if len(mockDB.MarkIndexCompleteFunc.History()) != 0 {
		t.Errorf("unexpected call to MarkIndexComplete")
	}
	if len(mockDB.MarkIndexErroredFunc.History()) != 1 {
		t.Errorf("expected call to MarkIndexErrored")
	} else if errText := mockDB.MarkIndexErroredFunc.History()[0].Arg2; errText != "process failure" {
		t.Errorf("unexpected failure text. want=%q have=%q", "process failure", errText)
	}
	if len(mockDB.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockDB.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error to Done: %s", doneErr)
	}
}

func TestDequeueAndProcessMarkIndexErrorFailure(t *testing.T) {
	index := db.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockDB := dbmocks.NewMockDB()
	mockDB.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockProcessor := NewMockProcessor()
	mockDB.DequeueIndexFunc.SetDefaultReturn(index, mockDB, true, nil)
	mockDB.MarkIndexErroredFunc.SetDefaultReturn(fmt.Errorf("db failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("failed"))

	indexer := &Indexer{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewIndexerMetrics(metrics.TestRegisterer),
	}

	_, err := indexer.dequeueAndProcess(context.Background())
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
	index := db.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockDB := dbmocks.NewMockDB()
	mockProcessor := NewMockProcessor()
	mockDB.DequeueIndexFunc.SetDefaultReturn(index, mockDB, true, nil)
	mockDB.DoneFunc.SetDefaultReturn(fmt.Errorf("db failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("failed"))

	indexer := &Indexer{
		db:        mockDB,
		processor: mockProcessor,
		metrics:   NewIndexerMetrics(metrics.TestRegisterer),
	}

	_, err := indexer.dequeueAndProcess(context.Background())
	if err == nil || !strings.Contains(err.Error(), "db failure") {
		t.Errorf("unexpected error to Done. want=%q have=%q", "db failure", err)
	}
}

//
//

func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}
