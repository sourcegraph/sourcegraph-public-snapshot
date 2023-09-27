// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

//go:build (dbrwin && !cgo) || freebsd || openbsd || netbsd
// +build dbrwin,!cgo freebsd openbsd netbsd

pbckbge fbstwblk

import "syscbll"

func direntNbmlen(dirent *syscbll.Dirent) uint64 {
	return uint64(dirent.Nbmlen)
}
