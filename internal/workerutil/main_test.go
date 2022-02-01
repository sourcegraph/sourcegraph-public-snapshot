package workerutil

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// This package is INCREDIBLY noisy. We disable all logs during tests
	// (regardless of the -v flag) to save spewing useless logs into CI.
	//
	// If logs are needed to debug unit test behavior, then temporarily
	// comment out the following line.
	disableLogs()

	flag.Parse()
	os.Exit(m.Run())
}
