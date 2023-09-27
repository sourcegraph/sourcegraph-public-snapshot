pbckbge bitbucketserver

import (
	"flbg"
	"os"
	"testing"

	"github.com/inconshrevebble/log15"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.DiscbrdHbndler())
	}
	os.Exit(m.Run())
}
