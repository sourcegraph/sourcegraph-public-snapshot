// Package goarch isolates code from runtime.GOARCH=wasm in a way that avoids
// cyclic dependencies when re-used from other packages.
package goarch

import (
	"context"
	"encoding/binary"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/custom"
	"github.com/tetratelabs/wazero/internal/gojs/util"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// StubFunction stubs functions not used in Go's main source tree.
// This traps (unreachable opcode) to ensure the function is never called.
func StubFunction(name string) *wasm.HostFunc {
	return &wasm.HostFunc{
		ExportName: name,
		Name:       name,
		ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
		ParamNames: []string{"sp"},
		Code:       wasm.Code{GoFunc: api.GoModuleFunc(func(ctx context.Context, _ api.Module, stack []uint64) {})},
	}
}

var le = binary.LittleEndian

type Stack interface {
	// Name is the function name being invoked.
	Name() string

	Param(i int) uint64

	// ParamBytes reads a byte slice, given its memory offset and length (stack
	// positions i, i+1)
	ParamBytes(mem api.Memory, i int) []byte

	// ParamString reads a string, given its memory offset and length (stack
	// positions i, i+1)
	ParamString(mem api.Memory, i int) string

	ParamInt32(i int) int32

	ParamUint32(i int) uint32

	// Refresh the stack from the current stack pointer (SP).
	//
	// Note: This is needed prior to storing a value when in an operation that
	// can trigger a Go event handler.
	Refresh(api.Module)

	SetResult(i int, v uint64)

	SetResultBool(i int, v bool)

	SetResultI32(i int, v int32)

	SetResultI64(i int, v int64)

	SetResultUint32(i int, v uint32)
}

func NewStack(name string, mem api.Memory, sp uint32) Stack {
	names := custom.NameSection[name]
	s := &stack{name: name, paramCount: len(names.ParamNames), resultCount: len(names.ResultNames)}
	s.refresh(mem, sp)
	return s
}

type stack struct {
	name                    string
	paramCount, resultCount int
	buf                     []byte
}

// Name implements Stack.Name
func (s *stack) Name() string {
	return s.name
}

// Param implements Stack.Param
func (s *stack) Param(i int) (res uint64) {
	pos := i << 3
	res = le.Uint64(s.buf[pos:])
	return
}

// ParamBytes implements Stack.ParamBytes
func (s *stack) ParamBytes(mem api.Memory, i int) (res []byte) {
	offset := s.ParamUint32(i)
	byteCount := s.ParamUint32(i + 1)
	return util.MustRead(mem, s.name, i, offset, byteCount)
}

// ParamString implements Stack.ParamString
func (s *stack) ParamString(mem api.Memory, i int) string {
	return string(s.ParamBytes(mem, i)) // safe copy of guest memory
}

// ParamInt32 implements Stack.ParamInt32
func (s *stack) ParamInt32(i int) int32 {
	return int32(s.Param(i))
}

// ParamUint32 implements Stack.ParamUint32
func (s *stack) ParamUint32(i int) uint32 {
	return uint32(s.Param(i))
}

// Refresh implements Stack.Refresh
func (s *stack) Refresh(mod api.Module) {
	s.refresh(mod.Memory(), GetSP(mod))
}

func (s *stack) refresh(mem api.Memory, sp uint32) {
	count := uint32(s.paramCount + s.resultCount)
	buf, ok := mem.Read(sp+8, count<<3)
	if !ok {
		panic("out of memory reading stack")
	}
	s.buf = buf
}

// SetResult implements Stack.SetResult
func (s *stack) SetResult(i int, v uint64) {
	pos := (s.paramCount + i) << 3
	le.PutUint64(s.buf[pos:], v)
}

// SetResultBool implements Stack.SetResultBool
func (s *stack) SetResultBool(i int, v bool) {
	if v {
		s.SetResultUint32(i, 1)
	} else {
		s.SetResultUint32(i, 0)
	}
}

// SetResultI32 implements Stack.SetResultI32
func (s *stack) SetResultI32(i int, v int32) {
	s.SetResult(i, uint64(v))
}

// SetResultI64 implements Stack.SetResultI64
func (s *stack) SetResultI64(i int, v int64) {
	s.SetResult(i, uint64(v))
}

// SetResultUint32 implements Stack.SetResultUint32
func (s *stack) SetResultUint32(i int, v uint32) {
	s.SetResult(i, uint64(v))
}

// GetSP gets the stack pointer, which is needed prior to storing a value when
// in an operation that can trigger a Go event handler.
//
// See https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js#L210-L213
func GetSP(mod api.Module) uint32 {
	// Cheat by reading global[0] directly instead of through a function proxy.
	// https://github.com/golang/go/blob/go1.20/src/runtime/rt0_js_wasm.s#L87-L90
	return uint32(mod.(*wasm.ModuleInstance).GlobalVal(0))
}

func NewFunc(name string, goFunc Func) *wasm.HostFunc {
	return util.NewFunc(name, (&stackFunc{name: name, f: goFunc}).Call)
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
