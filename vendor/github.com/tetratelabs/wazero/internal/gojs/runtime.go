package gojs

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/custom"
	"github.com/tetratelabs/wazero/internal/gojs/goarch"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// Debug has unknown use, so stubbed.
//
// See https://github.com/golang/go/blob/go1.20/src/cmd/link/internal/wasm/asm.go#L131-L136
var Debug = goarch.StubFunction(custom.NameDebug)

// TODO: should this call runtime.Breakpoint()?

// WasmExit implements runtime.wasmExit which supports runtime.exit.
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/sys_wasm.go#L24
var WasmExit = goarch.NewFunc(custom.NameRuntimeWasmExit, wasmExit)

func wasmExit(ctx context.Context, mod api.Module, stack goarch.Stack) {
	code := stack.ParamUint32(0)

	getState(ctx).close()
	_ = mod.CloseWithExitCode(ctx, code)
}

// WasmWrite implements runtime.wasmWrite which supports runtime.write and
// runtime.writeErr. This implements `println`.
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/os_js.go#L30
var WasmWrite = goarch.NewFunc(custom.NameRuntimeWasmWrite, wasmWrite)

func wasmWrite(_ context.Context, mod api.Module, stack goarch.Stack) {
	fd := stack.ParamInt32(0)
	p := stack.ParamBytes(mod.Memory(), 1 /*, 2 */)

	fsc := mod.(*wasm.ModuleInstance).Sys.FS()
	if f, ok := fsc.LookupFile(fd); ok {
		_, errno := f.File.Write(p)
		switch errno {
		case 0:
			return // success
		case syscall.ENOSYS:
			return // e.g. unimplemented for write
		case syscall.EBADF:
			return // e.g. not opened for write
		default:
			panic(fmt.Errorf("error writing p: %w", errno))
		}
	} else {
		panic(fmt.Errorf("fd %d invalid", fd))
	}
}

// ResetMemoryDataView signals wasm.OpcodeMemoryGrow happened, indicating any
// cached view of memory should be reset.
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/mem_js.go#L82
var ResetMemoryDataView = goarch.NewFunc(custom.NameRuntimeResetMemoryDataView, resetMemoryDataView)

func resetMemoryDataView(context.Context, api.Module, goarch.Stack) {
	// context state does not cache a memory view, and all byte slices used
	// are safely copied. Also, user-defined functions are not supported.
	// Hence, there's currently no known reason to reset anything.
}

// Nanotime1 implements runtime.nanotime which supports time.Since.
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/sys_wasm.s#L117
var Nanotime1 = goarch.NewFunc(custom.NameRuntimeNanotime1, nanotime1)

func nanotime1(_ context.Context, mod api.Module, stack goarch.Stack) {
	nsec := mod.(*wasm.ModuleInstance).Sys.Nanotime()

	stack.SetResultI64(0, nsec)
}

// Walltime implements runtime.walltime which supports time.Now.
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/sys_wasm.s#L121
var Walltime = goarch.NewFunc(custom.NameRuntimeWalltime, walltime)

func walltime(_ context.Context, mod api.Module, stack goarch.Stack) {
	sec, nsec := mod.(*wasm.ModuleInstance).Sys.Walltime()

	stack.SetResultI64(0, sec)
	stack.SetResultI32(1, nsec)
}

// ScheduleTimeoutEvent implements runtime.scheduleTimeoutEvent which supports
// runtime.notetsleepg used by runtime.signal_recv.
//
// Unlike other most functions prefixed by "runtime.", this both launches a
// goroutine and invokes code compiled into wasm "resume".
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/sys_wasm.s#L125
var ScheduleTimeoutEvent = goarch.NewFunc(custom.NameRuntimeScheduleTimeoutEvent, scheduleTimeoutEvent)

// Note: Signal handling is not implemented in GOOS=js.
func scheduleTimeoutEvent(ctx context.Context, mod api.Module, stack goarch.Stack) {
	ms := stack.Param(0)

	s := getState(ctx)
	id := s._nextCallbackTimeoutID
	stack.SetResultUint32(0, id)
	s._nextCallbackTimeoutID++

	cleared := make(chan bool)
	timeout := time.Millisecond * time.Duration(ms)
	s._scheduledTimeouts[id] = cleared

	// As wasm is currently not concurrent, a timeout on another goroutine may
	// not make sense. However, this implements what wasm_exec.js does anyway.
	go func() {
		select {
		case <-cleared: // do nothing
		case <-time.After(timeout):
			if _, err := mod.ExportedFunction("resume").Call(ctx); err != nil {
				println(err)
			}
		}
	}()
}

// ClearTimeoutEvent implements runtime.clearTimeoutEvent which supports
// runtime.notetsleepg used by runtime.signal_recv.
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/sys_wasm.s#L129
var ClearTimeoutEvent = goarch.NewFunc(custom.NameRuntimeClearTimeoutEvent, clearTimeoutEvent)

// Note: Signal handling is not implemented in GOOS=js.
func clearTimeoutEvent(ctx context.Context, _ api.Module, stack goarch.Stack) {
	id := stack.ParamUint32(0)
	s := getState(ctx)
	if cancel, ok := s._scheduledTimeouts[id]; ok {
		delete(s._scheduledTimeouts, id)
		cancel <- true
	}
}

// GetRandomData implements runtime.getRandomData, which initializes the seed
// for runtime.fastrand.
//
// See https://github.com/golang/go/blob/go1.20/src/runtime/sys_wasm.s#L133
var GetRandomData = goarch.NewFunc(custom.NameRuntimeGetRandomData, getRandomData)

func getRandomData(_ context.Context, mod api.Module, stack goarch.Stack) {
	r := stack.ParamBytes(mod.Memory(), 0 /*, 1 */)

	randSource := mod.(*wasm.ModuleInstance).Sys.RandSource()

	bufLen := len(r)
	if n, err := randSource.Read(r); err != nil {
		panic(fmt.Errorf("RandSource.Read(r /* len=%d */) failed: %w", bufLen, err))
	} else if n != bufLen {
		panic(fmt.Errorf("RandSource.Read(r /* len=%d */) read %d bytes", bufLen, n))
	}
}
