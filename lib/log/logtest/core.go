package logtest

import (
	"bytes"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"go.uber.org/zap/zapcore"
)

// newTestingCore creates a core with the same encoder as our usual logging, but writes to
// testing.TB instead.
func newTestingCore(t testing.TB, level zapcore.LevelEnabler, failOnErrorLogs bool) zapcore.Core {
	return zapcore.NewCore(
		encoders.BuildEncoder(encoders.OutputConsole, true),
		&testingWriter{
			t:          t,
			markFailed: failOnErrorLogs,
		},
		level,
	)
}

// testingWriter is a WriteSyncer that writes to the given testing.TB.
//
// Adapted from https://sourcegraph.com/github.com/uber-go/zap/-/blob/zaptest/logger.go
//
// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
type testingWriter struct {
	t testing.TB

	// If true, the test will be marked as failed if this testingWriter is
	// ever used.
	markFailed bool
}

func newTestingWriter(t testing.TB) testingWriter {
	return testingWriter{t: t}
}

// WithMarkFailed returns a copy of this testingWriter with markFailed set to
// the provided value.
func (w testingWriter) WithMarkFailed(v bool) testingWriter {
	w.markFailed = v
	return w
}

func (w testingWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	// Strip trailing newline because t.Log always adds one.
	p = bytes.TrimRight(p, "\n")

	// Note: t.Log is safe for concurrent use.
	w.t.Logf("%s", p)
	if w.markFailed {
		w.t.Fail()
	}

	return n, nil
}

func (w testingWriter) Sync() error {
	return nil
}
