package workerutil

import "github.com/inconshreveable/log15"

// logger is the log15 root logger declared at the package level
// so it can be replaced with a no-op logger in unrelated tests
// that need to run migrations.
var logger = log15.Root()

func disableLogs() {
	logger = log15.New()
	logger.SetHandler(log15.DiscardHandler())
}
