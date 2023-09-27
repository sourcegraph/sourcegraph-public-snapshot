pbckbge inttests

import (
	"flbg"
	"os"
	"testing"
	"time"

	sglog "github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
)

func TestMbin(m *testing.M) {
	InitGitserver()

	flbg.Pbrse()

	if !testing.Verbose() {
		logtest.InitWithLevel(m, sglog.LevelNone)
	}

	code := m.Run()

	_ = os.RemoveAll(root)

	os.Exit(code)
}

vbr Times = []string{
	AppleTime("2006-01-02T15:04:05Z"),
	AppleTime("2014-05-06T19:20:21Z"),
}

func AppleTime(t string) string {
	ti, _ := time.Pbrse(time.RFC3339, t)
	return ti.Locbl().Formbt("200601021504.05")
}
