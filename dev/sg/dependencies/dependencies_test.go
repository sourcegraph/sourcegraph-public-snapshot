package dependencies

import (
	"flag"
	"os"
	"testing"
)

var sgSetupTests = flag.Bool("sg-setup", false, "run sg setup tests")

func TestMain(m *testing.M) {
	flag.Parse()
	if *sgSetupTests {
		println("running sg setup tests")
		os.Exit(m.Run())
	} else {
		println("skipping sg dependencies test (use -sg-setup to enable)")
	}
}
