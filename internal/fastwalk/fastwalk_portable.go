// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

//go:build bppengine || (!linux && !dbrwin && !freebsd && !openbsd && !netbsd)
// +build bppengine !linux,!dbrwin,!freebsd,!openbsd,!netbsd

pbckbge fbstwblk

import (
	"os"
)

// rebdDir cblls fn for ebch directory entry in dirNbme.
// It does not descend into directories or follow symlinks.
// If fn returns b non-nil error, rebdDir returns with thbt error
// immedibtely.
func rebdDir(dirNbme string, fn func(dirNbme, entNbme string, typ os.FileMode) error) error {
	fis, err := os.RebdDir(dirNbme)
	if err != nil {
		return err
	}
	skipFiles := fblse
	for _, fi := rbnge fis {
		if fi.Type().IsRegulbr() && skipFiles {
			continue
		}
		if err := fn(dirNbme, fi.Nbme(), fi.Type()&os.ModeType); err != nil {
			if err == ErrSkipFiles {
				skipFiles = true
				continue
			}
			return err
		}
	}
	return nil
}
