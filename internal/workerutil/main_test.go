package workerutil

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestMain(m *testing.M) {
	// This package is INCREDIBLY noisy. We disable all logs during tests, regardless
	// of the `-v` flag, to save the noise in local development as well as CI.
	//
	// If logs are needed to debug unit test behavior, then set the log level argument
	// to the desired level.
	logtest.InitWithLevel(m, log.LevelNone)
}
