package schema

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xeipuuv/gojsonschema"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var testSchemaWithUUIDValidation string = `
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "test.schema.json#",
  "allowComments": true,
  "type": "object",
  "properties": {
    "uuid-escaped": {
      "description": "UUID -- with escaping of -",
      "type": "string",
      "pattern": "^\\{[0-9a-fA-F]{8}\\-[0-9a-fA-F]{4}\\-[0-9a-fA-F]{4}\\-[0-9a-fA-F]{4}\\-[0-9a-fA-F]{12}\\}$"
    },
    "uuid-unescaped": {
      "description": "UUID -- without escaping of - ",
      "type": "string",
      "pattern": "^\\{[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\\}$"
    }
  }
}
`

func TestSchemaValidationUUID(t *testing.T) {
	// This test validates that both regexes in the pattern behave the same way, with `\\-` or `-`.
	//
	// It's part of https://github.com/sourcegraph/sourcegraph/pull/54494, which fixes a regression for customers.
	//
	// This test should serve as an anti-regression-regression test, to make sure that we don't break something else.

	t.Run("valid input", func(t *testing.T) {
		input := `
{
	"uuid-escaped": "{fceb73c7-cef6-4abe-956d-e471281126bd}",
	"uuid-unescaped": "{fceb73c7-cef6-4abe-956d-e471281126bd}",
}
`
		if err := validateAgainstSchema(t, input, testSchemaWithUUIDValidation); err != nil {
			t.Fatalf("err should be nil, but is not: %s", err)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		input := `
{
	"uuid-escaped": "{fceb73c7+cef6-4abe-956d-e471281126bd}",
	"uuid-unescaped": "{fceb73c7+cef6-4abe-956d-e471281126bd}",
}
`
		err := validateAgainstSchema(t, input, testSchemaWithUUIDValidation)
		if err == nil {
			t.Fatal("expected err to not be nil, but is nil")
		}

		wantErr := `2 errors occurred:
	* uuid-escaped: Does not match pattern '^\{[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}\}$'
	* uuid-unescaped: Does not match pattern '^\{[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\}$'`

		// order of errors is not deterministic due to internal use of go
		// maps.
		sortedLines := func(s string) string {
			lines := strings.Split(s, "\n")
			sort.Strings(lines)
			return strings.Join(lines, "\n")
		}

		if diff := cmp.Diff(sortedLines(err.Error()), sortedLines(wantErr)); diff != "" {
			t.Fatalf("wrong error message: %s", diff)
		}
	})

}

// validateAgainstSchema does roughly what we do in
// `database.MakeValidateExternalServiceConfigFunc`, using same libraries.
//
// This is for testing our assumptions about schemas and how they work.
func validateAgainstSchema(t *testing.T, input, schema string) error {
	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(testSchemaWithUUIDValidation))
	if err != nil {
		t.Fatal(err)
	}

	normalized, err := jsonc.Parse(input)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
	if err != nil {
		t.Fatal(err)
	}

	var errs error
	for _, err := range res.Errors() {
		errString := err.String()
		// Remove `(root): ` from error formatting since these errors are
		// presented to users.
		errString = strings.TrimPrefix(errString, "(root): ")
		errs = errors.Append(errs, errors.New(errString))
	}

	return errs
}
