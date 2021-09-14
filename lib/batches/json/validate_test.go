package json

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalValidate(t *testing.T) {
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

	t.Run("bad schema", func(t *testing.T) {
		var target targetType
		if err := UnmarshalValidate("{", []byte(""), &target); err == nil {
			t.Error("unexpected nil error")
		} else if !strings.Contains(err.Error(), "failed to compile JSON schema") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		var target targetType
		if err := UnmarshalValidate(schema, []byte("b: bar"), &target); err == nil {
			t.Error("unexpected nil error")
		} else if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		input := `{"a": "hello", "b": 42}`

		var target targetType
		if err := UnmarshalValidate(schema, []byte(input), &target); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if diff := cmp.Diff(target, targetType{"hello", 42}); diff != "" {
			t.Errorf("unexpected target value:\n%s", diff)
		}
	})
}
