package log_test

import (
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestMain(m *testing.M) {
	logtest.Init(m, log.LevelDebug)
	os.Exit(m.Run())
}
