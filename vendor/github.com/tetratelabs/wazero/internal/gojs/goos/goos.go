// Package goos isolates code from runtime.GOOS=js in a way that avoids cyclic
// dependencies when re-used from other packages.
package goos

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/goarch"
	"github.com/tetratelabs/wazero/internal/gojs/util"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// Ref is used to identify a JavaScript value, since the value itself cannot
// be passed to WebAssembly.
//
// The JavaScript value "undefined" is represented by the value 0.
//
// A JavaScript number (64-bit float, except 0 and NaN) is represented by its
// IEEE 754 binary representation.
//
// All other values are represented as an IEEE 754 binary representation of NaN
// with bits 0-31 used as an ID and bits 32-34 used to differentiate between
// string, symbol, function and object.
type Ref uint64

const (
	// predefined

	IdValueNaN uint32 = iota
	IdValueZero
	IdValueNull
	IdValueTrue
	IdValueFalse
	IdValueGlobal
	IdJsGo

	// The below are derived from analyzing `*_js.go` source.

	IdObjectConstructor
	IdArrayConstructor
	IdJsProcess
	IdJsfs
	IdJsfsConstants
	IdUint8ArrayConstructor
	IdJsCrypto
	IdJsDateConstructor
	IdJsDate
	NextID
)

const (
	RefValueUndefined = Ref(0)
	RefValueNaN       = (NanHead|Ref(TypeFlagNone))<<32 | Ref(IdValueNaN)
	RefValueZero      = (NanHead|Ref(TypeFlagNone))<<32 | Ref(IdValueZero)
	RefValueNull      = (NanHead|Ref(TypeFlagNone))<<32 | Ref(IdValueNull)
	RefValueTrue      = (NanHead|Ref(TypeFlagNone))<<32 | Ref(IdValueTrue)
	RefValueFalse     = (NanHead|Ref(TypeFlagNone))<<32 | Ref(IdValueFalse)
	RefValueGlobal    = (NanHead|Ref(TypeFlagObject))<<32 | Ref(IdValueGlobal)
	RefJsGo           = (NanHead|Ref(TypeFlagObject))<<32 | Ref(IdJsGo)

	RefObjectConstructor     = (NanHead|Ref(TypeFlagFunction))<<32 | Ref(IdObjectConstructor)
	RefArrayConstructor      = (NanHead|Ref(TypeFlagFunction))<<32 | Ref(IdArrayConstructor)
	RefJsProcess             = (NanHead|Ref(TypeFlagObject))<<32 | Ref(IdJsProcess)
	RefJsfs                  = (NanHead|Ref(TypeFlagObject))<<32 | Ref(IdJsfs)
	RefJsfsConstants         = (NanHead|Ref(TypeFlagObject))<<32 | Ref(IdJsfsConstants)
	RefUint8ArrayConstructor = (NanHead|Ref(TypeFlagFunction))<<32 | Ref(IdUint8ArrayConstructor)
	RefJsCrypto              = (NanHead|Ref(TypeFlagFunction))<<32 | Ref(IdJsCrypto)
	RefJsDateConstructor     = (NanHead|Ref(TypeFlagFunction))<<32 | Ref(IdJsDateConstructor)
	RefJsDate                = (NanHead|Ref(TypeFlagObject))<<32 | Ref(IdJsDate)
)

type TypeFlag byte

// the type flags need to be in sync with gojs.js
const (
	TypeFlagNone TypeFlag = iota
	TypeFlagObject
	TypeFlagString
	TypeFlagSymbol //nolint
	TypeFlagFunction
)

func ValueRef(id uint32, typeFlag TypeFlag) Ref {
	return (NanHead|Ref(typeFlag))<<32 | Ref(id)
}

var le = binary.LittleEndian

// NanHead are the upper 32 bits of a Ref which are set if the value is not encoded as an IEEE 754 number (see above).
const NanHead = 0x7FF80000

func (ref Ref) ParseFloat() (v float64, ok bool) {
	if (ref>>32)&NanHead != NanHead {
		v = api.DecodeF64(uint64(ref))
		ok = true
	}
	return
}

// GetLastEventArgs returns the arguments to the last event created by
// custom.NameSyscallValueCall.
type GetLastEventArgs func(context.Context) []interface{}

type ValLoader func(context.Context, Ref) interface{}

type Stack interface {
	goarch.Stack

	ParamRef(i int) Ref

	ParamRefs(mem api.Memory, i int) []Ref

	ParamVal(ctx context.Context, i int, loader ValLoader) interface{}

	// ParamVals is used by functions whose final parameter is an arg array.
	ParamVals(ctx context.Context, mem api.Memory, i int, loader ValLoader) []interface{}

	SetResultRef(i int, v Ref)
}

type stack struct {
	s goarch.Stack
}

// Name implements the same method as documented on goarch.Stack
func (s *stack) Name() string {
	return s.s.Name()
}

// Param implements the same method as documented on goarch.Stack
func (s *stack) Param(i int) uint64 {
	return s.s.Param(i)
}

// ParamBytes implements the same method as documented on goarch.Stack
func (s *stack) ParamBytes(mem api.Memory, i int) []byte {
	return s.s.ParamBytes(mem, i)
}

