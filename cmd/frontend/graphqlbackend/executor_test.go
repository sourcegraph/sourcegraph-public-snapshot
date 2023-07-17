package graphqlbackend

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

func TestCalculateExecutorCompatibility(t *testing.T) {
	tests := []struct {
		name                  string
		executorVersion       string
		sourcegraphVersion    string
		isActive              bool
		expectedCompatibility *string
		expectedError         error
	}{
		{
			name:                  "Dev mode",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompatibility: nil,
		},
		{
			name:                  "Executor is inactive",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              false,
			expectedCompatibility: nil,
		},
		{
			name:                  "Executor is one minor version behind",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "3.42.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate.ToGraphQL(),
		},
		{
			name:                  "Executor is one minor version ahead",
			executorVersion:       "3.42.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate.ToGraphQL(),
		},
		{
			name:                  "Executor is the same version as the Sourcegraph instance",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate.ToGraphQL(),
		},
		{
			name:                  "Executor is the same version as the Sourcegraph instance (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-08-25_4.4-a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate.ToGraphQL(),
		},
		{
			name:                  "Executor is the same version as the Sourcegraph instance (insiders - old version)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-08-25_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate.ToGraphQL(),
		},
		{
			name:                  "Executor is multiple minor version behind",
			executorVersion:       "3.40.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityOutdated.ToGraphQL(),
		},
		{
			name:                  "Executor is major version behind",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "4.0.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityOutdated.ToGraphQL(),
		},
		{
			name:                  "Executor is multiple minor version ahead",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "3.40.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityVersionAhead.ToGraphQL(),
		},
		{
			executorVersion:       "4.0.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityVersionAhead.ToGraphQL(),
		},
		{
			name:                  "Executor is one release cycle behind (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-06-10_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-07-25_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityOutdated.ToGraphQL(),
		},
		{
			name:                  "Executor is one release cycle ahead (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-10-30_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-09-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityVersionAhead.ToGraphQL(),
		},
		{
			name:                  "Execcutor build date is greater than one release cycle + sourcegraph build date (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-08-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate.ToGraphQL(),
		},
		{
			name:                  "Sourcegrpah version mismatch",
			executorVersion:       "3.36.2",
			sourcegraphVersion:    "169135_2022-08-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: nil,
		},
		{
			name:                  "Executor version mismatch",
			executorVersion:       "169135_2022-08-15_a2b623dce148",
			sourcegraphVersion:    "3.39.2",
			isActive:              true,
			expectedCompatibility: nil,
		},
		{
			name:                  "Executor is in dev mode",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "3.39.2",
			isActive:              true,
			expectedCompatibility: nil,
		},
		{
			name:                  "Sourcegraph instance is in dev mode",
			executorVersion:       "3.39.2",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompatibility: nil,
		},
		{
			name:                  "Executor is in dev mode and Sourcegraph instance is on insiders version",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "169135_2022-08-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: nil,
		},
		{
			name:                  "Sourcegraph instance is in dev mode and executor is on insiders version",
			executorVersion:       "169135_2022-08-15_a2b623dce148",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompatibility: nil,
		},
		{
			name:                  "Executor version is an invalid semver",
			executorVersion:       "\n1.2",
			sourcegraphVersion:    "3.39.2",
			isActive:              true,
			expectedCompatibility: nil,
			expectedError:         errors.New("failed to parse executor version \"\\n1.2\": Invalid Semantic Version"),
		},
		{
			name:                  "Sourcegraph version is an invalid semver",
			executorVersion:       "4.0.1",
			sourcegraphVersion:    "\n1.2",
			isActive:              true,
			expectedCompatibility: nil,
			expectedError:         errors.New("failed to parse Sourcegraph version \"\\n1.2\": Invalid Semantic Version"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version.Mock(test.sourcegraphVersion)
			actual, err := calculateExecutorCompatibility(test.executorVersion)

			if test.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedError.Error(), err.Error())
				assert.Nil(t, actual)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedCompatibility, actual)
			}
		})
	}
}
