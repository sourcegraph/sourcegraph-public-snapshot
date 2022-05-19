package cliutil

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// fetchSchema returns the schema description of the given schema at the given version. If the version
// is not resolvable as a git rev-like, then an error is returned.
func fetchSchema(schemaName, version string) (schemaDescription descriptions.SchemaDescription, _ error) {
	url, err := getSchemaURL(schemaName, version)
	if err != nil {
		return schemaDescription, err
	}
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

// getSchemaURL returns the GitHub raw URL for the JSON-serialized version of the given schema at
// the given version.
func getSchemaURL(schemaName, version string) (string, error) {
	filename, err := getSchemaJSONFilename(schemaName, version)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://raw.githubusercontent.com/sourcegraph/sourcegraph/%s/internal/database/%s", version, filename), nil
}

// getSchemaJSONFilename returns the basename of the JSON-serialized schema in the sg/sg repository.
func getSchemaJSONFilename(schemaName, version string) (string, error) {
	switch schemaName {
	case "frontend":
		return "schema.json", nil
	case "codeintel":
		fallthrough
	case "codeinsights":
		return fmt.Sprintf("schema.%s.json", schemaName), nil
	}

	return "", errors.Newf("unknown schema name %q", schemaName)
}

var errOutOfSync = errors.Newf("database schema is out of sync")

func compareSchemaDescriptions(out *output.Output, actual, expected schemas.SchemaDescription) (err error) {
	missing := func(typeName, name string, value any) {
		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing %s %q", typeName, name)))
		jsonValue, _ := json.MarshalIndent(value, "", "    ")
		out.WriteMarkdown(fmt.Sprintf("```json\n%s\n```", string(jsonValue)))
		err = errOutOfSync
	}

	diff := func(typeName, fieldName, name string, a, b any) {
		diff := cmp.Diff(a, b)
		if diff != "" {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Mismatched %s of %s %q", fieldName, typeName, name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s```", diff))
			err = errOutOfSync
		}
	}

	//
	// Compare extensions

	actualExtensions := map[string]struct{}{}
	for _, extension := range actual.Extensions {
		actualExtensions[extension] = struct{}{}
	}
	expectedExtensions := map[string]struct{}{}
	for _, extension := range expected.Extensions {
		expectedExtensions[extension] = struct{}{}
	}
	for name := range expectedExtensions {
		if _, ok := actualExtensions[name]; !ok {
			missing("extension", name, nil)
		}
	}

	//
	// Compare enums

	actualEnums := map[string]schemas.EnumDescription{}
	for _, enum := range actual.Enums {
		actualEnums[enum.Name] = enum
	}
	expectedEnums := map[string]schemas.EnumDescription{}
	for _, enum := range expected.Enums {
		expectedEnums[enum.Name] = enum
	}
	for name, expectedEnum := range expectedEnums {
		if enum, ok := actualEnums[name]; !ok {
			missing("enum", name, expectedEnum)
		} else {
			diff("enum", "labels", name, expectedEnum.Labels, enum.Labels)
		}
	}

	//
	// Compare functions

	actualFunctions := map[string]schemas.FunctionDescription{}
	for _, function := range actual.Functions {
		actualFunctions[function.Name] = function
	}
	expectedFunctions := map[string]schemas.FunctionDescription{}
	for _, function := range expected.Functions {
		expectedFunctions[function.Name] = function
	}
	for name, expectedFunction := range expectedFunctions {
		if function, ok := actualFunctions[name]; !ok {
			missing("function", name, expectedFunction)
		} else {
			diff("function", "definition", name, expectedFunction.Definition, function.Definition)
		}
	}

	//
	// Compare sequences

	actualSequences := map[string]schemas.SequenceDescription{}
	for _, sequence := range actual.Sequences {
		actualSequences[sequence.Name] = sequence
	}
	expectedSequences := map[string]schemas.SequenceDescription{}
	for _, sequence := range expected.Sequences {
		expectedSequences[sequence.Name] = sequence
	}
	for name, expectedSequence := range expectedSequences {
		if sequence, ok := actualSequences[name]; !ok {
			missing("sequence", name, expectedSequence)
		} else {
			diff("sequence", "definition", name, expectedSequence, sequence)
		}
	}

	//
	// Compare tables

	actualTables := map[string]schemas.TableDescription{}
	for _, table := range actual.Tables {
		actualTables[table.Name] = table
	}
	expectedTables := map[string]schemas.TableDescription{}
	for _, Table := range expected.Tables {
		expectedTables[Table.Name] = Table
	}
	for name, expectedTable := range expectedTables {
		if table, ok := actualTables[name]; !ok {
			missing("table", name, expectedTable)
		} else {
			diff("table", "columns", name, expectedTable.Columns, table.Columns)
			diff("table", "columns", name, expectedTable.Constraints, table.Constraints)
			diff("table", "columns", name, expectedTable.Indexes, table.Indexes)
			diff("table", "columns", name, expectedTable.Triggers, table.Triggers)
		}
	}

	//
	// Compare views

	actualViews := map[string]schemas.ViewDescription{}
	for _, view := range actual.Views {
		actualViews[view.Name] = view
	}
	expectedViews := map[string]schemas.ViewDescription{}
	for _, view := range expected.Views {
		expectedViews[view.Name] = view
	}
	for name, expectedView := range expectedViews {
		if view, ok := actualViews[name]; !ok {
			missing("view", name, expectedView)
		} else {
			diff("view", "definition", name, expectedView.Definition, view.Definition)
		}
	}

	if err == nil {
		out.Write("No drift detected")
	}
	return err
}
