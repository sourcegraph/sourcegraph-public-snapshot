// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctxvfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestMap_openRoot(t *testing.T) {
	fs := Map(map[string][]byte{
		"foo/bar/three.txt": []byte("a"),
		"foo/bar.txt":       []byte("b"),
		"top.txt":           []byte("c"),
		"other-top.txt":     []byte("d"),
	})
	tests := []struct {
		path string
		want string
	}{
		{"/foo/bar/three.txt", "a"},
		{"foo/bar/three.txt", "a"},
		{"foo/bar.txt", "b"},
		{"top.txt", "c"},
		{"/top.txt", "c"},
		{"other-top.txt", "d"},
		{"/other-top.txt", "d"},
	}
	for _, tt := range tests {
		rsc, err := fs.Open(nil, tt.path)
		if err != nil {
			t.Errorf("Open(%q) = %v", tt.path, err)
			continue
		}
		slurp, err := ioutil.ReadAll(rsc)
		if err != nil {
			t.Error(err)
		}
		if string(slurp) != tt.want {
			t.Errorf("Read(%q) = %q; want %q", tt.path, tt.want, slurp)
		}
		rsc.Close()
	}

	_, err := fs.Open(nil, "/xxxx")
	if !os.IsNotExist(err) {
		t.Errorf("ReadDir /xxxx = %v; want os.IsNotExist error", err)
	}
}

func TestMap_Stat(t *testing.T) {
	fs := Map(map[string][]byte{
		"foo/bar/three.txt": []byte("333"),
		"foo/bar.txt":       []byte("22"),
		"top.txt":           []byte("top.txt file"),
		"other-top.txt":     []byte("other-top.txt file"),
	})
	tests := []struct {
		path string
		want os.FileInfo
	}{
		{path: "", want: dirInfo("/")},
		{path: "foo", want: dirInfo("foo")},
		{path: "foo/", want: dirInfo("foo")},
		{path: "foo/bar", want: dirInfo("bar")},
		{path: "foo/bar/", want: dirInfo("bar")},
		{path: "foo/bar/three.txt", want: fileInfo{"three.txt", 3}},
	}
	for _, leadingSlashOrEmpty := range []string{"", "/"} {
		for _, tt := range tests {
			path := leadingSlashOrEmpty + tt.path
			if path == "" {
				continue
			}
			fi, err := fs.Stat(nil, path)
			if err != nil {
				t.Errorf("Stat(%q) = %v", path, err)
				continue
			}
			if !reflect.DeepEqual(fi, tt.want) {
				t.Errorf("Stat(%q) = %#v; want %#v", path, fi, tt.want)
				continue
			}
		}
	}

	_, err := fs.Stat(nil, "/xxxx")
	if !os.IsNotExist(err) {
		t.Errorf("Stat /xxxx = %v; want os.IsNotExist error", err)
	}
}

func TestMap_ReadDir(t *testing.T) {
	fs := Map(map[string][]byte{
		"foo/bar/three.txt": []byte("333"),
		"foo/bar.txt":       []byte("22"),
		"top.txt":           []byte("top.txt file"),
		"other-top.txt":     []byte("other-top.txt file"),
	})
	tests := []struct {
		dir  string
		want []os.FileInfo
	}{
		{
			dir: "",
			want: []os.FileInfo{
				dirInfo("foo"),
				fileInfo{"other-top.txt", int64(len("other-top.txt file"))},
				fileInfo{"top.txt", int64(len("top.txt file"))},
			},
		},
		{
			dir: "foo",
			want: []os.FileInfo{
				dirInfo("bar"),
				fileInfo{"bar.txt", 2},
			},
		},
		{
			dir: "foo/",
			want: []os.FileInfo{
				dirInfo("bar"),
				fileInfo{"bar.txt", 2},
			},
		},
		{
			dir: "foo/bar",
			want: []os.FileInfo{
				fileInfo{"three.txt", 3},
			},
		},
	}
	for _, leadingSlashOrEmpty := range []string{"", "/"} {
		for _, tt := range tests {
			path := leadingSlashOrEmpty + tt.dir
			if path == "" {
				continue
			}
			fis, err := fs.ReadDir(nil, path)
			if err != nil {
				t.Errorf("ReadDir(%q) = %v", path, err)
				continue
			}
			if !reflect.DeepEqual(fis, tt.want) {
				t.Errorf("ReadDir(%q) = %#v; want %#v", path, fis, tt.want)
				continue
			}
		}

		if _, err := fs.ReadDir(nil, leadingSlashOrEmpty+"xxxx"); !os.IsNotExist(err) {
			t.Errorf("ReadDir %q = %v; want os.IsNotExist error", leadingSlashOrEmpty+"xxxx", err)
		}
	}
}

// BenchmarkReadDir-12    	    1000	   1467196 ns/op	  263818 B/op	    1090 allocs/op
//
// removing sort.Strings(ents) improves speed by ~15% but probably is not worth it
func BenchmarkMap_ReadDir(b *testing.B) {
	m := map[string][]byte{}
	fs := Map(m)

	addDirFiles := func(dir string, numFiles int) {
		for i := 0; i < numFiles; i++ {
			m[fmt.Sprintf("%s/f%d", dir, i)] = []byte("")
		}
	}
	const filesPerDir = 1000
	addDirFiles("/a", filesPerDir)
	addDirFiles("/a/b", filesPerDir)
	addDirFiles("/a/c", filesPerDir)
	addDirFiles("/a/b/c", filesPerDir)
	addDirFiles("/a/b/d", filesPerDir)
	addDirFiles("/a/b/d/e", filesPerDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := fs.ReadDir(nil, "/a"); err != nil {
			b.Error(err)
		}
	}
}
