package logtest

import (
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/log"
)

func TestMain(m *testing.M) {
	Init(m, log.LevelDebug)
	os.Exit(m.Run())
}
