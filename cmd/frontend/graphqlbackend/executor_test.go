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
		expectedCompatibility ExecutorCompatibility
		expectedError         error
	}{
		{
			name:                  "Dev mode",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompatibility: "",
		},
		{
			name:                  "Executor is inactive",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              false,
			expectedCompatibility: "",
		},
		{
			name:                  "Executor is one minor version behind",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "3.42.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor is one minor version behind",
			executorVersion:       "3.42.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor is the same version as the Sourcegraph instance",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor is the same version as the Sourcegraph instance (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-08-25_4.4-a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor is the same version as the Sourcegraph instance (insiders - old version)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-08-25_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor is multiple minor versions behind",
			executorVersion:       "3.40.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityOutdated,
		},
		{
			name:                  "Executor is major version behind",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "4.0.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityOutdated,
		},
		{
			name:                  "Executor is multiple patch versions behind",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "3.43.12",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor is multiple minor version ahead",
			executorVersion:       "3.43.0",
			sourcegraphVersion:    "3.40.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityVersionAhead,
		},
		{
			executorVersion:       "4.0.0",
			sourcegraphVersion:    "3.43.0",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityVersionAhead,
		},
		{
			name:                  "Executor is one release cycle behind (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-06-10_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-07-25_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityOutdated,
		},
		{
			name:                  "Executor is one release cycle ahead (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-10-30_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-09-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityVersionAhead,
		},
		{
			name:                  "Execcutor build date is greater than one release cycle + sourcegraph build date (insiders)",
			executorVersion:       "executor-patch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_patch",
			sourcegraphVersion:    "169135_2022-08-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Sourcegrpah version mismatch",
			executorVersion:       "3.36.2",
			sourcegraphVersion:    "169135_2022-08-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: "",
		},
		{
			name:                  "Executor version mismatch",
			executorVersion:       "169135_2022-08-15_a2b623dce148",
			sourcegraphVersion:    "3.39.2",
			isActive:              true,
			expectedCompatibility: "",
		},
		{
			name:                  "Executor is in dev mode",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "3.39.2",
			isActive:              true,
			expectedCompatibility: "",
		},
		{
			name:                  "Sourcegraph instance is in dev mode",
			executorVersion:       "3.39.2",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompatibility: "",
		},
		{
			name:                  "Executor is in dev mode and Sourcegraph instance is on insiders version",
			executorVersion:       "0.0.0+dev",
			sourcegraphVersion:    "169135_2022-08-15_a2b623dce148",
			isActive:              true,
			expectedCompatibility: "",
		},
		{
			name:                  "Sourcegraph instance is in dev mode and executor is on insiders version",
			executorVersion:       "169135_2022-08-15_a2b623dce148",
			sourcegraphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompatibility: "",
		},
		{
			name:                  "Executor version is an invalid semver",
			executorVersion:       "\n1.2",
			sourcegraphVersion:    "3.39.2",
			isActive:              true,
			expectedCompatibility: "",
			expectedError:         errors.New("failed to parse executor version \"\\n1.2\": Invalid Semantic Version"),
		},
		{
			name:                  "Sourcegraph version is an invalid semver",
			executorVersion:       "4.0.1",
			sourcegraphVersion:    "\n1.2",
			isActive:              true,
			expectedCompatibility: "",
			expectedError:         errors.New("failed to parse sourcegraph version \"\\n1.2\": Invalid Semantic Version"),
		},
		{
			name:                  "Executor release branch build",
			executorVersion:       "5.1_231128_2023-06-27_5.0-7ac9ba347103",
			sourcegraphVersion:    "5.0.3",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Sourcegraph release branch build",
			executorVersion:       "5.0.3",
			sourcegraphVersion:    "5.1_231128_2023-06-27_5.0-7ac9ba347103",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor release candidate",
			executorVersion:       "5.1.3-rc.1",
			sourcegraphVersion:    "5.1.3",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
		},
		{
			name:                  "Executor version missing patch",
			executorVersion:       "5.1",
			sourcegraphVersion:    "5.1.3",
			isActive:              true,
			expectedCompatibility: ExecutorCompatibilityUpToDate,
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
				// Once https://github.com/stretchr/testify/pull/1287 is merged, we can remove this and just use Equal.
				// When they are not equal we are just given the addresses which doesn't mean much to us, and tell us
				// how to fix the test.
				if test.expectedCompatibility != "" {
					require.NotNil(t, actual)
					assert.Equal(t, test.expectedCompatibility, ExecutorCompatibility(*actual))
				} else {
					assert.Nil(t, actual)
				}
			}
		})
	}
}
