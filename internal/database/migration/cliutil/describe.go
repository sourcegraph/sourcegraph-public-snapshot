package cliutil

import (
	"flag"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Describe(commandName string, factory RunnerFactory, outFactory func() *output.Output) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "db",
			Usage:    "The target `schema` to describe.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "format",
			Usage:    "The target output format.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "out",
			Usage:    "The file to write to. If not supplied, stdout is used.",
			Required: false,
		},
		&cli.BoolFlag{
			Name:     "force",
			Usage:    "Force write the file if it already exists.",
			Required: false,
		},
		&cli.BoolFlag{
			Name:     "no-color",
			Usage:    "If writing to stdout, disable output colorization.",
			Required: false,
		},
	}

	action := func(cmd *cli.Context) error {
		out := outFactory()

		if cmd.NArg() != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		var (
			schemaName     = cmd.String("db")
			format         = cmd.String("format")
			outputFilename = cmd.String("out")
			force          = cmd.Bool("force")
			noColor        = cmd.Bool("no-color")
		)

		ctx := cmd.Context
		r, err := factory(ctx, []string{schemaName})
		if err != nil {
			return err
		}
		store, err := r.Store(ctx, schemaName)
		if err != nil {
			return err
		}

		formatter := getFormatter(format)
		if formatter == nil {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: unrecognized format %q (must be json or psql)", format))
			return flag.ErrHelp
		}

		output, err := getOutput(out, outputFilename, force, noColor)
		if err != nil {
			return err
		}
		defer output.Close()

		schemas, err := store.Describe(ctx)
		if err != nil {
			return err
		}

		if _, err := io.Copy(output, strings.NewReader(formatter.Format(schemas["public"]))); err != nil {
			return err
		}

		return nil
	}

	return &cli.Command{
		Name:        "describe",
		Usage:       "Describe the current database schema",
		Description: ConstructLongHelp(),
		Flags:       flags,
		Action:      action,
	}
}

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

func getOutput(out *output.Output, filename string, force, noColor bool) (io.WriteCloser, error) {
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
