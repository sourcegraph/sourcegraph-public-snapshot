package reposource

import (
	"fmt"
	"testing"

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
			dep, err := ParseGoDependency(test.name)
			require.NoError(t, err)
			assert.Equal(t, api.RepoName(test.wantRepoName), dep.RepoName())
			assert.Equal(t, test.wantVersion, dep.PackageVersion())
		})
	}
}

func TestParseGoDependencyFromRepoName(t *testing.T) {
	tests := []struct {
		name string
		dep  *GoDependency
		err  string
	}{
		{
			name: "go/cloud.google.com/go",
			dep: NewGoDependency(module.Version{
				Path: "cloud.google.com/go",
			}),
		},
		{
			name: "go/cloud.google.com/go@v0.16.0",
			dep: NewGoDependency(module.Version{
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
			dep, err := ParseGoDependencyFromRepoName(test.name)

			assert.Equal(t, test.dep, dep)
			if test.err == "" {
				test.err = "<nil>"
			}
			assert.Equal(t, fmt.Sprint(err), test.err)
		})
	}
}
