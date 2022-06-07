package dependencies

import (
	"flag"
	"os"
	"testing"
)

var sgSetupTests = flag.String("sg-setup", "", "run sg setup tests for the designated platform")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var testArgs = CheckArgs{
	InRepo:     true,
	Teammate:   false,
	ConfigFile: "../../../sg.config.yaml",
}
