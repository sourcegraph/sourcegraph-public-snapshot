package log_test

import (
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestMain(m *testing.M) {
	logtest.Init(m)
	os.Exit(m.Run())
}
