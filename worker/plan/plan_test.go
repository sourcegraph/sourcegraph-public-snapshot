package plan

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/httptestutil"
)

func newTest() (context.Context, *httptestutil.MockClients) {
	var mocks httptestutil.MockClients
	mocks.Ctx = context.Background()

	// TODO(sqs): This makes the tests non-parallelizable.
	sourcegraph.MockNewClientFromContext(func(ctx context.Context) *sourcegraph.Client { return mocks.Client() })

	return mocks.Ctx, &mocks
}

func TestReadExistingConfig_notExist(t *testing.T) {
	ctx, mock := newTest()

	calledRepoTreeGet := mock.RepoTree.MockGet_NotFound(t)

	_, _, found, err := readExistingConfig(ctx, sourcegraph.RepoRevSpec{})
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Error("found")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestReadExistingConfig_oneAxis(t *testing.T) {
	ctx, mock := newTest()

	wantConfig := droneyaml.Config{
		Build: droneyaml.Builds{{
			Key: "",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "foo"},
				Commands:  []string{"echo foo"},
			},
		}},
	}
	wantYAML, err := yaml.Marshal(wantConfig)
	if err != nil {
		t.Fatal(err)
	}

	calledRepoTreeGet := mock.RepoTree.MockGet_Return_FileContents(t, ".drone.yml", string(wantYAML))

	config, axes, found, err := readExistingConfig(ctx, sourcegraph.RepoRevSpec{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*config, wantConfig) {
		t.Errorf("got config %q, want %q", config2yaml(*config), config2yaml(wantConfig))
	}
	if want := []matrix.Axis{{}}; !reflect.DeepEqual(axes, want) {
		t.Errorf("got axes %+v, want %+v", axes, want)
	}
	if !found {
		t.Error("!found")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestReadExistingConfig_multiAxes(t *testing.T) {
	ctx, mock := newTest()

	calledRepoTreeGet := mock.RepoTree.MockGet_Return_FileContents(t, ".drone.yml", `
build:
  image: foo:$$FOO_VERSION
  commands:
    - echo $$FOO_VERSION

matrix:
  FOO_VERSION:
    - 123
    - 456
`)

	// The Config doesn't have the matrix because that is dealt with
	// at an earlier stage.
	wantConfig := droneyaml.Config{
		Build: droneyaml.Builds{{
			Key: "",
			Build: droneyaml.Build{
				Container: droneyaml.Container{Image: "foo:$$FOO_VERSION"},
				Commands:  []string{"echo $$FOO_VERSION"},
			},
		}},
	}

	config, axes, found, err := readExistingConfig(ctx, sourcegraph.RepoRevSpec{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*config, wantConfig) {
		t.Errorf("got config %q, want %q", config2yaml(*config), config2yaml(wantConfig))
	}
	if want := []matrix.Axis{{"FOO_VERSION": "123"}, {"FOO_VERSION": "456"}}; !reflect.DeepEqual(axes, want) {
		t.Errorf("got axes %+v, want %+v", axes, want)
	}
	if !found {
		t.Error("!found")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func config2yaml(c droneyaml.Config) string {
	b, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(b)
}
