package app

import (
	"encoding/json"
	"html/template"
	"testing"
)

func TestRawJSON(t *testing.T) {
	tests := map[string]template.JS{
		"a":   "a",
		`"a"`: `"a"`,

		// Ensure that rawJSON's output is safe when included in an
		// HTML document; it should not be able to close its outer
		// <script> tag, or else there is an XSS vulnerability.
		`"</script>a"`: "\"\\u003c/script\\u003ea\"",
	}
	for input, want := range tests {
		inputJSON := json.RawMessage(input)
		got := rawJSON(&inputJSON)
		if got != want {
			t.Errorf("%q: got %q, want %q", input, got, want)
		}
	}
}
