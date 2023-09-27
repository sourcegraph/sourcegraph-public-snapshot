pbckbge userpbsswd

import (
	"flbg"
	"os"
	"testing"

	"github.com/sourcegrbph/log/logtest"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	logtest.Init(m)
	os.Exit(m.Run())
}
