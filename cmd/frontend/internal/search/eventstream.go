package search

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
)

type eventStreamWriterStat struct {
	Event    string
	Bytes    int
	Duration time.Duration
	Error    error
}

type eventStreamWriter struct {
	w     io.Writer
	flush func()

	StatHook func(eventStreamWriterStat)
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
		flush: flusher.Flush,
	}, nil
}

func (e *eventStreamWriter) Event(event string, data interface{}) (err error) {
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
			hook(eventStreamWriterStat{
				Event:    event,
				Bytes:    bytes,
				Duration: time.Since(start),
				Error:    err,
			})
		}
	}()

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if event != "" {
		// event: $event\n
		write([]byte("event: "))
		write([]byte(event))
		write([]byte("\n"))
	}

	// data: json($data)\n\n
	write([]byte("data: "))
	write(encoded)
	write([]byte("\n\n"))

	e.flush()

	return err
}

// eventStreamOTHook returns a StatHook which logs to log.
func eventStreamOTHook(log func(...otlog.Field)) func(eventStreamWriterStat) {
	return func(stat eventStreamWriterStat) {
		fields := []otlog.Field{
			otlog.String("event", stat.Event),
			otlog.Int("bytes", stat.Bytes),
			otlog.Int64("duration_ms", stat.Duration.Milliseconds()),
		}
		if stat.Error != nil {
			fields = append(fields, otlog.Error(stat.Error))
		}
		log(fields...)
	}
}
