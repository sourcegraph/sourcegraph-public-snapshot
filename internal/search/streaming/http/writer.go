package http

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type WriterStat struct {
	Event    string
	Bytes    int
	Duration time.Duration
	Error    error
}

type Writer struct {
	w     io.Writer
	flush func()

	StatHook func(WriterStat)
}

// NewWriter creates a text/event-stream writer that sets a bunch of appropriate
// headers for this use case. Note that once used, users should only interact
// with *Writer directly - for example, using (http.ResponseWriter).WriteHeader
// after NewWriter is used on it is invalid, raising internal errors in net/http:
//
//	http: WriteHeader called with both Transfer-Encoding of "chunked" and a Content-Length of ...
//
// In the WriteHeader case, it will also cause all further calls to (*Writer).Event
// and friends to return an error as well.
func NewWriter(w http.ResponseWriter) (*Writer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("http flushing not supported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// This informs nginx to not buffer. With buffering search responses will
	// be delayed until buffers get full, leading to worst case latency of the
	// full time a search takes to complete.
	w.Header().Set("X-Accel-Buffering", "no")

	return &Writer{
		w:     w,
		flush: flusher.Flush,
	}, nil
}

// Event writes event with data json marshalled.
func (e *Writer) Event(event string, data any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return e.EventBytes(event, encoded)
}

// EventBytes writes dataLine as an event. dataLine is not allowed to contain
// a newline.
func (e *Writer) EventBytes(event string, dataLine []byte) (err error) {
	if payloadSize := 16 /* event: \ndata: \n\n */ + len(event) + len(dataLine); payloadSize > maxPayloadSize {
		return errors.Errorf("payload size %d is greater than max payload size %d", payloadSize, maxPayloadSize)
	}

	// write is a helper to avoid error handling. Additionally it counts the
	// number of bytes written.
	start := time.Now()
	bytes := 0
	write := func(b []byte) {
		if err != nil {
			return
		}
		var n int
		n, err = e.w.Write(b)
		bytes += n
	}

	defer func() {
		if hook := e.StatHook; hook != nil {
			hook(WriterStat{
				Event:    event,
				Bytes:    bytes,
				Duration: time.Since(start),
				Error:    err,
			})
		}
	}()

	if event != "" {
		// event: $event\n
		write([]byte("event: "))
		write([]byte(event))
		write([]byte("\n"))
	}

	// data: json($data)\n\n
	write([]byte("data: "))
	write(dataLine)
	write([]byte("\n\n"))

	e.flush()

	return err
}
