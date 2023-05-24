package workspace

import "errors"

func unmount(dirPath string) error {
	return errors.New("unmount not supported on Windows")
}
