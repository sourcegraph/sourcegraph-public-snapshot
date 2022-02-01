package workerutil

import "github.com/inconshreveable/log15"

// logger is the log15 root logger declared at the package level so it can be
// replaced with a no-op logger in unrelated tests that use this code.
var logger = log15.Root()
