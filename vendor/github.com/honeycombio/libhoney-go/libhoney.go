// Copyright 2016 Honeycomb, Hound Technology, Inc. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package libhoney

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/honeycombio/libhoney-go/transmission"
	statsd "gopkg.in/alexcesaro/statsd.v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	transmission.Version = version
}

const (
	defaultSampleRate = 1
	defaultAPIHost    = "https://api.honeycomb.io/"
	defaultDataset    = "libhoney-go dataset"
	version           = "1.15.8"

	// DefaultMaxBatchSize how many events to collect in a batch
	DefaultMaxBatchSize = 50
	// DefaultBatchTimeout how frequently to send unfilled batches
	DefaultBatchTimeout = 100 * time.Millisecond
	// DefaultMaxConcurrentBatches how many batches to maintain in parallel
	DefaultMaxConcurrentBatches = 80
	// DefaultPendingWorkCapacity how many events to queue up for busy batches
	DefaultPendingWorkCapacity = 10000
)

var (
	ptrKinds = []reflect.Kind{reflect.Ptr, reflect.Slice, reflect.Map}
)

// globals to support default/singleton-like behavior
var (
	// singleton-like client used if you use package-level functions
	dc = &Client{}

	// responses is the interim channel to avoid breaking the API while
	// switching types to transmission.Response
	transitionResponses chan Response

	// oneResp protects the transitional responses channel from racing on
	// creation if multiple goroutines ask for the responses channel
	oneResp sync.Once
)

// default is a mute statsd; intended to be overridden
var sd, _ = statsd.New(statsd.Mute(true), statsd.Prefix("libhoney"))

// UserAgentAddition is a variable set at compile time via -ldflags to allow you
// to augment the "User-Agent" header that libhoney sends along with each event.
// The default User-Agent is "libhoney-go/<version>". If you set this variable, its
// contents will be appended to the User-Agent string, separated by a space. The
// expected format is product-name/version, eg "myapp/1.0"
var UserAgentAddition string

// Config specifies settings for initializing the library.
type Config struct {

	// APIKey is the Honeycomb authentication token. If it is specified during
	// libhoney initialization, it will be used as the default API key for all
	// events. If absent, API key must be explicitly set on a builder or
	// event. Find your team's API keys at https://ui.honeycomb.io/account
	APIKey string

	// WriteKey is the deprecated name for the Honeycomb authentication token.
	//
	// Deprecated: Use APIKey instead. If both are set, APIKey takes precedence.
	WriteKey string

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

	// BlockOnSend determines if libhoney should block or drop packets that exceed
	// the size of the send channel (set by PendingWorkCapacity). Defaults to
	// False - events overflowing the send channel will be dropped.
	BlockOnSend bool

	// BlockOnResponse determines if libhoney should block trying to hand
	// responses back to the caller. If this is true and there is nothing reading
	// from the Responses channel, it will fill up and prevent events from being
	// sent to Honeycomb. Defaults to False - if you don't read from the Responses
	// channel it will be ok.
	BlockOnResponse bool

	// Output is the deprecated method of manipulating how libhoney sends
	// events.
	//
	// Deprecated: Please use Transmission instead.
	Output Output

	// Transmission allows you to override what happens to events after you call
	// Send() on them. By default, events are asynchronously sent to the
	// Honeycomb API. You can use the MockOutput included in this package in
	// unit tests, or use the transmission.WriterSender to write events to
	// STDOUT or to a file when developing locally.
	Transmission transmission.Sender

	// Configuration for the underlying sender. It is safe (and recommended) to
	// leave these values at their defaults. You cannot change these values
	// after calling Init()
	MaxBatchSize         uint          // how many events to collect into a batch before sending. Overrides DefaultMaxBatchSize.
	SendFrequency        time.Duration // how often to send off batches. Overrides DefaultBatchTimeout.
	MaxConcurrentBatches uint          // how many batches can be inflight simultaneously. Overrides DefaultMaxConcurrentBatches.
	PendingWorkCapacity  uint          // how many events to allow to pile up. Overrides DefaultPendingWorkCapacity

	// Deprecated: Transport is deprecated and should not be used. To set the HTTP Transport
	// set the Transport elements on the Transmission Sender instead.
	Transport http.RoundTripper

	// Logger defaults to nil and the SDK is silent. If you supply a logger here
	// (or set it to &DefaultLogger{}), some debugging output will be emitted.
	// Intended for human consumption during development to understand what the
	// SDK is doing and diagnose trouble emitting events.
	Logger Logger
}

