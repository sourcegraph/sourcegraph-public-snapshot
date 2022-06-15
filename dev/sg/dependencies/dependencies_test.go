package dependencies

import (
	"flag"
	"os"
	"testing"
)

// WARNING: These tests attempt to modify your system! Run with care or exclusively in CI.
var sgSetupTests = flag.String("sg-setup", "", "run sg setup tests for the designated platform")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var testArgs = CheckArgs{
	Teammate:            false,
	ConfigFile:          "../../../sg.config.yaml",
	ConfigOverwriteFile: "../../../sg.config.overwrite.yaml",
}
