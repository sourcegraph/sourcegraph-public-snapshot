package git

import (
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/zap/ot"
)

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

func FromDiskPaths(op ot.WorkspaceOp) ot.WorkspaceOp {
	fromDiskPath := func(path string) string {
		if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "#") {
			panic(fmt.Sprintf("path %q must be an actual disk path", path))
		}
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

type FileBuffer struct {
	files map[string][]byte
}

func (b *FileBuffer) String() string {
	v := make(map[string]string, len(b.files))
	for name, data := range b.files {
		v[name] = string(data)
	}
	return fmt.Sprintf("%+v", v)
}

var _ FileSystem = &FileBuffer{}

func (b *FileBuffer) ReadFile(name string) ([]byte, error) {
	panicIfFileOrBufferPath(name)
	if data, ok := b.files[name]; ok {
		return data, nil
	}
	return nil, &os.PathError{Op: "ReadFile", Path: "(buf)" + name, Err: os.ErrNotExist}
}

func (b *FileBuffer) WriteFile(name string, data []byte, mode os.FileMode) error {
	panicIfFileOrBufferPath(name)
	if b.files == nil {
		b.files = map[string][]byte{}
	}
	b.files[name] = data
	return nil
}

func (b *FileBuffer) Rename(oldpath, newpath string) error {
	panicIfFileOrBufferPath(oldpath)
	panicIfFileOrBufferPath(newpath)
	if _, ok := b.files[oldpath]; !ok {
		return &os.PathError{Op: "Rename", Path: "(buf)" + oldpath, Err: os.ErrNotExist}
	}
	b.files[newpath] = b.files[oldpath]
	delete(b.files, oldpath)
	return nil
}

func (b *FileBuffer) Exists(name string) error {
	panicIfFileOrBufferPath(name)
	if _, ok := b.files[name]; !ok {
		return &os.PathError{Op: "Exists", Path: "(buf)" + name, Err: os.ErrNotExist}
	}
	return nil
}

func (b *FileBuffer) Remove(name string) error {
	panicIfFileOrBufferPath(name)
	if _, ok := b.files[name]; !ok {
		return &os.PathError{Op: "Remove", Path: "(buf)" + name, Err: os.ErrNotExist}
	}
	delete(b.files, name)
	return nil
}

func (b *FileBuffer) Copy() FileSystem {
	b2 := &FileBuffer{files: make(map[string][]byte, len(b.files))}
	for name, data := range b.files {
		tmp := make([]byte, len(data))
		copy(tmp, data)
		b2.files[name] = tmp
	}
	return b2
}

type FileSystemCopier interface {
	Copy() FileSystem
}
