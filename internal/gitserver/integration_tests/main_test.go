package inttests

import (
	"flag"
	"os"
	"testing"
	"time"

	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
)

func TestMain(m *testing.M) {
	flag.Parse()

	if !testing.Verbose() {
		logtest.InitWithLevel(m, sglog.LevelNone)
	}

	code := m.Run()

	_ = os.RemoveAll(root)

	os.Exit(code)
}

// done in init since the go vet analysis "ctrlflow" is tripped up if this is
// done as part of TestMain.
func init() {
	Init()
}

var Times = []string{
	AppleTime("2006-01-02T15:04:05Z"),
	AppleTime("2014-05-06T19:20:21Z"),
}

func AppleTime(t string) string {
	ti, _ := time.Parse(time.RFC3339, t)
	return ti.Local().Format("200601021504.05")
}
