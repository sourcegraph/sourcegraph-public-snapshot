package lockfiles

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseYarnLockFile(t *testing.T) {
	tests := []struct {
		lockfile  string
		wantDeps  string
		wantGraph string
	}{
		{
			lockfile: "testdata/parse/yarn.lock/yarn_graph.lock",
			wantDeps: `@types/tinycolor2 1.4.3
ansi-styles 4.3.0
chalk 4.1.2
color-convert 2.0.1
color-name 1.1.4
gradient-string 2.0.1
has-flag 4.0.0
supports-color 7.2.0
tinycolor2 1.4.2
tinygradient 1.1.5
`,
			wantGraph: `npm/gradient-string:
	npm/tinygradient:
		npm/types/tinycolor2
		npm/tinycolor2
	npm/chalk:
		npm/supports-color:
			npm/has-flag
		npm/ansi-styles:
			npm/color-convert:
				npm/color-name
`,
		},
		{
			lockfile: "testdata/parse/yarn.lock/yarn_normal.lock",
			wantDeps: `asap 2.0.6
jquery 3.4.1
promise 8.0.3
`,
			wantGraph: `npm/promise:
	npm/asap
npm/jquery
`,
		},
	}

	for _, tt := range tests {
		yarnLockFile, err := os.ReadFile(tt.lockfile)
		if err != nil {
			t.Fatal(err)
		}

		r := strings.NewReader(string(yarnLockFile))

		deps, graph, err := parseYarnLockFile(r)
		if err != nil {
			t.Fatal(err)
		}

		buf := bytes.Buffer{}
		for _, dep := range deps {
			_, err := fmt.Fprintf(&buf, "%s %s\n", dep.PackageSyntax(), dep.PackageVersion())
			if err != nil {
				t.Fatal()
			}
		}
		got := buf.String()

		if d := cmp.Diff(tt.wantDeps, got); d != "" {
			t.Fatalf("+want,-got\n%s", d)
		}

		gotGraph := graph.String()
		if d := cmp.Diff(tt.wantGraph, gotGraph); d != "" {
			t.Fatalf("+want,-got\n%s", d)
		}
	}
}
