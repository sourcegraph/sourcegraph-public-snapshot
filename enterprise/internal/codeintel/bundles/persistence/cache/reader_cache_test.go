package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	persistencemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

func TestReaderCache(t *testing.T) {
	opener := func(filename string) (persistence.Reader, error) {
		reader := persistencemocks.NewMockReader()
		reader.ReadMetaFunc.SetDefaultReturn(types.MetaData{NumResultChunks: 4}, nil)
		return reader, nil
	}

	var meta types.MetaData
	handler := func(reader persistence.Reader) (err error) {
		meta, err = reader.ReadMeta(context.Background())
		return err
	}

	cache := newReaderCache(opener)

	if err := cache.WithReader(context.Background(), "test", handler); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if meta.NumResultChunks != 4 {
		t.Errorf("unexpected reader result. want=%d have=%d", 4, meta.NumResultChunks)
	}
}

func TestReaderCacheError(t *testing.T) {
	opener := func(filename string) (persistence.Reader, error) {
		return persistencemocks.NewMockReader(), nil
	}

	expectedErr := fmt.Errorf("oops")
	handler := func(reader persistence.Reader) error {
		return expectedErr
	}

	cache := newReaderCache(opener)

	if err := cache.WithReader(context.Background(), "test", handler); err != expectedErr {
		t.Fatalf("unexpected error. want=%q have=%q", expectedErr, err)
	}
}

func TestReaderCacheInitError(t *testing.T) {
	expectedErr := fmt.Errorf("oops")
	returnedErr := expectedErr
	opener := func(filename string) (persistence.Reader, error) {
		temp := returnedErr
		returnedErr = nil
		return persistencemocks.NewMockReader(), temp
	}

	handler := func(reader persistence.Reader) error {
		return nil
	}

	cache := newReaderCache(opener)

	if err := cache.WithReader(context.Background(), "test", handler); err != expectedErr {
		t.Fatalf("unexpected error. want=%q have=%q", expectedErr, err)
	}

	// Wait for entry to drain
	waitForCondition(func() bool {
		cache.m.RLock()
		defer cache.m.RUnlock()
		return len(cache.readers) == 0
	})

	// Ensure we can recreate the reader without error
	if err := cache.WithReader(context.Background(), "test", handler); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestReaderCacheConcurrentRequests(t *testing.T) {
	t.Skip("Skipping because there seems to be race condition where the reported number of calls is wrong")

	calls := 0
	wait := make(chan struct{}) // blocks opener

	opener := func(filename string) (persistence.Reader, error) {
		calls++
		<-wait

		reader := persistencemocks.NewMockReader()
		reader.ReadMetaFunc.SetDefaultReturn(types.MetaData{NumResultChunks: 4}, nil)
		return reader, nil
	}

	n := 10
	errs := make(chan error)                // errors from withReader
	sync := make(chan struct{})             // signals goroutine has been scheduled
	values := make(chan types.MetaData, 10) // values from ReadMeta within handler

	handler := func(reader persistence.Reader) error {
		meta, err := reader.ReadMeta(context.Background())
		values <- meta
		return err
	}

	cache := newReaderCache(opener)

	for i := 0; i < n; i++ {
		go func() {
			sync <- struct{}{}
			errs <- cache.WithReader(context.Background(), "test", handler)
		}()
	}

	// Wait until all routines have scheduled
	for i := 0; i < n; i++ {
		<-sync
	}
	close(wait)

	// Read all errors
	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	close(errs)

	// Read all handler results
	for i := 0; i < n; i++ {
		if meta := <-values; meta.NumResultChunks != 4 {
			t.Errorf("unexpected reader result. want=%d have=%d", 4, meta.NumResultChunks)
		}
	}
	close(values)

	if calls != 1 {
		t.Errorf("unexpected number of calls. want=%d have=%d", 1, calls)
	}
}

func TestReaderCacheDisposedEntry(t *testing.T) {
	wait := make(chan struct{}) // blocks close from returning
	sync := make(chan struct{}) // signals close routine has been scheduled

	reader := persistencemocks.NewMockReader()
	reader.CloseFunc.PushHook(func(err error) error {
		close(sync)
		<-wait
		return err
	})

	opener := func(filename string) (persistence.Reader, error) {
		return reader, nil
	}

	handler := func(reader persistence.Reader) error {
		return nil
	}

	cache := newReaderCache(opener)

	if err := cache.WithReader(context.Background(), "test", handler); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	<-sync

	go func() {
		// Make disposed entry available for 25ms
		<-time.After(TestTickDuration)
		close(wait)
	}()

	if err := cache.WithReader(context.Background(), "test", handler); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestReaderCacheDisposedEntryContextCanceled(t *testing.T) {
	sync := make(chan struct{}) // signals close routine has been scheduled

	reader := persistencemocks.NewMockReader()
	reader.CloseFunc.PushHook(func(err error) error {
		close(sync)
		select {} // block forever
	})

	opener := func(filename string) (persistence.Reader, error) {
		return reader, nil
	}

	handler := func(reader persistence.Reader) error {
		return nil
	}

	cache := newReaderCache(opener)

	if err := cache.WithReader(context.Background(), "test", handler); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	<-sync

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cache.WithReader(ctx, "test", handler); err != context.Canceled {
		t.Fatalf("unexpected error. want=%q have=%q", context.Canceled, err)
	}
}

func TestReaderContextCanceled(t *testing.T) {
	wait := make(chan struct{})
	defer close(wait)

	opener := func(filename string) (persistence.Reader, error) {
		<-wait
		return persistencemocks.NewMockReader(), nil
	}

	handler := func(reader persistence.Reader) error {
		return nil
	}

	cache := newReaderCache(opener)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cache.WithReader(ctx, "test", handler); err != context.Canceled {
		t.Fatalf("unexpected error. want=%q have=%q", context.Canceled, err)
	}
}
