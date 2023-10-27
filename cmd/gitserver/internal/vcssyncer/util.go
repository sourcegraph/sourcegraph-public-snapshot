package vcssyncer

import (
	"fmt"
	"io"

	"github.com/sourcegraph/log"
)

// tryWrite tries to write the formatted string to the given writer, logging any errors
// to the logger.
func tryWrite(logger log.Logger, w io.Writer, format string, a ...any) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		logger.Error("failed to write log message", log.Error(err))
	}
}
