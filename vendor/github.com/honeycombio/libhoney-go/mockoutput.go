package libhoney

import "github.com/honeycombio/libhoney-go/transmission"

// MockOutput implements the Output interface and passes it along to the
// transmission.MockSender.
//
// Deprecated: Please use the transmission.MockSender directly instead.
// It is provided here for backwards compatibility and will be removed eventually.
type MockOutput struct {
	transmission.MockSender
}

func (w *MockOutput) Add(ev *Event) {
	transEv := &transmission.Event{
		APIHost:    ev.APIHost,
		APIKey:     ev.WriteKey,
		Dataset:    ev.Dataset,
		SampleRate: ev.SampleRate,
		Timestamp:  ev.Timestamp,
		Metadata:   ev.Metadata,
		Data:       ev.data,
	}
	w.MockSender.Add(transEv)
}

func (w *MockOutput) Events() []*Event {
	evs := []*Event{}
	for _, ev := range w.MockSender.Events() {
		transEv := &Event{
			APIHost:    ev.APIHost,
			WriteKey:   ev.APIKey,
			Dataset:    ev.Dataset,
			SampleRate: ev.SampleRate,
			Timestamp:  ev.Timestamp,
			Metadata:   ev.Metadata,
		}
		transEv.data = ev.Data
		evs = append(evs, transEv)
	}
	return evs
}
