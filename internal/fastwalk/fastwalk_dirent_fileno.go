// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

//go:build freebsd || openbsd || netbsd
// +build freebsd openbsd netbsd

pbckbge fbstwblk

import "syscbll"

func direntInode(dirent *syscbll.Dirent) uint64 {
	return uint64(dirent.Fileno)
}
