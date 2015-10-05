package metricutil

import (
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/fed"
)

type Worker struct {
	Buffer   []*sourcegraph.UserEvent
	Position int

	Channel chan *sourcegraph.UserEvent
	RootCtx context.Context
}

func (w *Worker) Work() {
	for {
		event := <-w.Channel
		if w.Position >= len(w.Buffer) {
			w.Flush()
		}
		w.Buffer[w.Position] = event
		w.Position += 1
	}
}

func (w *Worker) Flush() {
	if w.Position == 0 {
		return
	}
	eventList := &sourcegraph.UserEventList{Events: w.Buffer[:w.Position]}
	if fed.Config.IsRoot {
		ForwardEvents(w.RootCtx, eventList)
	} else {
		cl := sourcegraph.NewClientFromContext(w.RootCtx)
		_, err := cl.GraphUplink.PushEvents(w.RootCtx, eventList)
		if err != nil {
			log15.Error("GraphUplink.PushEvents failed", "error", err)
		}
	}
	w.Position = 0
}

type Logger struct {
	Channel chan *sourcegraph.UserEvent
	Worker  *Worker
}

func (l *Logger) Log(ctx context.Context, event *sourcegraph.UserEvent) {
	if !l.Filter(ctx, event) {
		return
	}

	select {
	case l.Channel <- event:
	case <-time.After(10 * time.Millisecond):
		// Discard log message
		log15.Debug("EventLogger discarding log event: buffer full")
	}
}

func (l *Logger) Filter(ctx context.Context, event *sourcegraph.UserEvent) bool {
	if event.Type != "grpc" {
		// all events that are not grpc calls are important
		return true
	}
	if fed.Config.IsRoot {
		// don't track grpc events on mothership
		return false
	}
	if event.UID == 0 && !authutil.ActiveFlags.AllowAnonymousReaders {
		// this is not a user initiated event
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

var ActiveLogger *Logger

// StartEventLogger sets up a buffered channel for posting events to, and workers that consume
// event messages from that channel.
// channelCapacity is the max number of events that the channel will hold. Newer events will be
// dropped when the channel is full.
// Each worker pulls events off the channel and pushes to it's buffer. workerBufferSize is the
// maximum number of buffered events after which the worker will flush the buffer upstream to
// the federation root via graph uplink.
func StartEventLogger(ctx context.Context, channelCapacity, workerBufferSize int) {
	rootCtx := ctx
	if !fed.Config.IsRoot {
		mothership, err := fed.Config.RootGRPCEndpoint()
		if err != nil {
			log15.Error("EventLogger could not identify the mothership", "error", err)
			return
		}
		rootCtx = sourcegraph.WithGRPCEndpoint(ctx, mothership)
	}

	ActiveLogger = &Logger{
		Channel: make(chan *sourcegraph.UserEvent, channelCapacity),
	}

	ActiveLogger.Worker = &Worker{
		Buffer:  make([]*sourcegraph.UserEvent, workerBufferSize),
		Channel: ActiveLogger.Channel,
		RootCtx: rootCtx,
	}

	go ActiveLogger.Worker.Work()

	log15.Debug("EventLogger initialized")
}

func LogEvent(ctx context.Context, event *sourcegraph.UserEvent) {
	if ActiveLogger != nil {
		if event.UID == 0 {
			event.UID = int32(authpkg.ActorFromContext(ctx).UID)
		}

		if event.CreatedAt == nil {
			ts := pbtypes.NewTimestamp(time.Now())
			event.CreatedAt = &ts
		}

		ActiveLogger.Log(ctx, event)
	}
}
