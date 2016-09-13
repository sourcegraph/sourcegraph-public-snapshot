package eventsutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sqs/pbtypes"
)

const AnalyticsAPIEndpoint = "https://analytics.sgdev.org/events"
const MaxRetries = 5

type Worker struct {
	Buffer   []*sourcegraph.Event
	Position int

	Channel chan *sourcegraph.Event
	AppURL  *url.URL

	retryCounter int
}

func (w *Worker) Work() {
	for {
		if w.Position >= len(w.Buffer) {
			if err := w.Flush(); err != nil {
				// Flush didn't succeed and buffer is full
				// so don't dequeue the new event.
				// Dequeue after a short interval to avoid
				// immediately hitting the analytics endpoint.
				time.Sleep(30 * time.Second)
				continue
			}
		}
		event := <-w.Channel
		if event.Type == "internal:flush" {
			w.Flush()
			continue
		}
		w.Buffer[w.Position] = event
		w.Position += 1
	}
}

// Flush pushes the local event buffer upstream to the analytics gateway,
// from all Sourcegraph instances.
// If Flush fails, the event buffer is not modified. Repeatedly failing
// to flush events will fill the local buffer and eventually newer events
// will start getting discarded.
func (w *Worker) Flush() error {
	if w.Position == 0 {
		return nil
	}
	eventList := &sourcegraph.EventList{
		Events:  w.Buffer[:w.Position],
		Version: buildvar.Version,
	}

	if w.AppURL != nil {
		eventList.AppURL = w.AppURL.String()
	}

	err := w.sendEvents(eventList)
	if err != nil && w.retryCounter < MaxRetries {
		w.retryCounter += 1
		return err
	}

	// Flush successful or max retries reached.
	w.Position = 0
	w.retryCounter = 0
	return nil
}

type PostData struct {
	ClientSecret string                `json:"client_secret,omitempty"`
	Events       sourcegraph.EventList `json:"event_data,omitempty"`
}

// sendEvents sends event data to the analytics gateway via an
// HTTP POST request.
func (w *Worker) sendEvents(srcEvents *sourcegraph.EventList) error {
	eventData := &PostData{
		ClientSecret: os.Getenv("SG_ANALYTICS_SECRET"),
		Events:       *srcEvents,
	}
	postData, err := json.Marshal(eventData)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", AnalyticsAPIEndpoint, bytes.NewBuffer([]byte(postData)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)
		log15.Error("Failed to send events to analytics gateway", "status", resp.Status, "error", bodyStr)
		return fmt.Errorf("%v: %v", resp.Status, bodyStr)
	}
	log15.Debug("Sent events to analytics gateway", "num", len(srcEvents.Events))
	return nil
}

type Logger struct {
	Channel chan *sourcegraph.Event
	Worker  *Worker
}

func (l *Logger) Log(event *sourcegraph.Event) {
	select {
	case l.Channel <- event:
	case <-time.After(10 * time.Millisecond):
		// Discard log message
		log15.Debug("Events logger discarding log event: buffer full")
	}
}

func (l *Logger) Uploader(flushInterval time.Duration) {
	// For the first 60 minutes after boot up, flush log every minute
	remainingMinutes := 60
	if flushInterval <= time.Minute {
		remainingMinutes = 0
	}
	for {
		if remainingMinutes > 0 {
			time.Sleep(time.Minute)
			remainingMinutes -= 1
		} else {
			time.Sleep(flushInterval)
		}
		l.Log(&sourcegraph.Event{Type: "internal:flush"})
	}
}

var ActiveLogger *Logger

// StartEventLogger sets up a buffered channel for posting events to, and workers that consume
// event messages from that channel.
// channelCapacity is the max number of events that the channel will hold. Newer events will be
// dropped when the channel is full.
// Each worker pulls events off the channel and pushes to it's buffer. workerBufferSize is the
// maximum number of buffered events after which the worker will flush the buffer upstream.
func StartEventLogger(ctx context.Context, channelCapacity, workerBufferSize int, flushInterval time.Duration) {
	ActiveLogger = &Logger{
		Channel: make(chan *sourcegraph.Event, channelCapacity),
	}

	ActiveLogger.Worker = &Worker{
		Buffer:  make([]*sourcegraph.Event, workerBufferSize),
		Channel: ActiveLogger.Channel,
		AppURL:  conf.AppURL(ctx),
	}

	go ActiveLogger.Worker.Work()
	go ActiveLogger.Uploader(flushInterval)
	log15.Debug("Events logger initialized")
}

// Log adds a sourcegraph.Event to the local log buffer, which
// will be periodically flushed upstream.
func Log(event *sourcegraph.Event) {
	if ActiveLogger == nil {
		return
	}
	if event.Timestamp == nil {
		ts := pbtypes.NewTimestamp(time.Now().UTC())
		event.Timestamp = &ts
	}

	ActiveLogger.Log(event)
	log15.Debug("Logged event", "event", event)
}
