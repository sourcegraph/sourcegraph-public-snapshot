package cliutil

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/urfave/cli/v2"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Drift(commandName string, factory RunnerFactory, outFactory func() *output.Output) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "db",
			Usage:    "The target `schema` to compare.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "version",
			Usage:    "The target schema version. Must be resolvable as a git revlike on the sourcegraph repository.",
			Required: true,
		},
	}

	action := func(cmd *cli.Context) error {
		out := outFactory()

		if cmd.NArg() != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		var (
			schemaName = cmd.String("db")
			version    = cmd.String("version")
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

		schemas, err := store.Describe(ctx)
		if err != nil {
			return err
		}
		schema := schemas["public"]
		descriptions.Canonicalize(schema)

		expected, err := fetchSchema(schemaName, version)
		if err != nil {
			return err
		}

		schema = prepareForSchemaComparison(schema, expected)

		if diff := cmp.Diff(schema, expected); diff != "" {
			out.Writef("Database schema drift detected: %s", diff)
		} else {
			out.Write("No drift detected!")
		}

		return nil
	}

	return &cli.Command{
		Name:        "drift",
		Usage:       "Detect differences between the current database schema and the expected schema",
		Description: ConstructLongHelp(),
		Flags:       flags,
		Action:      action,
	}
}

func fetchSchema(schemaName, version string) (schemaDescription descriptions.SchemaDescription, _ error) {
	name := "schema.json"
	if schemaName != "frontend" {
		name = fmt.Sprintf("schema.%s.json", schemaName)
	}
	url := fmt.Sprintf("https://raw.githubusercontent.com/sourcegraph/sourcegraph/%s/internal/database/%s", version, name)

	resp, err := http.Get(url)
	if err != nil {
		return schemaDescription, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&schemaDescription); err != nil {
		return schemaDescription, err
	}

	return schemaDescription, nil
}

func prepareForSchemaComparison(schemaDescription, expectedSchemaDescription descriptions.SchemaDescription) descriptions.SchemaDescription {
	var (
		expectedExtensions = map[string]struct{}{}
		expectedEnums      = map[string]struct{}{}
		expectedFunctions  = map[string]struct{}{}
		expectedSequences  = map[string]struct{}{}
		expectedTables     = map[string]struct{}{}
		expectedViews      = map[string]struct{}{}
	)

	for _, extension := range expectedSchemaDescription.Extensions {
		expectedExtensions[extension] = struct{}{}
	}
	for _, enum := range expectedSchemaDescription.Enums {
		expectedEnums[enum.Name] = struct{}{}
	}
	for _, function := range expectedSchemaDescription.Functions {
		expectedFunctions[function.Name] = struct{}{}
	}
	for _, sequence := range expectedSchemaDescription.Sequences {
		expectedSequences[sequence.Name] = struct{}{}
	}
	for _, table := range expectedSchemaDescription.Tables {
		expectedTables[table.Name] = struct{}{}
	}
	for _, view := range expectedSchemaDescription.Views {
		expectedViews[view.Name] = struct{}{}
	}

	var (
		filteredExtensions = schemaDescription.Extensions[:0]
		filteredEnums      = schemaDescription.Enums[:0]
		filteredFunctions  = schemaDescription.Functions[:0]
		filteredSequences  = schemaDescription.Sequences[:0]
		filteredTables     = schemaDescription.Tables[:0]
		filteredViews      = schemaDescription.Views[:0]
	)

	for _, extension := range schemaDescription.Extensions {
		if _, ok := expectedExtensions[extension]; ok {
			filteredExtensions = append(filteredExtensions, extension)
		}
	}

	for _, enum := range schemaDescription.Enums {
		if _, ok := expectedEnums[enum.Name]; ok {
			filteredEnums = append(filteredEnums, enum)
		}
	}

	for _, function := range schemaDescription.Functions {
		if _, ok := expectedFunctions[function.Name]; ok {
			filteredFunctions = append(filteredFunctions, function)
		}
	}

	for _, sequence := range schemaDescription.Sequences {
		if _, ok := expectedSequences[sequence.Name]; ok {
			filteredSequences = append(filteredSequences, sequence)
		}
	}

	for _, table := range schemaDescription.Tables {
		if _, ok := expectedTables[table.Name]; ok {
			filteredTables = append(filteredTables, table)
		}
	}

	for _, view := range schemaDescription.Views {
		if _, ok := expectedViews[view.Name]; ok {
			filteredViews = append(filteredViews, view)
		}
	}

	return descriptions.SchemaDescription{
		Extensions: filteredExtensions,
		Enums:      filteredEnums,
		Functions:  filteredFunctions,
		Sequences:  filteredSequences,
		Tables:     filteredTables,
		Views:      filteredViews,
	}
}
