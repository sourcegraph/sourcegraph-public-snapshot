package yaml

import (
	"testing"
)

func FuzzTestUnmarshalValidate(f *testing.F) {
	type targetType struct {
		A string
		B int
	}

	schema := `{
        "$schema": "http://json-schema.org/draft-07/schema#",
        "$id": "https://github.com/sourcegraph/sourcegraph/lib/batches/schema/test.schema.json",
        "type": "object",
        "properties": {
            "a": { "type": "string" },
            "b": { "type": "integer" }
        }
    }`

	input := `
            a: hello
            b: 42
        `
	f.Add(input)
	f.Fuzz(func(t *testing.T, input string) {

		var target targetType
		if err := UnmarshalValidate(schema, []byte(input), &target); err != nil {
			t.Logf("unexpected non-nil error: %v", err)
		}
	})
}
