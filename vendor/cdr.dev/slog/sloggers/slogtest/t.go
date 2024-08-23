// Package slogtest contains the slogger for use
// with Go's testing package.
//
// If imported, then all logs that go through the stdlib's
// default logger will go through slog.
package slogtest // import "cdr.dev/slog/sloggers/slogtest"

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"cdr.dev/slog"
	"cdr.dev/slog/internal/entryhuman"
	"cdr.dev/slog/sloggers/sloghuman"
)

// Ensure all stdlib logs go through slog.
func init() {
	l := slog.Make(sloghuman.Sink(os.Stderr))
	log.SetOutput(slog.Stdlib(context.Background(), l, slog.LevelInfo).Writer())
}

// Options represents the options for the logger returned
// by Make.
type Options struct {
	// IgnoreErrors causes the test logger to not fatal the test
	// on Fatal and not error the test on Error or Critical.
	IgnoreErrors bool
	// SkipCleanup skips adding a t.Cleanup call that prevents the logger from
	// logging after a test has exited. This is necessary because race
	// conditions exist when t.Log is called concurrently of a test exiting. Set
	// to true if you don't need this behavior.
	SkipCleanup bool
}

// Make creates a Logger that writes logs to tb in a human readable format.
func Make(tb testing.TB, opts *Options) slog.Logger {
	if opts == nil {
		opts = &Options{}
	}

	sink := &testSink{
		tb:   tb,
		opts: opts,
	}
	if !opts.SkipCleanup {
		tb.Cleanup(func() {
			sink.mu.Lock()
			defer sink.mu.Unlock()
			sink.testDone = true
		})
	}

	return slog.Make(sink)
}

type testSink struct {
	tb       testing.TB
	opts     *Options
	mu       sync.RWMutex
	testDone bool
}

func (ts *testSink) LogEntry(ctx context.Context, ent slog.SinkEntry) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// Don't log after the test this sink was created in has finished.
	if ts.testDone {
		return
	}

	// The testing package logs to stdout and not stderr.
	s := entryhuman.Fmt(os.Stdout, ent)

	switch ent.Level {
	case slog.LevelDebug, slog.LevelInfo, slog.LevelWarn:
		ts.tb.Log(s)
	case slog.LevelError, slog.LevelCritical:
		if ts.opts.IgnoreErrors {
			ts.tb.Log(s)
		} else {
			ts.tb.Error(s)
		}
	case slog.LevelFatal:
		ts.tb.Fatal(s)
	}
}

func (ts *testSink) Sync() {}

var ctx = context.Background()

func l(t testing.TB) slog.Logger {
	return Make(t, &Options{SkipCleanup: true})
}

// Debug logs the given msg and fields to t via t.Log at the debug level.
func Debug(t testing.TB, msg string, fields ...slog.Field) {
	slog.Helper()
	l(t).Debug(ctx, msg, fields...)
}

// Info logs the given msg and fields to t via t.Log at the info level.
func Info(t testing.TB, msg string, fields ...slog.Field) {
	slog.Helper()
	l(t).Info(ctx, msg, fields...)
}

// Error logs the given msg and fields to t via t.Error at the error level.
func Error(t testing.TB, msg string, fields ...slog.Field) {
	slog.Helper()
	l(t).Error(ctx, msg, fields...)
}

// Fatal logs the given msg and fields to t via t.Fatal at the fatal level.
func Fatal(t testing.TB, msg string, fields ...slog.Field) {
	slog.Helper()
	l(t).Fatal(ctx, msg, fields...)
}
