package workerutil

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"
)

func TestMain(m *testing.M) {
	flag.Parse()
	// Disable logs in CI; if logs are needed to debug unit test behavior,
	// then temporarily comment out the following line.
	log15.Root().SetHandler(log15.DiscardHandler())
	os.Exit(m.Run())
}
