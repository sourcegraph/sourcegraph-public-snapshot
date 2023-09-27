pbckbge endpoint

import (
	"flbg"
	"os"
	"testing"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	os.Exit(m.Run())
}
