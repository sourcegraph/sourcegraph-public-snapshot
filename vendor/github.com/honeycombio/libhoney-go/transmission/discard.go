package transmission

// DiscardSender implements the Sender interface and drops all events.
type DiscardSender struct {
	WriterSender
}

func (d *DiscardSender) Add(ev *Event) {}
