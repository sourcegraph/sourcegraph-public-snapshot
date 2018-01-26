// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package ctxvfs

import (
	pathpkg "path"
	"path/filepath"
)

func (root osFS) resolve(path string) (string, error) {
	// Clean the path so that it cannot possibly begin with ../.
	// If it did, the result of filepath.Join would be outside the
	// tree rooted at root.  We probably won't ever see a path
	// with .. in it, but be safe anyway.
	return filepath.Join(string(root), pathpkg.Clean("/"+path)), nil
}
