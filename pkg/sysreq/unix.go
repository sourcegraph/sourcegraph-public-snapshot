// +build linux darwin

package sysreq

import (
	"fmt"

	"golang.org/x/net/context"
	"golang.org/x/sys/unix"
)

func rlimitCheck(ctx context.Context) (*status, error) {
	const minLimit = 10000

	var limit unix.Rlimit
	err := unix.Getrlimit(unix.RLIMIT_NOFILE, &limit)
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
