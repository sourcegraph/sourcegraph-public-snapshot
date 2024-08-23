package mountinfo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/sys/mountinfo"
	sglog "github.com/sourcegraph/log"
)

// defined as a variable so that it can be redefined by test routines
var findSysfsMountpoint = func() (mountpoint string, err error) {
	fsinfo := func(info *mountinfo.Info) (skip, stop bool) {
		if info.FSType == "sysfs" {
			return false, true
		}
		return true, false
	}
	info, err := mountinfo.GetMounts(fsinfo)
	if err == nil && len(info) == 0 {
		err = errors.New("findSysfsMountpoint: no sysfs mountpoint found")
	}
	if err != nil {
		return "", fmt.Errorf("findSysfsMountpoint: %w", err)
	}
	// the provided sysfs mountpoint could itself be a symlink, so we
	// resolve it immediately so that future file path
	// evaluations / massaging doesn't break
	cleanedPath, err := filepath.EvalSymlinks(filepath.Clean(info[0].Mountpoint))
	if err != nil {
		return "", fmt.Errorf("findSysfsMountpoint: verifying sysfs mountpoint %q: failed to resolve symlink: %w", info[0].Mountpoint, err)
	}
	return cleanedPath, nil
}

func discoverSysfsDevicePath(sysfsMountPoint string, deviceNumber string) (string, error) {

	// /sys/dev/block/<device_number> symlinks to /sys/devices/.../block/.../<deviceName>
	symlink := filepath.Join(sysfsMountPoint, "dev", "block", deviceNumber)

	devicePath, err := filepath.EvalSymlinks(symlink)
	if err != nil {
		return "", fmt.Errorf("discoverSysfsDevicePath: failed to evaluate sysfs symlink %q: %w", symlink, err)
	}

	devicePath, err = filepath.Abs(devicePath)
	if err != nil {
		return "", fmt.Errorf("discoverSysfsDevicePath: failed to massage device path %q to absolute path: %w", devicePath, err)
	}

	return devicePath, nil
}

func getDeviceBlockName(sysfsMountPoint, devicePath string) (string, error) {

	// Check to see if devicePath points to a disk partition. If so, we need to find the parent
	// device.

	// massage the sysfs folder name to ensure that it always ends in a '/'
	// so that strings.HasPrefix does what we expect when checking to see if
	// we're still under the /sys sub-folder
	sysFolderPrefix := strings.TrimSuffix(sysfsMountPoint, string(os.PathSeparator))
	sysFolderPrefix = sysFolderPrefix + string(os.PathSeparator)

	for {
		if !strings.HasPrefix(devicePath, sysFolderPrefix) {
			// ensure that we're still under the /sys/ sub-folder
			return "", fmt.Errorf("getDeviceBlockName: device path %q isn't a subpath of %q", devicePath, sysFolderPrefix)
		}

		_, err := os.Stat(filepath.Join(devicePath, "partition"))
		if errors.Is(err, os.ErrNotExist) {
			break
		}

		parent := filepath.Dir(devicePath)
		devicePath = parent
	}

	// If this device is a block device, its device path should have a symlink
	// to the block subsystem.

	subsystemPath, err := filepath.EvalSymlinks(filepath.Join(devicePath, "subsystem"))
	if err != nil {
		return "", fmt.Errorf("getDeviceBlockName: failed to discover subsystem that device (path %q) is part of: %w", devicePath, err)
	}

	if filepath.Base(subsystemPath) != "block" {
		return "", fmt.Errorf("getDeviceBlockName: device (path %q) is not part of the block subsystem", devicePath)
	}

	return filepath.Base(filepath.Base(devicePath)), nil
}

// discoverDeviceName returns the name of the block device that filePath is
// stored on.
func discoverDeviceName(logger sglog.Logger, filePath string) (string, error) {
	// Note: It's quite involved to implement the device discovery logic for
	// every possible kind of storage device (e.x. logical volumes, NFS, etc.) See
	// https://unix.stackexchange.com/a/11312 for more information.
	//
	// As a result, this logic will only work correctly for filePaths that are either:
	// - stored directly on a block device
	// - stored on a block device's partition
	//
	// For all other device types, this logic will either:
	// - return an incorrect device name
	// - return an error
	//
	// This logic was implemented from information gathered from the following sources (amongst others):
	// - "The Linux Programming Interface" by Michael Kerrisk: Chapter 14
	// - "Linux Kernel Development" by Robert Love: Chapters 13, 17
	// - https://man7.org/linux/man-pages/man5/sysfs.5.html
	// - https://en.wikipedia.org/wiki/Sysfs
	// - https://unix.stackexchange.com/a/11312
	// - https://www.kernel.org/doc/ols/2005/ols2005v1-pages-321-334.pdf

	sysfsMountPoint, err := findSysfsMountpoint()
	if err != nil {
		return "", fmt.Errorf("finding sysfs mountpoint: %w", err)
	}

	deviceNumber, err := getDeviceNumber(filepath.Clean(filePath))
	if err != nil {
		return "", fmt.Errorf("discovering device number: %w", err)
	}

	logger.Debug(
		"discovered device number",
		sglog.String("deviceNumber", deviceNumber),
	)

	devicePath, err := discoverSysfsDevicePath(sysfsMountPoint, deviceNumber)
	if err != nil {
		return "", fmt.Errorf("discovering device path: %w", err)
	}

	logger.Debug("discovered device path",
		sglog.String("devicePath", devicePath),
	)

	name, err := getDeviceBlockName(sysfsMountPoint, devicePath)
	if err != nil {
		return "", fmt.Errorf("failed resolving block device name: %w", err)
	}
	return name, nil
}
