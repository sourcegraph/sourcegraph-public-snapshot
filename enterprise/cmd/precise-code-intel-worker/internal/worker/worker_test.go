package worker

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

func TestDequeueAndProcessNoUpload(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DequeueFunc.SetDefaultReturn(store.Upload{}, nil, false, nil)

	worker := newWorker(
		mockStore,
		mockProcessor,
		time.Minute,
		1,
		1000,
		metrics.NewWorkerMetrics(&observation.TestContext),
	)

	dequeued, err := worker.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing upload: %s", err)
	}
	if dequeued {
		t.Errorf("unexpected upload dequeued")
	}
}

func TestDequeueAndProcessSuccess(t *testing.T) {
	upload := store.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DequeueFunc.SetDefaultReturn(upload, mockStore, true, nil)

	worker := newWorker(
		mockStore,
		mockProcessor,
		time.Minute,
		1,
		1000,
		metrics.NewWorkerMetrics(&observation.TestContext),
	)

	dequeued, err := worker.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing upload: %s", err)
	}
	if !dequeued {
		t.Errorf("expected upload dequeue")
	}

	// Wait for processor to exit
	<-worker.semaphore

	if len(mockStore.MarkErroredFunc.History()) != 0 {
		t.Errorf("unexpected call to MarkErrored")
	}
	if len(mockStore.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockStore.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error passed to done: %s", doneErr)
	}
}

func TestDequeueAndProcessProcessFailure(t *testing.T) {
	upload := store.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DequeueFunc.SetDefaultReturn(upload, mockStore, true, nil)
	mockProcessor.ProcessFunc.SetDefaultReturn(false, fmt.Errorf("process failure"))

	worker := newWorker(
		mockStore,
		mockProcessor,
		time.Minute,
		1,
		1000,
		metrics.NewWorkerMetrics(&observation.TestContext),
	)

	dequeued, err := worker.dequeueAndProcess(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing and processing upload: %s", err)
	}
	if !dequeued {
		t.Errorf("expected upload dequeue")
	}

	// Wait for processor to exit
	<-worker.semaphore

	if len(mockStore.MarkErroredFunc.History()) != 1 {
		t.Errorf("expected call to MarkErrored")
	} else if errText := mockStore.MarkErroredFunc.History()[0].Arg2; errText != "process failure" {
		t.Errorf("unexpected failure text. want=%q have=%q", "process failure", errText)
	}
	if len(mockStore.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockStore.DoneFunc.History()[0].Arg0; doneErr != nil {
		t.Errorf("unexpected error passed to done: %s", doneErr)
	}
}

func TestHandleMarkErrorFailure(t *testing.T) {
	upload := store.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockStore := storemocks.NewMockStore()
	mockStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockProcessor := NewMockProcessor()
	mockStore.MarkErroredFunc.SetDefaultReturn(fmt.Errorf("store failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(false, fmt.Errorf("failed"))

	worker := newWorker(
		mockStore,
		mockProcessor,
		time.Minute,
		1,
		1000,
		metrics.NewWorkerMetrics(&observation.TestContext),
	)

	if err := worker.handle(context.Background(), mockStore, upload); err == nil || !strings.Contains(err.Error(), "store failure") {
		t.Errorf("unexpected error passed to done. want=%q have=%q", "store failure", err)
	}
	if len(mockStore.DoneFunc.History()) != 1 {
		t.Errorf("expected call to Done")
	} else if doneErr := mockStore.DoneFunc.History()[0].Arg0; doneErr != nil && !strings.Contains(doneErr.Error(), "store failure") {
		t.Errorf("unexpected error passed to done. want=%q have=%q", "store failure", doneErr)
	}
}

func TestHandleDoneFailure(t *testing.T) {
	upload := store.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockStore := storemocks.NewMockStore()
	mockProcessor := NewMockProcessor()
	mockStore.DoneFunc.SetDefaultReturn(fmt.Errorf("store failure"))
	mockProcessor.ProcessFunc.SetDefaultReturn(false, fmt.Errorf("failed"))

	worker := newWorker(
		mockStore,
		mockProcessor,
		time.Minute,
		1,
		1000,
		metrics.NewWorkerMetrics(&observation.TestContext),
	)

	if err := worker.handle(context.Background(), mockStore, upload); err == nil || !strings.Contains(err.Error(), "store failure") {
		t.Errorf("unexpected error passed to done. want=%q have=%q", "store failure", err)
	}
}
