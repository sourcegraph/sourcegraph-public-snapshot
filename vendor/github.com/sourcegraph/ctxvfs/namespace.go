// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctxvfs

import (
	"context"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"sort"
	"strings"
)

// Setting debugNS = true will enable debugging prints about
// name space translations.
const debugNS = false

// A NameSpace is a file system made up of other file systems
// mounted at specific locations in the name space.
//
// It is based on golang.org/x/tools/godoc/vfs Namespace and adds ctx
// method parameters and the ability to list ancestors of mount
// points. See that documentation for details.
type NameSpace map[string][]mountedFS

// A mountedFS handles requests for path by replacing
// a prefix 'old' with 'new' and then calling the fs methods.
type mountedFS struct {
	old string
	fs  FileSystem
	new string
}

// hasPathPrefix returns true if x == y or x == y + "/" + more
func hasPathPrefix(x, y string) bool {
	return x == y || strings.HasPrefix(x, y) && (strings.HasSuffix(y, "/") || strings.HasPrefix(x[len(y):], "/"))
}

// translate translates path for use in m, replacing old with new.
//
// mountedFS{"/src/pkg", fs, "/src"}.translate("/src/pkg/code") == "/src/code".
func (m mountedFS) translate(path string) string {
	path = pathpkg.Clean("/" + path)
	if !hasPathPrefix(path, m.old) {
		panic("translate " + path + " but old=" + m.old)
	}
	return pathpkg.Join(m.new, path[len(m.old):])
}

func (NameSpace) String() string {
	return "ns"
}

// Fprint writes a text representation of the name space to w.
func (ns NameSpace) Fprint(w io.Writer) {
	fmt.Fprint(w, "name space {\n")
	var all []string
	for mtpt := range ns {
		all = append(all, mtpt)
	}
	sort.Strings(all)
	for _, mtpt := range all {
		fmt.Fprintf(w, "\t%s:\n", mtpt)
		for _, m := range ns[mtpt] {
			fmt.Fprintf(w, "\t\t%s %s\n", m.fs, m.new)
		}
	}
	fmt.Fprint(w, "}\n")
}

// clean returns a cleaned, rooted path for evaluation.
// It canonicalizes the path so that we can use string operations
// to analyze it.
func (NameSpace) clean(path string) string {
	return pathpkg.Clean("/" + path)
}

type BindMode int

const (
	BindReplace BindMode = iota
	BindBefore
	BindAfter
)

func (ns NameSpace) Bind(old string, newfs FileSystem, new string, mode BindMode) {
	old = ns.clean(old)
	new = ns.clean(new)
	m := mountedFS{old, newfs, new}
	var mtpt []mountedFS
	switch mode {
	case BindReplace:
		mtpt = append(mtpt, m)
	case BindAfter:
		mtpt = append(mtpt, ns.resolve(old)...)
		mtpt = append(mtpt, m)
	case BindBefore:
		mtpt = append(mtpt, m)
		mtpt = append(mtpt, ns.resolve(old)...)
	}

	// Extend m.old, m.new in inherited mount point entries.
	for i := range mtpt {
		m := &mtpt[i]
		if m.old != old {
			if !hasPathPrefix(old, m.old) {
				// This should not happen.  If it does, panic so
				// that we can see the call trace that led to it.
				panic(fmt.Sprintf("invalid Bind: old=%q m={%q, %s, %q}", old, m.old, m.fs.String(), m.new))
			}
			suffix := old[len(m.old):]
			m.old = pathpkg.Join(m.old, suffix)
			m.new = pathpkg.Join(m.new, suffix)
		}
	}

	ns[old] = mtpt
}

// resolve resolves a path to the list of mountedFS to use for path.
func (ns NameSpace) resolve(path string) []mountedFS {
	path = ns.clean(path)
	for {
		if m := ns[path]; m != nil {
			if debugNS {
				fmt.Printf("resolve %s: %v\n", path, m)
			}
			return m
		}
		if path == "/" {
			break
		}
		path = pathpkg.Dir(path)
	}
	return nil
}

func (ns NameSpace) Open(ctx context.Context, path string) (ReadSeekCloser, error) {
	var err error
	for _, m := range ns.resolve(path) {
		if debugNS {
			fmt.Printf("tx %s: %v\n", path, m.translate(path))
		}
		tp := m.translate(path)
		r, err1 := m.fs.Open(ctx, tp)
		if err1 == nil {
			return r, nil
		}
		// IsNotExist errors in overlay FSes can mask real errors in
		// the underlying FS, so ignore them if there is another error.
		if err == nil || os.IsNotExist(err) {
			err = err1
		}
	}
	if err == nil {
		err = &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	return nil, err
}

// stat implements the FileSystem Stat and Lstat methods.
func (ns NameSpace) stat(ctx context.Context, path string, f func(FileSystem, context.Context, string) (os.FileInfo, error)) (os.FileInfo, error) {
	var err error
	for _, m := range ns.resolve(path) {
		fi, err1 := f(m.fs, ctx, m.translate(path))
		if err1 == nil {
			return fi, nil
		}
		if err == nil {
			err = err1
		}
	}

	// Check if path is an ancestor dir of a mount point.
	if err == nil || os.IsNotExist(err) && len(ns) > 0 {
		for old := range ns {
			if hasPathPrefix(old, path) && old != path {
				return dirInfo(pathpkg.Base(path)), nil
			}
		}
	}

	if err == nil {
		err = &os.PathError{Op: "stat", Path: path, Err: os.ErrNotExist}
	}
	return nil, err
}

func (ns NameSpace) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	return ns.stat(ctx, path, FileSystem.Stat)
}

func (ns NameSpace) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return ns.stat(ctx, path, FileSystem.Lstat)
}

func (ns NameSpace) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	path = ns.clean(path)

	var (
		haveName = map[string]bool{}
		all      []os.FileInfo
		err      error
	)

	for _, m := range ns.resolve(path) {
		dir, err1 := m.fs.ReadDir(ctx, m.translate(path))
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}

		if dir == nil {
			dir = []os.FileInfo{}
		}

		for _, d := range dir {
			name := d.Name()
			if !haveName[name] {
				haveName[name] = true
				all = append(all, d)
			}
		}
	}

	// Built union.  Add any missing directories needed to reach mount points.
	for old := range ns {
		if hasPathPrefix(old, path) && old != path {
			// Find next element after path in old.
			elem := old[len(path):]
			elem = strings.TrimPrefix(elem, "/")
			if i := strings.Index(elem, "/"); i >= 0 {
				elem = elem[:i]
			}
			if !haveName[elem] {
				haveName[elem] = true
				all = append(all, dirInfo(elem))
			}
		}
	}

	if len(all) == 0 {
		return nil, err
	}

	sort.Sort(byName(all))
	return all, nil
}

// byName implements sort.Interface.
type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
