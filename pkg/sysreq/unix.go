// +build linux darwin

package sysreq

import (
	"fmt"

	"context"

	"golang.org/x/sys/unix"
)

func rlimitCheck(ctx context.Context) (problem, fix string, err error) {
	const minLimit = 10000

	var limit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &limit); err != nil {
		return "", "", err
	}

	if limit.Cur < minLimit {
		return "Insufficient file descriptor limit", fmt.Sprintf(`Please increase the open file limit by running "ulimit -n %d".`, minLimit), nil
	}
	return
}
