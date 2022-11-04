package cliutil

import (
	"context"
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Drift(commandName string, factory RunnerFactory, outFactory OutputFactory, expectedSchemaFactories ...ExpectedSchemaFactory) *cli.Command {
	schemaNameFlag := &cli.StringFlag{
		Name:     "db",
		Usage:    "The target `schema` to compare.",
		Required: true,
	}
	versionFlag := &cli.StringFlag{
		Name:     "version",
		Usage:    "The target schema version. Must be resolvable as a git revlike on the Sourcegraph repository.",
		Required: false,
	}
	fileFlag := &cli.StringFlag{
		Name:     "file",
		Usage:    "The target schema description file.",
		Required: false,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		schemaName := schemaNameFlag.Get(cmd)
		version := versionFlag.Get(cmd)
		file := fileFlag.Get(cmd)

		if (version == "" && file == "") || (version != "" && file != "") {
			return errors.New("must supply exactly one of -version or -file")
		}

		_, store, err := setupStore(ctx, factory, schemaName)
		if err != nil {
			return err
		}
		schemas, err := store.Describe(ctx)
		if err != nil {
			return err
		}
		schema := schemas["public"]

		if file != "" {
			return compareByFile(schemaName, version, schema, out, file)
		}

		return compareByFactories(schemaName, version, schema, out, expectedSchemaFactories)
	})

	return &cli.Command{
		Name:        "drift",
		Usage:       "Detect differences between the current database schema and the expected schema",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNameFlag,
			versionFlag,
			fileFlag,
		},
	}
}

func compareByFile(
	schemaName string,
	version string,
	schema descriptions.SchemaDescription,
	out *output.Output,
	file string,
) error {
	expectedSchema, err := readDescriptionFromFile(file)
	if err != nil {
		return err
	}

	return compareSchemaDescriptions(out, schemaName, version, canonicalize(schema), canonicalize(expectedSchema))
}

func readDescriptionFromFile(file string) (descriptions.SchemaDescription, error) {
	f, err := os.Open(file)
	if err != nil {
		return descriptions.SchemaDescription{}, err
	}
	defer f.Close()

	var expectedSchema descriptions.SchemaDescription
	err = json.NewDecoder(f).Decode(&expectedSchema)
	return expectedSchema, err
}

func compareByFactories(
	schemaName string,
	version string,
	schema descriptions.SchemaDescription,
	out *output.Output,
	expectedSchemaFactories []ExpectedSchemaFactory,
) error {
	filename, err := getSchemaJSONFilename(schemaName)
	if err != nil {
		return err
	}

	for _, factory := range expectedSchemaFactories {
		expectedSchema, ok, err := factory(filename, version)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		return compareSchemaDescriptions(out, schemaName, version, canonicalize(schema), canonicalize(expectedSchema))
	}

	return errors.Newf("failed to determine squash schema for version %s (expected the file %s to exist)", version, filename)
}

func canonicalize(schemaDescription descriptions.SchemaDescription) descriptions.SchemaDescription {
	descriptions.Canonicalize(schemaDescription)

	filtered := schemaDescription.Tables[:0]
	for i, table := range schemaDescription.Tables {
		if table.Name == "migration_logs" {
			continue
		}

		for j := range table.Columns {
			schemaDescription.Tables[i].Columns[j].Index = -1
		}

		filtered = append(filtered, schemaDescription.Tables[i])
	}
	schemaDescription.Tables = filtered

	return schemaDescription
}
