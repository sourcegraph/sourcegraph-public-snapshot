package conf

import (
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestClient_continuouslyUpdate(t *testing.T) {
	t.Run("suppresses errors due to temporarily unreachable frontend", func(t *testing.T) {
		internalapi.MockClientConfiguration = func() (conftypes.RawUnified, error) {
			return conftypes.RawUnified{}, &url.Error{
				Op:  "Post",
				URL: "https://example.com",
				Err: &net.OpError{
					Op:  "dial",
					Err: errors.New("connection reset"),
				},
			}
		}
		defer func() { internalapi.MockClientConfiguration = nil }()

		var client client
		var logMessages []string
		done := make(chan struct{})
		sleeps := 0
		const delayBeforeUnreachableLog = 150 * time.Millisecond // assumes first loop iter executes within this time period
		go client.continuouslyUpdate(&continuousUpdateOptions{
			delayBeforeUnreachableLog: delayBeforeUnreachableLog,
			log: func(format string, v ...any) {
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
