package git

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/zap/ot"
)

// TODO(sqs): make callers of these instead call the pathutil funcs

func panicIfFileOrBufferPath(path string) {
	if strings.HasPrefix(path, "#") || strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("unexpected file or buffer path %q", path))
	}
}

func panicIfNotFileOrBufferPath(path string) {
	if !strings.HasPrefix(path, "#") && !strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("path %q is not a buffer or file path", path))
	}
}

func isBufferPath(path string) bool {
	panicIfNotFileOrBufferPath(path)
	return strings.HasPrefix(path, "#")
}

func isFilePath(path string) bool {
	panicIfNotFileOrBufferPath(path)
	return strings.HasPrefix(path, "/")
}

func stripBufferPath(path string) string {
	if !strings.HasPrefix(path, "#") {
		panic(fmt.Sprintf("expected path %q to have '#' prefix", path))
	}
	return strings.TrimPrefix(path, "#")
}

func stripFilePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("expected path %q to have '/' prefix", path))
	}
	return strings.TrimPrefix(path, "/")
}

func stripFileOrBufferPath(path string) string {
	panicIfNotFileOrBufferPath(path)
	return path[1:]
}

// FromDiskPaths takes an op whose file names are unprefixed file
// system paths and returns the op with all of those paths prefixed
// with "/". It is used because some Git operations are unaware of our
// buffer path convention (file name "#" prefixes). FromDiskPaths
// translates ops returned by those Git operations to ops that are
// valid in the rest of Zap.
//
// For example:
//
//   FromDiskPaths({Edit: {"foo": ["x"]}}) -> {Edit: {"/foo": ["x"]}}
//
func FromDiskPaths(ops ot.Ops) ot.Ops {
	fromDiskPath := func(path string) string {
		return "/" + path
	}

	op2 := ot.Ops{}
	for _, iop := range ops {
		switch op := iop.(type) {
		case ot.FileCopy:
			op2 = append(op2, ot.FileCopy{Src: fromDiskPath(op.Src), Dst: fromDiskPath(op.Dst)})
		case ot.FileRename:
			op2 = append(op2, ot.FileRename{Src: fromDiskPath(op.Src), Dst: fromDiskPath(op.Dst)})
		case ot.FileCreate:
			op2 = append(op2, ot.FileCreate{File: fromDiskPath(op.File)})
		case ot.FileDelete:
			op2 = append(op2, ot.FileDelete{File: fromDiskPath(op.File)})
		case ot.FileTruncate:
			op2 = append(op2, ot.FileTruncate{File: fromDiskPath(op.File)})
		case ot.FileEdit:
			op2 = append(op2, ot.FileEdit{File: fromDiskPath(op.File), Edits: op.Edits})
		case ot.GitHead:
			op2 = append(op2, op)
		}
	}

	return op2
}

// A FileBuffer is an in-memory file system for storing buffered
// (unsaved) file contents.
type FileBuffer map[string][]byte

var _ FileSystem = make(FileBuffer)

// ReadFile implements FileSystem.
func (b FileBuffer) ReadFile(name string) ([]byte, error) {
	panicIfFileOrBufferPath(name)
	if data, ok := b[name]; ok {
		return data, nil
	}
	return nil, &os.PathError{Op: "ReadFile", Path: "(buf)" + name, Err: os.ErrNotExist}
}

// WriteFile implements FileSystem.
func (b FileBuffer) WriteFile(name string, data []byte, mode os.FileMode) error {
	panicIfFileOrBufferPath(name)
	b[name] = data
	return nil
}

// Rename implements FileSystem.
func (b FileBuffer) Rename(oldpath, newpath string) error {
	panicIfFileOrBufferPath(oldpath)
	panicIfFileOrBufferPath(newpath)
	if _, ok := b[oldpath]; !ok {
		return &os.PathError{Op: "Rename", Path: "(buf)" + oldpath, Err: os.ErrNotExist}
	}
	b[newpath] = b[oldpath]
	delete(b, oldpath)
	return nil
}

// Exists implements FileSystem.
func (b FileBuffer) Exists(name string) error {
	panicIfFileOrBufferPath(name)
	if _, ok := b[name]; !ok {
		return &os.PathError{Op: "Exists", Path: "(buf)" + name, Err: os.ErrNotExist}
	}
	return nil
}

// Remove implements FileSystem.
func (b FileBuffer) Remove(name string) error {
	panicIfFileOrBufferPath(name)
	if _, ok := b[name]; !ok {
		return &os.PathError{Op: "Remove", Path: "(buf)" + name, Err: os.ErrNotExist}
	}
	delete(b, name)
	return nil
}

// Copy implements FileSystemCopier.
func (b FileBuffer) Copy() FileSystem {
	if b == nil {
		return FileBuffer(nil)
	}
	b2 := make(FileBuffer, len(b))
	for name, data := range b {
		tmp := make([]byte, len(data))
		copy(tmp, data)
		b2[name] = tmp
	}
	return b2
}

func (b FileBuffer) String() string {
	if b == nil {
		return "<nil>"
	}
	var buf bytes.Buffer
	i := 0
	for name, data := range b {
		if i != 0 {
			fmt.Fprint(&buf, " ")
		}
		var extra string
		if max := 7; len(data) > max {
			data = data[:max]
			extra = fmt.Sprintf("+%d", len(data)-max)
		}
		fmt.Fprintf(&buf, "%s:%q%s", name, data, extra)
	}
	return buf.String()
}

// A FileSystemCopier is a file system that can produce a deep copy of
// itself.
type FileSystemCopier interface {
	Copy() FileSystem
}
