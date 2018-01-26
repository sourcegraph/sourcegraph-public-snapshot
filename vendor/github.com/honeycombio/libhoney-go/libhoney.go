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

	"gopkg.in/alexcesaro/statsd.v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	defaultSampleRate = 1
	defaultAPIHost    = "https://api.honeycomb.io/"
	version           = "1.5.0"

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
	tx     Output
	txOnce sync.Once

	blockOnResponses = false
	sd, _            = statsd.New(statsd.Mute(true)) // init working default, to be overridden
	responses        = make(chan Response, 2*DefaultPendingWorkCapacity)
	defaultBuilder   = &Builder{
		APIHost:    defaultAPIHost,
		SampleRate: defaultSampleRate,
		dynFields:  make([]dynamicField, 0, 0),
		fieldHolder: fieldHolder{
			data: make(map[string]interface{}),
		},
	}
)

// UserAgentAddition is a variable set at compile time via -ldflags to allow you
// to augment the "User-Agent" header that libhoney sends along with each event.
// The default User-Agent is "libhoney-go/<version>". If you set this variable, its
// contents will be appended to the User-Agent string, separated by a space. The
// expected format is product-name/version, eg "myapp/1.0"
var UserAgentAddition string

// Config specifies settings for initializing the library.
type Config struct {

	// WriteKey is the Honeycomb authentication token. If it is specified during
	// libhoney initialization, it will be used as the default write key for all
	// events. If absent, write key must be explicitly set on a builder or
	// event. Find your team write key at https://ui.honeycomb.io/account
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

	// TODO add logger in an agnostic way

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

	// Output allows you to override what happens to events after you call
	// Send() on them. By default, events are asynchronously sent to the
	// Honeycomb API. You can use the MockOutput included in this package in
	// unit tests, or use the WriterOutput to write events to STDOUT or to a
	// file when developing locally.
	Output Output

	// Configuration for the underlying sender. It is safe (and recommended) to
	// leave these values at their defaults. You cannot change these values
	// after calling Init()
	MaxBatchSize         uint          // how many events to collect into a batch before sending. Overrides DefaultMaxBatchSize.
	SendFrequency        time.Duration // how often to send off batches. Overrides DefaultBatchTimeout.
	MaxConcurrentBatches uint          // how many batches can be inflight simultaneously. Overrides DefaultMaxConcurrentBatches.
	PendingWorkCapacity  uint          // how many events to allow to pile up. Overrides DefaultPendingWorkCapacity

	// Transport can be provided to the http.Client attempting to talk to
	// Honeycomb servers. Intended for use in tests in order to assert on
	// expected behavior.
	Transport http.RoundTripper
}

// VerifyWriteKey calls out to the Honeycomb API to validate the write key, so
// we can exit immediately if desired instead of happily sending events that
// are all rejected.
func VerifyWriteKey(config Config) (string, error) {
	if config.WriteKey == "" {
		return "", errors.New("Write key is empty")
	}
	if config.APIHost == "" {
		config.APIHost = defaultAPIHost
	}
	u, err := url.Parse(config.APIHost)
	if err != nil {
		return "", fmt.Errorf("Error parsing API URL: %s", err)
	}
	u.Path = path.Join(u.Path, "1", "team_slug")
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgentAddition)
	req.Header.Add("X-Honeycomb-Team", config.WriteKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("Write key provided is invalid")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(`Abnormal non-200 response verifying Honeycomb write key: %d
Response body: %s`, resp.StatusCode, string(body))
	}
	ret := map[string]string{}
	if err := json.Unmarshal(body, &ret); err != nil {
		return "", err
	}

	return ret["team_slug"], nil
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
}

// Marshaling an Event for batching up to the Honeycomb servers. Omits fields
// that aren't specific to this particular event, and allows for behavior like
// omitempty'ing a zero'ed out time.Time.
func (e *Event) MarshalJSON() ([]byte, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	tPointer := &(e.Timestamp)
	if e.Timestamp.IsZero() {
		tPointer = nil
	}

	// don't include sample rate if it's 1; this is the default
	sampleRate := e.SampleRate
	if sampleRate == 1 {
		sampleRate = 0
	}

	return json.Marshal(struct {
		Data       marshallableMap `json:"data"`
		SampleRate uint            `json:"samplerate,omitempty"`
		Timestamp  *time.Time      `json:"time,omitempty"`
	}{e.data, sampleRate, tPointer})
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
}

type fieldHolder struct {
	data marshallableMap
	lock sync.RWMutex
}

