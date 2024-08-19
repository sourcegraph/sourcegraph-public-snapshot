package libhoney

import (
	"errors"
	"sync"

	"github.com/honeycombio/libhoney-go/transmission"
)

// Client represents an object that can create new builders and events and send
// them somewhere. It maintains its own sending queue for events, distinct from
// both the package-level libhoney queue and any other client. Clients should be
// created with NewClient(config). A manually created Client{} will function as
// a nil output and drop everything handed to it (so can be safely used in dev
// and tests). For more complete testing you can create a Client with a
// MockOutput transmission then inspect the events it would have sent.
type Client struct {
	transmission transmission.Sender
	logger       Logger
	builder      *Builder

	oneTx      sync.Once
	oneLogger  sync.Once
	oneBuilder sync.Once
}

// ClientConfig is a subset of the global libhoney config that focuses on the
// configuration of the client itself. The other config options are specific to
// a given transmission Sender and should be specified there if the defaults
// need to be overridden.
type ClientConfig struct {
	// APIKey is the Honeycomb authentication token. If it is specified during
	// libhoney initialization, it will be used as the default API key for all
	// events. If absent, API key must be explicitly set on a builder or
	// event. Find your team's API keys at https://ui.honeycomb.io/account
	APIKey string

	// Dataset is the name of the Honeycomb dataset to which to send these events.
	// If it is specified during libhoney initialization, it will be used as the
	// default dataset for all events. If absent, dataset must be explicitly set
	// on a builder or event.
	Dataset string

	// SampleRate is the rate at which to sample this event. Default is 1,
	// meaning no sampling. If you want to send one event out of every 250 times
	// Send() is called, you would specify 250 here.
	SampleRate uint

	// APIHost is the hostname for the Honeycomb API server to which to send this
	// event. default: https://api.honeycomb.io/
	APIHost string

	// Transmission allows you to override what happens to events after you call
	// Send() on them. By default, events are asynchronously sent to the
	// Honeycomb API. You can use the MockOutput included in this package in
	// unit tests, or use the transmission.WriterSender to write events to
	// STDOUT or to a file when developing locally.
	Transmission transmission.Sender

	// Logger defaults to nil and the SDK is silent. If you supply a logger here
	// (or set it to &DefaultLogger{}), some debugging output will be emitted.
	// Intended for human consumption during development to understand what the
	// SDK is doing and diagnose trouble emitting events.
	Logger Logger
}

// NewClient creates a Client with defaults correctly set
func NewClient(conf ClientConfig) (*Client, error) {
	if conf.SampleRate == 0 {
		conf.SampleRate = defaultSampleRate
	}
	if conf.APIHost == "" {
		conf.APIHost = defaultAPIHost
	}
	if conf.Dataset == "" {
		conf.Dataset = defaultDataset
	}

	c := &Client{
		logger: conf.Logger,
	}
	c.ensureLogger()

	if conf.Transmission == nil {
		c.transmission = &transmission.Honeycomb{
			MaxBatchSize:         DefaultMaxBatchSize,
			BatchTimeout:         DefaultBatchTimeout,
			MaxConcurrentBatches: DefaultMaxConcurrentBatches,
			PendingWorkCapacity:  DefaultPendingWorkCapacity,
			UserAgentAddition:    UserAgentAddition,
			Logger:               c.logger,
			Metrics:              sd,
		}
	} else {
		c.transmission = conf.Transmission
	}
	if err := c.transmission.Start(); err != nil {
		c.logger.Printf("transmission client failed to start: %s", err.Error())
		return nil, err
	}

	c.builder = &Builder{
		WriteKey:   conf.APIKey,
		Dataset:    conf.Dataset,
		SampleRate: conf.SampleRate,
		APIHost:    conf.APIHost,
		dynFields:  make([]dynamicField, 0, 0),
		fieldHolder: fieldHolder{
			data: make(map[string]interface{}),
		},
		client: c,
	}

	return c, nil
}

func (c *Client) ensureTransmission() {
	c.oneTx.Do(func() {
		if c.transmission == nil {
			c.transmission = &transmission.DiscardSender{}
			c.transmission.Start()
		}
	})
}

func (c *Client) ensureLogger() {
	c.oneLogger.Do(func() {
		if c.logger == nil {
			c.logger = &nullLogger{}
		}
	})
}

func (c *Client) ensureBuilder() {
	c.oneBuilder.Do(func() {
		if c.builder == nil {
			c.builder = &Builder{
				SampleRate: 1,
				dynFields:  make([]dynamicField, 0, 0),
				fieldHolder: fieldHolder{
					data: make(map[string]interface{}),
				},
				client: c,
			}
		}
	})
}

// Close waits for all in-flight messages to be sent. You should
// call Close() before app termination.
func (c *Client) Close() {
	c.ensureLogger()
	c.logger.Printf("closing libhoney client")
	if c.transmission != nil {
		c.transmission.Stop()
	}
}

// Flush closes and reopens the Output interface, ensuring events
// are sent without waiting on the batch to be sent asyncronously.
// Generally, it is more efficient to rely on asyncronous batches than to
// call Flush, but certain scenarios may require Flush if asynchronous sends
// are not guaranteed to run (i.e. running in AWS Lambda)
func (c *Client) Flush() {
	c.ensureLogger()
	c.logger.Printf("flushing libhoney client")
	if c.transmission != nil {
		if err := c.transmission.Flush(); err != nil {
			c.logger.Printf("unable to flush: %v", err)
		}
	}
}

// TxResponses returns the channel from which the caller can read the responses
// to sent events.
func (c *Client) TxResponses() chan transmission.Response {
	c.ensureTransmission()
	return c.transmission.TxResponses()
}

// AddDynamicField takes a field name and a function that will generate values
// for that metric. The function is called once every time a NewEvent() is
// created and added as a field (with name as the key) to the newly created
// event.
func (c *Client) AddDynamicField(name string, fn func() interface{}) error {
	c.ensureTransmission()
	c.ensureBuilder()
	return c.builder.AddDynamicField(name, fn)
}

// AddField adds a Field to the Client's scope. This metric will be inherited by
// all builders and events.
func (c *Client) AddField(name string, val interface{}) {
	c.ensureTransmission()
	c.ensureBuilder()
	c.builder.AddField(name, val)
}

// Add adds its data to the Client's scope. It adds all fields in a struct or
// all keys in a map as individual Fields. These metrics will be inherited by
// all builders and events.
func (c *Client) Add(data interface{}) error {
	c.ensureTransmission()
	c.ensureBuilder()
	return c.builder.Add(data)
}

// NewEvent creates a new event prepopulated with any Fields present in the
// Client's scope.
func (c *Client) NewEvent() *Event {
	c.ensureTransmission()
	c.ensureBuilder()
	return c.builder.NewEvent()
}

// NewBuilder creates a new event builder. The builder inherits any Dynamic or
// Static Fields present in the Client's scope.
func (c *Client) NewBuilder() *Builder {
	c.ensureTransmission()
	c.ensureBuilder()
	return c.builder.Clone()
}

// sendResponse sends a dropped event response down the response channel
func (c *Client) sendDroppedResponse(e *Event, message string) {
	c.ensureTransmission()
	r := transmission.Response{
		Err:      errors.New(message),
		Metadata: e.Metadata,
	}
	c.transmission.SendResponse(r)

}
