package search

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type eventStreamWriter struct {
	w     io.Writer
	enc   *json.Encoder
	flush func()
}

func newEventStreamWriter(w http.ResponseWriter) (*eventStreamWriter, error) {
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

	return &eventStreamWriter{
		w:     w,
		enc:   json.NewEncoder(w),
		flush: flusher.Flush,
	}, nil
}

func (e *eventStreamWriter) Event(event string, data interface{}) error {
	if event != "" {
		// event: $event\n
		if _, err := e.w.Write([]byte("event: ")); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte(event)); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte("\n")); err != nil {
			return err
		}
	}

	// data: json(data)\n\n
	if _, err := e.w.Write([]byte("data: ")); err != nil {
		return err
	}
	if err := e.enc.Encode(data); err != nil {
		return err
	}
	// Encode writes a newline, so only need to write one newline.
	if _, err := e.w.Write([]byte("\n")); err != nil {
		return err
	}

	e.flush()

	return nil
}
