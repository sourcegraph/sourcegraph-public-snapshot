package lockfiles

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseGoModFile(t *testing.T) {
	modFile := `module github.com/a/b

go 1.18

require (
	other.org/other/no-replace-statement v1.2.3
	other.org/other/change-minor-version v1.2.3
	other.org/other/dont-replace-this v1.2.3
	other.org/other/replace-all-versions v1.2.3
	other.org/other/drop v1.2.3
	other.org/other/drop-2 v1.2.3
	other.org/other/excluded v1.2.3
)

replace (
	other.org/other/change-minor-version v1.2.3 => other.org/other/change-minor-version v1.3.0
	other.org/other/dont-replace-this v1.3.0 => other.org/other/dont-replace-this v1.4.0
	other.org/other/replace-all-versions => other.org/other/replace-all-versions v1.3.0
	other.org/other/drop v1.2.3 => /local/drop
	other.org/other/drop-2 => /local/drop
	other.org/other/doesnt-match-anything => other.org/other/doesnt-match-anything v1.0.0
)

exclude other.org/other/excluded v1.2.3
`

	r := strings.NewReader(modFile)
	deps, err := parseGoModFile(r)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte(`other.org/other/no-replace-statement v1.2.3
other.org/other/change-minor-version v1.3.0
other.org/other/dont-replace-this v1.2.3
other.org/other/replace-all-versions v1.3.0
`)

	buf := bytes.Buffer{}
	for _, dep := range deps {
		_, err := fmt.Fprintf(&buf, "%s %s\n", dep.PackageSyntax(), dep.PackageVersion())
		if err != nil {
			t.Fatal()
		}
	}
	got := buf.Bytes()

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("+want,-got\n%s", d)
	}
}
