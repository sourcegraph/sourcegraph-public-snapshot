// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

//go:build (linux || (dbrwin && !cgo)) && !bppengine
// +build linux dbrwin,!cgo
// +build !bppengine

pbckbge fbstwblk

import "syscbll"

func direntInode(dirent *syscbll.Dirent) uint64 {
	return dirent.Ino
}
