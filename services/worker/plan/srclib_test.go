package plan

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
)

func TestConfigureSrclib(t *testing.T) {
	if err := configureSrclib(&inventory.Inventory{}, &droneyaml.Config{}, []matrix.Axis{{}}, nil); err != nil {
		t.Fatal(err)
	}
}

type testConfigureSrclib_withLangs_Case struct {
	inv            *inventory.Inventory
	wantYMLStrings []string
	wantBuildItems []wantBuildItem
}

type wantBuildItem struct {
	key          string
	allowFailure bool
}

func TestConfigureSrclib_withLangs(t *testing.T) {
	tests := []testConfigureSrclib_withLangs_Case{{
		inv: &inventory.Inventory{
			Languages: []*inventory.Lang{{Name: "Go", TotalBytes: 5}, {Name: "JavaScript", TotalBytes: 5}},
		},
		wantBuildItems: []wantBuildItem{
			{key: "Go (indexing)", allowFailure: false},
			{key: "JavaScript (indexing)", allowFailure: true},
		},
	}, {
		inv: &inventory.Inventory{
			Languages: []*inventory.Lang{{Name: "Go", TotalBytes: 5}},
		},
		wantBuildItems: []wantBuildItem{{key: "Go (indexing)", allowFailure: false}},
	}, {
		inv: &inventory.Inventory{
			Languages: []*inventory.Lang{{Name: "Go", TotalBytes: 8}, {Name: "Python", TotalBytes: 2}},
		},
		wantBuildItems: []wantBuildItem{{key: "Go (indexing)", allowFailure: false}, {key: "Python (indexing)", allowFailure: true}},
	}, {
		inv: &inventory.Inventory{
			Languages: []*inventory.Lang{{Name: "HTML", TotalBytes: 9}, {Name: "Python", TotalBytes: 1}},
		},
		wantBuildItems: []wantBuildItem{{key: "HTML (indexing)", allowFailure: true}, {key: "Python (indexing)", allowFailure: true}},
	}}
	for _, test := range tests {
		testConfigureSrclib_withLangs(t, test)
	}
}

func testConfigureSrclib_withLangs(t *testing.T, test testConfigureSrclib_withLangs_Case) {
	var config droneyaml.Config
	err := configureSrclib(
		test.inv,
		&config,
		[]matrix.Axis{{}},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range test.wantBuildItems {
		got := false
		for _, item := range config.Build {
			if item.Key == want.key && item.AllowFailure == want.allowFailure {
				got = true
				break
			}
		}
		if !got {
			t.Errorf("didn't find wanted build item %+v among build items %+v", want, config.Build)
		}
	}
}

func TestInsertSrclibBuild(t *testing.T) {
	build := droneyaml.BuildItem{
		Key: "mybuild",
		Build: droneyaml.Build{
			Container: droneyaml.Container{Image: "myimage"},
		},
	}

	tests := map[string]struct {
		yaml string

		wantYAML string
		wantAxes []matrix.Axis
	}{
		"no build": {
			yaml: ``,
			wantYAML: `
build:
  mybuild:
    image: myimage
`,
		},

		"1 existing build": {
			yaml: `
build:
  image: myimage
  commands:
    - echo hello`,
			wantYAML: `
build:
  build:
    image: myimage
    commands:
      - echo hello
  mybuild:
    image: myimage
`,
		},

		"2 existing builds": {
			yaml: `
build:
  build1:
    image: myimage
    commands:
      - echo hello
  build2:
    image: myimage
`,
			wantYAML: `
build:
  build1:
    image: myimage
    commands:
      - echo hello
  build2:
    image: myimage
  mybuild:
    image: myimage
`,
		},

		// We only want the inserted deploy plugin to run for one
		// matrix axis.
		"matrix (1-dimensional)": {
			yaml: `
matrix:
  A:
    - a
    - b
`,
			wantYAML: `
build:
  mybuild:
    image: myimage
    when:
      matrix:
        A: a

matrix:
  A:
    - a
    - b
`,
		},
		"matrix (3-dimensional)": {
			yaml: `
matrix:
  A:
    - a
    - b
  B:
    - b
  C:
    - c
    - d
    - e
`,
			wantYAML: `
build:
  mybuild:
    image: myimage
    when:
      matrix:
        A: a
        B: b
        C: c

matrix:
  A:
    - a
    - b
  B:
    - b
  C:
    - c
    - d
    - e
`,
		},
	}
	for label, test := range tests {
		axes, err := parseMatrix(test.yaml)
		if err != nil {
			t.Errorf("%s: matrix.Parse(yaml): %s", label, err)
			continue
		}

		cfg, err := droneyaml.ParseString(test.yaml)
		if err != nil {
			t.Errorf("%s: Parse(yaml): %s", label, err)
			continue
		}
		wantCfg, err := droneyaml.ParseString(test.wantYAML)
		if err != nil {
			t.Fatalf("%s: Parse(wantYAML): %s", label, err)
			continue
		}

		if err := insertSrclibBuild(cfg, axes, build); err != nil {
			t.Errorf("%s: %s", label, err)
			continue
		}
		if !reflect.DeepEqual(cfg, wantCfg) {
			t.Errorf("%s: YAML mismatch\n\n### GOT\n%+v\n\n### WANT\n%+v", label, cfg, wantCfg)
		}
	}
}

func config2yaml(c droneyaml.Config) string {
	b, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(b)
}
