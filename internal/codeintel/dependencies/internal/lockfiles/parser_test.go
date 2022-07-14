package lockfiles

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func TestParse(t *testing.T) {
	files, err := filepath.Glob("testdata/parse/**/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, filePath := range files {
		if path.Ext(filePath) == ".golden" {
			continue
		}

		name := path.Base(filePath)
		lockFile := path.Dir(filePath)

		t.Run(name, func(t *testing.T) {
			f, err := os.Open(filePath)
			if err != nil {
				t.Fatal(err)
			}

			deps, graph, err := parse(lockFile, f)
			if err != nil {
				t.Fatal(err)
			}

			gotDepsList := make([]string, 0, len(deps))
			for _, dep := range deps {
				gotDepsList = append(gotDepsList, dep.VersionedPackageSyntax())
			}

			sort.Strings(gotDepsList)

			got := struct {
				Dependencies []string
				Graph        map[string]interface{}
			}{
				Dependencies: gotDepsList,
				Graph:        nil,
			}

			if graph != nil {
				got.Graph = graph.AsMap()
			}

			g := goldie.New(t, goldie.WithFixtureDir(lockFile))
			g.AssertJson(t, name, got)
		})
	}
}
