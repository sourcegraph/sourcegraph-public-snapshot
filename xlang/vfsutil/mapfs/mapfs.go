// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mapfs file provides an implementation of the FileSystem
// interface based on the contents of a map[string][]byte.
//
// It differs from golang.org/x/tools/godoc/vfs in that it uses []byte
// instead of string to hold file contents. This means that callers
// must synchronize writes (if any) to the byte arrays. It also
// expects the map to contain paths starting with "/".
package mapfs

import (
	"bytes"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/tools/godoc/vfs"
)

// TODO(sqs): optimize given that our use case is usually append only
// (i.e., we never remove entries)

// New returns a new FileSystem from the provided map.  Map keys
// should be forward slash-separated pathnames and must contain a
// leading slash.
func New(m map[string][]byte) vfs.FileSystem {
	return mapFS(m)
}

// mapFS is the map based implementation of FileSystem
type mapFS map[string][]byte

func (fs mapFS) String() string { return "mapfs" }

func (fs mapFS) Close() error { return nil }

func filename(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
}

func (fs mapFS) Open(p string) (vfs.ReadSeekCloser, error) {
	b, ok := fs[filename(p)]
	if !ok {
		return nil, os.ErrNotExist
	}
	return nopCloser{bytes.NewReader(b)}, nil
}

func fileInfo(name string, contents []byte) os.FileInfo {
	return mapFI{name: pathBase(name), size: len(contents)}
}

func dirInfo(name string) os.FileInfo {
	return mapFI{name: pathBase(name), dir: true}
}

func (fs mapFS) Lstat(p string) (os.FileInfo, error) {
	b, ok := fs[filename(p)]
	if ok {
		return fileInfo(p, b), nil
	}
	if p == "/" {
		return dirInfo("/"), nil
	}
	var ps string
	if strings.HasSuffix(p, "/") {
		ps = p
		p = strings.TrimSuffix(p, "/")
	} else {
		ps = p + "/"
	}
	for fn := range fs {
		if !strings.HasPrefix(fn, "/") {
			panic("filename has no leading slash: " + fn)
		}
		if strings.HasPrefix(fn, ps) {
			return dirInfo(p), nil
		}
	}
	return nil, os.ErrNotExist
}

func (fs mapFS) Stat(p string) (os.FileInfo, error) {
	return fs.Lstat(p)
}

// slashdir returns path.Dir(p), but special-cases paths not beginning
// with a slash to be in the root.
func slashdir(p string) string {
	d := pathDir(p)
	if d == "." {
		return "/"
	}
	if strings.HasPrefix(p, "/") {
		return d
	}
	return "/" + d
}

func (fs mapFS) ReadDir(p string) ([]os.FileInfo, error) {
	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	// if p != pathpkg.Clean(p) {
	// 	panic(fmt.Sprintf("unclean path %q", p))
	// }
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
					base = pathBase(fn)
				}
				if fim[base] == nil {
					var fi os.FileInfo
					if isFile {
						fi = fileInfo(fn, b)
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
				dir, lastBase = pathSplit(dir)
			}
		}
	}
	if len(ents) == 0 {
		return nil, os.ErrNotExist
	}

	sort.Strings(ents)
	var list []os.FileInfo
	for _, dir := range ents {
		list = append(list, fim[dir])
	}
	return list, nil
}

// pathDir is a faster and simpler version of path.Dir.
func pathDir(path string) string {
	path = path[:strings.LastIndex(path, "/")+1]
	if path == "" {
		return "/"
	}
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

// pathBase is a faster and simpler version of path.Dir.
func pathBase(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		path = path[i+1:]
	}
	if path == "" {
		return "/"
	}
	return path
}

func pathSplit(path string) (dir, file string) {
	i := strings.LastIndex(path, "/")
	dir = path[:i+1]
	if len(dir) > 1 && dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}
	if dir == "" {
		dir = "/"
	}
	file = path[i+1:]
	return
}

// mapFI is the map-based implementation of FileInfo.
type mapFI struct {
	name string
	size int
	dir  bool
}

func (fi mapFI) IsDir() bool        { return fi.dir }
func (fi mapFI) ModTime() time.Time { return time.Time{} }
func (fi mapFI) Mode() os.FileMode {
	if fi.IsDir() {
		return 0755 | os.ModeDir
	}
	return 0444
}
func (fi mapFI) Name() string     { return pathBase(fi.name) }
func (fi mapFI) Size() int64      { return int64(fi.size) }
func (fi mapFI) Sys() interface{} { return nil }

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
