package runner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSpec_Index(t *testing.T) {
	tests := []struct {
		name          string
		spec          runner.Spec
		expectedIndex int
		expectedErr   error
	}{
		{
			name: "Docker",
			spec: runner.Spec{
				CommandSpec: command.Spec{
					Key: "step.docker.step.1.pre",
				},
			},
			expectedIndex: 1,
		},
		{
			name: "Kubernetes",
			spec: runner.Spec{
				CommandSpec: command.Spec{
					Key: "step.kubernetes.step.1.pre",
				},
			},
			expectedIndex: 1,
		},
		{
			name: "No index",
			spec: runner.Spec{
				CommandSpec: command.Spec{
					Key: "yarn-install",
				},
			},
			expectedIndex: -1,
			expectedErr:   errors.New("no index found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			index, err := test.spec.Index()
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
				assert.Equal(t, test.expectedIndex, index)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedIndex, index)
			}
		})
	}
}
