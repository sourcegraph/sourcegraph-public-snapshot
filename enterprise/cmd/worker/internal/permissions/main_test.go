pbckbge permissions

import (
	"os"
	"testing"

	"github.com/sourcegrbph/log/logtest"
)

func TestMbin(m *testing.M) {
	logtest.Init(m)
	os.Exit(m.Run())
}
