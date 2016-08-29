package plan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
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

func TestImages_MemLimit(t *testing.T) {
	for lang, b := range langSrclibConfigs {
		if b.Build.Container.Memory == 0 {
			t.Errorf("no memory limit set for language %s", lang)
		}
	}
}

// TestImages_Version ensures we can parse all docker image names specified in
// srclib_images.go
func TestImages_Version(t *testing.T) {
	for lang, b := range langSrclibConfigs {
		image := b.Build.Container.Image
		_, err := versionHash(image)
		if err != nil {
			t.Errorf("could not parse %s hash %#v: %s", lang, image, err)
		}
	}
}

// TestImages_Exists ensures each image specified in srclib_images.go exists
// on the remote docker registry.
func TestImages_Exists(t *testing.T) {
	// We speak to an external service, so skip if we want short tests
	if testing.Short() {
		t.Skip("talks to external service (docker.io)")
	}

	wg := sync.WaitGroup{}
	for lang, b := range langSrclibConfigs {
		wg.Add(1)
		go func(lang, image string) {
			defer wg.Done()
			v, err := getImageSchemaVersion(image)
			if err != nil {
				t.Logf("WARNING: %s - failed to get docker image version of %s: %s", lang, image, err)
				return
			}
			if v == 0 {
				t.Errorf("%s: could not find image %s", lang, image)
			}
		}(lang, b.Build.Container.Image)
	}
	wg.Wait()
}

func getImageSchemaVersion(image string) (int, error) {
	parts := strings.SplitN(image, ":", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("missing tag in image name %s", image)
	}
	name, tag := parts[0], parts[1]

	token, err := getToken("registry.docker.io", "repository:"+name+":pull")
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://index.docker.io/v2/%s/manifests/%s", name, tag), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	m := struct {
		SchemaVersion int `json:"schemaVersion"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	return m.SchemaVersion, err
}

func getToken(service, scope string) (string, error) {
	// We need to get a token
	v := url.Values{}
	v.Set("service", service)
	v.Set("scope", scope)
	resp, err := http.Get("https://auth.docker.io/token?" + v.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	t := struct {
		Token string `json:"token"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&t)
	return t.Token, err
}
