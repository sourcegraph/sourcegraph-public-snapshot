package honey

// Dataset represents a Honeycomb dataset to which events can be sent.
// This provides an alternative to calling `honey.NewEvent`/`honey.NewEventWithFields`
// with a provided dataset name.
type Dataset struct {
	Name string
	// SetSampleRate overrides the global sample rate for events of this dataset.
	// Values less than or equal to 1 mean no sampling (aka all events are sent).
	// If you want to send one event out of every 250, you would specify 250 here.
	SampleRate uint
}

func (d *Dataset) Event() Event {
	event := NewEvent(d.Name)
	if d.SampleRate > 1 {
		event.SetSampleRate(d.SampleRate)
	}
	return event
}

func (d *Dataset) EventWithFields(fields map[string]any) Event {
	event := NewEventWithFields(d.Name, fields)
	if d.SampleRate > 1 {
		event.SetSampleRate(d.SampleRate)
	}
	return event
}
