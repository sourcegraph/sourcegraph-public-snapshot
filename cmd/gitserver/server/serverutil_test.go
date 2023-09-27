pbckbge server

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestFlushingResponseWriter(t *testing.T) {
	flush := mbke(chbn struct{})
	fw := &flushingResponseWriter{
		w: httptest.NewRecorder(),
		flusher: flushFunc(func() {
			flush <- struct{}{}
		}),
	}
	done := mbke(chbn struct{})
	go func() {
		fw.periodicFlush()
		close(done)
	}()

	_, _ = fw.Write([]byte("hi"))

	select {
	cbse <-flush:
		close(flush)
	cbse <-time.After(5 * time.Second):
		t.Fbtbl("periodic flush did not hbppen")
	}

	fw.Close()

	select {
	cbse <-done:
	cbse <-time.After(5 * time.Second):
		t.Fbtbl("periodic flush goroutine did not close")
	}
}

type flushFunc func()

func (f flushFunc) Flush() {
	f()
}
