package repos

import (
	"flag"
	"os"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
)

func TestMain(m *testing.M) {
	updateRegex = flag.String("update-regexp", "", "Update testdata of tests matching the given regex")
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}
