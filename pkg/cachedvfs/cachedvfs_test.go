package cachedvfs

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/golang/groupcache/lru"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

type mapCache struct {
	m map[lru.Key]interface{}
}

func newMapCache() *mapCache {
	return &mapCache{map[lru.Key]interface{}{}}
}

func (m *mapCache) Get(k lru.Key) (interface{}, bool) {
	v, ok := m.m[k]
	return v, ok
}
func (m *mapCache) Add(k lru.Key, v interface{}) { m.m[k] = v }

func TestHit(t *testing.T) {
	fs := mapfs.New(map[string]string{"foo": "bar"})
	cache := newMapCache()
	cfs := New(fs, cache, regexp.MustCompile("/foo"))
	assertFSEqual(t, fs, cfs)
	assertFSEqual(t, fs, cfs)
	if len(cache.m) != 3 {
		t.Errorf("Expected 3 items in cache (Open, Lstat, Stat, ReadDir): %v", cache.m)
	}
}

func TestReadDir(t *testing.T) {
	fs := mapfs.New(map[string]string{"foo": "bar"})
	cache := newMapCache()
	cfs := New(fs, cache, regexp.MustCompile("^/$"))
	assertFSEqual(t, fs, cfs)
	assertFSEqual(t, fs, cfs)
	if len(cache.m) != 1 {
		t.Errorf("Expected 1 items in cache (ReadDir): %v", cache.m)
	}
}

func TestMiss(t *testing.T) {
	fs := mapfs.New(map[string]string{"foo": "bar"})
	cache := newMapCache()
	cfs := New(fs, cache, regexp.MustCompile(".*bar.*"))
	assertFSEqual(t, fs, cfs)
	assertFSEqual(t, fs, cfs)
	if len(cache.m) != 0 {
		t.Errorf("Expected 0 items in cache: %v", cache.m)
	}
}

func assertFSEqual(t *testing.T, a, b vfs.FileSystem) {
	dirs := []string{"/"}
	for len(dirs) != 0 {
		d := dirs[0]
		dirs = dirs[1:]
		afis, err := a.ReadDir(d)
		if err != nil {
			t.Fatal(err)
		}
		bfis, err := b.ReadDir(d)
		if err != nil {
			t.Fatal(err)
		}
		if len(afis) != len(bfis) {
			t.Errorf("%s has different length fileinfos", d)
		}
		for i, fi := range afis {
			p := filepath.Join(d, fi.Name())
			assertFileInfoEqual(t, afis[i], bfis[i])
			assertPathEqual(t, p, a, b)
			if fi.IsDir() {
				dirs = append(dirs, p)
			}
		}
	}
}

func assertFileInfoEqual(t *testing.T, a, b os.FileInfo) {
	if a == nil || b == nil {
		if a != b {
			t.Errorf("os.FileInfo are both not nil %v != %v", a, b)
		}
		return
	}
	if a.Name() != b.Name() {
		t.Errorf("os.FileInfo.Name: %v != %v", a.Name(), b.Name())
	}
	if a.Size() != b.Size() {
		t.Errorf("os.FileInfo.Size: %v != %v", a.Size(), b.Size())
	}
	if a.Mode() != b.Mode() {
		t.Errorf("os.FileInfo.Mode: %v != %v", a.Mode(), b.Mode())
	}
	if a.ModTime() != b.ModTime() {
		t.Errorf("os.FileInfo.ModTime: %v != %v", a.ModTime(), b.ModTime())
	}
	if a.IsDir() != b.IsDir() {
		t.Errorf("os.FileInfo.IsDir: %v != %v", a.IsDir(), b.IsDir())
	}
	// We purposefully ignore Sys
}

func assertPathEqual(t *testing.T, p string, a, b vfs.FileSystem) {
	as, err := a.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := b.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	assertFileInfoEqual(t, as, bs)
	if as.IsDir() {
		return
	}

	as, err = a.Lstat(p)
	if err != nil {
		t.Fatal(err)
	}
	bs, err = b.Lstat(p)
	if err != nil {
		t.Fatal(err)
	}
	assertFileInfoEqual(t, as, bs)

	ab, err := vfs.ReadFile(a, p)
	if err != nil {
		t.Fatal(err)
	}
	bb, err := vfs.ReadFile(b, p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ab, bb) {
		t.Errorf("%s does not have equal file contents %v != %v", p, ab, bb)
	}
}
