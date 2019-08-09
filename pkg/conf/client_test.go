package conf

import (
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

func TestClient_continuouslyUpdate(t *testing.T) {
	t.Run("suppresses errors due to temporarily unreachable frontend", func(t *testing.T) {
		api.MockInternalClientConfiguration = func() (conftypes.RawUnified, error) {
			return conftypes.RawUnified{}, &url.Error{
				Op:  "Post",
				URL: "https://example.com",
				Err: &net.OpError{Op: "dial"},
			}
		}
		defer func() { api.MockInternalClientConfiguration = nil }()

		var client client
		var logMessages []string
		done := make(chan struct{})
		sleeps := 0
		const delayBeforeUnreachableLog = 150 * time.Millisecond // assumes first loop iter executes within this time period
		go client.continuouslyUpdate(&continuousUpdateOptions{
			delayBeforeUnreachableLog: delayBeforeUnreachableLog,
			log: func(format string, v ...interface{}) {
				logMessages = append(logMessages, fmt.Sprintf(format, v...))
			},
			sleep: func() {
				switch sleeps {
				case 0:
					if len(logMessages) > 0 {
						t.Errorf("got log messages (below), want no log before delayBeforeUnreachableLog\n\n%s", strings.Join(logMessages, "\n"))
					}
					time.Sleep(delayBeforeUnreachableLog)
					sleeps++
				case 1:
					if len(logMessages) != 1 {
						t.Errorf("got %d log messages, want exactly 1 log after delayBeforeUnreachableLog", len(logMessages))
					}

					// Exit goroutine after this test is done.
					close(done)
					runtime.Goexit()
				}
			},
		})
		<-done
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_717(size int) error {
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
