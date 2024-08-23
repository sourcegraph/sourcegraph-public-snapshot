package libhoney

import (
	"github.com/honeycombio/libhoney-go/transmission"
)

// WriterOutput implements the Output interface and passes it along to the
// transmission.WriterSender.
//
// Deprecated: Please use the transmission.WriterSender directly instead.
// It is provided here for backwards compatibility and will be removed eventually.
type WriterOutput struct {
	transmission.WriterSender
}

func (w *WriterOutput) Add(ev *Event) {
	transEv := &transmission.Event{
		APIHost:    ev.APIHost,
		APIKey:     ev.WriteKey,
		Dataset:    ev.Dataset,
		SampleRate: ev.SampleRate,
		Timestamp:  ev.Timestamp,
		Metadata:   ev.Metadata,
		Data:       ev.data,
	}
	w.WriterSender.Add(transEv)
}

// DiscardWriter implements the Output interface and drops all events.
//
// Deprecated: Please use the transmission.DiscardSender directly instead.
// It is provided here for backwards compatibility and will be removed eventually.
type DiscardOutput struct {
	WriterOutput
}

func (d *DiscardOutput) Add(ev *Event) {}
