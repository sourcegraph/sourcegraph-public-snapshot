package registry

import "testing"

func TestParseExtensionBundleFilename(t *testing.T) {
	tests := map[string]int64{
		"123.js":        123,
		"123-a-b.js":    123,
		"123-a-b-c.map": 123,
	}
	for input, want := range tests {
		got, err := parseExtensionBundleFilename(input)
		if err != nil {
			t.Errorf("%q: %s", input, err)
		}
		if got != want {
			t.Errorf("%q: got %d, want %d", input, got, want)
		}
	}
}