// Init is called on app initialization and passed a Config struct, which
// configures default behavior. Use of package-level functions (e.g. SendNow())
// require that WriteKey and Dataset are defined.
//
// Otherwise, if WriteKey and DataSet are absent or a Config is not provided,
// they may be specified later, either on a Builder or an Event. WriteKey,
// Dataset, SampleRate, and APIHost can all be overridden on a per-Builder or
// per-Event basis.
//
// Make sure to call Close() to flush buffers.
func Init(conf Config) error {

	// populate a client config to spin up the default package-level Client
	clientConf := ClientConfig{}

	// Use whichever one is set, but APIKey wins if both are set.
	switch {
	case conf.APIKey != "":
		clientConf.APIKey = conf.APIKey
	case conf.WriteKey != "":
		clientConf.APIKey = conf.WriteKey
	default:
	}

	clientConf.Dataset = conf.Dataset
	clientConf.SampleRate = conf.SampleRate
	clientConf.APIHost = conf.APIHost

	// set up default Logger because we're going to use it for the transmission
	if conf.Logger == nil {
		conf.Logger = &nullLogger{}
	}
	clientConf.Logger = conf.Logger

	// set up defaults for the Transmission
	if conf.MaxBatchSize == 0 {
		conf.MaxBatchSize = DefaultMaxBatchSize
	}
	if conf.SendFrequency == 0 {
		conf.SendFrequency = DefaultBatchTimeout
	}
	if conf.MaxConcurrentBatches == 0 {
		conf.MaxConcurrentBatches = DefaultMaxConcurrentBatches
	}
	if conf.PendingWorkCapacity == 0 {
		conf.PendingWorkCapacity = DefaultPendingWorkCapacity
	}

	// If both transmission and output are set, use transmission. If only one is
	// set, use it. If neither is set, use the Honeycomb transmission
	var t transmission.Sender
	switch {
	case conf.Transmission != nil:
		t = conf.Transmission
	case conf.Output != nil:
		t = &transitionOutput{
			Output:          conf.Output,
			blockOnResponse: conf.BlockOnResponse,
			responses:       make(chan transmission.Response, 2*conf.PendingWorkCapacity),
		}
	default:
		t = &transmission.Honeycomb{
			MaxBatchSize:         conf.MaxBatchSize,
			BatchTimeout:         conf.SendFrequency,
			MaxConcurrentBatches: conf.MaxConcurrentBatches,
			PendingWorkCapacity:  conf.PendingWorkCapacity,
			BlockOnSend:          conf.BlockOnSend,
			BlockOnResponse:      conf.BlockOnResponse,
			Transport:            conf.Transport,
			UserAgentAddition:    UserAgentAddition,
			Logger:               clientConf.Logger,
			Metrics:              sd,
		}
	}
	clientConf.Transmission = t
	var err error
	dc, err = NewClient(clientConf)
	return err
}

// Output was responsible for handling events after Send() is called. Implementations
// of Add() must be safe for concurrent calls.
//
// Deprecated: Output is deprecated; use Transmission instead.
type Output interface {
	Add(ev *Event)
	Start() error
	Stop() error
}

// transitionOutput allows us to use an Output as the transmission.Sender needed
// by the Client by adding the additional methods required to implement the
// Sender interface and embedding the original Output to handle its capabilities
type transitionOutput struct {
	Output
	blockOnResponse bool
	responses       chan transmission.Response
}

