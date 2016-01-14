// +build linux darwin

package sysreq

import (
	"fmt"
	"syscall"

	"golang.org/x/net/context"
)

func rlimitCheck(ctx context.Context) (*status, error) {
	var limit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit)
	if err != nil {
		return nil, err
	}
	const minLimit = 16384
	if limit.Cur < minLimit {
		return &status{
			problem: "Insufficient file descriptor limit",
			fix:     fmt.Sprintf(`Please increase the open file limit by running "ulimit -n %[1]d". On OS X you may need to first run "sudo launchctl limit maxfiles %[1]d".`, minLimit),
		}, nil
	}
	return nil, nil
}
