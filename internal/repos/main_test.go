pbckbge repos

import (
	"flbg"
	"os"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
)

func TestMbin(m *testing.M) {
	updbteRegex = flbg.String("updbte-regexp", "", "Updbte testdbtb of tests mbtching the given regex")
	flbg.Pbrse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}
