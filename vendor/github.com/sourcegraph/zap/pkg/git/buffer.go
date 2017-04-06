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
func FromDiskPaths(op ot.WorkspaceOp) ot.WorkspaceOp {
	fromDiskPath := func(path string) string {
		return "/" + path
	}

	var op2 ot.WorkspaceOp

	op2.Copy = make(map[string]string, len(op.Copy))
	for dst, src := range op.Copy {
		op2.Copy[fromDiskPath(dst)] = fromDiskPath(src)
	}

	for src, dst := range op.Rename {
		op2.Rename[fromDiskPath(src)] = fromDiskPath(dst)
	}

	op2.Create = make([]string, len(op.Create))
	for i, f := range op.Create {
		op2.Create[i] = fromDiskPath(f)
	}

	op2.Delete = make([]string, len(op.Delete))
	for i, f := range op.Delete {
		op2.Delete[i] = fromDiskPath(f)
	}

	op2.Truncate = make([]string, len(op.Truncate))
	for i, f := range op.Truncate {
		op2.Truncate[i] = fromDiskPath(f)
	}

	op2.Edit = make(map[string]ot.EditOps, len(op.Edit))
	for f, edits := range op.Edit {
		op2.Edit[fromDiskPath(f)] = edits
	}

	op2.GitHead = op.GitHead

	return ot.NormalizeWorkspaceOp(op2)
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
