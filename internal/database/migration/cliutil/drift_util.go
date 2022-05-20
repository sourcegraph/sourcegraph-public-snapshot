package cliutil

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// getSchemaJSONFilename returns the basename of the JSON-serialized schema in the sg/sg repository.
func getSchemaJSONFilename(schemaName string) (string, error) {
	switch schemaName {
	case "frontend":
		return "internal/database/schema.json", nil
	case "codeintel":
		fallthrough
	case "codeinsights":
		return fmt.Sprintf("internal/database/schema.%s.json", schemaName), nil
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
			diff("table", "constraints", name, expectedTable.Constraints, table.Constraints)
			diff("table", "indexes", name, expectedTable.Indexes, table.Indexes)
			diff("table", "triggers", name, expectedTable.Triggers, table.Triggers)
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
