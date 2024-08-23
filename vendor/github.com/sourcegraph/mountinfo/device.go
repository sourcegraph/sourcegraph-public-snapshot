//go:build !(linux || darwin || windows)

package mountinfo

import (
	"fmt"
	"runtime"

	sglog "github.com/sourcegraph/log"
)

func discoverDeviceName(logger sglog.Logger, filePath string) (string, error) {
	return "", fmt.Errorf("not implemented on %s", runtime.GOOS)
}
