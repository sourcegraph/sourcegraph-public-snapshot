package honey

import (
	"go.opentelemetry.io/otel/attribute"
)

type noopEvent struct{}

var _ Event = noopEvent{}

// NoopEvent returns an Event who's methods do nothing and
// return nil where applicable.
func NoopEvent() Event { return noopEvent{} }

func (noopEvent) Dataset() string { return "" }

func (noopEvent) AddField(_ string, _ any) {}

func (noopEvent) AddAttributes(_ []attribute.KeyValue) {}

func (noopEvent) Add(_ any) error { return nil }

func (noopEvent) Fields() map[string]any { return nil }

func (noopEvent) SetSampleRate(rate uint) {}

func (noopEvent) Send() error { return nil }
