package mountinfo

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	sglog "github.com/sourcegraph/log"
)

// discoverDeviceName returns the name of the block device that filePath is
// stored on.
func discoverDeviceName(logger sglog.Logger, filePath string) (string, error) {
	// on macOS (darwin), use the `stat` and `diskutil` OS tools
	// diskutil info $(stat -f '%Sd' <path>) | grep 'Part of Whole:' | awk '{print $NF}'

	// macOS does support using the `unix.Stat_t` struct and `unix.Stat` function,
	// but finding the device identifier name from the major + minor idendifiers proved difficult,
	// so just use `stat` to print out the partition identifier name, and `diskutil`
	// to find the disk identifier name from that

	filePath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s: %w", filePath, err)
	}
	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s: %w", filePath, err)
	}
	stat, err := exec.Command("/usr/bin/stat", "-f", "%Sd", filePath).Output()
	if err != nil {
		return "", fmt.Errorf("unable to stat %s: %w", filePath, err)
	}

	diskinfo, err := exec.Command("/usr/sbin/diskutil", "info", strings.TrimSpace(string(stat))).CombinedOutput()
	if err != nil {
		// log the output from `diskutil` instead of including it in the error message because it may be multiline
		logger.Error(fmt.Sprintf("unable to get disk info on %s. Output is (%s)", string(stat), string(diskinfo)))
		return "", fmt.Errorf("unable to get disk info on %s: %w", string(stat), err)
	}

	regex := regexp.MustCompile("Part of Whole:[ \t]+(?P<name>\\w+)")
	match := regex.FindSubmatch(diskinfo)
	if match == nil {
		// log the output from `diskutil` instead of including it in the error message because it may be multiline
		logger.Error(fmt.Sprintf("unable to find disk info in (%s)", string(diskinfo)))
		return "", fmt.Errorf("unable to find disk info on %s: %w", string(stat), err)
	}

	return string(match[1]), nil
}
