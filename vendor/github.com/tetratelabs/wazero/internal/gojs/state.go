package gojs

import (
	"context"
	"fmt"
	"math"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/config"
	"github.com/tetratelabs/wazero/internal/gojs/goos"
	"github.com/tetratelabs/wazero/internal/gojs/values"
)

func NewState(config *config.Config) *State {
	return &State{
		config:                 config,
		values:                 values.NewValues(),
		valueGlobal:            newJsGlobal(config),
		_nextCallbackTimeoutID: 1,
		_scheduledTimeouts:     map[uint32]chan bool{},
	}
}

// StateKey is a context.Context Value key. The value must be a state pointer.
type StateKey struct{}

func getState(ctx context.Context) *State {
	return ctx.Value(StateKey{}).(*State)
}

// GetLastEventArgs implements goos.GetLastEventArgs
func GetLastEventArgs(ctx context.Context) []interface{} {
	if ls := ctx.Value(StateKey{}).(*State)._lastEvent; ls != nil {
		if args := ls.args; args != nil {
			return args.slice
		}
	}
	return nil
}

type event struct {
	// id is the funcWrapper.id
	id     uint32
	this   goos.Ref
	args   *objectArray
	result interface{}
}

// Get implements the same method as documented on goos.GetFunction
func (e *event) Get(propertyKey string) interface{} {
	switch propertyKey {
	case "id":
		return e.id
	case "this": // ex fs
		return e.this
	case "args":
		return e.args
	}
	panic(fmt.Sprintf("TODO: event.%s", propertyKey))
}

var NaN = math.NaN()

// LoadValue reads up to 8 bytes at the memory offset `addr` to return the
// value written by storeValue.
//
// See https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L122-L133
func LoadValue(ctx context.Context, ref goos.Ref) interface{} { //nolint
	switch ref {
	case 0:
		return goos.Undefined
	case goos.RefValueNaN:
		return NaN
	case goos.RefValueZero:
		return float64(0)
	case goos.RefValueNull:
		return nil
	case goos.RefValueTrue:
		return true
	case goos.RefValueFalse:
		return false
	case goos.RefValueGlobal:
		return getState(ctx).valueGlobal
	case goos.RefJsGo:
		return getState(ctx)
	case goos.RefObjectConstructor:
		return objectConstructor
	case goos.RefArrayConstructor:
		return arrayConstructor
	case goos.RefJsProcess:
		return getState(ctx).valueGlobal.Get("process")
	case goos.RefJsfs:
		return getState(ctx).valueGlobal.Get("fs")
	case goos.RefJsfsConstants:
		return jsfsConstants
	case goos.RefUint8ArrayConstructor:
		return uint8ArrayConstructor
	case goos.RefJsCrypto:
		return jsCrypto
	case goos.RefJsDateConstructor:
		return jsDateConstructor
	case goos.RefJsDate:
		return jsDate
	default:
		if f, ok := ref.ParseFloat(); ok { // numbers are passed through as a Ref
			return f
		}
		return getState(ctx).values.Get(uint32(ref))
	}
}

// storeValue stores a value prior to returning to wasm from a host function.
// This returns 8 bytes to represent either the value or a reference to it.
// Any side effects besides memory must be cleaned up on wasmExit.
//
// See https://github.com/golang/go/blob/de4748c47c67392a57f250714509f590f68ad395/misc/wasm/wasm_exec.js#L135-L183
func storeValue(ctx context.Context, v interface{}) goos.Ref { //nolint
	// allow-list because we control all implementations
	if v == goos.Undefined {
		return goos.RefValueUndefined
	} else if v == nil {
		return goos.RefValueNull
	} else if r, ok := v.(goos.Ref); ok {
		return r
	} else if b, ok := v.(bool); ok {
		if b {
			return goos.RefValueTrue
		} else {
			return goos.RefValueFalse
		}
	} else if c, ok := v.(*jsVal); ok {
		return c.ref // already stored
	} else if _, ok := v.(*event); ok {
		id := getState(ctx).values.Increment(v)
		return goos.ValueRef(id, goos.TypeFlagFunction)
	} else if _, ok := v.(funcWrapper); ok {
		id := getState(ctx).values.Increment(v)
		return goos.ValueRef(id, goos.TypeFlagFunction)
	} else if _, ok := v.(jsFn); ok {
		id := getState(ctx).values.Increment(v)
		return goos.ValueRef(id, goos.TypeFlagFunction)
	} else if _, ok := v.(string); ok {
		id := getState(ctx).values.Increment(v)
		return goos.ValueRef(id, goos.TypeFlagString)
	} else if i32, ok := v.(int32); ok {
		return toFloatRef(float64(i32))
	} else if u32, ok := v.(uint32); ok {
		return toFloatRef(float64(u32))
	} else if i64, ok := v.(int64); ok {
		return toFloatRef(float64(i64))
	} else if u64, ok := v.(uint64); ok {
		return toFloatRef(float64(u64))
	} else if f64, ok := v.(float64); ok {
		return toFloatRef(f64)
	}
	id := getState(ctx).values.Increment(v)
	return goos.ValueRef(id, goos.TypeFlagObject)
}

func toFloatRef(f float64) goos.Ref {
	if f == 0 {
		return goos.RefValueZero
	}
	// numbers are encoded as float and passed through as a Ref
	return goos.Ref(api.EncodeF64(f))
}

// State holds state used by the "go" imports used by gojs.
// Note: This is module-scoped.
type State struct {
	config        *config.Config
	values        *values.Values
	_pendingEvent *event
	// _lastEvent was the last _pendingEvent value
	_lastEvent *event

	valueGlobal *jsVal

	_nextCallbackTimeoutID uint32
	_scheduledTimeouts     map[uint32]chan bool
}

// Get implements the same method as documented on goos.GetFunction
func (s *State) Get(propertyKey string) interface{} {
	switch propertyKey {
	case "_pendingEvent":
		return s._pendingEvent
	}
	panic(fmt.Sprintf("TODO: state.%s", propertyKey))
}

// call implements jsCall.call
func (s *State) call(_ context.Context, _ api.Module, _ goos.Ref, method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "_makeFuncWrapper":
		return funcWrapper(args[0].(float64)), nil
	}
	panic(fmt.Sprintf("TODO: state.%s", method))
}

// close releases any state including values and underlying slices for garbage
// collection.
func (s *State) close() {
	// _scheduledTimeouts may have in-flight goroutines, so cancel them.
	for k, cancel := range s._scheduledTimeouts {
		delete(s._scheduledTimeouts, k)
		cancel <- true
	}
	// Reset all state recursively to their initial values. This allows our
	// unit tests to check we closed everything.
	s.values.Reset()
	s._pendingEvent = nil
	s._lastEvent = nil
	s.valueGlobal = newJsGlobal(s.config)
	s._nextCallbackTimeoutID = 1
	s._scheduledTimeouts = map[uint32]chan bool{}
}

func toInt64(arg interface{}) int64 {
	if arg == goos.RefValueZero || arg == goos.Undefined {
		return 0
	} else if u, ok := arg.(int64); ok {
		return u
	}
	return int64(arg.(float64))
}

func toUint64(arg interface{}) uint64 {
	if arg == goos.RefValueZero || arg == goos.Undefined {
		return 0
	} else if u, ok := arg.(uint64); ok {
		return u
	}
	return uint64(arg.(float64))
}

// valueString returns the string form of JavaScript string, boolean and number types.
func valueString(v interface{}) string { //nolint
	if s, ok := v.(string); ok {
		return s
	} else {
		return fmt.Sprintf("%v", v)
	}
}
