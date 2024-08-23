package core

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const defaultStreamDelimiter = '\n'

// Streamer calls APIs and streams responses using a *Stream.
type Streamer[T any] struct {
	client  HTTPClient
	retrier *Retrier
}

// NewStreamer returns a new *Streamer backed by the given caller's HTTP client.
func NewStreamer[T any](caller *Caller) *Streamer[T] {
	return &Streamer[T]{
		client:  caller.client,
		retrier: caller.retrier,
	}
}

// StreamParams represents the parameters used to issue an API streaming call.
type StreamParams struct {
	URL          string
	Method       string
	Delimiter    string
	MaxAttempts  uint
	Headers      http.Header
	Client       HTTPClient
	Request      interface{}
	ErrorDecoder ErrorDecoder
}

// Stream issues an API streaming call according to the given stream parameters.
func (s *Streamer[T]) Stream(ctx context.Context, params *StreamParams) (*Stream[T], error) {
	req, err := newRequest(ctx, params.URL, params.Method, params.Headers, params.Request)
	if err != nil {
		return nil, err
	}

	// If the call has been cancelled, don't issue the request.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client := s.client
	if params.Client != nil {
		// Use the HTTP client scoped to the request.
		client = params.Client
	}

	var retryOptions []RetryOption
	if params.MaxAttempts > 0 {
		retryOptions = append(retryOptions, WithMaxAttempts(params.MaxAttempts))
	}

	resp, err := s.retrier.Run(
		client.Do,
		req,
		params.ErrorDecoder,
		retryOptions...,
	)
	if err != nil {
		return nil, err
	}

	// Check if the call was cancelled before we return the error
	// associated with the call and/or unmarshal the response data.
	if err := ctx.Err(); err != nil {
		defer resp.Body.Close()
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, decodeError(resp, params.ErrorDecoder)
	}

	var opts []StreamOption
	if params.Delimiter != "" {
		opts = append(opts, WithDelimiter(params.Delimiter))
	}
	return NewStream[T](resp, opts...), nil
}

// Stream represents a stream of messages sent from a server.
type Stream[T any] struct {
	reader streamReader
	closer io.Closer
}

// StreamOption adapts the behavior of the Stream.
type StreamOption func(*streamOptions)

// WithDelimiter overrides the delimiter for the Stream.
//
// By default, the Stream is newline-delimited.
func WithDelimiter(delimiter string) StreamOption {
	return func(opts *streamOptions) {
		opts.delimiter = delimiter
	}
}

// NewStream constructs a new Stream from the given *http.Response.
func NewStream[T any](response *http.Response, opts ...StreamOption) *Stream[T] {
	options := new(streamOptions)
	for _, opt := range opts {
		opt(options)
	}
	return &Stream[T]{
		reader: newStreamReader(response.Body, options.delimiter),
		closer: response.Body,
	}
}

// Recv reads a message from the stream, returning io.EOF when
// all the messages have been read.
func (s Stream[T]) Recv() (T, error) {
	var value T
	bytes, err := s.reader.ReadFromStream()
	if err != nil {
		return value, err
	}
	if err := json.Unmarshal(bytes, &value); err != nil {
		return value, err
	}
	return value, nil
}

// Close closes the Stream.
func (s Stream[T]) Close() error {
	return s.closer.Close()
}

// streamReader reads data from a stream.
type streamReader interface {
	ReadFromStream() ([]byte, error)
}

// newStreamReader returns a new streamReader based on the given
// delimiter.
//
// By default, the streamReader uses a simple a *bufio.Reader
// which splits on newlines, and otherwise use a *bufio.Scanner to
// split on custom delimiters.
func newStreamReader(reader io.Reader, delimiter string) streamReader {
	if len(delimiter) > 0 {
		return newScannerStreamReader(reader, delimiter)
	}
	return newBufferStreamReader(reader)
}

// bufferStreamReader reads data from a *bufio.Reader, which splits
// on newlines.
type bufferStreamReader struct {
	reader *bufio.Reader
}

func newBufferStreamReader(reader io.Reader) *bufferStreamReader {
	return &bufferStreamReader{
		reader: bufio.NewReader(reader),
	}
}

func (b *bufferStreamReader) ReadFromStream() ([]byte, error) {
	return b.reader.ReadBytes(defaultStreamDelimiter)
}

// scannerStreamReader reads data from a *bufio.Scanner, which allows for
// configurable delimiters.
type scannerStreamReader struct {
	scanner *bufio.Scanner
}

func newScannerStreamReader(reader io.Reader, delimiter string) *scannerStreamReader {
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), delimiter); i >= 0 {
			return i + len(delimiter), data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	return &scannerStreamReader{
		scanner: scanner,
	}
}

func (b *scannerStreamReader) ReadFromStream() ([]byte, error) {
	if b.scanner.Scan() {
		return b.scanner.Bytes(), nil
	}
	if err := b.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

type streamOptions struct {
	delimiter string
}
