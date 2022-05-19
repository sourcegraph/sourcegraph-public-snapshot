package cliutil

import (
	"io"
	"os"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// outputWriter is an io.WriteCloser over an *output.Output object.
type outputWriter struct {
	out     *output.Output
	noColor bool
}

func (w *outputWriter) Write(b []byte) (int, error) {
	// Color currently unimplemented
	w.out.Write(string(b))
	return len(b), nil
}

func (w *outputWriter) Close() error {
	return nil
}

// getFormatter returns the schema formatter with a given name, or nil if the given
// name is unrecognized.
func getFormatter(format string) descriptions.SchemaFormatter {
	switch format {
	case "json":
		return descriptions.NewJSONFormatter()
	case "psql":
		return descriptions.NewPSQLFormatter()
	default:
	}

	return nil
}

// getOutput return a write target from the given descriptions. If no filename is
// supplied, then the given output writer is used.
func getOutput(
	out *output.Output,
	filename string,
	force bool,
	noColor bool,
) (io.WriteCloser, error) {
	if filename == "" {
		return &outputWriter{out, noColor}, nil
	}

	if !force {
		if _, err := os.Stat(filename); err == nil {
			return nil, errors.Newf("file %q already exists", filename)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
}
