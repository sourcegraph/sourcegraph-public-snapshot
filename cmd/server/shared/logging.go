package shared

import (
	"fmt"
	"os"
)

func l(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "✱ "+format+"\n", args...)
}

var logLevelConverter = map[string]string{
	"dbug":  "debug",
	"info":  "info",
	"warn":  "warn",
	"error": "error",
	"crit":  "fatal",
}

// convertLogLevel converts a sourcegraph log level (dbug, info, warn, error, crit) into
// values postgres exporter accepts (debug, info, warn, error, fatal)
// If value cannot be converted returns "warn" which seems like a good middle-ground.
func convertLogLevel(level string) string {
	lvl, ok := logLevelConverter[level]
	if ok {
		return lvl
	}
	return "warn"
}
