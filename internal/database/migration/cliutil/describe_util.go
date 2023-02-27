package cliutil

import (
	"fmt"
	"io"
	"os"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// getOutput return a write target from the given descriptions. If no filename is
// supplied, then the given output writer is used. The shouldDecorate return value
// hints to the formatter that it will be rendered as markdown and should envelope
// its output if necessary.
func getOutput(out *output.Output, filename string, force, noColor bool) (_ io.WriteCloser, shouldDecorate bool, _ error) {
	if filename == "" {
		writeFunc := out.WriteMarkdown
		if noColor {
			writeFunc = func(s string, opts ...output.MarkdownStyleOpts) error {
				out.Write(s)
				return nil
			}
		}

		return &outputWriter{write: writeFunc}, !noColor, nil
	}

	if !force {
		if _, err := os.Stat(filename); err == nil {
			return nil, false, errors.Newf("file %q already exists", filename)
		} else if !os.IsNotExist(err) {
			return nil, false, err
		}
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	return f, false, err
}

// outputWriter is an io.WriteCloser over a single write method.
type outputWriter struct {
	write func(s string, opts ...output.MarkdownStyleOpts) error
}

func (w *outputWriter) Write(b []byte) (int, error) {
	if err := w.write(string(b)); err != nil {
		return 0, err
	}

	return len(b), nil
}

func (w *outputWriter) Close() error {
	return nil
}

// getFormatter returns the schema formatter with a given name, or nil if the given
// name is unrecognized. If shouldDecorate is true, then the output should be wrapped
// in a markdown envelope for rendering, if necessary.
func getFormatter(format string, shouldDecorate bool) descriptions.SchemaFormatter {
	switch format {
	case "json":
		jsonFormatter := descriptions.NewJSONFormatter()
		if shouldDecorate {
			jsonFormatter = markdownCodeFormatter{
				languageID: "json",
				formatter:  jsonFormatter,
			}
		}

		return jsonFormatter
	case "psql":
		return descriptions.NewPSQLFormatter()
	default:
	}

	return nil
}

type markdownCodeFormatter struct {
	languageID string
	formatter  descriptions.SchemaFormatter
}

func (f markdownCodeFormatter) Format(schemaDescription descriptions.SchemaDescription) string {
	return fmt.Sprintf(
		"%s%s\n%s\n%s",
		"```",
		f.languageID,
		f.formatter.Format(schemaDescription),
		"```",
	)
}
