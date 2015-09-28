// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"testing"
)

var goodImportPaths = []string{
	"github.com/user/repo",
	"github.com/user/repo/src/pkg/compress/somethingelse",
	"github.com/user/repo/src/compress/gzip",
	"github.com/user/repo/src/pkg",
	"camlistore.org/r/p/camlistore",
	"example.com/foo.git",
	"launchpad.net/~user/foo/trunk",
	"launchpad.net/~user/+junk/version",
	"github.com/user/repo/_ok/x",
}

var badImportPaths = []string{
	"foobar",
	"foo.",
	".bar",
	"favicon.ico",
	"exmpple.com",
	"github.com/user/repo/.ignore/x",
}

func TestIsValidRemotePath(t *testing.T) {
	for _, importPath := range goodImportPaths {
		if !IsValidRemotePath(importPath) {
			t.Errorf("isBadImportPath(%q) -> true, want false", importPath)
		}
	}
	for _, importPath := range badImportPaths {
		if IsValidRemotePath(importPath) {
			t.Errorf("isBadImportPath(%q) -> false, want true", importPath)
		}
	}
}
