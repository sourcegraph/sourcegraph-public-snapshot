package runner

import (
	"os"
	"strings"

	"github.com/inconshreveable/log15"
)

// logger is the log15 root logger declared at the package level so it can be
// replaced with a no-op logger in unrelated tests that need to run migrations.
var logger = log15.Root()

func init() {
	if strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/_test/") {
		// This package is INCREDIBLY noisy and imported by everything that
		// touches the database. We disable all logs during tests to save
		// spewing useless logs into CI.
		//
		// If logs are needed to debug unit test behavior, then temporarily
		// comment out the following lines.
		logger = log15.New()
		logger.SetHandler(log15.DiscardHandler())
	}
}
