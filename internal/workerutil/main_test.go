package workerutil

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"
)

func TestMain(m *testing.M) {
	// This package is INCREDIBLY noisy. We disable all logs during tests, regardless
	// of the `-v` flag, to save the noise in local development as well as CI.
	//
	// If logs are needed to debug unit test behavior, then temporarily
	// comment out the following lines.
	logger = log15.New()
	logger.SetHandler(log15.DiscardHandler())

	flag.Parse()
	os.Exit(m.Run())
}
