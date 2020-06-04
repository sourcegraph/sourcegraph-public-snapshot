package cache

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	persistencemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/mocks"
)

func TestCacheBasic(t *testing.T) {
	opener, openerCalls := testOpener()
	handler, handlerCalls := testHandler()
	cache := New(10, opener)

	for i := 0; i < 10; i++ {
		if err := cache.WithReader(context.Background(), "test", handler); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}

	if *openerCalls != 1 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 1, *openerCalls)
	}
	if *handlerCalls != 10 {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", 10, *handlerCalls)
	}
}

func TestCacheBasicEviction(t *testing.T) {
	opener, openerCalls := testOpener()
	handler, handlerCalls := testHandler()
	cache := New(3, opener)

	keys := []string{
		"foo",  // foo
		"foo",  // foo
		"foo",  // foo
		"bar",  // bar foo
		"baz",  // baz bar foo
		"foo",  // foo baz bar
		"bonk", // bonk foo baz
		"bar",  // bar bonk foo
		"quux", // quux bar bonk
		"foo",  // foo quux bar
		"bar",  // bar foo quux
		"baz",  // baz bar foo
	}

	for _, key := range keys {
		if err := cache.WithReader(context.Background(), key, handler); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}

	if *openerCalls != 8 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 8, *openerCalls)
	}
	if *handlerCalls != len(keys) {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", len(keys), *handlerCalls)
	}
}

func TestCacheInitializationTimeout(t *testing.T) {
	openerCalls := 0
	sync := make(chan struct{})
	wait := make(chan struct{})
	defer close(wait)

	opener := func(filename string) (persistence.Reader, error) {
		close(sync)
		openerCalls++
		<-wait
		return persistencemocks.NewMockReader(), nil
	}

	handler, handlerCalls := testHandler()
	cache := New(1, opener)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sync
		cancel()
	}()

	if err := cache.WithReader(ctx, "test", handler); err != ErrReaderInitializationDeadlineExceeded {
		t.Errorf("unexpected error. want=%q have=%q", ErrReaderInitializationDeadlineExceeded, err)
	}

	if openerCalls != 1 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 1, openerCalls)
	}
	if *handlerCalls != 0 {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", 0, handlerCalls)
	}
}

func TestCacheInitializationTimeoutSecondAttempt(t *testing.T) {
	wait := make(chan struct{})
	opener, openerCalls := testBlockingOpener(wait)
	handler, handlerCalls := testHandler()
	cache := New(1, opener)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cache.WithReader(ctx, "test", handler); err != ErrReaderInitializationDeadlineExceeded {
		t.Errorf("unexpected error. want=%q have=%q", ErrReaderInitializationDeadlineExceeded, err)
	}

	close(wait)

	if err := cache.WithReader(context.Background(), "test", handler); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if *openerCalls != 1 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 1, *openerCalls)
	}
	if *handlerCalls != 1 {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", 1, *handlerCalls)
	}
}

func TestCacheNotClosedWhileHeld(t *testing.T) {
	wait := make(chan struct{})
	opener, _, closeCalls := testClosingOpener()
	handler, _ := testBlockingHandler(wait)
	cache := New(1, opener)

	keys := []string{"foo", "bar", "baz", "bonk", "quux"}

	for _, key := range keys {
		go func(key string) {
			if err := cache.WithReader(context.Background(), key, handler); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		}(key)
	}

	if *closeCalls != 0 {
		t.Errorf("unexpected number of close calls. want=%d have=%d", 0, *closeCalls)
	}

	close(wait)

	for i := 0; i < 10; i++ {
		if *closeCalls == len(keys)-1 {
			return
		}

		// Wait for goroutines to finish up
		time.Sleep(time.Millisecond * 10)
	}

	t.Errorf("unexpected number of close calls. want=%d have=%d", len(keys)-1, *closeCalls)
}

func testOpener() (ReaderOpener, *int) {
	openerCalls := 0
	opener := func(filename string) (persistence.Reader, error) {
		openerCalls++
		return persistencemocks.NewMockReader(), nil
	}

	return opener, &openerCalls
}

func testBlockingOpener(ch <-chan struct{}) (ReaderOpener, *int) {
	openerCalls := 0
	opener := func(filename string) (persistence.Reader, error) {
		openerCalls++
		<-ch
		return persistencemocks.NewMockReader(), nil
	}

	return opener, &openerCalls
}

func testClosingOpener() (ReaderOpener, *int, *int) {
	closeCalls := 0
	mock := persistencemocks.NewMockReader()
	mock.CloseFunc.SetDefaultHook(func() error {
		closeCalls++
		return nil
	})

	openerCalls := 0
	opener := func(filename string) (persistence.Reader, error) {
		openerCalls++
		return mock, nil
	}

	return opener, &openerCalls, &closeCalls
}

func testHandler() (CacheHandler, *int) {
	handlerCalls := 0
	handler := func(r persistence.Reader) error {
		handlerCalls++
		return nil
	}

	return handler, &handlerCalls
}

func testBlockingHandler(ch <-chan struct{}) (CacheHandler, *int) {
	handlerCalls := 0
	handler := func(r persistence.Reader) error {
		handlerCalls++
		<-ch
		return nil
	}

	return handler, &handlerCalls
}
