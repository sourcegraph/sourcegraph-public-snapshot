// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mapfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestOpenRoot(t *testing.T) {
	fs := New(map[string][]byte{
		"/foo/bar/three.txt": []byte("a"),
		"/foo/bar.txt":       []byte("b"),
		"/top.txt":           []byte("c"),
		"/other-top.txt":     []byte("d"),
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
		rsc, err := fs.Open(tt.path)
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

	_, err := fs.Open("/xxxx")
	if !os.IsNotExist(err) {
		t.Errorf("ReadDir /xxxx = %v; want os.IsNotExist error", err)
	}
}

func TestStat(t *testing.T) {
	fs := New(map[string][]byte{
		"/foo/bar/three.txt": []byte("333"),
		"/foo/bar.txt":       []byte("22"),
		"/top.txt":           []byte("top.txt file"),
		"/other-top.txt":     []byte("other-top.txt file"),
	})
	tests := []struct {
		path string
		want os.FileInfo
	}{
		{path: "/", want: mapFI{name: "/", dir: true}},
		{path: "/foo", want: mapFI{name: "foo", dir: true}},
		{path: "/foo/", want: mapFI{name: "foo", dir: true}},
		{path: "/foo/bar", want: mapFI{name: "bar", dir: true}},
		{path: "/foo/bar/", want: mapFI{name: "bar", dir: true}},
		{path: "/foo/bar/three.txt", want: mapFI{name: "three.txt", size: 3}},
	}
	for _, tt := range tests {
		fi, err := fs.Stat(tt.path)
		if err != nil {
			t.Errorf("Stat(%q) = %v", tt.path, err)
			continue
		}
		if !reflect.DeepEqual(fi, tt.want) {
			t.Errorf("Stat(%q) = %#v; want %#v", tt.path, fi, tt.want)
			continue
		}
	}

	_, err := fs.Stat("/xxxx")
	if !os.IsNotExist(err) {
		t.Errorf("Stat /xxxx = %v; want os.IsNotExist error", err)
	}
}

func TestReadDir(t *testing.T) {
	fs := New(map[string][]byte{
		"/foo/bar/three.txt": []byte("333"),
		"/foo/bar.txt":       []byte("22"),
		"/top.txt":           []byte("top.txt file"),
		"/other-top.txt":     []byte("other-top.txt file"),
	})
	tests := []struct {
		dir  string
		want []os.FileInfo
	}{
		{
			dir: "/",
			want: []os.FileInfo{
				mapFI{name: "foo", dir: true},
				mapFI{name: "other-top.txt", size: len("other-top.txt file")},
				mapFI{name: "top.txt", size: len("top.txt file")},
			},
		},
		{
			dir: "/foo",
			want: []os.FileInfo{
				mapFI{name: "bar", dir: true},
				mapFI{name: "bar.txt", size: 2},
			},
		},
		{
			dir: "/foo/",
			want: []os.FileInfo{
				mapFI{name: "bar", dir: true},
				mapFI{name: "bar.txt", size: 2},
			},
		},
		{
			dir: "/foo/bar",
			want: []os.FileInfo{
				mapFI{name: "three.txt", size: 3},
			},
		},
	}
	for _, tt := range tests {
		fis, err := fs.ReadDir(tt.dir)
		if err != nil {
			t.Errorf("ReadDir(%q) = %v", tt.dir, err)
			continue
		}
		if !reflect.DeepEqual(fis, tt.want) {
			t.Errorf("ReadDir(%q) = %#v; want %#v", tt.dir, fis, tt.want)
			continue
		}
	}

	_, err := fs.ReadDir("/xxxx")
	if !os.IsNotExist(err) {
		t.Errorf("ReadDir /xxxx = %v; want os.IsNotExist error", err)
	}
}

// BenchmarkReadDir-12    	    1000	   1467196 ns/op	  263818 B/op	    1090 allocs/op
//
// removing sort.Strings(ents) improves speed by ~15% but probably is not worth it
func BenchmarkReadDir(b *testing.B) {
	m := map[string][]byte{}
	fs := New(m)

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
		if _, err := fs.ReadDir("/a"); err != nil {
			b.Error(err)
		}
	}
}
