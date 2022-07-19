package conf

import (
	"net"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestClientContinuouslyUpdate(t *testing.T) {
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
		logger, exportLogs := logtest.Captured(t)
		done := make(chan struct{})
		sleeps := 0
		const delayBeforeUnreachableLog = 150 * time.Millisecond // assumes first loop iter executes within this time period
		go client.continuouslyUpdate(&continuousUpdateOptions{
			delayBeforeUnreachableLog: delayBeforeUnreachableLog,
			logger:                    logger,
			sleepBetweenUpdates: func() {
				logMessages := exportLogs()
				switch sleeps {
				case 0:
					for _, message := range logMessages {
						require.NotEqual(t, message.Level, log.LevelError, "expected no error messages before delayBeforeUnreachableLog")
					}
					sleeps++
					time.Sleep(delayBeforeUnreachableLog)
				case 1:
					require.Len(t, logMessages, 2)
					assert.Contains(t, logMessages[0].Message, "checking")
					assert.Contains(t, logMessages[1].Message, "received error")

					// Exit goroutine after this test is done.
					close(done)
					runtime.Goexit()
				}
			},
		})
		<-done
	})
}
