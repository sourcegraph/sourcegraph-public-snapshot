package gojs

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/custom"
	"github.com/tetratelabs/wazero/internal/gojs/goarch"
	"github.com/tetratelabs/wazero/internal/gojs/goos"
	"github.com/tetratelabs/wazero/sys"
)

// FinalizeRef implements js.finalizeRef, which is used as a
// runtime.SetFinalizer on the given reference.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L61
var FinalizeRef = goos.NewFunc(custom.NameSyscallFinalizeRef, finalizeRef)

func finalizeRef(ctx context.Context, _ api.Module, stack goos.Stack) {
	r := stack.ParamRef(0)

	id := uint32(r) // 32-bits of the ref are the ID

	getState(ctx).values.Decrement(id)
}

// StringVal implements js.stringVal, which is used to load the string for
// `js.ValueOf(x)`. For example, this is used when setting HTTP headers.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L212
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L305-L308
var StringVal = goos.NewFunc(custom.NameSyscallStringVal, stringVal)

func stringVal(ctx context.Context, mod api.Module, stack goos.Stack) {
	x := stack.ParamString(mod.Memory(), 0)

	r := storeValue(ctx, x)

	stack.SetResultRef(0, r)
}

// ValueGet implements js.valueGet, which is used to load a js.Value property
// by name, e.g. `v.Get("address")`. Notably, this is used by js.handleEvent to
// get the pending event.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L295
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L311-L316
var ValueGet = goos.NewFunc(custom.NameSyscallValueGet, valueGet)

func valueGet(ctx context.Context, mod api.Module, stack goos.Stack) {
	v := stack.ParamVal(ctx, 0, LoadValue)
	p := stack.ParamString(mod.Memory(), 1 /*, 2 */)

	var result interface{}
	if g, ok := v.(goos.GetFunction); ok {
		result = g.Get(p)
	} else if e, ok := v.(error); ok {
		switch p {
		case "message": // js (GOOS=js) error, can be anything.
			result = e.Error()
		case "code": // syscall (GOARCH=wasm) error, must match key in mapJSError in fs_js.go
			result = ToErrno(e).Error()
		default:
			panic(fmt.Errorf("TODO: valueGet(v=%v, p=%s)", v, p))
		}
	} else {
		panic(fmt.Errorf("TODO: valueGet(v=%v, p=%s)", v, p))
	}

	r := storeValue(ctx, result)
	stack.SetResultRef(0, r)
}

// ValueSet implements js.valueSet, which is used to store a js.Value property
// by name, e.g. `v.Set("address", a)`. Notably, this is used by js.handleEvent
// set the event result.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L309
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L318-L322
var ValueSet = goos.NewFunc(custom.NameSyscallValueSet, valueSet)

func valueSet(ctx context.Context, mod api.Module, stack goos.Stack) {
	v := stack.ParamVal(ctx, 0, LoadValue)
	p := stack.ParamString(mod.Memory(), 1 /*, 2 */)
	x := stack.ParamVal(ctx, 3, LoadValue)

	if p := p; v == getState(ctx) {
		switch p {
		case "_pendingEvent":
			if x == nil { // syscall_js.handleEvent
				s := v.(*State)
				s._lastEvent = s._pendingEvent
				s._pendingEvent = nil
				return
			}
		}
	} else if e, ok := v.(*event); ok { // syscall_js.handleEvent
		switch p {
		case "result":
			e.result = x
			return
		}
	} else if m, ok := v.(*object); ok {
		m.properties[p] = x // e.g. opt.Set("method", req.Method)
		return
	}
	panic(fmt.Errorf("TODO: valueSet(v=%v, p=%s, x=%v)", v, p, x))
}

// ValueDelete is stubbed as it isn't used in Go's main source tree.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L321
var ValueDelete = goarch.StubFunction(custom.NameSyscallValueDelete)

// ValueIndex implements js.valueIndex, which is used to load a js.Value property
// by index, e.g. `v.Index(0)`. Notably, this is used by js.handleEvent to read
// event arguments
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L334
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L331-L334
var ValueIndex = goos.NewFunc(custom.NameSyscallValueIndex, valueIndex)

func valueIndex(ctx context.Context, _ api.Module, stack goos.Stack) {
	v := stack.ParamVal(ctx, 0, LoadValue)
	i := stack.ParamUint32(1)

	result := v.(*objectArray).slice[i]

	r := storeValue(ctx, result)
	stack.SetResultRef(0, r)
}

