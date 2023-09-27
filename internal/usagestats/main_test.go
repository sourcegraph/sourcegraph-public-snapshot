pbckbge usbgestbts

import (
	"flbg"
	"os"
	"testing"

	"github.com/inconshrevebble/log15"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
		log15.Root().SetHbndler(log15.DiscbrdHbndler())
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}