func (to *transitionOutput) Add(ev *transmission.Event) {
	origEvent := &Event{
		APIHost:     ev.APIHost,
		WriteKey:    ev.APIKey,
		Dataset:     ev.Dataset,
		SampleRate:  ev.SampleRate,
		Timestamp:   ev.Timestamp,
		Metadata:    ev.Metadata,
		fieldHolder: fieldHolder{data: ev.Data},
	}
	to.Output.Add(origEvent)
}

func (to *transitionOutput) Flush() error {
	if err := to.Stop(); err != nil {
		return err
	}
	return to.Stop()
}

func (to *transitionOutput) TxResponses() chan transmission.Response {
	return to.responses
}
func (to *transitionOutput) SendResponse(r transmission.Response) bool {
	if to.blockOnResponse {
		to.responses <- r
	} else {
		select {
		case to.responses <- r:
		default:
			return true
		}
	}
	return false
}

// VerifyWriteKey is the deprecated call to validate a Honeycomb API key.
//
// Deprecated: Please use VerifyAPIKey instead.
func VerifyWriteKey(config Config) (team string, err error) {
	return VerifyAPIKey(config)
}

// VerifyAPIKey calls out to the Honeycomb API to validate the API key so we can
// exit immediately if desired instead of happily sending events that are all
// rejected.
func VerifyAPIKey(config Config) (team string, err error) {
	dc.ensureLogger()
	defer func() { dc.logger.Printf("verify write key got back %s with err=%s", team, err) }()
	if config.APIKey == "" {
		if config.WriteKey == "" {
			return team, errors.New("config.APIKey and config.WriteKey are both empty; can't verify empty key")
		}
		config.APIKey = config.WriteKey
	}
	if config.APIHost == "" {
		config.APIHost = defaultAPIHost
	}
	u, err := url.Parse(config.APIHost)
	if err != nil {
		return team, fmt.Errorf("Error parsing API URL: %s", err)
	}
	u.Path = path.Join(u.Path, "1", "team_slug")
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return team, err
	}
	req.Header.Set("User-Agent", UserAgentAddition)
	req.Header.Add("X-Honeycomb-Team", config.APIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return team, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return team, errors.New("Write key provided is invalid")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return team, fmt.Errorf(`Abnormal non-200 response verifying Honeycomb write key: %d
Response body: %s`, resp.StatusCode, string(body))
	}
	ret := map[string]string{}
	if err := json.Unmarshal(body, &ret); err != nil {
		return team, err
	}

	return ret["team_slug"], nil
}

// Deprecated: Response is deprecated; please use transmission.Response instead.
type Response struct {
	transmission.Response
}

// Event is used to hold data that can be sent to Honeycomb. It can also
// specify overrides of the config settings.
type Event struct {
	// WriteKey, if set, overrides whatever is found in Config
	WriteKey string
	// Dataset, if set, overrides whatever is found in Config
	Dataset string
	// SampleRate, if set, overrides whatever is found in Config
	SampleRate uint
	// APIHost, if set, overrides whatever is found in Config
	APIHost string
	// Timestamp, if set, specifies the time for this event. If unset, defaults
	// to Now()
	Timestamp time.Time
	// Metadata is a field for you to add in data that will be handed back to you
	// on the Response object read off the Responses channel. It is not sent to
	// Honeycomb with the event.
	Metadata interface{}

	// fieldHolder contains fields (and methods) common to both events and builders
	fieldHolder

	// client is the Client to use to send events generated from this builder
	client *Client

	// sent is a bool indicating whether the event has been sent.  Once it's
	// been sent, all changes to the event should be ignored - any calls to Add
	// should just return immediately taking no action.
	sent     bool
	sendLock sync.Mutex
}

// Builder is used to create templates for new events, specifying default fields
// and override settings.
type Builder struct {
	// WriteKey, if set, overrides whatever is found in Config
	WriteKey string
	// Dataset, if set, overrides whatever is found in Config
	Dataset string
	// SampleRate, if set, overrides whatever is found in Config
	SampleRate uint
	// APIHost, if set, overrides whatever is found in Config
	APIHost string

	// fieldHolder contains fields (and methods) common to both events and builders
	fieldHolder

	// any dynamic fields to apply to each generated event
	dynFields     []dynamicField
	dynFieldsLock sync.RWMutex

	// client is the Client to use to send events generated from this builder
	client *Client
}

