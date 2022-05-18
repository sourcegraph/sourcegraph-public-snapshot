package cliutil

import (
	"flag"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
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
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: unrecognized format %q", format))
			return flag.ErrHelp
		}

		output, err := getOutput(outputFilename)
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

func getOutput(filename string) (io.WriteCloser, error) {
	if filename == "" {
		return nopCloser{os.Stdout}, nil
	}

	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
