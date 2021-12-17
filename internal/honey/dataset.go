package honey

// Dataset represents a Honeycomb dataset to which events can be sent.
// This provides an alternative to calling `honey.NewEvent`/`honey.NewEventWithFields`
// with a provided dataset name.
type Dataset struct {
	Name string
}

func (d *Dataset) Event() Event {
	return NewEvent(d.Name)
}

func (d *Dataset) EventWithFields(fields map[string]interface{}) Event {
	return NewEventWithFields(d.Name, fields)
}
