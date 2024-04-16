package codeintel

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestSanitizeRoot(t *testing.T) {
	testCases := map[string]string{
		"":                  "",
		".":                 "",
		"/":                 "",
		"foo/../bar":        "bar",
		"./bar/baz/../bonk": filepath.Join("bar", "bonk"),
	}

	for input, expectedOutput := range testCases {
		t.Run(fmt.Sprintf("input=%q", input), func(t *testing.T) {
			if SanitizeRoot(input) != expectedOutput {
				t.Errorf("unexpected root. want=%q have=%q", expectedOutput, SanitizeRoot(input))
			}
		})
	}
}
