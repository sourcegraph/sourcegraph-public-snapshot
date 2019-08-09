package redispool

import (
	"context"
	"time"

	"fmt"

	"github.com/sourcegraph/sourcegraph/pkg/sysreq"
)

func init() {
	sysreq.AddCheck("Redis", func(ctx context.Context) (problem, fix string, err error) {
		c := Store.Get()
		defer c.Close()

		timeout := 5 * time.Second
		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			_, err = c.Do("PING")
			if err == nil {
				// Success
				return "", "", nil
			}
			// Try again
			time.Sleep(100 * time.Millisecond)
		}
		return "Redis is unavailable or misconfigured",
			fmt.Sprintf("Start a Redis server listening at port %s", addrStore),
			err
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_873(size int) error {
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
