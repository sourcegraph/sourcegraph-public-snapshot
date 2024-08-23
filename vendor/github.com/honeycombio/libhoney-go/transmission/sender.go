package transmission

// Sender is responsible for handling events after Send() is called.
// Implementations of Add() must be safe for concurrent calls.
type Sender interface {

	// Add queues up an event to be sent
	Add(ev *Event)

	// Start initializes any background processes necessary to send events
	Start() error

	// Stop flushes any pending queues and blocks until everything in flight has
	// been sent. Once called, you cannot call Add unless Start has subsequently
	// been called.
	Stop() error

	// Flush flushes any pending queues and blocks until everything in flight has
	// been sent.
	Flush() error

	// Responses returns a channel that will contain a single Response for each
	// Event added. Note that they may not be in the same order as they came in
	TxResponses() chan Response

	// SendResponse adds a Response to the Responses queue. It should be added
	// for events handed to libhoney that are dropped before they even make it
	// to the Transmission Sender (for example because of sampling) to maintain
	// libhoney's guarantee that each event given to it will generate one event
	// in the Responses channel.
	SendResponse(Response) bool
}
