// +build !windows

package git

import (
	"fmt"
	"os"
	"syscall"
)

func init() {
	const minAcceptable = 10000

	// HACK: Increase open file limit. Our file watchers open a lot of
	// file descriptors. On macOS, this is particularly necessary as
	// the default limits are generally lower than on Linux.
	var rlimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "# warning: failed to query open file limit:", err)
		return
	}
	rlimit.Cur = rlimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		fmt.Fprintln(os.Stderr, "# warning: failed to increase open file limit:", err)
	}

	// Confirm that the increase succeeded.
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		fmt.Fprintln(os.Stderr, "# warning: failed to query open file limit:", err)
		return
	}
	if rlimit.Cur != rlimit.Max && rlimit.Cur < minAcceptable {
		fmt.Fprintf(os.Stderr, "# warning: failed to increase open file limit to %d (current limit is %d)\n", rlimit.Max, rlimit.Cur)
	}
}
