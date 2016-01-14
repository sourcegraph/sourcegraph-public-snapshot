// +build linux darwin

package sysreq

import (
	"fmt"
	"syscall"

	"golang.org/x/net/context"
)

func rlimitCheck(ctx context.Context) (*status, error) {
	const minLimit = 10000

	var limit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit)
	if err != nil {
		return nil, err
	}

	if limit.Cur < minLimit {
		return &status{
			problem: "Insufficient file descriptor limit",
			fix:     fmt.Sprintf(`Please increase the open file limit by running "ulimit -n %d".`, minLimit),
		}, nil
	}
	return nil, nil
}
