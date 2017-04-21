// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctxvfs

import (
	"bytes"
	"context"
	"os"
	"path"
	"sort"
	"strings"
)

// Map returns a new FileSystem from the provided map. Map keys
// should be forward slash-separated pathnames and not contain a
// leading slash.
//
// It differs from golang.org/x/tools/godoc/vfs/mapfs in the following
// ways:
//
// * It uses []byte instead of string to hold file contents. This
//   means that callers must synchronize writes (if any) to the byte
//   arrays.
// * It implements ctxvfs.FileSystem, not vfs.FileSystem.
func Map(m map[string][]byte) FileSystem {
	for k := range m {
		if strings.HasPrefix(k, "/") {
			panic("filename has illegal leading slash: " + k)
		}
	}

	return mapFS(m)
}

// mapFS is the map based implementation of FileSystem
type mapFS map[string][]byte

func (fs mapFS) String() string { return "mapfs" }

func (fs mapFS) Close() error { return nil }

func filename(p string) string {
	return strings.TrimPrefix(p, "/")
}

func (fs mapFS) Open(ctx context.Context, p string) (ReadSeekCloser, error) {
	b, ok := fs[filename(p)]
	if !ok {
		return nil, &os.PathError{Op: "Open", Path: p, Err: os.ErrNotExist}
	}
	return nopCloser{bytes.NewReader(b)}, nil
}

func (fs mapFS) Lstat(ctx context.Context, p string) (os.FileInfo, error) {
	if p == "/" {
		return dirInfo("/"), nil
	}
	p = filename(p)
	b, ok := fs[p]
	if ok {
		return fileInfo{path.Base(p), int64(len(b))}, nil
	}
	var ps string
	if strings.HasSuffix(p, "/") {
		ps = p
		p = strings.TrimSuffix(p, "/")
	} else {
		ps = p + "/"
	}
	for fn := range fs {
		if strings.HasPrefix(fn, "/") {
			panic("filename has illegal leading slash: " + fn)
		}
		if strings.HasPrefix(fn, ps) {
			return dirInfo(path.Base(p)), nil
		}
	}
	return nil, &os.PathError{Op: "Lstat", Path: p, Err: os.ErrNotExist}
}

func (fs mapFS) Stat(ctx context.Context, p string) (os.FileInfo, error) {
	return fs.Lstat(ctx, p)
}

// slashdir returns path.Dir(p), but special-cases paths not beginning
// with a slash to be in the root.
func slashdir(p string) string {
	d := path.Dir(p)
	if d == "." || d == "/" {
		return "/"
	}
	if strings.HasPrefix(p, "/") {
		return d
	}
	return "/" + d
}

func (fs mapFS) ReadDir(ctx context.Context, p string) ([]os.FileInfo, error) {
	p = path.Clean(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	var ents []string
	fim := make(map[string]os.FileInfo) // base -> fi
	for fn, b := range fs {
		dir := slashdir(fn)
		isFile := true
		var lastBase string
		for {
			if dir == p {
				base := lastBase
				if isFile {
					base = path.Base(fn)
				}
				if fim[base] == nil {
					var fi os.FileInfo
					if isFile {
						fi = fileInfo{base, int64(len(b))}
					} else {
						fi = dirInfo(base)
					}
					ents = append(ents, base)
					fim[base] = fi
				}
			}
			if dir == "/" {
				break
			} else {
				isFile = false
				dir, lastBase = path.Split(dir)
				if len(dir) > 1 {
					dir = dir[:len(dir)-1] // trim trailing "/"
				}
			}
		}
	}
	if len(ents) == 0 {
		return nil, &os.PathError{Op: "ReadDir", Path: p, Err: os.ErrNotExist}
	}

	sort.Strings(ents)
	var list []os.FileInfo
	for _, dir := range ents {
		list = append(list, fim[dir])
	}
	return list, nil
}
