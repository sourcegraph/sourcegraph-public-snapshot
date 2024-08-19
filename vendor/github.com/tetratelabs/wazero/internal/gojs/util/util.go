package util

import (
	"fmt"
	pathutil "path"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/custom"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// MustWrite is like api.Memory except that it panics if the offset
// is out of range.
func MustWrite(mem api.Memory, fieldName string, offset uint32, val []byte) {
	if ok := mem.Write(offset, val); !ok {
		panic(fmt.Errorf("out of memory writing %s", fieldName))
	}
}

// MustRead is like api.Memory except that it panics if the offset and
// byteCount are out of range.
func MustRead(mem api.Memory, funcName string, paramIdx int, offset, byteCount uint32) []byte {
	buf, ok := mem.Read(offset, byteCount)
	if ok {
		return buf
	}
	var paramName string
	if names, ok := custom.NameSection[funcName]; ok {
		if paramIdx < len(names.ParamNames) {
			paramName = names.ParamNames[paramIdx]
		}
	}
	if paramName == "" {
		paramName = fmt.Sprintf("%s param[%d]", funcName, paramIdx)
	}
	panic(fmt.Errorf("out of memory reading %s", paramName))
}

func NewFunc(name string, goFunc api.GoModuleFunc) *wasm.HostFunc {
	return &wasm.HostFunc{
		ExportName: name,
		Name:       name,
		ParamTypes: []api.ValueType{api.ValueTypeI32},
		ParamNames: []string{"sp"},
		Code:       wasm.Code{GoFunc: goFunc},
	}
}

// ResolvePath is needed when a non-absolute path is given to a function.
// Unlike other host ABI, GOOS=js maintains the CWD host side.
func ResolvePath(cwd, path string) (resolved string) {
	pathLen := len(path)
	switch {
	case pathLen == 0:
		return cwd
	case pathLen == 1 && path[0] == '.':
		return cwd
	case path[0] == '/':
		resolved = pathutil.Clean(path)
	default:
		resolved = pathutil.Join(cwd, path)
	}

	// If there's a trailing slash, we need to retain it for symlink edge
	// cases. See https://github.com/golang/go/issues/27225
	if len(resolved) > 1 && path[pathLen-1] == '/' {
		return resolved + "/"
	}
	return resolved
}
