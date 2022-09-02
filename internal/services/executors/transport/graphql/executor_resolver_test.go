package graphql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

func TestExecutorResolver(t *testing.T) {
	t.Run("isExecutorOutdated", func(t *testing.T) {
		testCases := []struct {
			executorVersion    string
			sourcegraphVersion string
			isActive           bool
			expected           bool
		}{
			// The executor isn't outdated when in dev mode.
			{
				executorVersion:    "0.0.0+dev",
				sourcegraphVersion: "0.0.0+dev",
				isActive:           true,
				expected:           false,
			},
			// The executor isn't outdated when it's inactive
			{
				executorVersion:    "0.0.0+dev",
				sourcegraphVersion: "0.0.0+dev",
				isActive:           false,
				expected:           false,
			},
			{
				executorVersion:    "3.43.0",
				sourcegraphVersion: "3.42.0",
				isActive:           false,
				expected:           false,
			},
			// The executor is not outdated when both sourcegraph and executor are the same (SEMVER).
			{
				executorVersion:    "3.43.0",
				sourcegraphVersion: "3.43.0",
				isActive:           true,
				expected:           false,
			},
			// The executor isn"t outdated when both sourcegraph and executor are the same (BuildDate).
			{
				executorVersion:    "executor-patch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_patch",
				sourcegraphVersion: "169135_2022-08-25_a2b623dce148",
				isActive:           true,
				expected:           false,
			},
			// The executor is outdated if the sourcegraph version is greater than theexecutor version (SEMVER).
			{
				executorVersion:    "3.42.0",
				sourcegraphVersion: "3.43.0",
				isActive:           true,
				expected:           true,
			},
			// The executor is outdated if the sourcegraph version is greater than the executor version (BuildDate).
			{
				executorVersion:    "executor-patch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_patch",
				sourcegraphVersion: "169135_2022-08-25_a2b623dce148",
				isActive:           true,
				expected:           true,
			},
			// The executor is not outdated if the executor version is greater than the sourcegraph version (SEMVER)
			{
				executorVersion:    "3.43.0",
				sourcegraphVersion: "3.42.0",
				isActive:           true,
				expected:           false,
			},
			// The executor is not outdated if the executor version is greater than the sourcegraph version (BuildDate)
			{
				executorVersion:    "executor-patch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_patch",
				sourcegraphVersion: "169135_2022-08-15_a2b623dce148",
				isActive:           true,
				expected:           false,
			},
		}

		for _, tc := range testCases {
			version.Mock(tc.sourcegraphVersion)
			want, err := isExecutorOutdated(tc.executorVersion, tc.isActive)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, want, fmt.Sprintf("ev: %s - expected: %t, got: %t", tc.executorVersion, tc.expected, want))
		}
	})
}
