package streamline

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"go.bobheadxi.dev/streamline/pipeline"
)

// LineReader is a reader that implements the ability to read up to a line delimiter.
// It is used for internal assertion only, to determine if io.Readers provided to
// streamline.New already implement the desired functionality - it is exported for
// reference, and should not be depended upon, since the interface may change in the
// future.
type LineReader interface {
	// ReadSlice should have the behaviour of (*bufio.Reader).ReadSlice, returning bytes
	// up to and including the delimiter.
	ReadSlice(delim byte) ([]byte, error)
}

// Stream enables live, line-by-line manipulation and handling of data through
// (*Stream).WithPipeline(...) and Stream's various aggregation methods. Stream also
// supports standard library features like io.Copy and io.ReadAll by implementing
// io.Reader and io.WriterTo.
//
// Stream's aggregation methods ((*Stream).Stream(...), (*Stream).Lines(...), etc) may
// only be used once. (*Stream).Read(...) may need to be called multiple times to consume
// all data but should not be used in conjunction with other methods - i.e. once Read is
// used, Stream behaves like an io.Reader.
//
// In general, (*Stream).Stream(...) and (*Stream).StreamBytes(...) has similar
// performance to ReadString and ReadSlice respectively from bufio.Reader, even with
// Pipelines configured (though certain Pipeline implementations may have higher overhead).
// Using bufio.Scanner from Text() may be more performant than (*Stream).Stream(...) if
// the line string is not used, but otherwise Stream generally performs similarly to its
// bufio.Scanner equivalents.
//
// Where Stream's overhead becomes more noticeable is in Stream's io.Reader implementation.
// Stream reads entire lines for Pipelines to process before populating data in
// (*Stream).Read(...), and retains a buffer for unread data. It is efficient for cases
// where you want to process data with Pipelines before providing it to a consumer, but
// if you are not using Pipelines or Stream's line-by-line aggregation methods, it may be
// better to provide your data directly to readers instead of wrapping it in Stream.
type Stream struct {
	// reader carries the input data and the current read state.
	reader LineReader

	// pipeline, if active, must be used to pre-process lines.
	pipeline pipeline.MultiPipeline

	// readBuffer is set by incremental consumers like Read to store unread data
	// from the reader.
	readBuffer *bytes.Buffer

	// lineSeparator is used as the read delimiter.
	lineSeparator byte
}

// New creates a Stream that consumes, processes, and emits data from the input. If the
// input also implements LineReader, then it will use the input directly - otherwise, it
// will wrap the input in a bufio.Reader.
func New(input io.Reader) *Stream {
	var reader LineReader
	if lr, ok := input.(LineReader); ok {
		reader = lr
	} else {
		reader = bufio.NewReader(input)
	}
	return &Stream{
		reader:        reader,
		lineSeparator: '\n',
	}
}

// WithPipeline configures this Stream to process the input data with the given Pipeline
// in all output methods ((*Stream).Stream(...), (*Stream).Lines(...), io.Copy, etc.).
//
// If one or more Pipelines are already configured on this Stream, the given Pipeline
// is applied sequentially after the preconfigured pipelines.
func (s *Stream) WithPipeline(p pipeline.Pipeline) *Stream {
	s.pipeline = append(s.pipeline, p)
	return s
}

// WithLineSeparator configures a custom line separator for this stream. The default is '\n'.
func (s *Stream) WithLineSeparator(separator byte) *Stream {
	s.lineSeparator = separator
	return s
}

// Stream passes lines read from the input to the handler as it processes them. It is
// intended for simple use cases - to be able to provide errors from the line handler, use
// StreamBytes instead.
//
// This method will block until the input returns an error. Unless the error is io.EOF,
// it will also propagate the error.
func (s *Stream) Stream(dst func(line string)) error {
	return s.StreamBytes(func(line []byte) error {
		dst(string(line))
		return nil
	})
}

