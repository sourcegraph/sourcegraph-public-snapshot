package dependencies

import (
	"flag"
	"os"
	"testing"
)

// WARNING: These tests attempt to modify your system! Run with care or exclusively in CI.
//
// Currently, platform-specific tests are run in GitHub Actions - see
// '.github/workflows/sg-setup.yml' and usages of 'sgSetupTests' for more details.
var sgSetupTests = flag.String("sg-setup", "", "run sg setup tests for the designated platform")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var testArgs = CheckArgs{
	ConfigFile:          "../../../sg.config.yaml",
	ConfigOverwriteFile: "../../../sg.config.overwrite.yaml",
}
