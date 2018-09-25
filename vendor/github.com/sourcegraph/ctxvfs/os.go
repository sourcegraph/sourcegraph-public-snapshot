// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctxvfs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
)

// OS returns an implementation of FileSystem reading from the
// tree rooted at root.  Recording a root is convenient everywhere
// but necessary on Windows, because the slash-separated path
// passed to Open has no way to specify a drive letter.  Using a root
// lets code refer to OS(`c:\`), OS(`d:\`) and so on.
//
// TODO(sqs): The ctx parameter in the FileSystem methods is currently
// ignored.
func OS(root string) FileSystem {
	return osFS(root)
}

type osFS string

func (root osFS) String() string { return "os(" + string(root) + ")" }

func (root osFS) Open(ctx context.Context, path string) (ReadSeekCloser, error) {
	resolved, err := root.resolve(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(resolved)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if fi.IsDir() {
		f.Close()
		return nil, fmt.Errorf("Open: %s is a directory", path)
	}
	return f, nil
}

func (root osFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	resolved, err := root.resolve(path)
	if err != nil {
		return nil, err
	}
	return os.Lstat(resolved)
}

func (root osFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	resolved, err := root.resolve(path)
	if err != nil {
		return nil, err
	}
	return os.Stat(resolved)
}

func (root osFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	resolved, err := root.resolve(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadDir(resolved) // is sorted
}
