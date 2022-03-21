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
			name:         "cloud.google.com/go v0.16.0/go.mod h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw=",
			wantRepoName: "gomod/cloud.google.com/go",
			wantVersion:  "0.16.0",
		},
		{
			name:         "cloud.google.com/go v0.44.1/go.mod",
			wantRepoName: "gomod/cloud.google.com/go",
			wantVersion:  "0.44.1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dep, err := ParseGoModDependency(test.name)
			require.NoError(t, err)
			assert.Equal(t, api.RepoName(test.wantRepoName), dep.RepoName())
			assert.Equal(t, test.wantVersion, dep.Version)
		})
	}
}