// StreamBytes passes lines read from the input to the handler as it processes them, and
// allows the handler to return an error.
//
// This method will block until the input returns an error. Unless the error is io.EOF,
// it will also propagate the error.
//
// Handlers must not retain line.
func (s *Stream) StreamBytes(dst func(line []byte) error) error {
	for {
		_, err := s.readLine(dst)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

// Lines collects all processed output as a slice of strings.
//
// This method will block until the input returns an error. Unless the error is io.EOF,
// it will also propagate the error.
func (s *Stream) Lines() ([]string, error) {
	lines := make([]string, 0, 10)
	return lines, s.Stream(func(line string) { lines = append(lines, line) })
}

// String collects all processed output as a string.
//
// This method will block until the input returns an error. Unless the error is io.EOF,
// it will also propagate the error.
func (s *Stream) String() (string, error) {
	var sb strings.Builder
	_, err := s.WriteTo(&sb)
	data := strings.TrimSuffix(sb.String(), string(s.lineSeparator))
	return data, err
}

// Bytes collects all processed output as a bytes slice.
//
// This method will block until the input returns an error. Unless the error is io.EOF,
// it will also propagate the error.
func (s *Stream) Bytes() ([]byte, error) {
	var b bytes.Buffer
	_, err := s.WriteTo(&b)
	data := bytes.TrimSuffix(b.Bytes(), []byte{s.lineSeparator})
	return data, err
}

var _ io.WriterTo = (*Stream)(nil)

// WriteTo writes processed data to dst. It allows Stream to effectively implement io.Copy
// handling. An additional new line will be added to the last line if there isn't already
// one.
func (s *Stream) WriteTo(dst io.Writer) (int64, error) {
	var totalWritten int64
	return totalWritten, s.StreamBytes(func(line []byte) error {
		n, err := dst.Write(append(line, s.lineSeparator))
		totalWritten += int64(n)
		return err
	})
}

var _ io.Reader = (*Stream)(nil)

// Read populates p with processed data. It allows Stream to effectively be compatible
// with anything that accepts an io.Reader. An additional new line will be added to the
// last line if there isn't already one.
//
// The implementation is designed to read entire lines of data and processing them with
// configured Pipelines before populating p, which adds some overhead - see the Stream
// docstrings for more details.
func (s *Stream) Read(p []byte) (int, error) {
	if s.readBuffer == nil {
		s.readBuffer = &bytes.Buffer{}
	}

	// If we have unread data, read it. Only return if and only if this fills p, otherwise
	// continue with line read.
	var written int
	if s.readBuffer.Len() > 0 {
		// Buffer read can only error with io.EOF, so we just discard it.
		written, _ = s.readBuffer.Read(p)
		if written == len(p) {
			return written, nil
		}
	}

	// Unread data has been read - we can reset the buffer now.
	s.readBuffer.Reset()

	// Next, written some lines into the buffer, keeping track of how much we have written
	// into p.
	for {
		var currentLine []byte
		skipped, err := s.readLine(func(next []byte) error {
			currentLine = append(next, s.lineSeparator)
			return nil
		})

		// If this was skipped line (different from an empty line), keep reading for more
		// data.
		if skipped && err == nil {
			continue
		}

		// Copy line byte by byte into b based on our current position, how much data is
		// in the line, and how much we have already written.
		for read := 0; read < len(currentLine); read++ {
			p[written] = currentLine[read]
			written++

			// We have filled up p, we are done.
			if written == len(p) {
				// If we weren't done reading the current line, write the
				// remainder into readBuffer - the next read will pick it up.
				if read < len(currentLine) {
					// Buffer writes will never error.
					_, _ = s.readBuffer.Write(currentLine[read+1:])
				}

				return written, err
			}
		}

		// Now that the read is complete, if an error has occurred (typically io.EOF) then
		// we are done.
		if err != nil {
			return written, err
		}

		// We were able to copy everything from currentLine into p, and p is not yet full,
		// and our data is not yet exhausted - continue.
	}
}

// readLine consumes a single line in the stream. The error returned, in order of
// precedence, is one of:
//
//   - processing error
//   - handler error
//   - read error
//
// The read error in particular may be io.EOF, which the caller should handle on a
// case-by-case basis.
func (s *Stream) readLine(handle func(line []byte) error) (skipped bool, err error) {
	var line []byte
	var readErr error
	for {
		// Each ReadSlice doesn't necessarily give us the entire line - it might
		// give us only part of it we get bufio.ErrBufferFull, so we keep reading
		// until we get a different result.
		data, err := s.reader.ReadSlice(s.lineSeparator)
		line = append(line, data...)
		if err == bufio.ErrBufferFull {
			continue
		}
		// Otherwise, err is the final result and we are done reading.
		readErr = err
		break
	}

	// If we got no data and encountered a read error, we can return immediately.
	// Generally, a non-nil readErr is an io.EOF if len(line) > 0, so after this point
	// we prefer to return other errors.
	if len(line) == 0 && readErr != nil {
		return true, readErr
	}

	// If the line ends with a newline, trim it before handling it - callers should add
	// it back as necessary.
	if line[len(line)-1] == s.lineSeparator {
		line = line[:len(line)-1]
	}

	// Run the line through any configured pipelines. Processing errors take precedence
	// over readErr still.
	if len(s.pipeline) > 0 {
		var processErr error
		if line, processErr = s.pipeline.ProcessLine(line); processErr != nil {
			return false, processErr
		}

		// Pipelines only return nil lines if the line should be skipped entirely.
		if line == nil {
			return true, nil
		}

		// Check if the Pipeline returned a multi-line line.
		if bytes.ContainsRune(line, rune(s.lineSeparator)) {
			// We do some custom handling here by giving the processed lines separately to
			// the handler, returning the handler error if we receive one - it continues
			// to take precedence over readErr.
			for _, subLine := range bytes.Split(line, []byte{s.lineSeparator}) {
				if dstErr := handle(subLine); dstErr != nil {
					return false, dstErr
				}
			}
			// We are done.
			return false, readErr
		}
	}

	// We give the processed line to the handler, returning the handler error if we
	// receive one - it continues to take precedence over readErr.
	if dstErr := handle(line); dstErr != nil {
		return false, dstErr
	}

	// Finally, if no other errors occur, we can return readErr.
	return false, readErr
}
