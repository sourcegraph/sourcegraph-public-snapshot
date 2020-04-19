package diskutil

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

// findMountPoint searches upwards starting from the directory d to find the mount point.
func findMountPoint(d string) (string, error) {
	d, err := filepath.Abs(d)
	if err != nil {
		return "", errors.Wrapf(err, "getting absolute version of %s", d)
	}

	for {
		m, err := isMount(d)
		if err != nil {
			return "", errors.Wrapf(err, "finding out if %s is a mount point", d)
		}
		if m {
			return d, nil
		}

		parent := filepath.Dir(d)
		if parent == d {
			return parent, nil
		}
		d = parent
	}
}

// isMount tells whether the directory d is a mount point.
func isMount(d string) (bool, error) {
	ddev, err := device(d)
	if err != nil {
		return false, errors.Wrapf(err, "getting device id for %s", d)
	}

	parent := filepath.Dir(d)
	if parent == d {
		// root of filesystem
		return true, nil
	}

	pdev, err := device(parent)
	if err != nil {
		return false, errors.Wrapf(err, "getting device id for %s", parent)
	}

	// root of device
	return pdev != ddev, nil
}

// device gets the device id of a file f.
func device(f string) (int64, error) {
	fi, err := os.Stat(f)
	if err != nil {
		return 0, errors.Wrapf(err, "running stat on %s", f)
	}
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get stat details for %s", f)
	}
	return int64(stat.Dev), nil
}