type fieldHolder struct {
	data marshallableMap
	lock sync.RWMutex
}

// Wrapper type for custom JSON serialization: individual values that can't be
// marshalled (or are null pointers) will be skipped, instead of causing
// marshalling to raise an error.

// TODO XMIT stop using this type and do the nil checks on Add instead of on marshal
type marshallableMap map[string]interface{}

func (m marshallableMap) MarshalJSON() ([]byte, error) {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	out := bytes.NewBufferString("{")

	first := true
	for _, k := range keys {
		b, ok := maybeMarshalValue(m[k])
		if ok {
			if first {
				first = false
			} else {
				out.WriteByte(',')
			}

			out.WriteByte('"')
			out.Write([]byte(k))
			out.WriteByte('"')
			out.WriteByte(':')
			out.Write(b)
		}
	}
	out.WriteByte('}')
	return out.Bytes(), nil
}

func maybeMarshalValue(v interface{}) ([]byte, bool) {
	if v == nil {
		return nil, false
	}
	val := reflect.ValueOf(v)
	kind := val.Type().Kind()
	for _, ptrKind := range ptrKinds {
		if kind == ptrKind && val.IsNil() {
			return nil, false
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	return b, true
}

type dynamicField struct {
	name string
	fn   func() interface{}
}

// Close waits for all in-flight messages to be sent. You should
// call Close() before app termination.
func Close() {
	dc.Close()
}

// Flush closes and reopens the Output interface, ensuring events
// are sent without waiting on the batch to be sent asyncronously.
// Generally, it is more efficient to rely on asyncronous batches than to
// call Flush, but certain scenarios may require Flush if asynchronous sends
// are not guaranteed to run (i.e. running in AWS Lambda)
// Flush is not thread safe - use it only when you are sure that no other
// parts of your program are calling Send
func Flush() {
	dc.Flush()
}

// Contrary to its name, SendNow does not block and send data
// immediately, but only enqueues to be sent asynchronously.
// It is equivalent to:
//   ev := libhoney.NewEvent()
//   ev.Add(data)
//   ev.Send()
//
// Deprecated: SendNow is deprecated and may be removed in a future major release.
func SendNow(data interface{}) error {
	dc.ensureLogger()
	ev := NewEvent()
	if err := ev.Add(data); err != nil {
		return err
	}
	err := ev.Send()
	dc.logger.Printf("SendNow enqueued event, err=%v", err)
	return err
}

// Responses returns the channel from which the caller can read the responses
// to sent events.
//
// Deprecated: Responses is deprecated; please use TxResponses instead.
func Responses() chan Response {
	oneResp.Do(func() {
		if transitionResponses == nil {
			txResponses := dc.TxResponses()
			transitionResponses = make(chan Response, cap(txResponses))
			go func() {
				for txResp := range txResponses {
					resp := Response{}
					resp.Response = txResp
					transitionResponses <- resp
				}
				close(transitionResponses)
			}()
		}
	})
	return transitionResponses
}

// TxResponses returns the channel from which the caller can read the responses
// to sent events.
func TxResponses() chan transmission.Response {
	return dc.TxResponses()
}

// AddDynamicField takes a field name and a function that will generate values
// for that metric. The function is called once every time a NewEvent() is
// created and added as a field (with name as the key) to the newly created
// event.
func AddDynamicField(name string, fn func() interface{}) error {
	return dc.AddDynamicField(name, fn)
}

// AddField adds a Field to the global scope. This metric will be inherited by
// all builders and events.
func AddField(name string, val interface{}) {
	dc.AddField(name, val)
}

// Add adds its data to the global scope. It adds all fields in a struct or all
// keys in a map as individual Fields. These metrics will be inherited by all
// builders and events.
func Add(data interface{}) error {
	return dc.Add(data)
}

// NewEvent creates a new event prepopulated with any Fields present in the
// global scope.
func NewEvent() *Event {
	return dc.NewEvent()
}

// NewBuilder creates a new event builder. The builder inherits any
// Dynamic or Static Fields present in the global scope.
func NewBuilder() *Builder {
	return dc.NewBuilder()
}

// AddField adds an individual metric to the event or builder on which it is
// called. Note that if you add a value that cannot be serialized to JSON (eg a
// function or channel), the event will fail to send.
func (f *fieldHolder) AddField(key string, val interface{}) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.data[key] = val
}

// Add adds a complex data type to the event or builder on which it's called.
// For structs, it adds each exported field. For maps, it adds each key/value.
// Add will error on all other types.
func (f *fieldHolder) Add(data interface{}) error {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Struct:
		return f.addStruct(data)
	case reflect.Map:
		return f.addMap(data)
	case reflect.Ptr:
		return f.Add(reflect.ValueOf(data).Elem().Interface())
	}
	return fmt.Errorf(
		"Couldn't add type %s content %+v",
		reflect.TypeOf(data).Kind(), data,
	)
}

