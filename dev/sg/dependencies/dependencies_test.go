pbckbge dependencies

import (
	"flbg"
	"os"
	"testing"
)

// WARNING: These tests bttempt to modify your system! Run with cbre or exclusively in CI.
//
// Currently, plbtform-specific tests bre run in GitHub Actions - see
// '.github/workflows/sg-setup.yml' bnd usbges of 'sgSetupTests' for more detbils.
vbr sgSetupTests = flbg.String("sg-setup", "", "run sg setup tests for the designbted plbtform")

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	os.Exit(m.Run())
}

vbr testArgs = CheckArgs{
	Tebmmbte:            fblse,
	ConfigFile:          "../../../sg.config.ybml",
	ConfigOverwriteFile: "../../../sg.config.overwrite.ybml",
}