// ParamRef implements Stack.ParamRef
func (s *stack) ParamRef(i int) Ref {
	return Ref(s.s.Param(i))
}

// ParamRefs implements Stack.ParamRefs
func (s *stack) ParamRefs(mem api.Memory, i int) []Ref {
	offset := s.s.ParamUint32(i)
	size := s.s.ParamUint32(i + 1)
	byteCount := size << 3 // size * 8

	result := make([]Ref, 0, size)

	buf := util.MustRead(mem, s.Name(), i, offset, byteCount)
	for pos := uint32(0); pos < byteCount; pos += 8 {
		ref := Ref(le.Uint64(buf[pos:]))
		result = append(result, ref)
	}
	return result
}

// ParamString implements the same method as documented on goarch.Stack
func (s *stack) ParamString(mem api.Memory, i int) string {
	return s.s.ParamString(mem, i)
}

// ParamInt32 implements the same method as documented on goarch.Stack
func (s *stack) ParamInt32(i int) int32 {
	return s.s.ParamInt32(i)
}

// ParamUint32 implements the same method as documented on goarch.Stack
func (s *stack) ParamUint32(i int) uint32 {
	return s.s.ParamUint32(i)
}

// ParamVal implements Stack.ParamVal
func (s *stack) ParamVal(ctx context.Context, i int, loader ValLoader) interface{} {
	ref := s.ParamRef(i)
	return loader(ctx, ref)
}

// ParamVals implements Stack.ParamVals
func (s *stack) ParamVals(ctx context.Context, mem api.Memory, i int, loader ValLoader) []interface{} {
	offset := s.s.ParamUint32(i)
	size := s.s.ParamUint32(i + 1)
	byteCount := size << 3 // size * 8

	result := make([]interface{}, 0, size)

	buf := util.MustRead(mem, s.Name(), i, offset, byteCount)
	for pos := uint32(0); pos < byteCount; pos += 8 {
		ref := Ref(le.Uint64(buf[pos:]))
		result = append(result, loader(ctx, ref))
	}
	return result
}

// Refresh implements the same method as documented on goarch.Stack
func (s *stack) Refresh(mod api.Module) {
	s.s.Refresh(mod)
}

// SetResult implements the same method as documented on goarch.Stack
func (s *stack) SetResult(i int, v uint64) {
	s.s.SetResult(i, v)
}

// SetResultBool implements the same method as documented on goarch.Stack
func (s *stack) SetResultBool(i int, v bool) {
	s.s.SetResultBool(i, v)
}

// SetResultI32 implements the same method as documented on goarch.Stack
func (s *stack) SetResultI32(i int, v int32) {
	s.s.SetResultI32(i, v)
}

// SetResultI64 implements the same method as documented on goarch.Stack
func (s *stack) SetResultI64(i int, v int64) {
	s.s.SetResultI64(i, v)
}

// SetResultRef implements Stack.SetResultRef
func (s *stack) SetResultRef(i int, v Ref) {
	s.s.SetResult(i, uint64(v))
}

// SetResultUint32 implements the same method as documented on goarch.Stack
func (s *stack) SetResultUint32(i int, v uint32) {
	s.s.SetResultUint32(i, v)
}

func NewFunc(name string, goFunc Func) *wasm.HostFunc {
	sf := &stackFunc{name: name, f: goFunc}
	return util.NewFunc(name, sf.Call)
}

type Func func(context.Context, api.Module, Stack)

type stackFunc struct {
	name string
	f    Func
}

// Call implements the same method as defined on api.GoModuleFunction.
func (f *stackFunc) Call(ctx context.Context, mod api.Module, wasmStack []uint64) {
	s := NewStack(f.name, mod.Memory(), uint32(wasmStack[0]))
	f.f(ctx, mod, s)
}

func NewStack(name string, mem api.Memory, sp uint32) *stack {
	return &stack{goarch.NewStack(name, mem, sp)}
}

var Undefined = struct{ name string }{name: "undefined"}

func ValueToUint32(arg interface{}) uint32 {
	if arg == RefValueZero || arg == Undefined {
		return 0
	} else if u, ok := arg.(uint32); ok {
		return u
	}
	return uint32(arg.(float64))
}

func ValueToInt32(arg interface{}) int32 {
	if arg == RefValueZero || arg == Undefined {
		return 0
	} else if u, ok := arg.(int); ok {
		return int32(u)
	}
	return int32(uint32(arg.(float64)))
}

// GetFunction allows getting a JavaScript property by name.
type GetFunction interface {
	Get(propertyKey string) interface{}
}

// ByteArray is a result of uint8ArrayConstructor which temporarily stores
// binary data outside linear memory.
//
// Note: This is a wrapper because a slice is not hashable.
type ByteArray struct {
	slice []byte
}

func WrapByteArray(buf []byte) *ByteArray {
	return &ByteArray{buf}
}

// Unwrap returns the underlying byte slice
func (a *ByteArray) Unwrap() []byte {
	return a.slice
}

// Get implements GetFunction
func (a *ByteArray) Get(propertyKey string) interface{} {
	switch propertyKey {
	case "byteLength":
		return uint32(len(a.slice))
	}
	panic(fmt.Sprintf("TODO: get byteArray.%s", propertyKey))
}
