package streaming

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	streamclient "github.com/sourcegraph/sourcegraph/internal/search/streaming/client"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// maxRequestDuration clamps any compute queries to run for at most 1 minute.
// It's possible to trigger longer-running queries with expensive operations,
// and this is best avoided on large instances like Sourcegraph.com
const maxRequestDuration = time.Minute

// NewComputeStreamHandler is an http handler which streams back compute results.
func NewComputeStreamHandler(logger log.Logger, db database.DB) http.Handler {
	return &streamHandler{
		logger:              logger,
		db:                  db,
		flushTickerInternal: 100 * time.Millisecond,
		pingTickerInterval:  5 * time.Second,
	}
}

type streamHandler struct {
	logger              log.Logger
	db                  database.DB
	flushTickerInternal time.Duration
	pingTickerInterval  time.Duration
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
	defer cancel()
	start := time.Now()

	args, err := parseURLQuery(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tr, ctx := trace.New(ctx, "compute.ServeStream", args.Query)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	computeQuery, err := compute.Parse(args.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	searchQuery, err := computeQuery.ToSearchQuery()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	progress := &streamclient.ProgressAggregator{
		Start:     start,
		RepoNamer: streamclient.RepoNamer(ctx, h.db),
		Trace:     trace.URL(trace.ID(ctx), conf.DefaultClient()),
	}

	sendProgress := func() {
		_ = eventWriter.Event("progress", progress.Current())
	}

	// Always send a final done event so clients know the stream is shutting
	// down.
	defer eventWriter.Event("done", map[string]any{})

	// Log events to trace
	eventWriter.StatHook = eventStreamOTHook(tr.LogFields)

	events, getResults := NewComputeStream(ctx, h.logger, h.db, searchQuery, computeQuery.Command)
	events = batchEvents(events, 50*time.Millisecond)

	// Store marshalled matches and flush periodically or when we go over
	// 32kb. 32kb chosen to be smaller than bufio.MaxTokenSize. Note: we can
	// still write more than that.
	matchesBuf := streamhttp.NewJSONArrayBuf(32*1024, func(data []byte) error {
		return eventWriter.EventBytes("results", data)
	})
	matchesFlush := func() {
		if err := matchesBuf.Flush(); err != nil {
			// EOF
			return
		}

		if progress.Dirty {
			sendProgress()
		}
	}
	flushTicker := time.NewTicker(h.flushTickerInternal)
	defer flushTicker.Stop()

	pingTicker := time.NewTicker(h.pingTickerInterval)
	defer pingTicker.Stop()

	first := true
	handleEvent := func(event Event) {
		progress.Dirty = true
		progress.Stats.Update(&event.Stats)

		for _, result := range event.Results {
			_ = matchesBuf.Append(result)
		}

		// Instantly send results if we have not sent any yet.
		if first && matchesBuf.Len() > 0 {
			first = false
			matchesFlush()
		}
	}

LOOP:
	for {
		select {
		case event, ok := <-events:
			if !ok {
				break LOOP
			}
			handleEvent(event)
		case <-flushTicker.C:
			matchesFlush()
		case <-pingTicker.C:
			sendProgress()
		}
	}

	matchesFlush()

	alert, err := getResults()
	if err != nil {
		_ = eventWriter.Event("error", streamhttp.EventError{Message: err.Error()})
		return
	}

	if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
		_ = eventWriter.Event("alert", streamhttp.EventAlert{
			Title:       "Incomplete data",
			Description: "This data is incomplete! We ran this query for 1 minute and we'd need more time to compute all the results. This isn't supported yet, so please reach out to support@sourcegraph.com if you're interested in running longer queries.",
		})
	}
	if alert != nil {
		var pqs []streamhttp.QueryDescription
		for _, pq := range alert.ProposedQueries {
			pqs = append(pqs, streamhttp.QueryDescription{
				Description: pq.Description,
				Query:       pq.QueryString(),
			})
		}
		_ = eventWriter.Event("alert", streamhttp.EventAlert{
			Title:           alert.Title,
			Description:     alert.Description,
			ProposedQueries: pqs,
		})
	}

	_ = eventWriter.Event("progress", progress.Final())
}

type args struct {
	Query   string
	Display int
}

func parseURLQuery(q url.Values) (*args, error) {
	get := func(k, def string) string {
		v := q.Get(k)
		if v == "" {
			return def
		}
		return v
	}

	a := args{
		Query: get("q", ""),
	}

	if a.Query == "" {
		return nil, errors.New("no query found")
	}

	display := get("display", "-1") // TODO(rvantonder): Currently unused; implement a limit for compute results.
	var err error
	if a.Display, err = strconv.Atoi(display); err != nil {
		return nil, errors.Errorf("display must be an integer, got %q: %w", display, err)
	}

	return &a, nil
}

// batchEvents takes an event stream and merges events that come through close in time into a single event.
// This makes downstream database and network operations more efficient by enabling batch reads.
func batchEvents(source <-chan Event, delay time.Duration) <-chan Event {
	results := make(chan Event)
	go func() {
		defer close(results)

		// Send the first event without a delay
		firstEvent, ok := <-source
		if !ok {
			return
		}
		results <- firstEvent

	OUTER:
		for {
			// Wait for a first event
			event, ok := <-source
			if !ok {
				return
			}

			// Wait up to the delay for more events to come through,
			// and merge any that do into the first event
			timer := time.After(delay)
			for {
				select {
				case newEvent, ok := <-source:
					if !ok {
						// Flush the buffered event and exit
						results <- event
						return
					}
					event.Results = append(event.Results, newEvent.Results...)
				case <-timer:
					results <- event
					continue OUTER
				}
			}
		}

	}()
	return results
}

// eventStreamOTHook returns a StatHook which logs to log.
func eventStreamOTHook(log func(...otlog.Field)) func(streamhttp.WriterStat) {
	return func(stat streamhttp.WriterStat) {
		fields := []otlog.Field{
			otlog.String("streamhttp.Event", stat.Event),
			otlog.Int("bytes", stat.Bytes),
			otlog.Int64("duration_ms", stat.Duration.Milliseconds()),
		}
		if stat.Error != nil {
			fields = append(fields, otlog.Error(stat.Error))
		}
		log(fields...)
	}
}
