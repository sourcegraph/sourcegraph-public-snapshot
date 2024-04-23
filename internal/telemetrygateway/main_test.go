package telemetrygateway

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
