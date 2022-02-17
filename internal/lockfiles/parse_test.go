package lockfiles

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		file string
		data string
		want []*Dependency
	}{
		{
			file: "package-lock.json",
			data: `{"dependencies": {
        "nan": {"version": "2.15.0"},
        "tree-sitter-cli": {"version": "0.20.4"}
      }}`,
			want: []*Dependency{
				{Name: "nan", Version: "2.15.0", Kind: KindNPM},
				{Name: "tree-sitter-cli", Version: "0.20.4", Kind: KindNPM},
			},
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			got, err := Parse(test.file, []byte(test.data))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("dependency mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
