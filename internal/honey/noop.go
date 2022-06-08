package honey

import "github.com/opentracing/opentracing-go/log"

type noopEvent struct {
	fields map[string]any
}

var _ Event = noopEvent{}

// NoopEvent returns an Event who's methods do nothing and
// return nil where applicable.
func NoopEvent() Event { return noopEvent{fields: map[string]any{}} }

func (noopEvent) Dataset() string { return "" }

func (e noopEvent) AddField(key string, val any) { e.fields[key] = val }

func (noopEvent) AddLogFields(_ []log.Field) {}

func (noopEvent) Add(_ any) error { return nil }

func (e noopEvent) Fields() map[string]any { return e.fields }

func (noopEvent) SetSampleRate(rate uint) {}

func (noopEvent) Send() error { return nil }
