package sse

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"unicode"
)

// The ResponseValidator type defines the type of the function
// that checks whether server responses are valid, before starting
// to read events from them. See the Client's documentation for more info.
//
// These errors are considered permanent and thus if the client is configured
// to retry on error no retry is attempted and the error is returned.
type ResponseValidator func(*http.Response) error

// The Client struct is used to initialize new connections to different servers.
// It is safe for concurrent use.
//
// After connections are created, the Connect method must be called to start
// receiving events.
type Client struct {
	// The HTTP client to be used. Defaults to http.DefaultClient.
	HTTPClient *http.Client
	// A callback that's executed whenever a reconnection attempt starts.
	// It receives the error that caused the retry and the reconnection time.
	OnRetry func(error, time.Duration)
	// A function to check if the response from the server is valid.
	// Defaults to a function that checks the response's status code is 200
	// and the content type is text/event-stream.
	//
	// If the error type returned has a Temporary or a Timeout method,
	// they will be used to determine whether to reattempt the connection.
	// Otherwise, the error will be considered permanent and no reconnections
	// will be attempted.
	ResponseValidator ResponseValidator
	// Backoff configures the backoff strategy. See the documentation of
	// each field for more information.
	Backoff Backoff
}

// Backoff configures the reconnection strategy of a Connection.
type Backoff struct {
	// The initial wait time before a reconnection is attempted.
	// Must be >0. Defaults to 500ms.
	InitialInterval time.Duration
	// How much should the reconnection time grow on subsequent attempts.
	// Must be >=1; 1 = constant interval. Defaults to 1.5.
	Multiplier float64
	// How much does the reconnection time vary relative to the base value.
	// This is useful to prevent multiple clients to reconnect at the exact
	// same time, as it makes the wait times distinct.
	// Must be in range (0, 1); -1 = no randomization. Defaults to 0.5.
	Jitter float64
	// How much can the wait time grow.
	// If <=0 = the wait time can infinitely grow. Defaults to infinite growth.
	MaxInterval time.Duration
	// How much time can retries be attempted.
	// For example, if this is 5 seconds, after 5 seconds the client
	// will stop retrying.
	// If <=0 = no limit. Defaults to no limit.
	MaxElapsedTime time.Duration
	// How many retries are allowed.
	// <0 = no retries, 0 = infinite. Defaults to infinite retries.
	MaxRetries int
}

// NewConnection initializes and configures a connection. On connect, the given
// request is sent and if successful the connection starts receiving messages.
// Use the request's context to stop the connection.
//
// If the request has a body, it is necessary to provide a GetBody function in order
// for the connection to be reattempted, in case of an error. Using readers
// such as bytes.Reader, strings.Reader or bytes.Buffer when creating a request
// using http.NewRequestWithContext will ensure this function is present on the request.
func (c *Client) NewConnection(r *http.Request) *Connection {
	if r == nil {
		panic("go-sse.client.NewConnection: request cannot be nil")
	}

	mergeDefaults(c)

	conn := &Connection{
		client:       *c,                   // we clone the client so the config cannot be modified from outside
		request:      r.Clone(r.Context()), // we clone the request so its fields cannot be modified from outside
		callbacks:    map[string]map[int]EventCallback{},
		callbacksAll: map[int]EventCallback{},
	}

	return conn
}

// DefaultValidator is the default client response validation function. As per the spec,
// It checks the content type to be text/event-stream and the response status code to be 200 OK.
//
// If this validator fails, errors are considered permanent. No retry attempts are made.
//
// See https://html.spec.whatwg.org/multipage/server-sent-events.html#sse-processing-model.
var DefaultValidator ResponseValidator = func(r *http.Response) error {
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status code %d %s, received %d %s", http.StatusOK, http.StatusText(http.StatusOK), r.StatusCode, http.StatusText(r.StatusCode))
	}
	cts := r.Header.Get("Content-Type")
	ct := contentType(cts)
	if expected := "text/event-stream"; ct != expected {
		return fmt.Errorf("expected content type to have %q, received %q", expected, cts)
	}
	return nil
}

