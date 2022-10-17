package graphqlbackend

import (
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
			expected           *string
			description        string
		}{
			// The executor isn't outdated when in dev mode.
			{
				executorVersion:    "0.0.0+dev",
				sourcegraphVersion: "0.0.0+dev",
				isActive:           true,
				expected:           nil,
				description:        "executor and the Sourcegraph instance are in dev mode",
			},
			// The executor isn't outdated when it's inactive
			{
				executorVersion:    "0.0.0+dev",
				sourcegraphVersion: "0.0.0+dev",
				isActive:           false,
				expected:           nil,
				description:        "executor is inactive",
			},
			// The executor is not outdated if it's one minor version behind the sourcegraph version (SEMVER)
			{
				executorVersion:    "3.43.0",
				sourcegraphVersion: "3.42.0",
				isActive:           true,
				expected:           ExecutorCompatibilityUpToDate.ToGraphQL(),
				description:        "executor is one version ahead of the Sourcegraph instance",
			},
			// The executor is not outdated if it's one minor version ahead of the sourcegraph version (SEMVER)
			{
				executorVersion:    "3.42.0",
				sourcegraphVersion: "3.43.0",
				isActive:           true,
				expected:           ExecutorCompatibilityUpToDate.ToGraphQL(),
				description:        "executor is one minor version ahead of the Sourcegraph instance",
			},
			// The executor is not outdated when both sourcegraph and executor are the same (SEMVER).
			{
				executorVersion:    "3.43.0",
				sourcegraphVersion: "3.43.0",
				isActive:           true,
				expected:           ExecutorCompatibilityUpToDate.ToGraphQL(),
				description:        "executor and the sourcegraph instance are the same version (SEMVER)",
			},
			// The executor is not outdated when both sourcegraph and executor are the same (insiders).
			{
				executorVersion:    "executor-patch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_patch",
				sourcegraphVersion: "169135_2022-08-25_a2b623dce148",
				isActive:           true,
				expected:           ExecutorCompatibilityUpToDate.ToGraphQL(),
				description:        "executor and the sourcegraph instance are the same version (insiders)",
			},
			// The executor is outdated if the sourcegraph version is more than one version
			// greater than the executor version (SEMVER).
			{
				executorVersion:    "3.40.0",
				sourcegraphVersion: "3.43.0",
				isActive:           true,
				expected:           ExecutorCompatibilityOutdated.ToGraphQL(),
				description:        "executor is more than one minor version behind the Sourcegraph instance (SEMVER)",
			},
			{
				executorVersion:    "3.43.0",
				sourcegraphVersion: "4.0.0",
				isActive:           true,
				expected:           ExecutorCompatibilityOutdated.ToGraphQL(),
				description:        "Sourcegraph instance is a major version ahead of the executor",
			},
			// The executor is too new if the executor is more than one version ahead of the sourcegraph version.
			{
				executorVersion:    "3.43.0",
				sourcegraphVersion: "3.40.0",
				isActive:           true,
				expected:           ExecutorCompatibilityVersionAhead.ToGraphQL(),
				description:        "executor is more than one minor version ahead of the Sourcegraph instance (SEMVER)",
			},
			{
				executorVersion:    "4.0.0",
				sourcegraphVersion: "3.43.0",
				isActive:           true,
				expected:           ExecutorCompatibilityVersionAhead.ToGraphQL(),
				description:        "executor is a major version ahead of the Sourcegraph instance",
			},
			// The executor is outdated if the sourcegraph version is greater than the executor version (insiders).
			{
				executorVersion:    "executor-patch-notest-es-ignite-debug_168065_2022-06-10_e94e18c4ebcc_patch",
				sourcegraphVersion: "169135_2022-07-25_a2b623dce148",
				isActive:           true,
				expected:           ExecutorCompatibilityOutdated.ToGraphQL(),
				description:        "executor version is less than the Sourcegraph build date - one release cycle  (insiders)",
			},
			// The executor is too new if the executor version is greater than the one release cycle + sourcegraph build date (insiders)
			{
				executorVersion:    "executor-patch-notest-es-ignite-debug_168065_2022-10-30_e94e18c4ebcc_patch",
				sourcegraphVersion: "169135_2022-09-15_a2b623dce148",
				isActive:           true,
				expected:           ExecutorCompatibilityVersionAhead.ToGraphQL(),
				description:        "executor version is greater than the one release cycle + Sourcegraph build date (insiders)",
			},
			// The executor is up to date if the build date isn't greater than one release cycle + sourcegraph build date (insiders)
			{
				executorVersion:    "executor-patch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_patch",
				sourcegraphVersion: "169135_2022-08-15_a2b623dce148",
				isActive:           true,
				expected:           ExecutorCompatibilityUpToDate.ToGraphQL(),
				description:        "executor version is a few days ahead of the Sourcegraph instance (insiders)",
			},
			// version mismatch
			{
				executorVersion:    "3.36.2",
				sourcegraphVersion: "169135_2022-08-15_a2b623dce148",
				isActive:           true,
				expected:           nil,
				description:        "Sourcegraph instance is on insiders version while the executor is tagged",
			},
			{
				executorVersion:    "169135_2022-08-15_a2b623dce148",
				sourcegraphVersion: "3.39.2",
				isActive:           true,
				expected:           nil,
				description:        "executor is on insiders version while the Sourcegraph instance is tagged",
			},
			{
				executorVersion:    "0.0.0+dev",
				sourcegraphVersion: "3.39.2",
				isActive:           true,
				expected:           nil,
				description:        "executor is in dev mode while Sourcegraph instance is tagged",
			},
			{
				executorVersion:    "3.39.2",
				sourcegraphVersion: "0.0.0+dev",
				isActive:           true,
				expected:           nil,
				description:        "Sourcegraph instance is in dev mode while executor is tagged",
			},
			{
				executorVersion:    "0.0.0+dev",
				sourcegraphVersion: "169135_2022-08-15_a2b623dce148",
				isActive:           true,
				expected:           nil,
				description:        "executor is in dev mode while Sourcegraph instance is on insiders version",
			},
			{
				executorVersion:    "169135_2022-08-15_a2b623dce148",
				sourcegraphVersion: "0.0.0+dev",
				isActive:           true,
				expected:           nil,
				description:        "Sourcegraph instance is in dev mode while executor is on insiders version",
			},
		}

		for _, tc := range testCases {
			version.Mock(tc.sourcegraphVersion)
			want, err := calculateExecutorCompatibility(tc.executorVersion)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, want, tc.description)
		}
	})
}
