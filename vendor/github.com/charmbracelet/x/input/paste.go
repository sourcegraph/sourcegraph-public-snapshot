package input

// PasteEvent is an event that is emitted when a terminal receives pasted text
// using bracketed-paste.
type PasteEvent string

// PasteStartEvent is an event that is emitted when a terminal enters
// bracketed-paste mode.
type PasteStartEvent struct{}

// PasteEvent is an event that is emitted when a terminal receives pasted text.
type PasteEndEvent struct{}
