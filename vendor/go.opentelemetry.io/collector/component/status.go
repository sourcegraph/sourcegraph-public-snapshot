// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package component // import "go.opentelemetry.io/collector/component"

import (
	"time"
)

type Status int32

// Enumeration of possible component statuses
const (
	StatusNone Status = iota
	StatusStarting
	StatusOK
	StatusRecoverableError
	StatusPermanentError
	StatusFatalError
	StatusStopping
	StatusStopped
)

// String returns a string representation of a Status
func (s Status) String() string {
	switch s {
	case StatusStarting:
		return "StatusStarting"
	case StatusOK:
		return "StatusOK"
	case StatusRecoverableError:
		return "StatusRecoverableError"
	case StatusPermanentError:
		return "StatusPermanentError"
	case StatusFatalError:
		return "StatusFatalError"
	case StatusStopping:
		return "StatusStopping"
	case StatusStopped:
		return "StatusStopped"
	}
	return "StatusNone"
}

// StatusEvent contains a status and timestamp, and can contain an error
type StatusEvent struct {
	status    Status
	err       error
	timestamp time.Time
}

// Status returns the Status (enum) associated with the StatusEvent
func (ev *StatusEvent) Status() Status {
	return ev.status
}

// Err returns the error associated with the StatusEvent.
func (ev *StatusEvent) Err() error {
	return ev.err
}

// Timestamp returns the timestamp associated with the StatusEvent
func (ev *StatusEvent) Timestamp() time.Time {
	return ev.timestamp
}

// NewStatusEvent creates and returns a StatusEvent with the specified status and sets the timestamp
// time.Now(). To set an error on the event for an error status use one of the dedicated
// constructors (e.g. NewRecoverableErrorEvent, NewPermanentErrorEvent, NewFatalErrorEvent)
func NewStatusEvent(status Status) *StatusEvent {
	return &StatusEvent{
		status:    status,
		timestamp: time.Now(),
	}
}

// NewRecoverableErrorEvent creates and returns a StatusEvent with StatusRecoverableError, the
// specified error, and a timestamp set to time.Now().
func NewRecoverableErrorEvent(err error) *StatusEvent {
	ev := NewStatusEvent(StatusRecoverableError)
	ev.err = err
	return ev
}

// NewPermanentErrorEvent creates and returns a StatusEvent with StatusPermanentError, the
// specified error, and a timestamp set to time.Now().
func NewPermanentErrorEvent(err error) *StatusEvent {
	ev := NewStatusEvent(StatusPermanentError)
	ev.err = err
	return ev
}

// NewFatalErrorEvent creates and returns a StatusEvent with StatusFatalError, the
// specified error, and a timestamp set to time.Now().
func NewFatalErrorEvent(err error) *StatusEvent {
	ev := NewStatusEvent(StatusFatalError)
	ev.err = err
	return ev
}

// AggregateStatus will derive a status for the given input using the following rules in order:
//  1. If all instances have the same status, there is nothing to aggregate, return it.
//  2. If any instance encounters a fatal error, the component is in a Fatal Error state.
//  3. If any instance is in a Permanent Error state, the component status is Permanent Error.
//  4. If any instance is Stopping, the component is in a Stopping state.
//  5. An instance is Stopped, but not all instances are Stopped, we must be in the process of Stopping the component.
//  6. If any instance is in a Recoverable Error state, the component status is Recoverable Error.
//  7. By process of elimination, the only remaining state is starting.
func AggregateStatus[K comparable](eventMap map[K]*StatusEvent) Status {
	seen := make(map[Status]struct{})
	for _, ev := range eventMap {
		seen[ev.Status()] = struct{}{}
	}

	// All statuses are the same. Note, this will handle StatusOK and StatusStopped as these two
	// cases require all components be in the same state.
	if len(seen) == 1 {
		for st := range seen {
			return st
		}
	}

	// Handle mixed status cases
	if _, isFatal := seen[StatusFatalError]; isFatal {
		return StatusFatalError
	}

	if _, isPermanent := seen[StatusPermanentError]; isPermanent {
		return StatusPermanentError
	}

	if _, isStopping := seen[StatusStopping]; isStopping {
		return StatusStopping
	}

	if _, isStopped := seen[StatusStopped]; isStopped {
		return StatusStopping
	}

	if _, isRecoverable := seen[StatusRecoverableError]; isRecoverable {
		return StatusRecoverableError
	}

	// By process of elimination, this is the last possible status; no check necessary.
	return StatusStarting
}

// StatusIsError returns true for error statuses (e.g. StatusRecoverableError,
// StatusPermanentError, or StatusFatalError)
func StatusIsError(status Status) bool {
	return status == StatusRecoverableError ||
		status == StatusPermanentError ||
		status == StatusFatalError
}

// AggregateStatusEvent returns a status event where:
//   - The status is set to the aggregate status of the events in the eventMap
//   - The timestamp is set to the latest timestamp of the events in the eventMap
//   - For an error status, the event will have same error as the most current event of the same
//     error type from the eventMap
func AggregateStatusEvent[K comparable](eventMap map[K]*StatusEvent) *StatusEvent {
	var lastEvent, lastMatchingEvent *StatusEvent
	aggregateStatus := AggregateStatus[K](eventMap)

	for _, ev := range eventMap {
		if lastEvent == nil || lastEvent.timestamp.Before(ev.timestamp) {
			lastEvent = ev
		}
		if aggregateStatus == ev.Status() &&
			(lastMatchingEvent == nil || lastMatchingEvent.timestamp.Before(ev.timestamp)) {
			lastMatchingEvent = ev
		}
	}

	// the effective status matches an existing event
	if lastEvent.Status() == aggregateStatus {
		return lastEvent
	}

	// the effective status requires a synthetic event
	aggregateEvent := &StatusEvent{
		status:    aggregateStatus,
		timestamp: lastEvent.timestamp,
	}
	if StatusIsError(aggregateStatus) {
		aggregateEvent.err = lastMatchingEvent.err
	}

	return aggregateEvent
}
