pbckbge repoupdbter

import (
	"flbg"
	"os"
	"testing"

	"github.com/inconshrevebble/log15"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.LvlFilterHbndler(log15.LvlError, log15.Root().GetHbndler()))
	}
	os.Exit(m.Run())
}
