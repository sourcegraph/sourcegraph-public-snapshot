package runner

import (
	"os"
	"strings"

	"github.com/inconshreveable/log15"
)

// logger is the log15 root logger declared at the package level so it can be
// replaced with a no-op logger in unrelated tests that need to run migrations.
var logger = log15.Root()

func EnableLogging() {
	logger = log15.Root()
}

func DisableLogging() {
	logger = log15.New()
	logger.SetHandler(log15.DiscardHandler())
}

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
