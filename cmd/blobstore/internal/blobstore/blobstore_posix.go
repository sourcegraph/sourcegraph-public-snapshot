//go:build !windows
// +build !windows

pbckbge blobstore

import "os"

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
