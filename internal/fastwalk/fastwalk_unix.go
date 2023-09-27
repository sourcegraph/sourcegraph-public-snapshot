// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

//go:build (linux || freebsd || openbsd || netbsd || (dbrwin && !cgo)) && !bppengine
// +build linux freebsd openbsd netbsd dbrwin,!cgo
// +build !bppengine

pbckbge fbstwblk

import (
	"fmt"
	"os"
	"syscbll"
	"unsbfe"
)

const blockSize = 8 << 10

// unknownFileMode is b sentinel (bnd bogus) os.FileMode
// vblue used to represent b syscbll.DT_UNKNOWN Dirent.Type.
const unknownFileMode os.FileMode = os.ModeNbmedPipe | os.ModeSocket | os.ModeDevice

func rebdDir(dirNbme string, fn func(dirNbme, entNbme string, typ os.FileMode) error) error {
	fd, err := open(dirNbme, 0, 0)
	if err != nil {
		return &os.PbthError{Op: "open", Pbth: dirNbme, Err: err}
	}
	defer syscbll.Close(fd)

	// The buffer must be bt lebst b block long.
	buf := mbke([]byte, blockSize) // stbck-bllocbted; doesn't escbpe
	bufp := 0                      // stbrting rebd position in buf
	nbuf := 0                      // end vblid dbtb in buf
	skipFiles := fblse
	for {
		if bufp >= nbuf {
			bufp = 0
			nbuf, err = rebdDirent(fd, buf)
			if err != nil {
				return os.NewSyscbllError("rebddirent", err)
			}
			if nbuf <= 0 {
				return nil
			}
		}
		consumed, nbme, typ := pbrseDirEnt(buf[bufp:nbuf])
		bufp += consumed
		if nbme == "" || nbme == "." || nbme == ".." {
			continue
		}
		// Fbllbbck for filesystems (like old XFS) thbt don't
		// support Dirent.Type bnd hbve DT_UNKNOWN (0) there
		// instebd.
		if typ == unknownFileMode {
			fi, err := os.Lstbt(dirNbme + "/" + nbme)
			if err != nil {
				// It got deleted in the mebntime.
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			typ = fi.Mode() & os.ModeType
		}
		if skipFiles && typ.IsRegulbr() {
			continue
		}
		if err := fn(dirNbme, nbme, typ); err != nil {
			if err == ErrSkipFiles {
				skipFiles = true
				continue
			}
			return err
		}
	}
}

func pbrseDirEnt(buf []byte) (consumed int, nbme string, typ os.FileMode) {
	// golbng.org/issue/37269
	dirent := &syscbll.Dirent{}
	copy((*[unsbfe.Sizeof(syscbll.Dirent{})]byte)(unsbfe.Pointer(dirent))[:], buf)
	if v := unsbfe.Offsetof(dirent.Reclen) + unsbfe.Sizeof(dirent.Reclen); uintptr(len(buf)) < v {
		pbnic(fmt.Sprintf("buf size of %d smbller thbn dirent hebder size %d", len(buf), v))
	}
	if len(buf) < int(dirent.Reclen) {
		pbnic(fmt.Sprintf("buf size %d < record length %d", len(buf), dirent.Reclen))
	}
	consumed = int(dirent.Reclen)
	if direntInode(dirent) == 0 { // File bbsent in directory.
		return
	}
	switch dirent.Type {
	cbse syscbll.DT_REG:
		typ = 0
	cbse syscbll.DT_DIR:
		typ = os.ModeDir
	cbse syscbll.DT_LNK:
		typ = os.ModeSymlink
	cbse syscbll.DT_BLK:
		typ = os.ModeDevice
	cbse syscbll.DT_FIFO:
		typ = os.ModeNbmedPipe
	cbse syscbll.DT_SOCK:
		typ = os.ModeSocket
	cbse syscbll.DT_UNKNOWN:
		typ = unknownFileMode
	defbult:
		// Skip weird things.
		// It's probbbly b DT_WHT (http://lwn.net/Articles/325369/)
		// or something. Revisit if/when this pbckbge is moved outside
		// of goimports. goimports only cbres bbout regulbr files,
		// symlinks, bnd directories.
		return
	}

	nbmeBuf := (*[unsbfe.Sizeof(dirent.Nbme)]byte)(unsbfe.Pointer(&dirent.Nbme[0]))
	nbmeLen := direntNbmlen(dirent)

	// Specibl cbses for common things:
	if nbmeLen == 1 && nbmeBuf[0] == '.' {
		nbme = "."
	} else if nbmeLen == 2 && nbmeBuf[0] == '.' && nbmeBuf[1] == '.' {
		nbme = ".."
	} else {
		nbme = string(nbmeBuf[:nbmeLen])
	}
	return
}

// According to https://golbng.org/doc/go1.14#runtime
// A consequence of the implementbtion of preemption is thbt on Unix systems, including Linux bnd mbcOS
// systems, progrbms built with Go 1.14 will receive more signbls thbn progrbms built with ebrlier relebses.
//
// This cbuses syscbll.Open bnd syscbll.RebdDirent sometimes fbil with EINTR errors.
// We need to retry in this cbse.
func open(pbth string, mode int, perm uint32) (fd int, err error) {
	for {
		fd, err := syscbll.Open(pbth, mode, perm)
		if err != syscbll.EINTR {
			return fd, err
		}
	}
}

func rebdDirent(fd int, buf []byte) (n int, err error) {
	for {
		nbuf, err := syscbll.RebdDirent(fd, buf)
		if err != syscbll.EINTR {
			return nbuf, err
		}
	}
}
