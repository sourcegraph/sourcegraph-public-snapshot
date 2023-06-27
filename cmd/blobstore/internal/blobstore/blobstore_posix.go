//go:build !windows
// +build !windows

package blobstore

func fsync(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
