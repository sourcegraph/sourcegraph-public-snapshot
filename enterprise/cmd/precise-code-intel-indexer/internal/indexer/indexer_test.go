package indexer

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
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
	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DequeueIndexFunc.SetDefaultReturn(store.Index{}, nil, false, nil)

	indexer := &Indexer{
		store:     mockStore,
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
	index := store.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DequeueIndexFunc.SetDefaultReturn(index, mockStore, true, nil)

	indexer := &Indexer{
		store:     mockStore,
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
	if len(mockStore.MarkIndexCompleteFunc.History()) != 1 {
		t.Errorf("expected call to MarkIndexComplete")
	}
	if len(mockStore.MarkIndexErroredFunc.History()) != 0 {
		t.Errorf("unexpected call to MarkIndexErrored")
	}
	if len(mockStore.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockStore.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error to Done: %s", doneErr)
	}
}

func TestDequeueAndProcessProcessFailure(t *testing.T) {
	index := store.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DequeueIndexFunc.SetDefaultReturn(index, mockStore, true, nil)
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("process failure"))

	indexer := &Indexer{
		store:     mockStore,
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
	if len(mockStore.MarkIndexCompleteFunc.History()) != 0 {
		t.Errorf("unexpected call to MarkIndexComplete")
	}
	if len(mockStore.MarkIndexErroredFunc.History()) != 1 {
		t.Errorf("expected call to MarkIndexErrored")
	} else if errText := mockStore.MarkIndexErroredFunc.History()[0].Arg2; errText != "process failure" {
		t.Errorf("unexpected failure text. want=%q have=%q", "process failure", errText)
	}
	if len(mockStore.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockStore.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error to Done: %s", doneErr)
	}
}

func TestDequeueAndProcessMarkIndexErrorFailure(t *testing.T) {
	index := store.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockStore := storemocks.NewMockStore()
	mockStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockProcessor := NewMockProcessor()
	mockStore.DequeueIndexFunc.SetDefaultReturn(index, mockStore, true, nil)
	mockStore.MarkIndexErroredFunc.SetDefaultReturn(fmt.Errorf("store failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("failed"))

	indexer := &Indexer{
		store:     mockStore,
		processor: mockProcessor,
		metrics:   NewIndexerMetrics(metrics.TestRegisterer),
	}

	_, err := indexer.dequeueAndProcess(context.Background())
	if err == nil || !strings.Contains(err.Error(), "store failure") {
		t.Errorf("unexpected error to Done. want=%q have=%q", "store failure", err)
	}
	if len(mockStore.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockStore.DoneFunc.History()[0].Arg0; doneErr != nil && !strings.Contains(doneErr.Error(), "store failure") {
		t.Errorf("unexpected error to Done. want=%q have=%q", "store failure", doneErr)
	}
}

func TestDequeueAndProcessDoneFailure(t *testing.T) {
	index := store.Index{
		ID:           42,
		Commit:       makeCommit(1),
		RepositoryID: 50,
	}

	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DequeueIndexFunc.SetDefaultReturn(index, mockStore, true, nil)
	mockStore.DoneFunc.SetDefaultReturn(fmt.Errorf("store failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(fmt.Errorf("failed"))

	indexer := &Indexer{
		store:     mockStore,
		processor: mockProcessor,
		metrics:   NewIndexerMetrics(metrics.TestRegisterer),
	}

	_, err := indexer.dequeueAndProcess(context.Background())
	if err == nil || !strings.Contains(err.Error(), "store failure") {
		t.Errorf("unexpected error to Done. want=%q have=%q", "store failure", err)
	}
}

//
//

func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}
