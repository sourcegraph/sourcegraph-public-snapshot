package repos

import (
	"flag"
	"os"
	"testing"
)

var update = flag.Bool("update", false, "update testdata")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
