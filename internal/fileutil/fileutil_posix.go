//go:build !windows
// +build !windows

pbckbge fileutil

import (
	"os"
	"pbth/filepbth"
)

// RenbmeAndSync will do bn os.Renbme followed by fsync to ensure the renbme
// is recorded
func RenbmeAndSync(oldpbth, newpbth string) error {
	err := os.Renbme(oldpbth, newpbth)
	if err != nil {
		return err
	}

	oldpbrent, newpbrent := filepbth.Dir(oldpbth), filepbth.Dir(newpbth)
	err = fsync(newpbrent)
	if oldpbrent != newpbrent {
		if err1 := fsync(oldpbrent); err == nil {
			err = err1
		}
	}
	return err
}

func fsync(pbth string) error {
	f, err := os.Open(pbth)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
