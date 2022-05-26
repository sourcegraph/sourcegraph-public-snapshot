package cliutil

import (
	"context"

	"github.com/urfave/cli/v2"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type ExpectedSchemaFactory func(repoName, version string) (descriptions.SchemaDescription, error)

func Drift(commandName string, factory RunnerFactory, outFactory OutputFactory, expectedSchemaFactory ExpectedSchemaFactory) *cli.Command {
	schemaNameFlag := &cli.StringFlag{
		Name:     "db",
		Usage:    "The target `schema` to compare.",
		Required: true,
	}
	versionFlag := &cli.StringFlag{
		Name:     "version",
		Usage:    "The target schema version. Must be resolvable as a git revlike on the sourcegraph repository.",
		Required: true,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		schemaName := schemaNameFlag.Get(cmd)

		_, store, err := setupStore(ctx, factory, schemaName)
		if err != nil {
			return err
		}
		schemas, err := store.Describe(ctx)
		if err != nil {
			return err
		}
		schema := schemas["public"]

		filename, err := getSchemaJSONFilename(schemaName)
		if err != nil {
			return err
		}
		expectedSchema, err := expectedSchemaFactory(filename, versionFlag.Get(cmd))
		if err != nil {
			return err
		}

		return compareSchemaDescriptions(out, schemaName, canonicalize(schema), canonicalize(expectedSchema))
	})

	return &cli.Command{
		Name:        "drift",
		Usage:       "Detect differences between the current database schema and the expected schema",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNameFlag,
			versionFlag,
		},
	}
}

func canonicalize(schemaDescription descriptions.SchemaDescription) descriptions.SchemaDescription {
	descriptions.Canonicalize(schemaDescription)

	for i, table := range schemaDescription.Tables {
		for j := range table.Columns {
			schemaDescription.Tables[i].Columns[j].Index = -1
		}
	}

	return schemaDescription
}
