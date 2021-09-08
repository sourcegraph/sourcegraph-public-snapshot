package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func TestServiceConnections(t *testing.T) {
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// We only test that we get something non-empty back.
	sc := serviceConnections()
	if reflect.DeepEqual(sc, conftypes.ServiceConnections{}) {
		t.Fatal("expected non-empty service connections")
	}
}

func TestReadSiteConfigFile(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		Name  string
		Files []string
		Want  string
		Err   string
	}{{
		Name:  "one",
		Files: []string{`{"hello": "world"}`},
		Want:  `{"hello": "world"}`,
	}, {
		Name: "two",
		Files: []string{
			`// leading comment
{
  // first comment
  "first": "file",
} // trailing comment
`, `{"second": "file"}`},
		Want: `// merged SITE_CONFIG_FILE
{

// BEGIN $tmp/0.json
  "first": "file",
// END $tmp/0.json

// BEGIN $tmp/1.json
  "second": "file",
// END $tmp/1.json
}
`,
	}, {
		Name: "parse-error",
		Files: []string{
			"{}",
			"{",
		},
		Err: "CloseBraceExpected",
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var paths []string
			for i, b := range c.Files {
				p := filepath.Join(dir, fmt.Sprintf("%d.json", i))
				paths = append(paths, p)
				if err := os.WriteFile(p, []byte(b), 0600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := readSiteConfigFile(paths)
			if c.Err != "" && !strings.Contains(fmt.Sprintf("%s", err), c.Err) {
				t.Fatalf("%s doesn't contain error substring %s", err, c.Err)
			}
			got = bytes.ReplaceAll(got, []byte(dir), []byte("$tmp"))
			if d := cmp.Diff(c.Want, string(got)); d != "" {
				t.Fatalf("unexpected merge (-want, +got):\n%s", d)
			}
		})
	}
}
