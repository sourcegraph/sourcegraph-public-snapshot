package reposource

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/module"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestParseGoDependency(t *testing.T) {
	tests := []struct {
		name         string
		wantRepoName string
		wantVersion  string
	}{
		{
			name:         "cloud.google.com/go",
			wantRepoName: "go/cloud.google.com/go",
			wantVersion:  "",
		},
		{
			name:         "cloud.google.com/go@v0.16.0",
			wantRepoName: "go/cloud.google.com/go",
			wantVersion:  "v0.16.0",
		},
		{
			name:         "cloud.google.com/go@v0.0.0-20180822173158-c12b1e7919c1",
			wantRepoName: "go/cloud.google.com/go",
			wantVersion:  "v0.0.0-20180822173158-c12b1e7919c1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dep, err := ParseGoVersionedPackage(test.name)
			require.NoError(t, err)
			assert.Equal(t, api.RepoName(test.wantRepoName), dep.RepoName())
			assert.Equal(t, test.wantVersion, dep.PackageVersion())
		})
	}
}

func TestParseGoDependencyFromRepoName(t *testing.T) {
	tests := []struct {
		name string
		dep  *GoVersionedPackage
		err  string
	}{
		{
			name: "go/cloud.google.com/go",
			dep: NewGoVersionedPackage(module.Version{
				Path: "cloud.google.com/go",
			}),
		},
		{
			name: "go/cloud.google.com/go@v0.16.0",
			dep: NewGoVersionedPackage(module.Version{
				Path:    "cloud.google.com/go",
				Version: "v0.16.0",
			}),
		},
		{
			name: "github.com/tsenart/vegeta",
			err:  "invalid go dependency repo name, missing go/ prefix",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dep, err := ParseGoDependencyFromRepoName(api.RepoName(test.name))

			assert.Equal(t, test.dep, dep)
			if test.err == "" {
				test.err = "<nil>"
			}
			assert.Equal(t, fmt.Sprint(err), test.err)
		})
	}
}

func TestGoDependency_Less(t *testing.T) {
	deps := []*GoVersionedPackage{
		parseGoDependencyOrPanic(t, "github.com/gorilla/mux@v1.1"),
		parseGoDependencyOrPanic(t, "github.com/go-kit/kit@v0.1.0"),
		parseGoDependencyOrPanic(t, "github.com/gorilla/mux@v1.8.0"),
		parseGoDependencyOrPanic(t, "github.com/go-kit/kit@v0.12.0"),
		parseGoDependencyOrPanic(t, "github.com/gorilla/mux@v1.6.1"),
		parseGoDependencyOrPanic(t, "github.com/gorilla/mux@v1.8.0-beta"),
	}

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Less(deps[j])
	})

	want := []string{
		"github.com/gorilla/mux@v1.8.0",
		"github.com/gorilla/mux@v1.8.0-beta",
		"github.com/gorilla/mux@v1.6.1",
		"github.com/gorilla/mux@v1.1",
		"github.com/go-kit/kit@v0.12.0",
		"github.com/go-kit/kit@v0.1.0",
	}

	have := make([]string, 0, len(deps))
	for _, d := range deps {
		have = append(have, d.VersionedPackageSyntax())
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("mismatch (-want, +have): %s", diff)
	}
}

func parseGoDependencyOrPanic(t *testing.T, value string) *GoVersionedPackage {
	dependency, err := ParseGoVersionedPackage(value)
	if err != nil {
		t.Fatalf("error=%s", err)
	}
	return dependency
}
