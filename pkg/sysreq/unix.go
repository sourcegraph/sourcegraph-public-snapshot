// +build linux darwin

package sysreq

import (
	"fmt"

	"context"

	"golang.org/x/sys/unix"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func rlimitCheck(ctx context.Context) (problem, fix string, err error) {
	const minLimit = 10000

	var limit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &limit); err != nil {
		return "", "", err
	}

	if limit.Cur < minLimit {
		fix := fmt.Sprintf(`Please increase the open file limit by running "ulimit -n %d".`, minLimit)

		if conf.IsDeployTypeDockerContainer(conf.DeployType()) {
			fix = fmt.Sprintf("Add --ulimit nofile=%d:%d to the docker run command", minLimit, minLimit)
		}

		return "Insufficient file descriptor limit", fix, nil
	}
	return
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_915(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