func (f *fieldHolder) addStruct(s interface{}) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	// TODO should we handle embedded structs differently from other deep structs?
	sType := reflect.TypeOf(s)
	sVal := reflect.ValueOf(s)
	// Iterate through the fields, adding each.
	for i := 0; i < sType.NumField(); i++ {
		fieldInfo := sType.Field(i)
		if fieldInfo.PkgPath != "" {
			// skipping unexported field in the struct
			continue
		}

		var fName string
		fTag := fieldInfo.Tag.Get("json")
		if fTag != "" {
			if fTag == "-" {
				// skip this field
				continue
			}
			// slice off options
			if idx := strings.Index(fTag, ","); idx != -1 {
				options := fTag[idx:]
				fTag = fTag[:idx]
				if strings.Contains(options, "omitempty") && isEmptyValue(sVal.Field(i)) {
					// skip empty values if omitempty option is set
					continue
				}
			}
			fName = fTag
		} else {
			fName = fieldInfo.Name
		}

		f.data[fName] = sVal.Field(i).Interface()
	}
	return nil
}

func (f *fieldHolder) addMap(m interface{}) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	mVal := reflect.ValueOf(m)
	mKeys := mVal.MapKeys()
	for _, key := range mKeys {
		// get a string representation of key
		var keyStr string
		switch key.Type().Kind() {
		case reflect.String:
			keyStr = key.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
			reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Complex64,
			reflect.Complex128:
			keyStr = fmt.Sprintf("%v", key.Interface())
		default:
			return fmt.Errorf("failed to add map: key type %s unaccepted", key.Type().Kind())
		}
		f.data[keyStr] = mVal.MapIndex(key).Interface()
	}
	return nil
}

// AddFunc takes a function and runs it repeatedly, adding the return values
// as fields.
// The function should return error when it has exhausted its values
func (f *fieldHolder) AddFunc(fn func() (string, interface{}, error)) error {
	for {
		key, rawVal, err := fn()
		if err != nil {
			// fn is done giving us data
			break
		}
		f.AddField(key, rawVal)
	}
	return nil
}

// Fields returns a reference to the map of fields that have been added to an
// event. Caution: it is not safe to manipulate the returned map concurrently
// with calls to AddField, Add or AddFunc.
func (f *fieldHolder) Fields() map[string]interface{} {
	return f.data
}

// returns a human friendly string representation of the fieldHolder
func (f *fieldHolder) String() string {
	return fmt.Sprint(f.data)
}

// mask the add functions on an event so that we can test the sent lock and noop
// if the event has been sent.

// AddField adds an individual metric to the event on which it is called. Note
// that if you add a value that cannot be serialized to JSON (eg a function or
// channel), the event will fail to send.
//
// Adds to an event that happen after it has been sent will return without
// having any effect.
func (e *Event) AddField(key string, val interface{}) {
	e.sendLock.Lock()
	defer e.sendLock.Unlock()
	if e.sent == true {
		return
	}
	e.fieldHolder.AddField(key, val)
}

