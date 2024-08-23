package run

import (
	"bytes"
	"context"
	"io"

	"go.bobheadxi.dev/streamline/pipeline"
)

// LineMap allows modifications of individual lines from Output and enables callbacks
// that operate on lines from Output. Bytes written to dst are collected and passed to
// subsequent LineMaps before being written to output aggregation, e.g. Output.Stream().
//
// The return value mirrors the signature of (Writer).Write(), and should be used to
// indicate what was written to dst.
//
// Errors interrupt line processing and are returned if and only if the command itself
// did not exit with an error.
type LineMap func(ctx context.Context, line []byte, dst io.Writer) (int, error)

// MapJQ creates a LineMap that executes a JQ query against each line and replaces the
// output with the result.
//
// Refer to https://github.com/itchyny/gojq for the specifics of supported syntax.
func MapJQ(query string) (LineMap, error) {
	jqCode, err := buildJQ(query)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, line []byte, dst io.Writer) (int, error) {
		b, err := execJQBytes(ctx, jqCode, line)
		if err != nil {
			return 0, err
		}
		return dst.Write(b)
	}, nil
}

type lineMapPipelineAdapter struct {
	ctx     context.Context
	buffer  *bytes.Buffer
	lineMap LineMap
}

var _ pipeline.Pipeline = &lineMapPipelineAdapter{}

func (l *lineMapPipelineAdapter) Inactive() bool { return l == nil || l.lineMap == nil }

func (l *lineMapPipelineAdapter) ProcessLine(line []byte) ([]byte, error) {
	// Use a shared buffer when applying this LineMap - it gets reset on each
	// line, and lines are processed synchronously.
	l.buffer.Reset()

	buf := tracedBuffer{Buffer: l.buffer}
	_, err := l.lineMap(l.ctx, line, &buf)
	if !buf.writeCalled || err != nil {
		return nil, err // omit the line or return the error
	}
	return buf.Bytes(), nil
}

type tracedBuffer struct {
	// writeCalled indicates that Write was called at all, even with empty input.
	writeCalled bool

	*bytes.Buffer
}

func (t *tracedBuffer) Write(b []byte) (int, error) {
	t.writeCalled = true
	return t.Buffer.Write(b)
}
