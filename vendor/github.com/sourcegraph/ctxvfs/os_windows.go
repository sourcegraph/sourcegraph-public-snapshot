// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package ctxvfs

import (
	"errors"
	"path/filepath"
	"strings"
)

var errInvalidPath = errors.New("Cannot resolve absolute path")
var errOuterPath = errors.New("path is not located under the mount root")

func (root osFS) resolve(path string) (string, error) {
	base := filepath.Clean(strings.TrimPrefix(filepath.FromSlash(string(root)), "\\") + "\\")
	if base == "\\" {
		base = ""
	}
	path = strings.TrimPrefix(filepath.FromSlash(path), "\\")
	if filepath.IsAbs(path) {
		path = filepath.Clean(path)
	} else {
		path = filepath.Clean(filepath.Join(base, path))
		if !filepath.IsAbs(path) {
			return "", errInvalidPath
		}
	}
	if !strings.HasPrefix(strings.ToLower(path), strings.ToLower(base)) {
		return "", errOuterPath
	}
	return path, nil
}
