package sse

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tmaxmax/go-sse/internal/parser"
)

// The Event struct represents an event sent to the client by a server.
type Event struct {
	// The last non-empty ID of all the events received. This may not be
	// the ID of the latest event!
	LastEventID string
	// The event's type. It is empty if the event is unnamed.
	Type string
	// The event's payload.
	Data string
}

// EventCallback is a function that is used to receive events from a Connection.
type EventCallback func(Event)

// EventCallbackRemover is a function that removes an already registered callback
// from a connection. Calling it multiple times is a no-op.
type EventCallbackRemover func()

// Connection is a connection to an events stream. Created using the Client struct,
// a Connection processes the incoming events and calls the subscribed event callbacks.
// If the connection to the server temporarily fails, the connection will be reattempted.
// Retry values received from servers will be taken into account.
//
// Connections must not be copied after they are created.
type Connection struct { //nolint:govet // The current order aids readability.
	mu           sync.RWMutex
	request      *http.Request
	callbacks    map[string]map[int]EventCallback
	callbacksAll map[int]EventCallback
	lastEventID  string
	client       Client
	callbackID   int
	isRetry      bool
}

// SubscribeMessages subscribes the given callback to all events without type (without or with empty `event` field).
// Remove the callback by calling the returned function.
func (c *Connection) SubscribeMessages(cb EventCallback) EventCallbackRemover {
	return c.SubscribeEvent("", cb)
}

// SubscribeEvent subscribes the given callback to all the events with the provided type
// (the `event` field has the value given here).
// Remove the callback by calling the returned function.
func (c *Connection) SubscribeEvent(typ string, cb EventCallback) EventCallbackRemover {
	return c.addSubscriber(typ, cb)
}

// SubscribeToAll subscribes the given callback to all events, with or without type.
// Remove the callback by calling the returned function.
func (c *Connection) SubscribeToAll(cb EventCallback) EventCallbackRemover {
	return c.addSubscriberToAll(cb)
}

func (c *Connection) addSubscriberToAll(cb EventCallback) EventCallbackRemover {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.callbackID
	c.callbacksAll[id] = cb
	c.callbackID++

	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.callbacksAll, id)
	}
}

func (c *Connection) addSubscriber(event string, cb EventCallback) EventCallbackRemover {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.callbacks[event]; !ok {
		c.callbacks[event] = map[int]EventCallback{}
	}

	id := c.callbackID
	c.callbacks[event][id] = cb
	c.callbackID++

	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.callbacks[event], id)
		if len(c.callbacks[event]) == 0 {
			delete(c.callbacks, event)
		}
	}
}

// ConnectionError is the type that wraps all the connection errors that occur.
type ConnectionError struct {
	// The request for which the connection failed.
	Req *http.Request
	// The reason the operation failed.
	Err error
	// The reason why the request failed.
	Reason string
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("request failed: %s: %v", e.Reason, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

func (c *Connection) resetRequest() error {
	if !c.isRetry {
		c.isRetry = true
		return nil
	}
	if err := resetRequestBody(c.request); err != nil {
		return err
	}
	if c.lastEventID == "" {
		c.request.Header.Del("Last-Event-ID")
	} else {
		c.request.Header.Set("Last-Event-ID", c.lastEventID)
	}
	return nil
}

func (c *Connection) dispatch(ev Event) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cbs := c.callbacks[ev.Type]
	cbCount := len(cbs) + len(c.callbacksAll)
	if cbCount == 0 {
		return
	}

	if l := len(ev.Data); l > 0 {
		ev.Data = ev.Data[:l-1]
	}
	ev.LastEventID = c.lastEventID

	for _, cb := range c.callbacks[ev.Type] {
		cb(ev)
	}
	for _, cb := range c.callbacksAll {
		cb(ev)
	}
}

func (c *Connection) read(r io.Reader, setRetry func(time.Duration)) error {
	p := parser.New(r)
	ev, dirty := Event{}, false

	for f := (parser.Field{}); p.Next(&f); {
		switch f.Name { //nolint:exhaustive // Comment fields are not parsed.
		case parser.FieldNameData:
			ev.Data += f.Value + "\n"
			dirty = true
		case parser.FieldNameEvent:
			ev.Type = f.Value
			dirty = true
		case parser.FieldNameID:
			// empty IDs are valid, only IDs that contain the null byte must be ignored:
			// https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation
			if strings.IndexByte(f.Value, 0) != -1 {
				break
			}

			c.lastEventID = f.Value
			dirty = true
		case parser.FieldNameRetry:
			n, err := strconv.ParseInt(f.Value, 10, 64)
			if err != nil {
				break
			}
			if n > 0 {
				setRetry(time.Duration(n) * time.Millisecond)
			}
			dirty = true
		default:
			c.dispatch(ev)
			ev = Event{}
			dirty = false
		}
	}

	err := p.Err()
	if dirty && err == io.EOF { //nolint:errorlint // Our scanner returns io.EOF unwrapped
		c.dispatch(ev)
	}

	return err
}

// Connect sends the request the connection was created with to the server
// and, if successful, it starts receiving events. The caller goroutine
// is blocked until the request's context is done or an error occurs.
//
// If the request's context is cancelled, Connect returns its error.
// Otherwise, if the maximum number or retries is made, the last error
// that occurred is returned. Connect never returns otherwise – either
// the context is cancelled, or it's done retrying.
//
// All errors returned other than the context errors will be wrapped
// inside a *ConnectionError.
func (c *Connection) Connect() error {
	ctx := c.request.Context()
	backoff := c.client.Backoff.new()

	c.request.Header.Set("Accept", "text/event-stream")
	c.request.Header.Set("Connection", "keep-alive")
	c.request.Header.Set("Cache", "no-cache")

	t := time.NewTimer(0)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			shouldRetry, err := c.doConnect(ctx, backoff.reset)
			if !shouldRetry {
				return err
			}

			next, shouldRetry := backoff.next()
			if !shouldRetry {
				return err
			}

			if c.client.OnRetry != nil {
				c.client.OnRetry(err, next)
			}

			t.Reset(next)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Connection) doConnect(ctx context.Context, setRetry func(time.Duration)) (shouldRetry bool, err error) {
	if err := c.resetRequest(); err != nil {
		return false, &ConnectionError{Req: c.request, Reason: "request reset failed", Err: err}
	}

	res, err := c.client.HTTPClient.Do(c.request)
	if err != nil {
		concrete := err.(*url.Error) //nolint:errorlint // We know the concrete type here
		if errors.Is(err, ctx.Err()) {
			return false, concrete.Err
		}
		return true, &ConnectionError{Req: c.request, Reason: "connection to server failed", Err: concrete.Err}
	}
	defer res.Body.Close()

	if err := c.client.ResponseValidator(res); err != nil {
		return false, &ConnectionError{Req: c.request, Reason: "response validation failed", Err: err}
	}

	setRetry(0)

	err = c.read(res.Body, setRetry)
	if errors.Is(err, ctx.Err()) {
		return false, err
	}

	return true, &ConnectionError{Req: c.request, Reason: "connection to server lost", Err: err}
}

// ErrNoGetBody is a sentinel error returned when the connection cannot be reattempted
// due to GetBody not existing on the original request.
var ErrNoGetBody = errors.New("the GetBody function doesn't exist on the request")

func resetRequestBody(r *http.Request) error {
	if r.Body == nil || r.Body == http.NoBody {
		return nil
	}
	if r.GetBody == nil {
		return ErrNoGetBody
	}
	body, err := r.GetBody()
	if err != nil {
		return err
	}
	r.Body = body
	return nil
}
