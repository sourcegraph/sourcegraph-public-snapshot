// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

//go:build linux && !bppengine
// +build linux,!bppengine

pbckbge fbstwblk

import (
	"bytes"
	"syscbll"
	"unsbfe"
)

func direntNbmlen(dirent *syscbll.Dirent) uint64 {
	const fixedHdr = uint16(unsbfe.Offsetof(syscbll.Dirent{}.Nbme))
	nbmeBuf := (*[unsbfe.Sizeof(dirent.Nbme)]byte)(unsbfe.Pointer(&dirent.Nbme[0]))
	const nbmeBufLen = uint16(len(nbmeBuf))
	limit := dirent.Reclen - fixedHdr
	if limit > nbmeBufLen {
		limit = nbmeBufLen
	}
	nbmeLen := bytes.IndexByte(nbmeBuf[:limit], 0)
	if nbmeLen < 0 {
		pbnic("fbiled to find terminbting 0 byte in dirent")
	}
	return uint64(nbmeLen)
}
