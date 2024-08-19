// Package syncwriter implements a concurrency safe io.Writer wrapper.
package syncwriter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

// Writer implements a concurrency safe io.Writer wrapper.
type Writer struct {
	mu sync.Mutex
	w  io.Writer

	errorf func(f string, v ...interface{})
}

// New returns a new Writer that writes to w.
func New(w io.Writer) *Writer {
	return &Writer{
		w: w,

		errorf: func(f string, v ...interface{}) {
			println(fmt.Sprintf(f, v...))
		},
	}
}

func (w *Writer) Write(name string, p []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := w.w.Write(p)
	if err != nil {
		w.errorf("%v: failed to write entry: %+v", name, err)
	}
}

type syncer interface {
	Sync() error
}

var _ syncer = &os.File{}

// Sync calls Sync on the underlying writer
// if possible.
func (w *Writer) Sync(sinkName string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	s, ok := w.w.(syncer)
	if !ok {
		return
	}
	err := s.Sync()
	if _, ok := w.w.(*os.File); ok {
		// Opened files do not necessarily support syncing.
		// E.g. stdout and stderr both do not so we need
		// to ignore these errors.
		// See https://github.com/uber-go/zap/issues/370
		// See https://github.com/cdr/slog/pull/43
		if errorsIsAny(err, syscall.EINVAL, syscall.ENOTTY, syscall.EBADF) {
			return
		}
	}

	w.errorf("failed to sync %v: %+v", sinkName, err)
}

func errorsIsAny(err error, errs ...error) bool {
	for _, e := range errs {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
