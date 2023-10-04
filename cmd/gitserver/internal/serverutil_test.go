package internal

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestFlushingResponseWriter(t *testing.T) {
	flush := make(chan struct{})
	fw := &flushingResponseWriter{
		w: httptest.NewRecorder(),
		flusher: flushFunc(func() {
			flush <- struct{}{}
		}),
	}
	done := make(chan struct{})
	go func() {
		fw.periodicFlush()
		close(done)
	}()

	_, _ = fw.Write([]byte("hi"))

	select {
	case <-flush:
		close(flush)
	case <-time.After(5 * time.Second):
		t.Fatal("periodic flush did not happen")
	}

	fw.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("periodic flush goroutine did not close")
	}
}

type flushFunc func()

func (f flushFunc) Flush() {
	f()
}
