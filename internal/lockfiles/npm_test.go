package lockfiles

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseNpm(t *testing.T) {
	tests := []struct {
		JSON string
		Want []Dependency
	}{
		{
			JSON: `{
            "dependencies": {
              "nan": {"version": "2.15.0"},
              "tree-sitter-cli": {"version": "0.20.4"}
            }}`,
			Want: []Dependency{
				{Name: "nan", Version: "2.15.0"},
				{Name: "tree-sitter-cli", Version: "0.20.4"},
			},
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			got, err := ParseNpm([]byte(test.JSON))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.Want, got); diff != "" {
				t.Fatalf("dependency mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