// NoopValidator is a client response validator function that treats all responses as valid.
var NoopValidator ResponseValidator = func(_ *http.Response) error {
	return nil
}

// DefaultClient is the client that is used when creating a new connection using the NewConnection function.
// Unset properties on new clients are replaced with the ones set for the default client.
var DefaultClient = &Client{
	HTTPClient:        http.DefaultClient,
	ResponseValidator: DefaultValidator,
	Backoff: Backoff{
		InitialInterval: time.Millisecond * 500,
		Multiplier:      1.5,
		Jitter:          0.5,
	},
}

// NewConnection creates a connection using the default client.
func NewConnection(r *http.Request) *Connection {
	return DefaultClient.NewConnection(r)
}

func mergeDefaults(c *Client) {
	if c.HTTPClient == nil {
		c.HTTPClient = DefaultClient.HTTPClient
	}
	if c.Backoff.InitialInterval <= 0 {
		c.Backoff.InitialInterval = DefaultClient.Backoff.InitialInterval
	}
	if c.Backoff.Multiplier < 1 {
		c.Backoff.Multiplier = DefaultClient.Backoff.Multiplier
	}
	if c.Backoff.Jitter <= 0 || c.Backoff.Jitter >= 1 {
		c.Backoff.Jitter = DefaultClient.Backoff.Jitter
	}
	if c.ResponseValidator == nil {
		c.ResponseValidator = DefaultClient.ResponseValidator
	}
}

func contentType(header string) string {
	cts := strings.FieldsFunc(header, func(r rune) bool {
		return unicode.IsSpace(r) || r == ';' || r == ','
	})
	if len(cts) == 0 {
		return ""
	}
	return strings.ToLower(cts[0])
}

type backoffController struct {
	start      time.Time
	rng        *rand.Rand
	b          *Backoff
	interval   time.Duration
	numRetries int
}

func (b *Backoff) new() backoffController {
	now := time.Now()
	return backoffController{
		start:      now,
		rng:        rand.New(rand.NewSource(now.UnixNano())),
		b:          b,
		interval:   b.InitialInterval,
		numRetries: 0,
	}
}

// reset the backoff to the initial state, i.e. as if no retries have occurred.
// If newInterval is greater than 0, the initial interval is changed to it.
func (c *backoffController) reset(newInterval time.Duration) {
	if newInterval > 0 {
		c.interval = newInterval
	} else {
		c.interval = c.b.InitialInterval
	}
	c.numRetries = 0
	c.start = time.Now()
}

func (c *backoffController) next() (interval time.Duration, shouldRetry bool) {
	if c.b.MaxRetries < 0 || (c.b.MaxRetries > 0 && c.numRetries == c.b.MaxRetries) {
		return 0, false
	}

	c.numRetries++
	elapsed := time.Since(c.start)
	next := nextInterval(c.b.Jitter, c.rng, c.interval)
	c.interval = growInterval(c.interval, c.b.MaxInterval, c.b.Multiplier)

	if c.b.MaxElapsedTime > 0 && elapsed+next > c.b.MaxElapsedTime {
		return 0, false
	}

	return next, true
}

func nextInterval(jitter float64, rng *rand.Rand, current time.Duration) time.Duration {
	if jitter == -1 {
		return current
	}

	delta := jitter * float64(current)
	minInterval := float64(current) - delta
	maxInterval := float64(current) + delta

	return time.Duration(minInterval + (rng.Float64() * (maxInterval - minInterval + 1)))
}

func growInterval(current, maxInterval time.Duration, mul float64) time.Duration {
	if maxInterval > 0 && float64(current) >= float64(maxInterval)/mul {
		return maxInterval
	}
	return time.Duration(float64(current) * mul)
}
