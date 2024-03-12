// Package streamio contains wrappers intended for turning gRPC streams
// that send/receive messages with a []byte field into io.Writers and
// io.Readers.
//
// This file is largely copied from the gitaly project, which is licensed
// under the MIT license. A copy of that license text can be found at
// https://mit-license.org/. The code this file was based off can be found
// at https://gitlab.com/gitlab-org/gitaly/-/blob/v1.87.0/streamio/stream.go
package streamio

import (
	"io"
	"sync"
)

// NewReader turns receiver into an io.Reader. Errors from the receiver
// function are passed on unmodified. This means receiver should emit
// io.EOF when done.
func NewReader(receiver func() ([]byte, error)) io.Reader {
	return &receiveReader{receiver: receiver}
}

type receiveReader struct {
	receiver func() ([]byte, error)
	data     []byte
	err      error
}

func (rr *receiveReader) Read(p []byte) (int, error) {
	if len(rr.data) == 0 && rr.err == nil {
		rr.data, rr.err = rr.receiver()
	}

	n := copy(p, rr.data)
	rr.data = rr.data[n:]

	// We want to return any potential error only in case we have no
	// buffered data left. Otherwise, it can happen that we do not relay
	// bytes when the reader returns both data and an error.
	if len(rr.data) == 0 {
		return n, rr.err
	}

	return n, nil
}

// NewWriter turns sender into an io.Writer. The sender callback will
// receive []byte arguments of length at most WriteBufferSize.
func NewWriter(sender func(p []byte) error) io.Writer {
	return &sendWriter{sender: sender}
}

// NewSyncWriter turns sender into an io.Writer. The sender callback will
// receive []byte arguments of length at most WriteBufferSize. All calls to the
// sender will be synchronized via the mutex.
func NewSyncWriter(m *sync.Mutex, sender func(p []byte) error) io.Writer {
	return &sendWriter{
		sender: func(p []byte) error {
			m.Lock()
			defer m.Unlock()

			return sender(p)
		},
	}
}

// WriteBufferSize is the largest []byte that Write() will pass to its
// underlying send function.
var WriteBufferSize = 128 * 1024

type sendWriter struct {
	sender func([]byte) error
}

func (sw *sendWriter) Write(p []byte) (int, error) {
	var sent int

	for len(p) > 0 {
		chunkSize := len(p)
		if chunkSize > WriteBufferSize {
			chunkSize = WriteBufferSize
		}

		if err := sw.sender(p[:chunkSize]); err != nil {
			return sent, err
		}

		sent += chunkSize
		p = p[chunkSize:]
	}

	return sent, nil
}
