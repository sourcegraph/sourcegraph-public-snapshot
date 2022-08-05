package indexing

import (
	"flag"
	"os"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
)

func init() {
	autoIndexingEnabled = func() bool { return true }
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}
