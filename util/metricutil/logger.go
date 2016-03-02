package metricutil

import (
	"bytes"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/env"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
)

type Worker struct {
	Buffer   []*sourcegraph.UserEvent
	Position int

	Channel              chan *sourcegraph.UserEvent
	Ctx                  context.Context
	forwardDestCtx       context.Context
	forwardDestAvailable bool
}

func (w *Worker) Work() {
	for {
		if w.Position >= len(w.Buffer) {
			if err := w.Flush(); err != nil {
				// Flush didn't succeed and buffer is full so don't
				// dequeue the new event.  Dequeue after a short
				// interval to avoid immediately re-connecting to the
				// forward dest server.
				time.Sleep(30 * time.Second)
				continue
			}
		}
		event := <-w.Channel
		if event.Type == "command" {
			switch event.Method {
			case "flush":
				w.Flush()
			default:
				log15.Debug("Metrics logger got unknown command", "command", event.Method)
			}
			continue
		}
		w.Buffer[w.Position] = event
		w.Position += 1
	}
}

// connectToForwardDestServer connects to the server that the worker
// is configured to forward metrics to.
func (w *Worker) connectToForwardDestServer() error {
	url, err := url.Parse(config.ForwardURL)
	if err != nil {
		return err
	}
	w.forwardDestCtx = sourcegraph.WithGRPCEndpoint(w.Ctx, url)
	w.forwardDestAvailable = true
	return nil
}

// Flush pushes the local event buffer upstream from client instances
// or forwards to EventForwarder from servers that can store metrics.
// If Flush fails, the event buffer is not modified. Repeatedly failing
// to flush events will fill the local buffer and eventually newer events
// will start getting discarded.
func (w *Worker) Flush() error {
	if w.Position == 0 {
		return nil
	}
	eventList := &sourcegraph.UserEventList{Events: w.Buffer[:w.Position]}
	if config.StoreURL != "" {
		StoreEvents(w.Ctx, eventList)
	}
	if config.ForwardURL != "" {
		if !w.forwardDestAvailable {
			if err := w.connectToForwardDestServer(); err != nil {
				log15.Error("Metrics logger flush failed to locate server to forward to", "error", err)
				return err
			}
		}
		cl, err := sourcegraph.NewClientFromContext(w.forwardDestCtx)
		if err != nil {
			return err
		}

		// Send events.
		_, err = cl.GraphUplink.PushEvents(w.forwardDestCtx, eventList)
		if err != nil {
			log15.Error("GraphUplink.PushEvents failed", "error", err, "forwardURL", config.ForwardURL)
			// Force the connection to be re-established on the next flush.
			w.forwardDestAvailable = false
			return err
		}

		// Send metrics.
		buf := &bytes.Buffer{}
		mfs := snapshotMetricFamilies()
		if err := mfs.Marshal(buf); err != nil {
			log15.Error("Error marshaling metric families", "error", err)
			return err
		}
		log15.Debug("GraphUplink sending metrics snapshot", "forwardURL", config.ForwardURL, "numMetrics", len(mfs))
		snapshot := sourcegraph.MetricsSnapshot{
			Type:          sourcegraph.TelemetryType_PrometheusDelimited0dot0dot4,
			TelemetryData: buf.Bytes(),
		}
		if _, err := cl.GraphUplink.Push(w.forwardDestCtx, &snapshot); err != nil {
			log15.Error("GraphUplink push failed", "error", err, "forwardURL", config.ForwardURL)
			// Force the connection to be re-established on the next flush.
			w.forwardDestAvailable = false
			return err
		}

		log15.Debug("Forwarded metrics and events", "forwardURL", config.ForwardURL, "numMetrics", len(mfs), "numEvents", len(eventList.Events))
	}
	// Flush successful
	w.Position = 0
	return nil
}

type logger struct {
	Channel chan *sourcegraph.UserEvent
	Worker  *Worker
}

