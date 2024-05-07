package honey

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/honeycombio/libhoney-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
)

// Event represents a mockable/noop-able single event in Honeycomb terms, as per
// https://docs.honeycomb.io/getting-started/events-metrics-logs/#structured-events.
type Event interface {
	// Dataset returns the destination dataset of this event
	Dataset() string
	// AddField adds a single key-value pair to this event.
	AddField(key string, val any)
	// AddAttributes adds each otel/attribute key-value field to this event.
	AddAttributes([]attribute.KeyValue)
	// Fields returns all the added fields of the event. The returned map is not safe to
	// be modified concurrently with calls Add/AddField/AddLogFields.
	Fields() map[string]any
	// SetSampleRate overrides the global sample rate for this event. Default is 1,
	// meaning no sampling. If you want to send one event out of every 250 times
	// Send() is called, you would specify 250 here.
	SetSampleRate(rate uint)
	// Send dispatches the event to be sent to Honeycomb, sampling if necessary.
	Send() error
}

type eventWrapper struct {
	event *libhoney.Event
	// contains a map of keys whose values have been slice wrapped aka
	// added more than once already. If theres no entry in sliceWrapped
	// but there is in event for a key, then the to-be-added value is
	// sliceWrapped before insertion and true inserted into sliceWrapped for that key
	sliceWrapped map[string]bool
}

var _ Event = eventWrapper{}

func (w eventWrapper) Dataset() string {
	return w.event.Dataset
}

func (w eventWrapper) AddField(name string, val any) {
	data, ok := w.Fields()[name]
	if !ok {
		data = val
	} else if ok && !w.sliceWrapped[name] {
		data = sliceWrapper{data, val}
		w.sliceWrapped[name] = true
	} else {
		data = append(data.(sliceWrapper), val)
	}
	w.event.AddField(name, data)
}

func (w eventWrapper) AddAttributes(attrs []attribute.KeyValue) {
	for _, attr := range attrs {
		w.AddField(string(attr.Key), attr.Value.AsInterface())
	}
}

func (w eventWrapper) Add(data any) error {
	return w.event.Add(data)
}

func (w eventWrapper) Fields() map[string]any {
	return w.event.Fields()
}

func (w eventWrapper) SetSampleRate(rate uint) {
	w.event.SampleRate = rate
}

func (w eventWrapper) Send() error {
	if local {
		var fields []string
		for k, v := range w.event.Fields() {
			fields = append(fields, fmt.Sprintf("  %s: %v", k, v))
		}
		slices.Sort(fields)
		_, _ = fmt.Fprintf(os.Stderr, "EVENT %s\n%s\n", w.event.Dataset, strings.Join(fields, "\n"))
		return nil
	}
	return w.event.Send()
}

// NewEvent creates an event for logging to dataset. If Enabled() would return false,
// NewEvent returns a noop event. NewEvent.Send will only work if
// Enabled() returns true.
func NewEvent(dataset string) Event {
	ev, _ := newEvent(dataset)
	return ev
}

// NewEventWithFields creates an event for logging to the given dataset. The given
// fields are assigned to the event.
func NewEventWithFields(dataset string, fields map[string]any) Event {
	ev, enabled := newEvent(dataset)
	if enabled {
		for key, value := range fields {
			ev.AddField(key, value)
		}
	}
	return ev
}

// newEvent is a helper used by NewEvent* which returns true if the event is
// not a noop event.
func newEvent(dataset string) (Event, bool) {
	if !Enabled() {
		metricNewEvent.WithLabelValues("false", dataset).Inc()
		return NonSendingEvent(), false
	}
	metricNewEvent.WithLabelValues("true", dataset).Inc()

	ev := libhoney.NewEvent()
	ev.Dataset = dataset + suffix
	return eventWrapper{
		event:        ev,
		sliceWrapped: map[string]bool{},
	}, true
}

// metricNewEvent will help us understand traffic we send to honeycomb as well
// as identify services wanting to log to honeycomb but missing the requisit
// environment variables.
var metricNewEvent = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_honey_event_total",
	Help: "The total number of honeycomb events created (before sampling).",
}, []string{"enabled", "dataset"})
