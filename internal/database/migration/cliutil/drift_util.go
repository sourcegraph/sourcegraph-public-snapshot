package cliutil

import (
	"encoding/json"
	"fmt"
	"net/http"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

// prepareForSchemaComparison modifies the given schema description to minimize the _spurious_ diff
// between the actual and the given expected schema description.
//
// This function currently:
//   - removes entire objects that are not present in the expected schema to allow additional extensions
//     enums, functions, sequences, tables, and views that may be co-located (dev environments for one).
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

	newSchema := descriptions.SchemaDescription{
		Extensions: filteredExtensions,
		Enums:      filteredEnums,
		Functions:  filteredFunctions,
		Sequences:  filteredSequences,
		Tables:     filteredTables,
		Views:      filteredViews,
	}

	descriptions.Canonicalize(newSchema)
	return newSchema
}
