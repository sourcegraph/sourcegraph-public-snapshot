package cache

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	persistencemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/mocks"
)

func TestReaderCache(t *testing.T) {
	openerCalls := uint32(0)
	opener := func(filename string) (persistence.Reader, error) {
		atomic.AddUint32(&openerCalls, 1)
		return persistencemocks.NewMockReader(), nil
	}

	ch := make(chan time.Time)
	defer close(ch)

	cache := newReaderCache(time.Minute, ch, opener)

	keys := []string{
		"foo",
		"bar",
		"baz",
	}

	// Create entries
	for _, key := range keys {
		if err := cache.WithReader(context.Background(), key, noopHandler); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}

	if openerCalls != 3 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 3, openerCalls)
	}

	// Re-use entries
	for _, key := range keys {
		if err := cache.WithReader(context.Background(), key, noopHandler); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}

	if openerCalls != 3 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 3, openerCalls)
	}

	var chs []chan struct{}
	for _, key := range keys {
		chs = append(chs, cache.entries[key].drained)
	}

	// Clear all entries
	ch <- time.Now().UTC().Add(time.Minute * 2)

	// Wait for entries to drain
	for _, drained := range chs {
		<-drained
	}

	// Re-create entries
	for _, key := range keys {
		if err := cache.WithReader(context.Background(), key, noopHandler); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}

	if openerCalls != 6 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 6, openerCalls)
	}
}

func TestReaderCacheInitError(t *testing.T) {
	openErr := fmt.Errorf("open error")
	opener := func(filename string) (persistence.Reader, error) {
		tmp := openErr
		openErr = nil
		return persistencemocks.NewMockReader(), tmp
	}

	ch := make(chan time.Time)
	defer close(ch)

	cache := newReaderCache(time.Minute, ch, opener)

	if err := cache.WithReader(context.Background(), "test", noopHandler); err == nil {
		t.Errorf("unexpected nil error")
	} else if err.Error() != "open error" {
		t.Errorf("unexpected error. want=%q have=%q", "open error", err)
	}

	// Ensure we can recreate entry after previous error
	if err := cache.WithReader(context.Background(), "test", noopHandler); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestReaderCacheDraining(t *testing.T) {
	wait := make(chan struct{})
	sync := make(chan struct{})

	opener := func(filename string) (persistence.Reader, error) {
		reader := persistencemocks.NewMockReader()
		reader.CloseFunc.SetDefaultHook(func() error {
			close(sync)
			<-wait
			return nil
		})
		return reader, nil
	}

	ch := make(chan time.Time)
	defer close(ch)

	cache := newReaderCache(time.Minute, ch, opener)

	if err := cache.WithReader(context.Background(), "test", noopHandler); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// Clear all entries, wait until reader begins to close
	ch <- time.Now().UTC().Add(time.Minute * 2)
	<-sync

	// Closed after we re-create a value
	opened := make(chan struct{})

	go func() {
		defer close(opened)

		if err := cache.WithReader(context.Background(), "test", noopHandler); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}()

	select {
	case <-opened:
		// Created a new value while old value drains
		t.Errorf("unexpected value on channel")
	case <-time.After(time.Millisecond * 50):
	}

	// Release old value
	close(wait)

	select {
	case <-opened:
	case <-time.After(time.Millisecond * 50):
		t.Errorf("expected value on channel")
	}
}

func TestReaderCacheContextCanceled(t *testing.T) {
	wait := make(chan struct{})
	opener := func(filename string) (persistence.Reader, error) {
		<-wait
		return persistencemocks.NewMockReader(), nil
	}

	ch := make(chan time.Time)
	defer close(ch)

	cache := newReaderCache(time.Minute, ch, opener)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cache.WithReader(ctx, "test", noopHandler); err != ErrReaderInitializationDeadlineExceeded {
		t.Errorf("unexpected error. want=%q have=%q", ErrReaderInitializationDeadlineExceeded, err)
	}
}

func noopHandler(reader persistence.Reader) error {
	return nil
}
