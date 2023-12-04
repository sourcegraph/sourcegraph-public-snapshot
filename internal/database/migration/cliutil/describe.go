package cliutil

import (
	"context"
	"io"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Describe(commandName string, factory RunnerFactory, outFactory OutputFactory) *cli.Command {
	schemaNameFlag := &cli.StringFlag{
		Name:     "schema",
		Usage:    "The target `schema` to describe. Possible values are 'frontend', 'codeintel' and 'codeinsights'",
		Required: true,
		Aliases:  []string{"db"},
	}
	formatFlag := &cli.StringFlag{
		Name:     "format",
		Usage:    "The target output format.",
		Required: true,
	}
	outFlag := &cli.StringFlag{
		Name:     "out",
		Usage:    "The file to write to. If not supplied, stdout is used.",
		Required: false,
	}
	forceFlag := &cli.BoolFlag{
		Name:     "force",
		Usage:    "Force write the file if it already exists.",
		Required: false,
	}
	noColorFlag := &cli.BoolFlag{
		Name:     "no-color",
		Usage:    "If writing to stdout, disable output colorization.",
		Required: false,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) (err error) {
		w, shouldDecorate, err := getOutput(out, outFlag.Get(cmd), forceFlag.Get(cmd), noColorFlag.Get(cmd))
		if err != nil {
			return err
		}
		defer w.Close()

		formatter := getFormatter(formatFlag.Get(cmd), shouldDecorate)
		if formatter == nil {
			return flagHelp(out, "unrecognized format %q (must be json or psql)", formatFlag.Get(cmd))
		}

		schemaName := TranslateSchemaNames(schemaNameFlag.Get(cmd), out)
		store, err := setupStore(ctx, factory, schemaName)
		if err != nil {
			return err
		}

		pending := out.Pending(output.Linef("", output.StylePending, "Describing database %s...", schemaName))
		defer func() {
			if err == nil {
				pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Description of %s written to target", schemaName))
			} else {
				pending.Destroy()
			}
		}()

		schemas, err := store.Describe(ctx)
		if err != nil {
			return err
		}
		schema := schemas["public"]

		if _, err := io.Copy(w, strings.NewReader(formatter.Format(schema))); err != nil {
			return err
		}

		return nil
	})

	return &cli.Command{
		Name:        "describe",
		Usage:       "Describe the current database schema",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNameFlag,
			formatFlag,
			outFlag,
			forceFlag,
			noColorFlag,
		},
	}
}