// ValueSetIndex is stubbed as it is only used for js.ValueOf when the input is
// []interface{}, which doesn't appear to occur in Go's source tree.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L348
var ValueSetIndex = goarch.StubFunction(custom.NameSyscallValueSetIndex)

// ValueCall implements js.valueCall, which is used to call a js.Value function
// by name, e.g. `document.Call("createElement", "div")`.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L394
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L343-L358
var ValueCall = goos.NewFunc(custom.NameSyscallValueCall, valueCall)

func valueCall(ctx context.Context, mod api.Module, stack goos.Stack) {
	mem := mod.Memory()
	vRef := stack.ParamRef(0)
	m := stack.ParamString(mem, 1 /*, 2 */)
	args := stack.ParamVals(ctx, mem, 3 /*, 4 */, LoadValue)
	// 5 = padding

	v := LoadValue(ctx, vRef)
	c, isCall := v.(jsCall)
	if !isCall {
		panic(fmt.Errorf("TODO: valueCall(v=%v, m=%s, args=%v)", v, m, args))
	}

	var res goos.Ref
	var ok bool
	if result, err := c.call(ctx, mod, vRef, m, args...); err != nil {
		res = storeValue(ctx, err)
	} else {
		res = storeValue(ctx, result)
		ok = true
	}

	stack.Refresh(mod)
	stack.SetResultRef(0, res)
	stack.SetResultBool(1, ok)
}

// ValueInvoke is stubbed as it isn't used in Go's main source tree.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L413
var ValueInvoke = goarch.StubFunction(custom.NameSyscallValueInvoke)

// ValueNew implements js.valueNew, which is used to call a js.Value, e.g.
// `array.New(2)`.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L432
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L378-L392
var ValueNew = goos.NewFunc(custom.NameSyscallValueNew, valueNew)

func valueNew(ctx context.Context, mod api.Module, stack goos.Stack) {
	mem := mod.Memory()
	vRef := stack.ParamRef(0)
	args := stack.ParamVals(ctx, mem, 1 /*, 2 */, LoadValue)
	// 3 = padding

	var res goos.Ref
	var ok bool
	switch vRef {
	case goos.RefArrayConstructor:
		result := &objectArray{}
		res = storeValue(ctx, result)
		ok = true
	case goos.RefUint8ArrayConstructor:
		var result interface{}
		a := args[0]
		if n, ok := a.(float64); ok {
			result = goos.WrapByteArray(make([]byte, uint32(n)))
		} else if _, ok := a.(*goos.ByteArray); ok {
			// In case of wrapping, increment the counter of the same ref.
			//	uint8arrayWrapper := uint8Array.New(args[0])
			result = stack.ParamRefs(mem, 1)[0]
		} else {
			panic(fmt.Errorf("TODO: valueNew(v=%v, args=%v)", vRef, args))
		}
		res = storeValue(ctx, result)
		ok = true
	case goos.RefObjectConstructor:
		result := &object{properties: map[string]interface{}{}}
		res = storeValue(ctx, result)
		ok = true
	case goos.RefJsDateConstructor:
		res = goos.RefJsDate
		ok = true
	default:
		panic(fmt.Errorf("TODO: valueNew(v=%v, args=%v)", vRef, args))
	}

	stack.Refresh(mod)
	stack.SetResultRef(0, res)
	stack.SetResultBool(1, ok)
}

// ValueLength implements js.valueLength, which is used to load the length
// property of a value, e.g. `array.length`.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L372
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L395-L398
var ValueLength = goos.NewFunc(custom.NameSyscallValueLength, valueLength)

func valueLength(ctx context.Context, _ api.Module, stack goos.Stack) {
	v := stack.ParamVal(ctx, 0, LoadValue)

	len := len(v.(*objectArray).slice)

	stack.SetResultUint32(0, uint32(len))
}

// ValuePrepareString implements js.valuePrepareString, which is used to load
// the string for `o.String()` (via js.jsString) for string, boolean and
// number types. Notably, http.Transport uses this in RoundTrip to coerce the
// URL to a string.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L531
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L401-L406
var ValuePrepareString = goos.NewFunc(custom.NameSyscallValuePrepareString, valuePrepareString)