// Wrapper type for custom JSON serialization: individual values that can't be
// marshalled (or are null pointers) will be skipped, instead of causing
// marshalling to raise an error.
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
func Init(config Config) error {
	// Default sample rate should be 1. 0 is invalid.
	if config.SampleRate == 0 {
		config.SampleRate = defaultSampleRate
	}
	if config.APIHost == "" {
		config.APIHost = defaultAPIHost
	}
	if config.MaxBatchSize == 0 {
		config.MaxBatchSize = DefaultMaxBatchSize
	}
	if config.SendFrequency == 0 {
		config.SendFrequency = DefaultBatchTimeout
	}
	if config.MaxConcurrentBatches == 0 {
		config.MaxConcurrentBatches = DefaultMaxConcurrentBatches
	}
	if config.PendingWorkCapacity == 0 {
		config.PendingWorkCapacity = DefaultPendingWorkCapacity
	}

	blockOnResponses = config.BlockOnResponse

	if config.Output == nil {
		// reset the global transmission
		tx = &txDefaultClient{
			maxBatchSize:         config.MaxBatchSize,
			batchTimeout:         config.SendFrequency,
			maxConcurrentBatches: config.MaxConcurrentBatches,
			pendingWorkCapacity:  config.PendingWorkCapacity,
			blockOnSend:          config.BlockOnSend,
			blockOnResponses:     config.BlockOnResponse,
			transport:            config.Transport,
		}
	} else {
		tx = config.Output
	}
	if err := tx.Start(); err != nil {
		return err
	}

	sd, _ = statsd.New(statsd.Prefix("libhoney"))
	responses = make(chan Response, config.PendingWorkCapacity*2)

	defaultBuilder = &Builder{
		WriteKey:   config.WriteKey,
		Dataset:    config.Dataset,
		SampleRate: config.SampleRate,
		APIHost:    config.APIHost,
		dynFields:  make([]dynamicField, 0, 0),
		fieldHolder: fieldHolder{
			data: make(map[string]interface{}),
		},
	}

	return nil
}

// Close waits for all in-flight messages to be sent. You should
// call Close() before app termination.
func Close() {
	tx.Stop()
	close(responses)
}

// SendNow is a shortcut to create an event, add data, and send the event.
func SendNow(data interface{}) error {
	ev := NewEvent()
	if err := ev.Add(data); err != nil {
		return err
	}
	if err := ev.Send(); err != nil {
		return err
	}
	return nil
}

// Responses returns the channel from which the caller can read the responses
// to sent events.
func Responses() chan Response {
	return responses
}

// AddDynamicField takes a field name and a function that will generate values
// for that metric. The function is called once every time a NewEvent() is
// created and added as a field (with name as the key) to the newly created
// event.
func AddDynamicField(name string, fn func() interface{}) error {
	return defaultBuilder.AddDynamicField(name, fn)
}

// AddField adds a Field to the global scope. This metric will be inherited by
// all builders and events.
func AddField(name string, val interface{}) {
	defaultBuilder.AddField(name, val)
}

// Add adds its data to the global scope. It adds all fields in a struct or all
// keys in a map as individual Fields. These metrics will be inherited by all
// builders and events.
func Add(data interface{}) error {
	return defaultBuilder.Add(data)
}

// NewEvent creates a new event prepopulated with any Fields present in the
// global scope.
func NewEvent() *Event {
	return defaultBuilder.NewEvent()
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
func (e *Event) Send() error {
	if shouldDrop(e.SampleRate) {
		sd.Increment("sampled")
		sendDroppedResponse(e, "event dropped due to sampling")
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
func (e *Event) SendPresampled() error {
	e.lock.RLock()
	defer e.lock.RUnlock()
	if len(e.data) == 0 {
		return errors.New("No metrics added to event. Won't send empty event.")
	}
	if e.APIHost == "" {
		return errors.New("No APIHost for Honeycomb. Can't send to the Great Unknown.")
	}
	if e.WriteKey == "" {
		return errors.New("No WriteKey specified. Can't send event.")
	}
	if e.Dataset == "" {
		return errors.New("No Dataset for Honeycomb. Can't send datasetless.")
	}

	txOnce.Do(func() {
		if tx == nil {
			tx = &txDefaultClient{
				maxBatchSize:         DefaultMaxBatchSize,
				batchTimeout:         DefaultBatchTimeout,
				maxConcurrentBatches: DefaultMaxConcurrentBatches,
				pendingWorkCapacity:  DefaultPendingWorkCapacity,
			}
			tx.Start()
		}
	})

	tx.Add(e)
	return nil
}

// sendResponse sends a dropped event response down the response channel
func sendDroppedResponse(e *Event, message string) {
	r := Response{
		Err:      errors.New(message),
		Metadata: e.Metadata,
	}
	if blockOnResponses {
		responses <- r
	} else {
		select {
		case responses <- r:
		default:
		}
	}
}

// returns true if the sample should be dropped
func shouldDrop(rate uint) bool {
	return rand.Intn(int(rate)) != 0
}

// NewBuilder creates a new event builder. The builder inherits any
// Dynamic or Static Fields present in the global scope.
func NewBuilder() *Builder {
	return defaultBuilder.Clone()
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

// SendNow is a shortcut to create an event from this builder, add data, and
// send the event.
func (b *Builder) SendNow(data interface{}) error {
	ev := b.NewEvent()
	if err := ev.Add(data); err != nil {
		return err
	}
	if err := ev.Send(); err != nil {
		return err
	}
	return nil
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
		dynFields:  make([]dynamicField, 0, len(b.dynFields)),
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
	for _, dynFd := range b.dynFields {
		newB.dynFields = append(newB.dynFields, dynFd)
	}
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
