package runner

import (
	"os"
	"strings"
)

func EnableLogging() {}

func DisableLogging() {}

// This package is INCREDIBLY noisy and imported by basically any package
// that touches the database. We disable all logs during tests to save the noise
// in local development as well as CI.
//
// If logs are needed to debug unit test behavior, then temporarily
// comment out the following lines.
func init() {
	if strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/_test/") {
		DisableLogging()
	}
}
