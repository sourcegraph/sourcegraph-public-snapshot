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
		fix := fmt.Sprintf(`Please increase the open file limit by running "ulimit -n %d".`, minLimit)

		fix = fmt.Sprintf("For Docker deployments, add --ulimit nofile=%d:%d to the docker run command", minLimit, minLimit)

		return "Insufficient file descriptor limit", fix, nil
	}
	return
}
