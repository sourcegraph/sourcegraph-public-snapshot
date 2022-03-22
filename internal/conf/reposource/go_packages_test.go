package reposource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestParseGoModDependency(t *testing.T) {
	t.Run("not enough fields", func(t *testing.T) {
		_, err := ParseGoModDependency("cloud.google.com/go")
		assert.Error(t, err)
	})

	tests := []struct {
		name         string
		wantRepoName string
		wantVersion  string
	}{
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
			dep, err := ParseGoModDependency(test.name)
			require.NoError(t, err)
			assert.Equal(t, api.RepoName(test.wantRepoName), dep.RepoName())
			assert.Equal(t, test.wantVersion, dep.PackageVersion())
		})
	}
}
