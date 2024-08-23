package gojs

import (
	"encoding/binary"
	"errors"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/util"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// Constants about memory layout. See REFERENCE.md
const (
	endOfPageZero     = uint32(4096)                      // runtime.minLegalPointer
	maxArgsAndEnviron = uint32(8192)                      // ld.wasmMinDataAddr - runtime.minLegalPointer
	wasmMinDataAddr   = endOfPageZero + maxArgsAndEnviron // ld.wasmMinDataAddr
)

var le = binary.LittleEndian

// WriteArgsAndEnviron writes arguments and environment variables to memory, so
// they can be read by main, Go compiles as the function export "run".
func WriteArgsAndEnviron(mod api.Module) (argc, argv uint32, err error) {
	mem := mod.Memory()
	sysCtx := mod.(*wasm.ModuleInstance).Sys
	args := sysCtx.Args()
	environ := sysCtx.Environ()

	argc = uint32(len(args))
	offset := endOfPageZero

	strPtr := func(val []byte, field string, i int) (ptr uint32) {
		// TODO: return err and format "%s[%d], field, i"
		ptr = offset
		util.MustWrite(mem, field, offset, append(val, 0))
		offset += uint32(len(val) + 1)
		if pad := offset % 8; pad != 0 {
			offset += 8 - pad
		}
		return
	}

	argvPtrLen := len(args) + 1 + len(environ) + 1
	argvPtrs := make([]uint32, 0, argvPtrLen)
	for i, arg := range args {
		argvPtrs = append(argvPtrs, strPtr(arg, "args", i))
	}
	argvPtrs = append(argvPtrs, 0)

	for i, env := range environ {
		argvPtrs = append(argvPtrs, strPtr(env, "env", i))
	}
	argvPtrs = append(argvPtrs, 0)

	argv = offset

	stop := uint32(argvPtrLen << 3) // argvPtrLen * 8
	if offset+stop >= wasmMinDataAddr {
		err = errors.New("total length of command line and environment variables exceeds limit")
	}

	buf, ok := mem.Read(argv, stop)
	if !ok {
		panic("out of memory reading argvPtrs")
	}
	pos := uint32(0)
	for _, ptr := range argvPtrs {
		le.PutUint64(buf[pos:], uint64(ptr))
		pos += 8
	}

	return
}