func (l *logger) Log(ctx context.Context, event *sourcegraph.UserEvent) {
	if !l.Filter(ctx, event) {
		return
	}

	select {
	case l.Channel <- event:
	case <-time.After(10 * time.Millisecond):
		// Discard log message
		log15.Debug("Metrics logger discarding log event: buffer full")
	}
}

func (l *logger) Filter(ctx context.Context, event *sourcegraph.UserEvent) bool {
	if event.Type != "grpc" {
		// all events that are not grpc calls are important
		return true
	}
	if event.UID == 0 && !authutil.ActiveFlags.AllowAnonymousReaders {
		// this is not a user initiated grpc call
		return false
	}
	switch event.Service {
	case "GraphUplink":
		return false
	case "Builds":
		switch event.Method {
		case "Create", "Update":
			return true
		default:
			return false
		}
	default:
		return true
	}
}

func (l *logger) Uploader(ctx context.Context, flushInterval time.Duration) {
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
		l.Log(ctx, &sourcegraph.UserEvent{
			Type:   "command",
			Method: "flush",
		})
	}
}

var activeLogger *logger

// startEventLogger sets up a buffered channel for posting events to, and workers that consume
// event messages from that channel.
// channelCapacity is the max number of events that the channel will hold. Newer events will be
// dropped when the channel is full.
// Each worker pulls events off the channel and pushes to it's buffer. workerBufferSize is the
// maximum number of buffered events after which the worker will flush the buffer upstream.
func startEventLogger(ctx context.Context, channelCapacity, workerBufferSize int, flushInterval time.Duration) {
	activeLogger = &logger{
		Channel: make(chan *sourcegraph.UserEvent, channelCapacity),
	}

	activeLogger.Worker = &Worker{
		Buffer:  make([]*sourcegraph.UserEvent, workerBufferSize),
		Channel: activeLogger.Channel,
		Ctx:     ctx,
	}

	go activeLogger.Worker.Work()

	go activeLogger.Uploader(ctx, flushInterval)

	log15.Debug("Metrics logger initialized")
}

// LogEvent adds a sourcegraph.UserEvent to the local log buffer, which
// will be periodically flushed upstream.
func LogEvent(ctx context.Context, event *sourcegraph.UserEvent) {
	if activeLogger != nil {
		if event.UID == 0 {
			event.UID = int32(authpkg.ActorFromContext(ctx).UID)
		}

		if event.ClientID == "" {
			event.ClientID = authpkg.ActorFromContext(ctx).ClientID
		}

		if event.CreatedAt == nil {
			ts := pbtypes.NewTimestamp(time.Now().UTC())
			event.CreatedAt = &ts
		}

		if event.Version == "" {
			event.Version = buildvar.Version
		}

		activeLogger.Log(ctx, event)
	}
}

// LogConfig dumps config info about the current server into the local
// log buffer, to push upstream for diagnostic purposes.
// The config dump contains:
//
//   1. Build information about the current src binary.
//   2. Commandline flags to `src serve`.
//   3. Env variables of the current process, relevant to the Sourcegraph
//      installation.
//
// The flag data must be sanitized of secrets before passing in to this function.
func LogConfig(ctx context.Context, clientID, flagsSafe string) {
	if activeLogger != nil {
		LogEvent(ctx, &sourcegraph.UserEvent{
			Type:     "notif",
			ClientID: clientID,
			Service:  "config",
			Method:   "buildvars",
			Message:  fmt.Sprintf("%+v", buildvar.All),
		})

		LogEvent(ctx, &sourcegraph.UserEvent{
			Type:     "notif",
			ClientID: clientID,
			Service:  "config",
			Method:   "flags",
			Message:  flagsSafe,
		})

		env := env.GetWhitelistedEnvironment()
		env = append(env, "GOOS="+runtime.GOOS)
		env = append(env, "GOARCH="+runtime.GOARCH)

		LogEvent(ctx, &sourcegraph.UserEvent{
			Type:     "notif",
			ClientID: clientID,
			Service:  "config",
			Method:   "env",
			Message:  fmt.Sprintf("%+v", env),
		})
	}
}
