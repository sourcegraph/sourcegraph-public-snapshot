package honey

import "github.com/opentracing/opentracing-go/log"

type noopEvent struct{}

var _ Event = noopEvent{}

// NoopEvent returns an Event who's methods do nothing and
// return nil where applicable.
func NoopEvent() Event { return noopEvent{} }

func (noopEvent) Dataset() string { return "" }

func (noopEvent) AddField(_ string, _ any) {}

func (noopEvent) AddLogFields(_ []log.Field) {}

func (noopEvent) Add(_ any) error { return nil }

func (noopEvent) Fields() map[string]any { return nil }

func (noopEvent) SetSampleRate(rate uint) {}

func (noopEvent) Send() error { return nil }
