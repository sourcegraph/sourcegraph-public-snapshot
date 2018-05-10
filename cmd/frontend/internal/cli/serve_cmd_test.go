package cli

import "testing"

func TestParseStringOrBool(t *testing.T) {
	defaultValue := "default"
	// parsedValue -> stringOrBool
	cases := map[string]interface{}{
		defaultValue: nil,
		"":           "",
		"hi":         "hi",
		"on":         true,
		"off":        false,
	}
	for want, v := range cases {
		got := parseStringOrBool(v, defaultValue)
		if got != want {
			t.Errorf("parseStringOrBool(%q) got %q want %q", v, got, want)
		}
	}
}