func valuePrepareString(ctx context.Context, _ api.Module, stack goos.Stack) {
	v := stack.ParamVal(ctx, 0, LoadValue)

	s := valueString(v)

	sRef := storeValue(ctx, s)
	sLen := uint32(len(s))

	stack.SetResultRef(0, sRef)
	stack.SetResultUint32(1, sLen)
}

// ValueLoadString implements js.valueLoadString, which is used copy a string
// value for `o.String()`.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L533
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L409-L413
var ValueLoadString = goos.NewFunc(custom.NameSyscallValueLoadString, valueLoadString)

func valueLoadString(ctx context.Context, mod api.Module, stack goos.Stack) {
	v := stack.ParamVal(ctx, 0, LoadValue)
	b := stack.ParamBytes(mod.Memory(), 1 /*, 2 */)

	s := valueString(v)
	copy(b, s)
}

// ValueInstanceOf is stubbed as it isn't used in Go's main source tree.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L543
var ValueInstanceOf = goarch.StubFunction(custom.NameSyscallValueInstanceOf)

// CopyBytesToGo copies a JavaScript managed byte array to linear memory.
// For example, this is used to read an HTTP response body.
//
// # Results
//
//   - n is the count of bytes written.
//   - ok is false if the src was not a uint8Array.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L569
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L437-L449
var CopyBytesToGo = goos.NewFunc(custom.NameSyscallCopyBytesToGo, copyBytesToGo)

func copyBytesToGo(ctx context.Context, mod api.Module, stack goos.Stack) {
	dst := stack.ParamBytes(mod.Memory(), 0 /*, 1 */)
	// padding = 2
	src := stack.ParamVal(ctx, 3, LoadValue)

	var n uint32
	var ok bool
	if src, isBuf := src.(*goos.ByteArray); isBuf {
		n = uint32(copy(dst, src.Unwrap()))
		ok = true
	}

	stack.SetResultUint32(0, n)
	stack.SetResultBool(1, ok)
}

// CopyBytesToJS copies linear memory to a JavaScript managed byte array.
// For example, this is used to read an HTTP request body.
//
// # Results
//
//   - n is the count of bytes written.
//   - ok is false if the dst was not a uint8Array.
//
// See https://github.com/golang/go/blob/go1.20/src/syscall/js/js.go#L583
// and https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L438-L448
var CopyBytesToJS = goos.NewFunc(custom.NameSyscallCopyBytesToJS, copyBytesToJS)

func copyBytesToJS(ctx context.Context, mod api.Module, stack goos.Stack) {
	dst := stack.ParamVal(ctx, 0, LoadValue)
	src := stack.ParamBytes(mod.Memory(), 1 /*, 2 */)
	// padding = 3

	var n uint32
	var ok bool
	if dst, isBuf := dst.(*goos.ByteArray); isBuf {
		if dst != nil { // empty is possible on EOF
			n = uint32(copy(dst.Unwrap(), src))
		}
		ok = true
	}

	stack.SetResultUint32(0, n)
	stack.SetResultBool(1, ok)
}

// funcWrapper is the result of go's js.FuncOf ("_makeFuncWrapper" here).
//
// This ID is managed on the Go side an increments (possibly rolling over).
type funcWrapper uint32

// invoke implements jsFn
func (f funcWrapper) invoke(ctx context.Context, mod api.Module, args ...interface{}) (interface{}, error) {
	e := &event{id: uint32(f), this: args[0].(goos.Ref)}

	if len(args) > 1 { // Ensure arguments are hashable.
		e.args = &objectArray{args[1:]}
		for i, v := range e.args.slice {
			if s, ok := v.([]byte); ok {
				args[i] = goos.WrapByteArray(s)
			} else if s, ok := v.([]interface{}); ok {
				args[i] = &objectArray{s}
			} else if e, ok := v.(error); ok {
				args[i] = e
			}
		}
	}

	getState(ctx)._pendingEvent = e // Note: _pendingEvent reference is cleared during resume!

	if _, err := mod.ExportedFunction("resume").Call(ctx); err != nil {
		if _, ok := err.(*sys.ExitError); ok {
			return nil, nil // allow error-handling to unwind when wasm calls exit due to a panic
		} else {
			return nil, err
		}
	}

	return e.result, nil
}
