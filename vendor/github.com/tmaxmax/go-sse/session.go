package sse

import (
	"errors"
	"net/http"
)

// ResponseWriter is a http.ResponseWriter augmented with a Flush method.
type ResponseWriter interface {
	http.ResponseWriter
	Flush() error
}

// MessageWriter is a special kind of response writer used by providers to
// send Messages to clients.
type MessageWriter interface {
	// Sens sends the message to the client.
	// To make sure it is sent, call Flush.
	Send(m *Message) error
	// Flush sends any buffered messages to the client.
	Flush() error
}

// A Session is an HTTP request from an SSE client.
// Create one using the Upgrade function.
//
// Using a Session you can also access the initial HTTP request,
// get the last event ID, or write data to the client.
type Session struct {
	// The response writer for the request. Can be used to write an error response
	// back to the client. Must not be used after the Session was subscribed!
	Res ResponseWriter
	// The initial HTTP request. Can be used to retrieve authentication data,
	// topics, or data from context – a logger, for example.
	Req *http.Request
	// Last event ID of the client. It is unset if no ID was provided in the Last-Event-Id
	// request header.
	LastEventID EventID

	didUpgrade bool
}

// Send sends the given event to the client. It returns any errors that occurred while writing the event.
func (s *Session) Send(e *Message) error {
	if err := s.doUpgrade(); err != nil {
		return err
	}
	if _, err := e.WriteTo(s.Res); err != nil {
		return err
	}
	return nil
}

// Flush sends any buffered messages to the client.
func (s *Session) Flush() error {
	prevDidUpgrade := s.didUpgrade
	if err := s.doUpgrade(); err != nil {
		return err
	}
	if prevDidUpgrade == s.didUpgrade {
		return s.Res.Flush()
	}
	return nil
}

func (s *Session) doUpgrade() error {
	if !s.didUpgrade {
		s.Res.Header()[headerContentType] = headerContentTypeValue
		if err := s.Res.Flush(); err != nil {
			return err
		}
		s.didUpgrade = true
	}
	return nil
}

// Upgrade upgrades an HTTP request to support server-sent events.
// It returns a Session that's used to send events to the client, or an
// error if the upgrade failed.
//
// The headers required by the SSE protocol are only sent when calling
// the Send method for the first time. If other operations are done before
// sending messages, other headers and status codes can safely be set.
func Upgrade(w http.ResponseWriter, r *http.Request) (*Session, error) {
	rw := getResponseWriter(w)
	if rw == nil {
		return nil, ErrUpgradeUnsupported
	}

	id := EventID{}
	// Clients must not send empty Last-Event-Id headers:
	// https://html.spec.whatwg.org/multipage/server-sent-events.html#sse-processing-model
	if h := r.Header[headerLastEventID]; len(h) != 0 && h[0] != "" {
		// We ignore the validity flag because if the given ID is invalid then an unset ID will be returned,
		// which providers are required to ignore.
		id, _ = NewID(h[0])
	}

	return &Session{Req: r, Res: rw, LastEventID: id}, nil
}

// ErrUpgradeUnsupported is returned when a request can't be upgraded to support server-sent events.
var ErrUpgradeUnsupported = errors.New("go-sse.server: upgrade unsupported")

// Canonicalized header keys.
const (
	headerLastEventID = "Last-Event-Id"
	headerContentType = "Content-Type"
)

// Pre-allocated header value.
var headerContentTypeValue = []string{"text/event-stream"}

// Logic below is similar to Go 1.20's ResponseController.
// We can't use that because we need to check if the request supports
// flushing messages before we subscribe it to the event stream.

type writeFlusher interface {
	http.ResponseWriter
	http.Flusher
}

type writeFlusherError interface {
	http.ResponseWriter
	FlushError() error
}

type rwUnwrapper interface {
	Unwrap() http.ResponseWriter
}

func getResponseWriter(w http.ResponseWriter) ResponseWriter {
	for {
		switch v := w.(type) {
		case writeFlusherError:
			return flusherErrorWrapper{v}
		case writeFlusher:
			return flusherWrapper{v}
		case rwUnwrapper:
			w = v.Unwrap()
		default:
			return nil
		}
	}
}

type flusherWrapper struct {
	writeFlusher
}

func (f flusherWrapper) Flush() error {
	f.writeFlusher.Flush()
	return nil
}

type flusherErrorWrapper struct {
	writeFlusherError
}

func (f flusherErrorWrapper) Flush() error { return f.FlushError() }
