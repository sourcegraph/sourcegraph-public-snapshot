package cache

import (
	"context"
	"sync/atomic"
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

	if val := atomic.LoadUint32(openerCalls); val != 1 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 1, val)
	}
	if val := atomic.LoadUint32(handlerCalls); val != 10 {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", 10, val)
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

	if val := atomic.LoadUint32(openerCalls); val != 8 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 8, val)
	}
	if val := atomic.LoadUint32(handlerCalls); val != uint32(len(keys)) {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", len(keys), val)
	}
}

func TestCacheInitializationTimeout(t *testing.T) {
	openerCalls := uint32(0)
	sync := make(chan struct{})
	wait := make(chan struct{})
	defer close(wait)

	opener := func(filename string) (persistence.Reader, error) {
		close(sync)
		atomic.AddUint32(&openerCalls, 1)
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

	if val := atomic.LoadUint32(&openerCalls); val != 1 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 1, val)
	}
	if val := atomic.LoadUint32(handlerCalls); val != 0 {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", 0, val)
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

	if val := atomic.LoadUint32(openerCalls); val != 1 {
		t.Errorf("unexpected number of opener calls. want=%d have=%d", 1, val)
	}
	if val := atomic.LoadUint32(handlerCalls); val != 1 {
		t.Errorf("unexpected number of handler calls. want=%d have=%d", 1, val)
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

	if val := atomic.LoadUint32(closeCalls); val != 0 {
		t.Errorf("unexpected number of close calls. want=%d have=%d", 0, val)
	}

	close(wait)

	for i := 0; i < 10; i++ {
		if atomic.LoadUint32(closeCalls) == uint32(len(keys)-1) {
			return
		}

		// Wait for goroutines to finish up
		time.Sleep(time.Millisecond * 10)
	}

	t.Errorf("unexpected number of close calls. want=%d have=%d", len(keys)-1, atomic.LoadUint32(closeCalls))
}

func testOpener() (ReaderOpener, *uint32) {
	openerCalls := uint32(0)
	opener := func(filename string) (persistence.Reader, error) {
		atomic.AddUint32(&openerCalls, 1)
		return persistencemocks.NewMockReader(), nil
	}

	return opener, &openerCalls
}

func testBlockingOpener(ch <-chan struct{}) (ReaderOpener, *uint32) {
	openerCalls := uint32(0)
	opener := func(filename string) (persistence.Reader, error) {
		atomic.AddUint32(&openerCalls, 1)
		<-ch
		return persistencemocks.NewMockReader(), nil
	}

	return opener, &openerCalls
}

func testClosingOpener() (ReaderOpener, *uint32, *uint32) {
	closeCalls := uint32(0)
	mock := persistencemocks.NewMockReader()
	mock.CloseFunc.SetDefaultHook(func() error {
		atomic.AddUint32(&closeCalls, 1)
		return nil
	})

	openerCalls := uint32(0)
	opener := func(filename string) (persistence.Reader, error) {
		atomic.AddUint32(&openerCalls, 1)
		return mock, nil
	}

	return opener, &openerCalls, &closeCalls
}

func testHandler() (CacheHandler, *uint32) {
	handlerCalls := uint32(0)
	handler := func(r persistence.Reader) error {
		atomic.AddUint32(&handlerCalls, 1)
		return nil
	}

	return handler, &handlerCalls
}

func testBlockingHandler(ch <-chan struct{}) (CacheHandler, *uint32) {
	handlerCalls := uint32(0)
	handler := func(r persistence.Reader) error {
		atomic.AddUint32(&handlerCalls, 1)
		<-ch
		return nil
	}

	return handler, &handlerCalls
}