// Add adds a complex data type to the event on which it's called.
// For structs, it adds each exported field. For maps, it adds each key/value.
// Add will error on all other types.
//
// Adds to an event that happen after it has been sent will return without
// having any effect.
func (e *Event) Add(data interface{}) error {
	e.sendLock.Lock()
	defer e.sendLock.Unlock()
	if e.sent == true {
		return nil
	}
	return e.fieldHolder.Add(data)
}

// AddFunc takes a function and runs it repeatedly, adding the return values
// as fields.
// The function should return error when it has exhausted its values
//
// Adds to an event that happen after it has been sent will return without
// having any effect.
func (e *Event) AddFunc(fn func() (string, interface{}, error)) error {
	e.sendLock.Lock()
	defer e.sendLock.Unlock()
	if e.sent == true {
		return nil
	}
	return e.fieldHolder.AddFunc(fn)
}

// Send dispatches the event to be sent to Honeycomb, sampling if necessary.
//
// If you have sampling enabled
// (i.e. SampleRate >1), Send will only actually transmit data with a
// probability of 1/SampleRate. No error is returned whether or not traffic
// is sampled, however, the Response sent down the response channel will
// indicate the event was sampled in the errors Err field.
//
// Send inherits the values of required fields from Config. If any required
// fields are specified in neither Config nor the Event, Send will return an
// error.  Required fields are APIHost, WriteKey, and Dataset. Values specified
// in an Event override Config.
//
// Once you Send an event, any addition calls to add data to that event will
// return without doing anything. Once the event is sent, it becomes immutable.
func (e *Event) Send() error {
	if e.client == nil {
		e.client = &Client{}
	}
	e.client.ensureLogger()
	if shouldDrop(e.SampleRate) {
		e.client.logger.Printf("dropping event due to sampling")
		sd.Increment("sampled")
		e.client.sendDroppedResponse(e, "event dropped due to sampling")
		return nil
	}
	return e.SendPresampled()
}

// SendPresampled dispatches the event to be sent to Honeycomb.
//
// Sampling is assumed to have already happened. SendPresampled will dispatch
// every event handed to it, and pass along the sample rate. Use this instead of
// Send() when the calling function handles the logic around which events to
// drop when sampling.
//
// SendPresampled inherits the values of required fields from Config. If any
// required fields are specified in neither Config nor the Event, Send will
// return an error.  Required fields are APIHost, WriteKey, and Dataset. Values
// specified in an Event override Config.
//
// Once you Send an event, any addition calls to add data to that event will
// return without doing anything. Once the event is sent, it becomes immutable.
func (e *Event) SendPresampled() (err error) {
	if e.client == nil {
		e.client = &Client{}
	}
	e.client.ensureLogger()
	defer func() {
		if err != nil {
			e.client.logger.Printf("Failed to send event. err: %s, event: %+v", err, e)
		} else {
			e.client.logger.Printf("Send enqueued event: %+v", e)
		}
	}()

	// Lock the sent bool before taking the event lock, to match the order in
	// the Add methods.
	e.sendLock.Lock()
	defer e.sendLock.Unlock()

	e.fieldHolder.lock.RLock()
	defer e.fieldHolder.lock.RUnlock()
	if len(e.data) == 0 {
		return errors.New("No metrics added to event. Won't send empty event.")
	}

	// if client.transmission is transmission.Honeycomb or a pointer to same,
	// then we should verify that APIHost and WriteKey are set. For
	// non-Honeycomb based Sender implementations (eg STDOUT) it's totally
	// possible to send events without an API key etc

	senderType := reflect.TypeOf(e.client.transmission).String()
	isHoneycombSender := strings.HasSuffix(senderType, "transmission.Honeycomb")
	isMockSender := strings.HasSuffix(senderType, "transmission.MockSender")
	if isHoneycombSender || isMockSender {
		if e.APIHost == "" {
			return errors.New("No APIHost for Honeycomb. Can't send to the Great Unknown.")
		}
		if e.WriteKey == "" {
			return errors.New("No WriteKey specified. Can't send event.")
		}
	}
	if e.Dataset == "" {
		return errors.New("No Dataset for Honeycomb. Can't send datasetless.")
	}

	// Mark the event as sent, no more field changes will be applied.
	e.sent = true

	e.client.ensureTransmission()
	txEvent := &transmission.Event{
		APIHost:    e.APIHost,
		APIKey:     e.WriteKey,
		Dataset:    e.Dataset,
		SampleRate: e.SampleRate,
		Timestamp:  e.Timestamp,
		Metadata:   e.Metadata,
		Data:       e.data,
	}
	e.client.transmission.Add(txEvent)
	return nil
}

