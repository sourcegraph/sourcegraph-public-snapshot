package runner_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
)

func TestDockerRunner_Setup(t *testing.T) {
	tests := []struct {
		name               string
		dockerAuthConfig   types.DockerAuthConfig
		expectedDockerAuth string
		expectedErr        error
	}{
		{
			name:             "Setup",
			dockerAuthConfig: types.DockerAuthConfig{},
		},
		{
			name: "Setup",
			dockerAuthConfig: types.DockerAuthConfig{
				Auths: map[string]types.DockerAuthConfigAuth{
					"index.docker.io": {
						Auth: []byte(`{"field": "value"}`),
					},
				},
			},
			expectedDockerAuth: `{"auths":{"index.docker.io":{"auth":"eyJmaWVsZCI6ICJ2YWx1ZSJ9"}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dockerRunner := runner.NewDockerRunner(nil, nil, "", command.DockerOptions{}, test.dockerAuthConfig)

			ctx := context.Background()
			err := dockerRunner.Setup(ctx)
			defer dockerRunner.Teardown(ctx)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				fmt.Println(dockerRunner.TempDir())
				entries, err := os.ReadDir(dockerRunner.TempDir())
				require.NoError(t, err)
				if len(test.expectedDockerAuth) == 0 {
					require.Len(t, entries, 0)
				} else {
					require.Len(t, entries, 1)
					dockerAuthEntries, err := os.ReadDir(filepath.Join(dockerRunner.TempDir(), entries[0].Name()))
					require.NoError(t, err)
					require.Len(t, dockerAuthEntries, 1)
					f, err := os.ReadFile(filepath.Join(dockerRunner.TempDir(), entries[0].Name(), dockerAuthEntries[0].Name()))
					require.NoError(t, err)
					assert.Equal(t, test.expectedDockerAuth, string(f))
				}
			}
		})
	}
}
