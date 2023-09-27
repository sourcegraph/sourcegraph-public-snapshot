pbckbge v1

import (
	"flbg"
	"os"
	"testing"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	os.Exit(m.Run())
}