// returns a human friendly string representation of the event
func (e *Event) String() string {
	masked := e.WriteKey
	if e.WriteKey != "" && len(e.WriteKey) > 4 {
		len := len(e.WriteKey) - 4
		masked = strings.Repeat("X", len) + e.WriteKey[len:]
	}
	return fmt.Sprintf("{WriteKey:%s Dataset:%s SampleRate:%d APIHost:%s Timestamp:%v fieldHolder:%+v sent:%t}", masked, e.Dataset, e.SampleRate, e.APIHost, e.Timestamp, e.fieldHolder.String(), e.sent)
}

// returns true if the sample should be dropped
func shouldDrop(rate uint) bool {
	if rate <= 1 {
		return false
	}

	return rand.Intn(int(rate)) != 0
}

// AddDynamicField adds a dynamic field to the builder. Any events
// created from this builder will get this metric added.
func (b *Builder) AddDynamicField(name string, fn func() interface{}) error {
	b.dynFieldsLock.Lock()
	defer b.dynFieldsLock.Unlock()
	dynFn := dynamicField{
		name: name,
		fn:   fn,
	}
	b.dynFields = append(b.dynFields, dynFn)
	return nil
}

// Contrary to its name, SendNow does not block and send data
// immediately, but only enqueues to be sent asynchronously.
// It is equivalent to:
//   ev := builder.NewEvent()
//   ev.Add(data)
//   ev.Send()
//
// Deprecated: SendNow is deprecated and may be removed in a future major release.
func (b *Builder) SendNow(data interface{}) error {
	ev := b.NewEvent()
	if err := ev.Add(data); err != nil {
		return err
	}
	err := ev.Send()
	return err
}

// NewEvent creates a new Event prepopulated with fields, dynamic
// field values, and configuration inherited from the builder.
func (b *Builder) NewEvent() *Event {
	e := &Event{
		WriteKey:   b.WriteKey,
		Dataset:    b.Dataset,
		SampleRate: b.SampleRate,
		APIHost:    b.APIHost,
		Timestamp:  time.Now(),
		client:     b.client,
	}
	e.data = make(map[string]interface{})

	b.lock.RLock()
	defer b.lock.RUnlock()
	for k, v := range b.data {
		e.data[k] = v
	}
	// create dynamic metrics
	b.dynFieldsLock.RLock()
	defer b.dynFieldsLock.RUnlock()
	for _, dynField := range b.dynFields {
		e.AddField(dynField.name, dynField.fn())
	}
	return e
}

// Clone creates a new builder that inherits all traits of this builder and
// creates its own scope in which to add additional static and dynamic fields.
func (b *Builder) Clone() *Builder {
	newB := &Builder{
		WriteKey:   b.WriteKey,
		Dataset:    b.Dataset,
		SampleRate: b.SampleRate,
		APIHost:    b.APIHost,
		client:     b.client,
	}
	newB.data = make(map[string]interface{})
	b.lock.RLock()
	defer b.lock.RUnlock()
	for k, v := range b.data {
		newB.data[k] = v
	}
	// copy dynamic metric generators
	b.dynFieldsLock.RLock()
	defer b.dynFieldsLock.RUnlock()
	newB.dynFields = make([]dynamicField, len(b.dynFields))
	copy(newB.dynFields, b.dynFields)
	return newB
}

// Helper lifted from Go stdlib encoding/json
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
