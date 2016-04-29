// Package cachedvfs provides a vfs which caches responses
package cachedvfs

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cache"
)

// CachedVFS is a wrapper around an underlying FS. If a read operation happens
// on something which matches PathRe, it will use the Cache to Get or Add the
// return value. Note: None of the operations cache error responses.
type CachedVFS struct {
	FS     vfs.FileSystem
	Cache  cache.Cache
	PathRe *regexp.Regexp
}

type key struct {
	Path string
	Op   string
}

// New constructs a new CachedVFS
func New(fs vfs.FileSystem, c cache.Cache, pathRe *regexp.Regexp) vfs.FileSystem {
	return &CachedVFS{fs, c, pathRe}
}

// See vfs.FileSystem godoc
func (v *CachedVFS) Open(name string) (vfs.ReadSeekCloser, error) {
	canCache := v.PathRe.MatchString(name)
	if !canCache {
		return v.FS.Open(name)
	}
	k := key{Path: name, Op: "open"}
	if data, ok := v.Cache.Get(k); ok {
		return download{bytes.NewReader(data.([]byte))}, nil
	}
	r, err := v.FS.Open(name)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	v.Cache.Add(k, data)
	return download{bytes.NewReader(data)}, nil
}

// See vfs.FileSystem godoc
func (v *CachedVFS) Lstat(path string) (os.FileInfo, error) {
	canCache := v.PathRe.MatchString(path)
	if !canCache {
		return v.FS.Lstat(path)
	}
	k := key{Path: path, Op: "lstat"}
	if fi, ok := v.Cache.Get(k); ok {
		return fi.(os.FileInfo), nil
	}
	fi, err := v.FS.Lstat(path)
	fi = newFileInfo(fi)
	if err == nil {
		v.Cache.Add(k, fi)
	}
	return fi, err
}

// See vfs.FileSystem godoc
func (v *CachedVFS) Stat(path string) (os.FileInfo, error) {
	canCache := v.PathRe.MatchString(path)
	if !canCache {
		return v.FS.Stat(path)
	}
	k := key{Path: path, Op: "stat"}
	if fi, ok := v.Cache.Get(k); ok {
		return fi.(os.FileInfo), nil
	}
	fi, err := v.FS.Stat(path)
	fi = newFileInfo(fi)
	if err == nil {
		v.Cache.Add(k, fi)
	}
	return fi, err
}

// See vfs.FileSystem godoc
func (v *CachedVFS) ReadDir(path string) ([]os.FileInfo, error) {
	canCache := v.PathRe.MatchString(path)
	if !canCache {
		return v.FS.ReadDir(path)
	}
	k := key{Path: path, Op: "readdir"}
	if fi, ok := v.Cache.Get(k); ok {
		return fi.([]os.FileInfo), nil
	}
	fis, err := v.FS.ReadDir(path)
	fis = newFileInfos(fis)
	if err == nil {
		v.Cache.Add(k, fis)
	}
	return fis, err
}

// See vfs.FileSystem godoc
func (v *CachedVFS) String() string {
	return "cachedvfs(" + v.FS.String() + ")"
}
